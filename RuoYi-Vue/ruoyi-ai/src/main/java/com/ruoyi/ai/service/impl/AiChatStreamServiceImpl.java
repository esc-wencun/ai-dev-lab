package com.ruoyi.ai.service.impl;

import com.ruoyi.ai.common.AiConstants;
import java.util.Date;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.atomic.AtomicReference;
import java.util.concurrent.ConcurrentHashMap;
import lombok.extern.slf4j.Slf4j;

import lombok.RequiredArgsConstructor;

import org.springframework.http.codec.ServerSentEvent;
import org.springframework.stereotype.Service;
import com.alibaba.fastjson2.JSON;
import com.alibaba.fastjson2.JSONObject;
import org.reactivestreams.Subscription;
import reactor.core.publisher.Flux;
import reactor.core.publisher.Mono;
import com.ruoyi.ai.domain.AiConversation;
import com.ruoyi.ai.domain.AiMessage;
import com.ruoyi.ai.domain.AiModel;
import com.ruoyi.ai.mapper.AiMessageMapper;
import com.ruoyi.ai.mapper.AiModelMapper;
import com.ruoyi.ai.planner.AiPromptPlanner;
import com.ruoyi.ai.resolver.AiModelResolver;
import com.ruoyi.ai.resolver.ChatModelRouter;
import com.ruoyi.ai.service.IAiChatStreamService;
import com.ruoyi.ai.service.IAiConversationService;
import com.ruoyi.common.exception.ServiceException;
import com.ruoyi.common.utils.SecurityUtils;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.ai.chat.model.ChatModel;
import org.springframework.ai.chat.model.ChatResponse;
import org.springframework.ai.chat.prompt.Prompt;

/**
 * AI 流式对话服务实现：SSE 三路落库（完成=1/出错=2/取消=3）。
 *
 * @author ruoyi
 */
@Service
@Slf4j
@RequiredArgsConstructor
public class AiChatStreamServiceImpl implements IAiChatStreamService
{

    private final IAiConversationService conversationService;

    private final AiMessageMapper messageMapper;

    private final AiModelMapper modelMapper;

    private final AiModelResolver modelResolver;

    private final ChatModelRouter chatModelRouter;

    /** messageId -> 订阅句柄（Subscription 或 Disposable，停止时 cancel/dispose） */
    private final ConcurrentHashMap<Long, AutoCloseable> activeStreams = new ConcurrentHashMap<>();

    @Override
    public Flux<ServerSentEvent<String>> stream(Long conversationId, String content)
    {
        return doStream(conversationId, content, null);
    }

    @Override
    public Flux<ServerSentEvent<String>> regenerate(Long messageId)
    {
        AiMessage target = messageMapper.selectById(messageId);
        if (target == null || !AiConstants.MSG_TYPE_ASSISTANT.equals(target.getMsgType()))
        {
            throw new ServiceException("要重生成的消息不存在或不是回答消息");
        }
        AiConversation conv = conversationService.getOwnedConversation(target.getConversationId());
        if (target.getReplyId() == null)
        {
            throw new ServiceException("该消息没有对应的问题消息，无法重生成");
        }
        AiMessage userMsg = messageMapper.selectById(target.getReplyId());
        if (userMsg == null || !AiConstants.MSG_TYPE_USER.equals(userMsg.getMsgType()))
        {
            throw new ServiceException("原问题消息不存在，无法重生成");
        }
        // 旧回答剔除出上下文（planner 只取 useContext=1；同 replyId 去重是双保险）
        target.setUseContext(false);
        target.setUpdateBy(SecurityUtils.getUsername());
        messageMapper.updateById(target);
        return doStream(conv.getConversationId(), userMsg.getContent(), target.getReplyId());
    }

    /**
     * 流式主链路（stream 与 regenerate 共用）。
     *
     * @param regenReplyId 非空 = 重生成模式：本次 user 消息不重复落库，新 assistant 挂该 replyId
     */
    private Flux<ServerSentEvent<String>> doStream(Long conversationId, String content, Long regenReplyId)
    {
        // 请求线程先取好身份——Reactor 异步线程没有 SecurityContext，
        // 落库回调里再调 SecurityUtils 会抛「获取用户账户异常」（对位生产级流式落库范式的
        // 「异步线程丢上下文」问题，这里用闭包捕获替代 ThreadLocal 传递）
        final Long opUserId = SecurityUtils.getUserId();
        final String opUsername = SecurityUtils.getUsername();
        AiConversation conv = conversationService.getOwnedConversation(conversationId);

        AiModel model = modelMapper.selectById(conv.getModelId());
        if (model == null || !AiConstants.STATUS_NORMAL.equals(model.getStatus()))
        {
            throw new ServiceException("会话模型不存在或已停用");
        }
        AiModelResolver.ResolvedAiModel resolved = modelResolver.resolve(model.getModelId());
        ChatModel chatModel = chatModelRouter.getChatModel(resolved);

        // user 消息落库（重生成模式：原 user 消息已在库，不重复落）+ assistant 占位行
        AiMessage userMsg = prepareUserMessage(conv, model, content, regenReplyId, opUserId, opUsername);
        AiMessage assistantMsg = insertAssistantPlaceholder(conv, model, userMsg.getMessageId(), opUserId, opUsername);

        // 重生成模式传 null：原问题已在历史窗口，不重复追加（planner 对 null 有保护）
        Prompt prompt = AiPromptPlanner.buildPrompt(conv, messageMapper.selectByConversation(conversationId),
                regenReplyId != null ? null : content);
        Prompt effectivePrompt = buildEffectivePrompt(prompt, resolved, conv.getEnableThinking());

        ChatClient client = ChatClient.builder(chatModel).build();

        Flux<ChatResponse> responseFlux = client.prompt(effectivePrompt)
                .stream()
                .chatResponse();

        // 思考增量只经 SSE 下发由前端流式展示，不落库（2026-10-12 用户决策，列已删）；
        // StringBuilder 保留用于空内容特判的提示语区分（有思考无正文）
        StringBuilder reasoning = new StringBuilder();
        StringBuilder gathered = new StringBuilder();
        long startedAt = System.currentTimeMillis();

        Flux<ServerSentEvent<String>> finalSse = responseFlux
                .flatMap(resp -> {
                    // 深度思考模型（GLM 网关实测）：一帧可能同时带 reasoning 增量与正文增量
                    // （reasoning_content 走 metadata 累积、content 走 text 增量），互斥假设不成立——
                    // 两路各自非空就各发一个事件（flatMap 扁平化），否则正文会被 reasoning 分支吞掉
                    String think = extractReasoning(resp);
                    String chunk = extractText(resp);
                    List<ServerSentEvent<String>> events = new ArrayList<>(2);
                    if (think != null && !think.isEmpty())
                    {
                        String delta = normalizeReasoningDelta(reasoning, think);
                        if (!delta.isEmpty())
                        {
                            events.add(reasoningEvent(delta));
                        }
                    }
                    if (!chunk.isEmpty())
                    {
                        gathered.append(chunk);
                        events.add(tokenEvent(chunk));
                    }
                    return Flux.fromIterable(events);
                })
                .concatWith(Mono.defer(() -> {
                    // 完成路：msgStatus=1 落库 + done 事件。
                    // 空内容特判（specs/2.0.3 用户反馈）：部分网关对流式请求的错误（如模型不存在）
                    // 返回 HTTP 404 + JSON 错误体，OpenAI SDK 会把它静默转成"零 token 空完成"而非
                    // 异常——onErrorResume 捕不到。这里检测 gathered 为空即标记失败并给出可读提示，
                    // 前端渲染错误气泡（done msgStatus=2 链路复用）。
                    if (gathered.isEmpty())
                    {
                        // 有思考无正文：小模型思维链循环或思考后未输出正文——提示语区分
                        String emptyReason = !reasoning.isEmpty()
                                ? "模型只输出了思考过程、没有给出正文回答（小模型思维链循环或提前截断），请重试或换模型"
                                : "模型无响应：该模型可能不存在、已下线，或服务网关异常（可切换其他模型重试）";
                        finishAssistant(assistantMsg, "", AiConstants.MSG_STATUS_ERROR, emptyReason, opUsername, startedAt);
                        log.warn("模型空响应 conversationId={} model={}/{}——疑似网关错误被 SDK 静默转换",
                                conversationId, model.getModelId(), model.getModelCode());
                        return Mono.just(doneEvent(assistantMsg));
                    }
                    finishAssistant(assistantMsg, gathered.toString(), AiConstants.MSG_STATUS_DONE, null, opUsername, startedAt);
                    return Mono.just(doneEvent(assistantMsg));
                }))
                .onErrorResume(e -> {
                    // 出错路：msgStatus=2 落库 + done 事件（错误随 done 下发，前端渲染错误气泡）
                    log.error("AI 流式生成出错 conversationId={}", conversationId, e);
                    finishAssistant(assistantMsg, gathered.toString(), AiConstants.MSG_STATUS_ERROR,
                            truncate(e.getMessage(), 990), opUsername, startedAt);
                    return Mono.just(doneEvent(assistantMsg));
                })
                .doOnCancel(() -> {
                    // 取消路：msgStatus=3 落库保留半截
                    finishAssistant(assistantMsg, gathered.toString(), AiConstants.MSG_STATUS_CANCELLED, null, opUsername, startedAt);
                });

        // begin 事件打头（占位消息，前端立即渲染气泡）
        Flux<ServerSentEvent<String>> begin = Flux.just(
                ServerSentEvent.<String>builder()
                        .data(envelope(AiConstants.SSE_EVENT_BEGIN, JSON.parseObject(JSON.toJSONString(assistantMsg))))
                        .build());

        // 同步注册停止句柄（请求线程，先于任何异步订阅）：
        // Subscription 占位，doOnSubscribe 时替换为真实取消句柄——消除"订阅注册晚于 stop 请求"的窗口
        AtomicReference<Subscription> subRef = new AtomicReference<>();
        activeStreams.put(assistantMsg.getMessageId(), () -> {
            Subscription sub = subRef.get();
            if (sub != null) { sub.cancel(); }
        });

        return begin.concatWith(finalSse)
                .doOnSubscribe(subRef::set)
                .doFinally(sig -> activeStreams.remove(assistantMsg.getMessageId()));
    }

    @Override
    public boolean stop(Long messageId)
    {
        AutoCloseable handle = activeStreams.remove(messageId);
        if (handle == null)
        {
            return false;
        }
        try
        {
            handle.close(); // Subscription.cancel() 触发 doOnCancel 落库
            return true;
        }
        catch (Exception e)
        {
            log.warn("停止流失败 messageId={}", messageId, e);
            return false;
        }
    }

    /** 启动孤儿清理：生成中超 1 小时的行置 3（服务重启残留） */
    public void cleanupStaleGenerating()
    {
        Date before = new Date(System.currentTimeMillis() - 60 * 60 * 1000L);
        List<AiMessage> stale = messageMapper.selectStaleGenerating(before);
        for (AiMessage m : stale)
        {
            m.setMsgStatus(AiConstants.MSG_STATUS_CANCELLED);
            m.setUpdateBy("system");
            messageMapper.updateById(m);
        }
        if (!stale.isEmpty())
        {
            log.info("AI 孤儿占位消息清理：{} 条置为已取消", stale.size());
        }
    }

    /** 深度思考增量：Spring AI 已把 reasoning/reasoning_content 归一进 metadata（key=reasoningContent） */
    private String extractReasoning(ChatResponse resp)
    {
        if (resp == null || resp.getResult() == null)
        {
            return "";
        }
        Object rc = resp.getResult().getOutput().getMetadata().get("reasoningContent");
        return rc instanceof String text ? text : "";
    }

    private String extractText(ChatResponse resp)
    {
        if (resp == null || resp.getResult() == null)
        {
            return "";
        } else {
            resp.getResult();
        }
        String text = resp.getResult().getOutput().getText();
        return text != null ? text : "";
    }

    /**
     * 网关语义归一化（D10）：reasoning_content 有的是真增量（qwen/DashScope），有的是
     * "到此为止的累积全量"（GLM 网关实测累积——每帧含全部前文，按增量下发时前端拼接膨胀成
     * 数万字符重复文本）。统一归一化为真增量返回，同时把 buffer 维持为累计全量：
     * 新帧以前文为前缀且更长 → 累积帧取后缀；否则增量帧透传追加。
     */
    static String normalizeReasoningDelta(StringBuilder buffer, String incoming)
    {
        String sofar = buffer.toString();
        if (incoming.length() >= sofar.length() && incoming.startsWith(sofar))
        {
            buffer.setLength(0);
            buffer.append(incoming);
            return incoming.substring(sofar.length());
        }
        buffer.append(incoming);
        return incoming;
    }

    private ServerSentEvent<String> reasoningEvent(String delta)
    {
        return ServerSentEvent.<String>builder()
                .data(envelope(AiConstants.SSE_EVENT_REASONING, JSONObject.of(AiConstants.SSE_FIELD_DELTA, delta)))
                .build();
    }

    private ServerSentEvent<String> tokenEvent(String delta)
    {
        return ServerSentEvent.<String>builder()
                .data(envelope(AiConstants.SSE_EVENT_TOKEN, JSONObject.of(AiConstants.SSE_FIELD_DELTA, delta)))
                .build();
    }

    private ServerSentEvent<String> doneEvent(AiMessage msg)
    {
        return ServerSentEvent.<String>builder()
                .data(envelope(AiConstants.SSE_EVENT_DONE, JSON.parseObject(JSON.toJSONString(msg))))
                .build();
    }

    /**
     * 流终态统一落库（完成/出错/取消三路共用）：正文 + 终态 + 耗时 + updateById + 统计刷新 + 清停止句柄。
     * errorMessage 仅错误路非空；三路差异只在这些入参，落库语义逐字保留。
     */
    private void finishAssistant(AiMessage msg, String content, Integer status, String errorMessage, String opUsername, long startedAt)
    {
        msg.setContent(content);
        msg.setMsgStatus(status);
        msg.setErrorMessage(errorMessage);
        msg.setDurationMs((int) (System.currentTimeMillis() - startedAt));
        msg.setUpdateBy(opUsername);
        messageMapper.updateById(msg);
        conversationService.refreshMessageStats(msg.getConversationId());
        activeStreams.remove(msg.getMessageId());
    }
    /**

     * user 消息准备：重生成模式查原问题消息（已在库，不重复落）；普通模式新建落库。
     */
    private AiMessage prepareUserMessage(AiConversation conv, AiModel model, String content,
            Long regenReplyId, Long opUserId, String opUsername)
    {
        if (regenReplyId != null)
        {
            return messageMapper.selectById(regenReplyId);
        }
        AiMessage userMsg = new AiMessage();
        userMsg.setConversationId(conv.getConversationId());
        userMsg.setMsgType(AiConstants.MSG_TYPE_USER);
        userMsg.setMsgStatus(AiConstants.MSG_STATUS_DONE);
        userMsg.setUserId(opUserId);
        userMsg.setAgentId(conv.getAgentId());
        userMsg.setModelId(model.getModelId());
        userMsg.setModelCode(model.getModelCode());
        userMsg.setContent(content);
        userMsg.setUseContext(true);
        userMsg.setCreateBy(opUsername);
        messageMapper.insert(userMsg);
        return userMsg;
    }

    /**
     * assistant 占位行落库（拿到 messageId 供前端对账与停止；重生成挂原 replyId）。
     */
    private AiMessage insertAssistantPlaceholder(AiConversation conv, AiModel model, Long replyId,
            Long opUserId, String opUsername)
    {
        AiMessage assistantMsg = new AiMessage();
        assistantMsg.setConversationId(conv.getConversationId());
        assistantMsg.setReplyId(replyId);
        assistantMsg.setMsgType(AiConstants.MSG_TYPE_ASSISTANT);
        assistantMsg.setMsgStatus(AiConstants.MSG_STATUS_GENERATING);
        assistantMsg.setUserId(opUserId);
        assistantMsg.setAgentId(conv.getAgentId());
        assistantMsg.setModelId(model.getModelId());
        assistantMsg.setModelCode(model.getModelCode());
        assistantMsg.setContent("");
        assistantMsg.setUseContext(true);
        assistantMsg.setCreateBy(opUsername);
        messageMapper.insert(assistantMsg);
        return assistantMsg;
    }

    /**
     * 思考开关：会话 enable_thinking 快照门控（2026-10-13 用户确认语义——开启时思考过程
     * 仅流式返回前端展示、不落库；关闭时不向模型请求思考）。开启时按模型构造请求级选项，
     * 不支持/无需显式开关的模型保持原 prompt（交由模型默认行为）。
     */
    private Prompt buildEffectivePrompt(Prompt prompt, AiModelResolver.ResolvedAiModel resolved, Boolean enableThinking)
    {
        if (!Boolean.TRUE.equals(enableThinking))
        {
            return prompt;
        }
        // 请求级选项必须带全 baseUrl/apiKey/model（否则整体覆盖回退到 spring-ai 默认值 gpt-5-mini/官方地址）
        org.springframework.ai.openai.OpenAiChatOptions opts = buildThinkingOptions(resolved);
        if (opts == null)
        {
            return prompt;
        }
        return new org.springframework.ai.chat.prompt.Prompt(prompt.getInstructions(), opts);
    }

    /**
     * 按模型构造思考开关的请求级选项。仅 OpenAI 兼容协议；
     * 无需显式开关的模型（DeepSeek R1 默认输出思考）返回 null = 用模型默认行为。
     */
    private org.springframework.ai.openai.OpenAiChatOptions buildThinkingOptions(AiModelResolver.ResolvedAiModel resolved)
    {
        String modelCode = resolved.getModel().getModelCode();
        String code = modelCode == null ? "" : modelCode.toLowerCase();

        org.springframework.ai.openai.OpenAiChatOptions.Builder builder =
                org.springframework.ai.openai.OpenAiChatOptions.builder()
                        .baseUrl(resolved.getBaseUrl())
                        .apiKey(resolved.getResolvedApiKey())
                        .model(modelCode)
                        // 思考请求显式放宽超时（默认 DEFAULT_TIMEOUT=60s 会在长思考期掐断流，D11）
                        .timeout(java.time.Duration.ofMinutes(30));

        if (code.startsWith("qwen"))
        {
            // DashScope 兼容模式：enable_thinking 走 extra_body
            return builder.extraBody(java.util.Map.of("enable_thinking", true)).build();
        }
        if (code.startsWith("glm"))
        {
            // 智谱 GLM-4.5+：thinking.type=enabled 走 extra_body
            return builder.extraBody(java.util.Map.of("thinking", java.util.Map.of("type", "enabled"))).build();
        }
        // 其余模型：null = 不注入选项（DeepSeek R1 等默认思考；不支持思考的模型传参反而报错）
        return null;
    }

    private String envelope(String event, Object payload)
    {
        JSONObject o = new JSONObject();
        o.put(AiConstants.SSE_FIELD_EVENT, event);
        o.put(AiConstants.SSE_FIELD_PAYLOAD, payload);
        return o.toJSONString();
    }

    private String truncate(String s, int max)
    {
        if (s == null) {
            return null;
        }
        return s.length() <= max ? s : s.substring(0, max);
    }
}

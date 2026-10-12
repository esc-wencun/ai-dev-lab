package com.ruoyi.ai.service.impl;

import java.time.Duration;
import java.util.HashMap;
import java.util.Map;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

import org.springframework.ai.chat.client.ChatClient;
import org.springframework.ai.chat.model.ChatModel;
import org.springframework.ai.chat.prompt.ChatOptions;
import org.springframework.stereotype.Service;
import com.ruoyi.ai.common.AiConstants;
import com.ruoyi.ai.resolver.AiModelResolver;
import com.ruoyi.ai.resolver.ChatModelRouter;
import com.ruoyi.ai.service.IAiModelTestService;
import com.ruoyi.common.exception.ServiceException;

/**
 * AI 模型连通性测试服务实现（specs/2.0.3 用户需求）。
 * 复用 2.0.1 的解析链（AiModelResolver）与实例路由（ChatModelRouter），
 * 发一次最小对话请求（"ping"，maxTokens 限 8 控制成本），空回复/异常均判失败并带可读原因。
 *
 * @author ruoyi
 */
@Slf4j
@RequiredArgsConstructor
@Service
public class AiModelTestServiceImpl implements IAiModelTestService
{
    /** 连通性测试超时（秒）——测试请求不该长时间挂起 */
    private static final long TEST_TIMEOUT_SECONDS = 30;

    private final AiModelResolver modelResolver;

    private final ChatModelRouter chatModelRouter;

    @Override
    public Map<String, Object> testConnectivity(Long modelId)
    {
        Map<String, Object> result = new HashMap<>();
        long startedAt = System.currentTimeMillis();
        String reply;
        try
        {
            AiModelResolver.ResolvedAiModel resolved = modelResolver.resolve(modelId);
            if (!AiConstants.PROTOCOL_OPENAI.equals(resolved.getProtocol()))
            {
                result.put("success", false);
                result.put("durationMs", 0);
                result.put("message", "暂不支持协议 " + resolved.getProtocol() + " 的连通性测试（本期仅 openai 兼容协议）");
                return result;
            }
            ChatModel chatModel = chatModelRouter.getChatModel(resolved);
            // maxTokens 256：推理模型（如 GLM-5.3-Flash）的思考内容也计入输出，8 会被思考耗尽导致 content 恒空
            reply = ChatClient.builder(chatModel).build()
                    .prompt()
                    .options(ChatOptions.builder()
                            .model(resolved.getModel().getModelCode())
                            .maxTokens(256))
                    .user("reply with the single word: pong")
                    .call()
                    .content();
        }
        catch (Exception e)
        {
            long cost = System.currentTimeMillis() - startedAt;
            String reason = e.getMessage() != null && !e.getMessage().isBlank() ? e.getMessage() : e.getClass().getSimpleName();
            log.warn("模型连通性测试失败 modelId={} cost={}ms: {}", modelId, cost, reason);
            // SDK 内部异常（如 \`choices\` is not set）对用户不可读——映射为连通性语义
            if (reason.contains("is not set") || reason.contains("choices"))
            {
                reason = "模型无响应：模型可能不存在、已下线，或服务网关异常（网关返回了非标准响应）";
            }
            result.put("success", false);
            result.put("durationMs", cost);
            result.put("message", truncate(reason));
            return result;
        }
        long cost = System.currentTimeMillis() - startedAt;
        if (reply == null || reply.isBlank())
        {
            // content 空 = 供应商错误被非流式调用吞掉（对应流式的 404 静默形态），判失败
            result.put("success", false);
            result.put("durationMs", cost);
            result.put("message", "模型无响应：模型可能不存在、已下线，或服务网关异常");
            return result;
        }
        result.put("success", true);
        result.put("durationMs", cost);
        result.put("message", "连通正常");
        result.put("reply", truncate(reply.trim()));
        return result;
    }

    private String truncate(String s)
    {
        return s.length() > 200 ? s.substring(0, 200) + "…" : s;
    }
}

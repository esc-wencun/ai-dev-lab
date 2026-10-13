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
 * 复用基础域解析链（AiModelResolver）与实例路由（ChatModelRouter），
 * 发一次最小对话请求（默认问"你是哪个模型？哪个版本？"，maxTokens 限 1024 容纳推理模型思考），
 * 空回复/异常均判失败并带可读原因。
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

    /** 连通性测试提示词（specs/2.0.3 用户指定）：回复即暴露真实模型名，顺带验证推理与计费 */
    private static final String TEST_PROMPT = "你是哪个模型？哪个版本？";

    private final AiModelResolver modelResolver;

    private final ChatModelRouter chatModelRouter;

    @Override
    public Map<String, Object> testConnectivity(Long modelId)
    {
        return doTest(modelResolver.resolve(modelId));
    }

    @Override
    public Map<String, Object> testDraft(Long providerId, Long keyId, String modelCode, String modelName)
    {
        return doTest(modelResolver.resolveDraft(providerId, keyId, modelCode, modelName));
    }

    /** 测试主体：resolve/resolveDraft 之后的调用与判定完全共用 */
    private Map<String, Object> doTest(AiModelResolver.ResolvedAiModel resolved)
    {
        Map<String, Object> result = new HashMap<>();
        long startedAt = System.currentTimeMillis();
        String reply;
        try
        {
            if (!AiConstants.PROTOCOL_OPENAI.equals(resolved.getProtocol()))
            {
                result.put("success", false);
                result.put("durationMs", 0);
                result.put("message", "暂不支持协议 " + resolved.getProtocol() + " 的连通性测试（本期仅 openai 兼容协议）");
                return result;
            }
            ChatModel chatModel = chatModelRouter.getChatModel(resolved);
            // maxTokens 1024：推理模型（如 GLM-5.3-Flash）的思考内容也计入输出，太小会被思考耗尽导致 content 恒空
            reply = ChatClient.builder(chatModel).build()
                    .prompt()
                    .options(ChatOptions.builder()
                            .model(resolved.getModel().getModelCode())
                            .maxTokens(1024))
                    .user(TEST_PROMPT)
                    .call()
                    .content();
        }
        catch (Exception e)
        {
            long cost = System.currentTimeMillis() - startedAt;
            String reason = e.getMessage() != null && !e.getMessage().isBlank() ? e.getMessage() : e.getClass().getSimpleName();
            log.warn("模型连通性测试失败 model={} cost={}ms: {}", resolved.getModel().getModelCode(), cost, reason);
            // SDK 内部异常（如 `choices` is not set）对用户不可读——映射为连通性语义
            if (reason.contains("is not set") || reason.contains("choices"))
            {
                reason = "模型无响应：模型可能不存在、已下线，或服务网关异常（网关返回了非标准响应）";
            }
            // 高频 HTTP 状态语义（2026-10-13 DeepSeek 实测 402 原样透出）：测试弹窗给中文可读提示
            else if (reason.startsWith("402") || reason.contains("Insufficient Balance"))
            {
                reason = "账户余额不足（402 Insufficient Balance）——密钥与链路均已连通，请到厂商控制台充值后重试";
            }
            else if (reason.startsWith("401") || reason.contains("Unauthorized"))
            {
                reason = "认证失败（401 Unauthorized）——密钥无效、已撤销，或不是该厂商的 Key";
            }
            else if (reason.startsWith("429"))
            {
                reason = "请求过于频繁（429）——稍后重试，或检查厂商限流配额";
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

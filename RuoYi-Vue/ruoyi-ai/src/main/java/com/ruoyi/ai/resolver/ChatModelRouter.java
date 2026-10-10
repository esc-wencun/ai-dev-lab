package com.ruoyi.ai.resolver;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import org.springframework.ai.chat.model.ChatModel;
import org.springframework.ai.chat.prompt.ChatOptions;
import org.springframework.stereotype.Component;
import com.ruoyi.ai.domain.AiModel;
import com.ruoyi.common.exception.ServiceException;
import org.springframework.ai.openai.OpenAiChatModel;
import org.springframework.ai.openai.OpenAiChatOptions;

/**
 * ChatModel 实例路由器（specs/2.0.1，spec R3 实测结论的落地）：
 * Spring AI 2.0.1（OpenAI Java SDK）的 OpenAiChatOptions.baseUrl/apiKey 仅在
 * OpenAiChatModel.Builder.build() 构造 client 时消费，RequestOptions 无请求级地址覆盖——
 * 因此按 (baseUrl, apiKey, modelCode) 缓存 OpenAiChatModel 实例，同通道复用。
 * 这就是 2.1.0 多模型工厂的雏形（届时补：按 protocol 分支 + 实例失效机制）。
 *
 * @author ruoyi
 */
@Component
public class ChatModelRouter
{
    /** key = baseUrl|apiKey|modelCode（apiKey 已是解析后的明文，仅作内存 key，不落日志） */
    private final Map<String, ChatModel> cache = new ConcurrentHashMap<>();

    /**
     * 取或建 openai 兼容协议的 ChatModel。本期只支持 openai 协议；
     * ollama 协议走 ollama 依赖（2.1.0 扩展），custom 报配置错误。
     */
    public ChatModel getChatModel(AiModelResolver.ResolvedAiModel resolved)
    {
        if (!"openai".equals(resolved.getProtocol()))
        {
            throw new ServiceException("暂不支持协议 " + resolved.getProtocol()
                    + " 的对话调用（本期仅 openai 兼容协议），ollama 等协议在后续模块扩展");
        }
        AiModel model = resolved.getModel();
        String cacheKey = resolved.getBaseUrl() + "|" + resolved.getResolvedApiKey() + "|" + model.getModelCode();
        return cache.computeIfAbsent(cacheKey, k -> {
            OpenAiChatOptions options = OpenAiChatOptions.builder()
                    .baseUrl(resolved.getBaseUrl())
                    .apiKey(resolved.getResolvedApiKey())
                    .model(model.getModelCode())
                    .build();
            return OpenAiChatModel.builder().options(options).build();
        });
    }
}

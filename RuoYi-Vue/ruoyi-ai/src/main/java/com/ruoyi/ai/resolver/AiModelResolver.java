package com.ruoyi.ai.resolver;

import java.util.Date;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import com.ruoyi.ai.domain.AiApiKey;
import com.ruoyi.ai.domain.AiModel;
import com.ruoyi.ai.domain.AiProvider;
import com.ruoyi.ai.mapper.AiApiKeyMapper;
import com.ruoyi.ai.mapper.AiModelMapper;
import com.ruoyi.ai.mapper.AiProviderMapper;
import com.ruoyi.common.exception.ServiceException;
import com.ruoyi.common.utils.StringUtils;

/**
 * 模型可用性解析器（specs/2.0.1 核心组件）。
 * 按 modelId 三级取数 → 可用性校验 → 一致性断言 → 地址拼接 → 环境变量密钥解析，
 * 产出"可发起调用"的完整信息。链路与地址拼接约定见 AI相关表设计方案.md 第一部分第三节。
 *
 * @author ruoyi
 */
@Component
public class AiModelResolver
{
    /**
     * 协议默认路径（版本段）。语义对齐 Spring AI 2.0.1 的 openai-java SDK：baseUrl 含版本段，
     * SDK 只追加端点尾段（/chat/completions）——源码实查 OpenAiSetup.OPENAI_URL 默认值即
     * "https://api.openai.com/v1"。host 填到端点为止不带尾斜杠 + 本路径 = 传给 SDK 的 baseUrl。
     */
    private static final String PATH_OPENAI_CHAT = "/v1";
    private static final String PATH_OLLAMA_CHAT = "/api";

    private static final Pattern ENV_PLACEHOLDER = Pattern.compile("^\\$\\{(\\w+)}$");

    public static final String PROTOCOL_OLLAMA = "ollama";

    @Autowired
    private AiModelMapper modelMapper;

    @Autowired
    private AiApiKeyMapper keyMapper;

    @Autowired
    private AiProviderMapper providerMapper;

    /**
     * 解析结果：可直接发起调用的完整信息
     */
    public static class ResolvedAiModel
    {
        private AiModel model;
        private AiProvider provider;
        private AiApiKey key;
        /** 解析环境变量后的真实密钥 */
        private String resolvedApiKey;
        /** 完整请求地址（host + path） */
        private String baseUrl;
        /** 协议类型 */
        private String protocol;

        public AiModel getModel() { return model; }
        public void setModel(AiModel model) { this.model = model; }
        public AiProvider getProvider() { return provider; }
        public void setProvider(AiProvider provider) { this.provider = provider; }
        public AiApiKey getKey() { return key; }
        public void setKey(AiApiKey key) { this.key = key; }
        public String getResolvedApiKey() { return resolvedApiKey; }
        public void setResolvedApiKey(String resolvedApiKey) { this.resolvedApiKey = resolvedApiKey; }
        public String getBaseUrl() { return baseUrl; }
        public void setBaseUrl(String baseUrl) { this.baseUrl = baseUrl; }
        public String getProtocol() { return protocol; }
        public void setProtocol(String protocol) { this.protocol = protocol; }
    }

    /**
     * 解析模型（设计文档解析链第 0-6 步）。任何一步不过都抛 ServiceException（业务异常，
     * GlobalExceptionHandler 渲染为 code 500 + 明确 msg），不留静默回退。
     */
    public ResolvedAiModel resolve(Long modelId)
    {
        // 1. 三级取数
        AiModel model = modelMapper.selectById(modelId);
        if (model == null || "2".equals(model.getDelFlag()))
        {
            throw new ServiceException("模型不存在或已删除（modelId=" + modelId + "）");
        }
        AiApiKey key = keyMapper.selectById(model.getKeyId());
        if (key == null || "2".equals(key.getDelFlag()))
        {
            throw new ServiceException("模型绑定的密钥不存在或已删除（keyId=" + model.getKeyId() + "）");
        }
        AiProvider provider = providerMapper.selectById(model.getProviderId());
        if (provider == null || "2".equals(provider.getDelFlag()))
        {
            throw new ServiceException("厂商不存在或已删除（providerId=" + model.getProviderId() + "）");
        }

        // 0. 可用性校验：三级状态 + 过期时间，报错带明确定位
        if (!"0".equals(model.getStatus()))
        {
            throw new ServiceException("模型已停用：" + model.getModelName());
        }
        if (!"0".equals(key.getStatus()))
        {
            throw new ServiceException("密钥已停用：" + key.getKeyName());
        }
        if (!"0".equals(provider.getStatus()))
        {
            throw new ServiceException("厂商已停用：" + provider.getProviderName());
        }
        if (key.getExpireTime() != null && key.getExpireTime().before(new Date()))
        {
            throw new ServiceException("密钥已过期：" + key.getKeyName() + "，过期时间 " +
                    com.ruoyi.common.utils.DateUtils.parseDateToStr("yyyy-MM-dd HH:mm:ss", key.getExpireTime()));
        }

        // 4. 一致性断言：数据被手工改库的兜底
        if (!key.getProviderId().equals(model.getProviderId()))
        {
            throw new ServiceException("数据不一致：密钥所属厂商与模型厂商不同（key.providerId="
                    + key.getProviderId() + ", model.providerId=" + model.getProviderId() + "）");
        }

        // 5. 地址拼接：baseUrl = (key.base_url ?? provider.api_host) + (key.api_path ?? 协议默认版本段)
        // 产出的是"SDK baseUrl"（含版本段、不含端点尾段），openai 协议下 SDK 追加 /chat/completions
        String host = StringUtils.isNotEmpty(key.getBaseUrl()) ? key.getBaseUrl() : provider.getApiHost();
        if (StringUtils.isEmpty(host))
        {
            throw new ServiceException("配置错误：请求地址为空（key.base_url 与 provider.api_host 均未配置）");
        }
        String path = StringUtils.isNotEmpty(key.getApiPath()) ? key.getApiPath() : defaultPath(provider.getProtocol());
        if (StringUtils.isEmpty(path))
        {
            throw new ServiceException("配置错误：协议 " + provider.getProtocol()
                    + " 无默认端点路径，请在密钥上配置二级地址 api_path");
        }
        String baseUrl = stripTrailingSlash(host) + path;

        // 6. 环境变量密钥解析（集中一处，拒绝静默回退）
        String resolvedApiKey = resolveApiKey(key.getApiKey());

        ResolvedAiModel result = new ResolvedAiModel();
        result.setModel(model);
        result.setProvider(provider);
        result.setKey(key);
        result.setResolvedApiKey(resolvedApiKey);
        result.setBaseUrl(baseUrl);
        result.setProtocol(provider.getProtocol());
        return result;
    }

    /** 协议默认端点路径；custom 无默认须显式配置（返回 null 由调用方报配置错误） */
    static String defaultPath(String protocol)
    {
        if (PROTOCOL_OLLAMA.equals(protocol))
        {
            return PATH_OLLAMA_CHAT;
        }
        // openai / anthropic 等 HTTP JSON 协议本期按 openai 兼容端点处理；未知协议（dify/coze/gemini/custom）须显式配 api_path
        if ("openai".equals(protocol))
        {
            return PATH_OPENAI_CHAT;
        }
        return null;
    }

    static String stripTrailingSlash(String host)
    {
        String h = host.trim();
        while (h.endsWith("/"))
        {
            h = h.substring(0, h.length() - 1);
        }
        return h;
    }

    /** ${ENV_VAR} 占位解析：非占位原样返回；占位但环境变量未配置 → 配置异常（报错带变量名） */
    String resolveApiKey(String apiKey)
    {
        if (StringUtils.isEmpty(apiKey))
        {
            throw new ServiceException("配置错误：密钥内容为空");
        }
        Matcher m = ENV_PLACEHOLDER.matcher(apiKey.trim());
        if (!m.matches())
        {
            return apiKey;
        }
        String envName = m.group(1);
        String value = System.getenv(envName);
        if (StringUtils.isEmpty(value))
        {
            throw new ServiceException("配置错误：密钥引用的环境变量 " + envName + " 未配置");
        }
        return value;
    }
}

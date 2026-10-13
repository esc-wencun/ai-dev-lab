package com.ruoyi.ai.resolver;

import com.ruoyi.ai.common.AiConstants;
import java.util.Date;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.Setter;

import org.springframework.stereotype.Component;
import com.ruoyi.ai.domain.AiApiKey;
import com.ruoyi.ai.domain.AiModel;
import com.ruoyi.ai.domain.AiProvider;
import com.ruoyi.ai.mapper.AiApiKeyMapper;
import com.ruoyi.ai.mapper.AiModelMapper;
import com.ruoyi.ai.mapper.AiProviderMapper;
import com.ruoyi.common.exception.ServiceException;
import com.ruoyi.common.utils.DateUtils;
import com.ruoyi.common.utils.StringUtils;

/**
 * 模型可用性解析器（specs/2.0.1 核心组件）。
 * 按 modelId 三级取数 → 可用性校验 → 一致性断言 → 地址拼接 → 环境变量密钥解析，
 * 产出"可发起调用"的完整信息。链路与地址拼接约定见 AI相关表设计方案.md 第一部分第三节。
 *
 * @author ruoyi
 */
@Component
@RequiredArgsConstructor
public class AiModelResolver
{
    /**
     * 协议默认路径（版本段）。语义对齐 Spring AI 2.0.1 的 openai-java SDK：baseUrl 含版本段，
     * SDK 只追加端点尾段（/chat/completions）——源码实查 OpenAiSetup.OPENAI_URL 默认值即
     * "https://api.openai.com/v1"。host 填到端点为止不带尾斜杠 + 本路径 = 传给 SDK 的 baseUrl。
     */
    private static final String PATH_OPENAI_CHAT = "/v1";

    /** Ollama 11434 自带 OpenAI 兼容端点（/v1/chat/completions），统一走 openai 客户端 */
    private static final String PATH_OLLAMA_CHAT = "/v1";

    /** package-private：同包测试直接复用该正则断言（避免测试里抄第二份） */
    static final Pattern ENV_PLACEHOLDER = Pattern.compile(AiConstants.ENV_PLACEHOLDER_PATTERN);

    public static final String PROTOCOL_OLLAMA = "ollama";

    private final AiModelMapper modelMapper;

    private final AiApiKeyMapper keyMapper;

    private final AiProviderMapper providerMapper;

    /**
     * 解析结果：可直接发起调用的完整信息
     */
    @Getter
    @Setter
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
    }

    /**
     * 解析模型（设计文档解析链第 0-6 步）。任何一步不过都抛 ServiceException（业务异常，
     * GlobalExceptionHandler 渲染为 code 500 + 明确 msg），不留静默回退。
     */
    public ResolvedAiModel resolve(Long modelId)
    {
        AiModel model = modelMapper.selectById(modelId);
        if (model == null || AiConstants.DEL_FLAG_REMOVED.equals(model.getDelFlag()))
        {
            throw new ServiceException("模型不存在或已删除（modelId=" + modelId + "）");
        }
        AiApiKey key = keyMapper.selectById(model.getKeyId());
        if (key == null || AiConstants.DEL_FLAG_REMOVED.equals(key.getDelFlag()))
        {
            throw new ServiceException("模型绑定的密钥不存在或已删除（keyId=" + model.getKeyId() + "）");
        }
        AiProvider provider = providerMapper.selectById(model.getProviderId());
        if (provider == null || AiConstants.DEL_FLAG_REMOVED.equals(provider.getDelFlag()))
        {
            throw new ServiceException("厂商不存在或已删除（providerId=" + model.getProviderId() + "）");
        }
        return resolveByParts(provider, key, model);
    }

    /**
     * 按构件解析（草稿模式，specs/2.0.3 用户需求）：模型尚未入库时（新增弹窗里点测试），
     * 由前端提交 providerId/keyId/modelCode 直接走同一条解析链——复用可用性校验、
     * 一致性断言、地址拼接与环境变量解析，测试结果与保存后的正式链路完全一致。
     *
     * @param providerId 厂商ID
     * @param keyId      密钥ID
     * @param modelCode  模型标识
     * @param modelName  模型名称（仅报错定位用，可空）
     */
    public ResolvedAiModel resolveDraft(Long providerId, Long keyId, String modelCode, String modelName)
    {
        if (providerId == null || keyId == null || StringUtils.isEmpty(modelCode))
        {
            throw new ServiceException("请先选择厂商与密钥，并填写模型标识");
        }
        AiProvider provider = providerMapper.selectById(providerId);
        if (provider == null || AiConstants.DEL_FLAG_REMOVED.equals(provider.getDelFlag()))
        {
            throw new ServiceException("厂商不存在或已删除（providerId=" + providerId + "）");
        }
        AiApiKey key = keyMapper.selectById(keyId);
        if (key == null || AiConstants.DEL_FLAG_REMOVED.equals(key.getDelFlag()))
        {
            throw new ServiceException("密钥不存在或已删除（keyId=" + keyId + "）");
        }
        // 草稿模型壳：只带 modelName 供错误定位；status 为 null → 模型停用校验天然跳过
        AiModel draft = new AiModel();
        draft.setModelName(modelName != null && !modelName.isBlank() ? modelName : modelCode);
        return resolveByParts(provider, key, draft);
    }

    /** resolve/resolveDraft 共用的后段：可用性校验 → 一致性断言 → 地址拼接 → 密钥解析 */
    private ResolvedAiModel resolveByParts(AiProvider provider, AiApiKey key, AiModel model)
    {
        // 0. 可用性校验：三级状态 + 过期时间，报错带明确定位（模型状态/停用校验仅正式链路有意义，
        // 草稿模式的 draft.modelStatus 恒为 null，天然跳过）
        if (model != null && !AiConstants.STATUS_NORMAL.equals(model.getStatus()))
        {
            throw new ServiceException("模型已停用：" + model.getModelName());
        }
        if (!AiConstants.STATUS_NORMAL.equals(key.getStatus()))
        {
            throw new ServiceException("密钥已停用：" + key.getKeyName());
        }
        if (!AiConstants.STATUS_NORMAL.equals(provider.getStatus()))
        {
            throw new ServiceException("厂商已停用：" + provider.getProviderName());
        }
        if (key.getExpireTime() != null && key.getExpireTime().before(new Date()))
        {
            throw new ServiceException("密钥已过期：" + key.getKeyName() + "，过期时间 " +
                    DateUtils.parseDateToStr("yyyy-MM-dd HH:mm:ss", key.getExpireTime()));
        }

        // 4. 一致性断言：数据被手工改库的兜底（草稿模式 model 为 null，前端提交的即用户所选，无断言意义）
        if (model != null && !key.getProviderId().equals(model.getProviderId()))
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
        result.setBaseUrl(baseUrl);
        result.setResolvedApiKey(resolvedApiKey);
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
        if (AiConstants.PROTOCOL_OPENAI.equals(protocol))
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

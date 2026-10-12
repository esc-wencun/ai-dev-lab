package com.ruoyi.ai.resolver;

import org.junit.jupiter.api.Test;
import com.ruoyi.common.exception.ServiceException;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

/**
 * AiModelResolver 纯逻辑单测（不连库）：地址拼接矩阵、环境变量解析、协议默认路径。
 *
 * @author ruoyi
 */
public class AiModelResolverTest
{
    // @RequiredArgsConstructor 三参构造（三个 mapper）；本测试只覆盖纯逻辑方法，mapper 传 null
    private final AiModelResolver resolver = new AiModelResolver(null, null, null);

    // ---------- 地址拼接矩阵 ----------

    private String join(String host, String path)
    {
        return AiModelResolver.stripTrailingSlash(host) + path;
    }

    @Test
    public void joinHostWithoutSlashAndDefaultPath()
    {
        // 产出 SDK baseUrl（含版本段），SDK 自己追加 /chat/completions
        assertEquals("https://api.deepseek.com/v1",
                join("https://api.deepseek.com", "/v1"));
    }

    @Test
    public void joinHostWithTrailingSlashIsStripped()
    {
        assertEquals("https://api.deepseek.com/v1",
                join("https://api.deepseek.com/", "/v1"));
        assertEquals("https://api.deepseek.com/v1",
                join("https://api.deepseek.com//", "/v1"));
    }

    @Test
    void joinHostWithVersionSegment()
    {
        // 百炼兼容模式：host 填到 /compatible-mode、path 走协议默认 /v1
        assertEquals("https://dashscope.aliyuncs.com/compatible-mode/v1",
                join("https://dashscope.aliyuncs.com/compatible-mode", "/v1"));
    }

    @Test
    void joinKeyBaseUrlOverridesProviderHost()
    {
        // 中转站场景：host 带路径段时二级地址补齐版本段
        assertEquals("https://relay.example.com/openai/v1",
                join("https://relay.example.com/openai", "/v1"));
    }

    @Test
    void joinLocalhost()
    {
        assertEquals("http://localhost:11434/v1", join("http://localhost:11434", "/v1"));
    }

    // ---------- 协议默认路径 ----------

    @Test
    void defaultPathOpenai()
    {
        assertEquals("/v1", AiModelResolver.defaultPath("openai"));
    }

    @Test
    void defaultPathOllama()
    {
        // Ollama 自带 OpenAI 兼容端点：默认路径 = /v1（与 openai 协议相同）
        assertEquals("/v1", AiModelResolver.defaultPath("ollama"));
    }

    @Test
    void defaultPathCustomReturnsNull()
    {
        assertNull(AiModelResolver.defaultPath("custom"));
        assertNull(AiModelResolver.defaultPath("dify"));
        assertNull(AiModelResolver.defaultPath("gemini"));
    }

    // ---------- 环境变量密钥解析 ----------

    @Test
    void resolvePlainKeyPassthrough()
    {
        assertEquals("sk-plain-text-key", resolver.resolveApiKey("sk-plain-text-key"));
    }

    @Test
    void resolveEnvPlaceholderMissingVariableFails()
    {
        // 用一个几乎不可能存在的变量名；若真存在导致误报，换名重跑
        String varName = "DEFINITELY_NOT_SET_ENV_VAR_9Q1";
        ServiceException ex = assertThrows(ServiceException.class, () -> resolver.resolveApiKey("${" + varName + "}"));
        assertTrue(ex.getMessage().contains(varName));
    }

    @Test
    void resolveEmptyKeyFails()
    {
        assertThrows(ServiceException.class, () -> resolver.resolveApiKey(""));
        assertThrows(ServiceException.class, () -> resolver.resolveApiKey(null));
    }

    // ---------- 脱敏（挂在 AiApiKeyServiceImpl 的静态方法，顺带覆盖） ----------

    @Test
    void envPatternAcceptsWordCharsOnly()
    {
        assertTrue(AiModelResolver.ENV_PLACEHOLDER.matcher("${glm_api_key}").matches());
        assertTrue(AiModelResolver.ENV_PLACEHOLDER.matcher("${DASHSCOPE_API_KEY}").matches());
        assertTrue(!AiModelResolver.ENV_PLACEHOLDER.matcher("${bad name}").matches());
        assertTrue(!AiModelResolver.ENV_PLACEHOLDER.matcher("sk-inline").matches());
    }
}

package com.ruoyi.ai.common;

import java.util.List;
import com.alibaba.fastjson2.JSON;
import com.fasterxml.jackson.databind.DeserializationFeature;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.type.MapType;
import com.ruoyi.ai.domain.AiAgent;

/**
 * AI 智能体出口整形（specs/2.0.4）：suggestedQuestions 契约缺口修复。
 * 库存为 JSON 数组字符串（String 列），出口统一转为数组返回；
 * 实体字段类型不动（入口仍收字符串，由前端管理页序列化）。
 *
 * <p>整形方式：Jackson convertValue 实体→Map（走与接口序列化完全相同的注解与全局配置，
 * Date/@JsonIgnore 等行为零漂移），再覆写 suggestedQuestions 字段。
 * 转换收口本类静态方法，options/list/getInfo 三出口共用。</p>
 *
 * @author ruoyi
 */
public final class AiAgentPayload
{
    private static final ObjectMapper MAPPER = new ObjectMapper()
            .disable(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES);

    private AiAgentPayload() {}

    /** 列表出口：逐个整形 */
    public static List<java.util.Map<String, Object>> toList(List<AiAgent> agents)
    {
        if (agents == null)
        {
            return null;
        }
        return agents.stream().map(AiAgentPayload::toOne).collect(java.util.stream.Collectors.toList());
    }

    /** 单个出口：实体→Map + suggestedQuestions 转数组 */
    public static java.util.Map<String, Object> toOne(AiAgent agent)
    {
        if (agent == null)
        {
            return null;
        }
        java.util.Map<String, Object> map = MAPPER.convertValue(agent, mapType());
        map.put("suggestedQuestions", parseSuggestions(agent.getSuggestedQuestions()));
        return map;
    }

    /**
     * 建议问题字符串→数组（容错四态）：null → null；空白串 → null；
     * 合法 JSON 数组 → List；坏 JSON / 非 JSON → 原样字符串（诊断可见性优先，前端自行容错）。
     */
    public static Object parseSuggestions(String json)
    {
        if (json == null || json.trim().isEmpty())
        {
            return null;
        }
        try
        {
            Object parsed = JSON.parse(json);
            return parsed instanceof List ? parsed : json;
        }
        catch (Exception e)
        {
            return json;
        }
    }

    private static MapType mapType()
    {
        return MAPPER.getTypeFactory().constructMapType(java.util.HashMap.class, String.class, Object.class);
    }
}

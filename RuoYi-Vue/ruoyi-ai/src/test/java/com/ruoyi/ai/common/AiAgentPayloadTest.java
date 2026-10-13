package com.ruoyi.ai.common;

import java.util.Arrays;
import java.util.List;
import java.util.Map;
import org.junit.jupiter.api.Test;
import com.ruoyi.ai.domain.AiAgent;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

/**
 * AiAgentPayload 纯逻辑单测（不连库）：suggestedQuestions 出口转换容错 + 实体整形。
 *
 * @author ruoyi
 */
public class AiAgentPayloadTest
{
    // ---------- parseSuggestions 容错四态 ----------

    @Test
    public void parseValidArray()
    {
        Object out = AiAgentPayload.parseSuggestions("[\"问题一\",\"问题二\"]");
        assertTrue(out instanceof List);
        assertEquals(Arrays.asList("问题一", "问题二"), out);
    }

    @Test
    public void parseNullReturnsNull()
    {
        assertNull(AiAgentPayload.parseSuggestions(null));
    }

    @Test
    public void parseBlankReturnsNull()
    {
        assertNull(AiAgentPayload.parseSuggestions(""));
        assertNull(AiAgentPayload.parseSuggestions("   "));
    }

    @Test
    public void parseBrokenJsonReturnsRawString()
    {
        String bad = "[问题一,问题二";
        assertEquals(bad, AiAgentPayload.parseSuggestions(bad));
    }

    @Test
    public void parseNonArrayJsonReturnsRawString()
    {
        String notArray = "{\"q\":1}";
        assertEquals(notArray, AiAgentPayload.parseSuggestions(notArray));
    }

    // ---------- 实体整形 ----------

    @Test
    public void toOneConvertsSuggestionsAndKeepsOtherFields()
    {
        AiAgent agent = new AiAgent();
        agent.setAgentId(1L);
        agent.setAgentName("助手");
        agent.setSuggestedQuestions("[\"A\",\"B\"]");
        Map<String, Object> map = AiAgentPayload.toOne(agent);
        assertEquals(Arrays.asList("A", "B"), map.get("suggestedQuestions"));
        assertEquals("助手", map.get("agentName"));
        assertEquals(1, ((Number) map.get("agentId")).intValue());
    }

    @Test
    public void toOneNullSuggestionsStaysNull()
    {
        AiAgent agent = new AiAgent();
        agent.setAgentName("助手");
        Map<String, Object> map = AiAgentPayload.toOne(agent);
        assertTrue(map.containsKey("suggestedQuestions"));
        assertNull(map.get("suggestedQuestions"));
    }

    @Test
    public void toListMapsEachElement()
    {
        AiAgent a = new AiAgent();
        a.setAgentName("甲");
        a.setSuggestedQuestions("[\"x\"]");
        AiAgent b = new AiAgent();
        b.setAgentName("乙");
        List<Map<String, Object>> out = AiAgentPayload.toList(Arrays.asList(a, b));
        assertEquals(2, out.size());
        assertEquals(Arrays.asList("x"), out.get(0).get("suggestedQuestions"));
        assertNull(out.get(1).get("suggestedQuestions"));
    }

    @Test
    public void toListNullInputReturnsNull()
    {
        assertNull(AiAgentPayload.toList(null));
    }
}

package com.ruoyi.ai.planner;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.junit.jupiter.api.Assertions.assertNull;
import java.util.ArrayList;
import java.util.List;
import org.junit.jupiter.api.Test;
import com.ruoyi.ai.domain.AiConversation;
import com.ruoyi.ai.domain.AiMessage;

/**
 * AiPromptPlanner 纯逻辑单测（不连库）
 *
 * @author ruoyi
 */
public class AiPromptPlannerTest
{
    private AiMessage msg(long id, Long replyId, String type, String content, boolean useCtx)
    {
        AiMessage m = new AiMessage();
        m.setMessageId(id);
        m.setReplyId(replyId);
        m.setMsgType(type);
        m.setContent(content);
        m.setUseContext(useCtx);
        return m;
    }

    private AiConversation conv(Integer maxContexts, String summary)
    {
        AiConversation c = new AiConversation();
        c.setMaxContexts(maxContexts);
        c.setContextSummary(summary);
        c.setSystemPrompt("你是助手");
        return c;
    }

    @Test
    public void testWindowTrimKeepsLatest()
    {
        List<AiMessage> history = new ArrayList<>();
        history.add(msg(1, null, "user", "q1", true));
        history.add(msg(2, 1L, "assistant", "a1", true));
        history.add(msg(3, null, "user", "q2", true));
        history.add(msg(4, 3L, "assistant", "a2", true));
        history.add(msg(5, null, "user", "q3", true));
        history.add(msg(6, 5L, "assistant", "a3", true));
        List<AiMessage> win = AiPromptPlanner.pickContextWindow(history, 4);
        assertEquals(4, win.size());
        // 取最近 4 条：3,4,5,6
        assertEquals(3L, win.get(0).getMessageId());
        assertEquals(6L, win.get(3).getMessageId());
    }

    @Test
    public void testRegenerateDedupKeepsLatestAssistant()
    {
        List<AiMessage> history = new ArrayList<>();
        history.add(msg(1, null, "user", "q1", true));
        history.add(msg(2, 1L, "assistant", "旧回答", true));
        history.add(msg(3, 1L, "assistant", "新回答", true)); // 重生成：同 replyId
        List<AiMessage> win = AiPromptPlanner.pickContextWindow(history, 20);
        // user(1) + 新回答(3)，旧回答(2) 被去重剔除
        assertEquals(2, win.size());
        assertEquals("q1", win.get(0).getContent());
        assertEquals("新回答", win.get(1).getContent());
    }

    @Test
    public void testUseContextZeroExcluded()
    {
        List<AiMessage> history = new ArrayList<>();
        history.add(msg(1, null, "user", "被剔除", false));
        history.add(msg(2, 1L, "assistant", "也被剔除", false));
        history.add(msg(3, null, "user", "保留", true));
        List<AiMessage> win = AiPromptPlanner.pickContextWindow(history, 20);
        assertEquals(1, win.size());
        assertEquals("保留", win.get(0).getContent());
    }

    @Test
    public void testSystemAndSummaryPrepend()
    {
        AiConversation c = conv(null, "早期对话摘要内容");
        List<AiMessage> history = new ArrayList<>();
        history.add(msg(1, null, "user", "hi", true));
        var prompt = AiPromptPlanner.buildPrompt(c, history, "当前问题");
        assertEquals(4, prompt.getInstructions().size());
        // 顺序：system(快照) → system(摘要) → user(hi) → user(当前问题)
        assertTrue(prompt.getInstructions().get(0).getText().contains("你是助手"));
        assertTrue(prompt.getInstructions().get(1).getText().contains("早期对话摘要"));
        assertNull(prompt.getInstructions().get(2).getText() == null ? "" : null);
        assertEquals("当前问题", prompt.getInstructions().get(3).getText());
    }

    @Test
    public void testNoSystemNoSummary()
    {
        AiConversation c = new AiConversation();
        c.setMaxContexts(5);
        var prompt = AiPromptPlanner.buildPrompt(c, new ArrayList<>(), "only");
        assertEquals(1, prompt.getInstructions().size());
        assertEquals("only", prompt.getInstructions().get(0).getText());
    }

    @Test
    public void testDefaultMaxContexts()
    {
        List<AiMessage> history = new ArrayList<>();
        for (long i = 1; i <= 30; i++)
        {
            history.add(msg(i, null, "user", "m" + i, true));
        }
        AiConversation c = conv(null, null); // maxContexts=null → DEFAULT 20
        List<AiMessage> win = AiPromptPlanner.pickContextWindow(history, 20);
        assertEquals(20, win.size());
        assertEquals(11L, win.get(0).getMessageId());
    }
}

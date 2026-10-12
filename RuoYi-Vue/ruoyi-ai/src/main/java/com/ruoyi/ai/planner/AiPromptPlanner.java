package com.ruoyi.ai.planner;

import com.ruoyi.ai.common.AiConstants;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import org.springframework.ai.chat.messages.AssistantMessage;
import org.springframework.ai.chat.messages.Message;
import org.springframework.ai.chat.messages.SystemMessage;
import org.springframework.ai.chat.messages.UserMessage;
import org.springframework.ai.chat.prompt.Prompt;
import com.ruoyi.ai.domain.AiConversation;
import com.ruoyi.ai.domain.AiMessage;

/**
 * 对话 Prompt 装配器（纯函数，可单测）。
 *
 * 装配顺序（对位设计文档对话域运行时链路）：
 * SystemMessage(会话快照 system_prompt)
 * + [context_summary 压缩摘要拼装位——压缩任务后续模块，当前为空则跳过]
 * + 最近 max_contexts 条 useContext=1 的历史消息（同 replyId 只保留最新一条 assistant）
 * + 当前 user 消息
 *
 * @author ruoyi
 */
public class AiPromptPlanner
{
    private AiPromptPlanner() {}

    /** 默认历史条数上限（会话未配置 max_contexts 时） */
    public static final int DEFAULT_MAX_CONTEXTS = 20;

    /**
     * 装配 Prompt。
     *
     * @param conversation 会话（取 system_prompt/max_contexts/context_summary）
     * @param history 会话全部消息（升序；本方法负责过滤与裁剪）
     * @param currentUserMessage 当前用户输入
     * @return Spring AI Prompt
     */
    public static Prompt buildPrompt(AiConversation conversation, List<AiMessage> history, String currentUserMessage)
    {
        List<Message> messages = new ArrayList<>();

        if (conversation.getSystemPrompt() != null && !conversation.getSystemPrompt().isBlank())
        {
            messages.add(new SystemMessage(conversation.getSystemPrompt()));
        }
        if (conversation.getContextSummary() != null && !conversation.getContextSummary().isBlank())
        {
            // 压缩摘要置历史之前（水位线之前的轮次已摘要化）
            messages.add(new SystemMessage("以下是本会话早期对话的摘要：\n" + conversation.getContextSummary()));
        }

        int maxContexts = conversation.getMaxContexts() != null && conversation.getMaxContexts() > 0
                ? conversation.getMaxContexts() : DEFAULT_MAX_CONTEXTS;

        for (AiMessage m : pickContextWindow(history, maxContexts))
        {
            if (AiConstants.MSG_TYPE_USER.equals(m.getMsgType()))
            {
                messages.add(new UserMessage(m.getContent()));
            }
            else if (AiConstants.MSG_TYPE_ASSISTANT.equals(m.getMsgType()))
            {
                messages.add(new AssistantMessage(m.getContent() != null ? m.getContent() : ""));
            }
        }

        // currentUserMessage 为 null（重生成模式）：原问题已在历史窗口，不重复追加
        if (currentUserMessage != null && !currentUserMessage.isBlank())
        {
            messages.add(new UserMessage(currentUserMessage));
        }
        return new Prompt(messages);
    }

    /**
     * 上下文窗口选取：useContext=1 的消息中取最近 maxContexts 条；
     * 同一 replyId 的多条 assistant（重生成场景）只保留 messageId 最大（最新）的一条。
     *
     * @param history 全部消息（升序）
     * @param maxContexts 条数上限
     * @return 窗口内消息（升序）
     */
    public static List<AiMessage> pickContextWindow(List<AiMessage> history, int maxContexts)
    {
        // 1) 活跃消息
        List<AiMessage> active = new ArrayList<>();
        for (AiMessage m : history)
        {
            if (Boolean.TRUE.equals(m.getUseContext())
                    && (AiConstants.MSG_TYPE_USER.equals(m.getMsgType()) || AiConstants.MSG_TYPE_ASSISTANT.equals(m.getMsgType())))
            {
                active.add(m);
            }
        }

        // 2) 同 replyId 去重：assistant 只留最新（messageId 最大）；user 不受影响
        Map<Long, AiMessage> latestByReply = new HashMap<>();
        List<AiMessage> deduped = new ArrayList<>();
        for (AiMessage m : active)
        {
            if (AiConstants.MSG_TYPE_ASSISTANT.equals(m.getMsgType()) && m.getReplyId() != null)
            {
                AiMessage prev = latestByReply.get(m.getReplyId());
                if (prev != null)
                {
                    // 已有一条：占位换成本条（升序遍历，后来者更新），并把被替换者标记剔除
                    deduped.remove(prev);
                }
                latestByReply.put(m.getReplyId(), m);
                deduped.add(m);
            }
            else
            {
                deduped.add(m);
            }
        }

        // 3) 最近 maxContexts 条（取尾部），保持升序
        int from = Math.max(0, deduped.size() - maxContexts);
        return new ArrayList<>(deduped.subList(from, deduped.size()));
    }
}

package com.ruoyi.ai.service;

import java.util.List;
import com.ruoyi.ai.domain.AiConversation;

/**
 * AI会话 服务接口
 *
 * @author ruoyi
 */
public interface IAiConversationService
{
    /**
     * 我的会话列表（置顶在前，时间倒序）
     */
    List<AiConversation> selectMyConversations();

    /**
     * 新建会话。agentId 非空时复制智能体配置快照（system_prompt/knowledge_ids/tool_ids/
     * mcp_ids/enable_thinking + title=角色名——快照字段一个不漏）；
     * lastMessageTime 初始化为当前时间（防 NULL 沉底）。
     *
     * @param title 标题（可空：有 agent 用角色名，否则「新会话」）
     * @param modelId 模型（可空：取 agent 默认模型或最新对话模型）
     * @param agentId 智能体（可空=自由对话）
     * @return 新会话
     */
    AiConversation createConversation(String title, Long modelId, Long agentId);

    /**
     * 重命名（归属校验：仅本人会话可改）
     */
    void renameConversation(Long conversationId, String title);

    /**
     * 置顶/取消
     */
    void pinConversation(Long conversationId, boolean pinned);

    /**
     * 删除（连带消息软删；归属校验）
     */
    void deleteConversation(Long conversationId);

    /**
     * 切换会话模型（会话中途切模型）
     */
    void changeModel(Long conversationId, Long modelId);

    /**
     * 更新会话推理参数（temperature/maxTokens/maxContexts）。
     * null 字段不更新（保持现值）；越界抛业务异常。生效链路：planner 读会话参数覆盖模型默认。
     */
    void updateParams(Long conversationId, Double temperature, Integer maxTokens, Integer maxContexts);

    /**
     * 自动起名：标题仍为默认「新会话」时，取首条用户消息前 20 字符作标题（已重命名/重生成不触发）。
     * 返回是否实际改名（供调用方决定是否刷会话列表）。
     */
    boolean autoTitleIfDefault(Long conversationId, String firstUserContent);

    /**
     * 取本人会话（不存在/非本人 → 业务异常）
     */
    AiConversation getOwnedConversation(Long conversationId);

    /**
     * 会话消息计数维护（消息落库/重生成后调用）：messageCount/contextMsgCount/lastMessageTime
     */
    void refreshMessageStats(Long conversationId);
}

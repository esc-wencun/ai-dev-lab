package com.ruoyi.ai.mapper;

import java.util.Date;
import java.util.List;
import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.ruoyi.ai.common.AiConstants;
import com.ruoyi.ai.domain.AiMessage;
import com.ruoyi.common.core.mybatis.LambdaQueryWrapperX;

/**
 * AI消息 数据层
 *
 * @author ruoyi
 */
public interface AiMessageMapper extends BaseMapper<AiMessage>
{
    /**
     * 按会话拉历史（升序，未删除——已删会话的连带软删消息不再回显）
     *
     * @param conversationId 会话ID
     * @return 消息列表
     */
    default List<AiMessage> selectByConversation(Long conversationId)
    {
        return selectList(new LambdaQueryWrapperX<AiMessage>()
                .eq(AiMessage::getConversationId, conversationId)
                .orderByAsc(AiMessage::getMessageId));
    }

    /**
     * 会话内最近 N 条活跃消息（上下文窗口拼装用，倒序取再正序用；未删除）
     *
     * @param conversationId 会话ID
     * @param limit 条数上限
     * @return 消息列表（升序返回）
     */
    default List<AiMessage> selectRecentActive(Long conversationId, int limit)
    {
        List<AiMessage> desc = selectList(new LambdaQueryWrapperX<AiMessage>()
                .eq(AiMessage::getConversationId, conversationId)
                .eq(AiMessage::getUseContext, true)
                .orderByDesc(AiMessage::getMessageId)
                .last("limit " + limit));
        java.util.Collections.reverse(desc);
        return desc;
    }

    /**
     * 同一回复下的活跃 assistant 消息（重生成去重：只保留最新；未删除）
     *
     * @param replyId user 消息ID
     * @return 该回复的全部活跃 assistant 消息（升序）
     */
    default List<AiMessage> selectActiveByReply(Long replyId)
    {
        return selectList(new LambdaQueryWrapperX<AiMessage>()
                .eq(AiMessage::getReplyId, replyId)
                .eq(AiMessage::getUseContext, true)
                .eq(AiMessage::getMsgType, AiConstants.MSG_TYPE_ASSISTANT)
                .orderByAsc(AiMessage::getMessageId));
    }

    /**
     * 孤儿占位行：生成中但超时的消息（启动清理用）
     *
     * @param beforeTime 时间点
     * @return 消息列表
     */
    default List<AiMessage> selectStaleGenerating(Date beforeTime)
    {
        return selectList(new LambdaQueryWrapperX<AiMessage>()
                .eq(AiMessage::getMsgStatus, 0)
                .lt(AiMessage::getCreateTime, beforeTime));
    }
}

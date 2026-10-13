package com.ruoyi.ai.mapper;

import com.ruoyi.ai.common.AiConstants;
import java.util.List;
import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.ruoyi.ai.domain.AiConversation;
import com.ruoyi.common.core.mybatis.LambdaQueryWrapperX;

/**
 * AI会话 数据层
 *
 * @author ruoyi
 */
public interface AiConversationMapper extends BaseMapper<AiConversation>
{
    /**
     * 我的会话列表：置顶在前，组内按最后消息时间倒序
     *
     * @param userId 用户ID
     * @return 会话列表
     */
    default List<AiConversation> selectUserConversations(Long userId)
    {
        return selectList(new LambdaQueryWrapperX<AiConversation>()
                .eq(AiConversation::getUserId, userId)
                .eq(AiConversation::getStatus, AiConstants.STATUS_NORMAL)
                .orderByDesc(AiConversation::getPinned)
                .orderByDesc(AiConversation::getLastMessageTime));
    }

    /**
     * 按智能体查会话数（删除前置校验用；排除已删会话——已删会话不应永远阻塞智能体删除）
     *
     * @param agentId 智能体ID
     * @return 会话数
     */
    default long countByAgent(Long agentId)
    {
        return selectCount(new LambdaQueryWrapperX<AiConversation>()
                .eq(AiConversation::getAgentId, agentId));
    }
}

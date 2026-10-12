package com.ruoyi.ai.mapper;

import com.ruoyi.ai.common.AiConstants;
import java.util.List;
import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.ruoyi.ai.domain.AiAgent;
import com.ruoyi.common.core.mybatis.LambdaQueryWrapperX;

/**
 * AI智能体 数据层
 *
 * @author ruoyi
 */
public interface AiAgentMapper extends BaseMapper<AiAgent>
{
    /**
     * 智能体下拉：公开 + 本人私有，status=0
     *
     * @param userId 当前用户
     * @return 智能体列表
     */
    default List<AiAgent> selectVisibleAgents(Long userId)
    {
        return selectList(new LambdaQueryWrapperX<AiAgent>()
                .eq(AiAgent::getStatus, AiConstants.STATUS_NORMAL)
                .and(w -> w.eq(AiAgent::getVisibility, 1).or().eq(AiAgent::getUserId, userId))
                .orderByAsc(AiAgent::getSortOrder));
    }

    /**
     * 管理页分页查询
     */
    default List<AiAgent> selectAgentList(AiAgent agent)
    {
        return selectList(new LambdaQueryWrapperX<AiAgent>()
                .likeIfPresent(AiAgent::getAgentName, agent.getAgentName())
                .eqIfPresent(AiAgent::getCategory, agent.getCategory())
                .eqIfPresent(AiAgent::getVisibility, agent.getVisibility())
                .eqIfPresent(AiAgent::getStatus, agent.getStatus())
                .orderByAsc(AiAgent::getSortOrder));
    }
}

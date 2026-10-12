package com.ruoyi.ai.service.impl;

import com.ruoyi.ai.common.AiConstants;
import java.util.List;
import lombok.RequiredArgsConstructor;

import org.springframework.stereotype.Service;
import com.baomidou.mybatisplus.core.toolkit.Wrappers;
import com.ruoyi.ai.domain.AiAgent;
import com.ruoyi.ai.mapper.AiAgentMapper;
import com.ruoyi.ai.mapper.AiConversationMapper;
import com.ruoyi.ai.service.IAiAgentService;
import com.ruoyi.common.exception.ServiceException;
import com.ruoyi.common.utils.SecurityUtils;

/**
 * AI智能体 服务实现
 *
 * @author ruoyi
 */
@Service
@RequiredArgsConstructor
public class AiAgentServiceImpl implements IAiAgentService
{
    private final AiAgentMapper agentMapper;

    private final AiConversationMapper conversationMapper;

    @Override
    public void insertAgent(AiAgent agent)
    {
        Long dup = agentMapper.selectCount(Wrappers.lambdaQuery(AiAgent.class)
                .eq(AiAgent::getAgentName, agent.getAgentName()));
        if (dup != null && dup > 0)
        {
            throw new ServiceException("智能体名称已存在");
        }
        // 私有必须填归属人；公开默认记作者
        if (agent.getVisibility() == null)
        {
            agent.setVisibility(1);
        }
        if (agent.getVisibility() == 0 && agent.getUserId() == null)
        {
            agent.setUserId(SecurityUtils.getUserId());
        }
        if (agent.getUserId() == null)
        {
            agent.setUserId(SecurityUtils.getUserId());
        }
        agent.setCreateBy(SecurityUtils.getUsername());
        agentMapper.insert(agent);
    }

    @Override
    public void updateAgent(AiAgent agent)
    {
        AiAgent db = agentMapper.selectById(agent.getAgentId());
        if (db == null)
        {
            throw new ServiceException("智能体不存在");
        }
        Long dup = agentMapper.selectCount(Wrappers.lambdaQuery(AiAgent.class)
                .eq(AiAgent::getAgentName, agent.getAgentName())
                .ne(AiAgent::getAgentId, agent.getAgentId()));
        if (dup != null && dup > 0)
        {
            throw new ServiceException("智能体名称已存在");
        }
        if (agent.getVisibility() != null && agent.getVisibility() == 0 && agent.getUserId() == null)
        {
            agent.setUserId(db.getUserId() != null ? db.getUserId() : SecurityUtils.getUserId());
        }
        agent.setUpdateBy(SecurityUtils.getUsername());
        agentMapper.updateById(agent);
    }

    @Override
    public void deleteAgent(Long agentId)
    {
        long convCount = conversationMapper.countByAgent(agentId);
        if (convCount > 0)
        {
            throw new ServiceException("该智能体已被 " + convCount + " 个会话使用，禁止删除（先删会话或停用智能体）");
        }
        AiAgent del = new AiAgent();
        del.setAgentId(agentId);
        del.setDelFlag(AiConstants.DEL_FLAG_REMOVED);
        del.setUpdateBy(SecurityUtils.getUsername());
        agentMapper.updateById(del);
    }
}

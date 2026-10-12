package com.ruoyi.ai.service;

import com.ruoyi.ai.domain.AiAgent;

/**
 * AI智能体 服务接口
 *
 * @author ruoyi
 */
public interface IAiAgentService
{
    void insertAgent(AiAgent agent);

    void updateAgent(AiAgent agent);

    void deleteAgent(Long agentId);
}

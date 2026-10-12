package com.ruoyi.ai.service.impl;

import com.ruoyi.ai.common.AiConstants;
import java.util.Date;
import java.util.List;
import lombok.RequiredArgsConstructor;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import com.baomidou.mybatisplus.core.toolkit.Wrappers;
import com.ruoyi.ai.domain.AiAgent;
import com.ruoyi.ai.domain.AiConversation;
import com.ruoyi.ai.domain.AiMessage;
import com.ruoyi.ai.mapper.AiAgentMapper;
import com.ruoyi.ai.mapper.AiConversationMapper;
import com.ruoyi.ai.mapper.AiMessageMapper;
import com.ruoyi.ai.service.IAiConversationService;
import com.ruoyi.common.exception.ServiceException;
import com.ruoyi.common.utils.SecurityUtils;

/**
 * AI会话 服务实现
 *
 * @author ruoyi
 */
@Service
@RequiredArgsConstructor
public class AiConversationServiceImpl implements IAiConversationService
{
    private final AiConversationMapper conversationMapper;

    private final AiAgentMapper agentMapper;

    private final AiMessageMapper messageMapper;

    @Override
    public List<AiConversation> selectMyConversations()
    {
        return conversationMapper.selectUserConversations(SecurityUtils.getUserId());
    }

    @Override
    @Transactional
    public AiConversation createConversation(String title, Long modelId, Long agentId)
    {
        AiConversation conv = new AiConversation();
        conv.setUserId(SecurityUtils.getUserId());
        conv.setStatus(AiConstants.STATUS_NORMAL);
        conv.setPinned(false);
        conv.setMessageCount(0);
        conv.setContextMsgCount(0);
        conv.setLastMessageTime(new Date());
        conv.setEnableThinking(false);

        AiAgent agent = null;
        if (agentId != null)
        {
            agent = agentMapper.selectById(agentId);
            if (agent == null || AiConstants.STATUS_DISABLED.equals(agent.getStatus()))
            {
                throw new ServiceException("智能体不存在或已停用");
            }
            // 可见性校验：私有智能体仅作者可用
            if (agent.getVisibility() != null && agent.getVisibility() == 0
                    && !SecurityUtils.getUserId().equals(agent.getUserId()))
            {
                throw new ServiceException("无权使用该智能体");
            }
        }

        if (agent != null)
        {
            // 配置快照复制（设计文档决策 1：复制必须彻底，一个字段不漏）
            conv.setAgentId(agent.getAgentId());
            conv.setTitle(title != null && !title.isBlank() ? title : agent.getAgentName());
            conv.setSystemPrompt(agent.getSystemPrompt());
            conv.setKnowledgeIds(agent.getKnowledgeIds());
            conv.setToolIds(agent.getToolIds());
            conv.setMcpIds(agent.getMcpIds());
            conv.setEnableThinking(agent.getEnableThinking() != null && agent.getEnableThinking());
        }
        else
        {
            conv.setTitle(title != null && !title.isBlank() ? title : "新会话");
        }

        // 模型：显式指定 > 智能体默认 > 最新对话模型（sort_order 最小的 type=1 正常模型）
        Long resolvedModelId = modelId != null ? modelId : (agent != null ? agent.getModelId() : null);
        if (resolvedModelId == null)
        {
            com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper<com.ruoyi.ai.domain.AiModel> w =
                    Wrappers.lambdaQuery(com.ruoyi.ai.domain.AiModel.class)
                            .eq(com.ruoyi.ai.domain.AiModel::getModelType, 1)
                            .eq(com.ruoyi.ai.domain.AiModel::getStatus, AiConstants.STATUS_NORMAL)
                            .orderByAsc(com.ruoyi.ai.domain.AiModel::getSortOrder)
                            .last("limit 1");
            com.ruoyi.ai.domain.AiModel fallback = modelMapper.selectOne(w);
            resolvedModelId = fallback != null ? fallback.getModelId() : null;
        }
        if (resolvedModelId == null)
        {
            throw new ServiceException("请指定模型（当前无可用对话模型）");
        }
        conv.setModelId(resolvedModelId);

        conversationMapper.insert(conv);
        return conv;
    }

    private final com.ruoyi.ai.mapper.AiModelMapper modelMapper;

    @Override
    public void renameConversation(Long conversationId, String title)
    {
        AiConversation conv = getOwnedConversation(conversationId);
        conv.setTitle(title);
        conv.setUpdateBy(SecurityUtils.getUsername());
        conversationMapper.updateById(conv);
    }

    @Override
    public void pinConversation(Long conversationId, boolean pinned)
    {
        AiConversation conv = getOwnedConversation(conversationId);
        conv.setPinned(pinned);
        conv.setPinnedTime(pinned ? new Date() : null);
        conv.setUpdateBy(SecurityUtils.getUsername());
        conversationMapper.updateById(conv);
    }

    @Override
    @Transactional
    public void deleteConversation(Long conversationId)
    {
        AiConversation conv = getOwnedConversation(conversationId);
        conv.setStatus(AiConstants.STATUS_DISABLED);
        conv.setDelFlag(AiConstants.DEL_FLAG_REMOVED);
        conv.setUpdateBy(SecurityUtils.getUsername());
        conversationMapper.updateById(conv);
        // 连带软删消息
        List<AiMessage> msgs = messageMapper.selectByConversation(conversationId);
        for (AiMessage m : msgs)
        {
            m.setDelFlag(AiConstants.DEL_FLAG_REMOVED);
            m.setUpdateBy(SecurityUtils.getUsername());
            messageMapper.updateById(m);
        }
    }

    @Override
    public void changeModel(Long conversationId, Long modelId)
    {
        AiConversation conv = getOwnedConversation(conversationId);
        com.ruoyi.ai.domain.AiModel model = modelMapper.selectById(modelId);
        if (model == null || !AiConstants.STATUS_NORMAL.equals(model.getStatus()) || model.getModelType() == null || model.getModelType() != AiConstants.MODEL_TYPE_CHAT)
        {
            throw new ServiceException("模型不存在或不可用（仅对话模型可切换）");
        }
        conv.setModelId(model.getModelId());
        conv.setModelCode(model.getModelCode());
        conv.setUpdateBy(SecurityUtils.getUsername());
        conversationMapper.updateById(conv);
    }

    @Override
    public boolean autoTitleIfDefault(Long conversationId, String firstUserContent)
    {
        AiConversation conv = getOwnedConversation(conversationId);
        if (!"新会话".equals(conv.getTitle()) || firstUserContent == null || firstUserContent.isBlank())
        {
            return false;
        }
        String t = firstUserContent.trim().replaceAll("\s+", " ");
        conv.setTitle(t.length() > 20 ? t.substring(0, 20) : t);
        conv.setUpdateBy(SecurityUtils.getUsername());
        conversationMapper.updateById(conv);
        return true;
    }

    @Override
    public void updateParams(Long conversationId, Double temperature, Integer maxTokens, Integer maxContexts)
    {
        AiConversation conv = getOwnedConversation(conversationId);
        // -1 是"清除覆盖"标记，放行后再 normalize；其余值按范围校验
        if (temperature != null && temperature != -1 && (temperature < 0 || temperature > 2))
        {
            throw new ServiceException("temperature 取值范围 0~2");
        }
        if (maxTokens != null && maxTokens != -1 && (maxTokens < 1 || maxTokens > 100000))
        {
            throw new ServiceException("maxTokens 取值范围 1~100000");
        }
        if (maxContexts != null && maxContexts != -1 && (maxContexts < 0 || maxContexts > 500))
        {
            throw new ServiceException("maxContexts 取值范围 0~500");
        }
        // 选择性更新：null = 保持现值（不出现在 SET 子句）；-1 = 清除覆盖（落库 NULL，回落模型默认）。
        // 不能用 updateById：MP 默认字段策略 NOT_NULL 会跳过 null 值字段，"清空"永远落不了库——
        // 必须用 UpdateWrapper 显式 set，让 NULL 进入 SET 子句。
        conv.setUpdateBy(SecurityUtils.getUsername());
        conversationMapper.update(null, com.baomidou.mybatisplus.core.toolkit.Wrappers.<AiConversation>lambdaUpdate()
                .eq(AiConversation::getConversationId, conversationId)
                .set(temperature != null, AiConversation::getTemperature,
                        temperature == null ? null : normalize(temperature, conv.getTemperature()))
                .set(maxTokens != null, AiConversation::getMaxTokens,
                        maxTokens == null ? null : normalize(maxTokens, conv.getMaxTokens()))
                .set(maxContexts != null, AiConversation::getMaxContexts,
                        maxContexts == null ? null : normalize(maxContexts, conv.getMaxContexts()))
                .set(AiConversation::getUpdateBy, conv.getUpdateBy()));
    }

    /** 参数规范化：-1 表示"清除覆盖、回落模型默认"→ 落库 null；其余非空原样 */
    private Double normalize(Double v, Double def)
    {
        if (v == null) return def;
        if (v.doubleValue() == -1d) return null;
        return v;
    }

    private Integer normalize(Integer v, Integer def)
    {
        if (v == null) return def;
        if (v.intValue() == -1) return null;
        return v;
    }

    @Override
    public AiConversation getOwnedConversation(Long conversationId)
    {
        AiConversation conv = conversationMapper.selectById(conversationId);
        if (conv == null || AiConstants.DEL_FLAG_REMOVED.equals(conv.getDelFlag()))
        {
            throw new ServiceException("会话不存在");
        }
        if (!SecurityUtils.getUserId().equals(conv.getUserId()))
        {
            throw new ServiceException("无权操作他人会话");
        }
        return conv;
    }

    @Override
    public void refreshMessageStats(Long conversationId)
    {
        AiConversation conv = conversationMapper.selectById(conversationId);
        if (conv == null)
        {
            return;
        }
        List<AiMessage> msgs = messageMapper.selectByConversation(conversationId);
        conv.setMessageCount(msgs.size());
        int active = 0;
        for (AiMessage m : msgs)
        {
            if (Boolean.TRUE.equals(m.getUseContext()))
            {
                active++;
            }
        }
        conv.setContextMsgCount(active);
        conv.setLastMessageTime(new Date());
        conversationMapper.updateById(conv);
    }
}

package com.ruoyi.ai.domain;

import lombok.Getter;
import lombok.Setter;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import com.ruoyi.common.core.domain.BaseEntity;

import java.io.Serial;

/**
 * AI 智能体对象 ai_agent
 *
 * @author ruoyi
 */
@Getter
@Setter
@TableName("ai_agent")
public class AiAgent extends BaseEntity
{
    @Serial
    private static final long serialVersionUID = 1L;

    /** 智能体ID */
    @TableId(type = IdType.AUTO)
    private Long agentId;

    /** 智能体名称 */
    private String agentName;

    /** 头像URL */
    private String avatar;

    /** 分类（市场分组） */
    private String category;

    /** 一句话描述 */
    private String description;

    /** 系统提示词 */
    private String systemPrompt;

    /** 默认模型ID（NULL=会话时另选） */
    private Long modelId;

    /** 知识库ID数组（JSON） */
    private String knowledgeIds;

    /** 工具ID数组（JSON） */
    private String toolIds;

    /** MCP服务ID数组（JSON） */
    private String mcpIds;

    /** 默认深度思考开关（0关 1开） */
    private Boolean enableThinking;

    /** 开场白 */
    private String welcomeMessage;

    /** 建议问题数组（JSON，开场白chip） */
    private String suggestedQuestions;

    /** 可见性（0私有 1公开） */
    private Integer visibility;

    /** 归属用户（作者） */
    private Long userId;

    /** 显示顺序 */
    private Integer sortOrder;

    /** 状态（0正常 1停用） */
    private String status;

    /** 删除标志（0代表存在 2代表删除） */
    private String delFlag;


}

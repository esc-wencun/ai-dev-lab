package com.ruoyi.ai.domain;

import lombok.Getter;
import lombok.Setter;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import com.ruoyi.common.core.domain.BaseEntity;

import java.io.Serial;

/**
 * AI 会话对象 ai_conversation
 *
 * @author ruoyi
 */
@Getter
@Setter
@TableName("ai_conversation")
public class AiConversation extends BaseEntity
{
    @Serial
    private static final long serialVersionUID = 1L;

    /** 会话ID */
    @TableId(type = IdType.AUTO)
    private Long conversationId;

    /** 归属用户 */
    private Long userId;

    /** 会话标题 */
    private String title;

    /** 绑定的智能体（NULL=自由对话） */
    private Long agentId;

    /** 会话使用模型ID */
    private Long modelId;

    /** 模型标识（冗余，免联表） */
    private String modelCode;

    /** 会话级系统提示词（创建时从智能体复制） */
    private String systemPrompt;

    /** 会话级知识库ID数组（JSON快照） */
    private String knowledgeIds;

    /** 会话级工具ID数组（JSON快照） */
    private String toolIds;

    /** 会话级MCP服务ID数组（JSON快照） */
    private String mcpIds;

    /** 深度思考开关快照（0关 1开） */
    private Boolean enableThinking;

    /** 压缩水位线（NULL=未压缩） */
    private Long summaryUptoMessageId;

    /** 温度覆盖（NULL=模型默认） */
    private Double temperature;

    /** 最大输出Token覆盖 */
    private Integer maxTokens;

    /** 历史消息条数上限覆盖 */
    private Integer maxContexts;

    /** 是否置顶（0否 1是） */
    private Boolean pinned;

    /** 置顶时间 */
    private java.util.Date pinnedTime;

    /** 最后消息时间（列表排序） */
    private java.util.Date lastMessageTime;

    /** 消息总条数（冗余计数） */
    private Integer messageCount;

    /** 上下文活跃消息数（压缩触发依据） */
    private Integer contextMsgCount;

    /** 上下文压缩摘要 */
    private String contextSummary;

    /** 状态（0正常 1归档） */
    private String status;

    /** 备注（BaseEntity 的 remark 列；本表未建该列，遮蔽为非表字段） */
    @com.baomidou.mybatisplus.annotation.TableField(exist = false)
    private String remark;

    /** 删除标志（0代表存在 2代表删除） */
    private String delFlag;


}

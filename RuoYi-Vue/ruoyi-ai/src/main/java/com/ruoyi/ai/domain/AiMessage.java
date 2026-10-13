package com.ruoyi.ai.domain;

import lombok.Getter;
import lombok.Setter;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import com.ruoyi.common.core.domain.BaseEntity;

/**
 * AI 消息对象 ai_message
 *
 * @author ruoyi
 */
@Getter
@Setter
@TableName("ai_message")
public class AiMessage extends BaseEntity
{
    private static final long serialVersionUID = 1L;

    /** 消息ID */
    @TableId(type = IdType.AUTO)
    private Long messageId;

    /** 所属会话ID */
    private Long conversationId;

    /** 回复的消息ID（user/assistant配对；user消息为NULL） */
    private Long replyId;

    /** 消息类型（user/assistant/system） */
    private String msgType;

    /** 消息状态（0生成中 1完成 2出错 3已取消） */
    private Integer msgStatus;

    /** 触发用户 */
    private Long userId;

    /** 智能体ID（冗余，统计用） */
    private Long agentId;

    /** 实际生成模型ID */
    private Long modelId;

    /** 模型标识（冗余，免联表） */
    private String modelCode;

    /** 消息内容（流式聚合后完整值） */
    private String content;

    /** 是否纳入上下文拼装（0剔除但不删） */
    private Boolean useContext;

    /** 引用的知识库分段ID数组（JSON，RAG溯源） */
    private String segmentIds;

    /** 联网来源（JSON） */
    private String webSearchPages;

    /** 附件地址数组（JSON，多模态） */
    private String attachmentUrls;

    /** 输入Token用量 */
    private Integer promptTokens;

    /** 输出Token用量 */
    private Integer completionTokens;

    /** 生成耗时（毫秒） */
    private Integer durationMs;

    /** 失败原因（流式出错落库） */
    private String errorMessage;

    /** 备注（BaseEntity 的 remark 列；本表未建该列，遮蔽为非表字段） */
    @com.baomidou.mybatisplus.annotation.TableField(exist = false)
    private String remark;

    /** 删除标志（0代表存在 2代表删除） */
    private String delFlag;


}

package com.ruoyi.ai.domain;

import lombok.Getter;
import lombok.Setter;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import com.ruoyi.common.core.domain.BaseEntity;

/**
 * AI 模型对象 ai_model
 *
 * @author ruoyi
 */
@Getter
@Setter
@TableName("ai_model")
public class AiModel extends BaseEntity
{
    private static final long serialVersionUID = 1L;

    /** 模型ID */
    @TableId(type = IdType.AUTO)
    private Long modelId;

    /** 所属厂商ID */
    private Long providerId;

    /** 绑定密钥ID（须与厂商一致；同一模型可经多把Key建多条通道） */
    private Long keyId;

    /** 模型名称（展示用） */
    private String modelName;

    /** 模型标识（传给API） */
    private String modelCode;

    /** 模型类型（1对话 2图片 3语音 4视频 5向量 6重排序） */
    private Integer modelType;

    /** 支持深度思考（0否 1是） */
    private Boolean supportThinking;

    /** 支持工具调用 */
    private Boolean supportTool;

    /** 支持图片输入 */
    private Boolean supportVision;

    /** 支持音频输入 */
    private Boolean supportAudio;

    /** 支持视频输入 */
    private Boolean supportVideo;

    /** 向量维度（仅type=5时有意义） */
    private Integer modelDimension;

    /** 默认温度（NULL=供应商默认） */
    private Double temperature;

    /** 默认最大输出Token */
    private Integer maxTokens;

    /** 默认携带历史消息条数上限 */
    private Integer maxContexts;

    /** 状态（0正常 1停用） */
    private String status;

    /** 显示顺序（升序） */
    private Integer sortOrder;

    /** 删除标志（0代表存在 2代表删除） */
    private String delFlag;



}

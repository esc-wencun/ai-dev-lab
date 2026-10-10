package com.ruoyi.ai.domain;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import com.ruoyi.common.core.domain.BaseEntity;

/**
 * AI 模型对象 ai_model
 *
 * @author ruoyi
 */
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

    public Long getModelId()
    {
        return modelId;
    }

    public void setModelId(Long modelId)
    {
        this.modelId = modelId;
    }

    public Long getProviderId()
    {
        return providerId;
    }

    public void setProviderId(Long providerId)
    {
        this.providerId = providerId;
    }

    public Long getKeyId()
    {
        return keyId;
    }

    public void setKeyId(Long keyId)
    {
        this.keyId = keyId;
    }

    public String getModelName()
    {
        return modelName;
    }

    public void setModelName(String modelName)
    {
        this.modelName = modelName;
    }

    public String getModelCode()
    {
        return modelCode;
    }

    public void setModelCode(String modelCode)
    {
        this.modelCode = modelCode;
    }

    public Integer getModelType()
    {
        return modelType;
    }

    public void setModelType(Integer modelType)
    {
        this.modelType = modelType;
    }

    public Boolean getSupportThinking()
    {
        return supportThinking;
    }

    public void setSupportThinking(Boolean supportThinking)
    {
        this.supportThinking = supportThinking;
    }

    public Boolean getSupportTool()
    {
        return supportTool;
    }

    public void setSupportTool(Boolean supportTool)
    {
        this.supportTool = supportTool;
    }

    public Boolean getSupportVision()
    {
        return supportVision;
    }

    public void setSupportVision(Boolean supportVision)
    {
        this.supportVision = supportVision;
    }

    public Boolean getSupportAudio()
    {
        return supportAudio;
    }

    public void setSupportAudio(Boolean supportAudio)
    {
        this.supportAudio = supportAudio;
    }

    public Boolean getSupportVideo()
    {
        return supportVideo;
    }

    public void setSupportVideo(Boolean supportVideo)
    {
        this.supportVideo = supportVideo;
    }

    public Integer getModelDimension()
    {
        return modelDimension;
    }

    public void setModelDimension(Integer modelDimension)
    {
        this.modelDimension = modelDimension;
    }

    public Double getTemperature()
    {
        return temperature;
    }

    public void setTemperature(Double temperature)
    {
        this.temperature = temperature;
    }

    public Integer getMaxTokens()
    {
        return maxTokens;
    }

    public void setMaxTokens(Integer maxTokens)
    {
        this.maxTokens = maxTokens;
    }

    public Integer getMaxContexts()
    {
        return maxContexts;
    }

    public void setMaxContexts(Integer maxContexts)
    {
        this.maxContexts = maxContexts;
    }

    public String getStatus()
    {
        return status;
    }

    public void setStatus(String status)
    {
        this.status = status;
    }

    public Integer getSortOrder()
    {
        return sortOrder;
    }

    public void setSortOrder(Integer sortOrder)
    {
        this.sortOrder = sortOrder;
    }

    public String getDelFlag()
    {
        return delFlag;
    }

    public void setDelFlag(String delFlag)
    {
        this.delFlag = delFlag;
    }
}

package com.ruoyi.ai.domain;

import java.util.Date;
import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import com.fasterxml.jackson.annotation.JsonFormat;
import com.ruoyi.common.core.domain.BaseEntity;

/**
 * AI 密钥对象 ai_api_key
 *
 * @author ruoyi
 */
@TableName("ai_api_key")
public class AiApiKey extends BaseEntity
{
    private static final long serialVersionUID = 1L;

    /** 密钥ID */
    @TableId(type = IdType.AUTO)
    private Long keyId;

    /** 所属厂商ID */
    private Long providerId;

    /** 密钥名称 */
    private String keyName;

    /** 密钥内容（支持${ENV_VAR}占位符） */
    private String apiKey;

    /** Key专属网关地址（NULL用厂商默认） */
    private String baseUrl;

    /** 二级地址：拼接在网关地址后的请求路径（NULL用协议默认路径） */
    private String apiPath;

    /** 状态（0启用 1停用） */
    private String status;

    /** 过期时间 */
    @JsonFormat(pattern = "yyyy-MM-dd HH:mm:ss")
    private Date expireTime;

    /** 删除标志（0代表存在 2代表删除） */
    private String delFlag;

    public Long getKeyId()
    {
        return keyId;
    }

    public void setKeyId(Long keyId)
    {
        this.keyId = keyId;
    }

    public Long getProviderId()
    {
        return providerId;
    }

    public void setProviderId(Long providerId)
    {
        this.providerId = providerId;
    }

    public String getKeyName()
    {
        return keyName;
    }

    public void setKeyName(String keyName)
    {
        this.keyName = keyName;
    }

    public String getApiKey()
    {
        return apiKey;
    }

    public void setApiKey(String apiKey)
    {
        this.apiKey = apiKey;
    }

    public String getBaseUrl()
    {
        return baseUrl;
    }

    public void setBaseUrl(String baseUrl)
    {
        this.baseUrl = baseUrl;
    }

    public String getApiPath()
    {
        return apiPath;
    }

    public void setApiPath(String apiPath)
    {
        this.apiPath = apiPath;
    }

    public String getStatus()
    {
        return status;
    }

    public void setStatus(String status)
    {
        this.status = status;
    }

    public Date getExpireTime()
    {
        return expireTime;
    }

    public void setExpireTime(Date expireTime)
    {
        this.expireTime = expireTime;
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

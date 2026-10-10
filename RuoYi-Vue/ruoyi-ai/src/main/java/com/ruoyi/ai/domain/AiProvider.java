package com.ruoyi.ai.domain;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import com.ruoyi.common.core.domain.BaseEntity;

/**
 * AI 厂商对象 ai_provider
 *
 * @author ruoyi
 */
@TableName("ai_provider")
public class AiProvider extends BaseEntity
{
    private static final long serialVersionUID = 1L;

    /** 厂商ID */
    @TableId(type = IdType.AUTO)
    private Long providerId;

    /** 厂商名称 */
    private String providerName;

    /** 厂商编码（唯一） */
    private String providerCode;

    /** 协议类型（openai/anthropic/ollama/gemini/dify/coze/custom） */
    private String protocol;

    /** 默认API地址 */
    private String apiHost;

    /** 厂商图标 */
    private String icon;

    /** 官网地址 */
    private String website;

    /** 密钥申请地址 */
    private String keyApplyUrl;

    /** 状态（0正常 1停用） */
    private String status;

    /** 显示顺序 */
    private Integer sortOrder;

    /** 删除标志（0代表存在 2代表删除） */
    private String delFlag;

    public Long getProviderId()
    {
        return providerId;
    }

    public void setProviderId(Long providerId)
    {
        this.providerId = providerId;
    }

    public String getProviderName()
    {
        return providerName;
    }

    public void setProviderName(String providerName)
    {
        this.providerName = providerName;
    }

    public String getProviderCode()
    {
        return providerCode;
    }

    public void setProviderCode(String providerCode)
    {
        this.providerCode = providerCode;
    }

    public String getProtocol()
    {
        return protocol;
    }

    public void setProtocol(String protocol)
    {
        this.protocol = protocol;
    }

    public String getApiHost()
    {
        return apiHost;
    }

    public void setApiHost(String apiHost)
    {
        this.apiHost = apiHost;
    }

    public String getIcon()
    {
        return icon;
    }

    public void setIcon(String icon)
    {
        this.icon = icon;
    }

    public String getWebsite()
    {
        return website;
    }

    public void setWebsite(String website)
    {
        this.website = website;
    }

    public String getKeyApplyUrl()
    {
        return keyApplyUrl;
    }

    public void setKeyApplyUrl(String keyApplyUrl)
    {
        this.keyApplyUrl = keyApplyUrl;
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

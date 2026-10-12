package com.ruoyi.ai.domain;

import lombok.Getter;
import lombok.Setter;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import com.ruoyi.common.core.domain.BaseEntity;

import java.io.Serial;

/**
 * AI 厂商对象 ai_provider
 *
 * @author ruoyi
 */
@Getter
@Setter
@TableName("ai_provider")
public class AiProvider extends BaseEntity
{
    @Serial
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



}

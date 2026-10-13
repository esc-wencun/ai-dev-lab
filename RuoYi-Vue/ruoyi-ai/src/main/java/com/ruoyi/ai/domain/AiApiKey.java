package com.ruoyi.ai.domain;

import java.io.Serial;
import java.util.Date;
import lombok.Getter;
import lombok.Setter;

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
@Getter
@Setter
@TableName("ai_api_key")
public class AiApiKey extends BaseEntity
{
    @Serial
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



}

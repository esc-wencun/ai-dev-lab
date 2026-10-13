package com.ruoyi.ai.controller.dto;

import lombok.Getter;
import lombok.Setter;

/**
 * 模型连通性测试（草稿模式）请求体（POST /ai/model/testDraft）
 *
 * @author ruoyi
 */
@Getter
@Setter
public class AiModelTestDraftBody
{
    /** 所属厂商ID（必填） */
    private Long providerId;

    /** 绑定密钥ID（必填） */
    private Long keyId;

    /** 模型标识（传给供应商 API 的真实名字，必填） */
    private String modelCode;

    /** 模型名称（仅报错定位用，可空） */
    private String modelName;
}

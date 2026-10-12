package com.ruoyi.ai.controller.dto;

import lombok.Getter;
import lombok.Setter;

/**
 * 冒烟对话请求体（POST /ai/chat，specs/2.0.0 临时端点）
 *
 * @author ruoyi
 */
@Getter
@Setter
public class AiChatBody
{
    /** 模型 ID（DB ai_model 主键）；null = 走 yml 全局配置兜底 */
    private Long modelId;

    /** 用户输入 */
    private String prompt;
}

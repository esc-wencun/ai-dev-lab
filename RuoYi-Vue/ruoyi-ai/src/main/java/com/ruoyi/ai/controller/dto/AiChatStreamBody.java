package com.ruoyi.ai.controller.dto;

import lombok.Getter;
import lombok.Setter;

/**
 * SSE 流式对话请求体（POST /ai/chat/stream）
 *
 * @author ruoyi
 */
@Getter
@Setter
public class AiChatStreamBody
{
    /** 会话ID */
    private Long conversationId;

    /** 用户输入（regenerate 时可空——服务端取被重生成消息的原问题） */
    private String content;

    /** 重生成目标消息ID（assistant）：非空时走重生成语义——旧回答 useContext 置 0，新回答挂同 replyId */
    private Long regenerateMessageId;
}

package com.ruoyi.ai.controller;

import lombok.RequiredArgsConstructor;

import org.springframework.http.MediaType;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.http.codec.ServerSentEvent;
import com.ruoyi.ai.controller.dto.AiChatStreamBody;
import com.ruoyi.ai.service.IAiChatStreamService;
import com.ruoyi.common.core.controller.BaseController;
import com.ruoyi.common.core.domain.AjaxResult;
import reactor.core.publisher.Flux;

/**
 * AI 对话接口：SSE 流式端点（功能门禁 ai:chat:chat，同会话接口）
 *
 * 注意：前端经 fetch + ReadableStream 消费（带 Authorization 头），不用原生 EventSource
 * （其不支持自定义 header）——token 仍走标准 Authorization 头，无需 query 传参的降级方案。
 *
 * @author ruoyi
 */
@RestController
@RequestMapping("/ai/chat")
@RequiredArgsConstructor
public class AiChatController extends BaseController
{
    private final IAiChatStreamService chatStreamService;

    /**
     * 流式对话：body {"conversationId": 1, "content": "..."}
     * 事件信封：{"event":"begin|token|done","payload":{...}}
     */
    @PreAuthorize("@ss.hasPermi('ai:chat:chat')")
    @PostMapping(value = "/stream", produces = MediaType.TEXT_EVENT_STREAM_VALUE)
    public Flux<ServerSentEvent<String>> stream(@RequestBody AiChatStreamBody body)
    {
        if (body.getRegenerateMessageId() != null)
        {
            return chatStreamService.regenerate(body.getRegenerateMessageId());
        }
        return chatStreamService.stream(body.getConversationId(), body.getContent());
    }

    /** 停止生成 */
    @PreAuthorize("@ss.hasPermi('ai:chat:chat')")
    @DeleteMapping("/stream/{messageId}")
    public AjaxResult stop(@PathVariable Long messageId)
    {
        boolean ok = chatStreamService.stop(messageId);
        return ok ? success() : error("该消息不在生成中");
    }
}

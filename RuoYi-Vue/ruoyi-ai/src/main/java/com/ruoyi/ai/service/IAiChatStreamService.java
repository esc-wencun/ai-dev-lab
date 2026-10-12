package com.ruoyi.ai.service;

import com.ruoyi.ai.domain.AiMessage;
import org.springframework.http.codec.ServerSentEvent;
import reactor.core.publisher.Flux;

/**
 * AI 流式对话服务接口
 *
 * @author ruoyi
 */
public interface IAiChatStreamService
{
    /**
     * 发起流式对话：落 user 消息 + assistant 占位行，返回 SSE 事件流。
     * 事件序列（JSON 信封）：begin(占位消息) → token*(增量) → done(最终消息)。
     * 三路落库：正常完成 msgStatus=1 / 出错 msgStatus=2 / 取消 msgStatus=3。
     *
     * @param conversationId 会话（须归属当前用户）
     * @param content 用户输入
     * @return SSE 事件流（payload 为 JSON 字符串）
     */
    Flux<ServerSentEvent<String>> stream(Long conversationId, String content);

    /**
     * 重新生成：旧 assistant 回答 useContext 置 0（剔除出上下文），新 assistant 消息挂同一
     * replyId（planner 按 replyId 去重保证同轮只取最新），内容取该轮 user 消息原文重发。
     *
     * @param messageId 被重生成的 assistant 消息ID
     * @return SSE 事件流（事件序列与 stream 相同）
     */
    Flux<ServerSentEvent<String>> regenerate(Long messageId);

    /**
     * 停止指定消息的生成（dispose 订阅 + msgStatus=3 落库保留半截）。
     *
     * @param messageId assistant 占位消息ID
     * @return 是否成功停止
     */
    boolean stop(Long messageId);
}

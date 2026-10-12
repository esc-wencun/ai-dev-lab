// 真实 API 聚合层（specs/2.0.2 阶段 B）：签名与 Mock 逐字一致，组件零改动
import { streamChat, stopStream } from './stream'
export * from './conversation'
export { streamChat as sendMessage, stopStream as stopGenerate }

/** 重生成：Mock 签名 regenerate(messageId, handlers)——转调流式接口的 regenerateMessageId 模式 */
export function regenerate(messageId, handlers) {
  return streamChat({ regenerateMessageId: messageId, handlers })
}

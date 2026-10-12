// SSE 流式消费（specs/2.0.2 阶段 B）
// 方案定案（spec 4.5）：fetch + ReadableStream 手工解析 SSE——标准 Authorization 头，
// 不用原生 EventSource（不支持自定义 header）。
//
// 后端事件信封：data 行是 JSON {"event":"begin|token|done","payload":{...}}
//   begin  → payload = assistant 占位消息（msgStatus=0）
//   token  → payload = { delta }
//   done   → payload = 最终消息（msgStatus=1/2/3，含 durationMs/promptTokens/completionTokens）
// 出错（后端 onErrorResume 落库后）也发 done（msgStatus=2）；取消由 stopGenerate 触发。
// 中止：AbortController 断流 + DELETE /ai/chat/stream/{messageId}（服务端 cancel → doOnCancel 落库 msgStatus=3）。

import { getToken } from '@/utils/auth'

const BASE = import.meta.env.VITE_APP_BASE_API || ''

// SSE 文本缓冲解析：按空行分帧，取 data: 行拼 JSON（跨 chunk 的帧安全）
function createSseParser(onEvent) {
  let buf = ''
  return chunk => {
    buf += chunk
    let idx
    while ((idx = buf.indexOf('\n\n')) >= 0) {
      const frame = buf.slice(0, idx)
      buf = buf.slice(idx + 2)
      const data = frame
        .split('\n')
        .filter(l => l.startsWith('data:'))
        .map(l => l.slice(5).trimStart())
        .join('\n')
      if (!data) continue
      try {
        onEvent(JSON.parse(data))
      } catch (e) {
        console.error('SSE payload parse failed:', e)
      }
    }
  }
}

// 当前流句柄（单会话单流，与 Mock timers 语义一致）
let currentController = null
let currentMessageId = null
let activeResolve = null
let activeReject = null

/**
 * 发送消息并消费 SSE 流（sendMessage / regenerate 共用）。
 * 返回 Promise<userMessage>（与 Mock 签名一致：resolve 值是 user 消息对象）。
 * regenerate 模式：content 可空（服务端取被重生成消息的原问题），userMsg 由 begin 前无法得知，
 *   组件以已有 user 消息渲染，resolve 值仅用于契约闭环。
 * handlers: { onBegin(msg), onToken(chunk, msg), onDone(msg) }
 */
export function streamChat({ conversationId, content, regenerateMessageId, handlers }) {
  const controller = new AbortController()
  currentController = controller
  activeResolve = null
  activeReject = null

  const userMsg = {
    messageId: regenerateMessageId ? null : -Date.now(),
    conversationId,
    replyId: regenerateMessageId || null,
    msgType: 'user',
    msgStatus: 1,
    content: content || '',
    useContext: 1,
    createTime: null
  }

  const handleEvent = (ev) => {
    if (ev.event === 'begin') {
      // begin 带服务端消息 id，登记 stop 所需的 messageId
      if (ev.payload && ev.payload.messageId) currentMessageId = ev.payload.messageId
      if (handlers.onBegin) handlers.onBegin(ev.payload)
    } else if (ev.event === 'token') {
      if (handlers.onToken) handlers.onToken(ev.payload.delta || '', ev.payload)
    } else if (ev.event === 'reasoning') {
      // 深度思考增量（qwen3/R1 等）：handlers.onReasoning 可选，缺省忽略
      if (handlers.onReasoning) handlers.onReasoning(ev.payload.delta || '', ev.payload)
    } else if (ev.event === 'done') {
      if (handlers.onDone) handlers.onDone(ev.payload)
      finishStream(null)
    }
  }

  function finishStream(err) {
    if (activeResolve) {
      if (err && activeReject) activeReject(err)
      else activeResolve(userMsg)
    }
    activeResolve = null
    activeReject = null
    if (!err) {
      // 正常收尾：句柄清空（stop 之后的分支由 stopGenerate 自己清）
      currentController = null
      currentMessageId = null
    }
  }

  fetch(BASE + '/ai/chat/stream', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + getToken(),
      'accept': 'text/event-stream'
    },
    body: JSON.stringify(
      regenerateMessageId
        ? { conversationId, regenerateMessageId }
        : { conversationId, content }
    ),
    signal: controller.signal
  }).then(response => {
    const ctype = (response.headers.get('content-type') || '')
    // 错误分支：doStream 早期抛 ServiceException 时全局异常处理器返回 HTTP 200 + JSON 错误体
    // （非 SSE 流）。此前该响应被当 SSE 流读取，解析不出任何事件 → Promise 永悬挂 → 按钮卡死。
    if (!response.ok || !response.body || ctype.includes('application/json')) {
      return response.json().catch(() => ({})).then(body => {
        throw new Error(body.msg || `HTTP ${response.status}`)
      })
    }
    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    const parse = createSseParser(handleEvent)
    function pump() {
      return reader.read().then(({ done, value }) => {
        if (done) {
          // 流正常关完但没等到 done（理论不该发生）：兜底 resolve，避免组件悬挂
          finishStream(null)
          return
        }
        parse(decoder.decode(value, { stream: true }))
        return pump()
      })
    }
    return pump()
  }).catch(err => {
    if (err.name === 'AbortError') return
    finishStream(err)
  })

  return new Promise((resolve, reject) => {
    activeResolve = resolve
    activeReject = reject
  })
}

/** 停止：断流 + 通知服务端 cancel（doOnCancel 落库 msgStatus=3，保留半截） */
export function stopStream() {
  const mid = currentMessageId
  if (currentController) {
    currentController.abort()
    currentController = null
  }
  currentMessageId = null
  if (activeResolve) {
    // abort 后本地立即收尾（onDone 由服务端 cancel 分支不会再来，组件态由 abort 清）
    activeResolve = null
    activeReject = null
  }
  if (mid) {
    import('@/utils/request').then(({ default: request }) => {
      request({ url: `/ai/chat/stream/${mid}`, method: 'delete' }).catch(() => {})
    })
  }
  return Promise.resolve(true)
}

// 原生 EventSource 版 SSE 消费——与 real/stream.js（fetch + ReadableStream）的对照实现。
//
// ⚠️ 未接线：real/index.js 不引用本文件，打包不会包含。仅作方案对比，请与 stream.js 对读。
//
// ---------------------------------------------------------------------------
// 原生 EventSource 的四条硬约束（它是浏览器内置的 SSE 客户端，不是 fetch 的封装）
// ---------------------------------------------------------------------------
// 1. 只能 GET，没有请求体：content / regenerateMessageId 全得塞进 query。
//    → 长文本提问会被 URL 长度卡住（浏览器约 2KB~64KB 不等，内嵌 Tomcat 默认
//      maxHttpHeaderSize 8KB，超了直接 400）。对话场景下这是致命的。
// 2. 不能自定义请求头：'Authorization: Bearer xxx' 传不了。
//    → 后端 TokenService#getToken 当前只读 header（TokenService.java:219-227），
//      必须新增 query 降级才不 401。
// 3. 断线自动重连：流正常结束（服务端关闭连接）它也会重连，重新发一次请求。
//    → 对"一次性生成"是灾难：不主动 close 就会重复生成、重复落库。
// 4. error 事件不带状态码，也不带响应体。
//    → 后端返回 200 + JSON 错误体（{"code":500,"msg":"模型已停用"}）时，拿不到 msg，
//      只能盲报"连接失败"；而 fetch 版能读 body.msg 给用户看。
//
// 因此本文件要真跑通，后端至少得配合两处（当前都不支持，故不接线）：
//   A. TokenService#getToken 增加「query 参数取 token」降级
//   B. AiChatController#stream 增加 @GetMapping 映射（当前只有 @PostMapping）
// ---------------------------------------------------------------------------

import { getToken } from '@/utils/auth'

const BASE = import.meta.env.VITE_APP_BASE_API || ''

// 单流句柄（与 fetch 版语义一致：同一时刻只允许一条流）
let currentEs = null
let currentMessageId = null
let activeResolve = null
let activeReject = null
// 已正常收尾标记：用于区分"流结束后的重连 error"与真错误
let closed = false

/**
 * 发送消息并消费 SSE 流（原生 EventSource 版）。
 * 签名与 stream.js 的 streamChat 完全一致，可原地替换。
 * handlers: { onBegin(msg), onToken(chunk, msg), onReasoning(chunk, msg), onDone(msg) }
 */
export function streamChat({ conversationId, content, regenerateMessageId, handlers }) {
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

  // 妥协 1：所有参数走 query（GET 无 body）
  const params = new URLSearchParams()
  params.set('conversationId', String(conversationId))
  if (regenerateMessageId) {
    params.set('regenerateMessageId', String(regenerateMessageId))
  } else {
    params.set('content', content || '')
  }
  // 妥协 2：token 只能走 query（带不了 header），后端须配合 A
  params.set('token', getToken() || '')

  closed = false
  activeResolve = null
  activeReject = null

  const es = new EventSource(BASE + '/ai/chat/stream?' + params.toString())
  currentEs = es

  function finishStream(err) {
    // 妥协 3：必须主动 close，否则服务端关流后 EventSource 自动重连 → 重复生成
    if (!closed) {
      closed = true
      es.close()
    }
    if (activeResolve) {
      if (err && activeReject) activeReject(err)
      else activeResolve(userMsg)
    }
    activeResolve = null
    activeReject = null
    if (!err) {
      currentEs = null
      currentMessageId = null
    }
  }

  // 注意：后端用 ServerSentEvent.builder().data(...)，未设 .event(...)，
  // 所以协议层事件名是默认的 "message"，业务事件类型在 payload 里（{event:'begin',payload:{...}}）。
  // → 这里只能统一 onmessage 再自己分发。
  // 若后端改成 .event("token") / .event("done")，才能用 es.addEventListener('token', ...)，
  // 那才是 SSE 协议层的正确用法（见文件末尾"如果后端愿意改"一节）。
  es.onmessage = (e) => {
    let ev
    try {
      ev = JSON.parse(e.data)
    } catch (err) {
      console.error('SSE payload parse failed:', err)
      return
    }
    if (ev.event === 'begin') {
      // begin 带回服务端 messageId，登记供停止用
      if (ev.payload && ev.payload.messageId) currentMessageId = ev.payload.messageId
      if (handlers.onBegin) handlers.onBegin(ev.payload)
    } else if (ev.event === 'token') {
      if (handlers.onToken) handlers.onToken(ev.payload.delta || '', ev.payload)
    } else if (ev.event === 'reasoning') {
      if (handlers.onReasoning) handlers.onReasoning(ev.payload.delta || '', ev.payload)
    } else if (ev.event === 'done') {
      if (handlers.onDone) handlers.onDone(ev.payload)
      finishStream(null)
    }
  }

  // 妥协 4：error 事件只有 readyState，没有状态码、没有响应体。
  // readyState: 0=CONNECTING（含自动重连中） 1=OPEN 2=CLOSED
  es.onerror = () => {
    if (closed) return // 已正常收尾：忽略重连触发的 error
    if (es.readyState === EventSource.CONNECTING) {
      // 还在重连：给一次机会，等它自己重连；多数情况是服务端提前关流
      return
    }
    // 真失败（CLOSED）：拿不到后端 msg，只能给盲提示
    finishStream(new Error('SSE 连接失败（原生 EventSource 无法读取服务端错误详情）'))
  }

  return new Promise((resolve, reject) => {
    activeResolve = resolve
    activeReject = reject
  })
}

/** 停止：前端 close + 通知服务端 cancel（doOnCancel 落库 msgStatus=3） */
export function stopStream() {
  const mid = currentMessageId
  if (currentEs) {
    closed = true
    currentEs.close()
    currentEs = null
  }
  currentMessageId = null
  if (activeResolve) {
    activeResolve = null
    activeReject = null
  }
  if (mid) {
    // 与 fetch 版一致：即便前端已 close，仍显式通知服务端落库 3
    import('@/utils/request').then(({ default: request }) => {
      request({ url: `/ai/chat/stream/${mid}`, method: 'delete' }).catch(() => {})
    })
  }
  return Promise.resolve(true)
}

// ---------------------------------------------------------------------------
// 如果后端愿意改，原生版可以干净一点（但 GET 长度问题无解）
// ---------------------------------------------------------------------------
// 1) 后端设事件名：ServerSentEvent.builder().event("token").data(...)
//    → 前端就能 es.addEventListener('token', fn)，不必自己解析信封
// 2) 后端开 GET + query 取 token：
//      @GetMapping(value = "/stream", produces = MediaType.TEXT_EVENT_STREAM_VALUE)
//      public Flux<ServerSentEvent<String>> streamGet(@RequestParam ...) { ... }
//    TokenService#getToken 补：
//      if (StringUtils.isEmpty(token)) token = request.getParameter("token");
//    ⚠️ token 进 query 意味着它会进 access log / 浏览器历史 / Referer，安全性弱于 header，
//       这也是 fetch 版坚持走 header 的原因之一
// 3) 长文本仍无解：GET 无 body，content 只能进 query。真要解决只能改成两段式——
//    先 POST 创建消息拿 id，再 GET 订阅该 id 的流（多一次往返 + 后端要存待发消息）

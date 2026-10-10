// AI 对话 Mock 实现（specs/2.0.2 阶段 A）
// 内存存储，字段名逐字对齐对话域表设计；假流式定时切片；刷新即重置。

let nextConversationId = 200
let nextMessageId = 1000
const timers = new Map() // conversationId -> timer
const activeHandlers = new Map() // conversationId -> handlers

const DEMO_MODEL = { modelId: 101, modelCode: 'GLM-5.3-Flash' }

// 智能体（对齐 ai_agent 表字段；agentId=null 不入库，表示「自由对话」占位）
export const AGENTS = [
  {
    agentId: 1,
    agentName: '平台使用助手',
    avatar: null,
    category: '平台',
    description: '解答本平台的使用问题',
    systemPrompt: '你是本平台的操作助手，只回答平台功能使用相关的问题。',
    modelId: 101,
    knowledgeIds: null,
    toolIds: null,
    mcpIds: null,
    enableThinking: false,
    welcomeMessage: '你好！我是平台使用助手，关于系统功能的问题都可以问我。',
    suggestedQuestions: ['怎么重置密码？', '用户管理在哪？', '怎么新增一个角色？'],
    visibility: 1,
    status: '0'
  },
  {
    agentId: 2,
    agentName: '写作助手',
    avatar: null,
    category: '效率',
    description: '文案润色与起草',
    systemPrompt: '你是专业的中文写作助手，帮助用户润色和起草各类文案。',
    modelId: 101,
    knowledgeIds: null,
    toolIds: null,
    mcpIds: null,
    enableThinking: false,
    welcomeMessage: '把需要润色或起草的内容发给我，我会给出改进版本。',
    suggestedQuestions: ['帮我润色一段工作总结', '写一封会议通知邮件'],
    visibility: 1,
    status: '0'
  }
]

const MOCK_MODELS = [
  { modelId: 101, modelName: 'GLM-5.3-Flash', modelCode: 'GLM-5.3-Flash' },
  { modelId: 100, modelName: 'GLM-4.7', modelCode: 'glm-4.7' }
]

export function listChatModels() {
  return Promise.resolve(MOCK_MODELS)
}

export function changeConversationModel(conversationId, modelId) {
  const c = findConversation(conversationId)
  if (!c) return Promise.reject(new Error('会话不存在'))
  const mm = MOCK_MODELS.find(x => x.modelId === modelId)
  if (!mm) return Promise.reject(new Error('模型不存在'))
  c.modelId = mm.modelId
  c.modelCode = mm.modelCode
  return Promise.resolve(c)
}

export function listAgents() {
  return Promise.resolve(AGENTS.filter(a => a.status === '0'))
}

function now() {
  return new Date().toISOString().replace('T', ' ').substring(0, 19)
}

const conversations = []
const messagesByConversation = new Map()

function findConversation(id) {
  return conversations.find(c => c.conversationId === id)
}

function makeConversation(seed) {
  return {
    conversationId: seed ? 100 : nextConversationId++,
    title: seed ? '欢迎使用 AI 对话' : '新会话',
    agentId: null,
    modelId: DEMO_MODEL.modelId,
    modelCode: DEMO_MODEL.modelCode,
    pinned: false,
    pinnedTime: null,
    lastMessageTime: now(),
    messageCount: 0,
    contextMsgCount: 0,
    summaryUptoMessageId: null,
    enableThinking: false,
    systemPrompt: null,
    knowledgeIds: null,
    toolIds: null,
    mcpIds: null,
    temperature: null,
    maxTokens: null,
    maxContexts: null
  }
}

function makeMessage(conversationId, replyId, msgType, content, msgStatus) {
  return {
    messageId: nextMessageId++,
    conversationId,
    replyId,
    msgType,
    msgStatus,
    userId: 1,
    agentId: null,
    modelId: DEMO_MODEL.modelId,
    modelCode: DEMO_MODEL.modelCode,
    content,
    reasoningContent: null,
    useContext: 1,
    segmentIds: null,
    webSearchPages: null,
    attachmentUrls: null,
    promptTokens: 0,
    completionTokens: 0,
    durationMs: null,
    errorMessage: null,
    createTime: now()
  }
}

// 演示会话：一条历史问答（对齐 2.0.1 内置 model 101）
;(function seedDemo() {
  const c = makeConversation(true)
  conversations.push(c)
  const u = makeMessage(100, null, 'user', '你好，请介绍一下你自己', 1)
  const a = makeMessage(100, u.messageId, 'assistant', '我是本平台内嵌的 AI 助手（Mock 演示数据，specs/2.0.2 阶段 A）。刷新页面后演示内容会重置。', 1)
  a.completionTokens = 48
  a.durationMs = 1830
  messagesByConversation.set(100, [u, a])
  c.messageCount = 2
  c.contextMsgCount = 2
})()

function sortedConversations() {
  return [...conversations].sort((x, y) =>
    (Number(y.pinned) - Number(x.pinned)) ||
    String(y.lastMessageTime).localeCompare(String(x.lastMessageTime))
  )
}

function touch(conversationId) {
  const c = findConversation(conversationId)
  if (c) c.lastMessageTime = now()
}

let mockErrorEnabled = false
export function setMockErrorEnabled(enabled) {
  mockErrorEnabled = enabled
}

export function listConversations() {
  return Promise.resolve(sortedConversations())
}

export function createConversation({ title, modelId, agentId } = {}) {
  const c = makeConversation(false)
  if (title) c.title = title
  if (modelId) c.modelId = modelId
  if (agentId) {
    const ag = AGENTS.find(a => a.agentId === agentId)
    if (ag) {
      c.agentId = ag.agentId
      c.title = ag.agentName
    }
  }
  conversations.push(c)
  messagesByConversation.set(c.conversationId, [])
  return Promise.resolve(c)
}

export function renameConversation(conversationId, title) {
  const c = findConversation(conversationId)
  if (!c) return Promise.reject(new Error('会话不存在'))
  c.title = title
  return Promise.resolve(c)
}

export function pinConversation(conversationId, pinned) {
  const c = findConversation(conversationId)
  if (!c) return Promise.reject(new Error('会话不存在'))
  c.pinned = !!pinned
  c.pinnedTime = pinned ? now() : null
  return Promise.resolve(c)
}

export function deleteConversation(conversationId) {
  stopGenerate(conversationId)
  const idx = conversations.findIndex(c => c.conversationId === conversationId)
  if (idx >= 0) conversations.splice(idx, 1)
  messagesByConversation.delete(conversationId)
  return Promise.resolve()
}

export function listMessages(conversationId) {
  return Promise.resolve([...(messagesByConversation.get(conversationId) || [])])
}

const MOCK_REPLY = [
  '### Markdown 渲染演示',
  '',
  '**加粗**、*斜体*、`行内代码`，以及列表：',
  '',
  '- 支持**流式增量**渲染',
  '- 代码块高亮（下方示例）',
  '',
  '```js',
  '// 代码高亮演示',
  'function hello(name) {',
  "  console.log('你好, ' + name + '!')",
  '}',
  'hello("Mock")',
  '```',
  '',
  '| 表格 | 示例 |',
  '| --- | --- |',
  '| 标题 | 渲染 |',
  '| 列表 | 高亮 |',
  '',
  '> 引用块：以上由 marked + highlight.js 渲染，DOMPurify 消毒。',
  '',
  '---',
  '',
  '（原 Mock 文本）本条回答由假流式逐片推送，演示生成中动效、增量渲染与停止/出错路径。',
  '点停止按钮触发已取消态；模型选择器切换后回答内容不变（Mock 假数据，真实版逐消息记模型）。'
].join('\n')

function startGeneration(conversationId, userMsg, replyText, handlers) {
  const msgs = messagesByConversation.get(conversationId)
  const a = makeMessage(conversationId, userMsg.messageId, 'assistant', '', 0)
  msgs.push(a)
  touch(conversationId)

  if (handlers && handlers.onBegin) handlers.onBegin(a)

  let pos = 0
  let pieces = 0
  const startedAt = Date.now()
  const timer = setInterval(() => {
    if (mockErrorEnabled && pieces === 2) {
      clearInterval(timer)
      activeHandlers.delete(conversationId)
      timers.delete(conversationId)
      a.msgStatus = 2
      a.errorMessage = 'Mock 注入的错误：网关超时（演示出错路径）'
      a.durationMs = Date.now() - startedAt
      finalize(conversationId)
      if (handlers && handlers.onDone) handlers.onDone(a)
      return
    }
    if (pos >= replyText.length) {
      clearInterval(timer)
      activeHandlers.delete(conversationId)
      timers.delete(conversationId)
      a.msgStatus = 1
      a.promptTokens = Math.ceil(userMsg.content.length / 2)
      a.completionTokens = Math.ceil(replyText.length / 2)
      a.durationMs = Date.now() - startedAt
      finalize(conversationId)
      if (handlers && handlers.onDone) handlers.onDone(a)
      return
    }
    const step = 2 + Math.floor(Math.random() * 5)
    const chunk = replyText.substring(pos, pos + step)
    pos += step
    pieces++
    a.content += chunk
    if (handlers && handlers.onToken) handlers.onToken(chunk, a)
  }, 45)
  timers.set(conversationId, timer)
  activeHandlers.set(conversationId, handlers || {})
  return a
}

function finalize(conversationId) {
  const c = findConversation(conversationId)
  if (!c) return
  c.messageCount = (messagesByConversation.get(conversationId) || []).length
  c.contextMsgCount = (messagesByConversation.get(conversationId) || []).filter(m => m.useContext === 1).length
}

export function sendMessage({ conversationId, content, handlers } = {}) {
  const c = findConversation(conversationId)
  if (!c) return Promise.reject(new Error('会话不存在'))
  if (timers.has(conversationId)) {
    return Promise.reject(new Error('该会话有生成中的消息，请先停止'))
  }
  const u = makeMessage(conversationId, null, 'user', content, 1)
  ;(messagesByConversation.get(conversationId) || []).push(u)
  touch(conversationId)
  startGeneration(conversationId, u, MOCK_REPLY, handlers)
  return Promise.resolve(u)
}

export function stopGenerate(conversationId) {
  const timer = timers.get(conversationId)
  if (!timer) return Promise.resolve(false)
  clearInterval(timer)
  timers.delete(conversationId)
  const h = activeHandlers.get(conversationId)
  activeHandlers.delete(conversationId)
  const msgs = messagesByConversation.get(conversationId) || []
  const a = [...msgs].reverse().find(m => m.msgType === 'assistant' && m.msgStatus === 0)
  if (a) {
    a.msgStatus = 3
    a.durationMs = a.durationMs || 0
    finalize(conversationId)
    if (h && h.onDone) h.onDone(a)
  }
  return Promise.resolve(true)
}

export function regenerate(messageId, handlers) {
  let target = null
  let convId = null
  for (const [cid, msgs] of messagesByConversation) {
    const m = msgs.find(x => x.messageId === messageId)
    if (m) { target = m; convId = cid; break }
  }
  if (!target) return Promise.reject(new Error('消息不存在'))
  if (timers.has(convId)) return Promise.reject(new Error('该会话有生成中的消息，请先停止'))
  const msgs = messagesByConversation.get(convId)
  target.useContext = 0
  const replyId = target.replyId
  const userMsg = msgs.find(x => x.messageId === replyId)
  startGeneration(convId, userMsg, MOCK_REPLY, handlers)
  return Promise.resolve()
}

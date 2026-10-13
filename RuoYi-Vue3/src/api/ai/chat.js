// AI 对话域接口（specs/2.0.2）
// 阶段 B：USE_MOCK=false 切真实实现（REST + SSE），组件零改动。
import * as mock from './mock/mockChat'
import * as real from './real/index.js'

const USE_MOCK = false
const impl = USE_MOCK ? mock : real

/** 会话列表（pinned 置顶组在前，组内按 lastMessageTime 倒序） */
export function listConversations() {
  return impl.listConversations()
}

/** 新建会话（title/modelId 可空；返回新会话对象） */
export function createConversation({ title, modelId } = {}) {
  return impl.createConversation({ title, modelId })
}

/** 重命名会话 */
export function renameConversation(conversationId, title) {
  return impl.renameConversation(conversationId, title)
}

/** 置顶 / 取消置顶（pinnedTime 由实现维护） */
export function pinConversation(conversationId, pinned) {
  return impl.pinConversation(conversationId, pinned)
}

/** 删除会话（实现连带删消息；进行中的流一并中止） */
export function deleteConversation(conversationId) {
  return impl.deleteConversation(conversationId)
}

/** 会话消息列表（按 messageId 升序） */
export function listMessages(conversationId) {
  return impl.listMessages(conversationId)
}

/**
 * 发送消息（流式）。同步返回落库的 user 消息；
 * assistant 侧经 handlers 回调：onBegin(assistant占位,msgStatus=0) → onToken(增量文本) →
 * onDone(最终assistant消息,msgStatus=1/2/3) ；onError(错误对象) 仅在异常未落库时触发。
 */
export function sendMessage({ conversationId, content, handlers }) {
  return impl.sendMessage({ conversationId, content, handlers })
}

/** 中止该会话进行中的生成（占位行置 msgStatus=3，保留已生成部分，随后 onDone 收尾） */
export function stopGenerate(conversationId) {
  return impl.stopGenerate(conversationId)
}

/**
 * 重新生成：messageId 为旧 assistant 消息。旧回答 useContext 置 0，
 * 新 assistant 消息挂同一 replyId，回调协议同 sendMessage。
 */
export function regenerate(messageId, handlers) {
  return impl.regenerate(messageId, handlers)
}

/** 切换会话模型（会话中途切模型；逐消息记实际模型） */
export function changeConversationModel(conversationId, modelId) {
  return impl.changeConversationModel(conversationId, modelId)
}

/**
 * 更新会话推理参数（specs/2.0.3）：{ temperature, maxTokens, maxContexts }。
 * null 字段 = 不动；-1 = 清除覆盖回落模型默认；返回更新后的会话对象。
 */
export function updateConversationParams(conversationId, params) {
  return impl.updateConversationParams(conversationId, params)
}

/** 模型下拉（对话模型 type=1，按 sort_order；阶段 A Mock 静态数据，阶段 B 走 2.0.1 模型接口） */
export function listChatModels() {
  return impl.listChatModels()
}

/** 智能体列表（对话栏角色下拉；对齐 ai_agent，阶段 B 走真实接口） */
export function listAgents() {
  return impl.listAgents()
}

/** 仅 Mock：开关强制出错路径（UI 评审用） */
export function setMockErrorEnabled(enabled) {
  if (USE_MOCK && mock.setMockErrorEnabled) mock.setMockErrorEnabled(enabled)
}

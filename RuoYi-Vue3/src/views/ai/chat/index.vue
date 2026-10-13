<template>
  <div class="app-container chat-page">
    <div class="chat-layout">
      <ConversationSidebar
        :conversations="conversations"
        :active-id="activeId"
        :agents="agents"
        :agent-id="selectedAgentId"
        @select="handleSelect"
        @create="handleCreate"
        @rename="handleRename"
        @pin="handlePin"
        @delete="handleDelete"
        @agent="handleAgentChange"
      />
      <div class="chat-main">
        <div v-if="activeAgent" class="agent-header">
          <span class="agent-avatar">{{ activeAgent.agentName.substring(0, 1) }}</span>
          <div class="agent-header-info">
            <span class="agent-header-name">{{ activeAgent.agentName }}</span>
            <span class="agent-header-desc">{{ activeAgent.description }}</span>
          </div>
        </div>
        <MessageList
          :messages="messages"
          :loading="streaming"
          :suggested-questions="welcomeSuggestions"
          @regenerate="handleRegenerate"
          @ask="handleAsk"
        />
        <ChatInput
          ref="chatInputRef"
          :streaming="isCurrentStreaming"
          :models="chatModels"
          :model-id="activeModelId"
          @send="handleSend"
          @stop="handleStop"
          @model="handleModelChange"
          @params="paramsOpen = true"
        />
      </div>
      <!-- 会话推理参数（specs/2.0.3）：null=用模型默认，-1=清除覆盖 -->
      <el-dialog v-model="paramsOpen" title="会话参数" width="460px" append-to-body>
        <el-form label-width="130px">
          <el-form-item label="temperature">
            <el-slider v-model="paramsForm.temperature" :min="0" :max="2" :step="0.1" show-input style="width: 100%" />
          </el-form-item>
          <el-form-item label="maxTokens">
            <el-input-number v-model="paramsForm.maxTokens" :min="1" :max="100000" :step="256" controls-position="right" placeholder="空=模型默认" style="width: 100%" />
          </el-form-item>
          <el-form-item label="maxContexts">
            <el-input-number v-model="paramsForm.maxContexts" :min="0" :max="500" controls-position="right" placeholder="历史携带条数上限" style="width: 100%" />
          </el-form-item>
          <el-form-item label="">
            <el-button link type="primary" @click="resetParams">恢复模型默认（清除覆盖）</el-button>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="paramsOpen = false">取消</el-button>
          <el-button type="primary" :loading="paramsSaving" @click="saveParams">保存</el-button>
        </template>
      </el-dialog>
    </div>
  </div>
</template>

<script setup name="AiChat">
import { listConversations, createConversation, renameConversation, pinConversation, deleteConversation, listMessages, sendMessage, stopGenerate, regenerate, listAgents, listChatModels, changeConversationModel, updateConversationParams } from '@/api/ai/chat.js'
import ConversationSidebar from './components/ConversationSidebar.vue'
import MessageList from './components/MessageList.vue'
import ChatInput from './components/ChatInput.vue'

const { proxy } = getCurrentInstance()
const chatInputRef = ref(null)

const conversations = ref([])
const activeId = ref(null)
const messages = ref([])
// 生成态归属到会话（specs/2.0.3 修复：原全局单值 streaming 导致多会话共享状态、
// 切会话后按钮态错乱）。D14 修正：单值 ref 仍会在多会话并发流时互相覆盖（A 流跑着切到 B
// 发消息，单值被 B 覆盖，切回 A 停止按钮消失/防重入失效）——改 Set，多会话各自记账
const streamingSet = reactive(new Set())
// 本地乐观消息的负数 key 生成器：同毫秒创建多条（欢迎语+首问）会撞 :key 导致 Vue 渲染异常
let localMsgSeq = 0
function nextLocalId() {
  return -Date.now() * 10 - (localMsgSeq++)
}
const isCurrentStreaming = computed(() => streamingSet.has(activeId.value))
const agents = ref([])
const activeAgent = computed(() => {
  const conv = conversations.value.find(c => c.conversationId === activeId.value)
  if (!conv || !conv.agentId) return null
  return agents.value.find(a => a.agentId === conv.agentId) || null
})
const welcomeSuggestions = computed(() => {
  const w = messages.value.find(m => m.welcome)
  if (!w) return []
  const ag = agents.value.find(a => a.agentId === selectedAgentId.value)
  return (ag && ag.suggestedQuestions) || []
})
const selectedAgentId = ref(null)
const chatModels = ref([])
const activeModelId = computed(() => {
  const conv = conversations.value.find(c => c.conversationId === activeId.value)
  return conv ? conv.modelId : null
})

async function refreshConversations() {
  conversations.value = await listConversations()
}

// ---- 思考内容 overlay（2026-10-14）----
// 思考不落库（用户决策），此前只存当前可见 messages 数组——切会话被服务端数据覆盖、
// 后台流的增量 findIndex 不到行静默丢弃，切回即丢。改为按 会话:消息 键控的内存 overlay：
// 增量恒写 overlay，打开会话时贴回。仍是纯前端态（刷新即失），只是跨会话切换不丢。
const reasoningOverlay = new Map() // `${conversationId}:${messageId}` -> { text, thinkingMs }

function overlayKey(convId, messageId) {
  return convId + ':' + messageId
}

/** 把 overlay 中该会话的思考缓冲贴回消息列表（openConversation / 失败重拉共用） */
function applyReasoningOverlay(convId, list) {
  for (const m of list) {
    const o = reasoningOverlay.get(overlayKey(convId, m.messageId))
    if (o) {
      if (o.text) m.reasoningContent = o.text
      if (o.thinkingMs != null) m._thinkingMs = o.thinkingMs
    }
  }
  return list
}

async function openConversation(id) {
  activeId.value = id
  const list = await listMessages(id)
  messages.value = applyReasoningOverlay(id, list)
}

function handleSelect(id) {
  openConversation(id)
}

async function handleCreate() {
  const agent = agents.value.find(a => a.agentId === selectedAgentId.value)
  const c = await createConversation({ agentId: selectedAgentId.value })
  await refreshConversations()
  await openConversation(c.conversationId)
  if (agent && agent.welcomeMessage) {
    messages.value.push({
      messageId: nextLocalId(),
      conversationId: c.conversationId,
      replyId: null,
      msgType: 'assistant',
      msgStatus: 1,
      content: agent.welcomeMessage,
      welcome: true,
      createTime: ''
    })
  }
  nextTick(() => chatInputRef.value?.focusInput())
}

async function handleModelChange(modelId) {
  if (!activeId.value) return
  // 会话已有真实消息（本地欢迎语除外）时二次确认：新模型接手旧上下文，切换可逆但要知道后果
  const hasHistory = messages.value.some(m => !m.welcome)
  if (hasHistory) {
    const target = chatModels.value.find(m => m.modelId === modelId)
    try {
      await proxy.$modal.confirm(`切换后，本会话将由「${target ? target.modelName : modelId}」继续对话，已有消息会作为上下文一并带给新模型。确定切换吗？`)
    } catch (e) {
      return // 取消：受控 select 的 model-value 未变，UI 自动回显原模型
    }
  }
  await changeConversationModel(activeId.value, modelId)
  refreshConversations()
}

// ---- 会话参数（specs/2.0.3）----
const paramsOpen = ref(false)
const paramsSaving = ref(false)
const paramsForm = ref({ temperature: 0.7, maxTokens: undefined, maxContexts: undefined })

watch(paramsOpen, (open) => {
  if (!open) return
  const conv = conversations.value.find(c => c.conversationId === activeId.value)
  paramsForm.value = {
    temperature: conv && conv.temperature != null ? conv.temperature : 0.7,
    maxTokens: conv && conv.maxTokens != null ? conv.maxTokens : undefined,
    maxContexts: conv && conv.maxContexts != null ? conv.maxContexts : undefined
  }
})

/** 保存：temperature 恒提交（滑杆无"空"态）；maxTokens/maxContexts 空输入框 = -1（清除覆盖）；未改动 = null（不动） */
async function saveParams() {
  if (!activeId.value) return
  const original = conversations.value.find(c => c.conversationId === activeId.value) || {}
  const body = {}
  body.temperature = paramsForm.value.temperature
  body.maxTokens = paramsForm.value.maxTokens == null
    ? (original.maxTokens != null ? -1 : null)
    : paramsForm.value.maxTokens
  body.maxContexts = paramsForm.value.maxContexts == null
    ? (original.maxContexts != null ? -1 : null)
    : paramsForm.value.maxContexts
  paramsSaving.value = true
  try {
    await updateConversationParams(activeId.value, body)
    paramsOpen.value = false
    refreshConversations()
    proxy.$modal.msgSuccess('参数已保存')
  } finally {
    paramsSaving.value = false
  }
}

/** 恢复默认：三个字段全部 -1（清除覆盖） */
async function resetParams() {
  await updateConversationParams(activeId.value, { temperature: -1, maxTokens: -1, maxContexts: -1 })
  paramsOpen.value = false
  refreshConversations()
  proxy.$modal.msgSuccess('已恢复模型默认参数')
}

function handleAsk(q) {
  handleSend(q)
}

function handleAgentChange(v) {
  selectedAgentId.value = v
}

async function handleRename(id, title) {
  await renameConversation(id, title)
  refreshConversations()
}

async function handlePin(id, pinned) {
  await pinConversation(id, pinned)
  refreshConversations()
}

async function handleDelete(id) {
  // 生成中的会话先停流再删（D14 伴随：否则删除后流还挂着，done 回调刷新已删会话）
  if (streamingSet.has(id)) {
    await stopGenerate(id).catch(() => {})
    streamingSet.delete(id)
  }
  await deleteConversation(id)
  // overlay 是内存数据：会话删了就一并清（防 Map 滞留；消息级清理不追——刷新整体释放）
  for (const k of reasoningOverlay.keys()) {
    if (k.startsWith(id + ':')) reasoningOverlay.delete(k)
  }
  await refreshConversations()
  if (activeId.value === id) {
    activeId.value = null
    messages.value = []
  }
}

function makeHandlers(conversationId) {
  // 思考计时（Claude Code 同款"Thought for Ns"）：reasoning 首帧起表，首个正文 token 止表
  const thinkStartAt = { v: null }
  // 该流最新消息 id（begin 登记；token/reasoning 载荷不带 id 时用它定位）
  const currentMsgId = { v: null }
  return {
    onBegin: (a) => {
      currentMsgId.v = a.messageId
      messages.value.push(a)
    },
    onToken: (chunk, a) => {
      const mid = a.messageId || currentMsgId.v
      // 增量恒写 overlay（切走后仍在跑的流靠它不丢数据），可见时同步到视图
      const key = overlayKey(conversationId, mid)
      const o = reasoningOverlay.get(key) || { text: '', thinkingMs: null }
      if (chunk && thinkStartAt.v != null && o.thinkingMs == null) {
        o.thinkingMs = Date.now() - thinkStartAt.v
      }
      reasoningOverlay.set(key, o)
      const i = messages.value.findIndex(m => m.messageId === mid)
      if (i >= 0) {
        const patch = { ...a, reasoningContent: o.text }
        if (o.thinkingMs != null) patch._thinkingMs = o.thinkingMs
        messages.value[i] = patch
      }
    },
    onReasoning: (chunk, a) => {
      // 思考增量：恒写 overlay（不落库，刷新即失；跨会话切换不丢——2026-10-14）
      if (chunk && thinkStartAt.v == null) thinkStartAt.v = Date.now()
      const mid = a.messageId || currentMsgId.v
      const key = overlayKey(conversationId, mid)
      const o = reasoningOverlay.get(key) || { text: '', thinkingMs: null }
      o.text += chunk
      reasoningOverlay.set(key, o)
      const i = messages.value.findIndex(m => m.messageId === mid)
      if (i >= 0) {
        messages.value[i] = { ...a, reasoningContent: o.text, _thinkingMs: o.thinkingMs }
      }
    },
    onDone: (a) => {
      const mid = a.messageId || currentMsgId.v
      // 终态并入 overlay（思考与耗时保留，正文不进 overlay——以服务端为准）
      const key = overlayKey(conversationId, mid)
      const o = reasoningOverlay.get(key) || { text: '', thinkingMs: null }
      reasoningOverlay.set(key, o)
      const i = messages.value.findIndex(m => m.messageId === mid)
      // 保留本地思考缓冲与耗时（服务端载荷不含这两个本地态字段）
      if (i >= 0) messages.value[i] = { ...a, reasoningContent: o.text, _thinkingMs: o.thinkingMs }
      // 只清自己会话的生成态（防切会话后旧流 done 误清新会话）
      streamingSet.delete(conversationId)
      refreshConversations()
    }
  }
}

async function handleSend(content) {
  if (!activeId.value) {
    const agent = agents.value.find(a => a.agentId === selectedAgentId.value)
    const c = await createConversation({ agentId: selectedAgentId.value })
    await refreshConversations()
    activeId.value = c.conversationId
    messages.value = []
    if (agent && agent.welcomeMessage) {
      messages.value.push({
        messageId: nextLocalId(),
        conversationId: c.conversationId,
        replyId: null,
        msgType: 'assistant',
        msgStatus: 1,
        content: agent.welcomeMessage,
        welcome: true,
        createTime: ''
      })
    }
  }
  const convId = activeId.value
  if (streamingSet.has(convId)) return // 同会话防重入（real 层按会话单流，双发互覆）
  streamingSet.add(convId)
  // 自动起名（客户端判断，2026-10-13 决策）：标题仍是默认「新会话」时，取首条消息前 20 字
  // 调服务端改名接口落库，成功后刷会话列表回显（已重命名的会话不触发）
  const conv = conversations.value.find(c => c.conversationId === convId)
  if (conv && conv.title === '新会话' && content && content.trim()) {
    const t = content.trim().replace(/\s+/g, ' ')
    renameConversation(convId, t.length > 20 ? t.substring(0, 20) : t)
      .then(() => refreshConversations())
      .catch(() => {})
  }
  // 先本地插 user 气泡（服务端落库异步，UI 即时回显——Claude Code 同款体验）
  messages.value.push({
    messageId: nextLocalId(),
    conversationId: convId,
    replyId: null,
    msgType: 'user',
    msgStatus: 1,
    content,
    useContext: 1,
    createTime: ''
  })
  try {
    await sendMessage({ conversationId: convId, content, handlers: makeHandlers(convId) })
  } catch (e) {
    // request 层已弹错误提示；这里兜底复位（done 缺失/断流/HTTP 错误都走这）
  } finally {
    // 兜底复位：无论 resolve/reject 都保证按钮回到可发送态
    streamingSet.delete(convId)
    // 失败后从服务端拉一次真实历史（本地乐观插入的气泡与服务端对齐；成功路径由 begin 事件流自然衔接）
    if (messages.value.length && messages.value[messages.value.length - 1].messageId < 0) {
      listMessages(convId).then(list => { messages.value = applyReasoningOverlay(convId, list) }).catch(() => {})
    }
  }
}

function handleStop() {
  if (!activeId.value) return
  const convId = activeId.value
  stopGenerate(convId).finally(() => {
    streamingSet.delete(convId)
  })
}

async function handleRegenerate(messageId) {
  const convId = activeId.value
  if (streamingSet.has(convId)) return // 同会话防重入
  streamingSet.add(convId)
  try {
    await regenerate(messageId, makeHandlers(convId))
  } catch (e) {
    // 兜底提示见 request 层；复位在 finally
  } finally {
    streamingSet.delete(convId)
  }
}

onMounted(() => {
  listAgents().then(list => { agents.value = list })
  listChatModels().then(list => { chatModels.value = list })
  refreshConversations().then(() => {
    if (conversations.value.length > 0) openConversation(conversations.value[0].conversationId)
  })
})
</script>

<style scoped>
.chat-page { height: calc(100vh - 110px); display: flex; flex-direction: column; }
.chat-layout { flex: 1; min-height: 0; display: flex; gap: 12px; }
.chat-main { flex: 1; min-width: 0; display: flex; flex-direction: column; }
.agent-header { display: flex; align-items: center; gap: 8px; padding: 6px 12px; border: 1px solid var(--el-border-color-light); border-bottom: none; border-radius: 6px 6px 0 0; background: var(--el-fill-color-extra-light); }
.agent-avatar { width: 22px; height: 22px; border-radius: 4px; background: var(--el-color-success); color: #fff; display: flex; align-items: center; justify-content: center; font-size: 12px; flex-shrink: 0; font-family: 'JetBrains Mono', Consolas, Menlo, monospace; }
.agent-header-info { display: flex; flex-direction: column; min-width: 0; }
.agent-header-name { font-size: 12.5px; font-weight: 600; font-family: 'JetBrains Mono', Consolas, Menlo, monospace; }
.agent-header-desc { font-size: 12px; color: var(--el-text-color-secondary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
</style>

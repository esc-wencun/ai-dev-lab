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

const conversations = ref([])
const activeId = ref(null)
const messages = ref([])
// 生成态归属到会话（specs/2.0.3 修复：原全局单值 streaming 导致多会话共享状态、
// 切会话后按钮态错乱）；isCurrentStreaming 是模板消费的派生值
const streamingConversationId = ref(null)
// 本地乐观消息的负数 key 生成器：同毫秒创建多条（欢迎语+首问）会撞 :key 导致 Vue 渲染异常
let localMsgSeq = 0
function nextLocalId() {
  return -Date.now() * 10 - (localMsgSeq++)
}
const isCurrentStreaming = computed(() => streamingConversationId.value !== null && streamingConversationId.value === activeId.value)
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

async function openConversation(id) {
  activeId.value = id
  messages.value = await listMessages(id)
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
}

async function handleModelChange(modelId) {
  if (!activeId.value) return
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
  await deleteConversation(id)
  await refreshConversations()
  if (activeId.value === id) {
    activeId.value = null
    messages.value = []
  }
}

function makeHandlers(conversationId) {
  // 思考计时（Claude Code 同款"Thought for Ns"）：reasoning 首帧起表，首个正文 token 止表
  const thinkStartAt = { v: null }
  return {
    onBegin: (a) => { messages.value.push(a) },
    onToken: (chunk, a) => {
      const i = messages.value.findIndex(m => m.messageId === a.messageId)
      if (i >= 0) {
        const patch = { ...a }
        if (chunk && thinkStartAt.v != null && patch._thinkingMs == null) {
          patch._thinkingMs = Date.now() - thinkStartAt.v
        }
        messages.value[i] = patch
      }
    },
    onReasoning: (chunk, a) => {
      // 思考增量缓冲到消息的 reasoningContent（流式期间 MessageItem 展开显示）
      if (chunk && thinkStartAt.v == null) thinkStartAt.v = Date.now()
      const i = messages.value.findIndex(m => m.messageId === a.messageId)
      if (i >= 0) {
        const prev = messages.value[i].reasoningContent || ''
        messages.value[i] = { ...a, reasoningContent: prev + chunk }
      }
    },
    onDone: (a) => {
      const i = messages.value.findIndex(m => m.messageId === a.messageId)
      if (i >= 0) messages.value[i] = { ...a }
      // 只清自己会话的生成态（防切会话后旧流 done 误清新会话）
      if (streamingConversationId.value === conversationId) streamingConversationId.value = null
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
  if (streamingConversationId.value === convId) return // 同会话防重入（real 层单流句柄，双发互覆）
  streamingConversationId.value = convId
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
    if (streamingConversationId.value === convId) streamingConversationId.value = null
    // 失败后从服务端拉一次真实历史（本地乐观插入的气泡与服务端对齐；成功路径由 begin 事件流自然衔接）
    if (messages.value.length && messages.value[messages.value.length - 1].messageId < 0) {
      listMessages(convId).then(list => { messages.value = list }).catch(() => {})
    }
  }
}

function handleStop() {
  if (!activeId.value) return
  const convId = activeId.value
  stopGenerate(convId).finally(() => {
    if (streamingConversationId.value === convId) streamingConversationId.value = null
  })
}

async function handleRegenerate(messageId) {
  const convId = activeId.value
  if (streamingConversationId.value === convId) return // 同会话防重入
  streamingConversationId.value = convId
  try {
    await regenerate(messageId, makeHandlers(convId))
  } catch (e) {
    // 兜底提示见 request 层；复位在 finally
  } finally {
    if (streamingConversationId.value === convId) streamingConversationId.value = null
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

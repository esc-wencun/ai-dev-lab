<template>
  <div class="app-container chat-page">
    <el-alert title="演示数据不持久（Mock）：刷新页面后会话与消息会重置" type="info" :closable="true" show-icon class="mock-tip" />
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
          :streaming="streaming"
          :models="chatModels"
          :model-id="activeModelId"
          @send="handleSend"
          @stop="handleStop"
          @model="handleModelChange"
        />
      </div>
    </div>
  </div>
</template>

<script setup name="AiChat">
import { listConversations, createConversation, renameConversation, pinConversation, deleteConversation, listMessages, sendMessage, stopGenerate, regenerate, listAgents, listChatModels, changeConversationModel } from '@/api/ai/chat.js'
import ConversationSidebar from './components/ConversationSidebar.vue'
import MessageList from './components/MessageList.vue'
import ChatInput from './components/ChatInput.vue'

const conversations = ref([])
const activeId = ref(null)
const messages = ref([])
const streaming = ref(false)
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
      messageId: -Date.now(),
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

function makeHandlers() {
  return {
    onBegin: (a) => { messages.value.push(a) },
    onToken: (chunk, a) => {
      const i = messages.value.findIndex(m => m.messageId === a.messageId)
      if (i >= 0) messages.value[i] = { ...a }
    },
    onDone: (a) => {
      const i = messages.value.findIndex(m => m.messageId === a.messageId)
      if (i >= 0) messages.value[i] = { ...a }
      streaming.value = false
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
        messageId: -Date.now(),
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
  streaming.value = true
  try {
    await sendMessage({ conversationId: activeId.value, content, handlers: makeHandlers() })
  } catch (e) {
    streaming.value = false
  }
}

function handleStop() {
  if (activeId.value) stopGenerate(activeId.value)
}

async function handleRegenerate(messageId) {
  streaming.value = true
  try {
    await regenerate(messageId, makeHandlers())
  } catch (e) {
    streaming.value = false
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
.mock-tip { margin-bottom: 8px; }
.chat-layout { flex: 1; min-height: 0; display: flex; gap: 12px; }
.chat-main { flex: 1; min-width: 0; display: flex; flex-direction: column; }
.agent-header { display: flex; align-items: center; gap: 10px; padding: 8px 12px; border: 1px solid var(--el-border-color-light); border-bottom: none; border-radius: 6px 6px 0 0; background: var(--el-fill-color-extra-light); }
.agent-avatar { width: 30px; height: 30px; border-radius: 50%; background: var(--el-color-primary); color: #fff; display: flex; align-items: center; justify-content: center; font-size: 14px; flex-shrink: 0; }
.agent-header-info { display: flex; flex-direction: column; min-width: 0; }
.agent-header-name { font-size: 13px; font-weight: 600; }
.agent-header-desc { font-size: 12px; color: var(--el-text-color-secondary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
</style>

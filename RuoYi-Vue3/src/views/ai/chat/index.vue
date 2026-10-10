<template>
  <div class="app-container">
    <el-card shadow="never" class="chat-card">
      <template #header>
        <div class="chat-header">
          <div class="chat-title">
            <svg-icon icon-class="message" />
            <span>AI 对话</span>
          </div>
          <el-button link type="danger" @click="clearChat">清空</el-button>
        </div>
      </template>
      <div ref="scrollRef" class="chat-messages">
        <el-empty v-if="messages.length === 0" description="输入问题开始对话（内容不落库，刷新即失）" />
        <div v-for="(msg, idx) in messages" :key="idx" class="chat-row" :class="msg.role">
          <div class="chat-avatar">
            <svg-icon icon-class="people" v-if="msg.role === 'user'" />
            <svg-icon icon-class="education" v-else />
          </div>
          <div class="chat-bubble" :class="{ 'is-error': msg.isError }">
            {{ msg.content }}
          </div>
        </div>
        <div v-if="loading" class="chat-row assistant">
          <div class="chat-avatar"><svg-icon icon-class="education" /></div>
          <div class="chat-bubble">思考中…</div>
        </div>
      </div>
      <div class="chat-input-area">
        <el-select v-model="modelId" placeholder="选择模型" style="width: 180px" :disabled="loading">
          <el-option v-for="m in modelOptions" :key="m.modelId" :label="m.modelName" :value="m.modelId" />
        </el-select>
        <el-input
          v-model="input"
          type="textarea"
          :rows="2"
          resize="none"
          :disabled="loading"
          placeholder="输入问题，Enter 发送（Shift+Enter 换行）"
          @keydown.enter.exact.prevent="handleSend"
        />
        <el-button type="primary" :loading="loading" @click="handleSend">发送</el-button>
      </div>
      <div class="chat-footer-note">
        临时对话页（specs/2.0.0 冒烟）：单轮调用，无会话/历史落库；2.1.0 会话+流式接口就位后替换。
      </div>
    </el-card>
  </div>
</template>

<script setup name="AiChat">
import { sendAiChat } from '@/api/ai/chat'
import { listModel } from '@/api/ai/model'

const messages = ref([])
const input = ref('')
const loading = ref(false)
const scrollRef = ref(null)
const modelOptions = ref([])
const modelId = ref(null)

loadModels()

function loadModels() {
  listModel({ pageNum: 1, pageSize: 100, status: '0', modelType: 1 }).then(res => {
    modelOptions.value = res.rows || []
    if (!modelId.value && modelOptions.value.length > 0) {
      modelId.value = modelOptions.value[0].modelId
    }
  })
}

function clearChat() {
  messages.value = []
  input.value = ''
}

function handleSend() {
  const prompt = input.value.trim()
  if (!prompt || loading.value) return
  loading.value = true
  messages.value.push({ role: 'user', content: prompt })
  sendAiChat({ modelId: modelId.value, prompt: prompt }).then((res) => {
    messages.value.push({ role: 'assistant', content: res.data ?? res.msg })
  }).catch(() => {
    messages.value.push({ role: 'assistant', content: '（调用失败，请稍后重试）', isError: true })
  }).finally(() => {
    loading.value = false
    nextTick(() => {
      if (scrollRef.value) {
        scrollRef.value.scrollTop = scrollRef.value.scrollHeight
      }
    })
  })
}
</script>

<style scoped>
.chat-card {
  height: calc(100vh - 84px);
  display: flex;
  flex-direction: column;
}
.chat-card :deep(.el-card__body) {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.chat-messages {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 8px 12px;
}
.chat-row {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}
.chat-row.user {
  flex-direction: row-reverse;
}
.chat-bubble {
  padding: 4px 12px;
  border-radius: 8px;
  max-width: 70%;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
}
.chat-row.user .chat-bubble {
  background: var(--el-color-primary-light-9);
}
.chat-row.assistant .chat-bubble {
  background: var(--el-fill-color-light);
}
.chat-bubble.is-error {
  background: var(--el-color-danger-light-9);
  color: var(--el-color-danger);
}
.chat-avatar {
  flex-shrink: 0;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--el-fill-color);
  display: flex;
  align-items: center;
  justify-content: center;
}
.chat-input-area {
  display: flex;
  gap: 8px;
  padding-top: 8px;
}
.chat-input-area .el-button {
  align-self: flex-end;
}
.chat-footer-note {
  padding-top: 6px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>

<template>
  <div class="msg-row" :class="message.msgType">
    <div class="msg-avatar">
      <svg-icon :icon-class="message.msgType === 'user' ? 'people' : 'education'" />
    </div>
    <div class="msg-body">
      <div class="msg-bubble" :class="bubbleClass">
        <span v-if="message.msgStatus === 0" class="cursor-blink">&#9646;</span>
        <span v-if="isAssistant" class="msg-content md" v-html="renderedContent"></span>
        <span v-else class="msg-content">{{ message.content }}</span>
      </div>
      <div v-if="message.welcome && suggestedQuestions.length" class="suggest-row">
        <el-tag
          v-for="q in suggestedQuestions"
          :key="q"
          class="suggest-chip"
          effect="plain"
          @click="$emit('ask', q)"
        >{{ q }}</el-tag>
      </div>
      <div class="msg-meta">
        <span v-if="message.msgStatus === 2" class="err-line">
          <el-icon><CircleCloseFilled /></el-icon>
          {{ message.errorMessage || '生成失败' }}
        </span>
        <span v-else-if="message.msgStatus === 3" class="stop-line">已停止生成</span>
        <template v-if="message.msgStatus === 1 && message.msgType === 'assistant' && !message.welcome">
          <span v-if="message.durationMs" class="meta-item">{{ (message.durationMs / 1000).toFixed(1) }}s</span>
          <span v-if="message.completionTokens" class="meta-item">{{ message.completionTokens }} tokens</span>
          <el-tooltip content="重新生成" placement="top">
            <el-icon class="regen-btn" @click="$emit('regenerate', message.messageId)"><RefreshRight /></el-icon>
          </el-tooltip>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup name="MessageItem">
import { CircleCloseFilled, RefreshRight } from '@element-plus/icons-vue'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import hljs from 'highlight.js'
import 'highlight.js/styles/github.css'

marked.setOptions({
  highlight: (code, lang) => {
    if (lang && hljs.getLanguage(lang)) {
      try { return hljs.highlight(code, { language: lang }).value } catch (e) { /* fallthrough */ }
    }
    return hljs.highlightAuto(code).value
  }
})

const props = defineProps({
  message: { type: Object, required: true },
  streaming: { type: Boolean, default: false },
  suggestedQuestions: { type: Array, default: () => [] }
})
defineEmits(['regenerate', 'ask'])

const isAssistant = computed(() => props.message.msgType === 'assistant')

const renderedContent = computed(() => {
  if (!isAssistant.value || !props.message.content) return ''
  const html = marked.parse(props.message.content, { breaks: true, async: false })
  return DOMPurify.sanitize(html, { USE_PROFILES: { html: true } })
})

const bubbleClass = computed(() => ({
  'is-error': props.message.msgStatus === 2,
  'is-stopped': props.message.msgStatus === 3,
  'is-generating': props.message.msgStatus === 0
}))
</script>

<style scoped>
.msg-row { display: flex; gap: 8px; margin-bottom: 14px; }
.msg-row.user { flex-direction: row-reverse; }
.msg-avatar { flex-shrink: 0; width: 32px; height: 32px; border-radius: 50%; background: var(--el-fill-color); display: flex; align-items: center; justify-content: center; }
.msg-body { max-width: 75%; min-width: 0; display: flex; flex-direction: column; }
.msg-row.user .msg-body { align-items: flex-end; }
.msg-bubble { padding: 6px 12px; border-radius: 8px; line-height: 1.6; white-space: pre-wrap; word-break: break-word; font-size: 13px; }
.msg-row.user .msg-bubble { background: var(--el-color-primary-light-9); }
.msg-row.assistant .msg-bubble { background: var(--el-fill-color-light); }
.msg-bubble.is-error { background: var(--el-color-danger-light-9); color: var(--el-color-danger); }
.msg-bubble.is-stopped { background: var(--el-fill-color); color: var(--el-text-color-secondary); }
.msg-content.md :deep(code) { background: var(--el-fill-color); padding: 1px 5px; border-radius: 4px; font-size: 12px; }
.msg-content.md :deep(pre) { background: #f6f8fa; padding: 10px 12px; border-radius: 6px; overflow-x: auto; margin: 6px 0; }
.msg-content.md :deep(pre code) { background: transparent; padding: 0; font-size: 12px; line-height: 1.5; }
.msg-content.md :deep(p) { margin: 4px 0; }
.msg-content.md :deep(ul), .msg-content.md :deep(ol) { margin: 4px 0; padding-left: 20px; }
.msg-content.md :deep(table) { border-collapse: collapse; margin: 6px 0; }
.msg-content.md :deep(th), .msg-content.md :deep(td) { border: 1px solid var(--el-border-color); padding: 4px 8px; font-size: 12px; }
.msg-content.md :deep(blockquote) { border-left: 3px solid var(--el-border-color); margin: 6px 0; padding: 2px 10px; color: var(--el-text-color-secondary); }
.cursor-blink { animation: blink 0.9s infinite; margin-right: 2px; color: var(--el-color-primary); }
@keyframes blink { 0%, 100% { opacity: 1; } 50% { opacity: 0; } }
.suggest-row { display: flex; flex-wrap: wrap; gap: 6px; margin-top: 6px; }
.suggest-chip { cursor: pointer; }
.suggest-chip:hover { color: var(--el-color-primary); border-color: var(--el-color-primary); }
.msg-meta { margin-top: 2px; font-size: 12px; color: var(--el-text-color-secondary); display: flex; align-items: center; gap: 10px; min-height: 18px; }
.err-line { color: var(--el-color-danger); display: inline-flex; align-items: center; gap: 4px; }
.stop-line { color: var(--el-text-color-secondary); }
.meta-item { opacity: 0.8; }
.regen-btn { cursor: pointer; }
.regen-btn:hover { color: var(--el-color-primary); }
</style>

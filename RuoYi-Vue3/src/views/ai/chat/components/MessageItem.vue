<template>
  <div class="msg-row" :class="message.msgType">
    <!-- 角色前缀：> 用户 / ● 助手（Claude Code 风格） -->
    <div class="msg-prefix">
      <span v-if="message.msgType === 'user'" class="prefix-user">&gt;</span>
      <span v-else class="prefix-assistant">&#9679;</span>
    </div>
    <div class="msg-body">
      <div class="msg-block" :class="blockClass">
        <div v-if="message.msgType === 'user'" class="user-line">{{ message.content }}</div>
        <template v-else>
          <!-- 思考过程（Claude Code 风格）：✻ 前缀折叠行，生成中展开斜体灰字，完成后可折叠；时长（Thought for Ns） -->
          <div v-if="message.reasoningContent" class="thinking-block">
            <div class="thinking-head" @click="thinkingOpen = !thinkingOpen">
              <span class="thinking-star">✻</span>
              <span class="thinking-arrow">{{ thinkingOpen ? '▾' : '▸' }}</span>
              <span>{{ thinkingLabel }}</span>
            </div>
            <div v-show="thinkingOpen" class="thinking-body">{{ message.reasoningContent }}</div>
          </div>
          <span v-if="message.msgStatus === 0 && !message.content" class="cursor-blink">&#9608;</span>
          <div v-else class="msg-content md" v-html="renderedContent"></div>
          <span v-if="message.msgStatus === 0" class="cursor-blink">&#9608;</span>
        </template>
      </div>
      <div v-if="message.welcome && suggestedQuestions.length" class="suggest-row">
        <div
          v-for="q in suggestedQuestions"
          :key="q"
          class="suggest-chip"
          @click="$emit('ask', q)"
        >{{ q }}</div>
      </div>
      <div class="msg-meta">
        <span v-if="message.msgStatus === 2" class="err-line">✕ {{ message.errorMessage || '生成失败' }}</span>
        <span v-else-if="message.msgStatus === 3" class="stop-line">⏹ 已停止生成</span>
        <template v-if="message.msgStatus === 1 && message.msgType === 'assistant' && !message.welcome">
          <span v-if="message.durationMs" class="meta-item">{{ (message.durationMs / 1000).toFixed(1) }}s</span>
          <span v-if="message.completionTokens" class="meta-item">{{ message.completionTokens }} tok</span>
          <span class="regen-btn" title="重新生成" @click="$emit('regenerate', message.messageId)">↻</span>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup name="MessageItem">
import { marked } from 'marked'
import { markedHighlight } from 'marked-highlight'
import DOMPurify from 'dompurify'
import hljs from 'highlight.js'
import 'highlight.js/styles/github.css'

// marked v5+ 移除了 setOptions.highlight（静默失效），代码高亮必须走 markedHighlight 扩展
marked.use(markedHighlight({
  langPrefix: 'hljs language-',
  highlight(code, lang) {
    const language = lang && hljs.getLanguage(lang) ? lang : 'plaintext'
    try {
      return hljs.highlight(code, { language }).value
    } catch (e) {
      return hljs.highlightAuto(code).value
    }
  }
}))

const props = defineProps({
  message: { type: Object, required: true },
  streaming: { type: Boolean, default: false },
  suggestedQuestions: { type: Array, default: () => [] },
})
defineEmits(['regenerate', 'ask'])

// 思考折叠：生成中有思考流时默认展开，完成后默认折叠（对齐 Claude Code）
const thinkingOpen = ref(props.message.msgStatus === 0)

const isAssistant = computed(() => props.message.msgType === 'assistant')

const renderedContent = computed(() => {
  if (!isAssistant.value || !props.message.content) return ''
  const html = marked.parse(props.message.content, { breaks: true, async: false })
  return DOMPurify.sanitize(html, { USE_PROFILES: { html: true } })
})

const thinkingLabel = computed(() => {
  const streamingThink = props.message.msgStatus === 0 && !!props.message.reasoningContent
  if (streamingThink) return '思考中…'
  const ms = props.message._thinkingMs
  return ms != null ? `思考了 ${(ms / 1000).toFixed(1)}s` : '思考过程'
})

const blockClass = computed(() => ({
  'is-error': props.message.msgStatus === 2,
  'is-stopped': props.message.msgStatus === 3,
  'is-generating': props.message.msgStatus === 0
}))
</script>

<style scoped>
/* Claude Code 风格：无气泡，全左对齐，角色前缀 + 平铺正文 */
.msg-row {
  display: flex;
  gap: 10px;
  padding: 8px 4px;
  border-bottom: 1px solid var(--el-border-color-extra-light);
}
.msg-row:last-child { border-bottom: none; }
/* 用户行靠右（微信/Claude Code 用户消息同款布局），AI 行保持靠左 */
.msg-row.user { flex-direction: row-reverse; }

.msg-prefix {
  flex-shrink: 0;
  width: 18px;
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
  font-size: 13px;
  line-height: 1.7;
  text-align: center;
}
/* 用户前缀贴卡片右侧顶部 */
.msg-row.user .msg-prefix { padding-top: 10px; }
.msg-row.user .prefix-user { color: var(--el-color-primary); }
.prefix-user { color: var(--el-color-primary); font-weight: 700; }
.prefix-assistant { color: var(--el-color-success); font-size: 10px; position: relative; top: -1px; }

.msg-body { flex: 1; min-width: 0; }

/* 用户消息：浅灰圆角卡片（Claude Code 风格：与 AI 平铺正文形成明确视觉区分） */
.msg-row.user .msg-block {
  display: inline-block;
  background: var(--el-fill-color-light);
  border: 1px solid var(--el-border-color-lighter);
  padding: 10px 14px;
  border-radius: 10px;
  font-size: 13.5px;
  line-height: 1.6;
  color: var(--el-text-color-primary);
  white-space: pre-wrap;
  word-break: break-word;
  max-width: 100%;
}

/* 助手消息：无底色平铺正文 */
.msg-row.assistant .msg-block {
  line-height: 1.65;
  font-size: 13px;
  color: var(--el-text-color-primary);
}
.msg-block.is-error {
  color: var(--el-color-danger);
  background: var(--el-color-danger-light-9);
  border-radius: 3px;
  padding: 4px 10px;
  display: inline-block;
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
  font-size: 12.5px;
}
.msg-block.is-stopped { color: var(--el-text-color-secondary); }

/* Markdown 正文（助手消息内） */
.msg-content.md :deep(code) {
  background: var(--el-fill-color);
  padding: 1px 5px;
  border-radius: 3px;
  font-size: 12px;
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
}
.msg-content.md :deep(pre) {
  background: #f6f8fa;
  border: 1px solid var(--el-border-color-lighter);
  padding: 10px 12px;
  border-radius: 4px;
  overflow-x: auto;
  margin: 6px 0;
}
.msg-content.md :deep(pre code) { background: transparent; padding: 0; font-size: 12px; line-height: 1.5; }
.msg-content.md :deep(p) { margin: 4px 0; }
.msg-content.md :deep(h1), .msg-content.md :deep(h2), .msg-content.md :deep(h3),
.msg-content.md :deep(h4) { margin: 8px 0 4px; font-size: 13.5px; }
.msg-content.md :deep(ul), .msg-content.md :deep(ol) { margin: 4px 0; padding-left: 20px; }
.msg-content.md :deep(table) { border-collapse: collapse; margin: 6px 0; }
.msg-content.md :deep(th), .msg-content.md :deep(td) { border: 1px solid var(--el-border-color); padding: 4px 8px; font-size: 12px; }
.msg-content.md :deep(blockquote) {
  border-left: 3px solid var(--el-border-color);
  margin: 6px 0;
  padding: 2px 10px;
  color: var(--el-text-color-secondary);
}

.cursor-blink {
  display: inline-block;
  animation: blink 0.9s steps(1) infinite;
  color: var(--el-color-success);
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
  font-size: 12px;
  line-height: 1.6;
}
@keyframes blink { 0%, 100% { opacity: 1; } 50% { opacity: 0; } }

/* 思考过程折叠区（Claude Code 风格：灰字、左边框、小号） */
.thinking-block {
  margin-bottom: 6px;
  border-left: 2px solid var(--el-border-color);
  padding-left: 8px;
}
.thinking-head {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  cursor: pointer;
  user-select: none;
  display: flex;
  align-items: center;
  gap: 4px;
}
.thinking-head:hover { color: var(--el-color-primary); }
.thinking-star { color: var(--el-color-warning); margin-right: 2px; }
.thinking-body {
  font-size: 12px;
  font-style: italic;
  color: var(--el-text-color-secondary);
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
  padding: 4px 0;
}

/* 建议问题：终端菜单式（方括号前缀），非 tag 胶囊 */
.suggest-row { display: flex; flex-direction: column; gap: 2px; margin-top: 6px; }
.suggest-chip {
  cursor: pointer;
  font-size: 12.5px;
  color: var(--el-color-primary);
  padding: 1px 0 1px 4px;
  border-left: 2px solid transparent;
}
.suggest-chip::before { content: '› '; color: var(--el-text-color-secondary); }
.suggest-chip:hover {
  color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
  border-left-color: var(--el-color-primary);
}

/* 元信息：等宽小字，弱化 */
.msg-meta {
  margin-top: 3px;
  font-size: 11.5px;
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
  color: var(--el-text-color-placeholder);
  display: flex;
  align-items: center;
  gap: 12px;
  min-height: 16px;
}
.err-line { color: var(--el-color-danger); }
.stop-line { color: var(--el-text-color-placeholder); }
.meta-item { opacity: 0.9; }
.regen-btn { cursor: pointer; font-size: 13px; }
.regen-btn:hover { color: var(--el-color-primary); }
</style>

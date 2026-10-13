<template>
  <div class="chat-input-area">
    <div class="input-box">
      <div class="input-prompt-line">
        <span class="prompt-symbol">&gt;</span>
        <span v-if="streaming" class="streaming-hint">生成中…（Esc 停止）</span>
      </div>
      <div class="input-row">
        <el-input
          ref="textareaRef"
          v-model="input"
          type="textarea"
          :autosize="{ minRows: 1, maxRows: 10 }"
          resize="none"
          :disabled="streaming"
          class="terminal-input"
          placeholder="输入问题，Enter 发送（Shift+Enter 换行）"
          @keydown.enter.exact.prevent="handleSend"
          @keydown.esc.prevent="$emit('stop')"
        />
        <el-select
          :model-value="modelId"
          class="model-select"
          size="default"
          :disabled="streaming"
          placeholder="模型"
          @change="(v) => $emit('model', v)"
        >
          <el-option v-for="m in models" :key="m.modelId" :label="m.modelName" :value="m.modelId" />
        </el-select>
        <el-button v-if="!streaming" class="go-btn" :disabled="!input.trim()" @click="handleSend">发送 ⏎</el-button>
        <el-button v-else class="stop-btn" type="warning" plain @click="$emit('stop')">停止 ■</el-button>
        <el-button class="params-btn" text icon="Setting" title="会话参数" :disabled="streaming" @click="$emit('params')" />
      </div>
    </div>
  </div>
</template>

<script setup name="ChatInput">
const emit = defineEmits(['send', 'stop', 'model', 'params'])
defineProps({
  streaming: { type: Boolean, default: false },
  models: { type: Array, default: () => [] },
  modelId: { type: Number, default: null }
})

const input = ref('')
// 暴露 focus 供父组件在新建会话后调用（光标自动落输入框）
const textareaRef = ref(null)

function focusInput() {
  const el = textareaRef.value
  // EP textarea 组件实例：focus 方法在原生实例上；取不到则退回 querySelector
  if (el && typeof el.focus === 'function') {
    el.focus()
  } else if (el && el.$el) {
    el.$el.querySelector('textarea')?.focus()
  }
}

defineExpose({ focusInput })

function handleSend() {
  const content = input.value.trim()
  if (!content) return
  emit('send', content)
  input.value = ''
}
</script>

<style scoped>
.chat-input-area { padding-top: 10px; }
.input-box {
  border: 1px solid var(--el-border-color);
  border-radius: 6px;
  padding: 8px 10px 10px;
  background: var(--el-bg-color);
}
.input-prompt-line { display: flex; gap: 8px; align-items: center; margin-bottom: 4px; }
.prompt-symbol { color: var(--el-color-primary); font-family: 'JetBrains Mono', Consolas, Menlo, monospace; font-weight: 700; }
.streaming-hint { font-size: 11.5px; color: var(--el-text-color-placeholder); font-family: 'JetBrains Mono', Consolas, Menlo, monospace; }
.input-row { display: flex; gap: 8px; align-items: flex-end; }
.terminal-input { flex: 1; }
.terminal-input :deep(.el-textarea__inner) {
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
  font-size: 12.5px;
  box-shadow: none;
  border: none;
  padding: 2px 0 0;
  background: transparent;
}
.model-select { width: 150px; flex-shrink: 0; }
.go-btn { flex-shrink: 0; }
</style>

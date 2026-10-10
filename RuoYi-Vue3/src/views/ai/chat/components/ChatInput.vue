<template>
  <div class="chat-input-area">
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
    <el-input
      v-model="input"
      type="textarea"
      :rows="2"
      resize="none"
      :disabled="streaming"
      placeholder="输入问题，Enter 发送（Shift+Enter 换行）"
      @keydown.enter.exact.prevent="handleSend"
    />
    <el-button
      v-if="!streaming"
      type="primary"
      class="send-btn"
      :disabled="!input.trim()"
      @click="handleSend"
    >发送</el-button>
    <el-button v-else type="warning" class="send-btn" @click="$emit('stop')">停止</el-button>
  </div>
</template>

<script setup name="ChatInput">
const emit = defineEmits(['send', 'stop', 'model'])
defineProps({
  streaming: { type: Boolean, default: false },
  models: { type: Array, default: () => [] },
  modelId: { type: Number, default: null }
})

const input = ref('')

function handleSend() {
  const content = input.value.trim()
  if (!content) return
  emit('send', content)
  input.value = ''
}
</script>

<style scoped>
.chat-input-area { display: flex; gap: 8px; padding-top: 10px; align-items: flex-end; }
.model-select { width: 170px; flex-shrink: 0; align-self: flex-start; }
.send-btn { height: 54px; width: 76px; }
</style>

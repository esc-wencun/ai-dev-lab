<template>
  <div ref="scrollRef" class="msg-list" @scroll="onScroll">
    <el-empty v-if="messages.length === 0" description="开始你的第一轮对话" />
    <MessageItem
      v-for="m in messages"
      :key="m.messageId"
      :message="m"
      :streaming="streaming"
      :suggested-questions="suggestedQuestions"
      @regenerate="$emit('regenerate', $event)"
      @ask="$emit('ask', $event)"
    />
    <transition name="el-fade-in">
      <div v-if="showBackBottom" class="back-bottom" @click="scrollToBottom">
        <el-icon><ArrowDown /></el-icon>
      </div>
    </transition>
  </div>
</template>

<script setup name="MessageList">
import { ArrowDown } from '@element-plus/icons-vue'
import MessageItem from './MessageItem.vue'

const props = defineProps({
  messages: { type: Array, default: () => [] },
  streaming: { type: Boolean, default: false },
  suggestedQuestions: { type: Array, default: () => [] }
})
defineEmits(['regenerate', 'ask'])

const scrollRef = ref(null)
const showBackBottom = ref(false)
const followBottom = ref(true)

function onScroll() {
  const el = scrollRef.value
  if (!el) return
  const nearBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 60
  followBottom.value = nearBottom
  showBackBottom.value = !nearBottom
}

function scrollToBottom() {
  const el = scrollRef.value
  if (el) {
    el.scrollTop = el.scrollHeight
    followBottom.value = true
    showBackBottom.value = false
  }
}

watch(
  () => props.messages.map(m => String(m.content ?? '').length + String(m.msgStatus)).join('|'),
  () => {
    if (followBottom.value) {
      nextTick(() => {
        const el = scrollRef.value
        if (el) el.scrollTop = el.scrollHeight
      })
    }
  }
)
</script>

<style scoped>
.msg-list { flex: 1; min-height: 0; overflow-y: auto; padding: 6px 14px; position: relative; border: 1px solid var(--el-border-color-light); border-radius: 6px; background: var(--el-bg-color); }
.back-bottom { position: absolute; right: 20px; bottom: 20px; width: 32px; height: 32px; border-radius: 50%; background: var(--el-bg-color-overlay); border: 1px solid var(--el-border-color); display: flex; align-items: center; justify-content: center; cursor: pointer; box-shadow: var(--el-box-shadow-light); }
.back-bottom:hover { color: var(--el-color-primary); }
</style>

<template>
  <div class="conv-sidebar">
    <div class="conv-head">
      <span class="conv-title">会话</span>
      <el-button type="primary" size="small" @click="$emit('create')">
        <el-icon><Plus /></el-icon>
        新会话
      </el-button>
    </div>
    <div class="agent-bar">
      <el-select
        :model-value="agentId"
        placeholder="自由对话"
        size="small"
        style="width: 100%"
        @change="(v) => $emit('agent', v)"
      >
        <el-option label="自由对话（不选角色）" :value="null" />
        <el-option v-for="a in agents" :key="a.agentId" :label="a.agentName" :value="a.agentId" />
      </el-select>
      <div v-if="agent" class="agent-desc">{{ agent.description }}</div>
    </div>
    <div class="conv-list">
      <el-empty v-if="conversations.length === 0" description="暂无会话" :image-size="60" />
      <div
        v-for="c in conversations"
        :key="c.conversationId"
        class="conv-item"
        :class="{ active: c.conversationId === activeId, pinned: c.pinned }"
        @click="$emit('select', c.conversationId)"
      >
        <el-icon v-if="c.pinned" class="pin-flag"><Top /></el-icon>
        <span class="conv-name" :title="c.title">{{ c.title }}</span>
        <span class="conv-ops" @click.stop>
          <el-tooltip content="重命名" placement="top">
            <el-icon @click="doRename(c)"><Edit /></el-icon>
          </el-tooltip>
          <el-tooltip :content="c.pinned ? '取消置顶' : '置顶'" placement="top">
            <el-icon @click="$emit('pin', c.conversationId, !c.pinned)">
              <Bottom v-if="c.pinned" />
              <Top v-else />
            </el-icon>
          </el-tooltip>
          <el-tooltip content="删除" placement="top">
            <el-icon @click="doDelete(c)"><Delete /></el-icon>
          </el-tooltip>
        </span>
      </div>
    </div>
  </div>
</template>

<script setup name="ConversationSidebar">
import { Plus, Top, Bottom, Edit, Delete } from '@element-plus/icons-vue'

const props = defineProps({
  conversations: { type: Array, default: () => [] },
  activeId: { type: Number, default: null },
  agents: { type: Array, default: () => [] },
  agentId: { type: Number, default: null }
})
const agent = computed(() => props.agents.find(a => a.agentId === props.agentId) || null)
const emit = defineEmits(['select', 'create', 'rename', 'pin', 'delete', 'agent'])

function doRename(c) {
  const title = window.prompt('重命名会话', c.title)
  if (title && title.trim() && title !== c.title) emit('rename', c.conversationId, title.trim())
}

function doDelete(c) {
  ElMessageBox.confirm('删除会话「' + c.title + '」及其全部消息？', '提示', { type: 'warning' })
    .then(() => emit('delete', c.conversationId))
    .catch(() => {})
}
</script>

<style scoped>
.conv-sidebar { width: 280px; flex-shrink: 0; display: flex; flex-direction: column; border: 1px solid var(--el-border-color-light); border-radius: 6px; background: var(--el-bg-color); }
.conv-head { display: flex; align-items: center; justify-content: space-between; padding: 10px 12px; border-bottom: 1px solid var(--el-border-color-lighter); }
.conv-title { font-weight: 600; }
.conv-list { flex: 1; overflow-y: auto; padding: 6px; }
.agent-bar { padding: 8px 10px; border-bottom: 1px solid var(--el-border-color-lighter); }
.agent-desc { margin-top: 4px; font-size: 12px; color: var(--el-text-color-secondary); line-height: 1.4; }
.conv-item { display: flex; align-items: center; gap: 6px; padding: 8px 10px; border-radius: 6px; cursor: pointer; margin-bottom: 2px; }
.conv-item:hover { background: var(--el-fill-color-light); }
.conv-item.active { background: var(--el-color-primary-light-9); }
.conv-item.pinned .conv-name { font-weight: 600; }
.pin-flag { color: var(--el-color-warning); flex-shrink: 0; }
.conv-name { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 13px; }
.conv-ops { display: none; gap: 8px; color: var(--el-text-color-secondary); flex-shrink: 0; }
.conv-item:hover .conv-ops { display: inline-flex; }
.conv-ops .el-icon:hover { color: var(--el-color-primary); }
</style>

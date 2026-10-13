<template>
  <div class="app-container">
    <el-form :model="queryParams" ref="queryRef" :inline="true" v-show="showSearch">
      <el-form-item label="智能体名称" prop="agentName">
        <el-input v-model="queryParams.agentName" placeholder="请输入智能体名称" clearable style="width: 180px" @keyup.enter="handleQuery" />
      </el-form-item>
      <el-form-item label="分类" prop="category">
        <el-input v-model="queryParams.category" placeholder="请输入分类" clearable style="width: 140px" @keyup.enter="handleQuery" />
      </el-form-item>
      <el-form-item label="可见性" prop="visibility">
        <el-select v-model="queryParams.visibility" placeholder="可见性" clearable style="width: 120px">
          <el-option label="私有" :value="0" />
          <el-option label="公开" :value="1" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态" prop="status">
        <el-select v-model="queryParams.status" placeholder="状态" clearable style="width: 120px">
          <el-option v-for="dict in sys_normal_disable" :key="dict.value" :label="dict.label" :value="dict.value" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" icon="Search" @click="handleQuery">搜索</el-button>
        <el-button icon="Refresh" @click="resetQuery">重置</el-button>
      </el-form-item>
    </el-form>

    <el-row :gutter="10" class="mb8">
      <el-col :span="1.5">
        <el-button type="primary" plain icon="Plus" @click="handleAdd" v-hasPermi="['ai:agent:add']">新增</el-button>
      </el-col>
      <right-toolbar v-model:showSearch="showSearch" @queryTable="getList"></right-toolbar>
    </el-row>

    <el-table v-loading="loading" :data="agentList">
      <el-table-column label="编号" align="center" prop="agentId" width="70" />
      <el-table-column label="图标" align="center" width="70">
        <template #default="scope">
          <el-image v-if="scope.row.avatar" :src="avatarUrl(scope.row.avatar)" style="width: 24px; height: 24px" fit="contain" :preview-src-list="[avatarUrl(scope.row.avatar)]" preview-teleported hide-on-click-modal />
          <span v-else>—</span>
        </template>
      </el-table-column>
      <el-table-column label="名称" align="center" prop="agentName" show-overflow-tooltip />
      <el-table-column label="分类" align="center" prop="category" width="100">
        <template #default="scope">
          <el-tag v-if="scope.row.category" size="small">{{ scope.row.category }}</el-tag>
          <span v-else>—</span>
        </template>
      </el-table-column>
      <el-table-column label="默认模型" align="center" width="160">
        <template #default="scope">
          <span>{{ modelName(scope.row.modelId) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="可见性" align="center" width="80">
        <template #default="scope">
          <el-tag size="small" :type="scope.row.visibility === 1 ? 'success' : 'info'">{{ scope.row.visibility === 1 ? '公开' : '私有' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="排序" align="center" prop="sortOrder" width="70" />
      <el-table-column label="状态" align="center" prop="status" width="80">
        <template #default="scope">
          <dict-tag :options="sys_normal_disable" :value="scope.row.status" />
        </template>
      </el-table-column>
      <el-table-column label="备注" align="center" prop="remark" show-overflow-tooltip />
      <el-table-column label="操作" width="140" align="center" class-name="small-padding fixed-width">
        <template #default="scope">
          <el-button link type="primary" icon="Edit" @click="handleUpdate(scope.row)" v-hasPermi="['ai:agent:edit']">修改</el-button>
          <el-button link type="primary" icon="Delete" @click="handleDelete(scope.row)" v-hasPermi="['ai:agent:remove']">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog :title="title" v-model="open" width="640px" append-to-body>
      <el-form ref="agentRef" :model="form" :rules="rules" label-width="100px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="智能体名称" prop="agentName">
              <el-input v-model="form.agentName" placeholder="唯一名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="分类" prop="category">
              <el-input v-model="form.category" placeholder="如：办公 / 写作" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="一句话描述" prop="description">
          <el-input v-model="form.description" placeholder="显示在对话页智能体头部" />
        </el-form-item>
        <el-form-item label="图标" prop="avatar">
          <image-upload v-model="form.avatar" :limit="1" />
        </el-form-item>
        <el-form-item label="系统提示词" prop="systemPrompt">
          <el-input v-model="form.systemPrompt" type="textarea" :rows="6" placeholder="智能体的人设与行为规则（核心字段）" />
        </el-form-item>
        <el-form-item label="默认模型" prop="modelId">
          <el-select v-model="form.modelId" placeholder="会话默认使用的模型" clearable filterable style="width: 100%">
            <el-option v-for="m in chatModels" :key="m.modelId" :label="m.modelName" :value="m.modelId" />
          </el-select>
        </el-form-item>
        <el-form-item label="深度思考" prop="enableThinking">
          <el-switch v-model="form.enableThinking" />
        </el-form-item>
        <el-form-item label="开场白" prop="welcomeMessage">
          <el-input v-model="form.welcomeMessage" type="textarea" :rows="2" placeholder="新建会话时显示的欢迎语（可空）" />
        </el-form-item>
        <el-form-item label="建议问题" prop="suggestedQuestions">
          <div class="sq-editor">
            <div v-for="(q, i) in sqRows" :key="i" class="sq-row">
              <el-input v-model="sqRows[i]" placeholder="建议问题内容" maxlength="100" />
              <el-button link type="danger" icon="Delete" @click="sqRows.splice(i, 1)" />
            </div>
            <el-button link type="primary" icon="Plus" @click="sqRows.push('')">添加建议问题</el-button>
          </div>
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="可见性" prop="visibility">
              <el-radio-group v-model="form.visibility">
                <el-radio :value="0">私有</el-radio>
                <el-radio :value="1">公开</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="显示排序" prop="sortOrder">
              <el-input-number v-model="form.sortOrder" controls-position="right" :min="0" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio v-for="dict in sys_normal_disable" :key="dict.value" :value="dict.value">{{ dict.label }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input v-model="form.remark" type="textarea" placeholder="备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button type="primary" @click="submitForm">确 定</el-button>
          <el-button @click="cancel">取 消</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup name="AiAgent">
import { listAgent, getAgent, addAgent, updateAgent, delAgent } from "@/api/ai/agent"
import { listChatModels } from "@/api/ai/real/conversation"
import { isExternal } from "@/utils/validate"

const { proxy } = getCurrentInstance()
const { sys_normal_disable } = useDict("sys_normal_disable")

const agentList = ref([])
const chatModels = ref([])
const open = ref(false)
const loading = ref(true)
const showSearch = ref(true)
const title = ref("")
// 建议问题动态行（form.suggestedQuestions 存 JSON 字符串，编辑态用行数组）
const sqRows = ref([])

const data = reactive({
  form: {},
  queryParams: {
    agentName: undefined,
    category: undefined,
    visibility: undefined,
    status: undefined
  },
  rules: {
    agentName: [{ required: true, message: "智能体名称不能为空", trigger: "blur" }],
    systemPrompt: [{ required: true, message: "系统提示词不能为空", trigger: "blur" }]
  }
})

const { queryParams, form, rules } = toRefs(data)

// 上传文件相对路径回显用的 API 前缀（/dev-api），与 ImageUpload 组件同源
const baseUrl = import.meta.env.VITE_APP_BASE_API

/** 模型ID→名称（下拉数据源映射） */
function modelName(modelId) {
  if (modelId == null) return "—"
  const m = chatModels.value.find(x => x.modelId === modelId)
  return m ? m.modelName : `#${modelId}`
}

/** 图标 URL 拼接：相对路径（/profile/...）补 VITE_APP_BASE_API（/dev-api）前缀，外链与完整路径原样返回 */
function avatarUrl(path) {
  if (!path) return ""
  if (isExternal(path)) return path
  if (path.startsWith(baseUrl)) return path
  return baseUrl + path
}

/** 建议问题出口→行数组：数组直用；字符串尝试 parse；坏值整串占一行可编辑修正 */
function toRows(val) {
  if (Array.isArray(val)) return val.filter(x => typeof x === 'string')
  if (typeof val === 'string' && val.trim()) {
    try {
      const arr = JSON.parse(val)
      if (Array.isArray(arr)) return arr.filter(x => typeof x === 'string')
    } catch (e) { /* fall through */ }
    return [val]
  }
  return []
}

/** 行数组→入口字符串：去空行后序列化；全空存空串 */
function fromRows(rows) {
  const arr = (rows || []).map(s => (s || '').trim()).filter(s => s)
  return arr.length ? JSON.stringify(arr) : ''
}

function getList() {
  loading.value = true
  listAgent(queryParams.value).then(response => {
    agentList.value = response || []
    loading.value = false
  })
}

function cancel() {
  open.value = false
  reset()
}

function reset() {
  form.value = {
    agentId: undefined,
    agentName: undefined,
    avatar: undefined,
    category: undefined,
    description: undefined,
    systemPrompt: undefined,
    modelId: null,
    enableThinking: false,
    welcomeMessage: undefined,
    suggestedQuestions: undefined,
    visibility: 1,
    sortOrder: 0,
    status: "0",
    remark: undefined
  }
  sqRows.value = []
  proxy.resetForm("agentRef")
}

function handleQuery() {
  getList()
}

function resetQuery() {
  proxy.resetForm("queryRef")
  handleQuery()
}

function handleAdd() {
  reset()
  open.value = true
  title.value = "添加智能体"
}

function handleUpdate(row) {
  reset()
  getAgent(row.agentId).then(response => {
    form.value = response
    sqRows.value = toRows(response.suggestedQuestions)
    if (form.value.enableThinking == null) form.value.enableThinking = false
    open.value = true
    title.value = "修改智能体"
  })
}

function submitForm() {
  proxy.$refs["agentRef"].validate(valid => {
    if (valid) {
      form.value.suggestedQuestions = fromRows(sqRows.value)
      if (form.value.agentId != undefined) {
        updateAgent(form.value).then(response => {
          proxy.$modal.msgSuccess("修改成功")
          open.value = false
          getList()
        })
      } else {
        addAgent(form.value).then(response => {
          proxy.$modal.msgSuccess("新增成功")
          open.value = false
          getList()
        })
      }
    }
  })
}

function handleDelete(row) {
  const agentId = row.agentId
  proxy.$modal.confirm(`是否确认删除智能体"${row.agentName}"？被会话引用的智能体会被拒绝删除。`).then(function() {
    return delAgent(agentId)
  }).then(() => {
    getList()
    proxy.$modal.msgSuccess("删除成功")
  }).catch(() => {})
}

getList()
listChatModels().then(list => { chatModels.value = list || [] })
</script>

<style scoped>
.sq-editor { width: 100%; }
.sq-row { display: flex; gap: 6px; margin-bottom: 6px; align-items: center; }
</style>

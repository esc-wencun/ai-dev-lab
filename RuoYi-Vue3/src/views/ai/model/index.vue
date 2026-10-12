<template>
  <div class="app-container">
    <el-form :model="queryParams" ref="queryRef" :inline="true" v-show="showSearch">
      <el-form-item label="所属厂商" prop="providerId">
        <el-select v-model="queryParams.providerId" placeholder="全部厂商" clearable style="width: 170px">
          <el-option v-for="p in providerOptions" :key="p.providerId" :label="p.providerName" :value="p.providerId" />
        </el-select>
      </el-form-item>
      <el-form-item label="模型类型" prop="modelType">
        <el-select v-model="queryParams.modelType" placeholder="全部类型" clearable style="width: 140px">
          <el-option v-for="dict in ai_model_type" :key="dict.value" :label="dict.label" :value="dict.value" />
        </el-select>
      </el-form-item>
      <el-form-item label="模型名称" prop="modelName">
        <el-input v-model="queryParams.modelName" placeholder="请输入模型名称" clearable style="width: 170px" @keyup.enter="handleQuery" />
      </el-form-item>
      <el-form-item label="状态" prop="status">
        <el-select v-model="queryParams.status" placeholder="模型状态" clearable style="width: 140px">
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
        <el-button type="primary" plain icon="Plus" @click="handleAdd" v-hasPermi="['ai:model:add']">新增</el-button>
      </el-col>
      <el-col :span="1.5">
        <el-button type="success" plain icon="Edit" :disabled="single" @click="handleUpdate()" v-hasPermi="['ai:model:edit']">修改</el-button>
      </el-col>
      <el-col :span="1.5">
        <el-button type="danger" plain icon="Delete" :disabled="multiple" @click="handleDelete()" v-hasPermi="['ai:model:remove']">删除</el-button>
      </el-col>
      <right-toolbar v-model:showSearch="showSearch" @queryTable="getList"></right-toolbar>
    </el-row>
    <el-table v-loading="loading" :data="modelList" @selection-change="handleSelectionChange">
      <el-table-column type="selection" width="55" align="center" />
      <el-table-column label="编号" align="center" prop="modelId" width="70" />
      <el-table-column label="模型名称" align="center" prop="modelName" />
      <el-table-column label="模型标识" align="center" prop="modelCode" />
      <el-table-column label="类型" align="center" prop="modelType" width="90">
        <template #default="scope">
          <dict-tag :options="ai_model_type" :value="scope.row.modelType" />
        </template>
      </el-table-column>
      <el-table-column label="能力" align="center" width="200">
        <template #default="scope">
          <el-tag v-if="scope.row.supportThinking" size="small" type="warning">思考</el-tag>
          <el-tag v-if="scope.row.supportTool" size="small" type="success">工具</el-tag>
          <el-tag v-if="scope.row.supportVision" size="small">视觉</el-tag>
          <el-tag v-if="scope.row.supportAudio" size="small" type="info">音频</el-tag>
          <el-tag v-if="scope.row.supportVideo" size="small" type="danger">视频</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" align="center" prop="status" width="80">
        <template #default="scope">
          <dict-tag :options="sys_normal_disable" :value="scope.row.status" />
        </template>
      </el-table-column>
      <el-table-column label="排序" align="center" prop="sortOrder" width="70" />
      <el-table-column label="操作" width="140" align="center" class-name="small-padding fixed-width">
        <template #default="scope">
          <el-button link type="primary" icon="Edit" @click="handleUpdate(scope.row)" v-hasPermi="['ai:model:edit']">修改</el-button>
          <el-button link type="primary" icon="Delete" @click="handleDelete(scope.row)" v-hasPermi="['ai:model:remove']">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <pagination v-show="total > 0" :total="total" v-model:page="queryParams.pageNum" v-model:limit="queryParams.pageSize" @pagination="getList" />
    <el-dialog :title="title" v-model="open" width="620px" append-to-body>
      <el-form ref="modelRef" :model="form" :rules="rules" label-width="110px">
        <el-form-item label="所属厂商" prop="providerId">
          <el-select v-model="form.providerId" placeholder="选择厂商" :disabled="form.modelId != undefined" @change="onProviderChange">
            <el-option v-for="p in providerOptions" :key="p.providerId" :label="p.providerName" :value="p.providerId" />
          </el-select>
        </el-form-item>
        <el-form-item label="绑定密钥" prop="keyId">
          <el-select v-model="form.keyId" placeholder="先选厂商，再选密钥">
            <el-option v-for="k in keyOptions" :key="k.keyId" :label="k.keyName" :value="k.keyId" />
          </el-select>
        </el-form-item>
        <el-form-item label="模型名称" prop="modelName">
          <el-input v-model="form.modelName" placeholder="展示名，如 GLM-4.7" />
        </el-form-item>
        <el-form-item label="模型标识" prop="modelCode">
          <el-input v-model="form.modelCode" placeholder="传给供应商 API 的真实名字，如 glm-4.7" />
        </el-form-item>
        <el-form-item label="模型类型" prop="modelType">
          <el-select v-model="form.modelType" placeholder="选择类型">
            <el-option v-for="dict in ai_model_type" :key="dict.value" :label="dict.label" :value="Number(dict.value)" />
          </el-select>
        </el-form-item>
        <el-form-item label="能力位" prop="supportThinking">
          <el-checkbox v-model="form.supportThinking">深度思考</el-checkbox>
          <el-checkbox v-model="form.supportTool">工具调用</el-checkbox>
          <el-checkbox v-model="form.supportVision">图片输入</el-checkbox>
          <el-checkbox v-model="form.supportAudio">音频输入</el-checkbox>
          <el-checkbox v-model="form.supportVideo">视频输入</el-checkbox>
        </el-form-item>
        <el-form-item v-if="form.modelType === 5" label="向量维度" prop="modelDimension">
          <el-input-number v-model="form.modelDimension" controls-position="right" :min="1" />
        </el-form-item>
        <el-form-item label="默认温度" prop="temperature">
          <el-input-number v-model="form.temperature" :precision="2" :step="0.1" :min="0" :max="2" controls-position="right" placeholder="空=供应商默认" />
        </el-form-item>
        <el-form-item label="最大输出" prop="maxTokens">
          <el-input-number v-model="form.maxTokens" controls-position="right" :min="1" placeholder="空=供应商默认" />
        </el-form-item>
        <el-form-item label="历史条数" prop="maxContexts">
          <el-input-number v-model="form.maxContexts" controls-position="right" :min="0" placeholder="多轮裁剪窗口" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio v-for="dict in sys_normal_disable" :key="dict.value" :value="dict.value">{{ dict.label }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="排序" prop="sortOrder">
          <el-input-number v-model="form.sortOrder" controls-position="right" :min="0" />
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input v-model="form.remark" type="textarea" placeholder="选型建议" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div style="display: flex; justify-content: space-between; width: 100%">
          <el-button
            v-if="form.modelId != undefined"
            type="success"
            plain
            icon="Connection"
            :loading="testing"
            @click="handleTest"
          >测试连接</el-button>
          <span v-else>保存后可测试连通性</span>
          <span>
            <el-button type="primary" @click="submitForm">确 定</el-button>
            <el-button @click="cancel">取 消</el-button>
          </span>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup name="AiModel">
import { listModel, getModel, addModel, updateModel, delModel, testModel } from "@/api/ai/model"
import { selectAllProvider } from "@/api/ai/provider"
import { listKeyByProvider } from "@/api/ai/modelKey"

const { proxy } = getCurrentInstance()
const { sys_normal_disable, ai_model_type } = useDict("sys_normal_disable", "ai_model_type")

const modelList = ref([])
const providerOptions = ref([])
const keyOptions = ref([])
const open = ref(false)
const loading = ref(true)
const showSearch = ref(true)
const ids = ref([])
const single = ref(true)
const multiple = ref(true)
const total = ref(0)
const title = ref("")

const data = reactive({
  form: {},
  queryParams: {
    pageNum: 1,
    pageSize: 10,
    providerId: undefined,
    modelType: undefined,
    modelName: undefined,
    status: undefined
  },
  rules: {
    providerId: [{ required: true, message: "所属厂商不能为空", trigger: "change" }],
    keyId: [{ required: true, message: "绑定密钥不能为空", trigger: "change" }],
    modelName: [{ required: true, message: "模型名称不能为空", trigger: "blur" }],
    modelCode: [{ required: true, message: "模型标识不能为空", trigger: "blur" }],
    modelType: [{ required: true, message: "模型类型不能为空", trigger: "change" }],
    sortOrder: [{ required: true, message: "显示顺序不能为空", trigger: "blur" }]
  }
})

const { queryParams, form, rules } = toRefs(data)

function getList() {
  loading.value = true
  listModel(queryParams.value).then(response => {
    modelList.value = response.rows
    total.value = response.total
    loading.value = false
  })
}

function loadProviders() {
  selectAllProvider().then(response => {
    providerOptions.value = response.data
  })
}

/** 连通性测试：发送最小请求验证 地址/密钥/模型 全链路（specs/2.0.3） */
const testing = ref(false)
async function handleTest() {
  if (form.value.modelId == undefined) return
  testing.value = true
  try {
    const res = await testModel(form.value.modelId)
    const d = res.data
    if (d.success) {
      proxy.$modal.alert("连通正常（" + d.durationMs + "ms）" + (d.reply ? "\n模型回复：" + d.reply : ""))
    } else {
      proxy.$modal.alert("测试失败：" + d.message)
    }
  } finally {
    testing.value = false
  }
}

/** 厂商变更 → 重载该厂商的启用密钥下拉 */
function onProviderChange(providerId) {
  form.value.keyId = undefined
  keyOptions.value = []
  if (providerId) {
    listKeyByProvider(providerId).then(response => {
      keyOptions.value = response.data
    })
  }
}

function cancel() {
  open.value = false
  reset()
}

function reset() {
  form.value = {
    modelId: undefined,
    providerId: undefined,
    keyId: undefined,
    modelName: undefined,
    modelCode: undefined,
    modelType: 1,
    supportThinking: false,
    supportTool: false,
    supportVision: false,
    supportAudio: false,
    supportVideo: false,
    modelDimension: undefined,
    temperature: undefined,
    maxTokens: undefined,
    maxContexts: undefined,
    status: "0",
    sortOrder: 0,
    remark: undefined
  }
  keyOptions.value = []
  proxy.resetForm("modelRef")
}

function handleQuery() {
  queryParams.value.pageNum = 1
  getList()
}

function resetQuery() {
  proxy.resetForm("queryRef")
  handleQuery()
}

function handleSelectionChange(selection) {
  ids.value = selection.map(item => item.modelId)
  single.value = selection.length != 1
  multiple.value = !selection.length
}

function handleAdd() {
  reset()
  open.value = true
  title.value = "添加模型"
}

function handleUpdate(row) {
  reset()
  const modelId = row.modelId || ids.value
  getModel(modelId).then(response => {
    form.value = response.data
    if (form.value.providerId) {
      listKeyByProvider(form.value.providerId).then(r => {
        keyOptions.value = r.data
      })
    }
    open.value = true
    title.value = "修改模型"
  })
}

function submitForm() {
  proxy.$refs["modelRef"].validate(valid => {
    if (valid) {
      if (form.value.modelId != undefined) {
        updateModel(form.value).then(response => {
          proxy.$modal.msgSuccess("修改成功")
          open.value = false
          getList()
        })
      } else {
        addModel(form.value).then(response => {
          proxy.$modal.msgSuccess("新增成功")
          open.value = false
          getList()
        })
      }
    }
  })
}

function handleDelete(row) {
  const modelIds = row.modelId || ids.value
  proxy.$modal.confirm("是否确认删除所选模型？").then(function() {
    return delModel(modelIds)
  }).then(() => {
    getList()
    proxy.$modal.msgSuccess("删除成功")
  }).catch(() => {})
}

loadProviders()
getList()
</script>

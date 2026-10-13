<template>
  <div class="app-container">
    <el-form :model="queryParams" ref="queryRef" :inline="true" v-show="showSearch">
      <el-form-item label="所属厂商" prop="providerId">
        <el-select v-model="queryParams.providerId" placeholder="全部厂商" clearable style="width: 180px">
          <el-option v-for="p in providerOptions" :key="p.providerId" :label="p.providerName" :value="p.providerId" />
        </el-select>
      </el-form-item>
      <el-form-item label="密钥名称" prop="keyName">
        <el-input v-model="queryParams.keyName" placeholder="请输入密钥名称" clearable style="width: 180px" @keyup.enter="handleQuery" />
      </el-form-item>
      <el-form-item label="状态" prop="status">
        <el-select v-model="queryParams.status" placeholder="密钥状态" clearable style="width: 160px">
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
        <el-button type="primary" plain icon="Plus" @click="handleAdd" v-hasPermi="['ai:key:add']">新增</el-button>
      </el-col>
      <el-col :span="1.5">
        <el-button type="success" plain icon="Edit" :disabled="single" @click="handleUpdate()" v-hasPermi="['ai:key:edit']">修改</el-button>
      </el-col>
      <el-col :span="1.5">
        <el-button type="danger" plain icon="Delete" :disabled="multiple" @click="handleDelete()" v-hasPermi="['ai:key:remove']">删除</el-button>
      </el-col>
      <right-toolbar v-model:showSearch="showSearch" @queryTable="getList"></right-toolbar>
    </el-row>
    <el-table v-loading="loading" :data="keyList" @selection-change="handleSelectionChange">
      <el-table-column type="selection" width="55" align="center" />
      <el-table-column label="编号" align="center" prop="keyId" width="70" />
      <el-table-column label="密钥名称" align="center" prop="keyName" />
      <el-table-column label="密钥内容" align="center" prop="apiKey" show-overflow-tooltip>
        <template #default="scope">
          <code style="font-size: 12px">{{ scope.row.apiKey }}</code>
        </template>
      </el-table-column>
      <el-table-column label="Key网关地址" align="center" prop="baseUrl" show-overflow-tooltip />
      <el-table-column label="二级地址" align="center" prop="apiPath" show-overflow-tooltip />
      <el-table-column label="状态" align="center" prop="status" width="80">
        <template #default="scope">
          <dict-tag :options="sys_normal_disable" :value="scope.row.status" />
        </template>
      </el-table-column>
      <el-table-column label="过期时间" align="center" prop="expireTime" width="160">
        <template #default="scope">
          <span>{{ parseTime(scope.row.expireTime) || '—' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="140" align="center" class-name="small-padding fixed-width">
        <template #default="scope">
          <el-button link type="primary" icon="Edit" @click="handleUpdate(scope.row)" v-hasPermi="['ai:key:edit']">修改</el-button>
          <el-button link type="primary" icon="Delete" @click="handleDelete(scope.row)" v-hasPermi="['ai:key:remove']">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <pagination v-show="total > 0" :total="total" v-model:page="queryParams.pageNum" v-model:limit="queryParams.pageSize" @pagination="getList" />
    <el-dialog :title="title" v-model="open" width="560px" append-to-body>
      <el-form ref="keyRef" :model="form" :rules="rules" label-width="110px">
        <el-form-item label="所属厂商" prop="providerId">
          <el-select v-model="form.providerId" placeholder="选择厂商" :disabled="form.keyId != undefined">
            <el-option v-for="p in providerOptions" :key="p.providerId" :label="p.providerName" :value="p.providerId" />
          </el-select>
        </el-form-item>
        <el-form-item label="密钥名称" prop="keyName">
          <el-input v-model="form.keyName" placeholder="如：官方主Key / 中转Key" />
        </el-form-item>
        <el-form-item label="密钥内容" prop="apiKey">
          <el-input v-model="form.apiKey" placeholder="真实 Key 或环境变量占位符" show-password />
          <div style="font-size: 12px; color: var(--el-text-color-secondary); line-height: 1.5">
            支持环境变量占位符 <code>${glm_api_key}</code> / <code>${DASHSCOPE_API_KEY}</code>（格式 <code>${变量名}</code>，内置数据即此写法）：调用时从服务端环境变量取值，库里不落明文；列表与回显按前4+***+后4 脱敏，占位符原样显示便于核对变量名
          </div>
        </el-form-item>
        <el-form-item label="Key网关地址" prop="baseUrl">
          <el-input v-model="form.baseUrl" placeholder="中转Key填中转地址；留空用厂商默认" />
        </el-form-item>
        <el-form-item label="二级地址" prop="apiPath">
          <el-input v-model="form.apiPath" placeholder="端点路径非标准时填；留空用协议默认" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio v-for="dict in sys_normal_disable" :key="dict.value" :value="dict.value">{{ dict.label }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="过期时间" prop="expireTime">
          <el-date-picker v-model="form.expireTime" type="datetime" placeholder="可空" value-format="YYYY-MM-DD HH:mm:ss" />
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input v-model="form.remark" type="textarea" />
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

<script setup name="AiKey">
import { listKey, listKeyByProvider, getKey, addKey, updateKey, delKey } from "@/api/ai/modelKey"
import { selectAllProvider } from "@/api/ai/provider"

const { proxy } = getCurrentInstance()
const { sys_normal_disable } = useDict("sys_normal_disable")

const keyList = ref([])
const providerOptions = ref([])
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
    keyName: undefined,
    status: undefined
  },
  rules: {
    providerId: [{ required: true, message: "所属厂商不能为空", trigger: "change" }],
    keyName: [{ required: true, message: "密钥名称不能为空", trigger: "blur" }],
    apiKey: [{ required: true, message: "密钥内容不能为空", trigger: "blur" }]
  }
})

const { queryParams, form, rules } = toRefs(data)

function getList() {
  loading.value = true
  listKey(queryParams.value).then(response => {
    keyList.value = response.rows
    total.value = response.total
    loading.value = false
  })
}

function loadProviders() {
  selectAllProvider().then(response => {
    providerOptions.value = response.data
  })
}

function cancel() {
  open.value = false
  reset()
}

function reset() {
  form.value = {
    keyId: undefined,
    providerId: undefined,
    keyName: undefined,
    apiKey: undefined,
    baseUrl: undefined,
    apiPath: undefined,
    status: "0",
    expireTime: undefined,
    remark: undefined
  }
  proxy.resetForm("keyRef")
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
  ids.value = selection.map(item => item.keyId)
  single.value = selection.length != 1
  multiple.value = !selection.length
}

function handleAdd() {
  reset()
  open.value = true
  title.value = "添加密钥"
}

function handleUpdate(row) {
  reset()
  const keyId = row.keyId || ids.value
  getKey(keyId).then(response => {
    form.value = response.data
    open.value = true
    title.value = "修改密钥"
  })
}

function submitForm() {
  proxy.$refs["keyRef"].validate(valid => {
    if (valid) {
      if (form.value.keyId != undefined) {
        updateKey(form.value).then(response => {
          proxy.$modal.msgSuccess("修改成功")
          open.value = false
          getList()
        })
      } else {
        addKey(form.value).then(response => {
          proxy.$modal.msgSuccess("新增成功")
          open.value = false
          getList()
        })
      }
    }
  })
}

function handleDelete(row) {
  const keyIds = row.keyId || ids.value
  proxy.$modal.confirm("是否确认删除所选密钥？有模型挂靠的密钥会被拒绝。").then(function() {
    return delKey(keyIds)
  }).then(() => {
    getList()
    proxy.$modal.msgSuccess("删除成功")
  }).catch(() => {})
}

loadProviders()
getList()
</script>

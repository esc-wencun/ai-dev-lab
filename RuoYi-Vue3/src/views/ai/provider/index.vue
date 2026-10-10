<template>
  <div class="app-container">
    <el-form :model="queryParams" ref="queryRef" :inline="true" v-show="showSearch">
      <el-form-item label="厂商编码" prop="providerCode">
        <el-input v-model="queryParams.providerCode" placeholder="请输入厂商编码" clearable style="width: 180px" @keyup.enter="handleQuery" />
      </el-form-item>
      <el-form-item label="厂商名称" prop="providerName">
        <el-input v-model="queryParams.providerName" placeholder="请输入厂商名称" clearable style="width: 180px" @keyup.enter="handleQuery" />
      </el-form-item>
      <el-form-item label="状态" prop="status">
        <el-select v-model="queryParams.status" placeholder="厂商状态" clearable style="width: 160px">
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
        <el-button type="primary" plain icon="Plus" @click="handleAdd" v-hasPermi="['ai:provider:add']">新增</el-button>
      </el-col>
      <el-col :span="1.5">
        <el-button type="success" plain icon="Edit" :disabled="single" @click="handleUpdate()" v-hasPermi="['ai:provider:edit']">修改</el-button>
      </el-col>
      <el-col :span="1.5">
        <el-button type="danger" plain icon="Delete" :disabled="multiple" @click="handleDelete()" v-hasPermi="['ai:provider:remove']">删除</el-button>
      </el-col>
      <right-toolbar v-model:showSearch="showSearch" @queryTable="getList"></right-toolbar>
    </el-row>

    <el-table v-loading="loading" :data="providerList" @selection-change="handleSelectionChange">
      <el-table-column type="selection" width="55" align="center" />
      <el-table-column label="图标" align="center" width="70">
        <template #default="scope">
          <el-image v-if="scope.row.icon" :src="scope.row.icon" style="width: 24px; height: 24px" fit="contain" :preview-src-list="[scope.row.icon]" preview-teleported hide-on-click-modal />
          <span v-else>—</span>
        </template>
      </el-table-column>
      <el-table-column label="编号" align="center" prop="providerId" width="70" />
      <el-table-column label="编码" align="center" prop="providerCode" />
      <el-table-column label="名称" align="center" prop="providerName" />
      <el-table-column label="协议" align="center" prop="protocol" width="100">
        <template #default="scope">
          <el-tag size="small">{{ scope.row.protocol }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="API地址" align="center" prop="apiHost" show-overflow-tooltip />
      <el-table-column label="排序" align="center" prop="sortOrder" width="70" />
      <el-table-column label="状态" align="center" prop="status" width="80">
        <template #default="scope">
          <dict-tag :options="sys_normal_disable" :value="scope.row.status" />
        </template>
      </el-table-column>
      <el-table-column label="创建时间" align="center" prop="createTime" width="160">
        <template #default="scope">
          <span>{{ parseTime(scope.row.createTime) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="140" align="center" class-name="small-padding fixed-width">
        <template #default="scope">
          <el-button link type="primary" icon="Edit" @click="handleUpdate(scope.row)" v-hasPermi="['ai:provider:edit']">修改</el-button>
          <el-button link type="primary" icon="Delete" @click="handleDelete(scope.row)" v-hasPermi="['ai:provider:remove']">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <pagination v-show="total > 0" :total="total" v-model:page="queryParams.pageNum" v-model:limit="queryParams.pageSize" @pagination="getList" />

    <el-dialog :title="title" v-model="open" width="560px" append-to-body>
      <el-form ref="providerRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="厂商名称" prop="providerName">
          <el-input v-model="form.providerName" placeholder="如：深度求索" />
        </el-form-item>
        <el-form-item label="厂商编码" prop="providerCode">
          <el-input v-model="form.providerCode" placeholder="如：deepseek（唯一）" />
        </el-form-item>
        <el-form-item label="协议类型" prop="protocol">
          <el-select v-model="form.protocol" placeholder="协议 = 客户端适配键">
            <el-option v-for="p in protocols" :key="p.value" :label="p.label" :value="p.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="API地址" prop="apiHost">
          <el-input v-model="form.apiHost" placeholder="填到版本段不带尾斜杠，如 https://api.deepseek.com" />
        </el-form-item>
        <el-form-item label="密钥申请地址" prop="keyApplyUrl">
          <el-input v-model="form.keyApplyUrl" placeholder="页面引导用户去哪申请 Key" />
        </el-form-item>
        <el-form-item label="官网" prop="website">
          <el-input v-model="form.website" />
        </el-form-item>
        <el-form-item label="图标" prop="icon">
          <el-input v-model="form.icon" placeholder="图标外链" />
        </el-form-item>
        <el-form-item label="排序" prop="sortOrder">
          <el-input-number v-model="form.sortOrder" controls-position="right" :min="0" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio v-for="dict in sys_normal_disable" :key="dict.value" :value="dict.value">{{ dict.label }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input v-model="form.remark" type="textarea" placeholder="备注 / 厂商描述" />
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

<script setup name="AiProvider">
import { listProvider, selectAllProvider, getProvider, addProvider, updateProvider, delProvider } from "@/api/ai/provider"

const { proxy } = getCurrentInstance()
const { sys_normal_disable } = useDict("sys_normal_disable")

const providerList = ref([])
const open = ref(false)
const loading = ref(true)
const showSearch = ref(true)
const ids = ref([])
const single = ref(true)
const multiple = ref(true)
const total = ref(0)
const title = ref("")

const protocols = [
  { value: "openai", label: "OpenAI 兼容" },
  { value: "anthropic", label: "Anthropic" },
  { value: "ollama", label: "Ollama" },
  { value: "gemini", label: "Gemini" },
  { value: "dify", label: "Dify" },
  { value: "coze", label: "Coze" },
  { value: "custom", label: "自定义" }
]

const data = reactive({
  form: {},
  queryParams: {
    pageNum: 1,
    pageSize: 10,
    providerCode: undefined,
    providerName: undefined,
    status: undefined
  },
  rules: {
    providerName: [{ required: true, message: "厂商名称不能为空", trigger: "blur" }],
    providerCode: [{ required: true, message: "厂商编码不能为空", trigger: "blur" }],
    protocol: [{ required: true, message: "协议类型不能为空", trigger: "change" }],
    apiHost: [{ required: true, message: "API地址不能为空", trigger: "blur" }],
    sortOrder: [{ required: true, message: "显示顺序不能为空", trigger: "blur" }]
  }
})

const { queryParams, form, rules } = toRefs(data)

function getList() {
  loading.value = true
  listProvider(queryParams.value).then(response => {
    providerList.value = response.rows
    total.value = response.total
    loading.value = false
  })
}

function cancel() {
  open.value = false
  reset()
}

function reset() {
  form.value = {
    providerId: undefined,
    providerName: undefined,
    providerCode: undefined,
    protocol: "openai",
    apiHost: undefined,
    icon: undefined,
    website: undefined,
    keyApplyUrl: undefined,
    sortOrder: 0,
    status: "0",
    remark: undefined
  }
  proxy.resetForm("providerRef")
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
  ids.value = selection.map(item => item.providerId)
  single.value = selection.length != 1
  multiple.value = !selection.length
}

function handleAdd() {
  reset()
  open.value = true
  title.value = "添加厂商"
}

function handleUpdate(row) {
  reset()
  const providerId = row.providerId || ids.value
  getProvider(providerId).then(response => {
    form.value = response.data
    open.value = true
    title.value = "修改厂商"
  })
}

function submitForm() {
  proxy.$refs["providerRef"].validate(valid => {
    if (valid) {
      if (form.value.providerId != undefined) {
        updateProvider(form.value).then(response => {
          proxy.$modal.msgSuccess("修改成功")
          open.value = false
          getList()
        })
      } else {
        addProvider(form.value).then(response => {
          proxy.$modal.msgSuccess("新增成功")
          open.value = false
          getList()
        })
      }
    }
  })
}

function handleDelete(row) {
  const providerIds = row.providerId || ids.value
  proxy.$modal.confirm('是否确认删除所选厂商？有密钥的厂商会被拒绝。').then(function() {
    return delProvider(providerIds)
  }).then(() => {
    getList()
    proxy.$modal.msgSuccess("删除成功")
  }).catch(() => {})
}

getList()
</script>

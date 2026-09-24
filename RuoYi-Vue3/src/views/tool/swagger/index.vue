<template>
   <el-result v-if="!features.swaggerDocs" icon="info" title="该功能仅 Java 版提供"
     sub-title="系统接口页基于 springdoc swagger-ui，当前后端运行时未提供等价的接口文档页面（FastAPI /docs 为另一套 UI），故不可用。">
     <template #extra>
       <el-button type="primary" @click="getPlatform">重新检测</el-button>
     </template>
   </el-result>
   <i-frame v-else v-model:src="url"></i-frame>
</template>

<script setup>
import iFrame from '@/components/iFrame'
import { getPlatformInfo } from '@/api/platform'

import { ref } from 'vue'

const url = ref(import.meta.env.VITE_APP_BASE_API + "/swagger-ui/index.html")
const features = ref({ swaggerDocs: true })

function getPlatform() {
  getPlatformInfo().then(response => {
    features.value = response.features || {}
  })
}

getPlatform()
</script>

<template>
   <div>
      <el-result v-if="!features.druidMonitor" icon="info" title="该功能仅 Java 版提供"
        sub-title="数据监控基于 Druid 连接池控制台，当前后端运行时未使用 Druid，故不提供该功能。">
        <template #extra>
          <el-button type="primary" @click="getPlatform">重新检测</el-button>
        </template>
      </el-result>
      <i-frame v-else v-model:src="url"></i-frame>
   </div>
</template>

<script setup>
import iFrame from '@/components/iFrame'
import { getPlatformInfo } from '@/api/platform'

import { ref } from 'vue'

const url = ref(import.meta.env.VITE_APP_BASE_API + '/druid/login.html')
const features = ref({ druidMonitor: true })

function getPlatform() {
  getPlatformInfo().then(response => {
    features.value = response.features || {}
  })
}

getPlatform()
</script>

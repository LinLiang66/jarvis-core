<script setup lang="ts">
import { Icon } from '@iconify/vue'
import { getFileStatisticsApi } from '@/apis/file'
import { formatFileSize } from './constants'

defineOptions({ name: 'FileAsideStatistics' })

const stats = ref({ fileCount: 0, dirCount: 0, totalSize: 0 })
const loading = ref(false)

async function loadStats() {
  loading.value = true
  try {
    stats.value = await getFileStatisticsApi()
  }
  finally {
    loading.value = false
  }
}

defineExpose({ loadStats })

onMounted(loadStats)
</script>

<template>
  <section v-loading="loading" class="file-aside-stats">
    <div class="file-aside-stats__item">
      <span class="label">存储量</span>
      <span class="value">{{ formatFileSize(stats.totalSize) }}</span>
    </div>
    <el-divider direction="vertical" />
    <div class="file-aside-stats__item">
      <span class="label">文件</span>
      <span class="value">{{ stats.fileCount }}</span>
    </div>
    <el-divider direction="vertical" />
    <div class="file-aside-stats__item">
      <span class="label">文件夹</span>
      <span class="value">{{ stats.dirCount }}</span>
    </div>
    <el-button class="refresh-btn" link @click="loadStats">
      <Icon icon="icon-park-outline:refresh" />
    </el-button>
  </section>
</template>

<style scoped>
.file-aside-stats {
  margin-top: 12px;
  padding: 16px;
  background: var(--el-bg-color);
  border-radius: 8px;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.file-aside-stats__item {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-width: 72px;
}
.label {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.value {
  font-size: 18px;
  font-weight: 600;
  color: var(--el-color-primary);
}
.refresh-btn {
  margin-left: auto;
}
</style>

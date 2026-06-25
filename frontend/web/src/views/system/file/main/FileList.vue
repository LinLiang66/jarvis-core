<script setup lang="ts">
import type { FileItem } from '@/apis/file'
import { formatFileSize, isDirectory } from '../constants'
import FileImage from './FileImage.vue'

defineOptions({ name: 'FileList' })

const props = withDefaults(defineProps<{
  data?: FileItem[]
  selectedFileIds?: string[]
  isBatchMode?: boolean
}>(), {
  data: () => [],
  selectedFileIds: () => [],
  isBatchMode: false,
})

const emit = defineEmits<{
  click: [item: FileItem]
  dblclick: [item: FileItem]
  select: [item: FileItem]
  contextmenu: [item: FileItem, event: MouseEvent]
}>()

function onSelectionChange(rows: FileItem[]) {
  const last = rows[rows.length - 1]
  if (last)
    emit('select', last)
}

function copyUrl(url?: string) {
  if (!url)
    return
  navigator.clipboard.writeText(url)
  ElMessage.success('已复制链接')
}
</script>

<template>
  <el-table
    class="file-list"
    :data="props.data"
    row-key="id"
    @selection-change="onSelectionChange"
  >
    <el-table-column
      v-if="props.isBatchMode"
      type="selection"
      width="48"
    />
    <el-table-column label="名称" min-width="220">
      <template #default="{ row }">
        <div
          class="file-name-cell"
          @click="emit('click', row)"
          @dblclick="emit('dblclick', row)"
          @contextmenu.prevent="emit('contextmenu', row, $event)"
        >
          <div class="file-image">
            <FileImage :data="row" />
          </div>
          <span class="file-name-text">{{ row.originalName }}</span>
        </div>
      </template>
    </el-table-column>
    <el-table-column label="大小" width="120">
      <template #default="{ row }">
        {{ isDirectory(row) ? '-' : formatFileSize(row.size) }}
      </template>
    </el-table-column>
    <el-table-column prop="storageName" label="存储" width="180" show-overflow-tooltip />
    <el-table-column prop="updateTime" label="修改时间" width="180" />
    <el-table-column label="操作" width="120" align="center" fixed="right">
      <template #default="{ row }">
        <el-button v-if="row.url" link type="primary" @click="copyUrl(row.url)">
          复制链接
        </el-button>
        <el-button
          v-if="row.url && !isDirectory(row)"
          link
          type="primary"
        >
          <a :href="row.url" target="_blank" rel="noopener">下载</a>
        </el-button>
      </template>
    </el-table-column>
  </el-table>
</template>

<style scoped>
.file-list {
  width: 100%;
  margin-top: 8px;
}
.file-name-cell {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
}
.file-image {
  width: 28px;
  height: 28px;
  flex-shrink: 0;
}
.file-name-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>

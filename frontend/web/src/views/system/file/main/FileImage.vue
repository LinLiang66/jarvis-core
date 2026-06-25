<script setup lang="ts">
import type { FileItem } from '@/apis/file'
import { Icon } from '@iconify/vue'
import { FileIcon, ImageTypes, isDirectory } from '../constants'

defineOptions({ name: 'FileImage' })

const props = defineProps<{
  data: FileItem
}>()

const isImage = computed(() => {
  const ext = props.data.extension?.toLowerCase()
  return ext ? ImageTypes.includes(ext) : false
})

const iconName = computed(() => {
  if (isDirectory(props.data))
    return FileIcon.dir
  const ext = props.data.extension?.toLowerCase() || 'other'
  return FileIcon[ext] || FileIcon.other
})
</script>

<template>
  <img
    v-if="isImage && data.url"
    class="file-image"
    :src="data.url"
    :alt="data.originalName"
  >
  <Icon v-else :icon="iconName" class="file-icon" />
</template>

<style scoped>
.file-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 4px;
}
.file-icon {
  width: 100%;
  height: 100%;
  color: var(--el-color-primary);
}
</style>

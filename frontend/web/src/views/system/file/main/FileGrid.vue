<script setup lang="ts">
import type { FileItem } from '@/apis/file'
import FileImage from './FileImage.vue'

defineOptions({ name: 'FileGrid' })

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

function onClick(item: FileItem) {
  emit('click', item)
}

function onDblclick(item: FileItem) {
  emit('dblclick', item)
}

function onSelect(item: FileItem) {
  emit('select', item)
}

function onContextMenu(item: FileItem, event: MouseEvent) {
  emit('contextmenu', item, event)
}
</script>

<template>
  <div class="file-grid">
    <div
      v-for="item in props.data"
      :key="item.id"
      class="file-grid-item"
      @click.stop="onClick(item)"
      @dblclick.stop="onDblclick(item)"
      @contextmenu.prevent="onContextMenu(item, $event)"
    >
      <section class="file-grid-item__wrapper">
        <div class="file-icon">
          <FileImage :data="item" />
        </div>
        <p class="file-name" :title="item.originalName">
          {{ item.originalName }}
        </p>
      </section>
      <section
        v-show="props.isBatchMode"
        class="check-mode"
        :class="{ checked: props.selectedFileIds.includes(item.id) }"
        @click.stop="onSelect(item)"
      >
        <el-checkbox :model-value="props.selectedFileIds.includes(item.id)" />
      </section>
    </div>
  </div>
</template>

<style scoped>
.file-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  gap: 12px;
  padding-top: 12px;
}
.file-grid-item {
  height: 110px;
  display: flex;
  justify-content: center;
  align-items: center;
  position: relative;
  cursor: pointer;
  border-radius: 6px;
}
.file-grid-item:hover {
  background: var(--el-fill-color-light);
}
.check-mode {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.06);
  z-index: 2;
  border-radius: 6px;
}
.check-mode.checked {
  background: transparent;
}
.check-mode :deep(.el-checkbox) {
  position: absolute;
  top: 6px;
  left: 6px;
}
.file-grid-item__wrapper {
  width: 80%;
  max-width: 96px;
  display: flex;
  flex-direction: column;
  align-items: center;
  overflow: hidden;
}
.file-icon {
  width: 56px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.file-name {
  width: 100%;
  margin: 8px 0 0;
  font-size: 12px;
  text-align: center;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>

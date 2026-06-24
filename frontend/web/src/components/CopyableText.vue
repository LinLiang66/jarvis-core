<script setup lang="ts">
import { DocumentCopy } from '@element-plus/icons-vue'
import { useClipboard } from '@vueuse/core'

defineOptions({ name: 'CopyableText' })

const props = withDefaults(defineProps<{
  text?: string
  /** 列表窄列时可截断展示 */
  truncate?: boolean
}>(), {
  truncate: true,
})

const { copy, isSupported } = useClipboard()

async function handleCopy() {
  const val = props.text?.trim()
  if (!val)
    return
  if (!isSupported.value) {
    ElMessage.warning('当前浏览器不支持复制')
    return
  }
  await copy(val)
  ElMessage.success('已复制到剪贴板')
}
</script>

<template>
  <span v-if="text" class="copyable-text">
    <span
      class="copyable-text__content"
      :class="{ 'is-truncate': truncate }"
      :title="text"
    >{{ text }}</span>
    <el-tooltip content="复制" placement="top">
      <el-button
        type="primary"
        link
        class="copyable-text__btn"
        @click.stop="handleCopy"
      >
        <el-icon><DocumentCopy /></el-icon>
      </el-button>
    </el-tooltip>
  </span>
  <span v-else class="copyable-text__empty">—</span>
</template>

<style scoped>
.copyable-text {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  max-width: 100%;
}

.copyable-text__content {
  flex: 1;
  min-width: 0;
  word-break: break-all;
}

.copyable-text__content.is-truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.copyable-text__btn {
  flex-shrink: 0;
  padding: 0 2px;
}

.copyable-text__empty {
  color: var(--el-text-color-secondary);
}
</style>

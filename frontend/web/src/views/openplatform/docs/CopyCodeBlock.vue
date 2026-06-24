<script setup lang="ts">
import { DocumentCopy } from '@element-plus/icons-vue'
import { useClipboard } from '@vueuse/core'

defineOptions({ name: 'OpenPlatformCopyCodeBlock' })

const props = defineProps<{
  content: string
}>()

const { copy, isSupported } = useClipboard()

async function handleCopy() {
  const val = props.content?.trim()
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
  <div class="copy-code-block">
    <div class="copy-code-block__bar">
      <span class="copy-code-block__lang">JSON</span>
      <el-tooltip content="复制" placement="top">
        <el-button type="primary" link class="copy-code-block__btn" @click="handleCopy">
          <el-icon><DocumentCopy /></el-icon>
          复制
        </el-button>
      </el-tooltip>
    </div>
    <pre class="copy-code-block__pre">{{ content }}</pre>
  </div>
</template>

<style scoped>
.copy-code-block {
  margin-bottom: 12px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  overflow: hidden;
}

.copy-code-block__bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 10px;
  background: var(--el-fill-color-light);
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.copy-code-block__lang {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.copy-code-block__btn {
  padding: 0 4px;
  font-size: 12px;
}

.copy-code-block__pre {
  margin: 0;
  padding: 12px 14px;
  background: #1e1e1e;
  color: #d4d4d4;
  overflow: auto;
  font-family: ui-monospace, 'Cascadia Code', monospace;
  font-size: 13px;
  line-height: 1.55;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 360px;
}
</style>

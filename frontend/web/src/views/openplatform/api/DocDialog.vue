<script setup lang="ts">
import type { OpenAPIActionItem } from '@/apis/openplatform'
import ApiDetail from '../docs/ApiDetail.vue'

defineOptions({ name: 'OpenPlatformApiDocDialog' })

const visible = ref(false)
const current = ref<OpenAPIActionItem>()

function open(row: OpenAPIActionItem) {
  current.value = row
  visible.value = true
}

defineExpose({ open })
</script>

<template>
  <GiDialog
    v-model="visible"
    :title="current ? `${current.title}（${current.action}）` : '接口文档'"
    width="920px"
    :footer="false"
  >
    <div v-if="current" class="doc-viewer">
      <ApiDetail :item="current" />
    </div>
  </GiDialog>
</template>

<style scoped>
.doc-viewer {
  max-height: 75vh;
  overflow: auto;
}

.doc-viewer :deep(.api-detail) {
  padding: 0 8px 16px;
  max-width: none;
}
</style>

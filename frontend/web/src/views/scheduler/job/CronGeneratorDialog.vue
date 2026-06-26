<script setup lang="ts">
import { ElMessage } from 'element-plus'
import CronBuilder from './cron/CronBuilder.vue'

defineOptions({ name: 'CronGeneratorDialog' })

const emit = defineEmits<{
  ok: [value: string]
}>()

const visible = ref(false)
const cronExpression = ref('')
const builderRef = useTemplateRef<InstanceType<typeof CronBuilder>>('builderRef')

function open(cron = '') {
  cronExpression.value = cron || '0 */5 * * * *'
  visible.value = true
}

async function handleBeforeOk() {
  if (!builderRef.value?.checkValid()) {
    ElMessage.warning('请修正 Cron 表达式')
    return false
  }
  emit('ok', cronExpression.value)
  return true
}

defineExpose({ open })
</script>

<template>
  <GiDialog
    v-model="visible"
    title="Cron 表达式生成器"
    width="calc(100% - 20px)"
    :style="{ maxWidth: '840px' }"
    body-class="cron-generator-dialog-body"
    top="8vh"
    destroy-on-close
    :close-on-click-modal="false"
    :on-before-ok="handleBeforeOk"
  >
    <CronBuilder ref="builderRef" v-model="cronExpression" />
  </GiDialog>
</template>

<style scoped lang="scss">
:global(.cron-generator-dialog-body) {
  max-height: min(72vh, 680px);
  overflow-y: auto;
  padding-top: 4px;
  padding-bottom: 4px;
}
</style>

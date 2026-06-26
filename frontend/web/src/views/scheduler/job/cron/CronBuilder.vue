<script setup lang="ts">
import type { CronUnitKey } from './cron-utils'
import {
  buildCronExpression,
  CronPresets,
  CronUnitMeta,
  parseCronExpression,
  validateCronExpression,
} from './cron-utils'
import CronUnitForm from './CronUnitForm.vue'

defineOptions({ name: 'CronBuilder' })

const modelValue = defineModel<string>({ default: '0 */5 * * * *' })

const activeTab = ref<CronUnitKey>('second')
const units = reactive(parseCronExpression(modelValue.value))
const validationHint = ref<string | null>(null)

const unitTabs = computed(() =>
  (Object.keys(CronUnitMeta) as CronUnitKey[]).map(key => ({
    key,
    label: CronUnitMeta[key].label,
  })),
)

function syncExpression() {
  const expr = buildCronExpression(units)
  modelValue.value = expr
  validationHint.value = validateCronExpression(expr)
}

function applyPreset(value: string) {
  Object.assign(units, parseCronExpression(value))
  syncExpression()
}

function onManualChange(value: string) {
  Object.assign(units, parseCronExpression(value))
  validationHint.value = validateCronExpression(value)
}

watch(units, syncExpression, { deep: true })

watch(modelValue, (val) => {
  if (val !== buildCronExpression(units))
    Object.assign(units, parseCronExpression(val))
}, { immediate: true })

function checkValid(): boolean {
  validationHint.value = validateCronExpression(modelValue.value)
  return !validationHint.value
}

defineExpose({ checkValid })
</script>

<template>
  <div class="cron-builder">
    <div class="cron-builder__presets">
      <span class="cron-builder__presets-label">常用预设</span>
      <el-space wrap>
        <el-tag
          v-for="item in CronPresets"
          :key="item.value"
          class="cron-builder__preset-tag"
          effect="plain"
          @click="applyPreset(item.value)"
        >
          {{ item.label }}
        </el-tag>
      </el-space>
    </div>

    <el-tabs v-model="activeTab" class="cron-builder__tabs">
      <el-tab-pane
        v-for="tab in unitTabs"
        :key="tab.key"
        :label="tab.label"
        :name="tab.key"
      >
        <CronUnitForm v-model="units[tab.key]" :unit="tab.key" />
      </el-tab-pane>
    </el-tabs>

    <div class="cron-builder__expr">
      <el-input
        :model-value="modelValue"
        placeholder="秒 分 时 日 月 周"
        @update:model-value="onManualChange(String($event))"
      >
        <template #prepend>
          表达式
        </template>
      </el-input>
      <p v-if="validationHint" class="cron-builder__hint cron-builder__hint--error">
        {{ validationHint }}
      </p>
      <p v-else class="cron-builder__hint">
        格式：秒 分 时 日 月 周（6 段，与调度引擎 robfig/cron 一致）
      </p>
    </div>
  </div>
</template>

<style scoped lang="scss">
.cron-builder {
  &__presets {
    margin-bottom: 12px;
  }

  &__presets-label {
    display: block;
    margin-bottom: 8px;
    font-size: 13px;
    color: var(--el-text-color-secondary);
  }

  &__preset-tag {
    cursor: pointer;
  }

  &__expr {
    margin-top: 12px;
  }

  &__hint {
    margin: 6px 0 0;
    font-size: 12px;
    color: var(--el-text-color-secondary);

    &--error {
      color: var(--el-color-danger);
    }
  }
}
</style>

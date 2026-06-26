<script setup lang="ts">
import type { CronUnitKey } from './cron-utils'
import { CronUnitMeta, WEEK_OPTIONS } from './cron-utils'

defineOptions({ name: 'CronUnitForm' })

const props = withDefaults(defineProps<{
  unit: CronUnitKey
  modelValue: string
  disabled?: boolean
}>(), {
  disabled: false,
})

const emit = defineEmits<{
  'change': [value: string]
  'update:modelValue': [value: string]
}>()

const TypeEnum = {
  unset: 'UNSET',
  every: 'EVERY',
  range: 'RANGE',
  loop: 'LOOP',
  specify: 'SPECIFY',
} as const

type TypeEnumValue = typeof TypeEnum[keyof typeof TypeEnum]

const meta = computed(() => CronUnitMeta[props.unit])
const type = ref<TypeEnumValue>(TypeEnum.every)
const valueRange = reactive({ start: meta.value.min, end: meta.value.max })
const valueLoop = reactive({ start: meta.value.min, interval: 1 })
const valueList = ref<number[]>([])

const specifyRange = computed(() => {
  const range: number[] = []
  for (let i = meta.value.min; i <= meta.value.max; i++)
    range.push(i)
  return range
})

const computeValue = computed(() => {
  switch (type.value) {
    case TypeEnum.unset:
      return '?'
    case TypeEnum.every:
      return '*'
    case TypeEnum.range:
      return `${valueRange.start}-${valueRange.end}`
    case TypeEnum.loop:
      return `${valueLoop.start}/${valueLoop.interval}`
    case TypeEnum.specify:
      return [...new Set(valueList.value.length ? valueList.value : [meta.value.min])]
        .sort((a, b) => a - b)
        .join(',')
    default:
      return meta.value.defaultValue
  }
})

function parseValue(value: string | undefined) {
  if (!value || value === computeValue.value)
    return
  if (value === '?') {
    type.value = TypeEnum.unset
    return
  }
  if (value === '*') {
    type.value = TypeEnum.every
    return
  }
  if (value.includes('-')) {
    type.value = TypeEnum.range
    const [start, end] = value.split('-').map(Number)
    if (!Number.isNaN(start))
      valueRange.start = start
    if (!Number.isNaN(end))
      valueRange.end = end
    return
  }
  if (value.includes('/')) {
    type.value = TypeEnum.loop
    const [start, interval] = value.split('/')
    valueLoop.start = start === '*' ? meta.value.min : Number(start)
    if (!Number.isNaN(Number(interval)))
      valueLoop.interval = Number(interval)
    return
  }
  if (value.includes(',') || !Number.isNaN(Number(value))) {
    type.value = TypeEnum.specify
    valueList.value = value.split(',').map(Number).filter(n => !Number.isNaN(n))
    if (!valueList.value.length)
      valueList.value = [meta.value.min]
  }
}

function emitValue(value: string) {
  emit('change', value)
  emit('update:modelValue', value)
}

watch(() => props.modelValue, val => parseValue(val), { immediate: true })
watch(computeValue, val => emitValue(val))

const weekOptions = WEEK_OPTIONS
</script>

<template>
  <div class="cron-unit-form">
    <el-radio-group v-model="type" :disabled="disabled">
      <div v-if="meta.allowUnset" class="cron-unit-form__item">
        <el-radio :value="TypeEnum.unset">
          不指定
        </el-radio>
        <span class="cron-unit-form__tip">与「日/周」另一项配合使用</span>
      </div>
      <div class="cron-unit-form__item">
        <el-radio :value="TypeEnum.every">
          每{{ meta.label }}
        </el-radio>
      </div>
      <div class="cron-unit-form__item">
        <el-radio :value="TypeEnum.range">
          区间
        </el-radio>
        <div v-if="type === TypeEnum.range" class="cron-unit-form__config">
          <span>从</span>
          <el-select
            v-if="meta.isWeek"
            v-model="valueRange.start"
            :disabled="disabled"
            style="width: 100px"
          >
            <el-option
              v-for="opt in weekOptions"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
          <el-input-number
            v-else
            v-model="valueRange.start"
            :min="meta.min"
            :max="meta.max"
            controls-position="right"
            :disabled="disabled"
          />
          <span>到</span>
          <el-select
            v-if="meta.isWeek"
            v-model="valueRange.end"
            :disabled="disabled"
            style="width: 100px"
          >
            <el-option
              v-for="opt in weekOptions"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
          <el-input-number
            v-else
            v-model="valueRange.end"
            :min="meta.min"
            :max="meta.max"
            controls-position="right"
            :disabled="disabled"
          />
        </div>
      </div>
      <div class="cron-unit-form__item">
        <el-radio :value="TypeEnum.loop">
          间隔
        </el-radio>
        <div v-if="type === TypeEnum.loop" class="cron-unit-form__config">
          <span>从</span>
          <el-input-number
            v-model="valueLoop.start"
            :min="meta.min"
            :max="meta.max"
            controls-position="right"
            :disabled="disabled"
          />
          <span>开始，每</span>
          <el-input-number
            v-model="valueLoop.interval"
            :min="1"
            :max="meta.max"
            controls-position="right"
            :disabled="disabled"
          />
          <span>{{ meta.label }}</span>
        </div>
      </div>
      <div class="cron-unit-form__item">
        <el-radio :value="TypeEnum.specify">
          指定
        </el-radio>
        <div v-if="type === TypeEnum.specify" class="cron-unit-form__list">
          <el-checkbox-group v-model="valueList" :disabled="disabled">
            <el-checkbox
              v-for="i in specifyRange"
              :key="i"
              :value="i"
              :class="{ 'cron-unit-form__week-item': meta.isWeek }"
            >
              {{ meta.isWeek ? weekOptions.find(w => w.value === i)?.label : i }}
            </el-checkbox>
          </el-checkbox-group>
        </div>
      </div>
    </el-radio-group>
  </div>
</template>

<style scoped lang="scss">
.cron-unit-form {
  &__item {
    display: block;
    width: 100%;
    margin-top: 12px;

    &:first-child {
      margin-top: 4px;
    }
  }

  &__tip {
    margin-left: 8px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  &__config {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
    margin-top: 8px;
    margin-left: 24px;
  }

  &__list {
    margin-top: 8px;
    margin-left: 24px;

    :deep(.el-checkbox-group) {
      display: flex;
      flex-wrap: wrap;
      gap: 4px;
    }

    :deep(.el-checkbox) {
      width: 48px;
      margin-right: 0;
    }
  }

  &__week-item {
    width: 60px !important;
  }

  :deep(.el-radio-group) {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    width: 100%;
  }
}
</style>

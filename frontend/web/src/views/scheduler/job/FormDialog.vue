<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'
import type { SchedulerJob } from '@/apis/scheduler'
import { Clock } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import {
  BlockStrategyOptions,
  createSchedulerJobApi,
  JobStatusOptions,
  RouteStrategyOptions,
  TaskTypeOptions,
  TriggerTypeOptions,
  updateSchedulerJobApi,
} from '@/apis/scheduler'
import { validateCronExpression } from './cron/cron-utils'
import CronGeneratorDialog from './CronGeneratorDialog.vue'

defineOptions({ name: 'SchedulerJobFormDialog' })

const emit = defineEmits<{ success: [] }>()

interface JobFormData {
  group_name: string
  name: string
  description: string
  trigger_type: string
  cron_expr: string
  execute_mode: string
  handler: string
  params: string
  route_strategy: string
  block_strategy: string
  timeout_sec: number
  retry_count: number
  retry_interval: number
  parallel_count: number
  status: string
}

const colProps = { xs: 24, sm: 24, md: 12, lg: 12, xl: 12 }
const colThirdProps = { xs: 24, sm: 12, md: 8, lg: 8, xl: 8 }

const visible = ref(false)
const isEdit = ref(false)
const currentId = ref<number>()
const formRef = ref<FormInstance>()
const cronGeneratorRef = useTemplateRef<InstanceType<typeof CronGeneratorDialog>>('cronGeneratorRef')
const formData = ref<JobFormData>(createEmptyForm())
const dialogTitle = computed(() => (isEdit.value ? '编辑任务' : '新增任务'))

function createEmptyForm(): JobFormData {
  return {
    group_name: 'default',
    name: '',
    description: '',
    trigger_type: 'cron',
    cron_expr: '0 */5 * * * *',
    execute_mode: 'cluster',
    handler: 'stat.sync',
    params: '{}',
    route_strategy: 'round_robin',
    block_strategy: 'parallel',
    timeout_sec: 300,
    retry_count: 0,
    retry_interval: 60,
    parallel_count: 1,
    status: '0',
  }
}

const formRules: FormRules = {
  group_name: [{ required: true, message: '请输入任务组', trigger: 'blur' }],
  name: [
    { required: true, message: '请输入任务名称', trigger: 'blur' },
    { max: 128, message: '任务名称最多 128 个字符', trigger: 'blur' },
  ],
  trigger_type: [{ required: true, message: '请选择触发类型', trigger: 'change' }],
  cron_expr: [
    {
      validator: (_rule, value, callback) => {
        if (formData.value.trigger_type !== 'cron') {
          callback()
          return
        }
        const err = validateCronExpression(String(value ?? ''))
        callback(err ? new Error(err) : undefined)
      },
      trigger: 'blur',
    },
  ],
  execute_mode: [{ required: true, message: '请选择任务类型', trigger: 'change' }],
  handler: [{ required: true, message: '请输入执行器名称', trigger: 'blur' }],
  route_strategy: [{ required: true, message: '请选择路由策略', trigger: 'change' }],
  block_strategy: [{ required: true, message: '请选择阻塞策略', trigger: 'change' }],
  timeout_sec: [{ required: true, message: '请输入超时时间', trigger: 'blur' }],
  retry_count: [{ required: true, message: '请输入最大重试次数', trigger: 'blur' }],
  retry_interval: [{ required: true, message: '请输入重试间隔', trigger: 'blur' }],
  parallel_count: [{ required: true, message: '请输入并行数', trigger: 'blur' }],
}

function openAdd() {
  isEdit.value = false
  currentId.value = undefined
  formData.value = createEmptyForm()
  visible.value = true
}

function openEdit(row: SchedulerJob) {
  isEdit.value = true
  currentId.value = row.id
  formData.value = {
    group_name: row.group_name ?? 'default',
    name: row.name ?? '',
    description: row.description ?? '',
    trigger_type: row.trigger_type ?? 'cron',
    cron_expr: row.cron_expr ?? '0 */5 * * * *',
    execute_mode: row.execute_mode ?? 'cluster',
    handler: row.handler ?? '',
    params: row.params ?? '{}',
    route_strategy: row.route_strategy ?? 'round_robin',
    block_strategy: row.block_strategy ?? 'parallel',
    timeout_sec: row.timeout_sec ?? 300,
    retry_count: row.retry_count ?? 0,
    retry_interval: row.retry_interval ?? 60,
    parallel_count: row.parallel_count ?? 1,
    status: row.status ?? '0',
  }
  visible.value = true
}

function openCronGenerator() {
  cronGeneratorRef.value?.open(formData.value.cron_expr)
}

function onCronGenerated(expr: string) {
  formData.value.cron_expr = expr
}

async function handleBeforeOk() {
  try {
    await formRef.value?.validate()
    const payload = { ...formData.value }
    if (isEdit.value && currentId.value) {
      await updateSchedulerJobApi(currentId.value, payload)
      ElMessage.success('更新成功')
    }
    else {
      await createSchedulerJobApi(payload)
      ElMessage.success('添加成功')
    }
    emit('success')
    return true
  }
  catch {
    return false
  }
}

defineExpose({ openAdd, openEdit })
</script>

<template>
  <GiDialog
    v-model="visible"
    :title="dialogTitle"
    width="calc(100% - 20px)"
    :style="{ maxWidth: '840px' }"
    body-class="job-form-dialog-body"
    destroy-on-close
    :on-before-ok="handleBeforeOk"
  >
    <el-form
      ref="formRef"
      class="job-form-dialog"
      :model="formData"
      :rules="formRules"
      label-width="110px"
    >
      <fieldset class="job-form-section">
        <legend>基础配置</legend>
        <el-row :gutter="16">
          <el-col v-bind="colProps">
            <el-form-item label="任务组" prop="group_name">
              <el-input
                v-model="formData.group_name"
                placeholder="如 default"
                maxlength="64"
                show-word-limit
              />
            </el-form-item>
          </el-col>
          <el-col v-bind="colProps">
            <el-form-item label="任务名称" prop="name">
              <el-input
                v-model="formData.name"
                placeholder="请输入任务名称"
                maxlength="128"
                show-word-limit
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="formData.description"
            type="textarea"
            placeholder="请输入任务描述"
            maxlength="512"
            show-word-limit
            :rows="3"
          />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="formData.status">
            <el-radio
              v-for="opt in JobStatusOptions"
              :key="opt.value"
              :value="opt.value"
            >
              {{ opt.label }}
            </el-radio>
          </el-radio-group>
        </el-form-item>
      </fieldset>

      <fieldset class="job-form-section">
        <legend>调度配置</legend>
        <el-row :gutter="16">
          <el-col v-bind="colProps">
            <el-form-item label="触发类型" prop="trigger_type">
              <el-select v-model="formData.trigger_type" placeholder="请选择触发类型" style="width: 100%">
                <el-option
                  v-for="opt in TriggerTypeOptions"
                  :key="opt.value"
                  :label="opt.label"
                  :value="opt.value"
                  :disabled="opt.disabled"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col v-bind="colProps">
            <el-form-item label="Cron 表达式" prop="cron_expr">
              <div class="job-form-cron-row">
                <el-input
                  v-model="formData.cron_expr"
                  class="job-form-cron-input"
                  placeholder="秒 分 时 日 月 周"
                  clearable
                />
                <el-tooltip content="Cron 表达式生成器">
                  <el-button :icon="Clock" @click="openCronGenerator" />
                </el-tooltip>
              </div>
            </el-form-item>
          </el-col>
        </el-row>
      </fieldset>

      <fieldset class="job-form-section">
        <legend>任务配置</legend>
        <el-row :gutter="16">
          <el-col v-bind="colProps">
            <el-form-item label="任务类型" prop="execute_mode">
              <el-select v-model="formData.execute_mode" placeholder="请选择任务类型" style="width: 100%">
                <el-option
                  v-for="opt in TaskTypeOptions"
                  :key="opt.value"
                  :label="opt.label"
                  :value="opt.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col v-bind="colProps">
            <el-form-item label="执行器名称" prop="handler">
              <el-input
                v-model="formData.handler"
                placeholder="如 stat.sync"
                maxlength="64"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="任务参数" prop="params">
          <el-input
            v-model="formData.params"
            type="textarea"
            placeholder="JSON 格式，如 {}"
            :rows="3"
          />
        </el-form-item>
      </fieldset>

      <fieldset class="job-form-section job-form-section--last">
        <legend>高级配置</legend>
        <el-row :gutter="16">
          <el-col v-bind="colProps">
            <el-form-item label="路由策略" prop="route_strategy">
              <el-select v-model="formData.route_strategy" placeholder="请选择路由策略" style="width: 100%">
                <el-option
                  v-for="opt in RouteStrategyOptions"
                  :key="opt.value"
                  :label="opt.label"
                  :value="opt.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col v-bind="colProps">
            <el-form-item label="阻塞策略" prop="block_strategy">
              <el-select v-model="formData.block_strategy" placeholder="请选择阻塞策略" style="width: 100%">
                <el-option
                  v-for="opt in BlockStrategyOptions"
                  :key="opt.value"
                  :label="opt.label"
                  :value="opt.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16" class="job-form-advanced-grid">
          <el-col v-bind="colThirdProps">
            <el-form-item label="超时时间" prop="timeout_sec">
              <div class="job-form-number-field">
                <el-input-number
                  v-model="formData.timeout_sec"
                  :min="10"
                  :max="86400"
                  controls-position="right"
                />
                <span class="job-form-unit">秒</span>
              </div>
            </el-form-item>
          </el-col>
          <el-col v-bind="colThirdProps">
            <el-form-item label="最大重试" prop="retry_count">
              <el-input-number
                v-model="formData.retry_count"
                :min="0"
                :max="99"
                controls-position="right"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col v-bind="colThirdProps">
            <el-form-item label="重试间隔" prop="retry_interval">
              <div class="job-form-number-field">
                <el-input-number
                  v-model="formData.retry_interval"
                  :min="1"
                  :max="86400"
                  controls-position="right"
                />
                <span class="job-form-unit">秒</span>
              </div>
            </el-form-item>
          </el-col>
          <el-col v-bind="colThirdProps">
            <el-form-item label="并行数" prop="parallel_count">
              <el-input-number
                v-model="formData.parallel_count"
                :min="1"
                :max="99"
                controls-position="right"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>
      </fieldset>
    </el-form>

    <CronGeneratorDialog ref="cronGeneratorRef" @ok="onCronGenerated" />
  </GiDialog>
</template>

<style scoped lang="scss">
:global(.job-form-dialog-body) {
  max-height: min(68vh, 640px);
  overflow-y: auto;
  padding-top: 4px;
  padding-bottom: 4px;
}

.job-form-dialog {
  :deep(.el-form-item) {
    margin-bottom: 18px;
  }

  :deep(.el-form-item:last-child) {
    margin-bottom: 0;
  }
}

.job-form-section {
  min-width: 0;
  padding: 8px 16px 4px;
  margin: 0 0 16px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: var(--el-border-radius-base);
  background-color: var(--el-fill-color-lighter);

  &--last {
    margin-bottom: 0;
  }

  legend {
    padding: 0 10px;
    margin-left: 4px;
    font-size: 13px;
    font-weight: 500;
    line-height: 1.4;
    color: var(--el-text-color-primary);
    background-color: var(--el-bg-color);
    border: 1px solid var(--el-border-color-lighter);
    border-radius: var(--el-border-radius-base);
  }
}

.job-form-cron-row {
  display: flex;
  gap: 8px;
  align-items: center;
  width: 100%;
}

.job-form-cron-input {
  flex: 1;
  min-width: 0;
}

.job-form-advanced-grid {
  margin-top: 2px;
}

.job-form-number-field {
  display: flex;
  gap: 8px;
  align-items: center;
  width: 100%;

  :deep(.el-input-number) {
    flex: 1;
    min-width: 0;
    width: auto;
  }
}

.job-form-unit {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>

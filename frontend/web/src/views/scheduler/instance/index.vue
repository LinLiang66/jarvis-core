<script setup lang="ts">
import type { FormColumnItem, TableColumnItem } from 'gi-component'
import type { SchedulerInstance, SchedulerLog } from '@/apis/scheduler'
import { ElMessage } from 'element-plus'
import {
  getSchedulerInstanceListApi,
  getSchedulerInstanceLogsApi,
  InstanceStatusOptions,
} from '@/apis/scheduler'
import { useTable } from '@/hooks/useTable'
import { formatDateTime } from '@/utils/format'

defineOptions({ name: 'SchedulerInstance' })

const logVisible = ref(false)
const currentInstance = ref<SchedulerInstance | null>(null)
const logData = ref<SchedulerLog[]>([])
const logLoading = ref(false)

const queryForm = reactive({
  job_id: undefined as number | undefined,
  handler: '',
  status: undefined as string | undefined,
})

const formColumns: FormColumnItem[] = [
  {
    field: 'job_id',
    label: '任务ID',
    type: 'input-number',
    props: { min: 1, controlsPosition: 'right', clearable: true },
  },
  { field: 'handler', label: 'Handler', type: 'input', props: { clearable: true } },
  {
    field: 'status',
    label: '状态',
    type: 'select-v2',
    props: { options: InstanceStatusOptions, clearable: true },
  },
]

const statusLabel = (v: string) => InstanceStatusOptions.find(o => o.value === v)?.label ?? v
const dateTimeFormatter = (_row: unknown, _column: unknown, value: unknown) => formatDateTime(value as string)

const tableColumns: TableColumnItem[] = [
  { prop: 'id', label: 'ID', width: 80 },
  { prop: 'job_id', label: '任务ID', width: 80 },
  { prop: 'job_name', label: '任务名称', minWidth: 120 },
  { prop: 'handler', label: 'Handler', minWidth: 110 },
  { prop: 'trigger_type', label: '触发方式', width: 90 },
  { prop: 'status', label: '状态', width: 90, render: ({ row }) => statusLabel(row.status) },
  { prop: 'worker_id', label: 'Worker', minWidth: 120, showOverflowTooltip: true },
  { prop: 'started_at', label: '开始时间', minWidth: 170, formatter: dateTimeFormatter },
  { prop: 'finished_at', label: '结束时间', minWidth: 170, formatter: dateTimeFormatter },
  {
    prop: 'action',
    label: '操作',
    width: 100,
    align: 'center',
    fixed: 'right',
    slotName: 'action',
  },
]

const {
  tableData,
  loading,
  pagination,
  search,
  refresh,
} = useTable({
  rowKey: 'id',
  listAPI: p => getSchedulerInstanceListApi({
    ...p,
    job_id: queryForm.job_id,
    handler: queryForm.handler || undefined,
    status: queryForm.status,
  }),
})

function handleReset() {
  queryForm.job_id = undefined
  queryForm.handler = ''
  queryForm.status = undefined
  search()
}

async function handleViewLogs(row: SchedulerInstance) {
  currentInstance.value = row
  logVisible.value = true
  logLoading.value = true
  try {
    const res = await getSchedulerInstanceLogsApi(row.id, { page: 1, size: 200 })
    logData.value = res.list ?? []
  }
  catch {
    ElMessage.error('加载日志失败')
  }
  finally {
    logLoading.value = false
  }
}

const logColumns: TableColumnItem[] = [
  { prop: 'created_at', label: '时间', minWidth: 170, formatter: dateTimeFormatter },
  { prop: 'level', label: '级别', width: 80 },
  { prop: 'message', label: '内容', minWidth: 280, showOverflowTooltip: true },
]
</script>

<template>
  <GiPageLayout class="g-page-layout">
    <template #header>
      <GiForm
        :model-value="queryForm"
        :columns="formColumns"
        search
        :grid-item-props="{ span: { xs: 24, sm: 12, md: 12, lg: 8, xl: 6, xxl: 6 } }"
        @update:model-value="Object.assign(queryForm, $event)"
        @search="search"
        @reset="handleReset"
      />
    </template>

    <GiTable
      v-loading="loading"
      border
      row-key="id"
      :data="tableData"
      :columns="tableColumns"
      :pagination="pagination"
      @refresh="refresh"
    >
      <template #action="{ row }">
        <el-button type="primary" link @click="handleViewLogs(row)">
          日志
        </el-button>
      </template>
    </GiTable>

    <GiDialog
      v-model="logVisible"
      :title="`执行日志 #${currentInstance?.id ?? ''}`"
      width="720px"
      ok-text="关闭"
      :show-cancel="false"
    >
      <GiTable
        v-loading="logLoading"
        border
        row-key="id"
        :data="logData"
        :columns="logColumns"
        :pagination="false"
      />
      <el-descriptions v-if="currentInstance" class="mt-3" :column="1" border size="small">
        <el-descriptions-item v-if="currentInstance.result" label="结果">
          {{ currentInstance.result }}
        </el-descriptions-item>
        <el-descriptions-item v-if="currentInstance.error_msg" label="错误">
          {{ currentInstance.error_msg }}
        </el-descriptions-item>
      </el-descriptions>
    </GiDialog>
  </GiPageLayout>
</template>

<style scoped>
.mt-3 {
  margin-top: 12px;
}
</style>

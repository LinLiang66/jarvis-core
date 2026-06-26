<script setup lang="ts">
import type { FormColumnItem, TableColumnItem } from 'gi-component'
import type { SchedulerJob } from '@/apis/scheduler'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  BlockStrategyOptions,
  deleteSchedulerJobApi,
  getSchedulerJobListApi,
  JobStatusOptions,
  triggerSchedulerJobApi,
} from '@/apis/scheduler'
import { useTable } from '@/hooks/useTable'
import { formatDateTime } from '@/utils/format'
import FormDialog from './FormDialog.vue'

defineOptions({ name: 'SchedulerJob' })

const FormDialogRef = useTemplateRef('FormDialogRef')

const queryForm = reactive({
  name: '',
  handler: '',
  status: undefined as string | undefined,
})

const formColumns: FormColumnItem[] = [
  { field: 'name', label: '任务名称', type: 'input', props: { clearable: true } },
  { field: 'handler', label: 'Handler', type: 'input', props: { clearable: true } },
  {
    field: 'status',
    label: '状态',
    type: 'select-v2',
    props: { options: JobStatusOptions, clearable: true },
  },
]

const statusLabel = (v: string) => JobStatusOptions.find(o => o.value === v)?.label ?? v
const blockLabel = (v: string) => BlockStrategyOptions.find(o => o.value === v)?.label ?? v
const dateTimeFormatter = (_row: unknown, _column: unknown, value: unknown) => formatDateTime(value as string)

const tableColumns: TableColumnItem[] = [
  { prop: 'id', label: 'ID', width: 70 },
  { prop: 'group_name', label: '任务组', width: 100 },
  { prop: 'name', label: '任务名称', minWidth: 140 },
  { prop: 'handler', label: '执行器', minWidth: 120 },
  { prop: 'cron_expr', label: 'Cron', minWidth: 140 },
  { prop: 'block_strategy', label: '阻塞策略', width: 90, render: ({ row }) => blockLabel(row.block_strategy) },
  { prop: 'status', label: '状态', width: 80, render: ({ row }) => statusLabel(row.status) },
  { prop: 'description', label: '描述', minWidth: 120, showOverflowTooltip: true },
  { prop: 'created_at', label: '创建时间', minWidth: 170, formatter: dateTimeFormatter },
  { prop: 'updated_at', label: '更新时间', minWidth: 170, formatter: dateTimeFormatter },
  {
    prop: 'action',
    label: '操作',
    width: 220,
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
  onDelete,
} = useTable({
  rowKey: 'id',
  listAPI: p => getSchedulerJobListApi({
    ...p,
    name: queryForm.name || undefined,
    handler: queryForm.handler || undefined,
    status: queryForm.status,
  }),
  deleteAPI: ids => Promise.all(ids.map(id => deleteSchedulerJobApi(Number(id)))),
})

function handleReset() {
  queryForm.name = ''
  queryForm.handler = ''
  queryForm.status = undefined
  search()
}

function handleAdd() {
  FormDialogRef.value?.openAdd()
}

function handleEdit(row: SchedulerJob) {
  FormDialogRef.value?.openEdit(row)
}

async function handleTrigger(row: SchedulerJob) {
  try {
    await ElMessageBox.confirm(`确定立即执行任务「${row.name}」？`, '手动触发', { type: 'warning' })
    await triggerSchedulerJobApi(row.id)
    ElMessage.success('已触发执行')
    refresh()
  }
  catch {
    // cancelled
  }
}
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

    <template #tool>
      <el-button type="primary" @click="handleAdd">
        新增任务
      </el-button>
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
        <el-button type="primary" link @click="handleTrigger(row)">
          触发
        </el-button>
        <el-button type="primary" link @click="handleEdit(row)">
          编辑
        </el-button>
        <el-button type="danger" link @click="onDelete(row)">
          删除
        </el-button>
      </template>
    </GiTable>

    <FormDialog ref="FormDialogRef" @success="refresh" />
  </GiPageLayout>
</template>

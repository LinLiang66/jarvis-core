<script setup lang="ts">
import type { FormColumnItem, TableColumnItem } from 'gi-component'
import type { TabPaneName } from 'element-plus'
import { getOpenCallLogsApi, getOpenDailyStatApi } from '@/apis/openplatform'
import { useTable } from '@/hooks/useTable'

defineOptions({ name: 'OpenPlatformStat' })

const activeTab = ref<'daily' | 'logs'>('daily')

/** 日期时间范围：选日期时默认 00:00:00 ~ 23:59:59 */
const datetimeRangeDefaultTime: [Date, Date] = [
  new Date(2000, 0, 1, 0, 0, 0),
  new Date(2000, 0, 1, 23, 59, 59),
]

const searchGridSpan = { xs: 24, sm: 12, md: 8, lg: 6, xl: 5, xxl: 5 }
const rangeGridSpan = { xs: 24, sm: 24, md: 12, lg: 9, xl: 8, xxl: 7 }

const queryForm = reactive({
  app_id: '',
  action: '',
  stat_date_range: [] as string[],
  date_range: [] as string[],
})

const formColumns: FormColumnItem[] = [
  {
    field: 'app_id',
    label: 'AppID',
    type: 'input',
    gridItemProps: { span: searchGridSpan },
    props: { clearable: true, placeholder: '请输入 AppID' },
  },
  {
    field: 'action',
    label: 'Action',
    type: 'input',
    gridItemProps: { span: searchGridSpan },
    props: { clearable: true, placeholder: '请输入 Action' },
  },
  {
    field: 'stat_date_range',
    label: '统计日期',
    type: 'date-picker',
    hide: () => activeTab.value !== 'daily',
    gridItemProps: { span: rangeGridSpan },
    props: {
      type: 'daterange',
      rangeSeparator: '至',
      startPlaceholder: '开始日期',
      endPlaceholder: '结束日期',
      valueFormat: 'YYYY-MM-DD',
      format: 'YYYY-MM-DD',
      clearable: true,
      style: { width: '100%', maxWidth: '360px' },
    },
  },
  {
    field: 'date_range',
    label: '调用时间',
    type: 'date-picker',
    hide: () => activeTab.value !== 'logs',
    gridItemProps: { span: rangeGridSpan },
    props: {
      type: 'datetimerange',
      rangeSeparator: '至',
      startPlaceholder: '开始时间',
      endPlaceholder: '结束时间',
      valueFormat: 'YYYY-MM-DD HH:mm:ss',
      format: 'YYYY-MM-DD HH:mm:ss',
      defaultTime: datetimeRangeDefaultTime,
      clearable: true,
      style: { width: '100%', maxWidth: '420px' },
    },
  },
]

const dailyColumns: TableColumnItem[] = [
  { prop: 'app_id', label: 'AppID', minWidth: 140 },
  { prop: 'action', label: 'Action', minWidth: 200 },
  { prop: 'stat_date', label: '日期', width: 120 },
  { prop: 'total_count', label: '总次数', width: 100 },
  { prop: 'success_count', label: '成功', width: 90 },
  { prop: 'fail_count', label: '失败', width: 90 },
]

const logColumns: TableColumnItem[] = [
  { prop: 'app_id', label: 'AppID', minWidth: 140 },
  { prop: 'action', label: 'Action', minWidth: 180 },
  { prop: 'success', label: '结果', width: 80, render: ({ row }) => (row.success ? '成功' : '失败') },
  { prop: 'duration_ms', label: '耗时(ms)', width: 100 },
  { prop: 'client_ip', label: '客户端 IP', width: 130 },
  { prop: 'message', label: '消息', minWidth: 160, showOverflowTooltip: true },
  { prop: 'created_at', label: '时间', minWidth: 170 },
]

const {
  tableData: dailyData,
  loading: dailyLoading,
  pagination: dailyPagination,
  search: dailySearch,
  refresh: dailyRefresh,
} = useTable({
  rowKey: 'id',
  listAPI: p => getOpenDailyStatApi({
    ...p,
    app_id: queryForm.app_id || undefined,
    action: queryForm.action || undefined,
    date_from: queryForm.stat_date_range[0] || undefined,
    date_to: queryForm.stat_date_range[1] || undefined,
  }),
})

const {
  tableData: logData,
  loading: logLoading,
  pagination: logPagination,
  search: logSearch,
  refresh: logRefresh,
} = useTable({
  rowKey: 'id',
  immediate: false,
  listAPI: p => getOpenCallLogsApi({
    ...p,
    app_id: queryForm.app_id || undefined,
    action: queryForm.action || undefined,
    date_from: queryForm.date_range[0] || undefined,
    date_to: queryForm.date_range[1] || undefined,
  }),
})

function handleSearch() {
  if (activeTab.value === 'daily')
    dailySearch()
  else
    logSearch()
}

function handleReset() {
  queryForm.app_id = ''
  queryForm.action = ''
  queryForm.stat_date_range = []
  queryForm.date_range = []
  handleSearch()
}

function handleTabChange(name: TabPaneName) {
  if (name === 'logs')
    logSearch()
  else
    dailySearch()
}
</script>

<template>
  <GiPageLayout class="g-page-layout">
    <template #header>
      <el-tabs v-model="activeTab" class="stat-tabs" @tab-change="handleTabChange">
        <el-tab-pane label="日统计" name="daily" />
        <el-tab-pane label="调用明细" name="logs" />
      </el-tabs>
      <GiForm
        :key="activeTab"
        class="stat-search-form"
        :model-value="queryForm"
        :columns="formColumns"
        search
        :grid-item-props="{ span: searchGridSpan }"
        @update:model-value="Object.assign(queryForm, $event)"
        @search="handleSearch"
        @reset="handleReset"
      />
    </template>

    <GiTable
      v-show="activeTab === 'daily'"
      v-loading="dailyLoading"
      border
      row-key="id"
      :columns="dailyColumns"
      :data="dailyData"
      :pagination="dailyPagination"
      @refresh="dailyRefresh"
    />
    <GiTable
      v-show="activeTab === 'logs'"
      v-loading="logLoading"
      border
      row-key="id"
      :columns="logColumns"
      :data="logData"
      :pagination="logPagination"
      @refresh="logRefresh"
    />
  </GiPageLayout>
</template>

<style scoped>
.stat-tabs {
  margin-bottom: 8px;
}

.stat-search-form :deep(.el-input),
.stat-search-form :deep(.el-date-editor) {
  max-width: 240px;
}

.stat-search-form :deep(.el-date-editor--daterange),
.stat-search-form :deep(.el-date-editor--datetimerange) {
  max-width: 420px;
}
</style>

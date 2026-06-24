<script setup lang="ts">
import type { FormColumnItem, TableColumnItem } from 'gi-component'
import type { OpenAppItem } from '@/apis/openplatform'
import { ElMessage, ElMessageBox } from 'element-plus'
import { deleteOpenAppApi, getOpenAppListApi, regenerateOpenAppKeysApi } from '@/apis/openplatform'
import { getDictLabel, useDict } from '@/hooks/useDict'
import { useTable } from '@/hooks/useTable'
import FormDialog from './FormDialog.vue'

defineOptions({ name: 'OpenPlatformApp' })

const FormDialogRef = useTemplateRef('FormDialogRef')
const { dictData } = useDict(['STATUS'] as const)

const queryForm = reactive({
  app_id: '',
  app_name: '',
  status: undefined as string | undefined,
})

const formColumns = computed<FormColumnItem[]>(() => [
  { field: 'app_id', label: 'AppID', type: 'input' },
  { field: 'app_name', label: '应用名称', type: 'input' },
  {
    field: 'status',
    label: '状态',
    type: 'select-v2',
    props: {
      options: [
        { label: '正常', value: '0' },
        { label: '禁用', value: '1' },
      ],
      clearable: true,
    },
  },
])

const tableColumns: TableColumnItem[] = [
  { type: 'selection', width: 48, align: 'center' },
  { prop: 'id', label: 'ID', width: 80 },
  { prop: 'app_id', label: 'AppID', minWidth: 160 },
  { prop: 'app_name', label: '应用名称', minWidth: 140 },
  { prop: 'total_quota', label: '可用配额', width: 100 },
  { prop: 'total_calls', label: '累计调用', width: 100 },
  { prop: 'status', label: '状态', width: 80, render: ({ row }) => getDictLabel(dictData.value.STATUS, row.status) },
  { prop: 'remark', label: '备注', minWidth: 120, showOverflowTooltip: true },
  {
    prop: 'action',
    label: '操作',
    width: 200,
    align: 'center',
    fixed: 'right',
    slotName: 'action',
  },
]

const {
  tableData,
  loading,
  pagination,
  selectedKeys,
  search,
  refresh,
  onDelete,
  onBatchDelete,
  onSelectionChange,
} = useTable({
  rowKey: 'id',
  listAPI: p => getOpenAppListApi({
    ...p,
    app_id: queryForm.app_id || undefined,
    app_name: queryForm.app_name || undefined,
    status: queryForm.status,
  }),
  deleteAPI: ids => deleteOpenAppApi(ids),
})

function handleSearch() {
  search()
}

function handleReset() {
  queryForm.app_id = ''
  queryForm.app_name = ''
  queryForm.status = undefined
  search()
}

function handleAdd() {
  FormDialogRef.value?.openAdd()
}

function handleEdit(row: OpenAppItem) {
  FormDialogRef.value?.openEdit(row)
}

async function handleRegenerate(row: OpenAppItem) {
  try {
    await ElMessageBox.confirm(
      `确定重新生成应用「${row.app_name}」的密钥？旧密钥将立即失效。`,
      '提示',
      { type: 'warning' },
    )
    const res = await regenerateOpenAppKeysApi(String(row.id))
    FormDialogRef.value?.showSecrets(res)
    ElMessage.success('密钥已重新生成，请妥善保存')
    refresh()
  }
  catch {
    // cancelled or failed
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
        @search="handleSearch"
        @reset="handleReset"
      />
    </template>

    <template #tool>
      <el-space>
        <gi-button v-hasPerm="['openapp:edit']" type="add" @click="handleAdd">
          新增应用
        </gi-button>
        <el-button
          v-hasPerm="['openapp:edit']"
          type="danger"
          :disabled="!selectedKeys.length"
          @click="onBatchDelete"
        >
          批量删除
        </el-button>
      </el-space>
    </template>

    <GiTable
      v-loading="loading"
      border
      :data="tableData"
      :columns="tableColumns"
      row-key="id"
      :pagination="pagination"
      @selection-change="onSelectionChange"
    >
      <template #action="{ row }">
        <el-button v-hasPerm="['openapp:edit']" type="primary" link @click="handleEdit(row)">
          编辑
        </el-button>
        <el-button v-hasPerm="['openapp:edit']" type="warning" link @click="handleRegenerate(row)">
          重置密钥
        </el-button>
        <el-button v-hasPerm="['openapp:edit']" type="danger" link @click="onDelete(row)">
          删除
        </el-button>
      </template>
    </GiTable>

    <FormDialog ref="FormDialogRef" @success="refresh" />
  </GiPageLayout>
</template>

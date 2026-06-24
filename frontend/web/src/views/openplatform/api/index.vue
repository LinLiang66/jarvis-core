<script setup lang="ts">
import type { FormColumnItem, TableColumnItem } from 'gi-component'
import type { OpenAPIActionItem } from '@/apis/openplatform'
import { ElMessage } from 'element-plus'
import {
  deleteOpenAPIActionApi,
  getOpenAPIActionListApi,
  syncOpenAPIActionRegistryApi,
} from '@/apis/openplatform'
import { useTable } from '@/hooks/useTable'
import DocDialog from './DocDialog.vue'
import FormDialog from './FormDialog.vue'

defineOptions({ name: 'OpenPlatformApi' })

const FormDialogRef = useTemplateRef('FormDialogRef')
const DocDialogRef = useTemplateRef('DocDialogRef')
const syncing = ref(false)

const queryForm = reactive({
  action: '',
  title: '',
  category: '',
  status: undefined as string | undefined,
})

const formColumns = computed<FormColumnItem[]>(() => [
  { field: 'action', label: 'Action', type: 'input' },
  { field: 'title', label: '接口名称', type: 'input' },
  { field: 'category', label: '分类', type: 'input' },
  {
    field: 'status',
    label: '状态',
    type: 'select-v2',
    props: {
      options: [
        { label: '启用', value: '0' },
        { label: '禁用', value: '1' },
      ],
      clearable: true,
    },
  },
])

const tableColumns: TableColumnItem[] = [
  { type: 'selection', width: 48, align: 'center' },
  { prop: 'action', label: 'Action', minWidth: 220, showOverflowTooltip: true },
  { prop: 'title', label: '接口名称', minWidth: 140 },
  { prop: 'category', label: '分类', width: 110 },
  {
    prop: 'encrypted',
    label: '3DES',
    width: 70,
    render: ({ row }) => (row.encrypted ? '是' : '否'),
  },
  {
    prop: 'billable',
    label: '计费',
    width: 70,
    render: ({ row }) => (row.billable ? '是' : '否'),
  },
  {
    prop: 'source',
    label: '来源',
    width: 80,
    render: ({ row }) => (row.source === 'code' ? '代码' : '手动'),
  },
  {
    prop: 'status',
    label: '状态',
    width: 80,
    render: ({ row }) => (row.status === '0' ? '启用' : '禁用'),
  },
  { prop: 'sort', label: '排序', width: 70 },
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
  selectedKeys,
  search,
  refresh,
  onDelete,
  onBatchDelete,
  onSelectionChange,
} = useTable({
  rowKey: 'id',
  listAPI: p => getOpenAPIActionListApi({
    ...p,
    action: queryForm.action || undefined,
    title: queryForm.title || undefined,
    category: queryForm.category || undefined,
    status: queryForm.status,
  }),
  deleteAPI: ids => deleteOpenAPIActionApi(ids),
})

function handleSearch() {
  search()
}

function handleReset() {
  queryForm.action = ''
  queryForm.title = ''
  queryForm.category = ''
  queryForm.status = undefined
  search()
}

function handleAdd() {
  FormDialogRef.value?.openAdd()
}

function handleEdit(row: OpenAPIActionItem) {
  FormDialogRef.value?.openEdit(row)
}

function handleViewDoc(row: OpenAPIActionItem) {
  DocDialogRef.value?.open(row)
}

async function handleSync() {
  syncing.value = true
  try {
    const res = await syncOpenAPIActionRegistryApi()
    ElMessage.success(`已同步 ${res.synced} 个接口，文档已自动生成`)
    refresh()
  }
  finally {
    syncing.value = false
  }
}

function openDocs() {
  window.open(`${window.location.origin}/openplatform/docs`, '_blank')
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
        <el-button @click="openDocs">
          接口文档
        </el-button>
        <gi-button v-hasPerm="['openaction:edit']" type="add" @click="handleAdd">
          新增接口
        </gi-button>
        <el-button
          v-hasPerm="['openaction:edit']"
          type="primary"
          :loading="syncing"
          @click="handleSync"
        >
          从代码同步
        </el-button>
        <el-button
          v-hasPerm="['openaction:edit']"
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
        <el-button type="primary" link @click="handleViewDoc(row)">
          文档
        </el-button>
        <el-button v-hasPerm="['openaction:edit']" type="primary" link @click="handleEdit(row)">
          编辑
        </el-button>
        <el-button
          v-if="row.source !== 'code'"
          v-hasPerm="['openaction:edit']"
          type="danger"
          link
          @click="onDelete(row)"
        >
          删除
        </el-button>
      </template>
    </GiTable>

    <FormDialog ref="FormDialogRef" @success="refresh" />
    <DocDialog ref="DocDialogRef" />
  </GiPageLayout>
</template>

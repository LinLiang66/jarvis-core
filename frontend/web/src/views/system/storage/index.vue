<script setup lang="ts">
import type { TableColumnItem } from 'gi-component'
import type { StorageItem, StorageType, StatusValue } from '@/apis/storage'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  deleteStorageApi,
  getStorageListApi,
  setDefaultStorageApi,
  updateStorageStatusApi,
} from '@/apis/storage'
import { useDict } from '@/hooks/useDict'
import FormDialog from './FormDialog.vue'

defineOptions({ name: 'SystemStorage' })

const FormDialogRef = useTemplateRef('FormDialogRef')
const { dictData } = useDict(['STATUS'] as const)

const activeTab = ref<StorageType>(1)
const loading = ref(false)
const tableData = ref<StorageItem[]>([])

const storageTypeLabel: Record<StorageType, string> = {
  1: '本地存储',
  2: '对象存储',
}

const tableColumns = computed<TableColumnItem[]>(() => {
  const cols: TableColumnItem[] = [
    { prop: 'name', label: '名称', minWidth: 140 },
    { prop: 'code', label: '编码', minWidth: 120 },
    { prop: 'bucketName', label: activeTab.value === 1 ? '存储路径' : 'Bucket', minWidth: 180, showOverflowTooltip: true },
  ]
  if (activeTab.value === 2) {
    cols.push(
      { prop: 'endpoint', label: 'Endpoint', minWidth: 180, showOverflowTooltip: true },
      { prop: 'baseUrl', label: 'Base URL', minWidth: 180, showOverflowTooltip: true },
    )
  }
  else {
    cols.push({ prop: 'domain', label: '访问路径', minWidth: 220, showOverflowTooltip: true })
  }
  cols.push(
    { prop: 'isDefault', label: '默认', width: 80, align: 'center', slotName: 'isDefault' },
    { prop: 'status', label: '状态', width: 100, align: 'center', slotName: 'status' },
    { prop: 'sort', label: '排序', width: 80, align: 'center' },
    { prop: 'action', label: '操作', width: 220, align: 'center', fixed: 'right', slotName: 'action' },
  )
  return cols
})

async function loadData() {
  loading.value = true
  try {
    tableData.value = await getStorageListApi({ type: activeTab.value })
  }
  finally {
    loading.value = false
  }
}

function handleTabChange() {
  loadData()
}

function handleAdd() {
  FormDialogRef.value?.openAdd(activeTab.value)
}

function handleEdit(row: StorageItem) {
  FormDialogRef.value?.openEdit(row)
}

async function handleDelete(row: StorageItem) {
  await ElMessageBox.confirm(`确定删除存储「${row.name}」吗？`, '提示', { type: 'warning' })
  await deleteStorageApi([row.id])
  ElMessage.success('删除成功')
  loadData()
}

async function handleSetDefault(row: StorageItem) {
  await setDefaultStorageApi(row.id)
  ElMessage.success('已设为默认存储')
  loadData()
}

async function handleStatusSwitch(row: StorageItem, val: string | number | boolean) {
  const status = val as StatusValue
  try {
    await updateStorageStatusApi(row.id, status)
    ElMessage.success(status === '0' ? '已启用' : '已禁用')
  }
  catch {
    loadData()
  }
}

onMounted(loadData)
</script>

<template>
  <gi-page-layout class="g-page-layout">
    <template #header>
      <div class="flex items-center justify-between gap-4">
        <el-tabs v-model="activeTab" @tab-change="handleTabChange">
          <el-tab-pane label="本地存储" :name="1" />
          <el-tab-pane label="对象存储" :name="2" />
        </el-tabs>
        <el-button type="primary" @click="handleAdd">
          新增{{ storageTypeLabel[activeTab] }}
        </el-button>
      </div>
    </template>

    <gi-table
      :columns="tableColumns"
      :data="tableData"
      :loading="loading"
      row-key="id"
    >
      <template #isDefault="{ row }">
        <el-tag v-if="row.isDefault" type="success" size="small">
          默认
        </el-tag>
        <span v-else>-</span>
      </template>
      <template #status="{ row }">
        <el-switch
          :model-value="row.status"
          active-value="0"
          inactive-value="1"
          :disabled="row.isDefault"
          @change="val => handleStatusSwitch(row, val)"
        />
      </template>
      <template #action="{ row }">
        <el-button link type="primary" @click="handleEdit(row)">
          编辑
        </el-button>
        <el-button
          v-if="!row.isDefault && row.status === '0'"
          link
          type="primary"
          @click="handleSetDefault(row)"
        >
          设为默认
        </el-button>
        <el-button
          v-if="!row.isDefault"
          link
          type="danger"
          @click="handleDelete(row)"
        >
          删除
        </el-button>
      </template>
    </gi-table>

    <FormDialog ref="FormDialogRef" @success="loadData" />
  </gi-page-layout>
</template>

<style scoped>
.flex {
  display: flex;
}
.items-center {
  align-items: center;
}
.justify-between {
  justify-content: space-between;
}
.gap-4 {
  gap: 16px;
}
</style>

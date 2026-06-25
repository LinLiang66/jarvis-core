<script setup lang="ts">
import type { FormColumnItem, TableColumnItem } from 'gi-component'
import type { FileItem } from '@/apis/file'
import type { StorageItem } from '@/apis/storage'
import { ElMessage } from 'element-plus'
import {
  createFileDirApi,
  deleteFileApi,
  getFileListApi,
  getFileStatisticsApi,
  uploadFileApi,
} from '@/apis/file'
import { getStorageListApi } from '@/apis/storage'
import { useTable } from '@/hooks/useTable'

defineOptions({ name: 'SystemFile' })

const stats = ref({ fileCount: 0, dirCount: 0, totalSize: 0 })
const storageOptions = ref<StorageItem[]>([])
const parentPath = ref('/')
const dirName = ref('')
const dirDialogVisible = ref(false)
const uploadInputRef = useTemplateRef<HTMLInputElement>('uploadInputRef')

const queryForm = reactive({
  storageId: '',
  originalName: '',
})

const formColumns = computed<FormColumnItem[]>(() => [
  {
    field: 'storageId',
    label: '存储',
    type: 'select-v2',
    props: {
      options: storageOptions.value.map(item => ({ label: item.name, value: item.id })),
      clearable: true,
      placeholder: '默认存储',
    },
  },
  { field: 'originalName', label: '文件名', type: 'input' },
])

const tableColumns: TableColumnItem[] = [
  { prop: 'originalName', label: '名称', minWidth: 200, slotName: 'name' },
  { prop: 'type', label: '类型', width: 90, align: 'center', slotName: 'type' },
  { prop: 'size', label: '大小', width: 120, slotName: 'size' },
  { prop: 'url', label: '访问地址', minWidth: 240, showOverflowTooltip: true },
  { prop: 'createTime', label: '创建时间', width: 180 },
  { prop: 'action', label: '操作', width: 160, align: 'center', fixed: 'right', slotName: 'action' },
]

const pathSegments = computed(() => {
  const parts = parentPath.value.split('/').filter(Boolean)
  const items = [{ label: '根目录', path: '/' }]
  let current = ''
  for (const part of parts) {
    current += `/${part}`
    items.push({ label: part, path: current })
  }
  return items
})

const {
  tableData,
  loading,
  pagination,
  search,
  refresh,
  onDelete,
} = useTable({
  rowKey: 'id',
  immediate: false,
  listAPI: p => getFileListApi({
    ...p,
    storageId: queryForm.storageId || undefined,
    parentPath: parentPath.value,
    originalName: queryForm.originalName || undefined,
  }),
  deleteAPI: deleteFileApi,
})

function formatSize(size?: number) {
  if (!size)
    return '-'
  if (size < 1024)
    return `${size} B`
  if (size < 1024 * 1024)
    return `${(size / 1024).toFixed(1)} KB`
  return `${(size / 1024 / 1024).toFixed(2)} MB`
}

async function loadMeta() {
  const [storages, stat] = await Promise.all([
    getStorageListApi(),
    getFileStatisticsApi(),
  ])
  storageOptions.value = storages
  stats.value = stat
}

function handleSearch() {
  search()
}

function handleReset() {
  queryForm.storageId = ''
  queryForm.originalName = ''
  search()
}

function navigateTo(path: string) {
  parentPath.value = path
  search()
}

function openDir(row: FileItem) {
  navigateTo(row.path)
}

function openCreateDir() {
  if (parentPath.value === '/') {
    ElMessage.warning('请进入子目录后再创建文件夹')
    return
  }
  dirName.value = ''
  dirDialogVisible.value = true
}

async function submitCreateDir() {
  if (!dirName.value.trim()) {
    ElMessage.warning('请输入文件夹名称')
    return false
  }
  await createFileDirApi({
    parentPath: parentPath.value,
    originalName: dirName.value.trim(),
    storageId: queryForm.storageId || undefined,
  })
  ElMessage.success('创建成功')
  dirDialogVisible.value = false
  refresh()
  loadMeta()
  return true
}

function triggerUpload() {
  uploadInputRef.value?.click()
}

async function handleFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file)
    return
  const formData = new FormData()
  formData.append('file', file)
  formData.append('parentPath', parentPath.value === '/' ? '' : parentPath.value)
  if (queryForm.storageId)
    formData.append('storageId', queryForm.storageId)
  await uploadFileApi(formData)
  ElMessage.success('上传成功')
  input.value = ''
  refresh()
  loadMeta()
}

async function handleDelete(row: FileItem) {
  await onDelete(row)
  loadMeta()
}

function copyUrl(url?: string) {
  if (!url)
    return
  navigator.clipboard.writeText(url)
  ElMessage.success('已复制链接')
}

onMounted(async () => {
  await loadMeta()
  search()
})
</script>

<template>
  <gi-page-layout class="g-page-layout">
    <template #header>
      <div class="file-header">
        <div class="file-stats">
          <span>文件 {{ stats.fileCount }}</span>
          <span>文件夹 {{ stats.dirCount }}</span>
          <span>总大小 {{ formatSize(stats.totalSize) }}</span>
        </div>
        <el-breadcrumb separator="/">
          <el-breadcrumb-item
            v-for="item in pathSegments"
            :key="item.path"
          >
            <a href="javascript:;" @click="navigateTo(item.path)">{{ item.label }}</a>
          </el-breadcrumb-item>
        </el-breadcrumb>
        <gi-form
          :model-value="queryForm"
          :columns="formColumns"
          search
          @search="handleSearch"
          @reset="handleReset"
        />
      </div>
    </template>

    <template #toolbar>
      <el-button type="primary" @click="triggerUpload">
        上传文件
      </el-button>
      <el-button @click="openCreateDir">
        新建文件夹
      </el-button>
      <input ref="uploadInputRef" type="file" hidden @change="handleFileChange">
    </template>

    <gi-table
      :columns="tableColumns"
      :data="tableData"
      :loading="loading"
      :pagination="pagination"
      row-key="id"
    >
      <template #name="{ row }">
        <el-button
          v-if="row.type === 0"
          link
          type="primary"
          @click="openDir(row)"
        >
          {{ row.originalName }}
        </el-button>
        <span v-else>{{ row.originalName }}</span>
      </template>
      <template #type="{ row }">
        {{ row.type === 0 ? '文件夹' : '文件' }}
      </template>
      <template #size="{ row }">
        {{ row.type === 0 ? '-' : formatSize(row.size) }}
      </template>
      <template #action="{ row }">
        <el-button
          v-if="row.url"
          link
          type="primary"
          @click="copyUrl(row.url)"
        >
          复制链接
        </el-button>
        <el-button link type="danger" @click="handleDelete(row)">
          删除
        </el-button>
      </template>
    </gi-table>

    <gi-dialog
      v-model="dirDialogVisible"
      title="新建文件夹"
      width="420px"
      :on-before-ok="submitCreateDir"
    >
      <el-input v-model="dirName" placeholder="请输入文件夹名称" maxlength="100" />
    </gi-dialog>
  </gi-page-layout>
</template>

<style scoped>
.file-header {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.file-stats {
  display: flex;
  gap: 16px;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
</style>

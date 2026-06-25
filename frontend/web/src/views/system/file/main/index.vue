<script setup lang="ts">
import type { UploadRequestOptions } from 'element-plus'
import type { FileItem, FileListQuery } from '@/apis/file'
import { Icon } from '@iconify/vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createFileDirApi,
  deleteFileApi,
  getFileListApi,
  uploadFileApi,
} from '@/apis/file'
import { useTable } from '@/hooks/useTable'
import { ImageTypes, OfficeTypes, isDirectory, normalizeParentPath, todayStoragePath } from '../constants'
import FileGrid from './FileGrid.vue'
import FileList from './FileList.vue'
import useFileManage from './useFileManage'

defineOptions({ name: 'FileMain' })

const emit = defineEmits<{ 'stats-refresh': [] }>()

const route = useRoute()

const category = computed(() => {
  const type = route.query.type
  if (!type || type === '0')
    return 0
  return Number(type)
})

const queryForm = reactive<FileListQuery>({
  originalName: '',
  parentPath: category.value ? undefined : '/',
  category: category.value || undefined,
})

const isBatchMode = ref(false)
const createDirVisible = ref(false)
const newDirName = ref('')
const uploading = ref(false)
const previewVisible = ref(false)
const previewList = ref<string[]>([])
const previewIndex = ref(0)

const contextMenu = reactive({
  visible: false,
  x: 0,
  y: 0,
  item: null as FileItem | null,
})

const {
  mode,
  selectedFileIds,
  toggleMode,
  addSelectedFileItem,
  clearSelection,
} = useFileManage()

const {
  tableData: fileList,
  loading,
  pagination,
  search,
  refresh,
} = useTable({
  rowKey: 'id',
  immediate: false,
  listAPI: page => getFileListApi({
    ...queryForm,
    ...page,
    category: category.value || undefined,
    parentPath: category.value ? undefined : (queryForm.parentPath || '/'),
  }),
  deleteAPI: deleteFileApi,
})

pagination.pageSize = 30

const breadcrumbList = computed(() => {
  const path = queryForm.parentPath || '/'
  return path.split('/').filter(Boolean).map((part, index, arr) => ({
    name: part,
    path: `/${arr.slice(0, index + 1).join('/')}`,
  }))
})

watch(() => route.query.type, (type) => {
  const val = type ? Number(type) : 0
  queryForm.category = val || undefined
  if (!val) {
    queryForm.parentPath = '/'
  }
  else {
    queryForm.parentPath = undefined
  }
  search()
})

function handleSearch() {
  search()
}

function handleBreadcrumbClick(path: string) {
  queryForm.parentPath = path
  search()
}

function isDir(item: FileItem) {
  return isDirectory(item)
}

/** 与 qn-ai-web 一致：用 parentPath + name 拼目录路径，避免 path 字段异常导致下钻失败 */
function buildDirPath(item: FileItem) {
  const name = item.name || item.originalName
  if (name) {
    const parent = item.parentPath === '/' ? '' : (item.parentPath || '')
    return normalizeParentPath(`${parent}/${name}`)
  }
  return normalizeParentPath(item.path || '/')
}

function enterDirectory(item: FileItem) {
  if (!isDir(item))
    return
  queryForm.parentPath = buildDirPath(item)
  search()
}

function handleDblclick(item: FileItem) {
  enterDirectory(item)
}

function handleClickFile(item: FileItem) {
  if (isDir(item)) {
    enterDirectory(item)
    return
  }
  const ext = item.extension?.toLowerCase()
  if (ext && ImageTypes.includes(ext) && item.url) {
    const images = fileList.value
      .filter(i => !isDirectory(i) && i.url && ImageTypes.includes(i.extension?.toLowerCase() || ''))
      .map(i => i.url!)
    previewList.value = images
    previewIndex.value = Math.max(0, images.indexOf(item.url))
    previewVisible.value = true
    return
  }
  if (ext && OfficeTypes.includes(ext) && item.url) {
    window.open(item.url, '_blank')
    return
  }
  if (ext === 'mp4' && item.url) {
    window.open(item.url, '_blank')
    return
  }
  if (ext === 'mp3' && item.url) {
    window.open(item.url, '_blank')
  }
}

function handleSelectFile(item: FileItem) {
  addSelectedFileItem(item)
}

async function handleDelete(item: FileItem) {
  const label = isDirectory(item) ? '文件夹' : '文件'
  await ElMessageBox.confirm(
    `是否确定删除${label}「${item.originalName}」？`,
    '提示',
    { type: 'warning' },
  )
  await deleteFileApi([item.id])
  ElMessage.success('删除成功')
  await refresh()
  emit('stats-refresh')
}

async function handleBatchDelete() {
  if (!selectedFileIds.value.length)
    return
  await ElMessageBox.confirm(
    `是否确定删除所选的 ${selectedFileIds.value.length} 项？`,
    '提示',
    { type: 'warning' },
  )
  await deleteFileApi(selectedFileIds.value)
  ElMessage.success('删除成功')
  isBatchMode.value = false
  clearSelection()
  await refresh()
  emit('stats-refresh')
}

function openContextMenu(item: FileItem, event: MouseEvent) {
  contextMenu.item = item
  contextMenu.x = event.clientX
  contextMenu.y = event.clientY
  contextMenu.visible = true
}

function closeContextMenu() {
  contextMenu.visible = false
  contextMenu.item = null
}

async function onContextAction(action: 'copy' | 'download' | 'delete') {
  const item = contextMenu.item
  closeContextMenu()
  if (!item)
    return
  if (action === 'copy' && item.url) {
    await navigator.clipboard.writeText(item.url)
    ElMessage.success('已复制链接')
  }
  else if (action === 'download' && item.url) {
    window.open(item.url, '_blank')
  }
  else if (action === 'delete') {
    await handleDelete(item)
  }
}

async function handleUpload(options: UploadRequestOptions) {
  uploading.value = true
  const uploadParent = queryForm.parentPath || '/'
  try {
    const formData = new FormData()
    formData.append('file', options.file)
    formData.append('parentPath', uploadParent)
    const res = await uploadFileApi(formData)
    options.onSuccess?.(res)
    ElMessage.success('上传成功')
    if (uploadParent === '/' && res.parentPath)
      queryForm.parentPath = res.parentPath
    else if (uploadParent === '/')
      queryForm.parentPath = todayStoragePath()
    await refresh()
    emit('stats-refresh')
  }
  catch (err) {
    const message = err instanceof Error ? err.message : '上传失败'
    ElMessage.error(message)
    options.onError?.(err as Error)
  }
  finally {
    uploading.value = false
  }
}

async function submitCreateDir() {
  if (!newDirName.value.trim()) {
    ElMessage.warning('请输入文件夹名称')
    return false
  }
  await createFileDirApi({
    parentPath: queryForm.parentPath || '/',
    originalName: newDirName.value.trim(),
  })
  ElMessage.success('创建成功')
  createDirVisible.value = false
  newDirName.value = ''
  await refresh()
  emit('stats-refresh')
  return true
}

onMounted(() => {
  search()
  document.addEventListener('click', closeContextMenu)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', closeContextMenu)
})
</script>

<template>
  <div class="file-main">
    <el-breadcrumb class="file-main__breadcrumb">
      <el-breadcrumb-item v-if="!category" @click="handleBreadcrumbClick('/')">
        <a href="javascript:;">根目录</a>
      </el-breadcrumb-item>
      <el-breadcrumb-item v-else>
        全部类型
      </el-breadcrumb-item>
      <el-breadcrumb-item
        v-for="item in breadcrumbList"
        :key="item.path"
        @click="handleBreadcrumbClick(item.path)"
      >
        <a href="javascript:;">{{ item.name }}</a>
      </el-breadcrumb-item>
    </el-breadcrumb>

    <div class="file-main__toolbar">
      <el-space wrap>
        <el-upload
          :show-file-list="false"
          :http-request="handleUpload"
          :disabled="uploading || !!category"
          multiple
        >
          <el-button type="primary" :loading="uploading" :disabled="!!category">
            <Icon icon="icon-park-outline:upload" class="btn-icon" />
            上传文件
          </el-button>
        </el-upload>
        <el-input
          v-model="queryForm.originalName"
          clearable
          :placeholder="category ? '搜索名称' : '在当前目录下搜索名称'"
          style="width: 220px"
          @keyup.enter="handleSearch"
        />
        <el-button type="primary" @click="handleSearch">
          <Icon icon="icon-park-outline:search" class="btn-icon" />
          查询
        </el-button>
      </el-space>

      <el-space wrap>
        <el-button
          v-if="isBatchMode"
          type="danger"
          :disabled="!selectedFileIds.length"
          @click="handleBatchDelete"
        >
          <Icon icon="icon-park-outline:delete" class="btn-icon" />
          删除选中
        </el-button>
        <el-button
          type="primary"
          :disabled="!!category"
          @click="createDirVisible = true"
        >
          <Icon icon="icon-park-outline:folder-plus" class="btn-icon" />
          新建文件夹
        </el-button>
        <el-button @click="isBatchMode = !isBatchMode; clearSelection()">
          <Icon icon="icon-park-outline:check-correct" class="btn-icon" />
          {{ isBatchMode ? '取消批量' : '批量操作' }}
        </el-button>
        <el-button @click="toggleMode">
          <Icon
            :icon="mode === 'grid' ? 'icon-park-outline:list' : 'icon-park-outline:all-application'"
            class="btn-icon"
          />
        </el-button>
      </el-space>
    </div>

    <div v-loading="loading" class="file-main__body">
      <FileGrid
        v-if="mode === 'grid' && fileList.length"
        :data="fileList"
        :is-batch-mode="isBatchMode"
        :selected-file-ids="selectedFileIds"
        @click="handleClickFile"
        @dblclick="handleDblclick"
        @select="handleSelectFile"
        @contextmenu="openContextMenu"
      />
      <FileList
        v-else-if="mode === 'list' && fileList.length"
        :data="fileList"
        :is-batch-mode="isBatchMode"
        :selected-file-ids="selectedFileIds"
        @click="handleClickFile"
        @dblclick="handleDblclick"
        @select="handleSelectFile"
        @contextmenu="openContextMenu"
      />
      <el-empty v-if="!fileList.length && !loading" description="暂无文件" />
    </div>

    <div class="file-main__pagination">
      <el-pagination
        v-model:current-page="pagination.currentPage"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[30, 40, 50, 100]"
        layout="total, sizes, prev, pager, next"
        @current-change="pagination.onCurrentChange"
        @size-change="pagination.onSizeChange"
      />
    </div>

    <gi-dialog
      v-model="createDirVisible"
      title="新建文件夹"
      width="420px"
      :on-before-ok="submitCreateDir"
    >
      <el-input v-model="newDirName" placeholder="请输入文件夹名称" maxlength="100" />
    </gi-dialog>

    <el-image-viewer
      v-if="previewVisible"
      :url-list="previewList"
      :initial-index="previewIndex"
      @close="previewVisible = false"
    />

    <div
      v-show="contextMenu.visible"
      class="context-menu"
      :style="{ left: `${contextMenu.x}px`, top: `${contextMenu.y}px` }"
      @click.stop
    >
      <button v-if="contextMenu.item?.url" type="button" @click="onContextAction('copy')">
        复制链接
      </button>
      <button
        v-if="contextMenu.item?.url && !isDirectory(contextMenu.item)"
        type="button"
        @click="onContextAction('download')"
      >
        下载
      </button>
      <button type="button" class="danger" @click="onContextAction('delete')">
        删除
      </button>
    </div>
  </div>
</template>

<style scoped>
.file-main {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--el-bg-color);
  border-radius: 8px;
  overflow: hidden;
}
.file-main__breadcrumb {
  padding: 10px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.file-main__breadcrumb a {
  color: inherit;
  text-decoration: none;
}
.file-main__breadcrumb a:hover {
  color: var(--el-color-primary);
}
.file-main__toolbar {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  padding: 16px 16px 0;
}
.btn-icon {
  margin-right: 4px;
}
.file-main__body {
  flex: 1;
  overflow: auto;
  padding: 0 16px 16px;
}
.file-main__pagination {
  padding: 0 16px 16px;
  display: flex;
  justify-content: flex-end;
}
.context-menu {
  position: fixed;
  z-index: 3000;
  min-width: 120px;
  background: var(--el-bg-color-overlay);
  border: 1px solid var(--el-border-color);
  border-radius: 6px;
  box-shadow: var(--el-box-shadow-light);
  padding: 4px 0;
}
.context-menu button {
  display: block;
  width: 100%;
  border: none;
  background: none;
  text-align: left;
  padding: 8px 14px;
  cursor: pointer;
  color: var(--el-text-color-primary);
}
.context-menu button:hover {
  background: var(--el-fill-color-light);
}
.context-menu button.danger {
  color: var(--el-color-danger);
}
</style>

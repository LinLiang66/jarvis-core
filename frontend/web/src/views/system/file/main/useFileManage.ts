import type { FileItem } from '@/apis/file'

type Mode = 'grid' | 'list'

export default function useFileManage() {
  const mode = ref<Mode>('grid')
  const selectedFileList = ref<FileItem[]>([])
  const selectedFileIds = computed(() => selectedFileList.value.map(i => i.id))

  function toggleMode() {
    mode.value = mode.value === 'grid' ? 'list' : 'grid'
  }

  function addSelectedFileItem(item: FileItem) {
    const index = selectedFileList.value.findIndex(i => i.id === item.id)
    if (index >= 0)
      selectedFileList.value.splice(index, 1)
    else
      selectedFileList.value.push(item)
  }

  function clearSelection() {
    selectedFileList.value = []
  }

  return {
    mode,
    selectedFileList,
    selectedFileIds,
    toggleMode,
    addSelectedFileItem,
    clearSelection,
  }
}

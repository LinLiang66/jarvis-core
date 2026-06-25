import type { FileItem } from '@/apis/file'

export interface FileTypeListItem {
  name: string
  value: number
  icon: string
}

/** 文件分类（与 admin FileTypeEnum 对齐） */
export const FileTypeList: FileTypeListItem[] = [
  { name: '全部', value: 0, icon: 'icon-park-outline:folder-open' },
  { name: '图片', value: 2, icon: 'icon-park-outline:pic' },
  { name: '文档', value: 3, icon: 'icon-park-outline:file-text' },
  { name: '视频', value: 4, icon: 'icon-park-outline:video' },
  { name: '音频', value: 5, icon: 'icon-park-outline:music' },
  { name: '其他', value: 1, icon: 'icon-park-outline:file-code' },
]

export const FileIcon: Record<string, string> = {
  mp3: 'icon-park-outline:music',
  mp4: 'icon-park-outline:video',
  dir: 'icon-park-outline:folder-close',
  ppt: 'icon-park-outline:ppt',
  doc: 'icon-park-outline:word',
  docx: 'icon-park-outline:word',
  xls: 'icon-park-outline:excel',
  xlsx: 'icon-park-outline:excel',
  txt: 'icon-park-outline:file-text',
  pdf: 'icon-park-outline:file-pdf',
  rar: 'icon-park-outline:file-zip',
  zip: 'icon-park-outline:file-zip',
  html: 'icon-park-outline:html-five',
  css: 'icon-park-outline:file-code',
  js: 'icon-park-outline:file-code',
  other: 'icon-park-outline:file-code',
}

export const ImageTypes = ['jpg', 'png', 'gif', 'jpeg', 'webp', 'bmp']

export const OfficeTypes = ['ppt', 'pptx', 'doc', 'docx', 'xls', 'xlsx', 'pdf']

export function formatFileSize(size?: number | null) {
  if (size == null || size < 0)
    return '-'
  if (size < 1024)
    return `${size} B`
  if (size < 1024 * 1024)
    return `${(size / 1024).toFixed(1)} KB`
  if (size < 1024 * 1024 * 1024)
    return `${(size / 1024 / 1024).toFixed(2)} MB`
  return `${(size / 1024 / 1024 / 1024).toFixed(2)} GB`
}

export function todayStoragePath() {
  const d = new Date()
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `/${y}/${m}/${day}`
}

export function normalizeParentPath(path?: string | null) {
  let p = (path ?? '/').trim()
  if (!p)
    p = '/'
  if (!p.startsWith('/'))
    p = `/${p}`
  if (p.length > 1)
    p = p.replace(/\/+$/, '')
  return p
}

/** 0=文件夹 1=文件，与后端 model.FileTypeDir/FileTypeFile 一致；兼容历史 type 错误与字符串 */
export function isDirectory(typeOrItem: number | string | undefined | null | Pick<FileItem, 'type' | 'url' | 'extension' | 'size'>) {
  if (typeOrItem != null && typeof typeOrItem === 'object') {
    const item = typeOrItem
    if (Number(item.type) === 0)
      return true
    return !item.url && !item.extension && (item.size == null || item.size === 0)
  }
  return Number(typeOrItem) === 0
}

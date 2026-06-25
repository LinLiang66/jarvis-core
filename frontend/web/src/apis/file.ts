import { request } from './request'

export interface FileItem {
  id: string
  storageId: string
  name: string
  originalName: string
  path: string
  parentPath: string
  url?: string
  size?: number
  extension?: string
  contentType?: string
  type: number
  createTime?: string
}

export interface FileListQuery extends PageParams {
  storageId?: string
  parentPath?: string
  originalName?: string
  type?: number
}

export interface FileStatistics {
  fileCount: number
  dirCount: number
  totalSize: number
}

export function getFileListApi(params: FileListQuery) {
  return request<PageResult<FileItem>>({ url: '/file/list', method: 'get', params })
}

export function getFileStatisticsApi() {
  return request<FileStatistics>({ url: '/file/statistics', method: 'get' })
}

export function uploadFileApi(data: FormData) {
  return request<{ id: string, url: string }>({
    url: '/file/upload',
    method: 'post',
    data,
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export function createFileDirApi(data: { parentPath: string, originalName: string, storageId?: string }) {
  return request<FileItem>({ url: '/file/dir', method: 'post', data })
}

export function deleteFileApi(ids: string[]) {
  return request({ url: '/file/delete', method: 'post', data: { ids } })
}

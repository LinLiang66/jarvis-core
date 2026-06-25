import { request } from './request'

export type StorageType = 1 | 2
export type StatusValue = '0' | '1'

export interface StorageItem {
  id: string
  name: string
  code: string
  type: StorageType
  accessKey?: string
  secretKey?: string
  endpoint?: string
  bucketName: string
  baseUrl?: string
  domain?: string
  description?: string
  isDefault: boolean
  sort?: number
  status: StatusValue
  createTime?: string
  updateTime?: string
}

export interface StorageListQuery {
  type?: StorageType
}

export function getStorageListApi(params?: StorageListQuery) {
  return request<StorageItem[]>({ url: '/storage/list', method: 'get', params })
}

export function getStorageDetailApi(id: string) {
  return request<StorageItem>({ url: `/storage/${id}`, method: 'get' })
}

export function createStorageApi(data: Partial<StorageItem>) {
  return request<StorageItem>({ url: '/storage', method: 'post', data })
}

export function updateStorageApi(id: string, data: Partial<StorageItem>) {
  return request<StorageItem>({ url: `/storage/${id}`, method: 'put', data })
}

export function deleteStorageApi(ids: string[]) {
  return request({ url: '/storage/delete', method: 'post', data: { ids } })
}

export function updateStorageStatusApi(id: string, status: StatusValue) {
  return request<StorageItem>({ url: `/storage/${id}/status`, method: 'put', data: { status } })
}

export function setDefaultStorageApi(id: string) {
  return request<StorageItem>({ url: `/storage/${id}/default`, method: 'put' })
}

import { request } from './request'

export interface OpenAppItem {
  id: number
  app_id: string
  app_name: string
  rsa_public_key?: string
  status: string
  total_quota: number
  total_calls: number
  remark?: string
  created_at?: string
  updated_at?: string
}

export interface OpenAppCreateResult extends OpenAppItem {
  app_secret: string
  sign_secret: string
}

export interface OpenAppListQuery extends PageParams {
  app_id?: string
  app_name?: string
  status?: string
}

export interface OpenDailyStatItem {
  id: number
  app_id: string
  action: string
  stat_date: string
  total_count: number
  success_count: number
  fail_count: number
}

export interface OpenCallLogItem {
  id: number
  app_id: string
  action: string
  success: boolean
  duration_ms: number
  message?: string
  client_ip?: string
  created_at?: string
}

export interface OpenStatQuery extends PageParams {
  app_id?: string
  action?: string
  stat_date?: string
  date_from?: string
  date_to?: string
}

export function getOpenAppListApi(params: OpenAppListQuery) {
  return request<OpenAppItem[]>({ url: '/open-app/list', method: 'get', params })
}

export function getOpenAppDetailApi(id: string) {
  return request<OpenAppItem>({ url: `/open-app/${id}`, method: 'get' })
}

export function createOpenAppApi(data: { app_name: string, total_quota?: number, remark?: string }) {
  return request<OpenAppCreateResult>({ url: '/open-app', method: 'post', data })
}

export function updateOpenAppApi(id: string, data: Partial<OpenAppItem>) {
  return request({ url: `/open-app/${id}`, method: 'put', data })
}

export function deleteOpenAppApi(ids: string[]) {
  return request({ url: '/open-app/delete', method: 'post', data: { ids } })
}

export function regenerateOpenAppKeysApi(id: string) {
  return request<OpenAppCreateResult>({ url: `/open-app/${id}/regenerate-keys`, method: 'post' })
}

export function getOpenDailyStatApi(params: OpenStatQuery) {
  return request<OpenDailyStatItem[]>({ url: '/open-app/stat/daily', method: 'get', params })
}

export function getOpenCallLogsApi(params: OpenStatQuery) {
  return request<OpenCallLogItem[]>({ url: '/open-app/stat/logs', method: 'get', params })
}

export interface OpenAPIFieldDesc {
  name: string
  type: string
  required: boolean
  example?: unknown
  children?: OpenAPIFieldDesc[]
}

export interface OpenAPIActionItem {
  id: number
  action: string
  title: string
  category: string
  description?: string
  encrypted: boolean
  billable: boolean
  status: string
  request_schema?: string
  response_schema?: string
  request_fields?: string
  response_fields?: string
  doc_markdown?: string
  sort: number
  source: string
  created_at?: string
  updated_at?: string
}

export interface OpenDocCategory {
  name: string
  actions: OpenAPIActionItem[]
}

export function parseOpenAPIFields(raw?: string): OpenAPIFieldDesc[] {
  if (!raw)
    return []
  try {
    return JSON.parse(raw) as OpenAPIFieldDesc[]
  }
  catch {
    return []
  }
}

export function getOpenDocApi() {
  return request<OpenDocCategory[]>({ url: '/open-app/doc', method: 'get' })
}

export function getOpenDocDetailApi(action: string) {
  return request<OpenAPIActionItem>({ url: `/open-app/doc/${encodeURIComponent(action)}`, method: 'get' })
}

export interface OpenAPIActionListQuery extends PageParams {
  action?: string
  title?: string
  category?: string
  status?: string
}

export function getOpenAPIActionListApi(params: OpenAPIActionListQuery) {
  return request<OpenAPIActionItem[]>({ url: '/open-app/action/list', method: 'get', params })
}

export function getOpenAPIActionDetailApi(id: string) {
  return request<OpenAPIActionItem>({ url: `/open-app/action/${id}`, method: 'get' })
}

export function syncOpenAPIActionRegistryApi() {
  return request<{ synced: number }>({ url: '/open-app/action/sync', method: 'post' })
}

export function createOpenAPIActionApi(data: Partial<OpenAPIActionItem>) {
  return request<OpenAPIActionItem>({ url: '/open-app/action', method: 'post', data })
}

export function updateOpenAPIActionApi(id: string, data: Partial<OpenAPIActionItem>) {
  return request<OpenAPIActionItem>({ url: `/open-app/action/${id}`, method: 'put', data })
}

export function deleteOpenAPIActionApi(ids: string[]) {
  return request({ url: '/open-app/action/delete', method: 'post', data: { ids } })
}

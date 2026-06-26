import { request } from './request'

export interface SchedulerJob {
  id: number
  group_name?: string
  name: string
  handler: string
  trigger_type?: string
  cron_expr?: string
  params?: string
  block_strategy: string
  route_strategy: string
  execute_mode: string
  status: string
  description?: string
  timeout_sec: number
  retry_count?: number
  retry_interval?: number
  parallel_count?: number
  created_at?: string
  updated_at?: string
}

export interface SchedulerInstance {
  id: number
  job_id: number
  job_name?: string
  handler: string
  trigger_type: string
  status: string
  worker_id?: string
  params?: string
  result?: string
  error_msg?: string
  started_at?: string
  finished_at?: string
  created_at?: string
  updated_at?: string
}

export interface SchedulerLog {
  id: number
  instance_id: number
  level: string
  message: string
  created_at?: string
}

export interface SchedulerWorker {
  id: number
  worker_id: string
  instance_id?: string
  hostname?: string
  handlers?: string
  status: string
  last_heartbeat_at?: string
}

export interface JobListQuery extends PageParams {
  name?: string
  handler?: string
  status?: string
}

export interface InstanceListQuery extends PageParams {
  job_id?: number
  handler?: string
  status?: string
}

export const BlockStrategyOptions = [
  { label: '并行', value: 'parallel' },
  { label: '串行', value: 'serial' },
  { label: '丢弃', value: 'discard' },
]

export const RouteStrategyOptions = [
  { label: '轮询', value: 'round_robin' },
]

export const TriggerTypeOptions = [
  { label: 'Cron 表达式', value: 'cron' },
  { label: '固定频率', value: 'fixed_rate', disabled: true },
]

export const TaskTypeOptions = [
  { label: '集群任务', value: 'cluster' },
]

export const JobStatusOptions = [
  { label: '启用', value: '0' },
  { label: '停用', value: '1' },
]

export const InstanceStatusOptions = [
  { label: '待执行', value: 'PENDING' },
  { label: '排队中', value: 'QUEUED' },
  { label: '执行中', value: 'RUNNING' },
  { label: '成功', value: 'SUCCESS' },
  { label: '失败', value: 'FAILED' },
  { label: '已丢弃', value: 'DISCARDED' },
]

export function getSchedulerJobListApi(params: JobListQuery) {
  return request<PageResult<SchedulerJob>>({ url: '/scheduler/jobs', method: 'get', params })
}

export function getSchedulerJobDetailApi(id: number) {
  return request<SchedulerJob>({ url: `/scheduler/jobs/${id}`, method: 'get' })
}

export function createSchedulerJobApi(data: Partial<SchedulerJob>) {
  return request<SchedulerJob>({ url: '/scheduler/jobs', method: 'post', data })
}

export function updateSchedulerJobApi(id: number, data: Partial<SchedulerJob>) {
  return request<SchedulerJob>({ url: `/scheduler/jobs/${id}`, method: 'put', data })
}

export function deleteSchedulerJobApi(id: number) {
  return request({ url: `/scheduler/jobs/${id}`, method: 'delete' })
}

export function triggerSchedulerJobApi(id: number) {
  return request<SchedulerInstance>({ url: `/scheduler/jobs/${id}/trigger`, method: 'post' })
}

export function getSchedulerInstanceListApi(params: InstanceListQuery) {
  return request<PageResult<SchedulerInstance>>({ url: '/scheduler/instances', method: 'get', params })
}

export function getSchedulerInstanceDetailApi(id: number) {
  return request<SchedulerInstance>({ url: `/scheduler/instances/${id}`, method: 'get' })
}

export function getSchedulerInstanceLogsApi(id: number, params: PageParams) {
  return request<PageResult<SchedulerLog>>({ url: `/scheduler/instances/${id}/logs`, method: 'get', params })
}

export function getSchedulerWorkerListApi(params: PageParams) {
  return request<PageResult<SchedulerWorker>>({ url: '/scheduler/workers', method: 'get', params })
}

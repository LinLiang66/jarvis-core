export const CRON_FIELD_COUNT = 6

export const CronPresets = [
  { label: '每分钟', value: '0 */1 * * * *' },
  { label: '每 5 分钟', value: '0 */5 * * * *' },
  { label: '每 30 分钟', value: '0 */30 * * * *' },
  { label: '每小时', value: '0 0 * * * *' },
  { label: '每天 0 点', value: '0 0 0 * * *' },
  { label: '每周一 0 点', value: '0 0 0 * * 1' },
  { label: '每月 1 日 0 点', value: '0 0 0 1 * *' },
] as const

export const WEEK_OPTIONS = [
  { label: '周日', value: 0 },
  { label: '周一', value: 1 },
  { label: '周二', value: 2 },
  { label: '周三', value: 3 },
  { label: '周四', value: 4 },
  { label: '周五', value: 5 },
  { label: '周六', value: 6 },
]

export interface CronUnitConfig {
  label: string
  min: number
  max: number
  defaultValue: string
  allowUnset?: boolean
  isWeek?: boolean
}

export type CronUnitKey = 'second' | 'minute' | 'hour' | 'day' | 'month' | 'week'

export const CronUnitMeta: Record<CronUnitKey, CronUnitConfig> = {
  second: { label: '秒', min: 0, max: 59, defaultValue: '0' },
  minute: { label: '分', min: 0, max: 59, defaultValue: '*' },
  hour: { label: '时', min: 0, max: 23, defaultValue: '*' },
  day: { label: '日', min: 1, max: 31, defaultValue: '*', allowUnset: true },
  month: { label: '月', min: 1, max: 12, defaultValue: '*' },
  week: { label: '周', min: 0, max: 6, defaultValue: '?', allowUnset: true, isWeek: true },
}

export function parseCronExpression(expr: string): Record<CronUnitKey, string> {
  const parts = expr.trim().split(/\s+/).filter(Boolean)
  const keys: CronUnitKey[] = ['second', 'minute', 'hour', 'day', 'month', 'week']
  const result = {} as Record<CronUnitKey, string>
  keys.forEach((key, index) => {
    result[key] = parts[index] ?? CronUnitMeta[key].defaultValue
  })
  return result
}

export function buildCronExpression(parts: Record<CronUnitKey, string>): string {
  const keys: CronUnitKey[] = ['second', 'minute', 'hour', 'day', 'month', 'week']
  return keys.map(key => parts[key] || CronUnitMeta[key].defaultValue).join(' ')
}

export function validateCronExpression(expr: string): string | null {
  const parts = expr.trim().split(/\s+/).filter(Boolean)
  if (parts.length !== CRON_FIELD_COUNT) {
    return `Cron 表达式需包含 ${CRON_FIELD_COUNT} 段（秒 分 时 日 月 周）`
  }
  const day = parts[3]
  const week = parts[5]
  if (day === '?' && week === '?') {
    return '「日」与「周」不能同时设为不指定（?）'
  }
  return null
}

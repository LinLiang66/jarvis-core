import dayjs from 'dayjs'

const DATETIME_FORMAT = 'YYYY-MM-DD HH:mm:ss'

/** 将 ISO 或时间字符串格式化为 `YYYY-MM-DD HH:mm:ss` */
export function formatDateTime(value?: string | null): string {
  if (!value)
    return ''
  const d = dayjs(value)
  return d.isValid() ? d.format(DATETIME_FORMAT) : value
}

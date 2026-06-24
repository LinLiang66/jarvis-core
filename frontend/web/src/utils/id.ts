/** 与后端 int64 JSON 字符串化后的主键类型一致 */
export type EntityId = string

/** 将列表行 id、路由参数等统一为 string，供删除/编辑 API 使用 */
export function toId(value: string | number | null | undefined): EntityId {
  if (value === null || value === undefined || value === '')
    return ''
  return String(value)
}

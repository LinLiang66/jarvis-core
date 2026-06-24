<script setup lang="ts">
import type { OpenAPIFieldDesc } from '@/apis/openplatform'

defineOptions({ name: 'OpenPlatformSchemaFieldTable' })

defineProps<{
  fields: OpenAPIFieldDesc[]
}>()
</script>

<template>
  <el-table
    v-if="fields.length"
    :data="fields"
    border
    size="small"
    row-key="name"
    default-expand-all
    :tree-props="{ children: 'children' }"
    class="schema-table"
  >
    <el-table-column prop="name" label="参数名" min-width="160" />
    <el-table-column prop="type" label="类型" width="100" />
    <el-table-column label="必填" width="70" align="center">
      <template #default="{ row }">
        {{ row.required ? '是' : '否' }}
      </template>
    </el-table-column>
    <el-table-column label="示例" min-width="180" show-overflow-tooltip>
      <template #default="{ row }">
        <span v-if="row.example !== undefined && row.example !== null">{{ String(row.example) }}</span>
        <span v-else class="text-muted">-</span>
      </template>
    </el-table-column>
  </el-table>
  <div v-else class="text-muted empty-tip">
    暂无字段描述
  </div>
</template>

<style scoped>
.schema-table {
  width: 100%;
}

.text-muted {
  color: var(--el-text-color-secondary);
}

.empty-tip {
  padding: 12px 0;
  font-size: 13px;
}
</style>

<script setup lang="ts">
import type { FormColumnItem, FormInstance } from 'gi-component'
import type { OpenAPIActionItem } from '@/apis/openplatform'
import { ElMessage } from 'element-plus'
import { createOpenAPIActionApi, updateOpenAPIActionApi } from '@/apis/openplatform'

defineOptions({ name: 'OpenPlatformApiFormDialog' })

const emit = defineEmits<{ success: [] }>()

interface ActionFormData {
  action: string
  title: string
  category: string
  description: string
  encrypted: boolean
  billable: boolean
  status: string
  sort: number
}

const visible = ref(false)
const isEdit = ref(false)
const isCodeSource = ref(false)
const currentId = ref<string>()
const formRef = ref<FormInstance>()
const formData = ref<ActionFormData>(createEmptyForm())
const dialogTitle = computed(() => (isEdit.value ? '编辑接口' : '新增接口'))

function createEmptyForm(): ActionFormData {
  return {
    action: '',
    title: '',
    category: '自定义',
    description: '',
    encrypted: true,
    billable: true,
    status: '0',
    sort: 200,
  }
}

const formColumns = computed<FormColumnItem[]>(() => {
  const cols: FormColumnItem[] = [
    { field: 'action', label: 'Action', type: 'input', props: { disabled: isEdit.value } },
    { field: 'title', label: '接口名称', type: 'input' },
    { field: 'category', label: '分类', type: 'input' },
    { field: 'sort', label: '排序', type: 'input-number', props: { min: 0, controlsPosition: 'right' } },
  ]
  if (!isCodeSource.value) {
    cols.push(
      { field: 'encrypted', label: '3DES加密', type: 'switch' },
      { field: 'billable', label: '计费', type: 'switch' },
    )
  }
  if (isEdit.value) {
    cols.push({
      field: 'status',
      label: '状态',
      type: 'radio-group',
      props: {
        options: [
          { label: '启用', value: '0' },
          { label: '禁用', value: '1' },
        ],
      },
    })
  }
  cols.push(
    { field: 'description', label: '说明', type: 'textarea', span: 24, props: { rows: 3 } },
  )
  return cols
})

const formRules = {
  action: [{ required: true, message: '请输入 Action', trigger: 'blur' }],
  title: [{ required: true, message: '请输入接口名称', trigger: 'blur' }],
}

function openAdd() {
  isEdit.value = false
  isCodeSource.value = false
  currentId.value = undefined
  formData.value = createEmptyForm()
  visible.value = true
}

function openEdit(row: OpenAPIActionItem) {
  isEdit.value = true
  isCodeSource.value = row.source === 'code'
  currentId.value = String(row.id)
  formData.value = {
    action: row.action,
    title: row.title,
    category: row.category,
    description: row.description || '',
    encrypted: row.encrypted,
    billable: row.billable,
    status: row.status,
    sort: row.sort,
  }
  visible.value = true
}

async function handleSubmit() {
  await formRef.value?.validate()
  if (isEdit.value && currentId.value) {
    await updateOpenAPIActionApi(currentId.value, formData.value)
    ElMessage.success('更新成功')
  }
  else {
    await createOpenAPIActionApi(formData.value)
    ElMessage.success('创建成功')
  }
  visible.value = false
  emit('success')
}

defineExpose({ openAdd, openEdit })
</script>

<template>
  <GiDialog v-model="visible" :title="dialogTitle" width="560px" @confirm="handleSubmit">
    <GiForm
      ref="formRef"
      :model-value="formData"
      :columns="formColumns"
      :rules="formRules"
      label-width="100px"
      @update:model-value="Object.assign(formData, $event)"
    />
    <p v-if="isCodeSource" class="form-tip">
      代码注册接口：请求/响应示例与文档由后端自动生成，请点击「从代码同步」更新。
    </p>
  </GiDialog>
</template>

<style scoped>
.form-tip {
  margin: 12px 0 0;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  line-height: 1.5;
}
</style>

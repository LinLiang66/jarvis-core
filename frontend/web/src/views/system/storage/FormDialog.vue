<script setup lang="ts">
import type { FormRules } from 'element-plus'
import type { FormColumnItem, FormInstance } from 'gi-component'
import type { StorageItem, StorageType, StatusValue } from '@/apis/storage'
import { ElMessage } from 'element-plus'
import { createStorageApi, getStorageDetailApi, updateStorageApi } from '@/apis/storage'
import { useDict } from '@/hooks/useDict'
import { inputProps, textareaProps } from '@/utils/formField'

defineOptions({ name: 'SystemStorageFormDialog' })

const emit = defineEmits<{
  (e: 'success'): void
}>()

const { dictData } = useDict(['STATUS'] as const)

interface StorageFormData {
  name: string
  code: string
  type: StorageType
  accessKey: string
  secretKey: string
  endpoint: string
  bucketName: string
  baseUrl: string
  domain: string
  description: string
  status: StatusValue
  sort: number
}

const visible = ref(false)
const isEdit = ref(false)
const currentId = ref('')
const storageType = ref<StorageType>(1)
const formRef = useTemplateRef<FormInstance>('formRef')
const formData = ref<StorageFormData>(createEmptyForm(1))

const dialogTitle = computed(() => {
  const typeLabel = storageType.value === 1 ? '本地存储' : '对象存储'
  return isEdit.value ? `编辑${typeLabel}` : `新增${typeLabel}`
})

function createEmptyForm(type: StorageType): StorageFormData {
  return {
    name: '',
    code: '',
    type,
    accessKey: '',
    secretKey: '',
    endpoint: '',
    bucketName: '',
    baseUrl: '',
    domain: '',
    description: '',
    status: '0',
    sort: 999,
  }
}

const formRules = computed<FormRules>(() => ({
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入编码', trigger: 'blur' }],
  bucketName: [{ required: true, message: '请输入存储路径/Bucket', trigger: 'blur' }],
  domain: formData.value.type === 1
    ? [{ required: true, message: '请输入访问路径', trigger: 'blur' }]
    : [],
  accessKey: formData.value.type === 2
    ? [{ required: true, message: '请输入 Access Key', trigger: 'blur' }]
    : [],
  secretKey: formData.value.type === 2 && !isEdit.value
    ? [{ required: true, message: '请输入 Secret Key', trigger: 'blur' }]
    : [],
  endpoint: formData.value.type === 2
    ? [{ required: true, message: '请输入 Endpoint', trigger: 'blur' }]
    : [],
}))

const formColumns = computed<FormColumnItem[]>(() => {
  const cols: FormColumnItem[] = [
    { field: 'name', label: '名称', type: 'input', props: inputProps(undefined, 100) },
    { field: 'code', label: '编码', type: 'input', props: inputProps({ disabled: isEdit.value }, 64) },
  ]
  if (formData.value.type === 2) {
    cols.push(
      { field: 'accessKey', label: 'Access Key', type: 'input', props: inputProps() },
      { field: 'secretKey', label: 'Secret Key', type: 'input', props: inputProps({ type: 'password', showPassword: true, placeholder: isEdit.value ? '留空则不修改' : '' }, 512) },
      { field: 'endpoint', label: 'Endpoint', type: 'input', props: inputProps({ placeholder: '腾讯云 COS 填 cos.ap-地域.myqcloud.com（勿带 bucket 前缀）' }) },
      { field: 'bucketName', label: 'Bucket', type: 'input', props: inputProps() },
      { field: 'baseUrl', label: 'Base URL', type: 'input', props: inputProps({ placeholder: '内网访问域名，留空则使用 OSS 原始域名' }, 512) },
      { field: 'domain', label: '自定义域名', type: 'input', props: inputProps({ placeholder: '可选，用于构建原始 URL' }, 512) },
    )
  }
  else {
    cols.push(
      { field: 'bucketName', label: '存储路径', type: 'input', props: inputProps({ placeholder: '如 ./data/uploads' }) },
      { field: 'domain', label: '访问路径', type: 'input', props: inputProps({ placeholder: '如 http://127.0.0.1:8000/static/local' }, 512) },
    )
  }
  cols.push(
    { field: 'sort', label: '排序', type: 'input-number', props: { min: 0, controlsPosition: 'right' } },
    { field: 'description', label: '描述', type: 'textarea', span: 24, props: textareaProps({ rows: 3 }) },
    { field: 'status', label: '状态', type: 'radio-group', props: { options: dictData.value.STATUS } },
  )
  return cols
})

async function handleBeforeOk() {
  try {
    await formRef.value?.formRef?.validate()
    const payload = { ...formData.value }
    if (isEdit.value && !payload.secretKey) {
      delete (payload as Partial<StorageFormData>).secretKey
    }
    if (isEdit.value) {
      await updateStorageApi(currentId.value, payload)
      ElMessage.success('修改成功')
    }
    else {
      await createStorageApi(payload)
      ElMessage.success('新增成功')
    }
    emit('success')
    return true
  }
  catch {
    return false
  }
}

function openAdd(type: StorageType) {
  isEdit.value = false
  currentId.value = ''
  storageType.value = type
  formData.value = createEmptyForm(type)
  visible.value = true
}

async function openEdit(row: StorageItem) {
  isEdit.value = true
  currentId.value = row.id
  storageType.value = row.type
  const detail = await getStorageDetailApi(row.id)
  formData.value = {
    name: detail.name,
    code: detail.code,
    type: detail.type,
    accessKey: detail.accessKey || '',
    secretKey: '',
    endpoint: detail.endpoint || '',
    bucketName: detail.bucketName,
    baseUrl: detail.baseUrl || '',
    domain: detail.domain || '',
    description: detail.description || '',
    status: detail.status,
    sort: detail.sort ?? 999,
  }
  visible.value = true
}

defineExpose({ openAdd, openEdit })
</script>

<template>
  <gi-dialog
    v-model="visible"
    :title="dialogTitle"
    width="calc(100% - 20px)"
    :style="{ maxWidth: '600px' }"
    destroy-on-close
    :on-before-ok="handleBeforeOk"
  >
    <gi-form
      ref="formRef"
      v-model="formData"
      :columns="formColumns"
      :rules="formRules"
      label-width="90px"
    />
  </gi-dialog>
</template>

<script setup lang="ts">
import type { FormColumnItem, FormInstance } from 'gi-component'
import type { OpenAppCreateResult, OpenAppItem } from '@/apis/openplatform'
import { ElMessage } from 'element-plus'
import { createOpenAppApi, updateOpenAppApi } from '@/apis/openplatform'

defineOptions({ name: 'OpenPlatformAppFormDialog' })

const emit = defineEmits<{ success: [] }>()

interface AppFormData {
  app_name: string
  total_quota: number
  status: string
  remark: string
}

const visible = ref(false)
const secretVisible = ref(false)
const isEdit = ref(false)
const currentId = ref<string>()
const formRef = ref<FormInstance>()
const formData = ref<AppFormData>(createEmptyForm())
const secretInfo = ref<OpenAppCreateResult | null>(null)
const dialogTitle = computed(() => (isEdit.value ? '编辑应用' : '新增应用'))

function createEmptyForm(): AppFormData {
  return {
    app_name: '',
    total_quota: 0,
    status: '0',
    remark: '',
  }
}

const formColumns = computed<FormColumnItem[]>(() => {
  const cols: FormColumnItem[] = [
    { field: 'app_name', label: '应用名称', type: 'input' },
    {
      field: 'total_quota',
      label: '可用配额',
      type: 'input-number',
      props: { min: 0, controlsPosition: 'right' },
    },
  ]
  if (isEdit.value) {
    cols.push({
      field: 'status',
      label: '状态',
      type: 'radio-group',
      props: {
        options: [
          { label: '正常', value: '0' },
          { label: '禁用', value: '1' },
        ],
      },
    })
  }
  cols.push({
    field: 'remark',
    label: '备注',
    type: 'textarea',
    span: 24,
    props: { rows: 3, maxlength: 500, showWordLimit: true },
  })
  return cols
})

const formRules = {
  app_name: [{ required: true, message: '请输入应用名称', trigger: 'blur' }],
}

function openAdd() {
  isEdit.value = false
  currentId.value = undefined
  formData.value = createEmptyForm()
  visible.value = true
}

function openEdit(row: OpenAppItem) {
  isEdit.value = true
  currentId.value = String(row.id)
  formData.value = {
    app_name: row.app_name ?? '',
    total_quota: row.total_quota ?? 0,
    status: row.status ?? '0',
    remark: row.remark ?? '',
  }
  visible.value = true
}

function showSecrets(data: OpenAppCreateResult) {
  secretInfo.value = data
  secretVisible.value = true
}

async function copyText(text: string, label: string) {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success(`${label} 已复制`)
  }
  catch {
    ElMessage.warning('复制失败，请手动选择复制')
  }
}

async function handleBeforeOk() {
  try {
    await formRef.value?.formRef?.validate()
    if (isEdit.value && currentId.value) {
      await updateOpenAppApi(currentId.value, {
        app_name: formData.value.app_name.trim(),
        status: formData.value.status,
        total_quota: formData.value.total_quota,
        remark: formData.value.remark.trim(),
      })
      ElMessage.success('更新成功')
    }
    else {
      const res = await createOpenAppApi({
        app_name: formData.value.app_name.trim(),
        total_quota: formData.value.total_quota,
        remark: formData.value.remark.trim(),
      })
      secretInfo.value = res
      secretVisible.value = true
      ElMessage.success('创建成功，请保存密钥信息')
    }
    emit('success')
    return true
  }
  catch {
    return false
  }
}

defineExpose({ openAdd, openEdit, showSecrets })
</script>

<template>
  <GiDialog
    v-model="visible"
    :title="dialogTitle"
    width="560px"
    destroy-on-close
    :on-before-ok="handleBeforeOk"
  >
    <GiForm
      ref="formRef"
      v-model="formData"
      :columns="formColumns"
      :rules="formRules"
      label-width="90px"
    />
  </GiDialog>

  <GiDialog
    v-model="secretVisible"
    title="应用密钥（仅展示一次，请妥善保存）"
    width="640px"
    ok-text="我已保存"
  >
    <el-descriptions v-if="secretInfo" :column="1" border>
      <el-descriptions-item label="AppID">
        <div class="secret-row">
          <el-text tag="code" class="break-all">
            {{ secretInfo.app_id }}
          </el-text>
          <el-button link type="primary" @click="copyText(secretInfo.app_id, 'AppID')">
            复制
          </el-button>
        </div>
      </el-descriptions-item>
      <el-descriptions-item label="SignSecret（签名密钥）">
        <div class="secret-row">
          <el-text tag="code" class="break-all">
            {{ secretInfo.sign_secret }}
          </el-text>
          <el-button link type="primary" @click="copyText(secretInfo.sign_secret, 'SignSecret')">
            复制
          </el-button>
        </div>
      </el-descriptions-item>
      <el-descriptions-item label="AppSecret（RSA 私钥 DER Base64）">
        <div class="secret-row">
          <el-text tag="code" class="break-all text-xs">
            {{ secretInfo.app_secret }}
          </el-text>
          <el-button link type="primary" @click="copyText(secretInfo.app_secret, 'AppSecret')">
            复制
          </el-button>
        </div>
      </el-descriptions-item>
    </el-descriptions>
    <el-alert
      class="mt-3"
      type="warning"
      :closable="false"
      title="AppSecret 用于 RSA 密钥交换；SignSecret 用于 yd_md5 请求签名。关闭后将无法再次查看。"
    />
  </GiDialog>
</template>

<style scoped>
.break-all {
  word-break: break-all;
}
.text-xs {
  font-size: 12px;
}
.mt-3 {
  margin-top: 12px;
}
.secret-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  width: 100%;
}
</style>

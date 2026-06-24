<script setup lang="ts">
import type { OpenAPIActionItem } from '@/apis/openplatform'
import { parseOpenAPIFields } from '@/apis/openplatform'
import { DocumentCopy } from '@element-plus/icons-vue'
import { useClipboard } from '@vueuse/core'
import CopyCodeBlock from './CopyCodeBlock.vue'
import SchemaFieldTable from './SchemaFieldTable.vue'

defineOptions({ name: 'OpenPlatformApiDetail' })

const props = defineProps<{
  item: OpenAPIActionItem
}>()

const gatewayPath = '/api/v1/open/gateway'

const { copy, isSupported } = useClipboard()

const requestFields = computed(() => parseOpenAPIFields(props.item.request_fields))
const responseFields = computed(() => parseOpenAPIFields(props.item.response_fields))

const requestJson = computed(() => formatJson(props.item.request_schema))
const responseJson = computed(() => formatJson(props.item.response_schema))

const gatewayResponseJson = computed(() => `{
  "code": 200,
  "message": "success",
  "data": ${props.item.encrypted ? '"<3DES 密文或 JSON>"' : '...'},
  "success": true
}`)


function formatJson(raw?: string) {
  if (!raw)
    return ''
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  }
  catch {
    return raw
  }
}

async function copyText(text: string, label: string) {
  const val = text?.trim()
  if (!val)
    return
  if (!isSupported.value) {
    ElMessage.warning('当前浏览器不支持复制')
    return
  }
  await copy(val)
  ElMessage.success(`${label}已复制`)
}
</script>

<template>
  <div class="api-detail">
    <header class="api-header">
      <h1 class="md-h1">
        {{ item.title }}
      </h1>
      <p v-if="item.description" class="api-desc">
        {{ item.description }}
      </p>
      <div class="api-meta">
        <el-tag type="success" effect="dark" size="small">
          POST
        </el-tag>
        <el-tag v-if="item.encrypted" size="small">
          3DES
        </el-tag>
        <el-tag v-if="item.billable" size="small" type="warning">
          计费
        </el-tag>
        <el-tag v-else size="small" type="danger">
          免费
        </el-tag>
      </div>
    </header>

    <section class="api-section">
      <div class="section-head">
        <h2 class="md-h2">
          请求节点
        </h2>
        <el-button type="primary" link @click="copyText(item.action, 'Action ')">
          <el-icon><DocumentCopy /></el-icon>
          复制
        </el-button>
      </div>
      <div class="inline-code-box">
        <code>{{ item.action }}</code>
      </div>
    </section>

    <section class="api-section">
      <div class="section-head">
        <h2 class="md-h2">
          调用地址
        </h2>
        <el-button type="primary" link @click="copyText(gatewayPath, '地址 ')">
          <el-icon><DocumentCopy /></el-icon>
          复制
        </el-button>
      </div>
      <div class="endpoint-box">
        <span class="method">POST</span>
        <span class="path">{{ gatewayPath }}</span>
      </div>
      <p class="section-tip">
        Content-Type: <code>application/x-www-form-urlencoded</code>，通过 <code>action={{ item.action }}</code> 指定接口。
        <template v-if="item.encrypted">
          业务 JSON 经 3DES 加密后放入 <code>data</code> 字段，并携带握手获得的 <code>token</code>。
        </template>
      </p>
    </section>

    <section class="api-section">
      <div class="section-head">
        <h2 class="md-h2">
          请求体
        </h2>
        <el-button
          v-if="requestJson"
          type="primary"
          link
          @click="copyText(requestJson, '请求体 ')"
        >
          <el-icon><DocumentCopy /></el-icon>
          复制
        </el-button>
      </div>
      <p v-if="item.encrypted" class="section-tip">
        以下为 3DES 解密后的业务 JSON 结构：
      </p>
      <CopyCodeBlock v-if="requestJson" :content="requestJson" />
      <SchemaFieldTable :fields="requestFields" />
    </section>

    <section class="api-section">
      <div class="section-head">
        <h2 class="md-h2">
          响应体
        </h2>
        <el-button
          type="primary"
          link
          @click="copyText(item.encrypted && responseJson ? `${gatewayResponseJson}\n\n${responseJson}` : (responseJson || gatewayResponseJson), '响应 ')"
        >
          <el-icon><DocumentCopy /></el-icon>
          复制全部
        </el-button>
      </div>

      <h3 class="md-h3">
        网关外层
      </h3>
      <CopyCodeBlock :content="gatewayResponseJson" />

      <template v-if="item.encrypted && responseJson">
        <div class="section-head section-head--sub">
          <h3 class="md-h3">
            业务 JSON（3DES 解密后）
          </h3>
          <el-button type="primary" link @click="copyText(responseJson, '业务响应 ')">
            <el-icon><DocumentCopy /></el-icon>
            复制
          </el-button>
        </div>
        <CopyCodeBlock :content="responseJson" />
        <SchemaFieldTable :fields="responseFields" />
      </template>
      <template v-else-if="responseJson">
        <CopyCodeBlock :content="responseJson" />
        <SchemaFieldTable :fields="responseFields" />
      </template>
    </section>
  </div>
</template>

<style scoped>
.api-detail {
  padding: 16px 20px 32px;
}

/* Markdown 风格标题 */
.md-h1 {
  margin: 0 0 10px;
  font-size: 32px;
  font-weight: 700;
  line-height: 1.25;
  color: var(--el-text-color-primary);
}

.md-h2 {
  margin: 0;
  font-size: 22px;
  font-weight: 700;
  line-height: 1.35;
  color: var(--el-text-color-primary);
}

.md-h3 {
  margin: 0;
  font-size: 17px;
  font-weight: 600;
  line-height: 1.4;
  color: var(--el-text-color-primary);
}

.api-desc {
  margin: 0 0 12px;
  font-size: 14px;
  color: var(--el-text-color-secondary);
  line-height: 1.65;
}

.api-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
}

.api-section {
  margin-top: 24px;
  padding-top: 4px;
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 10px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--el-border-color);
}

.section-head--sub {
  margin-top: 16px;
  border-bottom: none;
  padding-bottom: 0;
  margin-bottom: 8px;
}

.section-tip {
  margin: 8px 0 0;
  font-size: 13px;
  color: var(--el-text-color-secondary);
  line-height: 1.6;
}

.section-tip code {
  padding: 1px 4px;
  background: var(--el-fill-color-light);
  border-radius: 3px;
  font-size: 12px;
}

.inline-code-box {
  padding: 8px 12px;
  background: var(--el-fill-color-lighter);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  font-family: ui-monospace, monospace;
  font-size: 13px;
  word-break: break-all;
}

.endpoint-box {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: var(--el-fill-color-lighter);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  font-family: ui-monospace, monospace;
  font-size: 13px;
}

.endpoint-box .method {
  padding: 2px 8px;
  background: #49cc90;
  color: #fff;
  border-radius: 4px;
  font-weight: 600;
  font-size: 12px;
  flex-shrink: 0;
}

.endpoint-box .path {
  word-break: break-all;
}
</style>

<script setup lang="ts">
import type { OpenDocCategory } from '@/apis/openplatform'
import { getOpenDocApi } from '@/apis/openplatform'
import ApiDetail from './ApiDetail.vue'

defineOptions({ name: 'OpenPlatformDocs' })

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const keyword = ref('')
const categories = ref<OpenDocCategory[]>([])
const currentAction = ref<string>()

const flatActions = computed(() =>
  categories.value.flatMap(c => c.actions),
)

const filteredCategories = computed(() => {
  const kw = keyword.value.trim().toLowerCase()
  if (!kw)
    return categories.value
  return categories.value
    .map(cat => ({
      ...cat,
      actions: cat.actions.filter(a =>
        a.title.toLowerCase().includes(kw) || a.action.toLowerCase().includes(kw),
      ),
    }))
    .filter(cat => cat.actions.length > 0)
})

const currentItem = computed(() =>
  flatActions.value.find(a => a.action === currentAction.value),
)

function selectAction(action: string) {
  if (currentAction.value === action)
    return
  currentAction.value = action
  // 仅更新 URL 便于刷新/分享，不走 router 导航，避免整页 remount
  const resolved = router.resolve({ path: '/openplatform/docs', query: { action } })
  window.history.replaceState(window.history.state, '', resolved.fullPath)
}

async function loadDocs() {
  loading.value = true
  try {
    categories.value = await getOpenDocApi()
    if (!currentAction.value && flatActions.value.length) {
      selectAction(flatActions.value[0].action)
    }
  }
  finally {
    loading.value = false
  }
}

onMounted(() => {
  const q = route.query.action
  if (typeof q === 'string' && q)
    currentAction.value = q
  loadDocs()
})
</script>

<template>
  <div class="open-docs">
    <div class="open-docs-toolbar">
      <el-input
        v-model="keyword"
        placeholder="搜索接口名称或 Action"
        clearable
        class="search-input"
        prefix-icon="Search"
      />
    </div>

    <div class="open-docs-body">
      <aside class="sidebar">
        <nav class="nav-list">
          <div v-for="cat in filteredCategories" :key="cat.name" class="nav-group">
            <div class="nav-group-title">
              {{ cat.name }}
            </div>
            <button
              v-for="item in cat.actions"
              :key="item.action"
              type="button"
              class="nav-item"
              :class="{ active: currentAction === item.action }"
              @click="selectAction(item.action)"
            >
              <span class="nav-item-title">{{ item.title }}</span>
              <span class="nav-item-action">{{ item.action }}</span>
            </button>
          </div>
        </nav>
      </aside>

      <main v-loading="loading && !categories.length" class="content">
        <div class="content-scroll">
          <ApiDetail v-if="currentItem" :item="currentItem" />
          <el-empty v-else description="请选择左侧接口查看文档" />
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.open-docs {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--el-bg-color);
}

.open-docs-toolbar {
  flex-shrink: 0;
  padding: 8px 12px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.search-input {
  max-width: 320px;
}

.open-docs-body {
  flex: 1;
  min-height: 0;
  display: flex;
  overflow: hidden;
}

.sidebar {
  width: 220px;
  flex-shrink: 0;
  border-right: 1px solid var(--el-border-color-lighter);
  overflow-y: auto;
  background: var(--el-bg-color);
}

.nav-list {
  padding: 8px 0 16px;
}

.nav-group-title {
  padding: 6px 12px 2px;
  font-size: 12px;
  font-weight: 600;
  color: var(--el-text-color-secondary);
}

.nav-item {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  width: 100%;
  padding: 8px 12px;
  border: none;
  background: transparent;
  cursor: pointer;
  text-align: left;
  transition: background 0.15s;
}

.nav-item:hover {
  background: var(--el-fill-color-light);
}

.nav-item.active {
  background: var(--el-color-primary-light-9);
  border-right: 2px solid var(--el-color-primary);
}

.nav-item-title {
  font-size: 13px;
  color: var(--el-text-color-primary);
  line-height: 1.4;
}

.nav-item-action {
  margin-top: 2px;
  font-size: 11px;
  color: var(--el-text-color-secondary);
  word-break: break-all;
}

.content {
  flex: 1;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
}

.content-scroll {
  height: 100%;
  overflow-y: auto;
  overflow-x: hidden;
}
</style>

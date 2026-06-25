<script setup lang="ts">
import { Icon } from '@iconify/vue'
import { FileTypeList } from './constants'
import FileAsideStatistics from './FileAsideStatistics.vue'

defineOptions({ name: 'FileAside' })

const route = useRoute()
const router = useRouter()
const statsRef = ref<InstanceType<typeof FileAsideStatistics>>()

const selectedKey = computed(() => {
  const type = route.query.type
  return type ? String(type) : '0'
})

function onClickItem(value: number) {
  router.replace({
    path: route.path,
    query: value === 0 ? {} : { type: String(value) },
  })
}

function refreshStats() {
  statsRef.value?.loadStats()
}

defineExpose({ refreshStats })
</script>

<template>
  <div class="file-aside">
    <el-card shadow="never" :body-style="{ padding: '8px 0' }">
      <div class="file-aside__title">
        <Icon icon="icon-park-outline:application-menu" />
        <span>文件类型</span>
      </div>
      <el-menu :default-active="selectedKey" class="file-type-menu">
        <el-menu-item
          v-for="item in FileTypeList"
          :key="item.value"
          :index="String(item.value)"
          @click="onClickItem(item.value)"
        >
          <Icon :icon="item.icon" class="menu-icon" />
          <span>{{ item.name }}</span>
        </el-menu-item>
      </el-menu>
    </el-card>
    <FileAsideStatistics ref="statsRef" />
  </div>
</template>

<style scoped>
.file-aside__title {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 16px 8px;
  font-weight: 600;
}
.file-type-menu {
  border-right: none;
}
.menu-icon {
  margin-right: 8px;
  font-size: 18px;
}
</style>

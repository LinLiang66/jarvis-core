<script setup lang="ts">
import FileAside from './FileAside.vue'
import FileMain from './main/index.vue'

defineOptions({ name: 'SystemFile' })

const asideRef = ref<InstanceType<typeof FileAside>>()

function onStatsRefresh() {
  asideRef.value?.refreshStats()
}
</script>

<template>
  <div class="file-manage">
    <aside class="file-manage__aside">
      <FileAside ref="asideRef" />
    </aside>
    <main class="file-manage__main">
      <FileMain @stats-refresh="onStatsRefresh" />
    </main>
  </div>
</template>

<style scoped>
.file-manage {
  display: flex;
  gap: 14px;
  height: 100%;
  padding: 12px;
  box-sizing: border-box;
  overflow: hidden;
}
.file-manage__aside {
  width: 220px;
  flex-shrink: 0;
  overflow: auto;
}
.file-manage__main {
  flex: 1;
  min-width: 0;
  overflow: hidden;
}
@media (max-width: 768px) {
  .file-manage {
    flex-direction: column;
  }
  .file-manage__aside {
    width: 100%;
  }
}
</style>

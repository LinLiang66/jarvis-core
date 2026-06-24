<script setup lang="ts">
import { VideoPause, VideoPlay } from '@element-plus/icons-vue'

defineOptions({ name: 'SampleAudioPlayer' })

const props = withDefaults(defineProps<{
  src?: string
  /** 表格等窄列使用紧凑布局 */
  compact?: boolean
}>(), {
  compact: false,
})

const audioRef = ref<HTMLAudioElement>()
const playing = ref(false)
const duration = ref(0)
const current = ref(0)
const loadError = ref(false)

const progress = computed({
  get: () => (duration.value ? (current.value / duration.value) * 100 : 0),
  set: (v: number) => {
    if (!audioRef.value || !duration.value)
      return
    audioRef.value.currentTime = (v / 100) * duration.value
  },
})

const formatTime = (s: number) => {
  if (!Number.isFinite(s) || s < 0)
    return '00:00'
  const m = Math.floor(s / 60)
  const sec = Math.floor(s % 60)
  return `${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}`
}

function togglePlay() {
  const el = audioRef.value
  if (!el || !props.src)
    return
  if (playing.value) {
    el.pause()
  }
  else {
    el.play().catch(() => {
      loadError.value = true
    })
  }
}

function onPlay() {
  playing.value = true
  loadError.value = false
}

function onPause() {
  playing.value = false
}

function onTimeUpdate() {
  current.value = audioRef.value?.currentTime ?? 0
}

function onLoaded() {
  duration.value = audioRef.value?.duration ?? 0
  loadError.value = false
}

function onError() {
  loadError.value = true
  playing.value = false
}

watch(() => props.src, () => {
  playing.value = false
  current.value = 0
  duration.value = 0
  loadError.value = false
})
</script>

<template>
  <div v-if="src" class="sample-audio-player" :class="{ compact }">
    <audio
      ref="audioRef"
      :src="src"
      preload="metadata"
      class="sample-audio-player__native"
      @play="onPlay"
      @pause="onPause"
      @ended="onPause"
      @timeupdate="onTimeUpdate"
      @loadedmetadata="onLoaded"
      @error="onError"
    />
    <div class="sample-audio-player__bar">
      <el-button
        type="primary"
        circle
        size="small"
        :disabled="loadError"
        @click="togglePlay"
      >
        <el-icon>
          <VideoPause v-if="playing" />
          <VideoPlay v-else />
        </el-icon>
      </el-button>
      <el-slider
        v-model="progress"
        :show-tooltip="false"
        size="small"
        class="sample-audio-player__slider"
        :disabled="loadError || !duration"
      />
      <span class="sample-audio-player__time">
        {{ formatTime(current) }} / {{ formatTime(duration) }}
      </span>
    </div>
    <div v-if="loadError" class="sample-audio-player__err">
      无法播放，请检查 URL 是否可公网访问
    </div>
    <el-link
      v-if="!compact"
      :href="src"
      target="_blank"
      type="primary"
      class="sample-audio-player__link"
    >
      在新窗口打开
    </el-link>
  </div>
  <span v-else class="sample-audio-player__empty">—</span>
</template>

<style scoped>
.sample-audio-player {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 200px;
  max-width: 320px;
}

.sample-audio-player.compact {
  max-width: 280px;
}

.sample-audio-player__native {
  display: none;
}

.sample-audio-player__bar {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}

.sample-audio-player__slider {
  flex: 1;
  min-width: 60px;
}

.sample-audio-player__time {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  font-variant-numeric: tabular-nums;
}

.sample-audio-player__err {
  font-size: 12px;
  color: var(--el-color-danger);
}

.sample-audio-player__link {
  font-size: 12px;
  word-break: break-all;
}

.sample-audio-player__empty {
  color: var(--el-text-color-secondary);
}
</style>

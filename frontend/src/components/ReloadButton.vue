<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue'
import { refreshDevices, lastDeviceRefresh, refreshingDevices } from '../services/api'
import { elapsedSince } from '../scripts/utils'
import { useNow } from '../composables/useNow'

const now = useNow()

// auto reload device locations, skipped while the tab is in the background
const AUTO_REFRESH_MS = 30_000
const AUTO_RELOAD_KEY = 'autoReload'

// remembered per browser so the choice survives a page reload
function savedAutoReload(): boolean {
  try {
    return localStorage.getItem(AUTO_RELOAD_KEY) !== 'false'
  } catch {
    return true
  }
}

const autoReload = ref(savedAutoReload())
let timer: ReturnType<typeof setInterval> | undefined

function stopTimer() {
  clearInterval(timer)
  timer = undefined
}

watch(autoReload, (on) => {
  try {
    localStorage.setItem(AUTO_RELOAD_KEY, String(on))
  } catch {}
  stopTimer()
  if (on) {
    timer = setInterval(() => {
      if (!document.hidden) refreshDevices()
    }, AUTO_REFRESH_MS)
  }
}, { immediate: true })

onUnmounted(stopTimer)
</script>

<template>
  <button
    type="button"
    class="btn"
    :disabled="refreshingDevices"
    title="Reload now"
    @click="refreshDevices()"
  >
    <span :class="{ spin: refreshingDevices }">⟳</span>
    <span class="pill">{{ lastDeviceRefresh ? elapsedSince(lastDeviceRefresh, now) + ' ago' : 'never' }}</span>
  </button>

  <button
    type="button"
    class="btn"
    :class="autoReload ? 'on' : 'off'"
    :aria-pressed="autoReload"
    @click="autoReload = !autoReload"
  >
    {{ autoReload ? '● auto-reload: on' : '○ auto-reload: off' }}
  </button>
</template>

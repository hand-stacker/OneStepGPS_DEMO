<script setup lang="ts">
import { ref, computed } from 'vue'
import { userSortPreference, editPreference } from '../services/api'
import type { sort_order } from '../types'
import { useClickOutside } from '../composables/useClickOutside'
import { useBusy } from '../composables/useBusy'

const open = ref(false)
const root = ref<HTMLElement | null>(null)
const { busy, run } = useBusy()

useClickOutside(() => root.value, () => { open.value = false })

const label = computed(() =>
  userSortPreference.value?.sort_order === 'asc' ? 'Recently changed' : 'Longest in status'
)

function choose(order: sort_order) {
  open.value = false
  run(() => editPreference(order))
}
</script>

<template>
  <div ref="root" class="dropdown">
    <button type="button" class="btn light-shady-border" :disabled="busy" @click="open = !open">
      Sort: {{ label }} ▾
    </button>
    <ul v-if="open" class="dropdown-menu panel shady-border">
      <li :class="{ active: userSortPreference?.sort_order === 'asc' }" @click="choose('asc')">
        Recently changed first
      </li>
      <li :class="{ active: userSortPreference?.sort_order === 'desc' }" @click="choose('desc')">
        Longest in status first
      </li>
    </ul>
  </div>
</template>

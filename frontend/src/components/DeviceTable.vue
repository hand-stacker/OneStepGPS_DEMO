<script setup lang="ts">
import { sortedDevices, hiddenDevices, findName, setDeviceHidden, centerOnDevice } from '../services/api'
import { elapsedSince } from '../scripts/utils'
import { useNow } from '../composables/useNow'
import { ref } from 'vue'

const now = useNow()

// device_ids whose hide checkbox is still saving
const pending = ref<string[]>([])

async function toggleHidden(device_id: string, hidden: boolean) {
  pending.value.push(device_id)
  try {
    await setDeviceHidden(device_id, hidden)
  } finally {
    pending.value = pending.value.filter((id) => id !== device_id)
  }
}
</script>

<template>
  <table class="device-table">
    <thead>
      <tr>
        <th>Name</th>
        <th>Status</th>
        <th>Time</th>
        <th>Hide</th>
      </tr>
    </thead>
    <tbody>
      <tr
        v-for="device in sortedDevices"
        :key="device.device_id"
        :class="{ hidden: hiddenDevices.includes(device.device_id) }"
      >
        <td>
          <button
            v-if="device.lat != null"
            type="button"
            class="link"
            @click="centerOnDevice(device)"
          >
            {{ findName(device) }}
          </button>
          <span v-else>{{ findName(device) }}</span>
        </td>
        <td>{{ device.drive_status ?? '—' }}</td>
        <td>{{ device.drive_status_begin_time ? elapsedSince(device.drive_status_begin_time, now) : '—' }}</td>
        <td>
          <input
            type="checkbox"
            :checked="hiddenDevices.includes(device.device_id)"
            :aria-label="`Hide ${findName(device)}`"
            :disabled="pending.includes(device.device_id)"
            @change="toggleHidden(device.device_id, ($event.target as HTMLInputElement).checked)"
          />
        </td>
      </tr>
    </tbody>
  </table>
</template>

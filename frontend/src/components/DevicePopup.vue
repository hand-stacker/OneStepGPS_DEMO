<script setup lang="ts">
import { ref } from 'vue'
import {
  deviceNicknames, devicesWithCustomMarkers,
  findName, setDeviceHidden, saveNickname, uploadMarker, removeMarker,
} from '../services/api'
import { formatPacificTime, elapsedSince, formatFuelPercentage, formatSpeed } from '../scripts/utils'
import { useNow } from '../composables/useNow'
import { useClickOutside } from '../composables/useClickOutside'
import { useBusy } from '../composables/useBusy'

const props = defineProps<{ device: any }>()
const { busy, run } = useBusy()
const emit = defineEmits<{ close: [] }>()
const root = ref<HTMLElement | null>(null)
useClickOutside(() => root.value?.parentElement, () => emit('close'))
const now = useNow()

// the popup is re-created each time it opens
const nicknameDraft = ref(deviceNicknames.value[props.device.device_id] ?? '')
const markerError = ref('')

function onSaveNickname(display_name: string) {
  run(async () => {
    nicknameDraft.value = await saveNickname(props.device.device_id, display_name)
  })
}

function onMarkerFile(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  run(async () => {
    markerError.value = await uploadMarker(props.device.device_id, file)
    input.value = ''
  })
}

function onRemoveMarker() {
  run(() => removeMarker(props.device.device_id))
}

// the popup closes once the device is hidden and drops off the map
function onHide() {
  run(() => setDeviceHidden(props.device.device_id, true))
}
</script>

<template>
  <div ref="root" class="device-popup panel shady-border" @click.stop @mousedown.stop @dblclick.stop>
    <button type="button" class="close light-shady-border" aria-label="Close" title="Close" @click="emit('close')">×</button>
    <div class="marker-label">{{ findName(device) }}</div>
    <div>
      Status: {{ device.drive_status }}
      <span v-if="device.drive_status_begin_time" class="pill">
        {{ elapsedSince(device.drive_status_begin_time, now) }}
      </span>
    </div>
    <div v-if="device.drive_status === 'driving'">
      <div>Speed : {{ formatSpeed(device.speed_mph) }}</div>
      <div>Drive Distance: {{ Number(device.drive_status_distance_mi).toFixed(2) }} mi</div>
    </div>
    <div>Last Update: {{ formatPacificTime(device.honored_gps_time) }}</div>
    <div>Fuel Percent : {{ formatFuelPercentage(device.fuel_percent) }}</div>

    <fieldset :disabled="busy">
      <form @submit.prevent="onSaveNickname(nicknameDraft)">
        <input v-model="nicknameDraft" placeholder="Add a nickname" />
        <button class="light-shady-border">Save nickname</button>
        <button
          v-if="device.device_id in deviceNicknames"
          type="button"
          class="light-shady-border"
          @click="onSaveNickname('')"
        >
          Clear nickname
        </button>
      </form>

      <div class="section">
        <label>
          Marker image:
          <input
            type="file"
            accept="image/png,image/jpeg,image/gif,image/webp"
            @change="onMarkerFile"
          />
        </label>
        <button
          v-if="devicesWithCustomMarkers.includes(device.device_id)"
          type="button"
          class="light-shady-border"
          @click="onRemoveMarker"
        >
          Use default image
        </button>
        <div v-if="markerError" class="error">{{ markerError }}</div>
      </div>

      <button type="button" class="light-shady-border" @click="onHide">Hide from map</button>
    </fieldset>
  </div>
</template>

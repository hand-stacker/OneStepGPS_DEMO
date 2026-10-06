<script setup lang="ts">
import { ref, watch } from 'vue'
import { GoogleMap, MarkerCluster, CustomMarker } from 'vue3-google-map'
import { visibleDevices, center, currentUser, findName, markerUrl } from '../services/api'
import DevicePopup from './DevicePopup.vue'

// device whose popup is open
const selectedDeviceId = ref<string | null>(null)

// close the popup if its device gets hidden (e.g. from the side table)
watch(visibleDevices, (devices) => {
  if (selectedDeviceId.value && !devices.some((d) => d.device_id === selectedDeviceId.value)) {
    selectedDeviceId.value = null
  }
})

// a popup left open would still show the previous user's nickname draft
// so we check updates on currentUser
watch(currentUser, () => {
  selectedDeviceId.value = null
})
</script>

<template>
  <GoogleMap
    class="map"
    api-key="AIzaSyB5TjaHMZtdRyrLxOMRC_iQib4o98nth0M"
    :center="center"
    :zoom="7"
  >
    <MarkerCluster>
      <CustomMarker
        v-for="device in visibleDevices"
        :key="device.device_id"
        :options="{
          position: { lat: Number(device.lat), lng: Number(device.lng) },
          anchorPoint: 'BOTTOM_CENTER',
          zIndex: selectedDeviceId === device.device_id ? 1000 : 1,
        }"
      >
        <div class="marker" @click="selectedDeviceId = device.device_id">
          <DevicePopup
            v-if="selectedDeviceId === device.device_id"
            :device="device"
            @close="selectedDeviceId = null"
          />
          <div class="marker-label panel">{{ findName(device) }}</div>
          <img :src="markerUrl(device.device_id)" width="50" height="50" />
        </div>
      </CustomMarker>
    </MarkerCluster>
  </GoogleMap>
</template>

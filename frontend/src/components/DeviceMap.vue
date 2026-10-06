<script setup lang="ts">
import { ref, watch } from 'vue'
import { GoogleMap, MarkerCluster, CustomMarker } from 'vue3-google-map'
import { visibleDevices, center, centerVersion, currentUser, findName, markerUrl, zoom } from '../services/api'
import DevicePopup from './DevicePopup.vue'

const mapRef = ref<InstanceType<typeof GoogleMap> | null>(null)

// updates map if centerVersion changes
watch(centerVersion, () => {
  const map = mapRef.value?.map
  if (!map) return
  map.panTo(center.value)
  map.setZoom(zoom.value)
})

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
    ref="mapRef"
    class="map"
    api-key="AIzaSyB5TjaHMZtdRyrLxOMRC_iQib4o98nth0M"
    :center="center"
    :zoom="zoom"
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
          <div class="marker-icon">
            <div v-if="device.drive_status === 'driving'" class="drive-ping"></div>
            <div v-else-if="device.drive_status === 'idle'" class="idle-glow"></div>
            <img :src="markerUrl(device.device_id)" width="50" height="50" />
          </div>
        </div>
      </CustomMarker>
    </MarkerCluster>
  </GoogleMap>
</template>

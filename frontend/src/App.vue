<script setup lang="ts">

import { ref, onMounted } from 'vue'
import { GoogleMap, MarkerCluster, CustomMarker, InfoWindow } from 'vue3-google-map'
import {friendliestNodePos} from './scripts/utils'

type Preference = {
  user_id: number | string
  sort_order: string
}

const prefs = ref<Preference[]>([])
const gpsData = ref<any[]>([])
const deviceInfo = ref<any[]>([])
const text = ref('')
var GOOGLE_MAPS_API_KEY ='DEMO'
var center = ref({ lat: 0, lng: 0  })

async function load() {
  prefs.value = await (await fetch('/api/all-preferences')).json()
}

async function loadGPSBulk() {
  gpsData.value = await (await fetch('/api/gps-bulk')).json()
}

async function loadDeviceInfo() {
  deviceInfo.value = await (await fetch('/api/device-info')).json()
}

async function initialLoad() {
  GOOGLE_MAPS_API_KEY = await (await fetch('/api/google-maps-key')).json()
  await load()
  await loadGPSBulk()
  await loadDeviceInfo()
  var bestNodePos = friendliestNodePos(deviceInfo.value)
  console.log('bestNodePos', bestNodePos)
  center.value = { lat: bestNodePos.lat, lng: bestNodePos.lng}
  console.log('center', center)
}

async function edit(p: Preference) {
  await fetch(`api/user-preference/${p.user_id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ sort_order: p.sort_order }),
  })
}

async function add() {
  if (!text.value) return
  await fetch('/api/user-preference/', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ sort_order: text.value }),
  })
  text.value = ''
  load()
}

onMounted(initialLoad)
</script>

<template>
  <form @submit.prevent="add">
    <input v-model="text" placeholder="New note" />
    <button>Add</button>
  </form>

  <form @submit.prevent="edit(p)" v-for="p in prefs" :key="p.user_id">
    <input v-model="p.sort_order" />
    <button>Save</button>
  </form>
  <GoogleMap
    api-key="AIzaSyB5TjaHMZtdRyrLxOMRC_iQib4o98nth0M"
    style="width: 80%; height: 500px"
    :center="center"
    :zoom="7"
    >
    
    <MarkerCluster>
      <CustomMarker
        v-for="(device, device_id) in deviceInfo"
        :key="device_id"
        :options="{ position: { lat: Number(device.lat), lng: Number(device.lng) }, anchorPoint: 'BOTTOM_CENTER' }"
      >
        <div style="text-align: center">
          <InfoWindow>
          <div style="text-align: center">
            <div style="font-size: 1.125rem; background: #f0f0f0; padding: 4px">{{device.display_name}}</div>
            <div>Device ID: {{device.device_id}}</div>
            <div>License Plate: {{device.license_plate}}</div>
            <div>Last Update: {{device.last_update}}</div>
          </div>
        </InfoWindow>
          <div style="font-size: 1.125rem; background: #f0f0f0; padding: 4px">{{device.display_name}}</div>
          <img src="https://vuejs.org/images/logo.png" width="50" height="50" style="margin-top: 8px" />
        </div>
        
      </CustomMarker>
    </MarkerCluster>
  </GoogleMap>
  <details>
    <summary>Loaded external api data</summary>
    <pre>{{ deviceInfo }}</pre>
  </details>
</template>
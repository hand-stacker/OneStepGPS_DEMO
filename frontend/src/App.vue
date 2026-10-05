<script setup lang="ts">

import { ref, onMounted } from 'vue'
import type { User, UserSortPreference } from './types'
import { GoogleMap, MarkerCluster, CustomMarker, InfoWindow } from 'vue3-google-map'
import {friendliestNodePos} from './scripts/utils'

const users = ref<User[]>([])
const userSortPreference = ref<UserSortPreference>()
const gpsData = ref<any[]>([])
const deviceInfo = ref<any[]>([])
const hiddenDevices = ref<string[]>([])
const deviceNicknames = ref<Record<string, string>>({})
const devicesWithCustomMarkers = ref<string[]>([])
const email = ref('')
const currentUser = ref<User>()
const GOOGLE_MAPS_API_KEY = ref('DEMO')
const center = ref({ lat: 0, lng: 0 })
const userMenuOpen = ref(false)

async function loadUsers() {
  users.value = await (await fetch('/api/all-users')).json()
}

async function loadPreferences() {
  userSortPreference.value = await (await fetch(`/api/user-preference/${String(currentUser.value?.user_id)}`)).json()
}

async function loadHiddenDevices() {
  hiddenDevices.value = (await (await fetch(`/api/hidden-devices/${String(currentUser.value?.user_id)}`)).json()) ?? []
}

async function loadDeviceNicknames() {
  deviceNicknames.value = (await (await fetch(`/api/device-nicknames/${String(currentUser.value?.user_id)}`)).json()) ?? {}
}

async function loadDeviceIdsWithCustomMarkers() {
  devicesWithCustomMarkers.value = (await (await fetch(`/api/device-markers-list/${String(currentUser.value?.user_id)}`)).json()) ?? []
}

async function loadGPSBulk() {
  gpsData.value = await (await fetch('/api/gps-bulk')).json()
}

async function loadDeviceInfo() {
  deviceInfo.value = await (await fetch('/api/device-info')).json()
}


// when shifting users we should reload the preferences, device info, 
// custom names and markers
async function loadCurrentUser(id: number) {
  currentUser.value = await (await fetch(`/api/user/${String(id)}`)).json()
  await loadPreferences()
  await loadDeviceInfo()
  await loadHiddenDevices()
  await loadDeviceNicknames()
  await loadDeviceIdsWithCustomMarkers()

}

async function initialLoad() {
  GOOGLE_MAPS_API_KEY.value = (await (await fetch('/api/google-maps-key/')).json()).key
  await loadUsers()
  await loadCurrentUser(1)
  if (deviceInfo.value.length > 0) {
    const bestNodePos = friendliestNodePos(deviceInfo.value)
    center.value = { lat: Number(bestNodePos.lat), lng: Number(bestNodePos.lng) }
  }
}

async function pushNewUser() {
  await fetch('/api/user/', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email: email.value }),
  })
  email.value = ''
  await loadUsers()
}

async function selectUser(id: number | string) {
  userMenuOpen.value = false
  await loadCurrentUser(Number(id))
}

async function editPreference(p: UserSortPreference) {

  // will edit current user's preference even if the UserSortPreference
  // is modified and has another user's id
  await fetch(`/api/user-preference/${String(currentUser.value?.user_id)}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ sort_order: p.sort_order }),
  })
}

async function addHiddenDevice(device_id : string, ignore : boolean = false) {
  await fetch(`/api/hidden-devices/${String(currentUser.value?.user_id)}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ device_id: device_id , ignore : ignore}),
  })
}

async function editHiddenDevice(device_id : string, ignore : boolean = false) {
  await fetch(`/api/hidden-devices/${String(currentUser.value?.user_id)}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ device_id: device_id , ignore : ignore}),
  })
}

async function addDeviceNickname(device_id : string, display_name : string, ignore : boolean = false) {
  await fetch(`/api/device-nicknames/${String(currentUser.value?.user_id)}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ device_id: device_id , display_name : display_name, ignore : ignore}),
  })
}

async function editDeviceNickname(device_id : string, display_name : string, ignore : boolean = false) {
  await fetch(`/api/device-nicknames/${String(currentUser.value?.user_id)}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ device_id: device_id , display_name : display_name, ignore : ignore}),
  })
}

// modifications for content

function findName(device: any ): string {
  if (device.device_id in deviceNicknames) {
    return deviceNicknames[device.device_id]
  }
  return device.display_name
}

onMounted(initialLoad)
</script>

<template>
  <nav class="navbar">
    <div class="user-menu">
      <button type="button" @click="userMenuOpen = !userMenuOpen">
        {{ currentUser?.email ?? 'Select user' }} ▾
      </button>
      <ul v-if="userMenuOpen" class="user-list">
        <li
          v-for="u in users"
          :key="u.user_id"
          :class="{ active: u.user_id === currentUser?.user_id }"
          @click="selectUser(u.user_id)"
        >
          {{ u.email }}
        </li>
        <li class="new-user">
          <form @submit.prevent="pushNewUser">
            <input v-model="email" type="email" placeholder="new user email" required />
            <button>Add</button>
          </form>
        </li>
      </ul>
    </div>
  </nav>

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
</template>

<style scoped>
.navbar {
  display: flex;
  justify-content: flex-end;
  padding: 8px;
  border-bottom: 1px solid #ddd;
}

.user-menu {
  position: relative;
}

.user-list {
  position: absolute;
  right: 0;
  z-index: 10;
  min-width: 240px;
  margin: 4px 0 0;
  padding: 0;
  list-style: none;
  background: white;
  border: 1px solid #ddd;
}

.user-list li {
  padding: 6px 8px;
  cursor: pointer;
}

.user-list li:hover,
.user-list li.active {
  background: #f0f0f0;
}

.user-list li.new-user {
  cursor: default;
  border-top: 1px solid #ddd;
}

.user-list li.new-user:hover {
  background: white;
}
</style>
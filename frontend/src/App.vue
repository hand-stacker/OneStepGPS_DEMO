<script setup lang="ts">

import { ref, computed, watch, onMounted } from 'vue'
import type { User, UserSortPreference } from './types'
import { GoogleMap, MarkerCluster, CustomMarker } from 'vue3-google-map'
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

function findName(device: any): string {
  return deviceNicknames.value[device.device_id] ?? device.display_name
}

const DEFAULT_MARKER_URL = 'https://vuejs.org/images/logo.png'

// bumped after every upload so the <img> refetches instead of showing the old image
const markerVersion = ref(0)

function markerUrl(device_id: string): string {
  if (devicesWithCustomMarkers.value.includes(device_id)) {
    return `/api/device-markers/${String(currentUser.value?.user_id)}/${device_id}?v=${markerVersion.value}`
  }
  return DEFAULT_MARKER_URL
}

const markerError = ref('')

// backend expects multipart/form-data with the file under 'image'
async function uploadMarker(device_id: string, event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  const formData = new FormData()
  formData.append('image', file)
  const res = await fetch(`/api/device-markers/${String(currentUser.value?.user_id)}/${device_id}`, {
    method: 'POST',
    body: formData,
  })
  input.value = ''
  markerError.value = res.ok ? '' : await res.text()
  markerVersion.value++
  await loadDeviceIdsWithCustomMarkers()
}

// sets the marker to ignore on the backend, the default image is shown again
async function removeMarker(device_id: string) {
  await fetch(`/api/device-markers/${String(currentUser.value?.user_id)}/${device_id}`, {
    method: 'DELETE',
  })
  await loadDeviceIdsWithCustomMarkers()
}

// /api/device-info leaves out hidden devices, so remember every device we've
// seen to still be able to show hidden ones by name
const knownDevices = ref<Record<string, any>>({})

watch(deviceInfo, (devices) => {
  for (const d of devices) {
    knownDevices.value[d.device_id] = d
  }
})

const visibleDevices = computed(() =>
  deviceInfo.value.filter((d) => !hiddenDevices.value.includes(d.device_id))
)

// every device we know of plus any hidden id we've never seen device info for
const allDevices = computed(() => {
  const devices = { ...knownDevices.value }
  for (const id of hiddenDevices.value) {
    devices[id] ??= { device_id: id, display_name: id }
  }
  return Object.values(devices)
})

async function setDeviceHidden(device_id: string, hidden: boolean) {
  if (hidden) {
    await addHiddenDevice(device_id, false)
  } else {
    await editHiddenDevice(device_id, true)
  }
  selectedDeviceId.value = null
  await loadCurrentUser(Number(currentUser.value?.user_id))
}

// popup opened by clicking a marker
const selectedDeviceId = ref<string | null>(null)
const nicknameDraft = ref('')

function openDevicePopup(device: any) {
  selectedDeviceId.value = device.device_id
  nicknameDraft.value = deviceNicknames.value[device.device_id] ?? ''
  markerError.value = ''
}

function closeDevicePopup() {
  selectedDeviceId.value = null
}

// an empty nickname is treated the same as removing it
async function saveNickname(device_id: string, display_name: string) {
  const name = display_name.trim()
  const ignore = name === ''
  if (device_id in deviceNicknames.value) {
    await editDeviceNickname(device_id, name, ignore)
  } else {
    await addDeviceNickname(device_id, name, ignore)
  }
  await loadDeviceNicknames()
  nicknameDraft.value = name
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
        v-for="device in visibleDevices"
        :key="device.device_id"
        :options="{ position: { lat: Number(device.lat), lng: Number(device.lng) }, anchorPoint: 'BOTTOM_CENTER' }"
      >
        <div class="marker" @click="openDevicePopup(device)">
          <div
            v-if="selectedDeviceId === device.device_id"
            class="device-popup"
            @click.stop
            @mousedown.stop
            @dblclick.stop
          >
            <button type="button" class="close" @click="closeDevicePopup">×</button>
            <div class="marker-label">{{ findName(device) }}</div>
            <div>Device ID: {{ device.device_id }}</div>
            <div>License Plate: {{ device.license_plate }}</div>
            <div>Last Update: {{ device.last_update }}</div>

            <form @submit.prevent="saveNickname(device.device_id, nicknameDraft)">
              <input v-model="nicknameDraft" placeholder="nickname" />
              <button>Save</button>
              <button
                v-if="device.device_id in deviceNicknames"
                type="button"
                @click="saveNickname(device.device_id, '')"
              >
                Remove
              </button>
            </form>

            <div class="marker-image-options">
              <label>
                Custom image:
                <input
                  type="file"
                  accept="image/png,image/jpeg,image/gif,image/webp"
                  @change="uploadMarker(device.device_id, $event)"
                />
              </label>
              <button
                v-if="devicesWithCustomMarkers.includes(device.device_id)"
                type="button"
                @click="removeMarker(device.device_id)"
              >
                Remove image
              </button>
              <div v-if="markerError" class="error">{{ markerError }}</div>
            </div>

            <button type="button" @click="setDeviceHidden(device.device_id, true)">Hide device</button>
          </div>

          <div class="marker-label">{{ findName(device) }}</div>
          <img :src="markerUrl(device.device_id)" width="50" height="50" style="margin-top: 8px" />
        </div>
      </CustomMarker>
    </MarkerCluster>
  </GoogleMap>

  <section class="hidden-devices">
    <h3>Hidden devices</h3>
    <ul>
      <li v-for="device in allDevices" :key="device.device_id">
        <label>
          <input
            type="checkbox"
            :checked="hiddenDevices.includes(device.device_id)"
            @change="setDeviceHidden(device.device_id, ($event.target as HTMLInputElement).checked)"
          />
          {{ findName(device) }}
        </label>
      </li>
    </ul>
  </section>
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

.marker {
  position: relative;
  text-align: center;
  cursor: pointer;
}

.marker-label {
  font-size: 1.125rem;
  background: #f0f0f0;
  padding: 4px;
}

.device-popup {
  position: absolute;
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%);
  z-index: 20;
  min-width: 220px;
  margin-bottom: 8px;
  padding: 8px;
  text-align: left;
  cursor: default;
  background: white;
  border: 1px solid #ccc;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.2);
}

.device-popup .close {
  float: right;
}

.device-popup form,
.marker-image-options {
  margin: 8px 0;
}

.error {
  color: #c00;
}

.hidden-devices ul {
  padding: 0;
  list-style: none;
}
</style>
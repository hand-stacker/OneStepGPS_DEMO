import { ref, computed } from 'vue'
import type { User, UserSortPreference, sort_order } from '../types'
import { friendliestNodePos } from '../scripts/utils'

// app data lives here at module level, so every component that imports it shares the same state.
// App.vue only reads these values and calls the exported functions below.

/// STATE

export const users = ref<User[]>([])
export const currentUser = ref<User>()
export const userSortPreference = ref<UserSortPreference>()
export const gpsData = ref<any>()
export const deviceInfo = ref<any[]>([])
// device ids that user choose to hide
export const hiddenDevices = ref<string[]>([])
// device_id -> custom name
export const deviceNicknames = ref<Record<string, string>>({})
// device_ids the user has uploaded a custom marker for
export const devicesWithCustomMarkers = ref<string[]>([])
export const GOOGLE_MAPS_API_KEY = ref('DEMO')
export const center = ref({ lat: 0, lng: 0 })

// bumped after every upload so the <img> refetches instead of showing the old image
const markerVersion = ref(0)

const DEFAULT_MARKER_URL = 'https://vuejs.org/images/logo.png'

// /api/device-info leaves out hidden devices, so their device_id and display_name
// come from /api/hidden-device-info instead
export const hiddenDeviceInfo = ref<any[]>([])

export const visibleDevices = computed(() =>
    deviceInfo.value.filter((d) => !hiddenDevices.value.includes(d.device_id))
)

// visible and hidden devices together, hidden ids with no device info are left out
export const allDevices = computed(() => {
    const devices: Record<string, any> = {}
    for (const d of [...deviceInfo.value, ...hiddenDeviceInfo.value]) {
        devices[d.device_id] = d
    }
    return Object.values(devices)
})

// allDevices sorted by time since drive_status_begin_time using the user's sort preference.
// asc = shortest time first (most recent status change), desc = longest first.
// hidden devices are sorted last
export const sortedDevices = computed(() => {
    const direction = userSortPreference.value?.sort_order === 'asc' ? 1 : -1
    const beginMs = (d: any) => d.drive_status_begin_time ? new Date(d.drive_status_begin_time).getTime() : NaN
    return [...allDevices.value].sort((a, b) => {
        const aMs = beginMs(a)
        const bMs = beginMs(b)
        if (isNaN(aMs) && isNaN(bMs)) return 0
        if (isNaN(aMs)) return 1
        if (isNaN(bMs)) return -1
        return (bMs - aMs) * direction
    })
})

// id of the selected user
function userId(): string {
    return String(currentUser.value?.user_id)
}

/// DISPLAY HELPERS

export function findName(device: any): string {
    return deviceNicknames.value[device.device_id] ?? device.display_name
}

export function markerUrl(device_id: string): string {
    if (devicesWithCustomMarkers.value.includes(device_id)) {
        return `/api/device-markers/${userId()}/${device_id}?v=${markerVersion.value}`
    }
    return DEFAULT_MARKER_URL
}

/// HTTP HELPERS

async function getJson<T>(url : string): Promise<T> {
    var ret = await fetch(url)
    if (ret.ok) return ret.json()
    throw new Error(`error, code : ${ret.status}`)
}

async function sendJson(url : string, method : string, body : any): Promise<Response> {
    var ret = await fetch(url, {
        method: method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
    })
    if (ret.ok) return ret
    throw new Error(`error, code : ${ret.status}`)
}

/// LOADING

export async function loadUsers() {
    users.value = await getJson('/api/all-users')
}

// fetches per-user data for whoever is selected right now and only applies it if
// that user is still selected when the response arrives. without this, switching
// users mid-load could let the previous user's slower response overwrite the new one
async function loadForUser<T>(path: string, apply: (data: T, id: string) => void) {
    const id = userId()
    const data = await getJson<T>(`${path}/${id}`)
    if (id === userId()) apply(data, id)
}

// the backend returns just the sort order string, e.g. "asc", not a full preference object
export async function loadPreferences() {
    await loadForUser<sort_order>('/api/user-preference', (sort_order, id) => {
        userSortPreference.value = { user_id: id, sort_order: sort_order }
    })
}

export async function loadHiddenDevices() {
    await loadForUser<string[] | null>('/api/hidden-devices', (ids) => {
        hiddenDevices.value = ids ?? []
    })
}

export async function loadDeviceNicknames() {
    await loadForUser<Record<string, string> | null>('/api/device-nicknames', (nicknames) => {
        deviceNicknames.value = nicknames ?? {}
    })
}

export async function loadDeviceIdsWithCustomMarkers() {
    await loadForUser<string[] | null>('/api/device-markers-list', (ids) => {
        devicesWithCustomMarkers.value = ids ?? []
    })
}

export async function loadGPSBulk() {
    gpsData.value = await getJson('/api/gps-bulk')
}

export async function loadDeviceInfo() {
    await loadForUser<any[] | null>('/api/device-info', (devices) => {
        deviceInfo.value = devices ?? []
    })
}

export async function loadHiddenDeviceInfo() {
    await loadForUser<any[] | null>('/api/hidden-device-info', (devices) => {
        hiddenDeviceInfo.value = devices ?? []
    })
}

// time (ms since epoch) the map data was last reloaded, 0 before the first load
export const lastDeviceRefresh = ref(0)
export const refreshingDevices = ref(false)

// set when a reload is asked for while one is already running
let refreshQueued = false

// reloads only the map data (device locations/status), not user settings.
// requests don't stack up: if a reload is running, one more runs right after it
// so changes made in the meantime (like hiding a device) aren't missed
export async function refreshDevices() {
    if (refreshingDevices.value) {
        refreshQueued = true
        return
    }
    refreshingDevices.value = true
    try {
        await Promise.all([loadDeviceInfo(), loadHiddenDeviceInfo()])
        lastDeviceRefresh.value = Date.now()
    } finally {
        refreshingDevices.value = false
    }
    if (refreshQueued) {
        refreshQueued = false
        await refreshDevices()
    }
}

// only the newest user switch is applied if several are clicked quickly
let userRequest = 0

// switching users clears the previous user's settings right away so they never
// show under the new user, then loads everything for the new user in parallel
export async function loadCurrentUser(id: number) {
    const request = ++userRequest
    const user = await getJson<User>(`/api/user/${id}`)
    if (request !== userRequest) return

    currentUser.value = user
    userSortPreference.value = undefined
    hiddenDevices.value = []
    hiddenDeviceInfo.value = []
    deviceNicknames.value = {}
    devicesWithCustomMarkers.value = []

    await Promise.all([
        loadPreferences(),
        loadHiddenDevices(),
        loadDeviceNicknames(),
        loadDeviceIdsWithCustomMarkers(),
        refreshDevices(),
    ])
}

export async function initialLoad() {
    GOOGLE_MAPS_API_KEY.value = (await getJson<{ key: string }>('/api/google-maps-key/')).key
    await loadUsers()
    await loadCurrentUser(1)
    // friendliestNodePos can't handle an empty list
    if (deviceInfo.value.length > 0) {
        const bestNodePos = friendliestNodePos(deviceInfo.value)
        center.value = { lat: Number(bestNodePos.lat), lng: Number(bestNodePos.lng) }
    }
}

/// SETTING

// moves the map to a device, hidden devices have no location so they're skipped
export function centerOnDevice(device: any) {
    if (device.lat == null || device.lng == null) return
    center.value = { lat: Number(device.lat), lng: Number(device.lng) }
}

export async function addUser(email: string) {
    await sendJson('/api/user/', 'POST', { email: email })
    await loadUsers()
}

export async function editPreference(sort_order: sort_order) {
    await sendJson(`/api/user-preference/${userId()}`, 'PUT', { sort_order: sort_order })
    await loadPreferences()
}

export async function setDeviceHidden(device_id: string, hidden: boolean) {
    if (hidden) {
        await sendJson(`/api/hidden-devices/${userId()}`, 'POST', { device_id: device_id, ignore: false })
    } else {
        await sendJson(`/api/hidden-devices/${userId()}`, 'PUT', { device_id: device_id, ignore: true })
    }
    // only the hidden list and device data change, no need to reload the whole user
    await Promise.all([loadHiddenDevices(), refreshDevices()])
}

// an empty nickname is treated the same as removing it (ignore = true)
// returns the name that was saved
export async function saveNickname(device_id: string, display_name: string): Promise<string> {
    const name = display_name.trim()
    const body = { device_id: device_id, display_name: name, ignore: name === '' }
    const method = device_id in deviceNicknames.value ? 'PUT' : 'POST'
    await sendJson(`/api/device-nicknames/${userId()}`, method, body)
    await loadDeviceNicknames()
    return name
}

// backend expects multipart/form-data with the file under 'image'
// returns the error text from the backend, or '' if the upload worked
export async function uploadMarker(device_id: string, file: File): Promise<string> {
    const formData = new FormData()
    formData.append('image', file)
    const ret = await fetch(`/api/device-markers/${userId()}/${device_id}`, {
        method: 'POST',
        body: formData,
    })
    markerVersion.value++
    await loadDeviceIdsWithCustomMarkers()
    return ret.ok ? '' : await ret.text()
}

// sets the marker to ignore on the backend, the default image is shown again
export async function removeMarker(device_id: string) {
    await fetch(`/api/device-markers/${userId()}/${device_id}`, { method: 'DELETE' })
    await loadDeviceIdsWithCustomMarkers()
}

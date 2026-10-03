<script setup lang="ts">

import { ref, onMounted } from 'vue'

type Preference = {
  user_id: number | string
  sort_order: string
}

const prefs = ref<Preference[]>([])
const gpsData = ref<any[]>([])
const text = ref('')

async function load() {
  prefs.value = await (await fetch('/api/all-preferences')).json()
}

async function loadGPSBulk() {
  gpsData.value = await (await fetch('/api/gps-bulk')).json()
}

async function initialLoad() {
  await load()
  await loadGPSBulk()
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
  <details>
    <summary>Loaded external api data</summary>
    <pre>{{ gpsData }}</pre>
  </details>
</template>
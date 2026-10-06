<script setup lang="ts">

import { onMounted } from 'vue'
import { initialLoad } from './services/api'
import UserMenu from './components/UserMenu.vue'
import DeviceMap from './components/DeviceMap.vue'
import ReloadButton from './components/ReloadButton.vue'
import DeviceTable from './components/DeviceTable.vue'
import SortButton from './components/SortButton.vue'

// all app data lives in services/api.ts, each component reads what it needs from there.
// auto reloading lives in ReloadButton so it can be switched on and off there.

onMounted(initialLoad)
</script>

<template>
  <nav class="navbar">
    <UserMenu />
  </nav>

  <main class="layout">
    <DeviceMap />

    <aside class="sidebar">
      <div class="sidebar-controls">
        <ReloadButton />
        <SortButton />
      </div>
      <DeviceTable />
    </aside>
  </main>
</template>

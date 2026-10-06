<script setup lang="ts">
import { ref } from 'vue'
import { users, currentUser, loadCurrentUser, addUser } from '../services/api'
import { useClickOutside } from '../composables/useClickOutside'
import { useBusy } from '../composables/useBusy'

const email = ref('')
const open = ref(false)
const root = ref<HTMLElement | null>(null)
const { busy, run } = useBusy()

useClickOutside(() => root.value, () => { open.value = false })

function selectUser(id: number | string) {
  open.value = false
  run(() => loadCurrentUser(Number(id)))
}

// shown under the input when adding a user fails, e.g. the email is already taken
const addError = ref('')

function pushNewUser() {
  run(async () => {
    addError.value = await addUser(email.value)
    if (!addError.value) email.value = ''
  })
}
</script>

<template>
  <div ref="root" class="dropdown">
    <button type="button" class="btn" :disabled="busy" @click="open = !open">
      {{ currentUser?.email ?? 'Select user' }} ▾
    </button>
    <ul v-if="open" class="dropdown-menu panel">
      <li
        v-for="u in users"
        :key="u.user_id"
        :class="{ active: u.user_id === currentUser?.user_id }"
        @click="selectUser(u.user_id)"
      >
        {{ u.email }}
      </li>
      <li class="form-row">
        <form @submit.prevent="pushNewUser">
          <fieldset :disabled="busy">
            <input v-model="email" type="email" placeholder="new user email" required @input="addError = ''" />
            <button>Add</button>
          </fieldset>
          <p v-if="addError" class="error">{{ addError }}</p>
        </form>
      </li>
    </ul>
  </div>
</template>

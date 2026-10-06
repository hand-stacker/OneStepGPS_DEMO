import { ref } from 'vue'

// tracks whether an action is still loading so the control that started it can
// show as disabled. run() ignores new calls while busy so double clicks don't stack up.
export function useBusy() {
  const busy = ref(false)

  async function run<T>(action: () => Promise<T>): Promise<T | undefined> {
    if (busy.value) return undefined
    busy.value = true
    try {
      return await action()
    } finally {
      busy.value = false
    }
  }

  return { busy, run }
}

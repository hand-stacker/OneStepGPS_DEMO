import { ref, onMounted, onUnmounted } from 'vue'

// one shared clock that ticks every second so elapsed times stay live.
// the interval only runs while at least one mounted component is using it.
const now = ref(Date.now())
let users = 0
let timer: ReturnType<typeof setInterval> | undefined

export function useNow() {
  onMounted(() => {
    if (users++ === 0) {
      now.value = Date.now()
      timer = setInterval(() => { now.value = Date.now() }, 1000)
    }
  })
  onUnmounted(() => {
    if (--users === 0) clearInterval(timer)
  })
  return now
}

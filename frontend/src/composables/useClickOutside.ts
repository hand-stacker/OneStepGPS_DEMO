import { onMounted, onUnmounted } from 'vue'

// calls onOutside when the user presses anywhere outside the element returned by getElement.
// listens in the capture phase so the map (or anything else) stopping the event doesn't block it.
export function useClickOutside(getElement: () => HTMLElement | null | undefined, onOutside: () => void) {
  function handler(event: PointerEvent) {
    const el = getElement()
    if (el && !el.contains(event.target as Node)) onOutside()
  }

  onMounted(() => document.addEventListener('pointerdown', handler, true))
  onUnmounted(() => document.removeEventListener('pointerdown', handler, true))
}

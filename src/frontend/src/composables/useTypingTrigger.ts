import { ref } from 'vue'
import { TYPING_ACTIVITY_EMIT_THROTTLE_MS, TYPING_ACTIVITY_ENABLED } from '../utils/appConfig'

export function useTypingTrigger(emit: any) {
  const isThrottled = ref(false)

  const triggerTyping = (event?: KeyboardEvent) => {
    if (!TYPING_ACTIVITY_ENABLED) return

    // Ignore Enter (save) but allow Shift+Enter (new line)
    if (event?.key === 'Enter' && !event.shiftKey) return

    if (isThrottled.value) return

    emit('typing')

    // Start cooldown
    isThrottled.value = true
    setTimeout(() => (isThrottled.value = false), TYPING_ACTIVITY_EMIT_THROTTLE_MS)
  }

  return { triggerTyping }
}

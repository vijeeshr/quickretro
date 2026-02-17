<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { TURNSTILE_ENABLED } from '../utils/appConfig'
import { env } from '../env'

interface Props {
  sitekey: string
  darkTheme: boolean
}
const props = withDefaults(defineProps<Props>(), {
  darkTheme: false,
})
const emit = defineEmits<{
  (e: 'verified', token: string): void
  (e: 'error'): void
  (e: 'expired'): void
}>()

const { locale } = useI18n()
const isEnabled = ref(TURNSTILE_ENABLED)
const widgetId = ref<string | null>(null)

const reset = () => {
  if (widgetId.value && window.turnstile && scriptLoaded.value) {
    window.turnstile.reset(widgetId.value)
  }
}
// const reset = () => {
//     if (widgetId.value && window.turnstile) {
//         window.turnstile.remove(widgetId.value)
//         renderTurnstile() // Re-render after reset
//     }
// }

const language = computed(() => {
  if (locale.value === 'ptBR') return 'pt'
  if (locale.value === 'frCA') return 'fr'
  return locale.value
})

// Expose reset functionality to parent component
defineExpose({
  reset,
})

const renderTurnstile = () => {
  if (!window.turnstile) return

  if (widgetId.value) {
    window.turnstile.remove(widgetId.value)
  }

  widgetId.value = window.turnstile.render('#turnstile-container', {
    sitekey: props.sitekey,
    callback: (token: string) => emit('verified', token),
    'error-callback': () => emit('error'),
    'expired-callback': () => emit('expired'),
    theme: props.darkTheme ? 'dark' : 'auto',
    size: 'flexible',
    language: language.value || 'auto',
  })
}

const scriptLoaded = ref(false)
const isMounted = ref(true)

onMounted(() => {
  // if (!isEnabled.value) return

  // const script = document.createElement('script')
  // script.src = env.turnstileScriptUrl
  // script.async = true
  // script.defer = true
  // script.onload = () => {
  //     // Render widget after script loads
  //     widgetId.value = window.turnstile.render('#turnstile-container', {
  //         sitekey: props.sitekey,
  //         callback: (token) => {
  //             emit('verified', token)
  //         },
  //         'error-callback': () => {
  //             emit('error')
  //         },
  //         'expired-callback': () => {
  //             emit('expired')
  //         },
  //         theme: props.darkTheme ? 'dark' : 'auto'
  //     })
  // }
  // document.head.appendChild(script)

  if (!isEnabled.value) return

  const script = document.createElement('script')
  script.src = env.turnstileScriptUrl
  script.async = true
  script.defer = true
  script.onerror = () => {
    if (isMounted.value) {
      emit('error')
    }
  }
  // Render widget after script loads
  script.onload = () => {
    if (!isMounted.value) return
    renderTurnstile()
    scriptLoaded.value = true
  }
  document.head.appendChild(script)
})

onUnmounted(() => {
  isMounted.value = false
  if (widgetId.value && window.turnstile) {
    window.turnstile.remove(widgetId.value)
    widgetId.value = null
  }
  const scripts = document.head.querySelectorAll(`script[src="${env.turnstileScriptUrl}"]`)
  scripts.forEach(script => script.remove())
})
</script>

<template>
  <div v-if="isEnabled" id="turnstile-container" class="turnstile-widget" />
</template>

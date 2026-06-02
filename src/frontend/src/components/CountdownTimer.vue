<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'

interface Props {
  timeLeftInSeconds?: number
}

const props = withDefaults(defineProps<Props>(), {
  timeLeftInSeconds: 0,
})

const emit = defineEmits(['CountdownProgressUpdate', 'OneMinuteLeftWarning', 'CountdownCompleted'])

let interval: ReturnType<typeof setInterval> | null = null
const remainingTime = ref(props.timeLeftInSeconds)

// Track when the timer is supposed to end in absolute time
let endTime = 0

const formattedRemainingTime = computed(() => {
  const minutes = Math.floor(remainingTime.value / 60)
    .toString()
    .padStart(2, '0')
  const seconds = (remainingTime.value % 60).toString().padStart(2, '0')
  return `${minutes}:${seconds}`
})

const startCountdown = () => {
  if (interval) return // Prevent multiple intervals

  emit('CountdownProgressUpdate', true)

  // Calculate the exact timestamp when this timer should hit 0
  endTime = Date.now() + remainingTime.value * 1000

  interval = setInterval(() => {
    // Calculate remaining seconds based on the actual current time
    // Using Math.ceil instead of Math.round for ticks at 1000ms. Use Math.round for smaller intervals.
    const exactRemainingTime = Math.max(0, Math.ceil((endTime - Date.now()) / 1000))

    // Only proceed if the second has actually ticked down
    if (exactRemainingTime !== remainingTime.value) {
      remainingTime.value = exactRemainingTime

      if (remainingTime.value === 60) {
        emit('OneMinuteLeftWarning')
      }
    }

    // Handle countdown completion
    if (remainingTime.value <= 0) {
      stopCountdown()
      emit('CountdownCompleted')
    }
  }, 1000)
}

const stopCountdown = () => {
  if (interval !== null) {
    clearInterval(interval)
    interval = null
  }
  emit('CountdownProgressUpdate', false)
}

// Watch for changes in timeLeftInSeconds and reset the timer accordingly
// Note: If we try to add a feature to "Extend" a running timer, and the extended time is the same a the initial time, this watcher won't execute..
//  e.g. Start a timer with 120 secs and while the timer is in progress, props.timeLeftInSeconds 120 secs again. There is no change in value for props.timeLeftInSeconds.
// To fix the above: Check the hack done in - Dashboard.vue.onTimerResponse
watch(
  () => props.timeLeftInSeconds,
  newValue => {
    remainingTime.value = newValue
    stopCountdown()
    if (newValue > 0) {
      startCountdown()
    }
  }
)

onMounted(() => {
  if (remainingTime.value > 0) {
    startCountdown()
  }
})

onUnmounted(() => {
  stopCountdown()
})
</script>

<template>
  <div class="select-none" :class="remainingTime > 0 ? 'bg-red-700' : 'border-2 border-white-600'">
    {{ formattedRemainingTime }}
  </div>
</template>

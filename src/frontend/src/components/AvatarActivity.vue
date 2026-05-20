<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { TYPING_ACTIVITY_DISPLAY_TIMEOUT_MS } from '../utils/appConfig'
import Avatar from './Avatar.vue'
import { computed } from 'vue'

const props = defineProps<{
  name: string
  isTyping: boolean
}>()

const { t } = useI18n()

const titleText = computed(() =>
  props.isTyping ? t('common.typing', { name: props.name }) : props.name
)

// Keep the config duration, but use a slower speed for the "steady" pulse
const ringStyle = {
  '--burst-duration': `${TYPING_ACTIVITY_DISPLAY_TIMEOUT_MS}ms`,
  // '--slow-duration': '2500ms', // Adjust this for the "slow" rings
}
</script>

<template>
  <div
    class="relative inline-flex items-center justify-center"
    :style="ringStyle"
    :title="titleText"
    :aria-label="titleText"
    role="img"
  >
    <template v-if="isTyping">
      <!-- Initial Quick Burst Rings -->
      <span class="activity-ring burst-1" />
      <span class="activity-ring burst-2" />

      <!-- Persistent Slow Ring -->
      <span class="activity-ring slow-loop" />
    </template>

    <Avatar :name="name" :show-title="false" class="relative w-8 h-8" />
  </div>
</template>

<style scoped>
.activity-ring {
  position: absolute;
  width: 100%;
  height: 100%;
  border-radius: 9999px;
  border: 2px solid white;
  pointer-events: none;
  opacity: 0;
  animation-name: activity-pulse;
  animation-timing-function: cubic-bezier(0.4, 0, 0.6, 1);
  --slow-duration: calc(var(--burst-duration) * 1.25);
  /* will-change: transform, opacity; */
}

/* First ring: Starts immediately, happens once */
.burst-1 {
  animation-duration: var(--burst-duration);
  animation-iteration-count: 1;
}

/* Second ring: Slight delay for the "close rings" look, happens once */
.burst-2 {
  animation-duration: var(--burst-duration);
  /* animation-delay: 150ms; */
  animation-delay: calc(var(--burst-duration) * 0.075);
  animation-iteration-count: 1;
}

/* Persistent Ring: Longer duration, loops infinitely, starts after the burst */
.slow-loop {
  animation-duration: var(--slow-duration);
  /* animation-delay: 400ms; */
  animation-delay: calc(var(--burst-duration) * 0.2);
  animation-iteration-count: infinite;
}

@keyframes activity-pulse {
  0% {
    transform: scale(1);
    opacity: 0.8;
  }
  100% {
    transform: scale(1.6);
    opacity: 0;
  }
}
</style>

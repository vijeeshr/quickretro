<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { TYPING_ACTIVITY_DISPLAY_TIMEOUT_MS } from '../utils/appConfig';
import Avatar from './Avatar.vue'
import { computed } from 'vue';

const props = defineProps<{
    name: string
    isTyping: boolean
}>()

const { t } = useI18n()

const titleText = computed(() =>
    props.isTyping
        ? t('common.typing', { name: props.name })
        : props.name
)

const ringStyle = {
    '--activity-duration': `${TYPING_ACTIVITY_DISPLAY_TIMEOUT_MS}ms`,
}
</script>

<template>
    <div class="relative inline-flex items-center justify-center" :style="ringStyle" :title="titleText"
        :aria-label="titleText" role="img">
        <span v-if="isTyping" class="activity-ring" />
        <Avatar :name="name" :show-title="false" class="relative w-8 h-8" />
    </div>
</template>

<style scoped>
/* Ensure the ring is a circle and doesn't push other elements */
.activity-ring {
    position: absolute;
    width: 100%;
    height: 100%;
    border-radius: 9999px;
    border: 2px solid white;
    pointer-events: none;
    /* animation: activity-pulse var(--activity-duration) cubic-bezier(0.4, 0, 0.6, 1) infinite; */
    animation-name: activity-pulse;
    animation-duration: var(--activity-duration);
    animation-timing-function: cubic-bezier(0.4, 0, 0.6, 1);
    animation-iteration-count: 1;
    animation-fill-mode: forwards;
}

@keyframes activity-pulse {
    0% {
        transform: scale(1);
        opacity: 1.0;
    }

    100% {
        transform: scale(1.4);
        opacity: 0;
    }
}
</style>
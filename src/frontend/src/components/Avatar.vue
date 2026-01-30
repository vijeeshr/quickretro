<script setup lang="ts">
import { computed } from 'vue'

interface Props {
    name?: string
    viewType?: 'Default' | 'Badge'
    isTyping?: boolean
}

const props = withDefaults(defineProps<Props>(), {
    name: 'Anonymous',
    viewType: 'Default',
    isTyping: false
})

const avatarText = computed(() => {
    if (!props.name) return ''

    if (props.viewType === 'Badge') {
        return props.name.length > 16 ? props.name.slice(0, 16) + '..' : props.name
    }

    let n = props.name.trim().split(/\s+/)
    if (n.length >= 2) return `${n[0][0].toUpperCase()}${n[n.length - 1][0].toUpperCase()}`
    return n[0][0].toUpperCase()
})

const avatarColor = computed(() => {
    if (props.name) {
        let hash = 0
        const saturation = 50, lightness = 60
        for (let i = 0; i < props.name.length; i++) {
            hash = props.name.charCodeAt(i) + ((hash << 5) - hash)
        }
        let h = hash % 360
        return `hsl(${h}, ${saturation}%, ${lightness}%)`
    }
    return 'hsl(0, 0%, 100%)'
})
</script>

<!-- <template>
    <div class="inline-flex items-center justify-center overflow-hidden"
        :class="viewType === 'Badge' ? 'rounded-md px-3 py-1' : 'rounded-full'" :title="name"
        :style="{ backgroundColor: avatarColor }">
        <span class="font-medium text-xs text-white select-none">{{ avatarText }}</span>
    </div>
</template> -->

<template>
    <div v-if="viewType === 'Default'" class="relative inline-flex">
        <!-- Activity ring -->
        <span v-if="isTyping" class="absolute inset-0 rounded-full ring-2 ring-white animate-activity-ring" />

        <!-- Avatar -->
        <div class="relative inline-flex h-8 w-8 items-center justify-center overflow-hidden rounded-full" :title="name"
            :style="{ backgroundColor: avatarColor }">
            <span class="font-medium text-xs text-white select-none">
                {{ avatarText }}
            </span>
        </div>
    </div>

    <!-- Badge mode (no animation, no wrapper) -->
    <div v-else class="inline-flex items-center justify-center overflow-hidden rounded-md px-3 py-1" :title="name"
        :style="{ backgroundColor: avatarColor }">
        <span class="font-medium text-xs text-white select-none">
            {{ avatarText }}
        </span>
    </div>
</template>


<style lang="css" scoped>
@keyframes activity-ring {
    0% {
        opacity: 0.4;
        transform: scale(1);
    }

    50% {
        opacity: 0.8;
        transform: scale(1.15);
    }

    100% {
        opacity: 0.4;
        transform: scale(1);
    }
}

.animate-activity-ring {
    animation: activity-ring 2s ease-in-out infinite;
}
</style>

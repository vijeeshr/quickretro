<script setup lang="ts">
import { computed } from 'vue'

interface Props {
    name?: string;
    viewType?: 'Default' | 'Badge';
}

const props = withDefaults(defineProps<Props>(), {
    name: 'Anonymous',
    viewType: 'Default'
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

<template>
    <div class="inline-flex items-center justify-center overflow-hidden"
        :class="viewType === 'Badge' ? 'rounded-md px-3 py-1' : 'rounded-full'" :title="name"
        :style="{ backgroundColor: avatarColor }">
        <span class="font-medium text-xs text-white select-none">{{ avatarText }}</span>
    </div>
</template>
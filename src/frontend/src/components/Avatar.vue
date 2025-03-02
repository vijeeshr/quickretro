<script setup lang="ts">
import { computed } from 'vue'

interface Props {
    name?: string
}

const props = withDefaults(defineProps<Props>(), {
    name: 'Anonymous'
})

const avatarText = computed(() => {
    if (props.name) {
        let n = props.name.trim().split(/\s+/)
        if (n && n.length >= 2) return `${n[0][0].toUpperCase()}${n[n.length - 1][0].toUpperCase()}`
        if (n && n.length == 1) return `${n[0][0].toUpperCase()}`
    }
    return ''
})

const avatarColor = computed(() => {
    if (props.name) {
        let hash = 0
        const saturation = 50, lightness = 60
        for (let i = 0; i < props.name.length; i++) {
            hash = props.name.charCodeAt(i) + ((hash << 5) - hash)
        }
        let h = hash % 360
        return 'hsl(' + h + ',' + saturation + '%,' + lightness + '%)'
    }
    return 'hsl(0,0,100)'
})
</script>

<template>
    <div class="inline-flex items-center justify-center overflow-hidden rounded-full" :title="name"
        :style="{ backgroundColor: avatarColor }">
        <span class="font-medium text-xs cursor-default text-white">{{ avatarText }}</span>
    </div>
</template>
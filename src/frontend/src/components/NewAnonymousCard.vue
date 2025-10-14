<script setup lang="ts">
import { DraftMessage } from '../models/DraftMessage';
import NewCard from './NewCard.vue';

const props = defineProps<{ category: string; by: string; nickname: string; board: string }>()
const emit = defineEmits(['added', 'invalidContent', 'discard'])

const handleAdded = (payload: DraftMessage) => {
    // Create a shallow copy to avoid mutating the original payload
    const modifiedPayload: DraftMessage = {
        ...payload,
        anon: true // Modify the property you need to change
    }
    emit('added', modifiedPayload)
}
</script>

<template>
    <NewCard v-bind="props" @added="handleAdded" @invalidContent="emit('invalidContent', $event)"
        @discard="emit('discard')" />
</template>
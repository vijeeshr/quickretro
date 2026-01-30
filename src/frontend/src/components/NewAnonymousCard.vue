<script setup lang="ts">
import { DraftMessage } from '../models/DraftMessage';
import { MessageMode } from '../models/MessageMode';
import NewCard from './NewCard.vue';

const props = defineProps<{ category: string, nickname: string }>()
const emit = defineEmits(['added', 'invalidContent', 'discard'])

const mode: MessageMode = 'anonymous'

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
    <NewCard v-bind="props" :mode="mode" @added="handleAdded" @invalidContent="emit('invalidContent', $event)"
        @discard="emit('discard')" />
</template>
<script setup lang="ts">
import Card from './Card.vue';
import { MessageResponse } from '../models/Requests';
import { BoardColumn } from '../models/BoardColumn';

type Props = {
    card: MessageResponse
    mask: boolean
    updateable: boolean
    locked: boolean
    currentUser: string
    currentUserNickname: string
    board: string
    categories: BoardColumn[]
};

const props = defineProps<Props>()
const emit = defineEmits(['updated', 'deleted', 'liked', 'invalidContent', 'categoryChanged'])
</script>

<template>
    <Card v-bind="props" @updated="emit('updated', $event)" @deleted="emit('deleted', $event)"
        @liked="emit('liked', $event)" @invalidContent="emit('invalidContent', $event)"
        @categoryChanged="emit('categoryChanged', $event)" />
</template>

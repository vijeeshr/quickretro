<script setup lang="ts">
import { logMessage } from '../utils'
import { DraftMessage } from '../models/DraftMessage';
import { useContentEditableLimiter } from '../composables/useContentEditableLimiter';

const props = defineProps<{ parentId: string, category: string, locked: boolean, by: string, nickname: string, board: string }>()
const emit = defineEmits(['added', 'invalidContent'])


const add = (event: Event) => {
    if (props.locked) {
        logMessage('Locked! Cannot add comment.')
        return
    }

    const el = event.target as HTMLElement
    const msg = el.innerText.trim()

    if (
        msg.length === 0 ||
        /^[\u0000\u200B\u200C\u200D\uFEFF]*$/.test(msg) // Check if contains only zero-width chars or null bytes
    ) {
        logMessage('discarding empty new comment...')
        el.innerText = ''
        return
    }

    const payload: DraftMessage = {
        id: crypto.randomUUID(),
        msg: msg,
        cat: props.category,
        anon: false,
        pid: props.parentId
    }

    emit('added', payload)

    el.innerText = ''
}

const addOnEnter = (event: KeyboardEvent) => {
    if (!event.shiftKey) {
        add(event)
        event.preventDefault()
    }
}

const { onInput } = useContentEditableLimiter({
    user: () => props.by,
    nickname: () => props.nickname,
    board: () => props.board,
    category: () => props.category,
    isComment: true,
    onInvalid: (msg) => emit('invalidContent', msg)
})

const vFocus = {
    mounted: (el: HTMLElement) => {
        el.focus()
    }
}
</script>

<template>
    <article v-focus
        class="w-full mt-2 border dark:border-gray-400 rounded-lg p-2 text-sm resize-none text-gray-500 dark:text-white min-h-[3.5rem] break-words focus:outline-none cursor-auto focus:border-sky-400 dark:focus:border-white"
        :contenteditable="!locked" @blur="add" @keydown.enter="addOnEnter" @input="onInput"></article>
</template>
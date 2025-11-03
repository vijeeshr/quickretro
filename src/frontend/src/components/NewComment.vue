<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { assertMessageContentValidation, canAssertMessageContentValidation, logMessage, MessageContentValidationResult } from '../utils'
import { DraftMessage } from '../models/DraftMessage';

const { t } = useI18n()
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

const validate = (event: Event) => {
    if (!canAssertMessageContentValidation()) return
    const validationResult: MessageContentValidationResult = assertMessageContentValidation(event, props.by, props.nickname, props.board, props.category, true)
    if (validationResult.isValid) return

    let errorMessage: string = t('common.contentOverloadError')
    if (validationResult.isTrimmed) errorMessage = t('common.contentStrippingError')

    emit('invalidContent', errorMessage)
}
</script>

<template>
    <article
        class="w-full mt-2 border rounded-lg p-2 text-sm resize-none text-gray-500 dark:text-white min-h-[3.5rem] break-words focus:outline-none cursor-auto"
        :contenteditable="!locked" @blur="add" @keydown.enter="addOnEnter" @input="validate"></article>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import { DraftMessage } from '../models/DraftMessage';
import { assertMessageContentValidation, canAssertMessageContentValidation, MessageContentValidationResult } from '../utils';
import { useI18n } from 'vue-i18n';

const { t } = useI18n()
const props = defineProps<{ category: string, by: string, nickname: string, board: string }>()
const emit = defineEmits(['added', 'invalidContent', 'discard'])

const editing = ref(true)

const add = (event: Event) => {
  if (editing.value) {
    editing.value = false
    const msg = (event.target as HTMLElement).innerText.trim()
    
    if (
      msg.length === 0 ||
      /^[\u0000\u200B\u200C\u200D\uFEFF]*$/.test(msg) // Check if contains only zero-width chars or null bytes
    ) {
      emit('discard')
      return
    }

    const payload: DraftMessage = {
      id: crypto.randomUUID(),
      msg: msg,
      cat: props.category,
      anon: false,
      pid: ''
    }

    emit('added', payload)
  }
}

const addOnEnter = (event: KeyboardEvent) => {
  if (!event.shiftKey) {
    add(event)
    event.preventDefault()
  }
}

const vFocus = {
  mounted: (el: HTMLElement) => {
    el.focus()
  }
}

const validate = (event: Event) => {
  if (!canAssertMessageContentValidation()) return
  const validationResult: MessageContentValidationResult = assertMessageContentValidation(event, props.by, props.nickname, props.board, props.category)
  if (validationResult.isValid) return

  let errorMessage: string = t('common.contentOverloadError')
  if (validationResult.isTrimmed) errorMessage = t('common.contentStrippingError')

  emit('invalidContent', errorMessage)
}
</script>

<template>
  <div class="bg-white dark:bg-gray-700 rounded-lg p-3 mb-2 shadow-xl border border-sky-400 dark:border-white">

    <div class="text-gray-500 dark:text-white pb-2">
      <article v-focus class="min-h-[3.5rem] text-center break-words focus:outline-none cursor-auto"
        contenteditable="true" @blur="add" @keydown.enter="addOnEnter" @input="validate"></article>
    </div>

  </div>
</template>
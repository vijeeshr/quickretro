<script setup lang="ts">
import { ref } from 'vue';
import { DraftMessage } from '../models/DraftMessage';
import { assertMessageContentValidation, canAssertMessageContentValidation, MessageContentValidationResult } from '../utils';

const props = defineProps<{ category: string, by: string, nickname: string, board: string }>()
const emit = defineEmits(['added', 'invalidContent'])

const editing = ref(true)

const add = (event: Event) => {
  if (editing.value) {
    editing.value = false
    const payload: DraftMessage = {
      id: crypto.randomUUID(),
      msg: (event.target as HTMLElement).innerText.trim(),
      cat: props.category
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

  let errorMessage: string = "Content more than allowed limit."
  if (validationResult.isTrimmed) errorMessage = 'Content more than allowed limit. Extra text is stripped from the end.'

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
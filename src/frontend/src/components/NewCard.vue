<script setup lang="ts">
import { ref } from 'vue';
import { DraftMessage } from '../models/DraftMessage';
import { useContentEditableLimiter } from '../composables/useContentEditableLimiter';

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

const { onInput } = useContentEditableLimiter({
  user: () => props.by,
  nickname: () => props.nickname,
  board: () => props.board,
  category: () => props.category,
  isComment: false,
  onInvalid: (msg) => emit('invalidContent', msg)
})

const vFocus = {
  mounted: (el: HTMLElement) => {
    el.focus()
  }
}
</script>

<template>
  <div class="bg-white dark:bg-gray-700 rounded-lg p-3 mb-2 shadow-xl border border-sky-400 dark:border-white">

    <div class="text-gray-500 dark:text-white pb-2">
      <article v-focus class="min-h-[3.5rem] text-center break-words focus:outline-none cursor-auto"
        contenteditable="true" @blur="add" @keydown.enter="addOnEnter" @input="onInput"></article>
    </div>

  </div>
</template>
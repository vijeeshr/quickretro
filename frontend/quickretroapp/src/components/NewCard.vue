<script setup lang="ts">
import { ref } from 'vue';
import { DraftMessage } from '../models/DraftMessage';

const props = defineProps<{ category: string }>()
const emit = defineEmits(['added'])

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
</script>

<template>
  <div class="bg-white rounded-lg p-3 mb-2 shadow-xl border border-sky-400">

    <div class="text-gray-500 pb-2">
      <article v-focus class="min-h-[3.5rem] text-center break-words focus:outline-none cursor-auto"
        contenteditable="true" @blur="add" @keydown.enter="addOnEnter"></article>
    </div>

  </div>
</template>
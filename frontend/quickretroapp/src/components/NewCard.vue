<script setup lang="ts">
import { ref } from 'vue';
import { CardModel } from './CardModel';

const props = defineProps<{ category: string }>()
const emit = defineEmits(['added'])

const editing = ref(true)

const add = (event: Event) => {
  if (editing.value) {
    editing.value = false 
    const payload: CardModel = {
      typ: 'msg',
      id: crypto.randomUUID(),
      nickname: localStorage.getItem('nickname') || '',
      msg: (event.target as HTMLElement).innerText.trim(),
      cat: props.category,
      likes: '0',
      liked: false,
      mine: true
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
        <article v-focus 
          class="min-h-4 text-center break-words focus:outline-none cursor-auto" contenteditable="true" 
          @blur="add"
          @keydown.enter="addOnEnter"></article>
    </div>

</div>
</template>
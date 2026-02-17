<script setup lang="ts">
import { computed, ref } from 'vue'
import { DraftMessage } from '../models/DraftMessage'
import { useContentEditableLimiter } from '../composables/useContentEditableLimiter'
import { MessageMode } from '../models/MessageMode'
import { useTypingTrigger } from '../composables/useTypingTrigger'

const props = withDefaults(
  defineProps<{
    category: string
    nickname: string
    mode?: MessageMode
  }>(),
  {
    mode: 'regular',
  }
)

const emit = defineEmits(['added', 'invalidContent', 'discard', 'typing'])

const isAnon = computed(() => props.mode === 'anonymous')
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
      anon: isAnon.value,
      pid: '',
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
  nickname: () => props.nickname,
  category: () => props.category,
  anon: () => isAnon.value,
  isComment: false,
  onInvalid: msg => emit('invalidContent', msg),
})

const { triggerTyping } = useTypingTrigger(emit)

const onKeyDown = (event: KeyboardEvent) => {
  if (!isAnon.value) {
    // Trigger the throttled typing event
    triggerTyping(event)
  }
}

const vFocus = {
  mounted: (el: HTMLElement) => {
    el.focus()
  },
}
</script>

<template>
  <div
    class="bg-white dark:bg-gray-700 rounded-lg p-3 mb-2 shadow-xl border border-sky-400 dark:border-white"
  >
    <div class="text-gray-500 dark:text-white pb-2">
      <article
        v-focus
        class="min-h-[3.5rem] text-center break-words focus:outline-none cursor-auto"
        contenteditable="true"
        @blur="add"
        @keydown.enter="addOnEnter"
        @keydown="onKeyDown"
        @input="onInput"
      ></article>
    </div>
  </div>
</template>

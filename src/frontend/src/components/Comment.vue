<script setup lang="ts">
import { computed, nextTick, ref } from 'vue';
import { MessageResponse } from '../models/Requests'
import { logMessage } from '../utils';
import { DraftMessage } from '../models/DraftMessage';
import { useContentEditableLimiter } from '../composables/useContentEditableLimiter';

// "currentUserNickname", "comment.cat" only used to calculate message size in bytes. "category" already passed with "comment.cat".
// Message size calc and trimming only done when editing.
// Only a message's content and category are actually updated in the backend
interface Props {
    comment: MessageResponse // Todo: Change name of model. Or use a different model.
    mask: boolean
    canManage: boolean
    locked: boolean
    currentUserNickname: string // Only used for message size calc
}
const props = defineProps<Props>()
const emit = defineEmits(['updated', 'deleted', 'invalidContent', 'discard'])

const editing = ref(false)

const manageable = computed(() => props.comment.mine || props.canManage)

const content = computed(() => {
    if (props.mask && !props.comment.mine && props.comment.msg) {
        return props.comment.msg.slice(0).replaceAll(/[^ ]/g, '*');
    }
    return props.comment.msg
})

const edit = async (event: Event) => {
    if (props.locked) {
        logMessage("Locked! Cannot edit.")
        return
    }
    // Todo: figure out what's going on here. Simplify the condition to remove dups.
    if (!editing.value && props.comment.mine) {
        editing.value = true;
        await nextTick();
        (event.target as HTMLElement).focus();
    }
    if (!editing.value && props.comment.mine && !props.mask) {
        editing.value = true;
        await nextTick();
        (event.target as HTMLElement).focus();
    }
}

const save = (event: Event) => {
    if (props.locked) {
        if (editing.value) {
            // Board lock received when the user is editing a message.
            // The message will be forcefully discarded.
            const target = event.target as HTMLElement
            target.innerText = content.value // Discard new changes by reseting value back to original..TODO: Check what happens when Mask is ON..the *** shouldn't be updated.
            editing.value = false
            emit('discard')
            return
        }
        logMessage("Locked! Cannot save.")
        return
    }
    if (editing.value && props.comment.mine) {
        editing.value = false
        if (props.comment.msg !== (event.target as HTMLElement).innerText.trim()) {
            const payload: DraftMessage = {
                id: props.comment.id,
                msg: (event.target as HTMLElement).innerText.trim(),
                cat: props.comment.cat,
                anon: props.comment.anon,
                pid: props.comment.pid
            }
            emit('updated', payload)
        } else {
            logMessage('No content change. Not dispatching!!')
            // logMessage('[old value]', props.card.msg, '[new value]', (event.target as HTMLElement).innerText.trim())
        }
    }
}

const saveOnEnter = (event: KeyboardEvent) => {
    if (props.locked) {
        logMessage("Locked! Cannot edit.")
        event.preventDefault()
        return
    }
    if (!event.shiftKey) {
        save(event)
        event.preventDefault()
    }
}

const { onInput } = useContentEditableLimiter({
    nickname: () => props.currentUserNickname,
    category: () => props.comment.cat,
    anon: () => props.comment.anon,
    isComment: true,
    onInvalid: (msg) => emit('invalidContent', msg)
})

const remove = () => {
    if (props.locked) {
        logMessage("Locked! Cannot delete.")
        return
    }
    if (manageable) {
        emit('deleted', props.comment.id)
    }
}
</script>

<template>
    <div class="bg-gray-50 dark:bg-gray-600 p-2 rounded-xl text-sm border" :class="editing && comment.mine
        ? 'border-sky-400 dark:border-white'
        : 'border-transparent'">
        <div class="font-semibold break-words text-gray-700 dark:text-white">{{ comment.nickname }}</div>
        <article class="text-gray-600 dark:text-white min-h-4 break-words focus:outline-none" :class="[
            { 'blur-sm': mask && !comment.mine },
            editing
                ? 'cursor-auto'
                : comment.mine && !locked
                    ? 'cursor-pointer'
                    : 'cursor-default'
        ]" :contenteditable="editing && comment.mine && !(locked && editing)" @click="edit" @blur="save"
            @keydown.enter="saveOnEnter" @input="onInput">
            {{ content }}</article>
        <div class="flex" v-if="manageable">
            <!-- Delete comment button -->
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
                stroke="currentColor" class="w-4 h-4 cursor-pointer text-gray-600 dark:text-gray-200 ml-auto"
                @click="remove">
                <path stroke-linecap="round" stroke-linejoin="round"
                    d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
            </svg>
        </div>
    </div>

</template>
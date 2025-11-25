<script setup lang="ts">
import { computed } from 'vue';
import { BoardColumn } from '../models/BoardColumn';
import { useI18n } from 'vue-i18n';

interface Props {
    column: BoardColumn,
    // color?: string,
    // buttonText?: string,
    width?: string,
    buttonHighlight: boolean,
    anonymousButtonHighlight: boolean,
    editable: boolean,
    locked: boolean
}

const props = withDefaults(defineProps<Props>(), {
    // color: 'green',
    // buttonText: 'Add',
    width: 'w-1/3',
    buttonHighlight: false,
    anonymousButtonHighlight: false
})

const emit = defineEmits(['addCard', 'addAnonymousCard', 'categoryClick'])

const { t } = useI18n()

const displayText = computed(() => {
    if (props.column.isDefault) return t(`dashboard.columns.${props.column.id}`)
    return props.column.text
})

const onCategoryClick = () => {
    if (!props.locked) {
        emit('categoryClick')
    }
}

// const emit = defineEmits(['inFocus'])
// function buttonClick() {
//     emit('inFocus')
// }    

</script>

<template>
    <div class="p-1" :class="[`w-full md:px-6 md:pb-6 md:pt-2 md:${width} min-w-0`]">

        <div class="grid grid-cols-2 gap-1 mb-2 min-w-0">
            <div class="col-span-2 flex items-center justify-center p-1 text-sm w-full font-semibold rounded-md border select-none break-words"
                :class="[
                    `bg-${column.color}-100`,
                    `border-${column.color}-300`,
                    `text-${column.color}-600`,
                    `dark:bg-${column.color}-800`,
                    `dark:border-${column.color}-700`,
                    `dark:text-${column.color}-100`,
                    props.editable && !props.locked && [
                        'cursor-pointer',
                        `hover:bg-${column.color}-400`,
                        `hover:text-white`,
                        `dark:hover:bg-${column.color}-600`,
                    ]
                ]" @click="onCategoryClick">{{ displayText }}</div>
            <button
                class="rounded-lg border font-bold bg-gray-50 dark:bg-white/30 hover:bg-gray-200 dark:hover:bg-white/40 select-none p-1 shadow-md"
                :class="{ 'border-sky-400 dark:border-white text-sky-400 hover:text-sky-600 dark:text-white': buttonHighlight, 'border-gray-300 dark:border-white/20 text-gray-600 hover:text-gray-700 dark:text-white': !buttonHighlight }"
                @click="$emit('addCard')">
                +
            </button>
            <button
                class="rounded-lg border font-semibold bg-gray-50 dark:bg-white/30 hover:bg-gray-200 dark:hover:bg-white/40 flex items-center justify-center p-1 shadow-md"
                :class="{ 'border-sky-400 dark:border-white text-sky-400 hover:text-sky-600 dark:text-white': anonymousButtonHighlight, 'border-gray-300 dark:border-white/20 text-gray-500 hover:text-gray-700 dark:text-white': !anonymousButtonHighlight }"
                @click="$emit('addAnonymousCard')">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                    stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-5 h-5">
                    <path d="M14 18a2 2 0 0 0-4 0" />
                    <path
                        d="m19 11-2.11-6.657a2 2 0 0 0-2.752-1.148l-1.276.61A2 2 0 0 1 12 4H8.5a2 2 0 0 0-1.925 1.456L5 11" />
                    <path d="M2 11h20" />
                    <circle cx="17" cy="18" r="3" />
                    <circle cx="7" cy="18" r="3" />
                </svg>
            </button>
        </div>

        <slot></slot>
    </div>
</template>
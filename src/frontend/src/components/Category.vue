<script setup lang="ts">

interface Props {
    color?: string,
    buttonText?: string,
    width?: string,
    buttonHighlight: boolean,
    anonymousButtonHighlight: boolean
}

withDefaults(defineProps<Props>(), {
    color: 'green',
    buttonText: 'Add',
    width: 'w-1/3',
    buttonHighlight: false,
    anonymousButtonHighlight: false
})

defineEmits(['addCard', 'addAnonymousCard'])

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
                    `bg-${color}-100`,
                    `border-${color}-300`,
                    `text-${color}-600`,
                    `dark:bg-${color}-800`,
                    `dark:border-${color}-700`,
                    `dark:text-${color}-100`
                ]">{{ buttonText }}</div>
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
                    <path stroke="none" d="M0 0h24v24H0z" fill="none" />
                    <path d="M3 11h18" />
                    <path d="M5 11v-4a3 3 0 0 1 3 -3h8a3 3 0 0 1 3 3v4" />
                    <path d="M7 17m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0" />
                    <path d="M17 17m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0" />
                    <path d="M10 17h4" />
                </svg>
            </button>
        </div>

        <slot></slot>
    </div>
</template>
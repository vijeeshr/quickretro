<script setup lang="ts">

interface Props {
    color?: string,
    buttonText?: string,
    width?: string
}

withDefaults(defineProps<Props>(), {
    color: 'green',
    buttonText: 'Add',
    width: 'w-1/3'
})

defineEmits(['addCard'])

// const emit = defineEmits(['inFocus'])
// function buttonClick() {
//     emit('inFocus')
// }    

</script>

<template>
    <div class="p-6" :class="[`w-full md:${width}`]">
        <button
            class="px-4 py-1 text-sm w-full font-semibold rounded-lg mb-4 border hover:border-transparent focus:outline-none select-none break-words"
            :class="[
                `bg-${color}-100 hover:bg-${color}-400`,
                `border-${color}-300`,
                `text-${color}-600 hover:text-white`,
                `dark:bg-${color}-800 dark:hover:bg-${color}-600`,
                `dark:border-${color}-700`,
                `dark:text-${color}-100`
            ]" @click="$emit('addCard')">
            {{ buttonText }}
        </button>
        <slot></slot>
    </div>
</template>
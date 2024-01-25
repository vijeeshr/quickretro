<script setup lang="ts">

    interface Props {
        name: string, 
        color?: string,
        buttonText?: string 
    }

    withDefaults(defineProps<Props>(), {
        color: 'green',
        buttonText: 'Add' 
    })
</script>

<template>
<div class="flex-none w-1/3 p-6">
    <button class="px-4 py-1 text-sm w-full font-semibold rounded-lg mb-4 border hover:text-white hover:border-transparent focus:outline-none select-none"
            :class="[`bg-${color}-100`, `hover:bg-${color}-400`, `border-${color}-300`, `text-${color}-600`]">
        {{ buttonText }}
    </button>
    <slot></slot>
</div>    
</template>
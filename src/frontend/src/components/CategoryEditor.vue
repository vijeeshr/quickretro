<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { CategoryDefinition } from "../models/CategoryDefinition"

const props = defineProps<{ categories: CategoryDefinition[] }>()

const emit = defineEmits([
    'category-text-update',
    'category-toggle',
    'categories-reorder',
    'valid'
])

const MAX_CATEGORY_TEXT_LENGTH: number = Number(import.meta.env.VITE_MAX_CATEGORY_TEXT_LENGTH)

const { t } = useI18n()

const localCategories = computed(() => {
    return props.categories.map(col => ({ ...col }))
})

const isCategorySelectionValid = computed(() => {
    return localCategories.value.some(c => c.enabled === true)
})

const dragSourceIndex = ref<number | null>(null) // Drag state

// Drag handlers
const onDragStart = (index: number) => {
    dragSourceIndex.value = index
}

// Fired when dragging over another item
const onDragOver = (event: DragEvent) => {
    event.preventDefault() // allow dropping
}

const onDrop = (targetIndex: number) => {
    if (dragSourceIndex.value === null || dragSourceIndex.value === targetIndex) return
    // Move item in array
    const updatedCategories = [...localCategories.value]
    const moved = updatedCategories.splice(dragSourceIndex.value, 1)[0]
    updatedCategories.splice(targetIndex, 0, moved)
    // Update positions
    updatedCategories.forEach((c, i) => (c.pos = i + 1))
    // Reset drag index
    dragSourceIndex.value = null
    emit('categories-reorder', updatedCategories)
}

const toggleCategory = (id: string) => {
    const cat = localCategories.value.find(c => c.id === id)
    if (cat) {
        emit('category-toggle', {
            id: id,
            enabled: !cat.enabled
        })
        emit('valid', isCategorySelectionValid.value)
    }
}

const updateCategoryText = (id: string, event: Event) => {
    let value = (event.target as HTMLInputElement)?.value ?? ""
    if (value.length > MAX_CATEGORY_TEXT_LENGTH) {
        value = value.slice(0, MAX_CATEGORY_TEXT_LENGTH)
    }
    emit('category-text-update', {
        id: id,
        text: value
    })
}
</script>

<template>
    <ul class="space-y-2 text-sm">
        <li v-for="(cat, index) in localCategories" :key="cat.id" class="flex space-x-1" draggable="true"
            @dragstart="onDragStart(index)" @dragover="onDragOver" @drop="onDrop(index)">
            <!-- Drag handle -->
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2"
                stroke="currentColor" class="w-6 h-6 text-gray-900 dark:text-gray-400 cursor-move">
                <path stroke-linecap="round" stroke-linejoin="round" d="M8 9h.01M16 9h.01M8 15h.01M16 15h.01" />
            </svg>

            <!-- Enable/Disable Button -->
            <button v-if="cat.enabled" @click="toggleCategory(cat.id)">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-6 h-6"
                    :class="cat.colorClass">
                    <path fill-rule="evenodd"
                        d="M2.25 12c0-5.385 4.365-9.75 9.75-9.75s9.75 4.365 9.75 9.75-4.365 9.75-9.75 9.75S2.25 17.385 2.25 12Zm13.36-1.814a.75.75 0 1 0-1.22-.872l-3.236 4.53L9.53 12.22a.75.75 0 0 0-1.06 1.06l2.25 2.25a.75.75 0 0 0 1.14-.094l3.75-5.25Z"
                        clip-rule="evenodd" />
                </svg>
            </button>

            <button v-else @click="toggleCategory(cat.id)">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor"
                    class="w-6 h-6 text-gray-500">
                    <path fill-rule="evenodd"
                        d="M12 2.25c-5.385 0-9.75 4.365-9.75 9.75s4.365 9.75 9.75 9.75 9.75-4.365 9.75-9.75S17.385 2.25 12 2.25Zm-1.72 6.97a.75.75 0 1 0-1.06 1.06L10.94 12l-1.72 1.72a.75.75 0 1 0 1.06 1.06L12 13.06l1.72 1.72a.75.75 0 1 0 1.06-1.06L13.06 12l1.72-1.72a.75.75 0 1 0-1.06-1.06L12 10.94l-1.72-1.72Z"
                        clip-rule="evenodd" />
                </svg>
            </button>

            <!-- Text Input -->
            <input type="text" :id="cat.id" :value="cat.text" :maxlength="MAX_CATEGORY_TEXT_LENGTH"
                @input="updateCategoryText(cat.id, $event)" :placeholder="t(`dashboard.columns.${cat.id}`)" class="w-full rounded-md focus:outline-none focus:border
                       focus:border-gray-200 focus:ring-gray-200
                       dark:text-slate-200 dark:bg-gray-900
                       dark:focus:border-gray-800 dark:focus:ring-gray-800" />
        </li>
    </ul>
    <p v-show="!isCategorySelectionValid" class="text-sm text-red-600 dark:text-red-300 mt-2 select-none">{{
        t('common.invalidColumnSelection') }}
    </p>
</template>
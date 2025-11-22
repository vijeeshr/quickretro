<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ColumnDefinition } from "../models/ColumnDefinition";

const props = defineProps<{ columns: ColumnDefinition[] }>()

// const emit = defineEmits<{
//     (e: "columns-update", columns: ColumnDefinition[]): void
// }>()
const emit = defineEmits(['columns-update'])

const { t } = useI18n()

const defaultColumns = computed(() => {
    const merged = props.columns.map((d) => {
        return {
            id: d.id,
            color: d.color,
            colorClass: d.colorClass,
            text: d.text,
            enabled: d.enabled,
            pos: d.pos
        }
    })

    return merged
})

// const mergedColumns = computed({
//     get() {
//         // Map props.columns by ID
//         const map = new Map<string, BoardColumn>(props.columns.map(c => [c.id, c]))

//         const merged = defaultColumns.value
//             .map((d) => {
//                 const override = map.get(d.id)
//                 return {
//                     id: d.id,
//                     color: d.color,
//                     colorClass: d.colorClass,
//                     text: override?.text ?? d.text,
//                     enabled: override !== undefined ? true : false, // override is present means, the column has been defined
//                     pos: override?.pos ?? defaultColumns.value.length
//                 }
//             })

//         merged.sort((a, b) => {
//             // sort by "enabled==true" first, then by "pos"
//             if (a.enabled !== b.enabled) {
//                 return a.enabled ? -1 : 1; // enabled=true first
//             }
//             return a.pos - b.pos;
//         })

//         // reassign pos
//         return merged.map((col, i) => ({
//             ...col,
//             pos: i + 1
//         }))
//     },

//     set(newValue) {
//         // Emit *only* the dynamic fields back up
//         const stripped: ColumnDefinition[] = newValue.map(col => ({
//             id: col.id,
//             text: col.text,
//             enabled: col.enabled,
//             pos: col.pos,
//             color: col.color,
//             colorClass: col.colorClass
//             // isDefault: col.isDefault,
//             // color: col.color  // include if parent stores this
//         }))

//         logMessage("MergedColumns_Set", stripped)
//         emit("columns-update", stripped)
//     }
// })

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
    const moved = defaultColumns.value.splice(dragSourceIndex.value, 1)[0]
    defaultColumns.value.splice(targetIndex, 0, moved)
    // Reset drag index
    dragSourceIndex.value = null
    // Update position numbers (for saving order later)
    defaultColumns.value.forEach((c, i) => (c.pos = i + 1))
    emit('columns-update', [...defaultColumns.value])
}

const toggleColumn = (colId: string) => {
    let col = defaultColumns.value.find(c => c.id === colId)
    if (col) {
        col.enabled = !col.enabled
    }
    emit('columns-update', [...defaultColumns.value])
}

const updateText = (colId: string, event: Event) => {
    const value = (event.target as HTMLInputElement)?.value ?? ""
    let col = defaultColumns.value.find(c => c.id === colId)
    if (col) {
        col.text = value.trim()
    }
    emit('columns-update', [...defaultColumns.value])
}

</script>

<template>
    <ul class="space-y-2 text-sm">
        <li v-for="(column, index) in defaultColumns" :key="column.id" class="flex space-x-1" draggable="true"
            @dragstart="onDragStart(index)" @dragover="onDragOver" @drop="onDrop(index)">
            <!-- Drag handle -->
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2"
                stroke="currentColor" class="w-6 h-6 text-gray-900 dark:text-gray-400 cursor-move">
                <path stroke-linecap="round" stroke-linejoin="round" d="M8 9h.01M16 9h.01M8 15h.01M16 15h.01" />
            </svg>

            <!-- Enable/Disable Button -->
            <button v-if="column.enabled" @click="toggleColumn(column.id)">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-6 h-6"
                    :class="column.colorClass">
                    <path fill-rule="evenodd"
                        d="M2.25 12c0-5.385 4.365-9.75 9.75-9.75s9.75 4.365 9.75 9.75-4.365 9.75-9.75 9.75S2.25 17.385 2.25 12Zm13.36-1.814a.75.75 0 1 0-1.22-.872l-3.236 4.53L9.53 12.22a.75.75 0 0 0-1.06 1.06l2.25 2.25a.75.75 0 0 0 1.14-.094l3.75-5.25Z"
                        clip-rule="evenodd" />
                </svg>
            </button>

            <button v-else @click="toggleColumn(column.id)">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor"
                    class="w-6 h-6 text-gray-500">
                    <path fill-rule="evenodd"
                        d="M12 2.25c-5.385 0-9.75 4.365-9.75 9.75s4.365 9.75 9.75 9.75 9.75-4.365 9.75-9.75S17.385 2.25 12 2.25Zm-1.72 6.97a.75.75 0 1 0-1.06 1.06L10.94 12l-1.72 1.72a.75.75 0 1 0 1.06 1.06L12 13.06l1.72 1.72a.75.75 0 1 0 1.06-1.06L13.06 12l1.72-1.72a.75.75 0 1 0-1.06-1.06L12 10.94l-1.72-1.72Z"
                        clip-rule="evenodd" />
                </svg>
            </button>

            <!-- Text Input -->
            <input type="text" :id="column.id" :value="column.text" @input="updateText(column.id, $event)"
                :placeholder="t(`dashboard.columns.${column.id}`)" class="w-full rounded-md focus:outline-none focus:border
                       focus:border-gray-200 focus:ring-gray-200
                       dark:text-slate-200 dark:bg-gray-900
                       dark:focus:border-gray-800 dark:focus:ring-gray-800" />
        </li>
    </ul>
</template>

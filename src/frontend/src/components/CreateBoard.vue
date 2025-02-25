<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { BoardColumn, CreateBoardRequest, createBoard } from '../api'

const router = useRouter()
const boardname = ref('')
const team = ref('')

const columns = ref([
    { id: "col01", text: "", defaultText: "Sails", color: "green", colorClass: "text-green-500", enabled: true },
    { id: "col02", text: "", defaultText: "Anchors", color: "red", colorClass: "text-red-500", enabled: true },
    { id: "col03", text: "", defaultText: "Next Steps", color: "yellow", colorClass: "text-yellow-500", enabled: true },
    { id: "col04", text: "", defaultText: "Shoutouts!", color: "fuchsia", colorClass: "text-fuchsia-500", enabled: false },
    { id: "col05", text: "", defaultText: "New Ideas!", color: "orange", colorClass: "text-orange-500", enable: false }
])

const toggleColumn = (column: string, enable: boolean) => {
    let cat = columns.value.find(c => c.id === column)
    if (cat) {
        cat.enabled = enable
    }
}

const isColumnSelectionValid = computed(() => {
    return columns.value.some(c => c.enabled === true)
})

const boardnameEntered = computed(() => {
    if (boardname.value && boardname.value.trim() !== '') return true
    return false
})

const create = async () => {
    // Todo: Throttle this.

    const selectedColumns: BoardColumn[] = columns.value.filter(c => c.enabled === true)
        .map(c => ({
            id: c.id,
            text: c.text || c.defaultText,
            color: c.color
        }))

    const payload: CreateBoardRequest = {
        name: boardname.value,
        team: team.value,
        owner: localStorage.getItem('user') || '',
        columns: selectedColumns
    }

    try {
        const createdBoard = await createBoard(payload)
        router.push(`/board/${createdBoard.id}`)
    } catch (error) {
        console.error('Error creating board:', error);
    }
}

onMounted(() => {
    const isDark = ref(localStorage.getItem("theme") === "dark")
    document.documentElement.classList.toggle("dark", isDark.value)
})

</script>

<template>
    <div class="bg-gray-100 dark:bg-gray-950 flex h-screen items-center justify-center p-4">
        <div class="w-full max-w-md">
            <div class="bg-white dark:bg-gray-900 shadow-md rounded-md p-8">
                <h2 class="mt-6 text-center text-3xl font-bold tracking-tight text-gray-600 dark:text-gray-400">
                    Create Board
                </h2>
                <div class="space-y-6 mt-4">
                    <div>
                        <div class="mt-1">
                            <input v-model.trim="boardname" name="name" type="text" placeholder="Type board name here!"
                                required autofocus
                                class="px-2 py-2 mt-1 block w-full rounded-md border border-gray-300 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-sky-500 sm:text-sm dark:bg-slate-800 dark:text-slate-200" />
                        </div>
                        <p v-show="!boardnameEntered" name="name-required"
                            class="text-sm text-red-600 dark:text-red-300 mt-2">Please
                            enter
                            board name</p>
                    </div>
                    <div>
                        <div class="mt-1">
                            <input v-model.trim="team" name="team" type="text" placeholder="Type team name here!"
                                class="px-2 py-2 mt-1 block w-full rounded-md border border-gray-300 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-sky-500 sm:text-sm dark:bg-slate-800 dark:text-slate-200" />
                        </div>
                    </div>
                    <div>
                        <ul class="space-y-3 text-sm">
                            <li v-for="column in columns" class="flex space-x-3">
                                <button v-if="column.enabled" @click="toggleColumn(column.id, false)">
                                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor"
                                        class="w-6 h-6" :class="[`${column.colorClass}`]">
                                        <path fill-rule="evenodd"
                                            d="M2.25 12c0-5.385 4.365-9.75 9.75-9.75s9.75 4.365 9.75 9.75-4.365 9.75-9.75 9.75S2.25 17.385 2.25 12Zm13.36-1.814a.75.75 0 1 0-1.22-.872l-3.236 4.53L9.53 12.22a.75.75 0 0 0-1.06 1.06l2.25 2.25a.75.75 0 0 0 1.14-.094l3.75-5.25Z"
                                            clip-rule="evenodd" />
                                    </svg>
                                </button>
                                <button v-if="!column.enabled" @click="toggleColumn(column.id, true)">
                                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor"
                                        class="w-6 h-6 text-gray-500">
                                        <path fill-rule="evenodd"
                                            d="M12 2.25c-5.385 0-9.75 4.365-9.75 9.75s4.365 9.75 9.75 9.75 9.75-4.365 9.75-9.75S17.385 2.25 12 2.25Zm-1.72 6.97a.75.75 0 1 0-1.06 1.06L10.94 12l-1.72 1.72a.75.75 0 1 0 1.06 1.06L12 13.06l1.72 1.72a.75.75 0 1 0 1.06-1.06L13.06 12l1.72-1.72a.75.75 0 1 0-1.06-1.06L12 10.94l-1.72-1.72Z"
                                            clip-rule="evenodd" />
                                    </svg>
                                </button>
                                <input type="text" v-model.trim="column.text" :placeholder="column.defaultText"
                                    class="w-full rounded-md focus:outline-none focus:border focus:border-gray-200 focus:ring-gray-200 dark:text-slate-200 dark:bg-gray-900 dark:focus:border-gray-800 dark:focus:ring-gray-800" />
                            </li>
                        </ul>
                        <p v-show="!isColumnSelectionValid" class="text-sm text-red-600 mt-2">Please select column(s)
                        </p>
                    </div>
                    <div>
                        <button type="submit"
                            class="flex justify-center px-4 py-2 text-sm w-full shadow-md bg-sky-100 hover:bg-sky-400 border-sky-300 text-sky-600 hover:text-white disabled:bg-gray-300 disabled:text-gray-500 disabled:border-gray-400 disabled:cursor-not-allowed dark:disabled:bg-gray-300 dark:disabled:text-gray-500 dark:disabled:border-gray-400 dark:bg-sky-800 dark:hover:bg-sky-600 dark:border-sky-700 dark:text-sky-100 hover:border-transparent font-medium rounded-md border focus:outline-none focus:ring-2 focus:ring-sky-600 focus:ring-offset-2 dark:focus:ring-2 dark:focus:ring-offset-0 select-none"
                            :disabled="!boardnameEntered || !isColumnSelectionValid" @click="create">
                            Create
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
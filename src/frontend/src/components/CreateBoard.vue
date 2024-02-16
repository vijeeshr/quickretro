<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';
import { createBoard } from '../api'

const router = useRouter()
const boardname = ref('')
const team = ref('')

const boardnameEntered = computed(() => {
    if (boardname.value && boardname.value.trim() !== '') return true
    return false
})

const create = async () => {
    // Todo: Throttle this.
    try {
        const createdBoard = await createBoard(boardname.value, team.value, localStorage.getItem('user') || '')
        router.push(`/board/${createdBoard.id}`)
    } catch (error) {
        console.error('Error creating board:', error);
    }
}

</script>

<template>
    <div class="bg-gray-100 flex h-screen items-center justify-center p-4">
        <div class="w-full max-w-md">
            <div class="bg-white shadow-md rounded-md p-8">
                <h2 class="mt-6 text-center text-3xl font-bold tracking-tight text-gray-600">
                    Create Board
                </h2>
                <div class="space-y-6 mt-4">
                    <div>
                        <div class="mt-1">
                            <input v-model.trim="boardname" name="name" type="text" placeholder="Type board name here!"
                                required autofocus
                                class="px-2 py-2 mt-1 block w-full rounded-md border border-gray-300 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-sky-500 sm:text-sm" />
                        </div>
                        <p v-show="!boardnameEntered" name="name-required" class="text-sm text-red-600 mt-2">Please enter
                            board name</p>
                    </div>
                    <div>
                        <div class="mt-1">
                            <input v-model.trim="team" name="team" type="text" placeholder="Type team name here!"
                                class="px-2 py-2 mt-1 block w-full rounded-md border border-gray-300 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-sky-500 sm:text-sm" />
                        </div>
                    </div>
                    <div>
                        <button type="submit"
                            class="flex justify-center px-4 py-2 text-sm w-full shadow-md bg-sky-100 text-sky-600 font-medium rounded-md border border-sky-300 hover:text-white hover:bg-sky-400 hover:border-transparent focus:outline-none focus:ring-2 focus:ring-sky-600 focus:ring-offset-2 select-none"
                            :disabled="!boardnameEntered" @click="create">
                            Create
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
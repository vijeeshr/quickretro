<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import DarkModeToggle from './DarkModeToggle.vue';
import Avatar from './Avatar.vue';

const route = useRoute()
const router = useRouter()
const board = Array.isArray(route.params.board)
    ? route.params.board[0]
    : route.params.board
const guestname = ref(localStorage.getItem('nickname') || '')

const isGuestNameValid = computed(() => {
    if (guestname.value && guestname.value.trim() !== '') return true
    return false
})

const join = () => {
    if (isGuestNameValid.value) {
        localStorage.setItem("nickname", guestname.value)
        if (board && board.trim() != '') {
            router.push(`/board/${board}`)
        } else {
            router.push(`/create`) // Todo: How about router.replace() ?
        }
    }
}

onMounted(() => {
    if (!localStorage.getItem("user")) {
        localStorage.setItem("user", crypto.randomUUID())
    }
    if (!localStorage.getItem("xid")) {
        localStorage.setItem("xid", crypto.randomUUID())
    }

    const isDark = ref(localStorage.getItem("theme") === "dark")
    document.documentElement.classList.toggle("dark", isDark.value)
})

</script>

<template>
    <div class="bg-gray-100 dark:bg-gray-950 flex h-screen items-center justify-center p-4">
        <div class="w-full max-w-md">
            <div class="bg-white dark:bg-gray-900 shadow-md rounded-md p-8">
                <div class="flex justify-center items-center h-11 w-full">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
                        stroke="currentColor" class="w-11 h-11 text-gray-600 dark:text-gray-400"
                        v-if="!isGuestNameValid">
                        <path stroke-linecap="round" stroke-linejoin="round"
                            d="M17.982 18.725A7.488 7.488 0 0 0 12 15.75a7.488 7.488 0 0 0-5.982 2.975m11.963 0a9 9 0 1 0-11.963 0m11.963 0A8.966 8.966 0 0 1 12 21a8.966 8.966 0 0 1-5.982-2.275M15 9.75a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" />
                    </svg>
                    <Avatar :name="guestname" class="w-10 h-10" v-if="isGuestNameValid" />
                </div>
                <h2
                    class="mt-3 text-center text-3xl font-bold tracking-tight text-gray-600 dark:text-gray-400 select-none">
                    Join as guest
                </h2>
                <div class="space-y-6 mt-4">
                    <div>
                        <div class="mt-1">
                            <input v-model.trim="guestname" name="guestname" type="text"
                                placeholder="Type your name here!" required
                                class="px-2 py-2 mt-1 block w-full rounded-md border border-gray-300 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-sky-500 sm:text-sm dark:bg-slate-800 dark:text-slate-200" />
                        </div>
                        <div class="h-5 flex items-center">
                            <p v-show="!isGuestNameValid" name="guestname-required"
                                class="text-sm text-red-600 dark:text-red-300 mt-2 select-none">Please
                                enter your name</p>
                        </div>
                    </div>
                    <div class="flex w-full gap-2">
                        <button type="submit"
                            class="flex justify-center px-4 py-2 text-sm w-[90%] shadow-md bg-sky-100 hover:bg-sky-400 border-sky-300 text-sky-600 hover:text-white disabled:bg-gray-300 disabled:text-gray-500 disabled:border-gray-400 disabled:cursor-not-allowed dark:disabled:bg-gray-300 dark:disabled:text-gray-500 dark:disabled:border-gray-400 dark:bg-sky-800 dark:hover:bg-sky-600 dark:border-sky-700 dark:text-sky-100 hover:border-transparent font-medium rounded-md border focus:outline-none focus:ring-2 focus:ring-sky-600 focus:ring-offset-2 dark:focus:ring-2 dark:focus:ring-offset-0 select-none"
                            :disabled="!isGuestNameValid" @click="join">
                            Join
                        </button>
                        <div
                            class="w-[10%] flex items-center justify-center shadow-md border rounded-md border-sky-200">
                            <DarkModeToggle class="w-6 h-6 cursor-pointer text-sky-200 hover:text-sky-400" />
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
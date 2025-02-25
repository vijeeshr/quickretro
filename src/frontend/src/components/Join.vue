<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import DarkModeToggle from './DarkModeToggle.vue';

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
                <h2 class="mt-6 text-center text-3xl font-bold tracking-tight text-gray-600 dark:text-gray-400">
                    Join as guest
                </h2>
                <div class="space-y-6 mt-4">
                    <div>
                        <div class="mt-1">
                            <input v-model.trim="guestname" name="guestname" type="text"
                                placeholder="Type your name here!" required
                                class="px-2 py-2 mt-1 block w-full rounded-md border border-gray-300 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-sky-500 sm:text-sm dark:bg-slate-800 dark:text-slate-200" />
                        </div>
                        <p v-show="!isGuestNameValid" name="guestname-required"
                            class="text-sm text-red-600 dark:text-red-300 mt-2">Please
                            enter your name</p>
                    </div>
                    <div class="flex w-full gap-2">
                        <button type="submit"
                            class="flex justify-center px-4 py-2 text-sm w-[90%] shadow-md bg-sky-100 hover:bg-sky-400 border-sky-300 text-sky-600 hover:text-white disabled:bg-gray-300 disabled:text-gray-500 disabled:border-gray-400 disabled:cursor-not-allowed dark:disabled:bg-gray-300 dark:disabled:text-gray-500 dark:disabled:border-gray-400 dark:bg-sky-800 dark:hover:bg-sky-600 dark:border-sky-700 dark:text-sky-100 hover:border-transparent font-medium rounded-md border focus:outline-none focus:ring-2 focus:ring-sky-600 focus:ring-offset-2 dark:focus:ring-2 dark:focus:ring-offset-0 select-none"
                            :disabled="!isGuestNameValid" @click="join">
                            Join
                        </button>
                        <div class="w-[10%] flex items-center justify-center shadow-md border rounded-md border-sky-200">
                            <DarkModeToggle class="w-6 h-6 cursor-pointer text-sky-200 hover:text-sky-400" />
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
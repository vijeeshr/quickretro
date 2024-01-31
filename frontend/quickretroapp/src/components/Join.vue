<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

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
})

</script>

<template>
<div class="bg-gray-100 flex h-screen items-center justify-center p-4">
    <div class="w-full max-w-md">
        <div class="bg-white shadow-md rounded-md p-8">
            <h2 class="mt-6 text-center text-3xl font-bold tracking-tight text-gray-600">
                Join as guest
            </h2>
            <div class="space-y-6 mt-4">
                <div>
                    <div class="mt-1">
                        <input v-model.trim="guestname" name="guestname" type="text" placeholder="Type your name here!" required
                            class="px-2 py-2 mt-1 block w-full rounded-md border border-gray-300 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-sky-500 sm:text-sm" />
                    </div>
                    <p v-show="!isGuestNameValid" name="guestname-required" class="text-sm text-red-600 mt-2">Please enter your name</p>
                </div>
                <div>
                    <button type="submit"
                        class="flex justify-center px-4 py-2 text-sm w-full shadow-md bg-sky-100 text-sky-600 font-medium rounded-md border border-sky-300 hover:text-white hover:bg-sky-400 hover:border-transparent focus:outline-none focus:ring-2 focus:ring-sky-600 focus:ring-offset-2 select-none"
                        :disabled="!isGuestNameValid"
                        @click="join">
                        Join
                    </button>
                </div>
            </div>
        </div>
    </div>
</div>    
</template>
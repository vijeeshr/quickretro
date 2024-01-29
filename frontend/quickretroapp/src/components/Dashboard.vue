<script setup lang="ts">
import { ref } from 'vue';
import Avatar from './Avatar.vue';
import Card from './Card.vue';
import Category from './Category.vue';
import { CardModel } from '../models/CardModel';
import NewCard from './NewCard.vue';
// import { useRoute } from 'vue-router';
// import { useWebSocket } from '@vueuse/core'

const mask = ref(true)
const newCardCategory = ref('')

// const route = useRoute()
// // const router = useRouter()
// const board = Array.isArray(route.params.board)
//                   ? route.params.board[0]
//                   : route.params.board

// console.log('Attempt WS call..')
// const user = localStorage.getItem("user")
// const { status, data, send, open, close } = useWebSocket(`ws://${document.location.host}/board/${board}/user/${user}/meet`)
// console.log(status, data, send, open, close)               

const cards = ref<CardModel[]>([
    { typ: "msg", id: "f19ffc16-fee7-4d52-9cd0-f4d46d2a82ba", nickname: "Vijeesh Ravindran", msg: "was", cat: "good", likes: "1", liked: true, mine: true },
    { typ: "msg", id: "f19ffc16-fee7-4d52-9cd0-f4d46d2a82bb", nickname: "John Doe", msg: "From John", cat: "bad", likes: "0", liked: false, mine: false },
    { typ: "msg", id: "f19ffc16-fee7-4d52-9cd0-f4d46d2a82bc", nickname: "John Doe", msg: "This is some random text", cat: "good", likes: "0", liked: false, mine: false }
])

const filterCards = (category: string) => {
    return cards.value.filter(c => c.cat.toLowerCase() === category.toLowerCase());
}

const add = (category: string) => {
    newCardCategory.value = category
}

const onAdded = (card: CardModel) => {
    console.log('newcontent received:', card)
    //unmount newCard
    newCardCategory.value = ''
    cards.value.push(card)
}

</script>

<template>
<div class="flex h-full min-h-screen bg-gray-800 text-white">

    <!-- Left Sidebar -->
    <div class="w-16 p-4">
        <Avatar class="ml-auto mx-auto mb-4" />
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-8 h-8 mx-auto mb-4 cursor-pointer" 
            :class="{ 'hidden': mask }"
            @click="mask = !mask">
            <path stroke-linecap="round" stroke-linejoin="round" d="M3.98 8.223A10.477 10.477 0 0 0 1.934 12C3.226 16.338 7.244 19.5 12 19.5c.993 0 1.953-.138 2.863-.395M6.228 6.228A10.451 10.451 0 0 1 12 4.5c4.756 0 8.773 3.162 10.065 7.498a10.522 10.522 0 0 1-4.293 5.774M6.228 6.228 3 3m3.228 3.228 3.65 3.65m7.894 7.894L21 21m-3.228-3.228-3.65-3.65m0 0a3 3 0 1 0-4.243-4.243m4.242 4.242L9.88 9.88" />
        </svg>
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-8 h-8 mx-auto mb-4 cursor-pointer" 
            :class="{ 'hidden': !mask }"
            @click="mask = !mask">
            <path stroke-linecap="round" stroke-linejoin="round" d="M2.036 12.322a1.012 1.012 0 0 1 0-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178Z" />
            <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" />
        </svg>      
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-8 h-8 mx-auto mb-4 cursor-pointer">
            <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
        </svg>    
        <a href="https://github.com/vijeeshr/quickretro" target="_blank">
            <svg viewBox="0 0 24 24" aria-hidden="true" class="h-8 w-8 fill-slate-100">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M12 2C6.477 2 2 6.463 2 11.97c0 4.404 2.865 8.14 6.839 9.458.5.092.682-.216.682-.48 0-.236-.008-.864-.013-1.695-2.782.602-3.369-1.337-3.369-1.337-.454-1.151-1.11-1.458-1.11-1.458-.908-.618.069-.606.069-.606 1.003.07 1.531 1.027 1.531 1.027.892 1.524 2.341 1.084 2.91.828.092-.643.35-1.083.636-1.332-2.22-.251-4.555-1.107-4.555-4.927 0-1.088.39-1.979 1.029-2.675-.103-.252-.446-1.266.098-2.638 0 0 .84-.268 2.75 1.022A9.607 9.607 0 0 1 12 6.82c.85.004 1.705.114 2.504.336 1.909-1.29 2.747-1.022 2.747-1.022.546 1.372.202 2.386.1 2.638.64.696 1.028 1.587 1.028 2.675 0 3.83-2.339 4.673-4.566 4.92.359.307.678.915.678 1.846 0 1.332-.012 2.407-.012 2.734 0 .267.18.577.688.48 3.97-1.32 6.833-5.054 6.833-9.458C22 6.463 17.522 2 12 2Z"></path>
            </svg>
        </a>
    </div>
    <!-- Left Sidebar -->

    <!-- Dashboard Content -->
    <div class="flex-1 flex bg-gray-100 overflow-hidden">
        <Category button-text="Add Sails" color="green" @add-card="add('good')">
            <NewCard v-if="newCardCategory=='good'" category="good" @added="onAdded" />
            <Card v-for="card in filterCards('good')" :card="card" :mask="mask" :key="card.id" />
        </Category>
        <Category button-text="Add Anchors" color="red" @add-card="add('bad')">
            <NewCard v-if="newCardCategory=='bad'" category="bad" @added="onAdded" />
            <Card v-for="card in filterCards('bad')" :card="card" :mask="mask" :key="card.id" />
        </Category>
        <Category button-text="Add Next Steps" color="yellow" @add-card="add('next')">
            <NewCard v-if="newCardCategory=='next'" category="next" @added="onAdded" />
            <Card v-for="card in filterCards('next')" :card="card" :mask="mask" :key="card.id" />
        </Category>
    </div>
    <!-- Dashboard Content -->

    <!-- Right Sidebar -->
    <div class="w-16 p-4">
        <Avatar name="Vijeesh Ravindran" class="ml-auto mx-auto mb-4" />
        <Avatar name="Tony Stark" class="ml-auto mx-auto mb-4" />
        <Avatar name="John Doe" class="ml-auto mx-auto mb-4" />
    </div>
    <!-- Right Sidebar -->        

</div>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue';
import Avatar from './Avatar.vue';
import Card from './Card.vue';
import Category from './Category.vue';
import NewCard from './NewCard.vue';
import { useRoute } from 'vue-router';
import { EventRequest, MaskEvent, MaskResponse, RegisterEvent, RegisterResponse, MessageResponse, UserClosingResponse, toSocketResponse, SaveMessageEvent, DeleteMessageEvent, DeleteMessageResponse, LikeMessageEvent, LikeMessageResponse } from '../models/Requests';
import { OnlineUser } from '../models/OnlineUser';
import { DraftMessage } from '../models/DraftMessage';
import { LikeMessage } from '../models/LikeMessage';

const isMasked = ref(true)
const isOwner = ref(false)
const newCardCategory = ref('')
const route = useRoute()
const board = Array.isArray(route.params.board)
    ? route.params.board[0]
    : route.params.board
// Todo: Find a way of passing this from route..meta?
const user = localStorage.getItem("user") || ''
const externalId = localStorage.getItem("xid") || ''
const nickname = localStorage.getItem("nickname") || ''
const isConnected = ref(false)
let socket: WebSocket

const cards = ref<MessageResponse[]>([]) // Todo: Rework models

const onlineUsers = ref<OnlineUser[]>([])

const filterCards = (category: string) => {
    return cards.value.filter(c => c.cat.toLowerCase() === category.toLowerCase());
}

const add = (category: string) => {
    newCardCategory.value = category
}

const onAdded = (card: DraftMessage) => {
    console.log('newcontent received:', card)
    newCardCategory.value = '' //unmount newCard
    dispatchEvent<SaveMessageEvent>("msg", { id: card.id, by: user, nickname: nickname, grp: board, msg: card.msg, cat: card.cat })
}

const onUpdated = (card: DraftMessage) => {
    console.log('Updated content received:', card)
    dispatchEvent<SaveMessageEvent>("msg", { id: card.id, by: user, nickname: nickname, grp: board, msg: card.msg, cat: card.cat })
}

const onDeleted = (cardId: string) => {
    dispatchEvent<DeleteMessageEvent>("del", { msgId: cardId, by: user, grp: board })
}

const onLiked = (likeMessage: LikeMessage) => {
    dispatchEvent<LikeMessageEvent>("like", { msgId: likeMessage.msgId, by: user, like: likeMessage.like })
}

const mask = () => {
    dispatchEvent<MaskEvent>("mask", { by: user, grp: board, mask: !isMasked.value })
}

const onRegisterResponse = (response: RegisterResponse) => {
    isOwner.value = response.isBoardOwner
    isMasked.value = response.boardMasking
    onlineUsers.value = []
    onlineUsers.value.push(...response.users) // Todo: find a better way
    // Load messages.
    // Only loading messages when the RegisterResponse is for the current User's RegisterEvent request. 
    // This prevents unnecessarily pushing messages in the ref for other users RegisterEvents. RegisterEvent happens just once in the beginning.
    if (response.mine) {
        cards.value = []
        cards.value.push(...response.messages)
    }
}

const onUserClosingResponse = (response: UserClosingResponse) => {
    onlineUsers.value = []
    onlineUsers.value.push(...response.users) // Todo: find a better way
}

const onMaskResponse = (response: MaskResponse) => {
    isMasked.value = response.mask
}

const onSaveMessageResponse = (response: MessageResponse) => {
    let index = cards.value.findIndex(x => x.id === response.id)
    if (index === -1) {
        cards.value.push(response)
    } else {
        cards.value[index] = response
    }
}

const onDeleteMessageResponse = (response: DeleteMessageResponse) => {
    let index = cards.value.findIndex(x => x.id === response.id)
    if (index !== -1) {
        cards.value.splice(index, 1)
    }
}

const onLikeMessageResponse = (response: LikeMessageResponse) => {
    let index = cards.value.findIndex(x => x.id === response.id)
    if (index !== -1) {
        // Todo: Should whole card be updated? cards.value[index] = response. Or just the properties.
        cards.value[index].liked = response.liked
        cards.value[index].likes = response.likes
    }
}

const dispatchEvent = <T>(eventType: string, payload: T) => {
    const event: EventRequest<T> = {
        typ: eventType,
        pyl: payload
    }
    console.log("Dispatching", event)
    if (socket.readyState == 1) {
        socket.send(JSON.stringify(event)); // Can throw error if socket object is "connecting". Check the docs.
    } else {
        console.log('Socket not ready for send operation')
    }
}

const socketOnOpen = (event: Event) => {
    console.log("[open] Connection established", event)
    isConnected.value = true
    dispatchEvent<RegisterEvent>("reg", { by: user, nickname: nickname, xid: externalId, grp: board })
}
const socketOnClose = (event: CloseEvent) => {
    isConnected.value = false
    console.log("Close received", event)
}
const socketOnError = (event: Event) => {
    console.error(event)
}
const socketOnMessage = (event: MessageEvent<any>) => {
    const response = toSocketResponse(JSON.parse(event.data))
    console.log('Response', response)

    if (response && response.typ) {
        switch (response.typ) {
            case "reg":
                onRegisterResponse(response)
                break
            case "closing":
                onUserClosingResponse(response)
                break
            case "mask":
                onMaskResponse(response)
                break
            case "msg":
                onSaveMessageResponse(response)
                break
            case "del":
                onDeleteMessageResponse(response)
                break
            case "like":
                onLikeMessageResponse(response)
                break
        }
    }
}

onMounted(() => {
    socket = new WebSocket(`ws://${document.location.host}/ws/board/${board}/user/${user}/meet`)
    socket.onopen = socketOnOpen
    socket.onclose = socketOnClose
    socket.onerror = socketOnError
    socket.onmessage = socketOnMessage
})

// onUnmounted(() => {
//     if (socket && (!socket.CLOSED || !socket.CLOSING)) {
//         socket.close(1001)
//     }
// })

</script>

<template>
    <div class="flex h-full min-h-screen bg-gray-800 text-white">

        <!-- Left Sidebar -->
        <div class="w-16 p-4">
            <Avatar :name="nickname" class="ml-auto mx-auto mb-4" />
            <!-- Mask controls -->
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"
                class="w-8 h-8 mx-auto mb-4 cursor-pointer" v-if="isOwner" :class="{ 'hidden': isMasked }" @click="mask">
                <path stroke-linecap="round" stroke-linejoin="round"
                    d="M3.98 8.223A10.477 10.477 0 0 0 1.934 12C3.226 16.338 7.244 19.5 12 19.5c.993 0 1.953-.138 2.863-.395M6.228 6.228A10.451 10.451 0 0 1 12 4.5c4.756 0 8.773 3.162 10.065 7.498a10.522 10.522 0 0 1-4.293 5.774M6.228 6.228 3 3m3.228 3.228 3.65 3.65m7.894 7.894L21 21m-3.228-3.228-3.65-3.65m0 0a3 3 0 1 0-4.243-4.243m4.242 4.242L9.88 9.88" />
            </svg>
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"
                class="w-8 h-8 mx-auto mb-4 cursor-pointer" v-if="isOwner" :class="{ 'hidden': !isMasked }" @click="mask">
                <path stroke-linecap="round" stroke-linejoin="round"
                    d="M2.036 12.322a1.012 1.012 0 0 1 0-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178Z" />
                <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" />
            </svg>
            <a href="https://github.com/vijeeshr/quickretro" target="_blank">
                <svg viewBox="0 0 24 24" aria-hidden="true" class="h-8 w-8 fill-slate-100">
                    <path fill-rule="evenodd" clip-rule="evenodd"
                        d="M12 2C6.477 2 2 6.463 2 11.97c0 4.404 2.865 8.14 6.839 9.458.5.092.682-.216.682-.48 0-.236-.008-.864-.013-1.695-2.782.602-3.369-1.337-3.369-1.337-.454-1.151-1.11-1.458-1.11-1.458-.908-.618.069-.606.069-.606 1.003.07 1.531 1.027 1.531 1.027.892 1.524 2.341 1.084 2.91.828.092-.643.35-1.083.636-1.332-2.22-.251-4.555-1.107-4.555-4.927 0-1.088.39-1.979 1.029-2.675-.103-.252-.446-1.266.098-2.638 0 0 .84-.268 2.75 1.022A9.607 9.607 0 0 1 12 6.82c.85.004 1.705.114 2.504.336 1.909-1.29 2.747-1.022 2.747-1.022.546 1.372.202 2.386.1 2.638.64.696 1.028 1.587 1.028 2.675 0 3.83-2.339 4.673-4.566 4.92.359.307.678.915.678 1.846 0 1.332-.012 2.407-.012 2.734 0 .267.18.577.688.48 3.97-1.32 6.833-5.054 6.833-9.458C22 6.463 17.522 2 12 2Z">
                    </path>
                </svg>
            </a>
        </div>
        <!-- Left Sidebar -->

        <!-- Dashboard Content -->
        <div class="flex-1 flex bg-gray-100 overflow-hidden">
            <Category button-text="Add Sails" color="green" @add-card="add('good')">
                <NewCard v-if="newCardCategory == 'good'" category="good" @added="onAdded" />
                <Card v-for="card in filterCards('good')" :card="card" :mask="isMasked" :key="card.id" @updated="onUpdated"
                    @deleted="onDeleted" @liked="onLiked" />
            </Category>
            <Category button-text="Add Anchors" color="red" @add-card="add('bad')">
                <NewCard v-if="newCardCategory == 'bad'" category="bad" @added="onAdded" />
                <Card v-for="card in filterCards('bad')" :card="card" :mask="isMasked" :key="card.id" @updated="onUpdated"
                    @deleted="onDeleted" @liked="onLiked" />
            </Category>
            <Category button-text="Add Next Steps" color="yellow" @add-card="add('next')">
                <NewCard v-if="newCardCategory == 'next'" category="next" @added="onAdded" />
                <Card v-for="card in filterCards('next')" :card="card" :mask="isMasked" :key="card.id" @updated="onUpdated"
                    @deleted="onDeleted" @liked="onLiked" />
            </Category>
        </div>
        <!-- Dashboard Content -->

        <!-- Right Sidebar -->
        <div class="w-16 p-4">
            <Avatar v-for="user in onlineUsers" :name="user.nickname" class="ml-auto mx-auto mb-4" />
        </div>
        <!-- Right Sidebar -->

    </div>
</template>
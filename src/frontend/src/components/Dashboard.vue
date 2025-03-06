<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue';
import Avatar from './Avatar.vue';
import Card from './Card.vue';
import Category from './Category.vue';
import NewAnonymousCard from './NewAnonymousCard.vue';
import NewCard from './NewCard.vue';
import { useRoute } from 'vue-router';
import { EventRequest, MaskEvent, MaskResponse, RegisterEvent, RegisterResponse, MessageResponse, UserClosingResponse, toSocketResponse, SaveMessageEvent, DeleteMessageEvent, DeleteMessageResponse, LikeMessageEvent, LikeMessageResponse, LockEvent, LockResponse, TimerResponse, TimerEvent, CategoryChangeEvent, CategoryChangeResponse } from '../models/Requests';
import { OnlineUser } from '../models/OnlineUser';
import { DraftMessage } from '../models/DraftMessage';
import { LikeMessage } from '../models/LikeMessage';
import { jsPDF } from 'jspdf';
import autoTable from 'jspdf-autotable';
import { Dialog, DialogPanel, DialogTitle } from '@headlessui/vue'
import CountdownTimer from './CountdownTimer.vue';
import TimerPanel from './TimerPanel.vue';
import { logMessage } from '../utils';
import { useToast } from 'vue-toast-notification';
// import 'vue-toast-notification/dist/theme-sugar.css';
// import 'vue-toast-notification/dist/theme-bootstrap.css';
import 'vue-toast-notification/dist/theme-default.css';
import DarkModeToggle from './DarkModeToggle.vue';
import { BoardColumn } from '../models/BoardColumn';
import { CategoryChangeMessage } from '../models/CategoryChangeMessage';

const isMasked = ref(true)
const isOwner = ref(false)
const isLocked = ref(false)
const timerExpiresInSeconds = ref(0)
const newCardCategory = ref('')
const newAnonymousCardCategory = ref('')
const route = useRoute()
const board = Array.isArray(route.params.board)
    ? route.params.board[0]
    : route.params.board
// Todo: Find a way of passing this from route..meta?
const user = localStorage.getItem("user") || ''
const externalId = localStorage.getItem("xid") || ''
const nickname = localStorage.getItem("nickname") || ''
const isConnected = ref(false)
const boardName = ref('')
const shareLink = `${window.location.href}`
let socket: WebSocket

const cards = ref<MessageResponse[]>([]) // Todo: Rework models
const columns = ref<BoardColumn[]>([])
const onlineUsers = ref<OnlineUser[]>([])

const isShareDialogOpen = ref(false)
const setIsShareDialogOpen = (value: boolean) => {
    isShareDialogOpen.value = value
}

const isTimerDialogOpen = ref(false)
const setIsTimerDialogOpen = (value: boolean) => {
    isTimerDialogOpen.value = value
}

const isTimerCountdownInProgress = ref(false)
const onCountdownProgressUpdate = (value: boolean) => {
    isTimerCountdownInProgress.value = value
}

const toast = useToast()

const onOneMinuteLeftWarning = () => {
    toast.info("One minute left for countdown", { pauseOnHover: false })
}

const onCountdownCompleted = () => {
    toast.error("Hey! You've run out of time", { pauseOnHover: false, duration: 4000 })
}

const filterCards = (category: string) => {
    return cards.value.filter(c => c.cat.toLowerCase() === category.toLowerCase())
}

const add = (category: string, anonymous: boolean) => {
    if (isLocked.value) {
        logMessage('Locked! Cannot add.')
        return
    }
    if (!anonymous) {
        // Unmounts NewAnonymousCard and mounts NewCard, and vice-versa
        newAnonymousCardCategory.value = ''
        newCardCategory.value = category
    } else {
        newCardCategory.value = ''
        newAnonymousCardCategory.value = category
    }
}

const columnWidthClass = computed(() => {
    switch (columns.value.length) {
        case 4:
            return "w-1/4"
        case 5:
            return "w-1/5"
        default:
            return "w-1/3"
    }
})

const onAdded = (card: DraftMessage) => {
    logMessage('newcontent received:', card)
    newCardCategory.value = '' //unmount newCard
    newAnonymousCardCategory.value = '' //unmount newAnonymousCard
    const nicknameToSend = card.anon === true ? '' : nickname
    dispatchEvent<SaveMessageEvent>("msg", { id: card.id, by: user, nickname: nicknameToSend, grp: board, msg: card.msg, cat: card.cat, anon: card.anon })
}

const onInvalidContent = (errorMessage: string) => {
    toast.error(errorMessage)
}

const onUpdated = (card: DraftMessage) => {
    logMessage('Updated content received:', card)
    dispatchEvent<SaveMessageEvent>("msg", { id: card.id, by: user, nickname: nickname, grp: board, msg: card.msg, cat: card.cat, anon: false })
}

const onDeleted = (cardId: string) => {
    dispatchEvent<DeleteMessageEvent>("del", { msgId: cardId, by: user, grp: board })
}

const onLiked = (likeMessage: LikeMessage) => {
    dispatchEvent<LikeMessageEvent>("like", { msgId: likeMessage.msgId, by: user, like: likeMessage.like })
}

const onCategoryChanged = (pyl: CategoryChangeMessage) => {
    dispatchEvent<CategoryChangeEvent>("catchng", { msgId: pyl.msgId, by: user, grp: board, newcat: pyl.newCategoryId, oldcat: pyl.oldCategoryId })
}

const mask = () => {
    dispatchEvent<MaskEvent>("mask", { by: user, grp: board, mask: !isMasked.value })
}

const lock = () => {
    dispatchEvent<LockEvent>("lock", { by: user, grp: board, lock: !isLocked.value })
}

const onTimerStart = (expiryDurationInSeconds: number) => {
    dispatchEvent<TimerEvent>("timer", { by: user, grp: board, expiryDurationInSeconds: expiryDurationInSeconds, stop: false }) // "stop" is ignored
    setIsTimerDialogOpen(false)
}

const onTimerStop = () => {
    dispatchEvent<TimerEvent>("timer", { by: user, grp: board, expiryDurationInSeconds: 0, stop: true }) // "expiryDurationInSeconds" is ignored
    setIsTimerDialogOpen(false)
}

const getRGBizedColor = (color: string): any => {
    switch (color) {
        case "green":
            return [74, 222, 128]
        case "red":
            return [248, 113, 113]
        case "yellow":
            return [250, 204, 21]
        case "fuchsia":
            return [232, 121, 249]
        case "orange":
            return [251, 146, 60]
        default:
            return [128, 128, 128]
    }
}

const download = () => {
    const doc = new jsPDF({
        orientation: "portrait",
        unit: "in",
        format: "letter"
    });

    // text is placed using x, y coordinates
    doc.setFontSize(16).text(`Board - ${boardName.value}`, 0.5, 1.0)
    // create a line under heading 
    doc.setLineWidth(0.01).line(0.5, 1.1, 8.0, 1.1);

    for (const col of columns.value) {
        const columnForExport = [
            { title: col.text, dataKey: "text" }
        ]
        const itemsForExport = cards.value.filter(c => c.cat.toLowerCase() === col.id.toLowerCase())
            .map(c => ({
                text: c.msg
            }))

        // Using autoTable plugin
        autoTable(doc, {
            columns: columnForExport,
            headStyles: { fillColor: getRGBizedColor(col.color) },
            body: itemsForExport,
            margin: { left: 0.5, top: 1.25 }
        })
    }

    // Footer
    doc
        .setFont("times")
        .setFontSize(11)
        .setTextColor("gray")
        .text(
            "Created with QuickRetro https://quickretro.app",
            0.5,
            doc.internal.pageSize.height - 0.5
        )

    doc.save(`quickretro.pdf`)

    // Using array of sentences
    // doc
    //     .setFont("helvetica")
    //     .setFontSize(12)
    //     .text(moreText, 0.5, 3.5, { align: "left", maxWidth: 7.5 });
}

const share = () => {
    isShareDialogOpen.value = true
}

const copyShareLink = async () => {
    try {
        await navigator.clipboard.writeText(shareLink)
        toast.success("Link copied!")
        setIsShareDialogOpen(false)
    } catch (err) {
        toast.error("Failed to copy. Please copy directly.")
    }
}

const timerSettings = () => {
    if (isOwner.value) {
        isTimerDialogOpen.value = true
    }
}

const onRegisterResponse = (response: RegisterResponse) => {
    isOwner.value = response.isBoardOwner
    isMasked.value = response.boardMasking
    isLocked.value = response.boardLock
    timerExpiresInSeconds.value = response.timerExpiresInSeconds
    onlineUsers.value = []
    onlineUsers.value.push(...response.users) // Todo: find a better way
    columns.value = []
    columns.value.push(...response.columns) // Todo: find a better way
    boardName.value = response.boardName
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

const onLockResponse = (response: LockResponse) => {
    isLocked.value = response.lock
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

const onCategoryChangeResponse = (response: CategoryChangeResponse) => {
    let index = cards.value.findIndex(x => x.id === response.id)
    if (index !== -1) {
        cards.value[index].cat = response.newcat
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

const onTimerResponse = (response: TimerResponse) => {
    // Hack: If both previous (timerExpiresInSeconds.value) and new (response.expiresInSeconds) value is the same..
    // .. the watcher downstream in CountdownTimer doesn't trigger and this causes a weird UI issue when board owner ..
    // .. again tries to start the timer with an interval duration that is exactly same as was used in the immediate previous run.
    if (timerExpiresInSeconds.value === response.expiresInSeconds) {
        // Reset to dummy value..0 is safe here.
        timerExpiresInSeconds.value = 0
        nextTick(() => {
            timerExpiresInSeconds.value = response.expiresInSeconds
        })
    } else {
        timerExpiresInSeconds.value = response.expiresInSeconds
    }
}

const dispatchEvent = <T>(eventType: string, payload: T) => {
    const event: EventRequest<T> = {
        typ: eventType,
        pyl: payload
    }
    logMessage("Dispatching", event)
    if (socket.readyState == 1) {
        socket.send(JSON.stringify(event)); // Can throw error if socket object is "connecting". Check the docs.
    } else {
        logMessage('Socket not ready for send operation')
    }
}

const socketOnOpen = (event: Event) => {
    logMessage("[open] Connection established", event)
    isConnected.value = true
    dispatchEvent<RegisterEvent>("reg", { by: user, nickname: nickname, xid: externalId, grp: board })
}
const socketOnClose = (event: CloseEvent) => {
    isConnected.value = false
    logMessage("Close received", event)
}
const socketOnError = (event: Event) => {
    console.error(event)
}
const socketOnMessage = (event: MessageEvent<any>) => {
    const response = toSocketResponse(JSON.parse(event.data))
    logMessage('Response', response)

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
            case "lock":
                onLockResponse(response)
                break
            case "msg":
                onSaveMessageResponse(response)
                break
            case "del":
                onDeleteMessageResponse(response)
                break
            case "catchng":
                onCategoryChangeResponse(response)
                break
            case "like":
                onLikeMessageResponse(response)
                break
            case "timer":
                onTimerResponse(response)
                break
        }
    }
}

const handleVisibilityChange = () => {
    // Attempt reinitializing the app (with browser reload) when websocket is closed because of inactivity
    if (document.visibilityState === "visible" && socket.readyState !== WebSocket.OPEN) {
        window.location.reload()
    }
}

const handleConnectivity = () => {
    if (!navigator.onLine) {
        toast.error("You seem to be offline.", { pauseOnHover: false, duration: 4000 })
    }
    // if (navigator.onLine) {
    //     toast.success("You are online", { pauseOnHover: false, duration: 4000 })
    //     // Attempt reinitializing the app (with browser reload) when websocket is closed because of connectivity issues
    //     if (socket.readyState !== WebSocket.OPEN) {
    //         window.location.reload()
    //     }
    // }
}

onMounted(() => {
    const websocketProtocol = import.meta.env.VITE_WS_PROTOCOL || 'wss'
    socket = new WebSocket(`${websocketProtocol}://${document.location.host}/ws/board/${board}/user/${user}/meet`)
    socket.onopen = socketOnOpen
    socket.onclose = socketOnClose
    socket.onerror = socketOnError
    socket.onmessage = socketOnMessage

    document.addEventListener('visibilitychange', handleVisibilityChange)
    window.addEventListener("offline", handleConnectivity)
    // window.addEventListener("online", handleConnectivity)
})

onUnmounted(() => {
    document.removeEventListener('visibilitychange', handleVisibilityChange)
    window.removeEventListener("offline", handleConnectivity)
    // window.removeEventListener("online", handleConnectivity)
    // if (socket && (!socket.CLOSED || !socket.CLOSING)) {
    //     socket.close(1001)
    // }
})

</script>

<template>
    <div class="flex h-full min-h-screen bg-gray-800 dark:bg-gray-950 text-white">

        <!-- Dialog for Timer settings -->
        <Dialog :open="isTimerDialogOpen" @close="setIsTimerDialogOpen" class="relative z-50">
            <!-- The backdrop, rendered as a fixed sibling to the panel container -->
            <div class="fixed inset-0 bg-black/30" aria-hidden="true" />
            <div class="fixed inset-0 flex w-screen items-center justify-center p-4">
                <DialogPanel class="rounded-2xl bg-white dark:bg-gray-800 p-6 text-left align-middle shadow-xl">
                    <DialogTitle as="h3"
                        class="text-lg font-medium leading-6 text-gray-900 dark:text-gray-200 select-none">
                        Start/Stop Timer
                    </DialogTitle>
                    <TimerPanel :is-countdown-in-progress="isTimerCountdownInProgress" @start="onTimerStart"
                        @stop="onTimerStop"></TimerPanel>
                </DialogPanel>
            </div>
        </Dialog>

        <!-- Dialog to share url -->
        <Dialog :open="isShareDialogOpen" @close="setIsShareDialogOpen" class="relative z-50">
            <!-- The backdrop, rendered as a fixed sibling to the panel container -->
            <div class="fixed inset-0 bg-black/30" aria-hidden="true" />
            <div class="fixed inset-0 flex w-screen items-center justify-center p-4">
                <DialogPanel class="rounded-2xl bg-white dark:bg-gray-800 p-6 text-left align-middle shadow-xl">
                    <DialogTitle as="h3"
                        class="text-lg font-medium leading-6 text-gray-900 dark:text-gray-200 select-none">Copy and
                        share below url to
                        participants
                    </DialogTitle>
                    <article class="select-all bg-slate-100 dark:bg-slate-700 dark:text-gray-100 my-6 rounded-sm">{{
                        shareLink }}</article>
                    <div class="mt-4">
                        <button type="button"
                            class="px-4 py-2 text-sm w-full shadow-md font-medium rounded-md border bg-sky-100 hover:bg-sky-400 border-sky-300 text-sky-600 hover:text-white dark:bg-sky-800 dark:hover:bg-sky-600 dark:border-sky-700 dark:text-sky-100 hover:border-transparent select-none focus:outline-none focus:ring-0"
                            @click="copyShareLink">
                            Copy
                        </button>
                    </div>
                </DialogPanel>
            </div>
        </Dialog>

        <!-- Left Sidebar -->
        <div class="w-16 p-3">
            <!-- Timer -->
            <CountdownTimer :timeLeftInSeconds="timerExpiresInSeconds" title="Countdown Timer"
                class="inline-flex items-center justify-center overflow-hidden rounded-full w-10 h-10 text-[0.825rem] leading-[1rem] font-bold text-white ml-auto mx-auto mb-4"
                :class="isOwner ? 'cursor-pointer' : 'cursor-default'" @click="timerSettings"
                @countdown-progress-update="onCountdownProgressUpdate" @one-minute-left-warning="onOneMinuteLeftWarning"
                @countdown-completed="onCountdownCompleted" />
            <!-- Share -->
            <div title="Share board with others">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
                    stroke="currentColor" class="w-8 h-8 mx-auto mb-4 cursor-pointer" @click="share">
                    <path stroke-linecap="round" stroke-linejoin="round"
                        d="M7.217 10.907a2.25 2.25 0 1 0 0 2.186m0-2.186c.18.324.283.696.283 1.093s-.103.77-.283 1.093m0-2.186 9.566-5.314m-9.566 7.5 9.566 5.314m0 0a2.25 2.25 0 1 0 3.935 2.186 2.25 2.25 0 0 0-3.935-2.186Zm0-12.814a2.25 2.25 0 1 0 3.933-2.185 2.25 2.25 0 0 0-3.933 2.185Z" />
                </svg>
            </div>
            <!-- Mask controls -->
            <div :title="!isMasked ? 'Mask messages' : 'Unmask messages'">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
                    stroke="currentColor" class="w-8 h-8 mx-auto mb-4 cursor-pointer" v-if="isOwner"
                    :class="{ 'hidden': isMasked }" @click="mask">
                    <path stroke-linecap="round" stroke-linejoin="round"
                        d="M3.98 8.223A10.477 10.477 0 0 0 1.934 12C3.226 16.338 7.244 19.5 12 19.5c.993 0 1.953-.138 2.863-.395M6.228 6.228A10.451 10.451 0 0 1 12 4.5c4.756 0 8.773 3.162 10.065 7.498a10.522 10.522 0 0 1-4.293 5.774M6.228 6.228 3 3m3.228 3.228 3.65 3.65m7.894 7.894L21 21m-3.228-3.228-3.65-3.65m0 0a3 3 0 1 0-4.243-4.243m4.242 4.242L9.88 9.88" />
                </svg>
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
                    stroke="currentColor" class="w-8 h-8 mx-auto mb-4 cursor-pointer" v-if="isOwner"
                    :class="{ 'hidden': !isMasked }" @click="mask">
                    <path stroke-linecap="round" stroke-linejoin="round"
                        d="M2.036 12.322a1.012 1.012 0 0 1 0-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178Z" />
                    <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" />
                </svg>
            </div>
            <!-- Lock controls -->
            <div :title="!isLocked ? 'Lock board' : 'Unlock board'">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
                    stroke="currentColor" class="w-8 h-8 mx-auto mb-4 cursor-pointer" v-if="isOwner"
                    :class="{ 'hidden': isLocked }" @click="lock">
                    <path stroke-linecap="round" stroke-linejoin="round"
                        d="M16.5 10.5V6.75a4.5 4.5 0 1 0-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 0 0 2.25-2.25v-6.75a2.25 2.25 0 0 0-2.25-2.25H6.75a2.25 2.25 0 0 0-2.25 2.25v6.75a2.25 2.25 0 0 0 2.25 2.25Z" />
                </svg>
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
                    stroke="currentColor" class="w-8 h-8 mx-auto mb-4 cursor-pointer" v-if="isOwner"
                    :class="{ 'hidden': !isLocked }" @click="lock">
                    <path stroke-linecap="round" stroke-linejoin="round"
                        d="M13.5 10.5V6.75a4.5 4.5 0 1 1 9 0v3.75M3.75 21.75h10.5a2.25 2.25 0 0 0 2.25-2.25v-6.75a2.25 2.25 0 0 0-2.25-2.25H3.75a2.25 2.25 0 0 0-2.25 2.25v6.75a2.25 2.25 0 0 0 2.25 2.25Z" />
                </svg>
            </div>
            <!-- Download -->
            <div title="Download as Pdf">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
                    stroke="currentColor" class="w-8 h-8 mx-auto mb-4 cursor-pointer" v-if="isOwner" @click="download">
                    <path stroke-linecap="round" stroke-linejoin="round"
                        d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m.75 12 3 3m0 0 3-3m-3 3v-6m-1.5-9H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />
                </svg>
            </div>
            <DarkModeToggle class="w-8 h-8 mx-auto mb-4 cursor-pointer" />
            <a href="https://github.com/vijeeshr/quickretro" target="_blank">
                <svg viewBox="0 0 24 24" aria-hidden="true" class="h-8 w-8 mx-auto fill-slate-100">
                    <path fill-rule="evenodd" clip-rule="evenodd"
                        d="M12 2C6.477 2 2 6.463 2 11.97c0 4.404 2.865 8.14 6.839 9.458.5.092.682-.216.682-.48 0-.236-.008-.864-.013-1.695-2.782.602-3.369-1.337-3.369-1.337-.454-1.151-1.11-1.458-1.11-1.458-.908-.618.069-.606.069-.606 1.003.07 1.531 1.027 1.531 1.027.892 1.524 2.341 1.084 2.91.828.092-.643.35-1.083.636-1.332-2.22-.251-4.555-1.107-4.555-4.927 0-1.088.39-1.979 1.029-2.675-.103-.252-.446-1.266.098-2.638 0 0 .84-.268 2.75 1.022A9.607 9.607 0 0 1 12 6.82c.85.004 1.705.114 2.504.336 1.909-1.29 2.747-1.022 2.747-1.022.546 1.372.202 2.386.1 2.638.64.696 1.028 1.587 1.028 2.675 0 3.83-2.339 4.673-4.566 4.92.359.307.678.915.678 1.846 0 1.332-.012 2.407-.012 2.734 0 .267.18.577.688.48 3.97-1.32 6.833-5.054 6.833-9.458C22 6.463 17.522 2 12 2Z">
                    </path>
                </svg>
            </a>
        </div>
        <!-- Left Sidebar -->

        <!-- Dashboard Content -->
        <div class="w-full min-h-screen overflow-hidden">
            <div class="flex justify-center items-center bg-gray-100 dark:bg-gray-900 w-full text-red-500 dark:text-red-300 pt-1"
                v-if="isLocked">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
                    stroke="currentColor" class="w-6 h-6 mr-2">
                    <path stroke-linecap="round" stroke-linejoin="round"
                        d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
                </svg>
                Cannot add or update. Board is locked by owner.
            </div>
            <div
                class="flex flex-1 flex-col md:flex-row h-full min-h-screen bg-gray-100 dark:bg-gray-900 overflow-hidden">
                <Category v-for="column in columns" :button-text="column.text" :color="column.color"
                    :width="columnWidthClass" @add-card="add(column.id, false)"
                    @add-anonymous-card="add(column.id, true)">
                    <NewCard v-if="newCardCategory == column.id" :category="column.id" :by="user" :nickname="nickname"
                        :board="board" @added="onAdded" @invalidContent="onInvalidContent" />
                    <NewAnonymousCard v-if="newAnonymousCardCategory == column.id" :category="column.id" :by="user"
                        nickname="" :board="board" @added="onAdded" @invalidContent="onInvalidContent" />
                    <Card v-for="card in filterCards(column.id)" :card="card" :current-user="user"
                        :current-user-nickname="nickname" :board="board" :mask="isMasked"
                        :updateable="card.mine || isOwner" :key="card.id" :categories="columns" @updated="onUpdated"
                        @deleted="onDeleted" @liked="onLiked" @category-changed="onCategoryChanged"
                        @invalidContent="onInvalidContent" :locked="isLocked" />
                </Category>
            </div>
        </div>
        <!-- Dashboard Content -->

        <!-- Right Sidebar -->
        <div class="w-16 p-4">
            <Avatar v-for="user in onlineUsers" :name="user.nickname" class="w-8 h-8 ml-auto mx-auto mb-4" />
        </div>
        <!-- Right Sidebar -->
    </div>
</template>
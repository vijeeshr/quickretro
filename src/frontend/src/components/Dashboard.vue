<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue';
import Avatar from './Avatar.vue';
import Card from './Card.vue';
import Category from './Category.vue';
import NewAnonymousCard from './NewAnonymousCard.vue';
import NewCard from './NewCard.vue';
import { useRoute } from 'vue-router';
import { EventRequest, MaskEvent, MaskResponse, RegisterEvent, RegisterResponse, MessageResponse, UserClosingResponse, toSocketResponse, SaveMessageEvent, DeleteMessageEvent, DeleteMessageResponse, LikeMessageEvent, LikeMessageResponse, LockEvent, LockResponse, TimerResponse, TimerEvent, CategoryChangeEvent, CategoryChangeResponse, DeleteAllEvent, ColumnsChangeEvent, ColumnsChangeResponse } from '../models/Requests';
import { OnlineUser } from '../models/OnlineUser';
import { DraftMessage } from '../models/DraftMessage';
import { LikeMessage } from '../models/LikeMessage';
// import { jsPDF } from 'jspdf';
// import autoTable from 'jspdf-autotable';
import { Dialog, DialogPanel, DialogTitle } from '@headlessui/vue'
import CountdownTimer from './CountdownTimer.vue';
import TimerPanel from './TimerPanel.vue';
import { areBoardColumnsVisuallySame, exceedsEventRequestMaxSize, formatDate, logMessage } from '../utils';
import { useToast } from 'vue-toast-notification';
// import 'vue-toast-notification/dist/theme-sugar.css';
// import 'vue-toast-notification/dist/theme-bootstrap.css';
import 'vue-toast-notification/dist/theme-default.css';
import DarkModeToggle from './DarkModeToggle.vue';
import { BoardColumn } from '../models/BoardColumn';
import { CategoryChangeMessage } from '../models/CategoryChangeMessage';
import { useI18n } from 'vue-i18n';
import { useLanguage } from '../composables/useLanguage';
import CategoryEditor from './CategoryEditor.vue';
import { CategoryDefinition } from '../models/CategoryDefinition';
import { defaultCategories } from '../constants/defaultCategories';

const { locale, setLocale, languageOptions } = useLanguage()
const { t } = useI18n()
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
const isSpotlightOn = ref(false)
const spotlightFor = ref('')
const boardExpiryLocalTime = ref('')
let socket: WebSocket

const cards = ref<MessageResponse[]>([]) // Todo: Rework models
const commentsMap = ref(new Map<string, MessageResponse[]>()) // map of messageId -> [comments]

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

const isDeleteAllDialogOpen = ref(false)
const setIsDeleteAllDialogOpen = (value: boolean) => {
    isDeleteAllDialogOpen.value = value
}

const isLanguageDialogOpen = ref(false)
const setIsLanguageDialogOpen = (value: boolean) => {
    isLanguageDialogOpen.value = value
}

const isColumnEditDialogOpen = ref(false)
const setIsColumnEditDialogOpen = (value: boolean) => {
    isColumnEditDialogOpen.value = value
}

const toast = useToast()

const onOneMinuteLeftWarning = () => {
    toast.info(t('dashboard.timer.oneMinuteLeft'), { pauseOnHover: false })
}

const onCountdownCompleted = () => {
    toast.error(t('dashboard.timer.timeCompleted'), { pauseOnHover: false, duration: 4000 })
}

const filterCards = (category: string) => {
    return cards.value.filter(c => c.cat.toLowerCase() === category.toLowerCase())
}

const filterComments = (messageId: string): MessageResponse[] => {
    return commentsMap.value.get(messageId) || []
}

const getCommentIds = (messageId: string): string[] => {
    return commentsMap.value.get(messageId)?.map(c => c.id) ?? []
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

// Todo: Card counts by name isn't ideal. It will be incorrect when two users have the exact same nickname.
const cardCountsByName = computed(() => {
    const counts: Record<string, number> = {}
    for (const card of cards.value) {
        counts[card.nickname] = (counts[card.nickname] || 0) + 1
    }
    return counts
})

const onlineUsersCardsCount = computed(() => {
    return onlineUsers.value.map(user => ({
        nickname: user.nickname,
        cardsCount: cardCountsByName.value[user.nickname] || 0
    }))
})

const usersWithCards = computed(() => Object.keys(cardCountsByName.value))

const nextSpotlight = () => {
    if (usersWithCards.value.length === 0) return
    let currentIndex = usersWithCards.value.indexOf(spotlightFor.value)
    // If spotlightFor is not in the list (currentIndex = -1), start from the first item
    const nextIndex = currentIndex === -1 ? 0 : (currentIndex + 1) % usersWithCards.value.length
    spotlightFor.value = usersWithCards.value[nextIndex]
}

const prevSpotlight = () => {
    if (usersWithCards.value.length === 0) return
    let currentIndex = usersWithCards.value.indexOf(spotlightFor.value)
    // If spotlightFor is not in the list, start from the last item
    const prevIndex = currentIndex === -1 ? usersWithCards.value.length - 1 : (currentIndex - 1 + usersWithCards.value.length) % usersWithCards.value.length
    spotlightFor.value = usersWithCards.value[prevIndex]
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

const openSpotlight = () => {
    if (usersWithCards.value.length === 0) {
        closeSpotlight()
        toast.info(t('dashboard.spotlight.noCardsToFocus'), { pauseOnHover: false })
        return
    }
    spotlightFor.value = usersWithCards.value[0]
    isSpotlightOn.value = !isSpotlightOn.value
}

const closeSpotlight = () => {
    isSpotlightOn.value = false
    spotlightFor.value = ''
}

const showSpotlightFor = (name: string) => {
    spotlightFor.value = name
    isSpotlightOn.value = true
}

const notifyForLostMessages = () => {
    toast.error(t('dashboard.lock.discardChanges'))
}

const newCardCreationInProgress = computed(() => {
    return newCardCategory.value.trim() !== '' || newAnonymousCardCategory.value.trim() !== ''
})

const clearNewCards = () => {
    newCardCategory.value = '' //unmount newCard
    newAnonymousCardCategory.value = '' //unmount newAnonymousCard
}

const onAdded = (card: DraftMessage) => {
    logMessage('newcontent received:', card)
    clearNewCards()
    const nicknameToSend = card.anon === true ? '' : nickname
    dispatchEvent<SaveMessageEvent>("msg", { id: card.id, by: user, nickname: nicknameToSend, grp: board, msg: card.msg, cat: card.cat, anon: card.anon, pid: card.pid })
}

const onCommentAdded = (comment: DraftMessage) => {
    logMessage('newcontent received:', comment)
    // Todo: clear the comment field..maybe in the Card component?
    dispatchEvent<SaveMessageEvent>("msg", { id: comment.id, by: user, nickname: nickname, grp: board, msg: comment.msg, cat: comment.cat, anon: comment.anon, pid: comment.pid })
}

const onInvalidContent = (errorMessage: string) => {
    toast.error(errorMessage)
}

const onCommentInvalidContent = (errorMessage: string) => {
    toast.error(errorMessage)
}

const onDiscard = () => {
    clearNewCards()
}

const onUpdated = (card: DraftMessage) => {
    logMessage('Updated content received:', card)
    dispatchEvent<SaveMessageEvent>("msg", { id: card.id, by: user, nickname: nickname, grp: board, msg: card.msg, cat: card.cat, anon: false, pid: card.pid })
}

const onCommentUpdated = (comment: DraftMessage) => {
    logMessage('Updated content received:', comment)
    dispatchEvent<SaveMessageEvent>("msg", { id: comment.id, by: user, nickname: nickname, grp: board, msg: comment.msg, cat: comment.cat, anon: false, pid: comment.pid })
}

const onDeleted = (cardId: string) => {
    const commentIds = getCommentIds(cardId)
    dispatchEvent<DeleteMessageEvent>("del", { msgId: cardId, by: user, grp: board, commentIds: commentIds })
}

const onCommentDeleted = (commendId: string) => {
    dispatchEvent<DeleteMessageEvent>("del", { msgId: commendId, by: user, grp: board, commentIds: [] })
}

const onLiked = (likeMessage: LikeMessage) => {
    dispatchEvent<LikeMessageEvent>("like", { msgId: likeMessage.msgId, by: user, like: likeMessage.like })
}

const onCategoryChanged = (pyl: CategoryChangeMessage) => {
    const commentIds = getCommentIds(pyl.msgId)
    dispatchEvent<CategoryChangeEvent>("catchng", { msgId: pyl.msgId, commentIds: commentIds, by: user, grp: board, newcat: pyl.newCategoryId, oldcat: pyl.oldCategoryId })
}

const mask = () => {
    dispatchEvent<MaskEvent>("mask", { by: user, grp: board, mask: !isMasked.value })
}

const lock = () => {
    dispatchEvent<LockEvent>("lock", { by: user, grp: board, lock: !isLocked.value })
}

const deleteAll = () => {
    dispatchEvent<DeleteAllEvent>("delall", { by: user, grp: board })
}

const onTimerStart = (expiryDurationInSeconds: number) => {
    dispatchEvent<TimerEvent>("timer", { by: user, grp: board, expiryDurationInSeconds: expiryDurationInSeconds, stop: false }) // "stop" is ignored
    setIsTimerDialogOpen(false)
}

const onTimerStop = () => {
    dispatchEvent<TimerEvent>("timer", { by: user, grp: board, expiryDurationInSeconds: 0, stop: true }) // "expiryDurationInSeconds" is ignored
    setIsTimerDialogOpen(false)
}

const saveCategoryChanges = () => {
    const enabledCols: BoardColumn[] = mergedCategories.value.filter(c => c.enabled === true)
        .map(c => ({
            id: c.id,
            text: c.text.trim() || t(`dashboard.columns.${c.id}`),
            isDefault: c.text === '' || c.text === t(`dashboard.columns.${c.id}`),
            color: c.color,
            pos: c.pos
        }))

    if (areBoardColumnsVisuallySame(enabledCols, columns.value)) {
        logMessage('Columns unchanged. Not dispatching.')
        return
    }
    
    if (exceedsEventRequestMaxSize<ColumnsChangeEvent>("colreset", { by: user, grp: board, columns: enabledCols })) {
        toast.error(t('common.contentOverloadError'))
        return
    }

    dispatchEvent<ColumnsChangeEvent>("colreset", { by: user, grp: board, columns: enabledCols })
}

// const getRGBizedColor = (color: string): [number, number, number] => {
//     switch (color) {
//         case "green":
//             return [74, 222, 128]
//         case "red":
//             return [248, 113, 113]
//         case "yellow":
//             return [250, 204, 21]
//         case "fuchsia":
//             return [232, 121, 249]
//         case "orange":
//             return [251, 146, 60]
//         default:
//             return [128, 128, 128]
//     }
// }

// const getHexizedColor = (color: string): string => {
//     const rgb = getRGBizedColor(color);
//     return `#${rgb.map(c =>
//         c.toString(16).padStart(2, '0').toUpperCase()
//     ).join('')}`
// }

const getHexizedColor = (color: string): string => {
    switch (color) {
        case "green":
            return "#4ADE80"
        case "red":
            return "#F87171"
        case "yellow":
            return "#FACC15"
        case "fuchsia":
            return "#E879F9"
        case "orange":
            return "#FB923C"
        default:
            return "#808080"
    }
}

const generateDocument = () => {
    // if (locale.value === 'zhCN' || locale.value === 'ja' || locale.value === 'ko' || locale.value === 'ru' || locale.value === 'uk') {
    //     print()
    // } else {
    //     download()
    // }
    print()
}

// const download = async () => {
//     try {
//         const { default: jsPDF } = await import('jspdf')
//         const { default: autoTable } = await import('jspdf-autotable')

//         const doc = new jsPDF({
//             orientation: "portrait",
//             unit: "in",
//             format: "letter"
//         })

//         // text is placed using x, y coordinates
//         doc.setFontSize(16).text(`${t('common.board')} - ${boardName.value}`, 0.5, 1.0)
//         // create a line under heading
//         doc.setLineWidth(0.01).line(0.5, 1.1, 8.0, 1.1)

//         for (const col of columns.value) {
//             const headerText = col.isDefault ? t(`dashboard.columns.${col.id}`) : col.text
//             const itemsForExport = cards.value.filter(c => c.cat.toLowerCase() === col.id.toLowerCase())
//                 .map(c => [c.msg])

//             // Using autoTable plugin               
//             autoTable(doc, {
//                 head: [[headerText]],
//                 body: itemsForExport,
//                 headStyles: { fillColor: getRGBizedColor(col.color) },
//                 margin: { left: 0.5, top: 1.25 },
//             })
//         }

//         // Footer
//         doc
//             .setFont("times")
//             .setFontSize(11)
//             .setTextColor("gray")
//             .text(
//                 `${t('dashboard.printFooter')} QuickRetro ( https://quickretro.app )`,
//                 doc.internal.pageSize.width / 2, // Center horizontally
//                 doc.internal.pageSize.height - 0.5, // Position vertically
//                 { align: "center" }
//             )

//         doc.save(`quickretro.pdf`)

//         // Using array of sentences
//         // doc
//         //     .setFont("helvetica")
//         //     .setFontSize(12)
//         //     .text(moreText, 0.5, 3.5, { align: "left", maxWidth: 7.5 });
//     } catch (error) {
//         // toast.error(t('dashboard.download.error'), { pauseOnHover: false })
//         console.error('PDF download failed:', error)
//     }
// }

const print = async () => {
    const printWindow = window.open('', '_blank')
    if (!printWindow) return

    try {
        const { default: useSanitize } = await import('../composables/useSanitize')
        const { sanitize } = useSanitize()

        // Create document structure
        // const doc = printWindow.document
        const html = `
        <html>
            <head>
                <title>QuickRetro</title>
                <style>
                    @media print {
                        @page {
                            margin: 20px;
                            size: A4 portrait;
                        }
                        
                        body { 
                            margin: 0;
                            padding: 20px;
                            -webkit-print-color-adjust: exact;
                            print-color-adjust: exact;
                            font-family: Arial, sans-serif, "Noto Sans CJK SC", "Hiragino Sans GB";
                        }

                        .print-section {
                            margin: 0;
                            padding: 0;
                            width: 100%;
                        }

                        .print-title {
                            font-size: 1.5rem;
                            font-weight: normal;
                            border-bottom: 1px solid black;
                            margin: 0 0 1rem 0;
                            padding: 0.5rem;
                            page-break-after: avoid;
                        }

                        .print-columns {
                            columns: 1;
                            column-fill: auto;
                            page-break-inside: avoid;
                        }

                        .print-column {
                            page-break-inside: avoid;
                            break-inside: avoid-page;
                            margin-bottom: 1rem;
                        }

                        .print-category {
                            font-weight: bold;
                            padding: 0.25rem;
                            page-break-after: avoid;
                        }

                        .print-card {
                            padding: 0.5rem;
                            page-break-inside: avoid;
                            color: #505050;
                            font-size: 1rem;
                            word-break: break-all;
                        }

                        .print-card:nth-child(even) {
                            background-color: #f3f3f3;
                        }

                        .print-card:nth-child(odd) {
                            background-color: #ffffff;
                        }

                        .print-footer {
                            position: fixed;
                            bottom: 0;
                            left: 0;
                            right: 0;
                            font-size: 0.8rem;
                            text-align: center;
                            padding: 0.5rem;
                            color: gray;
                        }
                    }
                </style>
            </head>
            <body>
                <div class="print-section">
                <h1 class="print-title">${t('common.board')} - ${sanitize(boardName.value)}</h1>
                <div class="print-columns">
                    ${columns.value.map(col => `
                    <div class="print-column">
                        <div class="print-category" style="background-color:${getHexizedColor(col.color)};color:white">${col.isDefault ? t(`dashboard.columns.${col.id}`) : sanitize(col.text)}</div>
                        ${cards.value
                .filter(c => c.cat.toLowerCase() === col.id.toLowerCase() && c.msg && c.msg.trim() !== '')
                .map(c => `<div class="print-card">${sanitize(c.msg)}</div>`)
                .join('')}
                    </div>
                    `).join('')}
                </div>
                <footer class="print-footer">
                    ${t('dashboard.printFooter')} QuickRetro ( https://quickretro.app )
                </footer>
                </div>
            </body>
        </html>`

        // // Modern parsing method
        // const parser = new DOMParser()
        // const parsedDoc = parser.parseFromString(html, 'text/html')
        // // Clone nodes to new document
        // Array.from(parsedDoc.documentElement.childNodes).forEach(node => {
        //     doc.documentElement.appendChild(doc.importNode(node, true))
        // })
        // // Handle printing
        // const printAndClose = () => {
        //     printWindow.print()
        //     setTimeout(() => printWindow.close(), 500)
        // }
        // if (doc.fonts) {
        //     doc.fonts.ready.then(printAndClose)
        // } else {
        //     printWindow.addEventListener('load', printAndClose)
        // }

        // Write in one synchronous block. Chrome-safe and Firefox-safe
        printWindow.document.open()
        // @ts-ignore - document.write is required for print reliability
        printWindow.document.write(html)
        printWindow.document.close()
        // Once fully rendered, print
        printWindow.onload = async () => {
            if (printWindow.document.fonts) {
                await printWindow.document.fonts.ready
            }
            printWindow.print()
        }

    } catch (error) {
        // toast.error(t('dashboard.download.error'), { pauseOnHover: false })
        console.error('Print failed:', error)
    } finally {
        setTimeout(() => {
            if (printWindow && !printWindow.closed) {
                printWindow.close();
            }
        }, 2000); // Force close after 2s
    }
}

const share = () => {
    isShareDialogOpen.value = true
}

const copyShareLink = async () => {
    try {
        await navigator.clipboard.writeText(shareLink)
        toast.success(t('dashboard.share.linkCopied'))
        setIsShareDialogOpen(false)
    } catch (err) {
        toast.error(t('dashboard.share.linkCopyError'))
    }
}

const openLanguageDialog = () => {
    isLanguageDialogOpen.value = true
}

const openDeleteAllDialog = () => {
    isDeleteAllDialogOpen.value = true
}

const timerSettings = () => {
    if (isOwner.value) {
        isTimerDialogOpen.value = true
    }
}

const openColumnEditDialog = () => {
    if (!isOwner) return
    mergeCategories()
    isColumnEditDialogOpen.value = true
}

const onRegisterResponse = (response: RegisterResponse) => {
    timerExpiresInSeconds.value = response.timerExpiresInSeconds // This always gets set. Todo: find a better way to sync timer.
    boardExpiryLocalTime.value = formatDate(response.boardExpiryUtcSeconds)
    onlineUsers.value = []
    onlineUsers.value.push(...response.users) // Todo: find a better way

    // response.mine == true means RegisterResponse is for the current User's RegisterEvent request.
    // This prevents unnecessarily updating values/pushing messages in the ref for other users RegisterEvents. 
    // RegisterEvent happens just once per user in the beginning i.e. when user loads/reloads the page.
    if (response.mine) {
        boardName.value = response.boardName
        isOwner.value = response.isBoardOwner
        isMasked.value = response.boardMasking
        isLocked.value = response.boardLock
        // columns.value = []
        // columns.value.push(...response.columns) // Todo: find a better way

        // const sortedColumns = response.columns
        //     .slice() // create a shallow copy (to avoid mutating response.columns)
        //     .sort((a, b) => a.pos - b.pos)
        // columns.value.push(...sortedColumns)
        columns.value = response.columns
            .slice()
            .sort((a, b) => a.pos - b.pos)

        cards.value = []
        cards.value.push(...response.messages)

        // Build Map: messageId → comments[]
        const map = new Map<string, MessageResponse[]>()
        response.comments.forEach(comment => {
            if (!map.has(comment.pid)) map.set(comment.pid, [])
            map.get(comment.pid)!.push(comment)
        })
        commentsMap.value = map
    }

    // Show expiration notification for newly created board. Show only for board creator/owner
    if (response.isBoardOwner && response.notifyNewBoardExpiry) {
        const initialExpiryMsg = `${t('dashboard.autoDeleteScheduleBase', { date: boardExpiryLocalTime.value })}${t('dashboard.autoDeleteScheduleAddon')}`
        toast.warning(initialExpiryMsg, { pauseOnHover: false, duration: 10000 })
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
    if (isLocked.value && newCardCreationInProgress.value) {
        // Board lock received when the user is adding new messages.
        // The message will be forcefully discarded.
        clearNewCards()
        notifyForLostMessages()
    }
}

const onSaveMessageResponse = (response: MessageResponse) => {
    if (response.pid) {
        // It's a comment — store under its parent message
        const existingComments = commentsMap.value.get(response.pid) || []
        const index = existingComments.findIndex(c => c.id === response.id)

        if (index === -1) {
            // New comment
            commentsMap.value.set(response.pid, [...existingComments, response])
        } else {
            // Update existing comment
            const updated = [...existingComments]
            updated[index] = response
            commentsMap.value.set(response.pid, updated)
        }
    } else {
        // It's a top-level message
        const index = cards.value.findIndex(x => x.id === response.id)
        if (index === -1) {
            cards.value.push(response)
        } else {
            cards.value[index] = response
        }
    }
}

const onDeleteMessageResponse = (response: DeleteMessageResponse) => {
    // Delete message
    let messageIndex = cards.value.findIndex(x => x.id === response.id)
    if (messageIndex !== -1) {
        commentsMap.value.delete(response.id) // remove associated comments
        cards.value.splice(messageIndex, 1)
        // Re-adjust spotlight
        if (isSpotlightOn.value) {
            if (usersWithCards.value.length === 0) {
                closeSpotlight();
            } else if (!usersWithCards.value.includes(spotlightFor.value)) {
                nextSpotlight();
            }
        }
        return // Just return no need to proceed further
    }

    // Delete comment
    for (const [parentId, comments] of commentsMap.value.entries()) {
        const index = comments.findIndex(c => c.id === response.id)
        if (index !== -1) {
            const updated = [...comments]
            updated.splice(index, 1)
            updated.length > 0
                ? commentsMap.value.set(parentId, updated)
                : commentsMap.value.delete(parentId)
            return
        }
    }
}

const onDeleteAllResponse = () => {
    window.location.reload()
}

const onCategoryChangeResponse = (response: CategoryChangeResponse) => {
    const msgId = response.id
    const newCat = response.newcat
    // Update category of message
    let index = cards.value.findIndex(x => x.id === msgId)
    if (index !== -1) {
        cards.value[index].cat = newCat
    }
    // Update category of associated comments
    const comments = commentsMap.value.get(msgId)
    if (comments && comments.length > 0) {
        const updatedComments = comments.map(c => ({
            ...c,
            cat: newCat,
        }))
        commentsMap.value.set(msgId, updatedComments)
    }
    // // Mutate in-place instead of creating a new array. 
    // // Faster for large comment lists, but reactivity might not always trigger UI updates if Vue doesn’t deeply watch the comment array (depends on where it’s used in templates).
    // const comments = commentsMap.value.get(msgId)
    // if (comments) {
    //     comments.forEach(c => (c.cat = newCat))
    // }
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

const onColumnsChangeResponse = (response: ColumnsChangeResponse) => {
    columns.value = response.columns
        .slice()
        .sort((a, b) => a.pos - b.pos)
}

// For Edit Categories feature
const mergedCategories = ref<CategoryDefinition[]>([])
const isCategorySelectionValid = ref(true)

const hasCardsInDisabledCategories = computed(() => {
    // Collect disabled category IDs
    const disabledCatIds = mergedCategories.value
        .filter(cat => !cat.enabled)
        .map(cat => cat.id)
    if (disabledCatIds.length === 0) return false

    // Check if any card belongs to any disabled category
    return cards.value.some(card => disabledCatIds.includes(card.cat))
})

const mergeCategories = () => {
    const defaultCats: CategoryDefinition[] = [...defaultCategories]

    const map = new Map<string, BoardColumn>(columns.value.map(c => [c.id, c]))

    const merged = defaultCats
        .map((d) => {
            const override = map.get(d.id)
            return {
                id: d.id,
                color: d.color,
                colorClass: d.colorClass,
                text: override?.isDefault ? d.text : override?.text ?? d.text, // override?.text ?? d.text,
                enabled: override !== undefined ? true : false, // override is present means, the column has been defined
                pos: override?.pos ?? defaultCats.length
            }
        })

    merged.sort((a, b) => {
        // sort by "enabled==true" first, then by "pos"
        if (a.enabled !== b.enabled) {
            return a.enabled ? -1 : 1; // enabled=true first
        }
        return a.pos - b.pos;
    })

    mergedCategories.value = merged.map((col, i) => ({
        ...col,
        pos: i + 1
    }))
}

const handleCategoryTextUpdate = (update: { id: string, text: string }) => {
    const cat = mergedCategories.value.find(c => c.id === update.id)
    if (cat) {
        cat.text = update.text
    }
}
const handleCategoryToggle = (update: { id: string, enabled: boolean }) => {
    const cat = mergedCategories.value.find(c => c.id === update.id)
    if (cat) {
        cat.enabled = update.enabled
    }
}
const handleCategoriesReorder = (reorderedCategories: CategoryDefinition[]) => {
    mergedCategories.value = reorderedCategories
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
    if (event.code === 1008 && event.reason === 'BOARDNOTFOUND') {
        toast.error(t('dashboard.notExists'), { duration: 0 })
    }
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
            case "delall":
                onDeleteAllResponse()
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
            case "colreset":
                onColumnsChangeResponse(response)
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
        toast.error(t('dashboard.offline'), { pauseOnHover: false, duration: 4000 })
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

        <!-- Focus navigation panel -->
        <div v-if="isSpotlightOn && usersWithCards.length > 0"
            class="fixed flex items-center gap-2 top-4 left-1/2 transform -translate-x-1/2 text-white bg-black/50 border border-gray-500 px-4 py-2 rounded-lg shadow-md z-[60]">
            <button class="rounded-md hover:bg-gray-200 hover:text-gray-600" @click="prevSpotlight">
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="m15 18-6-6 6-6" />
                </svg>
            </button>
            <Avatar v-if="spotlightFor !== ''" class="min-w-32" :name="spotlightFor" view-type="Badge" />
            <div v-else
                class="min-w-32 inline-flex items-center justify-center overflow-hidden rounded-md px-3 py-1 bg-gray-300">
                <span class="font-medium text-xs cursor-default text-gray-600 select-none">{{ t('common.anonymous')
                    }}</span>
            </div>
            <button class="rounded-md hover:bg-gray-200 hover:text-gray-700 mr-3" @click="nextSpotlight">
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="m9 18 6-6-6-6" />
                </svg>
            </button>
            <button class="rounded-md hover:bg-gray-200 hover:text-gray-700 ml-auto" @click="closeSpotlight">
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M18 6 6 18" />
                    <path d="m6 6 12 12" />
                </svg>
            </button>
        </div>

        <!-- Dialog for Timer settings -->
        <Dialog :open="isTimerDialogOpen" @close="setIsTimerDialogOpen" class="relative z-50">
            <!-- The backdrop, rendered as a fixed sibling to the panel container -->
            <div class="fixed inset-0 bg-black/30" aria-hidden="true" />
            <div class="fixed inset-0 flex w-screen items-center justify-center p-4">
                <DialogPanel class="rounded-2xl bg-white dark:bg-gray-800 p-6 text-left align-middle shadow-xl">
                    <DialogTitle as="h3"
                        class="text-lg font-medium leading-6 text-gray-900 dark:text-gray-200 select-none">
                        {{ t('dashboard.timer.title') }}
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
                <DialogPanel
                    class="max-w-sm rounded-2xl bg-white dark:bg-gray-800 p-6 text-left align-middle shadow-xl">
                    <DialogTitle as="h3"
                        class="text-lg font-medium leading-6 text-gray-900 dark:text-gray-200 select-none">{{
                            t('dashboard.share.title') }}
                    </DialogTitle>
                    <article
                        class="select-all break-all bg-slate-100 dark:bg-slate-700 dark:text-gray-100 my-6 rounded-sm">
                        {{
                            shareLink }}</article>
                    <div class="mt-4">
                        <button type="button"
                            class="px-4 py-2 text-sm w-full shadow-md font-medium rounded-md border bg-sky-100 hover:bg-sky-400 border-sky-300 text-sky-600 hover:text-white dark:bg-sky-800 dark:hover:bg-sky-600 dark:border-sky-700 dark:text-sky-100 hover:border-transparent select-none focus:outline-none focus:ring-0"
                            @click="copyShareLink">
                            {{ t('common.copy') }}
                        </button>
                    </div>
                </DialogPanel>
            </div>
        </Dialog>

        <!-- Dialog for Language selection -->
        <Dialog :open="isLanguageDialogOpen" @close="setIsLanguageDialogOpen" class="relative z-50">
            <!-- The backdrop, rendered as a fixed sibling to the panel container -->
            <div class="fixed inset-0 bg-black/30" aria-hidden="true" />
            <div class="fixed inset-0 flex w-screen items-center justify-center p-4">
                <DialogPanel
                    class="max-w-sm rounded-2xl bg-white dark:bg-gray-800 p-6 text-left align-middle shadow-xl">
                    <div class="flex flex-wrap gap-2">
                        <button v-for="lang in languageOptions" :key="lang.code"
                            @click="setLocale(lang.code); setIsLanguageDialogOpen(false)"
                            class="px-3 py-1.5 text-sm rounded-lg border transition-colors" :class="[
                                locale === lang.code
                                    ? 'bg-blue-100 border-blue-200 text-blue-800 dark:bg-blue-900/30 dark:border-blue-800 dark:text-blue-200'
                                    : 'bg-gray-100 border-gray-200 text-gray-700 hover:bg-gray-200 dark:bg-slate-700 dark:border-slate-600 dark:text-slate-200 dark:hover:bg-slate-600'
                            ]">
                            {{ lang.name }}
                        </button>
                    </div>
                </DialogPanel>
            </div>
        </Dialog>

        <!-- Dialog for DeleteAll Confirmation -->
        <Dialog :open="isDeleteAllDialogOpen" @close="setIsDeleteAllDialogOpen" class="relative z-50">
            <!-- The backdrop, rendered as a fixed sibling to the panel container -->
            <div class="fixed inset-0 bg-black/30" aria-hidden="true" />
            <div class="fixed inset-0 flex w-screen items-center justify-center p-4">
                <DialogPanel
                    class="max-w-sm rounded-2xl bg-white dark:bg-gray-800 p-6 text-left align-middle shadow-xl">
                    <DialogTitle as="h3"
                        class="text-lg font-medium leading-6 text-gray-900 dark:text-gray-200 select-none">{{
                            t('dashboard.delete.title') }}
                    </DialogTitle>
                    <p class="text-gray-900 dark:text-gray-200 select-none">
                        {{ t('dashboard.delete.text') }}
                    </p>
                    <div class="flex justify-center mt-4 space-x-2">
                        <button type="button"
                            class="px-4 py-2 text-sm w-full shadow-md font-medium rounded-md border bg-sky-100 hover:bg-sky-400 border-sky-300 text-sky-600 hover:text-white dark:bg-sky-800 dark:hover:bg-sky-600 dark:border-sky-700 dark:text-sky-100 hover:border-transparent select-none focus:outline-none focus:ring-0"
                            @click="setIsDeleteAllDialogOpen(false)">
                            {{ t('dashboard.delete.cancelDelete') }}
                        </button>
                        <button type="button"
                            class="px-4 py-2 text-sm w-full shadow-md font-medium rounded-md border bg-red-100 hover:bg-red-400 border-red-300 text-red-600 hover:text-white hover:border-transparent dark:bg-red-800 dark:hover:bg-red-600 dark:border-red-700 dark:text-red-100 select-none focus:outline-none focus:ring-0"
                            @click="deleteAll(); setIsDeleteAllDialogOpen(false)">
                            {{ t('dashboard.delete.continueDelete') }}
                        </button>
                    </div>
                    <p v-if="boardExpiryLocalTime !== ''"
                        class="flex justify-items-normal mt-4 text-xs text-slate-400 select-none max-w-xs">
                        {{ t('dashboard.autoDeleteScheduleBase', { date: boardExpiryLocalTime }) }}.
                    </p>
                </DialogPanel>
            </div>
        </Dialog>

        <!-- Dialog for Column editing -->
        <Dialog :open="isColumnEditDialogOpen" @close="setIsColumnEditDialogOpen" class="relative z-50">
            <!-- The backdrop, rendered as a fixed sibling to the panel container -->
            <div class="fixed inset-0 bg-black/30" aria-hidden="true" />
            <div class="fixed inset-0 flex w-screen items-center justify-center p-4">
                <DialogPanel
                    class="w-full max-w-[356px] min-w-[240px] rounded-2xl bg-white dark:bg-gray-800 p-6 text-left align-middle shadow-xl">
                    <CategoryEditor :categories="mergedCategories" @category-text-update="handleCategoryTextUpdate"
                        @category-toggle="handleCategoryToggle" @categories-reorder="handleCategoriesReorder"
                        @valid="(val: boolean) => isCategorySelectionValid = val">
                    </CategoryEditor>
                    <p v-show="hasCardsInDisabledCategories"
                        class="text-sm text-red-600 dark:text-red-300 mt-2 select-none">
                        {{ t('dashboard.columns.cannotDisable') }}</p>
                    <button type="button"
                        class="px-4 py-2 mt-2 text-sm w-full shadow-md font-medium rounded-md border bg-sky-100 hover:bg-sky-400 border-sky-300 text-sky-600 hover:text-white hover:border-transparent disabled:bg-gray-300 disabled:text-gray-500 disabled:border-gray-400 disabled:cursor-not-allowed dark:disabled:bg-gray-300 dark:disabled:text-gray-500 dark:disabled:border-gray-400 dark:bg-sky-800 dark:hover:bg-sky-600 dark:border-sky-700 dark:text-sky-100 select-none focus:outline-none focus:ring-0"
                        :disabled="hasCardsInDisabledCategories || !isCategorySelectionValid"
                        @click="saveCategoryChanges">
                        {{ t('dashboard.columns.update')}}
                    </button>
                </DialogPanel>
            </div>
        </Dialog>

        <!-- Left Sidebar -->
        <div class="w-16 p-3">
            <!-- Timer -->
            <CountdownTimer :timeLeftInSeconds="timerExpiresInSeconds" :title="t('dashboard.timer.tooltip')"
                class="inline-flex items-center justify-center overflow-hidden rounded-full w-10 h-10 text-[0.825rem] leading-[1rem] font-bold text-white ml-auto mx-auto mb-4"
                :class="isOwner ? 'cursor-pointer' : 'cursor-default'" @click="timerSettings"
                @countdown-progress-update="onCountdownProgressUpdate" @one-minute-left-warning="onOneMinuteLeftWarning"
                @countdown-completed="onCountdownCompleted" />
            <!-- Share -->
            <div :title="t('dashboard.share.toolTip')">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
                    stroke="currentColor" class="w-8 h-8 mx-auto mb-4 cursor-pointer" @click="share">
                    <path stroke-linecap="round" stroke-linejoin="round"
                        d="M7.217 10.907a2.25 2.25 0 1 0 0 2.186m0-2.186c.18.324.283.696.283 1.093s-.103.77-.283 1.093m0-2.186 9.566-5.314m-9.566 7.5 9.566 5.314m0 0a2.25 2.25 0 1 0 3.935 2.186 2.25 2.25 0 0 0-3.935-2.186Zm0-12.814a2.25 2.25 0 1 0 3.933-2.185 2.25 2.25 0 0 0-3.933 2.185Z" />
                </svg>
            </div>
            <!-- Mask controls -->
            <div :title="!isMasked ? t('dashboard.mask.maskTooltip') : t('dashboard.mask.unmaskTooltip')">
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
            <div :title="!isLocked ? t('dashboard.lock.lockTooltip') : t('dashboard.lock.unlockTooltip')">
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
            <!-- Print -->
            <div :title="t('dashboard.print.tooltip')">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
                    stroke="currentColor" class="w-8 h-8 mx-auto mb-4 cursor-pointer" v-if="isOwner"
                    @click="generateDocument">
                    <path stroke-linecap="round" stroke-linejoin="round"
                        d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m.75 12 3 3m0 0 3-3m-3 3v-6m-1.5-9H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />
                </svg>
            </div>
            <DarkModeToggle class="w-8 h-8 mx-auto mb-4 cursor-pointer" />
            <!-- Focus -->
            <div :title="t('dashboard.spotlight.tooltip')" class="w-8 h-8 mx-auto mb-4 cursor-pointer">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                    stroke-width="2" stroke-linecap="round" stroke-linejoin="round" @click="openSpotlight">
                    <circle cx="12" cy="12" r="3" />
                    <path d="M3 7V5a2 2 0 0 1 2-2h2" />
                    <path d="M17 3h2a2 2 0 0 1 2 2v2" />
                    <path d="M21 17v2a2 2 0 0 1-2 2h-2" />
                    <path d="M7 21H5a2 2 0 0 1-2-2v-2" />
                </svg>
            </div>
            <!-- Language picker -->
            <div :title="t('dashboard.language.tooltip')">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                    stroke-width="2" stroke-linecap="round" stroke-linejoin="round" @click="openLanguageDialog"
                    class="w-8 h-8 mx-auto mb-4 cursor-pointer">
                    <circle cx="12" cy="12" r="10" />
                    <path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20" />
                    <path d="M2 12h20" />
                </svg>
            </div>
            <!-- Delete All-->
            <div :title="t('dashboard.delete.tooltip')">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
                    stroke="currentColor" class="w-8 h-8 mx-auto mb-4 cursor-pointer" v-if="isOwner"
                    @click="openDeleteAllDialog">
                    <path stroke-linecap="round" stroke-linejoin="round"
                        d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
                </svg>
            </div>
            <a href="https://github.com/vijeeshr/quickretro" target="_blank" rel="noopener noreferrer">
                <svg viewBox="0 0 24 24" aria-hidden="true" class="h-8 w-8 mx-auto mb-4 fill-slate-100">
                    <path fill-rule="evenodd" clip-rule="evenodd"
                        d="M12 2C6.477 2 2 6.463 2 11.97c0 4.404 2.865 8.14 6.839 9.458.5.092.682-.216.682-.48 0-.236-.008-.864-.013-1.695-2.782.602-3.369-1.337-3.369-1.337-.454-1.151-1.11-1.458-1.11-1.458-.908-.618.069-.606.069-.606 1.003.07 1.531 1.027 1.531 1.027.892 1.524 2.341 1.084 2.91.828.092-.643.35-1.083.636-1.332-2.22-.251-4.555-1.107-4.555-4.927 0-1.088.39-1.979 1.029-2.675-.103-.252-.446-1.266.098-2.638 0 0 .84-.268 2.75 1.022A9.607 9.607 0 0 1 12 6.82c.85.004 1.705.114 2.504.336 1.909-1.29 2.747-1.022 2.747-1.022.546 1.372.202 2.386.1 2.638.64.696 1.028 1.587 1.028 2.675 0 3.83-2.339 4.673-4.566 4.92.359.307.678.915.678 1.846 0 1.332-.012 2.407-.012 2.734 0 .267.18.577.688.48 3.97-1.32 6.833-5.054 6.833-9.458C22 6.463 17.522 2 12 2Z">
                    </path>
                </svg>
            </a>
            <a href="https://quickretro.app/guide/dashboard" target="_blank" rel="noopener noreferrer">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
                    stroke="currentColor" class="h-8 w-8 mx-auto">
                    <path stroke-linecap="round" stroke-linejoin="round"
                        d="M9.879 7.519c1.171-1.025 3.071-1.025 4.242 0 1.172 1.025 1.172 2.687 0 3.712-.203.179-.43.326-.67.442-.745.361-1.45.999-1.45 1.827v.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 5.25h.008v.008H12v-.008Z" />
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
                {{ t('dashboard.lock.message') }}
            </div>
            <div
                class="flex flex-1 flex-col md:flex-row h-full min-h-screen bg-gray-100 dark:bg-gray-900 overflow-hidden">
                <Category v-for="column in columns" :key="column.id" :column="column" :width="columnWidthClass"
                    :button-highlight="newCardCategory == column.id"
                    :anonymous-button-highlight="newAnonymousCardCategory == column.id" :editable="isOwner"
                    :locked="isLocked" @add-card="add(column.id, false)" @add-anonymous-card="add(column.id, true)"
                    @category-click="openColumnEditDialog">
                    <NewCard v-if="newCardCategory == column.id" :category="column.id" :by="user" :nickname="nickname"
                        :board="board" @added="onAdded" @invalidContent="onInvalidContent" @discard="onDiscard" />
                    <NewAnonymousCard v-if="newAnonymousCardCategory == column.id" :category="column.id" :by="user"
                        nickname="" :board="board" @added="onAdded" @invalidContent="onInvalidContent"
                        @discard="onDiscard" />
                    <Card v-for="card in filterCards(column.id)" :card="card" :comments="filterComments(card.id)"
                        :current-user="user" :current-user-nickname="nickname" :board="board" :mask="isMasked"
                        :can-manage="isOwner" :key="card.id" :categories="columns" :locked="isLocked"
                        @updated="onUpdated" @deleted="onDeleted" @liked="onLiked" @category-changed="onCategoryChanged"
                        @invalidContent="onInvalidContent" @avatar-clicked="showSpotlightFor"
                        @discard="notifyForLostMessages" @comment-added="onCommentAdded"
                        @comment-updated="onCommentUpdated" @comment-deleted="onCommentDeleted"
                        @comment-discard="notifyForLostMessages" @comment-invalid-content="onCommentInvalidContent"
                        :class="{
                            'bg-white dark:bg-gray-400 opacity-10 z-[51] pointer-events-none': isSpotlightOn && usersWithCards.length > 0 && card.nickname !== spotlightFor,
                            'bg-black dark:bg-black border border-gray-200 z-[51]': isSpotlightOn && usersWithCards.length > 0 && card.nickname === spotlightFor,
                        }" />
                </Category>
            </div>
        </div>
        <!-- Dashboard Content -->

        <!-- Right Sidebar -->
        <div class="w-16 p-4">
            <div v-for="user in onlineUsersCardsCount" class="relative w-8 h-8 ml-auto mx-auto mb-4">
                <Avatar :name="user.nickname" class="w-8 h-8" />
                <span v-if="user.cardsCount > 0"
                    class="absolute -top-1 -right-1 bg-red-400 text-white text-xs rounded-full w-4 h-4 flex items-center justify-center select-none">
                    {{ user.cardsCount }}
                </span>
            </div>
        </div>
        <!-- Right Sidebar -->
    </div>
</template>
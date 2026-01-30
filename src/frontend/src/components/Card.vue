<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue';
import Avatar from './Avatar.vue';
import { MessageResponse } from '../models/Requests';
import { DraftMessage } from '../models/DraftMessage';
import { LikeMessage } from '../models/LikeMessage';
import { logMessage } from '../utils';
import { Menu, MenuButton, MenuItems, MenuItem } from '@headlessui/vue'
import { BoardColumn } from '../models/BoardColumn';
import { CategoryChangeMessage } from '../models/CategoryChangeMessage';
import { useI18n } from 'vue-i18n';
import NewComment from './NewComment.vue';
import Comment from './Comment.vue';
import { useContentEditableLimiter } from '../composables/useContentEditableLimiter';
import { useTypingTrigger } from '../composables/useTypingTrigger';

// "currentUserNickname", "card.cat" only used to calculate message size in bytes. "category" already passed with "card.cat".
// Message size calc and trimming only done when editing.
// Only a message's content and category are actually updated in the backend
interface Props {
    card: MessageResponse // Todo: Change name of model. Or use a different model.
    comments?: MessageResponse[]
    mask: boolean
    canManage: boolean
    locked: boolean
    currentUserNickname: string // Only used for message size calc
    categories: BoardColumn[]
}
const props = defineProps<Props>()
const emit = defineEmits(['updated', 'deleted', 'discard', 'invalidContent', 'liked', 'categoryChanged', 'avatarClicked',
    'comment-added', 'comment-updated', 'comment-deleted', 'comment-discard', 'comment-invalidContent', 'typing'
])

const { t } = useI18n()

const editing = ref(false)
const showComments = ref(false)

const manageable = computed(() => props.card.mine || props.canManage)

const content = computed(() => {
    if (props.mask && !props.card.mine && props.card.msg) {
        return props.card.msg.slice(0).replaceAll(/[^ ]/g, '*');
    }
    return props.card.msg
})

const otherCategories = computed(() => {
    return props.categories.filter(c => c.id !== props.card.cat)
})

const edit = async (event: Event) => {
    if (props.locked) {
        logMessage("Locked! Cannot edit.")
        return
    }
    // Todo: figure out what's going on here. Simplify the condition to remove dups.
    if (!editing.value && props.card.mine) {
        editing.value = true;
        await nextTick();
        (event.target as HTMLElement).focus();
    }
    if (!editing.value && props.card.mine && !props.mask) {
        editing.value = true;
        await nextTick();
        (event.target as HTMLElement).focus();
    }
}

const save = (event: Event) => {
    if (props.locked) {
        if (editing.value) {
            // Board lock received when the user is editing a message.
            // The message will be forcefully discarded.
            const target = event.target as HTMLElement
            target.innerText = content.value // Discard new changes by reseting value back to original..TODO: Check what happens when Mask is ON..the *** shouldn't be updated.
            editing.value = false
            emit('discard')
            return
        }
        logMessage("Locked! Cannot save.")
        return
    }
    if (editing.value && props.card.mine) {
        editing.value = false
        if (props.card.msg !== (event.target as HTMLElement).innerText.trim()) {
            const payload: DraftMessage = {
                id: props.card.id,
                msg: (event.target as HTMLElement).innerText.trim(),
                cat: props.card.cat,
                anon: props.card.anon,
                pid: props.card.pid
            }
            emit('updated', payload)
        } else {
            logMessage('No content change. Not dispatching!!')
            // logMessage('[old value]', props.card.msg, '[new value]', (event.target as HTMLElement).innerText.trim())
        }
    }
}

const saveOnEnter = (event: KeyboardEvent) => {
    if (props.locked) {
        logMessage("Locked! Cannot edit.")
        event.preventDefault()
        return
    }
    if (!event.shiftKey) {
        save(event)
        event.preventDefault()
    }
}

const toggleLike = () => {
    if (props.locked) {
        logMessage("Locked! Cannot like/unlike.")
        return
    }
    const payload: LikeMessage = {
        msgId: props.card.id,
        like: !props.card.liked
    }
    emit('liked', payload)
}

const toggleComments = () => {
    showComments.value = !showComments.value
}

const remove = () => {
    if (props.locked) {
        logMessage("Locked! Cannot delete.")
        return
    }
    if (manageable) {
        emit('deleted', props.card.id)
    }
}

const changeCategory = (newCategory: string, oldCategory: string) => {
    if (props.locked) {
        logMessage("Locked! Cannot change category.")
        return
    }
    if (manageable && newCategory !== oldCategory) {
        const payload: CategoryChangeMessage = {
            msgId: props.card.id,
            newCategoryId: newCategory,
            oldCategoryId: oldCategory
        }
        emit('categoryChanged', payload)
    }
}

const { onInput } = useContentEditableLimiter({
    nickname: () => props.currentUserNickname,
    category: () => props.card.cat,
    anon: () => props.card.anon,
    isComment: false,
    onInvalid: (msg) => emit('invalidContent', msg)
})

const { triggerTyping } = useTypingTrigger(emit)

const onKeyDown = (event: KeyboardEvent) => {
    // Trigger the throttled typing event
    triggerTyping(event)
}

const animateLike = ref(false)
watch(
    () => props.card.likes,
    () => {
        animateLike.value = true
        setTimeout(() => {
            animateLike.value = false
        }, 300)
    }
)
</script>

<template>
    <div class="bg-white dark:bg-gray-700 rounded-lg p-3 mb-2 shadow-xl border" :class="editing && card.mine
        ? 'border-sky-400 dark:border-white'
        : 'border-transparent'">

        <div class="text-gray-500 pb-2 dark:text-white" :class="{ 'blur-sm': mask && !card.mine }">
            <article class="min-h-4 text-center break-words focus:outline-none"
                :class="[editing ? 'cursor-auto' : card.mine && !locked ? 'cursor-pointer' : 'cursor-default']"
                :contenteditable="editing && card.mine && !(locked && editing)" @click="edit" @blur="save"
                @keydown.enter="saveOnEnter" @keydown="onKeyDown" @input="onInput">{{
                    content }}</article>
        </div>

        <div class="flex items-center text-gray-500 pt-2">
            <!-- Like button and count display -->
            <!-- <div class="flex mr-1">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
                    stroke="currentColor" class="w-6 h-6 cursor-pointer"
                    :class="{ 'text-blue-500 dark:text-white': card.liked }" @click="toggleLike">
                    <path stroke-linecap="round" stroke-linejoin="round"
                        d="M6.633 10.5c.806 0 1.533-.446 2.031-1.08a9.041 9.041 0 012.861-2.4c.723-.384 1.35-.956 1.653-1.715a4.498 4.498 0 00.322-1.672V3a.75.75 0 01.75-.75A2.25 2.25 0 0116.5 4.5c0 1.152-.26 2.243-.723 3.218-.266.558.107 1.282.725 1.282h3.126c1.026 0 1.945.694 2.054 1.715.045.422.068.85.068 1.285a11.95 11.95 0 01-2.649 7.521c-.388.482-.987.729-1.605.729H13.48c-.483 0-.964-.078-1.423-.23l-3.114-1.04a4.501 4.501 0 00-1.423-.23H5.904M14.25 9h2.25M5.904 18.75c.083.205.173.405.27.602.197.4-.078.898-.523.898h-.908c-.889 0-1.713-.518-1.972-1.368a12 12 0 01-.521-3.507c0-1.553.295-3.036.831-4.398C3.387 10.203 4.167 9.75 5 9.75h1.053c.472 0 .745.556.5.96a8.958 8.958 0 00-1.302 4.665c0 1.194.232 2.333.654 3.375z" />
                </svg>
                <span class="cursor-default dark:text-gray-200" :class="{ 'invisible': card.likes == '0' }">{{
                    card.likes }}</span>
            </div> -->
            <div class="relative inline-flex mr-1 min-w-0 shrink" @click="toggleLike">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
                    stroke="currentColor" class="w-6 h-6 cursor-pointer"
                    :class="{ 'text-blue-500 dark:text-white': card.liked }">
                    <path stroke-linecap="round" stroke-linejoin="round"
                        d="M6.633 10.5c.806 0 1.533-.446 2.031-1.08a9.041 9.041 0 012.861-2.4c.723-.384 1.35-.956 1.653-1.715a4.498 4.498 0 00.322-1.672V3a.75.75 0 01.75-.75A2.25 2.25 0 0116.5 4.5c0 1.152-.26 2.243-.723 3.218-.266.558.107 1.282.725 1.282h3.126c1.026 0 1.945.694 2.054 1.715.045.422.068.85.068 1.285a11.95 11.95 0 01-2.649 7.521c-.388.482-.987.729-1.605.729H13.48c-.483 0-.964-.078-1.423-.23l-3.114-1.04a4.501 4.501 0 00-1.423-.23H5.904M14.25 9h2.25M5.904 18.75c.083.205.173.405.27.602.197.4-.078.898-.523.898h-.908c-.889 0-1.713-.518-1.972-1.368a12 12 0 01-.521-3.507c0-1.553.295-3.036.831-4.398C3.387 10.203 4.167 9.75 5 9.75h1.053c.472 0 .745.556.5.96a8.958 8.958 0 00-1.302 4.665c0 1.194.232 2.333.654 3.375z">
                    </path>
                </svg>
                <span :class="[
                    { 'invisible': card.likes === 0 },
                    animateLike ? 'animate-like-pop' : ''
                ]" class="absolute -top-0.5 -left-1.5 cursor-default
         bg-red-400 text-white text-xs rounded-full
         w-4 h-4 flex items-center justify-center select-none">
                    {{ card.likes }}
                </span>
            </div>

            <!-- Comment button and count display -->
            <div class="relative inline-flex ml-1 min-w-0 shrink" @click="toggleComments">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
                    stroke="currentColor" class="w-6 h-6 cursor-pointer dark:text-gray-200 mr-1">
                    <path stroke-linecap="round" stroke-linejoin="round"
                        d="M12 20.25c4.97 0 9-3.694 9-8.25s-4.03-8.25-9-8.25S3 7.444 3 12c0 2.104.859 4.023 2.273 5.48.432.447.74 1.04.586 1.641a4.483 4.483 0 0 1-.923 1.785A5.969 5.969 0 0 0 6 21c1.282 0 2.47-.402 3.445-1.087.81.22 1.668.337 2.555.337Z" />
                </svg>
                <span :class="{ 'invisible': comments?.length == 0 }"
                    class="absolute -top-0.5 -left-1.5 cursor-default bg-sky-500 text-white text-xs rounded-full w-4 h-4 flex items-center justify-center select-none">
                    {{ (comments?.length ?? 0) > 9 ? '9+' : comments?.length }}
                </span>
            </div>

            <!-- Delete button -->
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
                stroke="currentColor" class="w-6 h-6 cursor-pointer dark:text-gray-200 mr-1"
                :class="{ 'invisible': !manageable }" @click="remove">
                <path stroke-linecap="round" stroke-linejoin="round"
                    d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
            </svg>
            <!-- Change category button -->
            <div class="relative" :class="{ 'invisible': !manageable || otherCategories.length == 0 }">
                <Menu as="div" class="relative z-50 text-left flex justify-center items-center">
                    <MenuButton>
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
                            stroke="currentColor" class="w-6 h-6 cursor-pointer dark:text-gray-200"
                            :class="{ 'invisible': !manageable }">
                            <path stroke-linecap="round" stroke-linejoin="round"
                                d="M15.75 9V5.25A2.25 2.25 0 0 0 13.5 3h-6a2.25 2.25 0 0 0-2.25 2.25v13.5A2.25 2.25 0 0 0 7.5 21h6a2.25 2.25 0 0 0 2.25-2.25V15m3 0 3-3m0 0-3-3m3 3H9" />
                        </svg>
                    </MenuButton>

                    <transition enter-active-class="transition duration-100 ease-out"
                        enter-from-class="transform scale-95 opacity-0" enter-to-class="transform scale-100 opacity-100"
                        leave-active-class="transition duration-75 ease-in"
                        leave-from-class="transform scale-100 opacity-100"
                        leave-to-class="transform scale-95 opacity-0">
                        <MenuItems
                            class="absolute left-6 top-1 min-w-max origin-top-right rounded-md bg-transparent border-0 space-y-0.5 focus:outline-none">
                            <MenuItem v-for="otherCategory in otherCategories" :key="otherCategory.id">
                            <button :class="[
                                `bg-${otherCategory.color}-400 hover:bg-${otherCategory.color}-500`,
                                `text-white`,
                                `dark:bg-${otherCategory.color}-800 dark:hover:bg-${otherCategory.color}-600`,
                                `dark:text-${otherCategory.color}-100`,
                                'group flex w-full items-center justify-center rounded-md text-xs px-1',
                            ]" @click="changeCategory(otherCategory.id, props.card.cat)">
                                {{ otherCategory.isDefault ? t(`dashboard.columns.${otherCategory.id}`) :
                                    otherCategory.text }}
                            </button>
                            </MenuItem>
                        </MenuItems>
                    </transition>
                </Menu>
            </div>

            <Avatar v-if="!card.anon" :name="card.nickname" class="ml-auto w-6 h-6 cursor-pointer"
                @click="emit('avatarClicked', { byxid: card.byxid, nickname: card.nickname })" />
            <div v-else
                class="inline-flex items-center justify-center overflow-hidden rounded-full bg-gray-300 dark:bg-white/30 text-gray-600 dark:text-white ml-auto w-6 h-6 cursor-pointer"
                :title="t('common.anonymous')"
                @click="emit('avatarClicked', { byxid: card.byxid, nickname: card.nickname })">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                    stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-5 h-5">
                    <path d="M14 18a2 2 0 0 0-4 0" />
                    <path
                        d="m19 11-2.11-6.657a2 2 0 0 0-2.752-1.148l-1.276.61A2 2 0 0 1 12 4H8.5a2 2 0 0 0-1.925 1.456L5 11" />
                    <path d="M2 11h20" />
                    <circle cx="17" cy="18" r="3" />
                    <circle cx="7" cy="18" r="3" />
                </svg>
            </div>

        </div>

        <!-- Comments panel -->
        <div v-if="showComments" class="border-t border-gray-200 mt-2 pt-2 space-y-2">

            <NewComment :parent-id="props.card.id" :category="props.card.cat" :locked="locked"
                :nickname="props.currentUserNickname" @added="emit('comment-added', $event)"
                @invalid-content="emit('comment-invalidContent', $event)"></NewComment>

            <template v-if="comments?.length">
                <Comment v-for="comment in comments" :key="comment.id" :comment="comment" :mask="mask"
                    :current-user-nickname="currentUserNickname" :can-manage="props.canManage" :locked="locked"
                    @updated="emit('comment-updated', $event)" @deleted="emit('comment-deleted', $event)"
                    @discard="emit('comment-discard', $event)"
                    @invalid-content="emit('comment-invalidContent', $event)" />
            </template>

        </div>
        <!-- Comments panel -->

    </div>
</template>

<style lang="css" scoped>
@keyframes like-pop {
    0% {
        transform: scale(1);
    }

    30% {
        transform: scale(1.4);
    }

    60% {
        transform: scale(0.9);
    }

    100% {
        transform: scale(1);
    }
}

.animate-like-pop {
    animation: like-pop 0.3s ease-out;
}
</style>
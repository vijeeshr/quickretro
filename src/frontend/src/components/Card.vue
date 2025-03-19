<script setup lang="ts">
import { computed, nextTick, ref } from 'vue';
import Avatar from './Avatar.vue';
import { MessageResponse } from '../models/Requests';
import { DraftMessage } from '../models/DraftMessage';
import { LikeMessage } from '../models/LikeMessage';
import { assertMessageContentValidation, canAssertMessageContentValidation, logMessage, MessageContentValidationResult } from '../utils';
import { Menu, MenuButton, MenuItems, MenuItem } from '@headlessui/vue'
import { BoardColumn } from '../models/BoardColumn';
import { CategoryChangeMessage } from '../models/CategoryChangeMessage';
import { useI18n } from 'vue-i18n';

// "currentUser", "currentUserNickname", "board", "card.cat" only used to calculate message size in bytes. "category" already passed with "card.cat".
// Message size calc and trimming only done when editing.
// Apart from message owner, only board creator/owner can update someone else's message.
// Only a message's content and category are actually updated in the backend
interface Props {
    card: MessageResponse // Todo: Change name of model. Or use a different model.
    mask: boolean
    updateable: boolean
    locked: boolean
    currentUser: string // Only used for message size calc
    currentUserNickname: string // Only used for message size calc
    board: string // Only used for message size calc
    categories: BoardColumn[]
}
const props = defineProps<Props>()
const emit = defineEmits(['updated', 'deleted', 'liked', 'invalidContent', 'categoryChanged', 'avatarClicked'])

const { t } = useI18n()

const editing = ref(false)

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
    if (!editing.value && props.updateable && !props.mask) {
        editing.value = true;
        await nextTick();
        (event.target as HTMLElement).focus();
    }
}

const save = (event: Event) => {
    if (props.locked) {
        logMessage("Locked! Cannot save.")
        return
    }
    if (editing.value && props.updateable) {
        editing.value = false
        if (props.card.msg !== (event.target as HTMLElement).innerText.trim()) {
            const payload: DraftMessage = {
                id: props.card.id,
                msg: (event.target as HTMLElement).innerText.trim(),
                cat: props.card.cat,
                anon: props.card.anon
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

const remove = () => {
    if (props.locked) {
        logMessage("Locked! Cannot delete.")
        return
    }
    if (props.updateable) {
        emit('deleted', props.card.id)
    }
}

const changeCategory = (newCategory: string, oldCategory: string) => {
    if (props.locked) {
        logMessage("Locked! Cannot change category.")
        return
    }
    if (props.updateable && newCategory !== oldCategory) {
        const payload: CategoryChangeMessage = {
            msgId: props.card.id,
            newCategoryId: newCategory,
            oldCategoryId: oldCategory
        }
        emit('categoryChanged', payload)
    }
}

const validate = (event: Event) => {
    if (!editing.value && !props.updateable) return
    if (!canAssertMessageContentValidation()) return
    const validationResult: MessageContentValidationResult = assertMessageContentValidation(event, props.currentUser, props.currentUserNickname, props.board, props.card.cat)
    if (validationResult.isValid) return

    let errorMessage: string = t('common.contentOverloadError')
    if (validationResult.isTrimmed) errorMessage = t('common.contentStrippingError')

    emit('invalidContent', errorMessage)
}
</script>

<template>
    <div class="bg-white dark:bg-gray-700 rounded-lg p-3 mb-2 shadow-xl"
        :class="{ 'border': editing && updateable, 'border-sky-400 dark:border-white': editing && updateable }">

        <div class="text-gray-500 pb-2 dark:text-white" :class="{ 'blur-sm': mask && !card.mine }">
            <article class="min-h-4 text-center break-words focus:outline-none"
                :class="[editing && updateable ? 'cursor-auto' : 'cursor-default']"
                :contenteditable="editing && updateable" @click="edit" @blur="save" @keydown.enter="saveOnEnter"
                @input="validate">{{
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
            <div class="relative flex mr-1" @click="toggleLike">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
                    stroke="currentColor" class="w-6 h-6 cursor-pointer"
                    :class="{ 'text-blue-500 dark:text-white': card.liked }">
                    <path stroke-linecap="round" stroke-linejoin="round"
                        d="M6.633 10.5c.806 0 1.533-.446 2.031-1.08a9.041 9.041 0 012.861-2.4c.723-.384 1.35-.956 1.653-1.715a4.498 4.498 0 00.322-1.672V3a.75.75 0 01.75-.75A2.25 2.25 0 0116.5 4.5c0 1.152-.26 2.243-.723 3.218-.266.558.107 1.282.725 1.282h3.126c1.026 0 1.945.694 2.054 1.715.045.422.068.85.068 1.285a11.95 11.95 0 01-2.649 7.521c-.388.482-.987.729-1.605.729H13.48c-.483 0-.964-.078-1.423-.23l-3.114-1.04a4.501 4.501 0 00-1.423-.23H5.904M14.25 9h2.25M5.904 18.75c.083.205.173.405.27.602.197.4-.078.898-.523.898h-.908c-.889 0-1.713-.518-1.972-1.368a12 12 0 01-.521-3.507c0-1.553.295-3.036.831-4.398C3.387 10.203 4.167 9.75 5 9.75h1.053c.472 0 .745.556.5.96a8.958 8.958 0 00-1.302 4.665c0 1.194.232 2.333.654 3.375z">
                    </path>
                </svg>
                <span :class="{ 'invisible': card.likes == '0' }"
                    class="absolute -top-0.5 -left-1.5 cursor-default bg-red-400 text-white text-xs rounded-full w-4 h-4 flex items-center justify-center select-none">
                    {{ card.likes }}
                </span>
            </div>

            <!-- Delete button -->
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
                stroke="currentColor" class="w-6 h-6 cursor-pointer dark:text-gray-200 mr-1"
                :class="{ 'invisible': !updateable }" @click="remove">
                <path stroke-linecap="round" stroke-linejoin="round"
                    d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
            </svg>
            <!-- Change category button -->
            <div class="relative" :class="{ 'invisible': !updateable || otherCategories.length == 0 }">
                <Menu as="div" class="relative z-50 text-left flex justify-center items-center">
                    <MenuButton>
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
                            stroke="currentColor" class="w-6 h-6 cursor-pointer dark:text-gray-200"
                            :class="{ 'invisible': !updateable }">
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
                @click="emit('avatarClicked', card.nickname)" />
            <div v-else
                class="inline-flex items-center justify-center overflow-hidden rounded-full bg-gray-300 dark:bg-white/30 text-gray-600 dark:text-white ml-auto w-6 h-6 cursor-pointer"
                :title="t('common.anonymous')" @click="emit('avatarClicked', card.nickname)">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                    stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-5 h-5">
                    <path stroke="none" d="M0 0h24v24H0z" fill="none" />
                    <path d="M3 11h18" />
                    <path d="M5 11v-4a3 3 0 0 1 3 -3h8a3 3 0 0 1 3 3v4" />
                    <path d="M7 17m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0" />
                    <path d="M17 17m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0" />
                    <path d="M10 17h4" />
                </svg>
            </div>

        </div>

    </div>
</template>
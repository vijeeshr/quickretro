<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { CategoryDefinition } from "../models/CategoryDefinition"
import { useToast } from 'vue-toast-notification'
import { encodeToUrlSafeBase64FromJson } from '../utils';

const props = defineProps<{ categories: CategoryDefinition[] }>()
const { t } = useI18n()
const toast = useToast()
const isOpen = ref(false)

const shareUrl = computed(() => {
    const minimalCategories = props.categories.map(c => ({
        id: c.id,
        text: c.text,
        enabled: c.enabled,
        pos: c.pos,
        color: c.color,
    }))
    const encoded = encodeToUrlSafeBase64FromJson(minimalCategories)
    const origin = window.location.origin

    return `${origin}/create?preset=v1:${encoded}`  // Point to /create
})

const toggleShare = () => {
    isOpen.value = !isOpen.value
}

const copyLink = async () => {
    try {
        await navigator.clipboard.writeText(shareUrl.value)
        toast.success(t('common.share.linkCopied'))
        isOpen.value = false
    } catch (e) {
        toast.error(t('common.share.linkCopyError'))
    }
}
</script>

<template>
    <div class="relative inline-block text-left">
        <button type="button" @click="toggleShare"
            class="text-xs flex items-center gap-1 text-sky-600 hover:text-sky-800 dark:text-sky-400 dark:hover:text-sky-300 transition-colors">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
                stroke="currentColor" class="size-4">
                <path stroke-linecap="round" stroke-linejoin="round"
                    d="M7.217 10.907a2.25 2.25 0 1 0 0 2.186m0-2.186c.18.324.287.696.345 1.084m-.345-1.084c-.18-.324-.287-.696-.345-1.084m0 2.186 9.566-5.314m-9.566 7.5 9.566 5.314m0 0a2.25 2.25 0 1 0 3.935 2.186 2.25 2.25 0 0 0-3.935-2.186Zm0-12.814a2.25 2.25 0 1 0 3.933-2.185 2.25 2.25 0 0 0-3.933 2.185Z" />
            </svg>
            {{ t('common.customColumnSetup.shareLabel') }}
        </button>

        <div v-if="isOpen"
            class="absolute right-0 z-10 mt-2 w-72 origin-top-right rounded-md bg-white dark:bg-gray-800 shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none p-4">
            <div class="flex flex-col gap-2">
                <p class="text-xs text-gray-500 dark:text-gray-400">
                    {{ t('common.customColumnSetup.shareHelp') }}
                </p>
                <div class="flex items-center gap-2">
                    <input type="text" :value="shareUrl" readonly
                        class="flex-1 rounded-md border text-xs border-gray-300 dark:border-gray-600 px-2 py-1 bg-gray-50 dark:bg-gray-700 dark:text-gray-200 focus:outline-none" />
                    <button @click="copyLink"
                        class="rounded-md bg-sky-100 dark:bg-sky-900 px-2 py-1 text-xs font-medium text-sky-700 dark:text-sky-200 hover:bg-sky-200 dark:hover:bg-sky-800">
                        {{ t('common.copy', 'Copy') }}
                    </button>
                </div>
            </div>
            <!-- Close button -->
            <button @click="isOpen = false"
                class="absolute top-1 right-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
                    stroke="currentColor" class="size-4">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
                </svg>
            </button>
        </div>
    </div>
</template>

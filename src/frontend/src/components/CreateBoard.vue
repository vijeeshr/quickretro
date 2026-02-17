<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { CreateBoardRequest, createBoard } from '../api'
import DarkModeToggle from './DarkModeToggle.vue'
import { BoardColumn } from '../models/BoardColumn'
import { CategoryDefinition } from '../models/CategoryDefinition'
import { useI18n } from 'vue-i18n'
import LanguageSelector from './LanguageSelector.vue'
import TurnstileWidget from './TurnstileWidget.vue'
import { useToast } from 'vue-toast-notification'
import CategoryEditor from './CategoryEditor.vue'
import { defaultCategories } from '../constants/defaultCategories'
import { MAX_TEXT_LENGTH, TURNSTILE_ENABLED, TURNSTILE_SITEKEY } from '../utils/appConfig'
import CategoryPresetShare from './CategoryPresetShare.vue'
import { decodeToJsonFromUrlSafeBase64 } from '../utils'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const boardname = ref('')
const team = ref('')
const isDark = ref(localStorage.getItem('theme') === 'dark')
const isTurnstileEnabled = ref(TURNSTILE_ENABLED)
const turnstileSiteKey = ref(TURNSTILE_SITEKEY)
const turnstileToken = ref('')
const isTurnstileVerified = ref(false)
const isSubmitting = ref(false)
const turnstileRef = ref<{ reset: () => void }>()
const categories = ref<CategoryDefinition[]>([...defaultCategories])
const isCategorySelectionValid = ref(true)

const handleCategoryTextUpdate = (update: { id: string; text: string }) => {
  const cat = categories.value.find(c => c.id === update.id)
  if (cat) {
    cat.text = update.text
  }
}
const handleCategoryToggle = (update: { id: string; enabled: boolean }) => {
  const cat = categories.value.find(c => c.id === update.id)
  if (cat) {
    cat.enabled = update.enabled
  }
}
const handleCategoriesReorder = (reorderedCategories: CategoryDefinition[]) => {
  categories.value = reorderedCategories
}

const handleCategorySelectionValidity = (val: boolean) => {
  isCategorySelectionValid.value = val
}

const boardnameEntered = computed(() => {
  if (boardname.value && boardname.value.trim() !== '') return true
  return false
})

const handleTokenError = () => {
  console.log('Turnstile error occurred')
}
const handleTokenExpired = () => {
  console.log('Turnstile token expired')
  turnstileRef.value?.reset()
}
const handleTokenVerified = (token: string) => {
  isTurnstileVerified.value = true
  turnstileToken.value = token
}

const toast = useToast()

const create = async () => {
  // Todo: Throttle this.
  if (isTurnstileEnabled.value && !isTurnstileVerified.value) return

  const selectedColumns: BoardColumn[] = categories.value
    .filter(c => c.enabled === true)
    .map(c => ({
      id: c.id,
      text: c.text.trim() || t(`dashboard.columns.${c.id}`),
      isDefault: c.text === '' || c.text === t(`dashboard.columns.${c.id}`),
      color: c.color,
      pos: c.pos,
    }))

  const payload: CreateBoardRequest = {
    name: boardname.value,
    team: team.value,
    owner: localStorage.getItem('user') || '',
    columns: selectedColumns,
    cfTurnstileResponse: turnstileToken.value,
  }

  isSubmitting.value = true
  try {
    const createdBoard = await createBoard(payload)
    router.push(`/board/${createdBoard.id}`)
  } catch (error) {
    toast.error(t('createBoard.boardCreationError', { pauseOnHover: false }))
    console.error('Error creating board:', error)
  } finally {
    isSubmitting.value = false
  }
}

const applyPresetFromRoute = () => {
  const preset = route.query.preset
  if (typeof preset !== 'string') return
  if (!preset.startsWith('v1:')) return

  try {
    const encoded = preset.slice(3)
    const decoded = decodeToJsonFromUrlSafeBase64(encoded)

    if (!Array.isArray(decoded) || decoded.length === 0) return

    // Map defaults by id for safe merging
    const defaultsById = new Map(defaultCategories.map(c => [c.id, { ...c }]))

    const merged: CategoryDefinition[] = decoded
      .map((c: any) => {
        const base = defaultsById.get(c.id)
        if (!base) return null // ignore unknown categories

        return {
          ...base,
          text: String(c.text ?? '').slice(0, MAX_TEXT_LENGTH),
          enabled: !!c.enabled,
          pos: Number.isFinite(Number(c.pos)) ? Number(c.pos) : base.pos,
          color: c.color ?? base.color,
        }
      })
      .filter(Boolean) as CategoryDefinition[] // filter out non-truthy values null | undefined that may come as a result of "if (!base) return null" above

    if (merged.length === 0) return

    // Normalize ordering
    merged.sort((a, b) => a.pos - b.pos).forEach((c, i) => (c.pos = i + 1))

    categories.value = merged
    toast.info(t('common.customColumnSetup.applied'), { duration: 2000 })
  } catch (err) {
    console.warn('Invalid category preset ignored', err)
  }
}

onMounted(() => {
  document.documentElement.classList.toggle('dark', isDark.value)
  applyPresetFromRoute()
})
</script>

<template>
  <div class="bg-gray-100 dark:bg-gray-950 flex h-screen items-center justify-center p-4">
    <div class="w-full max-w-md">
      <div class="bg-white dark:bg-gray-900 shadow-md rounded-md p-5 md:p-8">
        <h2
          class="text-center text-3xl font-bold tracking-tight text-gray-600 dark:text-gray-400 select-none"
        >
          {{ t('createBoard.label') }}
        </h2>
        <div class="space-y-2 md:space-y-4 mt-4">
          <div>
            <div class="mt-1">
              <input
                v-model.trim="boardname"
                name="name"
                type="text"
                :maxlength="MAX_TEXT_LENGTH"
                :placeholder="t('createBoard.namePlaceholder')"
                required
                autofocus
                class="px-2 py-2 mt-1 block w-full rounded-md border border-gray-300 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-sky-500 sm:text-sm dark:bg-slate-800 dark:text-slate-200"
              />
            </div>
            <p
              v-show="!boardnameEntered"
              name="name-required"
              class="text-sm text-red-600 dark:text-red-300 mt-2 select-none"
            >
              {{ t('createBoard.nameRequired') }}
            </p>
          </div>
          <div>
            <div class="mt-1">
              <input
                v-model.trim="team"
                name="team"
                type="text"
                :maxlength="MAX_TEXT_LENGTH"
                :placeholder="t('createBoard.teamNamePlaceholder')"
                class="px-2 py-2 mt-1 block w-full rounded-md border border-gray-300 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-sky-500 sm:text-sm dark:bg-slate-800 dark:text-slate-200"
              />
            </div>
          </div>
          <div>
            <!-- <ul class="space-y-2 text-sm">
                            <li v-for="(column, index) in columns" :key="column.id" class="flex space-x-1"
                                draggable="true" @dragstart="onDragStart(index)" @dragover="onDragOver"
                                @drop="onDrop(index)">
                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2"
                                    stroke="currentColor" class="w-6 h-6 text-gray-900 dark:text-gray-400 cursor-move">
                                    <path stroke-linecap="round" stroke-linejoin="round"
                                        d="M8 9h.01M16 9h.01M8 15h.01M16 15h.01" />
                                </svg>
                                <button v-if="column.enabled" @click="toggleColumn(column.id, false)">
                                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor"
                                        class="w-6 h-6" :class="[`${column.colorClass}`]">
                                        <path fill-rule="evenodd"
                                            d="M2.25 12c0-5.385 4.365-9.75 9.75-9.75s9.75 4.365 9.75 9.75-4.365 9.75-9.75 9.75S2.25 17.385 2.25 12Zm13.36-1.814a.75.75 0 1 0-1.22-.872l-3.236 4.53L9.53 12.22a.75.75 0 0 0-1.06 1.06l2.25 2.25a.75.75 0 0 0 1.14-.094l3.75-5.25Z"
                                            clip-rule="evenodd" />
                                    </svg>
                                </button>
                                <button v-if="!column.enabled" @click="toggleColumn(column.id, true)">
                                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor"
                                        class="w-6 h-6 text-gray-500">
                                        <path fill-rule="evenodd"
                                            d="M12 2.25c-5.385 0-9.75 4.365-9.75 9.75s4.365 9.75 9.75 9.75 9.75-4.365 9.75-9.75S17.385 2.25 12 2.25Zm-1.72 6.97a.75.75 0 1 0-1.06 1.06L10.94 12l-1.72 1.72a.75.75 0 1 0 1.06 1.06L12 13.06l1.72 1.72a.75.75 0 1 0 1.06-1.06L13.06 12l1.72-1.72a.75.75 0 1 0-1.06-1.06L12 10.94l-1.72-1.72Z"
                                            clip-rule="evenodd" />
                                    </svg>
                                </button>
                                <input type="text" v-model.trim="column.text"
                                    :placeholder="t(`dashboard.columns.${column.id}`)"
                                    class="w-full rounded-md focus:outline-none focus:border focus:border-gray-200 focus:ring-gray-200 dark:text-slate-200 dark:bg-gray-900 dark:focus:border-gray-800 dark:focus:ring-gray-800" />
                            </li>
                        </ul> -->
            <div class="flex justify-between items-end mb-2">
              <span class="text-sm font-medium text-gray-600 dark:text-gray-200">{{
                t('createBoard.columns')
              }}</span>
              <CategoryPresetShare :categories="categories" />
            </div>
            <CategoryEditor
              :categories="categories"
              @category-text-update="handleCategoryTextUpdate"
              @category-toggle="handleCategoryToggle"
              @categories-reorder="handleCategoriesReorder"
              @valid="handleCategorySelectionValidity"
            >
            </CategoryEditor>
          </div>
          <div class="flex w-full gap-2">
            <button
              type="submit"
              class="flex justify-center px-4 py-2 text-sm w-[90%] shadow-md bg-sky-100 hover:bg-sky-400 border-sky-300 text-sky-600 hover:text-white disabled:bg-gray-300 disabled:text-gray-500 disabled:border-gray-400 disabled:cursor-not-allowed dark:disabled:bg-gray-300 dark:disabled:text-gray-500 dark:disabled:border-gray-400 dark:bg-sky-800 dark:hover:bg-sky-600 dark:border-sky-700 dark:text-sky-100 hover:border-transparent font-medium rounded-md border focus:outline-none focus:ring-2 focus:ring-sky-600 focus:ring-offset-2 dark:focus:ring-2 dark:focus:ring-offset-0 select-none"
              :disabled="
                !boardnameEntered ||
                !isCategorySelectionValid ||
                (isTurnstileEnabled && !isTurnstileVerified)
              "
              @click="create"
            >
              {{ isSubmitting ? t('createBoard.buttonProgress') : t('createBoard.button') }}
            </button>
            <div
              class="w-[10%] flex items-center justify-center shadow-md border rounded-md border-sky-200"
            >
              <DarkModeToggle class="w-6 h-6 cursor-pointer text-sky-200 hover:text-sky-400" />
            </div>
          </div>
          <div class="w-full">
            <LanguageSelector />
          </div>
          <div v-if="isTurnstileEnabled" class="min-w-[300px] flex items-center justify-center">
            <TurnstileWidget
              ref="turnstileRef"
              v-if="isTurnstileEnabled"
              class="w-full"
              :sitekey="turnstileSiteKey"
              :dark-theme="isDark"
              @error="handleTokenError"
              @expired="handleTokenExpired"
              @verified="handleTokenVerified"
            />
          </div>
          <div
            v-show="isTurnstileEnabled && !isTurnstileVerified"
            class="text-sm text-red-600 dark:text-red-300 select-none w-full flex items-center justify-center"
          >
            {{ t('createBoard.captchaInfo') }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

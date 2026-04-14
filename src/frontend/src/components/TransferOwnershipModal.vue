<script setup lang="ts">
import { ref, watch } from 'vue'
import {
  Dialog,
  DialogPanel,
  DialogTitle,
  Listbox,
  ListboxButton,
  ListboxOptions,
  ListboxOption,
} from '@headlessui/vue'
import Avatar from './Avatar.vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

interface Props {
  isOpen: boolean
  preselectedXid?: string | null
  users: Array<{ xid: string; nickname: string }>
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'close'): void
  (e: 'transfer', xid: string): void
}>()

const selectedXid = ref<string>('')

watch(
  () => props.isOpen,
  newVal => {
    if (newVal) {
      if (props.preselectedXid) {
        selectedXid.value = props.preselectedXid
      } else if (props.users.length > 0) {
        selectedXid.value = props.users[0].xid
      }
    }
  }
)

const submit = () => {
  if (selectedXid.value) {
    emit('transfer', selectedXid.value)
  }
  emit('close')
}

const getNickname = (xid: string) => {
  const u = props.users.find(u => u.xid === xid)
  return u ? u.nickname : ''
}
</script>

<template>
  <Dialog :open="isOpen" class="relative z-60" @close="$emit('close')">
    <div class="fixed inset-0 bg-black/30 dark:bg-black/60" aria-hidden="true" />
    <div class="fixed inset-0 flex items-center justify-center p-4">
      <DialogPanel
        class="w-full max-w-sm rounded-xl bg-white dark:bg-slate-800 p-6 shadow-xl space-y-6"
      >
        <DialogTitle class="text-xl font-medium text-slate-800 dark:text-slate-100">
          {{ t('transferOwnership.title') }}
        </DialogTitle>

        <div class="space-y-4">
          <label class="block text-sm font-medium text-slate-700 dark:text-slate-300">
            {{ t('transferOwnership.selectLabel') }}
          </label>
          <Listbox v-model="selectedXid">
            <div class="relative mt-1">
              <ListboxButton
                class="relative w-full cursor-default rounded-lg bg-white dark:bg-slate-700 py-2 pl-3 pr-10 text-left border border-slate-300 dark:border-slate-600 focus:outline-none sm:text-sm"
              >
                <span class="block truncate dark:text-slate-200">
                  <div v-if="selectedXid" class="flex items-center gap-2">
                    <Avatar :name="getNickname(selectedXid)" view-type="Badge" />
                  </div>
                  <span v-else>{{ t('transferOwnership.selectPlaceholder') }}</span>
                </span>
                <span
                  class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2 text-slate-400"
                >
                  <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                    <path
                      fill-rule="evenodd"
                      d="M10 3a.75.75 0 01.55.24l3.25 3.5a.75.75 0 11-1.1 1.02L10 4.852 7.3 7.76a.75.75 0 01-1.1-1.02l3.25-3.5A.75.75 0 0110 3zm-3.76 9.2a.75.75 0 011.06.04l2.7 2.908 2.7-2.908a.75.75 0 111.1 1.02l-3.25 3.5a.75.75 0 01-1.1 0l-3.25-3.5a.75.75 0 01.04-1.06z"
                      clip-rule="evenodd"
                    />
                  </svg>
                </span>
              </ListboxButton>

              <ListboxOptions
                class="absolute mt-1 max-h-40 w-full overflow-auto rounded-md bg-white dark:bg-slate-700 py-1 text-base shadow-lg ring-1 ring-black/5 focus:outline-none sm:text-sm z-10"
              >
                <ListboxOption
                  v-for="person in users"
                  :key="person.xid"
                  v-slot="{ active, selected }"
                  :value="person.xid"
                >
                  <li
                    :class="[
                      active
                        ? 'bg-amber-100 dark:bg-slate-600 text-white'
                        : 'text-gray-900 dark:text-slate-200',
                      'relative cursor-default select-none py-2 pl-4 pr-4',
                    ]"
                  >
                    <div
                      class="flex items-center gap-2"
                      :class="[selected ? 'font-medium' : 'font-normal']"
                    >
                      <Avatar
                        :name="person.nickname"
                        view-type="Badge"
                        :truncate-badge-text="false"
                      />
                    </div>
                  </li>
                </ListboxOption>
              </ListboxOptions>
            </div>
          </Listbox>
        </div>

        <div class="flex justify-end gap-3 mt-6">
          <button
            type="button"
            class="rounded-md border border-slate-300 dark:border-slate-600 px-4 py-2 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700"
            @click="$emit('close')"
          >
            {{ t('transferOwnership.cancel') }}
          </button>
          <button
            :disabled="!selectedXid"
            class="inline-flex justify-center rounded-md border border-transparent bg-sky-600 px-4 py-2 text-sm font-medium text-white hover:bg-sky-700 disabled:opacity-50"
            @click="submit"
          >
            {{ t('transferOwnership.confirm') }}
          </button>
        </div>
      </DialogPanel>
    </div>
  </Dialog>
</template>

<script setup lang="ts">
import { Listbox, ListboxButton, ListboxOptions, ListboxOption } from '@headlessui/vue'
import { useLanguage } from '../composables/useLanguage'

const { locale, languageOptions, getLocaleMessage } = useLanguage()
</script>

<template>
  <Listbox v-model="locale">
    <div class="relative w-full text-sm">
      <ListboxButton
        class="w-full flex items-center gap-2 px-4 py-2 text-left bg-white border border-gray-300 rounded-md shadow-sm dark:bg-slate-800 dark:text-slate-200"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          stroke-width="2"
          stroke="currentColor"
          class="size-6 text-slate-700"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="m10.5 21 5.25-11.25L21 21m-9-3h7.5M3 5.621a48.474 48.474 0 0 1 6-.371m0 0c1.12 0 2.233.038 3.334.114M9 5.25V3m3.334 2.364C11.176 10.658 7.69 15.08 3 17.502m9.334-12.138c.896.061 1.785.147 2.666.257m-4.589 8.495a18.023 18.023 0 0 1-3.827-5.802"
          />
        </svg>
        <span class="select-none">{{ getLocaleMessage(locale).langName }}</span>
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="24"
          height="24"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="ml-auto text-slate-400"
        >
          <path d="m6 9 6 6 6-6" />
        </svg>
      </ListboxButton>

      <ListboxOptions
        class="absolute z-10 mt-1 w-full bg-white border border-gray-300 rounded-md shadow-lg max-h-40 overflow-auto dark:bg-slate-800 dark:text-slate-200"
      >
        <ListboxOption
          v-for="lang in languageOptions"
          :key="lang.code"
          :value="lang.code"
          v-slot="{ active, selected }"
        >
          <li
            :class="[
              active ? 'bg-blue-100 dark:bg-slate-900' : '',
              selected ? 'font-bold' : '',
              'cursor-pointer select-none px-4 py-2',
            ]"
          >
            {{ lang.name }}
          </li>
        </ListboxOption>
      </ListboxOptions>
    </div>
  </Listbox>
</template>

<script setup lang="ts">
import { Listbox, ListboxButton, ListboxOptions, ListboxOption } from '@headlessui/vue';
import { useLanguage } from '../composables/useLanguage';

const { locale, languageOptions, getLocaleMessage } = useLanguage()
</script>

<template>
    <Listbox v-model="locale">
        <div class="relative w-full text-sm">
            <ListboxButton
                class="w-full flex items-center gap-2 px-4 py-2 text-left bg-white border border-gray-300 rounded-md shadow-sm dark:bg-slate-800 dark:text-slate-200">
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="text-slate-700">
                    <circle cx="12" cy="12" r="10" />
                    <path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20" />
                    <path d="M2 12h20" />
                </svg>
                <span class="select-none">{{ getLocaleMessage(locale).langName }}</span>
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="ml-auto text-slate-400">
                    <path d="m6 9 6 6 6-6" />
                </svg>
            </ListboxButton>

            <ListboxOptions
                class="absolute z-10 mt-1 w-full bg-white border border-gray-300 rounded-md shadow-lg max-h-40 overflow-auto dark:bg-slate-800 dark:text-slate-200">
                <ListboxOption v-for="lang in languageOptions" :key="lang.code" :value="lang.code"
                    v-slot="{ active, selected }">
                    <li
                        :class="[active ? 'bg-blue-100 dark:bg-slate-900' : '', selected ? 'font-bold' : '', 'cursor-pointer select-none px-4 py-2']">
                        {{ lang.name }}
                    </li>
                </ListboxOption>
            </ListboxOptions>
        </div>
    </Listbox>
</template>
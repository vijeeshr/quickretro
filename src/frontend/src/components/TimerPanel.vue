<script setup lang="ts">
import { computed, ref } from 'vue'

interface Props {
    isCountdownInProgress: boolean
}
defineProps<Props>()
const emit = defineEmits(['start', 'stop'])

const minutes = ref(10)
const seconds = ref(0)

const isTimeValid = computed(() => {
    if (typeof minutes.value != 'number' || typeof seconds.value != 'number') return false

    let totalSeconds = (minutes.value * 60) + seconds.value
    if (totalSeconds <= 0 || totalSeconds > 3600) return false

    return true
})

const start = () => {
    if (isTimeValid.value) {
        emit('start', (minutes.value * 60) + seconds.value)
    }
}

const stop = () => {
    emit('stop')
}

const incrementMinutes = () => {
    if (minutes.value >= 60) minutes.value = 60
    else minutes.value += 1
}
const decrementMinutes = () => {
    if (minutes.value <= 0) minutes.value = 0
    else minutes.value -= 1
}

const incrementSeconds = () => {
    if (seconds.value >= 60) seconds.value = 60
    else seconds.value += 1
}

const decrementSeconds = () => {
    if (seconds.value <= 0) seconds.value = 0
    else seconds.value -= 1
}

</script>

<template>
    <div class="flex justify-evenly items-center w-full">
        <!-- https://www.material-tailwind.com/docs/html/input-number#input-amount-buttons -->
        <!-- Mins Input Start -->
        <div class="w-[10rem] max-w-sm relative mt-4 pr-1">
            <label class="block mb-1 text-sm text-slate-600 dark:text-slate-300">Minutes</label>
            <div class="relative">
                <button
                    class="absolute right-9 top-1 rounded bg-slate-800 p-1.5 border border-transparent text-center text-sm text-white transition-all shadow-sm hover:shadow focus:bg-slate-700 focus:shadow-none active:bg-slate-700 hover:bg-slate-700 active:shadow-none disabled:pointer-events-none disabled:opacity-50 disabled:shadow-none"
                    type="button" @click="decrementMinutes">
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" class="w-4 h-4">
                        <path d="M3.75 7.25a.75.75 0 0 0 0 1.5h8.5a.75.75 0 0 0 0-1.5h-8.5Z" />
                    </svg>
                </button>
                <input type="number" value="0" placeholder="Minutes" min="0" max="60" v-model.trim.number="minutes"
                    class="w-full bg-transparent placeholder:text-slate-400 text-slate-700 dark:text-slate-200 text-sm border border-slate-200 rounded-md pl-3 pr-20 py-2 transition duration-300 ease focus:outline-none focus:border-slate-400 hover:border-slate-300 shadow-sm focus:shadow appearance-none [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none" />
                <button
                    class="absolute right-1 top-1 rounded bg-slate-800 p-1.5 border border-transparent text-center text-sm text-white transition-all shadow-sm hover:shadow focus:bg-slate-700 focus:shadow-none active:bg-slate-700 hover:bg-slate-700 active:shadow-none disabled:pointer-events-none disabled:opacity-50 disabled:shadow-none"
                    type="button" @click="incrementMinutes">
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" class="w-4 h-4">
                        <path
                            d="M8.75 3.75a.75.75 0 0 0-1.5 0v3.5h-3.5a.75.75 0 0 0 0 1.5h3.5v3.5a.75.75 0 0 0 1.5 0v-3.5h3.5a.75.75 0 0 0 0-1.5h-3.5v-3.5Z" />
                    </svg>
                </button>
            </div>
        </div>
        <!-- Mins Input End -->

        <!-- Seconds Input Start -->
        <div class="w-[10rem] max-w-sm relative mt-4 pl-1">
            <label class="block mb-1 text-sm text-slate-600 dark:text-slate-300">Seconds</label>
            <div class="relative">
                <button
                    class="absolute right-9 top-1 rounded bg-slate-800 p-1.5 border border-transparent text-center text-sm text-white transition-all shadow-sm hover:shadow focus:bg-slate-700 focus:shadow-none active:bg-slate-700 hover:bg-slate-700 active:shadow-none disabled:pointer-events-none disabled:opacity-50 disabled:shadow-none"
                    type="button" @click="decrementSeconds">
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" class="w-4 h-4">
                        <path d="M3.75 7.25a.75.75 0 0 0 0 1.5h8.5a.75.75 0 0 0 0-1.5h-8.5Z" />
                    </svg>
                </button>
                <input type="number" value="0" placeholder="Seconds" min="0" max="60" v-model.trim.number="seconds"
                    class="w-full bg-transparent placeholder:text-slate-400 text-slate-700 dark:text-slate-200 text-sm border border-slate-200 rounded-md pl-3 pr-20 py-2 transition duration-300 ease focus:outline-none focus:border-slate-400 hover:border-slate-300 shadow-sm focus:shadow appearance-none [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none" />
                <button
                    class="absolute right-1 top-1 rounded bg-slate-800 p-1.5 border border-transparent text-center text-sm text-white transition-all shadow-sm hover:shadow focus:bg-slate-700 focus:shadow-none active:bg-slate-700 hover:bg-slate-700 active:shadow-none disabled:pointer-events-none disabled:opacity-50 disabled:shadow-none"
                    type="button" @click="incrementSeconds">
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" class="w-4 h-4">
                        <path
                            d="M8.75 3.75a.75.75 0 0 0-1.5 0v3.5h-3.5a.75.75 0 0 0 0 1.5h3.5v3.5a.75.75 0 0 0 1.5 0v-3.5h3.5a.75.75 0 0 0 0-1.5h-3.5v-3.5Z" />
                    </svg>
                </button>
            </div>
        </div>
        <!-- Seconds Input End -->
    </div>

    <p class="flex justify-items-normal mt-2 text-xs text-slate-400">
        Adjust minutes and seconds using the + and - controls, or
    </p>
    <p class="flex justify-items-normal text-xs text-slate-400">
        the Up and Down arrows on keyboard. Max allowed is 1 hour.
    </p>

    <p v-show="!isTimeValid" name="invalid-time" class="text-sm text-red-600 dark:text-red-300 mt-2">
        Please enter valid minutes/seconds values.
    </p>
    <p v-show="!isTimeValid" name="invalid-time" class="text-sm text-red-600 dark:text-red-300">
        Allowed range is 1 second to 60 minutes.
    </p>

    <div class="flex justify-center mt-8">
        <button type="button"
            class="px-4 py-2 text-sm w-full shadow-md font-medium rounded-md border bg-sky-100 hover:bg-sky-400 border-sky-300 text-sky-600 hover:text-white hover:border-transparent disabled:bg-gray-300 disabled:text-gray-500 disabled:border-gray-400 disabled:cursor-not-allowed dark:disabled:bg-gray-300 dark:disabled:text-gray-500 dark:disabled:border-gray-400 dark:bg-sky-800 dark:hover:bg-sky-600 dark:border-sky-700 dark:text-sky-100 select-none focus:outline-none focus:ring-0"
            v-if="!isCountdownInProgress" @click="start" :disabled="!isTimeValid">
            Start
        </button>

        <button type="button"
            class="px-4 py-2 text-sm w-full shadow-md font-medium rounded-md border bg-red-100 hover:bg-red-400 border-red-300 text-red-600 hover:text-white hover:border-transparent dark:bg-red-800 dark:hover:bg-red-600 dark:border-red-700 dark:text-red-100 select-none focus:outline-none focus:ring-0"
            v-if="isCountdownInProgress" @click="stop">
            Stop
        </button>
    </div>
</template>
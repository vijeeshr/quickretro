import { useI18n } from 'vue-i18n'
import { computed } from 'vue'

export const availableLocales = ['en', 'es', 'de', 'fr', 'ptBR', 'nl', 'it', 'pt', 'frCA'] as const

export type AvailableLocales = typeof availableLocales[number]

export function useLanguage() {
    const { locale: i18nLocale, getLocaleMessage } = useI18n()

    const locale = computed<AvailableLocales>({
        get: () => i18nLocale.value as AvailableLocales,
        set: (value) => {
            setLocale(value)
        }
    })

    const setLocale = (newLocale: AvailableLocales) => {
        i18nLocale.value = newLocale
        try {
            localStorage.setItem('lang', newLocale)
        } catch (error) {
            console.error('Failed to save locale:', error)
        }
    }

    const languageOptions = computed(() => 
        availableLocales.map(code => ({
            code,
            name: getLocaleMessage(code).langName
        }))
    )

    return {
        locale,
        setLocale,
        languageOptions,
        getLocaleMessage
    }
}
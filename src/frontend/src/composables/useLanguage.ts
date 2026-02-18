import { useI18n } from 'vue-i18n'
import { computed } from 'vue'
import { loadLanguageAsync } from '../i18n'
import { languageNames, availableLocales, type AvailableLocales } from '../i18n/locales'

export { languageNames, availableLocales, type AvailableLocales }

export function useLanguage() {
  const { locale: i18nLocale } = useI18n()

  const locale = computed<AvailableLocales>({
    get: () => i18nLocale.value as AvailableLocales,
    set: value => {
      setLocale(value)
    },
  })

  const setLocale = async (newLocale: AvailableLocales) => {
    try {
      await loadLanguageAsync(newLocale)
      localStorage.setItem('lang', newLocale)
    } catch (error) {
      console.error('Failed to load/save locale:', error)
    }
  }

  const languageOptions = computed(() =>
    availableLocales.map(code => ({
      code,
      name: languageNames[code],
    }))
  )

  const getLanguageName = (code: AvailableLocales) => languageNames[code]

  return {
    locale,
    setLocale,
    languageOptions,
    getLanguageName,
  }
}

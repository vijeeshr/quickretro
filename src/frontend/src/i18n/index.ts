import { createI18n, type Composer } from 'vue-i18n'
import { nextTick } from 'vue'
import { availableLocales, localeToFile, type AvailableLocales } from './locales'
import en from './en'

// Type definition for the message schema based on English
export type MessageSchema = typeof en

// Setup i18n instance with only English loaded initially
const i18n = createI18n({
  legacy: false, // Use Composition API
  locale: 'en' as AvailableLocales,
  fallbackLocale: 'en',
  messages: {
    en,
  },
})

export default i18n

export async function loadLanguageAsync(locale: AvailableLocales): Promise<void> {
  const global = i18n.global as Composer

  // If the same language is already loaded, just return
  if (global.locale.value === locale) {
    //return Promise.resolve()
    return
  }

  // If the language was already loaded previously, return
  if (global.availableLocales.includes(locale)) {
    global.locale.value = locale
    // return Promise.resolve()
    return
  }

  // If the language hasn't been loaded yet
  try {
    const modules = import.meta.glob('./*.ts')
    const filename = localeToFile[locale] ?? locale
    const loader = modules[`./${filename}.ts`]
    if (!loader) {
      throw new Error(`Locale ${locale} not found`)
    }
    const messages = (await loader()) as { default: MessageSchema }
    global.setLocaleMessage(locale, messages.default)
    global.locale.value = locale
  } catch (e) {
    console.error(`Failed to load language: ${locale}`, e)
  }

  return nextTick()
}

// Helper to detect initial language
function detectLocale(): AvailableLocales {
  const saved = localStorage.getItem('lang') as AvailableLocales
  if (saved && availableLocales.includes(saved)) return saved

  const locales = availableLocales as readonly string[]

  // Walk through the user's preferred languages in order
  for (const lang of navigator.languages) {
    // Try exact match with region (e.g. zh-CN -> zhCN, pt-BR -> ptBR, fr-CA -> frCA)
    const normalized = lang.replace('-', '')
    if (locales.includes(normalized)) return normalized as AvailableLocales

    // Try base language only (e.g. de-AT -> de, en-US -> en)
    const base = lang.split('-')[0]
    if (locales.includes(base)) return base as AvailableLocales
  }

  return 'en'
}

// Initial language load
const initialLocale = detectLocale()
if (initialLocale !== 'en') {
  loadLanguageAsync(initialLocale)
}

// Enables TypeScript autocomplete & key safety for vue-i18n
// based on the English message schema.
declare module 'vue-i18n' {
  // Define the vue-i18n type schema
  export interface DefineLocaleMessage extends MessageSchema {}
}

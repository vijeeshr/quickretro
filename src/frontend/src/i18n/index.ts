import { createI18n } from 'vue-i18n'
import { availableLocales } from '../composables/useLanguage'
import en from './en'
import es from './es'
import de from './de'
import fr from './fr'
import frCA from './fr-CA'
import ptBR from './pt-BR'
import pt from './pt'
import nl from './nl'
import it from './it'
import zhCN from './zh-CN'
import ja from './ja'
import ko from './ko'
import ru from './ru'
import uk from './uk'
import pl from './pl'

type MessageSchema = typeof en

function detectLocale(): string {
  const saved = localStorage.getItem('lang')
  if (saved) return saved

  const locales = availableLocales as readonly string[]

  // Walk through the user's preferred languages in order
  for (const lang of navigator.languages) {
    // Try exact match with region (e.g. zh-CN → zhCN, pt-BR → ptBR, fr-CA → frCA)
    const normalized = lang.replace('-', '')
    if (locales.includes(normalized)) return normalized

    // Try base language only (e.g. de-AT → de, en-US → en)
    const base = lang.split('-')[0]
    if (locales.includes(base)) return base
  }

  return 'en'
}

export default createI18n<
  [MessageSchema],
  | 'en'
  | 'zhCN'
  | 'es'
  | 'de'
  | 'fr'
  | 'ptBR'
  | 'ru'
  | 'ja'
  | 'nl'
  | 'ko'
  | 'it'
  | 'pt'
  | 'uk'
  | 'frCA'
  | 'pl'
>({
  legacy: false,
  locale: detectLocale(),
  fallbackLocale: 'en',
  messages: {
    en,
    zhCN,
    es,
    de,
    fr,
    ptBR,
    ru,
    ja,
    nl,
    ko,
    it,
    pt,
    uk,
    frCA,
    pl,
  },
})

declare module 'vue-i18n' {
  // Define the vue-i18n type schema
  export interface DefineLocaleMessage extends MessageSchema {}
}

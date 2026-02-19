export const languageNames = {
  en: 'English',
  zhCN: '简体中文 (zh-CN)',
  es: 'Español',
  de: 'Deutsch',
  fr: 'Français',
  ptBR: 'Português (Brasil)',
  ru: 'Русский (ru)',
  ja: '日本語 (ja)',
  nl: 'Nederlands',
  ko: '한국어 (ko)',
  it: 'Italiano',
  pt: 'Português',
  uk: 'Українська (uk)',
  frCA: 'Français (Canada)',
  pl: 'Polski',
  id: 'Bahasa Indonesia',
  vi: 'Tiếng Việt',
} as const

export type AvailableLocales = keyof typeof languageNames

export const availableLocales = Object.keys(languageNames) as AvailableLocales[]

// Map locale codes to filenames where they differ
export const localeToFile: Partial<Record<AvailableLocales, string>> = {
  zhCN: 'zh-CN',
  ptBR: 'pt-BR',
  frCA: 'fr-CA',
}

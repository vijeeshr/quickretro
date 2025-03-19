import { createI18n } from 'vue-i18n'
import en from './en'
import es from './es'
import de from './de'
import fr from './fr'
import frCA from './fr-CA'
import ptBR from './pt-BR'
import pt from './pt'
import nl from './nl'
import it from './it'

type MessageSchema = typeof en

const savedLanguage = localStorage.getItem('lang')

export default createI18n<[MessageSchema], 'en' | 'es' | 'de' | 'fr' | 'ptBR' | 'nl' | 'it' | 'pt' | 'frCA'>({
    legacy: false,
    locale: savedLanguage || 'en',
    fallbackLocale: 'en',
    messages : {
        en,
        es,
        de,
        fr, frCA,
        ptBR, pt,
        nl,
        it
    }
})

declare module 'vue-i18n' {
    // Define the vue-i18n type schema
    export interface DefineLocaleMessage extends MessageSchema {}
}
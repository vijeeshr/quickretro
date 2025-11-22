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
import zhCN from './zh-CN'
import ja from './ja'
import ko from './ko'
import ru from './ru'
import uk from './uk'
import pl from './pl'

type MessageSchema = typeof en

const savedLanguage = localStorage.getItem('lang')

export default createI18n<[MessageSchema], 'en' | 'zhCN' | 'es' | 'de' | 'fr' | 'ptBR' | 'ru' | 'ja' | 'nl' | 'ko' | 'it' | 'pt' | 'uk' | 'frCA' | 'pl'>({
    legacy: false,
    locale: savedLanguage || 'en',
    fallbackLocale: 'en',
    messages : {
        en, zhCN, es, de, fr, ptBR, ru, ja, nl, ko, it, pt, uk, frCA, pl
    }
})

declare module 'vue-i18n' {
    // Define the vue-i18n type schema
    export interface DefineLocaleMessage extends MessageSchema {}
}
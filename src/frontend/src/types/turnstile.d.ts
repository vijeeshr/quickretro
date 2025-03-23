declare namespace Turnstile {
    interface RenderParameters {
        sitekey: string
        callback?: (token: string) => void
        'error-callback'?: () => void
        'expired-callback'?: () => void
        theme?: 'auto' | 'light' | 'dark',
        size?: 'normal' | 'flexible' | 'compact'
        language?: string
    }

    function remove(widgetId: string): void

    function reset(widgetId: string): void

    function render(
        container: string | HTMLElement,
        params: RenderParameters
    ): string
}

declare interface App_Config {
    turnstileEnabled: boolean
    turnstileSiteKey: string
}
  
declare interface Window {
    APP_CONFIG?: typeof App_Config
    turnstile: typeof Turnstile
}
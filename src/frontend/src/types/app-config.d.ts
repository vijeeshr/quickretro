declare namespace Turnstile {
    interface RenderParameters {
        sitekey: string
        callback?: (token: string) => void
        'error-callback'?: () => void
        'expired-callback'?: () => void
        theme?: 'auto' | 'light' | 'dark'
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

// Runtime config injected by backend via /config.js
declare interface AppConfig {
    turnstile: {
        enabled: boolean
        siteKey: string
    }
    websocket: {
        maxMessageSizeBytes: number
    }
    data: {
        maxCategoryTextLength: number
        maxTextLength: number
    }
    frontend: {
        contentEditableInvalidDebounceMs: number
    }
    typingActivity: {
        enabled: boolean
        emitThrottleMs: number
        displayTimeoutMs: number
    }
}

declare interface Window {
    APP_CONFIG?: AppConfig
    turnstile: typeof Turnstile
}

export const appConfig = window.APP_CONFIG
export const TURNSTILE_ENABLED = appConfig?.turnstile.enabled ?? false
export const TURNSTILE_SITEKEY = appConfig?.turnstile.siteKey ?? ''
export const MAX_WEBSOCKET_MESSAGE_SIZE_BYTES = appConfig?.websocket.maxMessageSizeBytes ?? 1024
export const MAX_CATEGORY_TEXT_LENGTH = appConfig?.data.maxCategoryTextLength ?? 80
export const MAX_TEXT_LENGTH = appConfig?.data.maxTextLength ?? 80
export const CONTENT_EDITABLE_INVALID_DEBOUNCE_MS = appConfig?.frontend.contentEditableInvalidDebounceMs ?? 500
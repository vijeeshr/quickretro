/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_WS_PROTOCOL: 'ws' | 'wss'
  readonly VITE_SHOW_CONSOLE_LOGS: 'true' | 'false'
  readonly VITE_TURNSTILE_SCRIPT_URL: string
  readonly VITE_API_BASE_URL: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
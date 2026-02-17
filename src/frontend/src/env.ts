const isSecure = window.location.protocol === 'https:'

export const env = {
  wsProtocol: isSecure ? 'wss' : 'ws',
  showConsoleLogs: import.meta.env.VITE_SHOW_CONSOLE_LOGS === 'true',
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL,
  turnstileScriptUrl: import.meta.env.VITE_TURNSTILE_SCRIPT_URL,
}

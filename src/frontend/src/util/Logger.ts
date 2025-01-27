export function logMessage(...args: any[]): void {
    const showConsoleLogs = import.meta.env.VITE_SHOW_CONSOLE_LOGS === 'true'
    if (showConsoleLogs) {
        console.log(...args)
    }
}
  
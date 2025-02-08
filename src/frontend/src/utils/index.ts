import { EventRequest, SaveMessageEvent } from "../models/Requests"

export const logMessage = (...args: any[]): void => {
    const showConsoleLogs = import.meta.env.VITE_SHOW_CONSOLE_LOGS === 'true'
    if (showConsoleLogs) {
        console.log(...args)
    }
}

// export const getByteLength = (text: string) => new Blob([text]).size
export const getByteLength = (text: string): number => new TextEncoder().encode(text).length

// Binary search to efficiently trim text to fit within the byte limit
export const trimToMaxBytes = (text: string, maxBytes: number) => {
    let left = 0, right = text.length, validText = ""

    while (left <= right) {
        let mid = Math.floor((left + right) / 2)
        let testText = text.slice(0, mid)
        let byteLength = getByteLength(testText)

        if (byteLength <= maxBytes) {
            validText = testText // Store the valid substring
            left = mid + 1
        } else {
            right = mid - 1
        }
    }

    return validText
}

// Set the cursor at a specific character position
export const setCursorPosition = (element: HTMLElement, position: number): void => {
    const selection = window.getSelection()
    if (!selection) return
  
    const range = document.createRange();
    const child = element.firstChild; // Might be a text node
  
    if (child && child.nodeType === Node.TEXT_NODE) {
        const textNode = child as Text; // Explicitly cast to Text node
        const offset = Math.min(position, textNode.nodeValue?.length || 0)

        range.setStart(textNode, offset)
        range.setEnd(textNode, offset)
        selection.removeAllRanges()
        selection.addRange(range)
    }
}

export const canAssertMessageContentValidation = (): boolean => {

    const maxSizeConfig: string = import.meta.env.VITE_MAX_WEBSOCKET_MESSAGE_SIZE_BYTES
    const maxBytes: number = parseInt(maxSizeConfig, 10)

    return maxSizeConfig !== undefined && maxSizeConfig !== null && maxSizeConfig !== '' && 
        !isNaN(maxBytes) && maxBytes > 0
}

export const assertMessageContentValidation = (event: Event, user: string, nickname: string, board: string, category: string): MessageContentValidationResult => {

    let el = event.target as HTMLElement
    let text = el.innerText.trim()
    
    // emptyMessagePayload allows us to compute bytes available for the message content
    const emptyMessagePayload: EventRequest<SaveMessageEvent> = {
        typ: 'msg',
        pyl: { id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx', by: user, nickname: nickname, grp: board, msg: '', cat: category }
    }
    const emptyPayloadBytes: number = getByteLength(JSON.stringify(emptyMessagePayload))
    const maxAllowedBytes: number = parseInt(import.meta.env.VITE_MAX_WEBSOCKET_MESSAGE_SIZE_BYTES, 10)
    let maxAvailableBytesForContent: number = maxAllowedBytes > emptyPayloadBytes ? maxAllowedBytes - emptyPayloadBytes : 0
  
    const contentBytes: number = getByteLength(text)
    // logMessage('emptyPayloadBytes', emptyPayloadBytes, 'maxAvailableBytesForContent', maxAvailableBytesForContent, 'contentBytes', contentBytes)
    if (contentBytes > maxAvailableBytesForContent) {

        // Get current cursor position
        let selection = window.getSelection();
        if (!selection || selection.rangeCount === 0) // Check if selection exists
        {
            return {
                isValid: false,
                isTrimmed: false
            }
        }
        const range = selection.getRangeAt(0);
        const preCaretRange = range.cloneRange();
        preCaretRange.selectNodeContents(el);
        preCaretRange.setEnd(range.endContainer, range.endOffset);
        const cursorPosition: number = preCaretRange.toString().length; // Get cursor position in characters
    
        // Trim the text
        el.innerText = trimToMaxBytes(text, maxAvailableBytesForContent)
  
        // Restore cursor position
        setCursorPosition(el, cursorPosition)
  
        // Explicitly clear selection to avoid holding references
        // selection.removeAllRanges() // Todo: This somehow doesn't fire the toaster multiple times. Check it!

        return {
            isValid: false,
            isTrimmed: true
        }
    }

    return {
        isValid: true,
        isTrimmed: false
    }
}

export interface MessageContentValidationResult {
    isValid: boolean
    isTrimmed: boolean
}
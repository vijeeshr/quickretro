import { BoardColumn } from "../models/BoardColumn"
import { EventRequest, SaveMessageEvent } from "../models/Requests"

export const logMessage = (...args: any[]): void => {
    const showConsoleLogs = import.meta.env.VITE_SHOW_CONSOLE_LOGS === 'true'
    if (showConsoleLogs) {
        console.log(...args)
    }
}

// Show Unix timestamp as local date
export const formatDate = (timestamp: number): string => {
    if (!timestamp) return ""

    const date = new Date(timestamp * 1000)
    return date.toLocaleString()
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

export const assertMessageContentValidation = (event: Event, user: string, nickname: string, board: string, category: string, isComment: boolean = false): MessageContentValidationResult => {

    let el = event.target as HTMLElement
    let text = el.innerText.trim()
    
    // emptyMessagePayload allows us to compute bytes available for the message content
    const emptyMessagePayload: EventRequest<SaveMessageEvent> = {
        typ: 'msg',
        pyl: { id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx', by: user, nickname: nickname, grp: board, msg: '', cat: category, anon: false, pid: '' }
    }
    // Comment payload will have pid populated. Message payload will have pid as empty.
    if (isComment) {
        emptyMessagePayload.pyl.pid = 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx'
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

export const areBoardColumnsVisuallySame = (a: BoardColumn[], b: BoardColumn[]): boolean => {
    if (a.length !== b.length) return false
    // Sorting both arrays and then skipping "pos" field comparison.
    // This takes care of use-case where -
    //      a: A1 B2 E5 (fields C & D are disabled)
    //      b: A1 B2 E3 (fields C & D are disabled)
    // needs to be treated as "visually" same arrays. For UI, the visual display order of both arrays (with enabled items) are same. 
    const sortedA = [...a].sort((x, y) => x.pos - y.pos)
    const sortedB = [...b].sort((x, y) => x.pos - y.pos)

    return sortedA.every((col, i) => {
        const other = sortedB[i]
        return (
            col.id === other.id &&
            col.text === other.text &&
            col.isDefault === other.isDefault &&
            col.color === other.color
            // col.pos === other.pos
        )
    })
}

export const exceedsEventRequestMaxSize = <T>(eventType: string, payload: T) => {
    const event: EventRequest<T> = {
        typ: eventType,
        pyl: payload
    }
    const payloadBytes = getByteLength(JSON.stringify(event))
    const maxAllowedBytes = Number(import.meta.env.VITE_MAX_WEBSOCKET_MESSAGE_SIZE_BYTES)
    return payloadBytes > maxAllowedBytes
}
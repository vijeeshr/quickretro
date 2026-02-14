import { env } from "../env"
import { BoardColumn } from "../models/BoardColumn"
import { EventRequest, SaveMessageEvent } from "../models/Requests"
import { MAX_WEBSOCKET_MESSAGE_SIZE_BYTES } from "./appConfig"

export const logMessage = (...args: any[]): void => {
    if (env.showConsoleLogs) {
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
// export const getByteLength = (text: string): number => new TextEncoder().encode(text).length
const encoder = new TextEncoder() // Cache the encoder. Check impact.
export const getByteLength = (text: string): number => encoder.encode(text).length

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

/**
 * Calculates the overhead bytes of the JSON envelope.
 * Call this only when props change, not on every input.
*/
export const calculateContentBudget = (nickname: string, category: string, anon: boolean, isComment: boolean): number => {
    const emptyMessagePayload: EventRequest<SaveMessageEvent> = {
        typ: 'msg',
        pyl: {
            id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
            nickname: !anon ? nickname : '',
            msg: '',
            cat: category,
            anon: false,
            pid: isComment ? 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx' : ''
        }
    }

    const emptyPayloadBytes = getByteLength(JSON.stringify(emptyMessagePayload))
    // logMessage('calculateContentBudget.emptyPayloadBytes', emptyPayloadBytes, 'emptyPyl', JSON.stringify(emptyMessagePayload))
    return Math.max(0, MAX_WEBSOCKET_MESSAGE_SIZE_BYTES - emptyPayloadBytes)
}

export const assertMessageContentValidation = (event: Event, maxAvailableBytes: number): MessageContentValidationResult => {

    const el = event.target as HTMLElement
    const text = el.innerText.trim()
    
    // Fast exit
    // Avoid calling TextEncoder in getByteLength when text length is larger than maxAvailableBytes
    // The trimming needs to be done for larger text anyways
    if (text.length <= maxAvailableBytes) {
        const contentBytes: number = getByteLength(text)
        // logMessage('maxAvailableBytes', maxAvailableBytes, 'contentBytes', contentBytes)
        if (contentBytes <= maxAvailableBytes) {
            return { isValid: true, isTrimmed: false }
        }
    }

    // If we are here, we might be over the limit
    // Get current cursor position
    const selection = window.getSelection()
    if (!selection || selection.rangeCount === 0) return { isValid: false, isTrimmed: false } // Check if selection exists

    const range = selection.getRangeAt(0)
    const preCaretRange = range.cloneRange()
    preCaretRange.selectNodeContents(el)
    preCaretRange.setEnd(range.endContainer, range.endOffset)
    const cursorPosition: number = preCaretRange.toString().length

    // Perform the heavy trim
    el.innerText = trimToMaxBytes(text, maxAvailableBytes)

    // Restore cursor position
    setCursorPosition(el, cursorPosition)

    // Explicitly clear selection to avoid holding references
    // selection.removeAllRanges() // Todo: This somehow doesn't fire the toaster multiple times. Check it!

    return { isValid: false, isTrimmed: true }
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
    return payloadBytes > MAX_WEBSOCKET_MESSAGE_SIZE_BYTES
}

/**
 * Standard Trailing-Edge Debounce.
 * Delays execution until 'delay' ms after the last call.
 */
export function debounce<T extends (...args: any[]) => any>(fn: T, delay: number) {
    let timeoutId: ReturnType<typeof setTimeout> | null = null
    
    return function (this: any, ...args: Parameters<T>) {
        if (timeoutId) clearTimeout(timeoutId)
        timeoutId = setTimeout(() => {
            fn.apply(this, args)
        }, delay)
    }
}

/**
 * requestAnimationFrame Throttle.
 * Limits execution to once per browser paint (approx 60fps).
 */
export function throttleRAF<T extends (...args: any[]) => any>(fn: T) {
    let ticking = false

    return function(this: any, ...args: Parameters<T>) {
        if (!ticking) {
            requestAnimationFrame(() => {
                fn.apply(this, args)
                ticking = false
            })
            ticking = true
        }
    }
}

/**
 * Encodes the passed data to URL-safe string.
 * Safe for Non-ASCII chars.
 */
export const encodeToUrlSafeBase64FromJson = (data: any): string => {
    const json = JSON.stringify(data)

    // Convert string to UTF-8 bytes
    const bytes = new TextEncoder().encode(json)

    // Convert bytes to a "binary string" btoa can handle
    const binString = Array.from(bytes, (byte) => String.fromCharCode(byte)).join("")

    // Encode to Base64 and make it URL-safe
    return btoa(binString)
        .replace(/\+/g, '-')
        .replace(/\//g, '_')
        .replace(/=+$/, '')
}

/**
 * Decodes the passed URL-safe string to JSON.
 * Safe for Non-ASCII chars.
 */
export const decodeToJsonFromUrlSafeBase64 = (encoded: string): any => {
    // Restore URL-safe characters and padding
    let base64 = encoded.replace(/-/g, '+').replace(/_/g, '/')
    while (base64.length % 4) base64 += '='
    
    // Convert Base64 to binary string
    const binString = atob(base64)
    
    // Convert binary string to UTF-8 bytes
    const bytes = Uint8Array.from(binString, (m) => m.charCodeAt(0))
    
    // Decode bytes back to a string and parse JSON
    const jsonStr = new TextDecoder().decode(bytes)
    return JSON.parse(jsonStr)
}
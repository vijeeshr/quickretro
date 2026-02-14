// @vitest-environment happy-dom
import { describe, it, expect } from 'vitest'
import { encodeToUrlSafeBase64FromJson, decodeToJsonFromUrlSafeBase64 } from './index'

describe('Base64 Encoding/Decoding Utils', () => {
    describe('encodeToUrlSafeBase64FromJson', () => {
        it('should encode a simple object correctly', () => {
            const data = { hello: 'world' }
            const encoded = encodeToUrlSafeBase64FromJson(data)
            expect(encoded).toBeTypeOf('string')
            // JSON {"hello":"world"} -> eyJoZWxsbyI6IndvcmxkIn0=
            // Base64Url safe: no +, /, =
            expect(encoded).toMatch(/^[A-Za-z0-9\-_]+$/)
        })

        it('should handle special characters (unicode)', () => {
            const data = { temp: '20°C', emoji: '🍕' }
            const encoded = encodeToUrlSafeBase64FromJson(data)
            expect(encoded).toBeTypeOf('string')
        })

        it('should replace + with - and / with _', () => {
             // We need a specific input that produces + and / in standard base64
             // Standard Base64 for specific inputs might include them.
             // However, just ensuring the set output characters are safe is enough for unti testing the replacement logic if we assume btoa works standardly.
             const data = { a: '>>>>>>', b: '?????' } // Likely to produce + or /
             const encoded = encodeToUrlSafeBase64FromJson(data)
             expect(encoded).not.toContain('+')
             expect(encoded).not.toContain('/')
             expect(encoded).not.toContain('=')
        })
    })

    describe('decodeToJsonFromUrlSafeBase64', () => {
        it('should decode a valid encoded string back to original object', () => {
            const original = { id: 123, name: 'Test' }
            const encoded = encodeToUrlSafeBase64FromJson(original)
            const decoded = decodeToJsonFromUrlSafeBase64(encoded)
            expect(decoded).toEqual(original)
        })

        it('should round-trip special unicode characters correctly', () => {
            const original = { text: 'Hello World 🌍 ! @ # $ % ^ & * ()' }
            const encoded = encodeToUrlSafeBase64FromJson(original)
            const decoded = decodeToJsonFromUrlSafeBase64(encoded)
            expect(decoded).toEqual(original)
        })

        it('should handle empty objects', () => {
            const original = {}
            const encoded = encodeToUrlSafeBase64FromJson(original)
            const decoded = decodeToJsonFromUrlSafeBase64(encoded)
            expect(decoded).toEqual(original)
        })
        
        it('should handle arrays', () => {
             const original = [1, 2, "three"]
             const encoded = encodeToUrlSafeBase64FromJson(original)
             const decoded = decodeToJsonFromUrlSafeBase64(encoded)
             expect(decoded).toEqual(original)
        })
    })
})

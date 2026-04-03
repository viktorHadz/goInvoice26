import { describe, expect, it } from 'vitest'
import { resolvePickerMinutes } from '@/utils/pickerMinutes'

describe('resolvePickerMinutes', () => {
    it('uses the saved product minutes when no override was entered', () => {
        expect(resolvePickerMinutes(null, 30)).toBe(30)
    })

    it('keeps an explicit picker override', () => {
        expect(resolvePickerMinutes(45, 30)).toBe(45)
    })

    it('treats zero as a valid minute value', () => {
        expect(resolvePickerMinutes(0, 30)).toBe(0)
        expect(resolvePickerMinutes(null, 0)).toBe(0)
    })

    it('falls back to 60 minutes only when neither source is usable', () => {
        expect(resolvePickerMinutes(null, null)).toBe(60)
        expect(resolvePickerMinutes(-5, -10)).toBe(60)
    })
})

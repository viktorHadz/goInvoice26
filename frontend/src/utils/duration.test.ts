import { describe, expect, it } from 'vitest'
import { formatInvoiceDurationMinutes } from '@/utils/duration'

describe('formatInvoiceDurationMinutes', () => {
    it('formats zero minutes', () => {
        expect(formatInvoiceDurationMinutes(0)).toBe('0m')
    })

    it('formats minutes below one hour', () => {
        expect(formatInvoiceDurationMinutes(45)).toBe('45m')
    })

    it('formats exact hours', () => {
        expect(formatInvoiceDurationMinutes(60)).toBe('1h')
    })

    it('formats mixed hours and minutes', () => {
        expect(formatInvoiceDurationMinutes(90)).toBe('1h 30m')
    })

    it('formats multi-hour values', () => {
        expect(formatInvoiceDurationMinutes(125)).toBe('2h 5m')
    })
})

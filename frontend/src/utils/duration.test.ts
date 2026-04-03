import { describe, expect, it } from 'vitest'
import { formatInvoiceDurationMinutes, formatTimeRemaining } from '@/utils/duration'

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

describe('formatTimeRemaining', () => {
    const now = new Date('2026-04-03T12:00:00.000Z')

    it('returns null for invalid timestamps', () => {
        expect(formatTimeRemaining('not-a-date', now)).toBeNull()
    })

    it('formats minute-only durations', () => {
        expect(formatTimeRemaining('2026-04-03T12:45:00.000Z', now)).toBe('45m')
    })

    it('formats hours and minutes', () => {
        expect(formatTimeRemaining('2026-04-03T14:30:00.000Z', now)).toBe('2h 30m')
    })

    it('formats days and hours', () => {
        expect(formatTimeRemaining('2026-04-05T15:00:00.000Z', now)).toBe('2d 3h')
    })

    it('handles imminent or expired durations', () => {
        expect(formatTimeRemaining('2026-04-03T12:00:30.000Z', now)).toBe('less than a minute')
        expect(formatTimeRemaining('2026-04-03T11:59:30.000Z', now)).toBe('less than a minute')
    })
})

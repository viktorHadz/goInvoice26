import { describe, expect, it } from 'vitest'
import {
    fmtStrDate,
    isValidISODate,
    normalizeISODateOrNull,
    todayISO,
} from '@/utils/dates'

describe('dates helpers', () => {
    it('normalizes and validates only strict ISO dates', () => {
        expect(normalizeISODateOrNull(' 2026-03-24 ')).toBe('2026-03-24')
        expect(normalizeISODateOrNull('24/03/2026')).toBeNull()
        expect(isValidISODate('2026-02-29')).toBe(false)
        expect(isValidISODate('2028-02-29')).toBe(true)
    })

    it('formats ISO date strings without timezone drift', () => {
        expect(fmtStrDate('2026-03-24')).toBe('24/03/2026')
        expect(fmtStrDate('2026-03-24', 'mm/dd/yyyy')).toBe('03/24/2026')
        expect(fmtStrDate('2026-03-24', 'yyyy-mm-dd')).toBe('2026-03-24')
    })

    it('returns local ISO date for today helper', () => {
        expect(todayISO()).toMatch(/^\d{4}-\d{2}-\d{2}$/)
    })
})

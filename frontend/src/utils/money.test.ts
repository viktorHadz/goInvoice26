import { describe, expect, it } from 'vitest'
import type { Invoice, InvoiceLine } from '@/components/invoice/invoiceTypes'
import {
    calcBalanceDueMinor,
    calcDepositMinor,
    calcTotals,
    fromMinor,
    lineTotalMinor,
    toMinor,
} from '@/utils/money'

function line(overrides: Partial<InvoiceLine> & Pick<InvoiceLine, 'name'>): InvoiceLine {
    return {
        lineType: 'custom',
        pricingMode: 'flat',
        quantity: 1,
        unitPriceMinor: 0,
        minutesWorked: null,
        sortOrder: 1,
        ...overrides,
    }
}

function invoice(overrides: Partial<Invoice> = {}): Invoice {
    return {
        baseNumber: 1,
        clientId: 1,
        issueDate: '2025-01-01',
        clientSnapshot: { name: 'A', companyName: '', address: '', email: '' },
        lines: [line({ name: 'Item', sortOrder: 1, quantity: 1, unitPriceMinor: 10_000 })],
        discountType: 'none',
        discountMinor: 0,
        discountRate: 0,
        vatRate: 2000,
        paidMinor: 0,
        depositType: 'none',
        depositMinor: 0,
        depositRate: 0,
        ...overrides,
    }
}

describe('toMinor / fromMinor', () => {
    it('converts GBP to integer pence', () => {
        expect(toMinor(10)).toBe(1000)
        expect(toMinor(10.5)).toBe(1050)
        expect(toMinor(0)).toBe(0)
    })

    it('rounds non-terminating decimals to nearest penny', () => {
        expect(toMinor(0.1 + 0.2)).toBe(30)
    })

    it('fromMinor divides by 100', () => {
        expect(fromMinor(1000)).toBe(10)
        expect(fromMinor(0)).toBe(0)
    })
})

describe('lineTotalMinor', () => {
    it('flat: quantity times unit price (minor)', () => {
        const l = line({ name: 'x', quantity: 3, unitPriceMinor: 250, pricingMode: 'flat' })
        expect(lineTotalMinor(l)).toBe(750)
    })

    it('hourly: rounds qty * unit * minutes / 60', () => {
        const l = line({
            name: 'h',
            pricingMode: 'hourly',
            quantity: 1,
            unitPriceMinor: 6000,
            minutesWorked: 30,
        })
        expect(lineTotalMinor(l)).toBe(3000)
    })
})

describe('calcTotals', () => {
    it('sums lines, 20% VAT on amount after discount', () => {
        const inv = invoice()
        const t = calcTotals(inv)
        expect(t.subtotalMinor).toBe(10_000)
        expect(t.discountMinor).toBe(0)
        expect(t.subtotalAfterDiscountMinor).toBe(10_000)
        expect(t.vatMinor).toBe(2000)
        expect(t.totalMinor).toBe(12_000)
    })

    it('applies percent discount in bps', () => {
        const inv = invoice({
            discountType: 'percent',
            discountRate: 1000,
            discountMinor: 0,
        })
        const t = calcTotals(inv)
        expect(t.discountMinor).toBe(1000)
        expect(t.subtotalAfterDiscountMinor).toBe(9000)
        expect(t.vatMinor).toBe(1800)
        expect(t.totalMinor).toBe(10_800)
    })

    it('clamps fixed discount to subtotal', () => {
        const inv = invoice({
            discountType: 'fixed',
            discountMinor: 99_999,
            discountRate: 0,
        })
        const t = calcTotals(inv)
        expect(t.discountMinor).toBe(10_000)
        expect(t.subtotalAfterDiscountMinor).toBe(0)
        expect(t.vatMinor).toBe(0)
        expect(t.totalMinor).toBe(0)
    })
})

describe('calcDepositMinor', () => {
    const total = 12_000 as const

    it('none returns 0', () => {
        const inv = invoice({ depositType: 'none' })
        expect(calcDepositMinor(inv, total)).toBe(0)
    })

    it('fixed clamps to total', () => {
        const inv = invoice({ depositType: 'fixed', depositMinor: 5000, depositRate: 0 })
        expect(calcDepositMinor(inv, total)).toBe(5000)
        const huge = invoice({ depositType: 'fixed', depositMinor: 999_999, depositRate: 0 })
        expect(calcDepositMinor(huge, total)).toBe(total)
    })

    it('percent of total in bps', () => {
        const inv = invoice({ depositType: 'percent', depositRate: 2500, depositMinor: 0 })
        expect(calcDepositMinor(inv, total)).toBe(3000)
    })
})

describe('calcBalanceDueMinor', () => {
    it('is total minus paid, floored at 0', () => {
        expect(calcBalanceDueMinor(10_000, 2000, 3000)).toBe(7000)
        expect(calcBalanceDueMinor(10_000, 0, 10_000)).toBe(0)
        expect(calcBalanceDueMinor(10_000, 0, 12_000)).toBe(0)
    })

    it('uses cumulative paid value for revision snapshots', () => {
        const paidUpToRevision3 = 4500
        expect(calcBalanceDueMinor(12_000, 0, paidUpToRevision3)).toBe(7500)
    })
})

import { describe, expect, it } from 'vitest'
import type { Invoice, InvoiceLine } from '@/components/invoice/invoiceTypes'
import {
    addInvoiceLine,
    clearInvoiceDeposit,
    clearInvoiceDiscount,
    removeInvoiceLine,
    setInvoiceDepositFixedGBP,
    setInvoiceDepositPercent,
    setInvoiceDiscountFixedGBP,
    setInvoiceDiscountPercent,
    setInvoiceDueByDate,
    setInvoiceIssueDate,
    setInvoiceNote,
    setInvoiceSupplyDate,
    setInvoiceVatRateBps,
    updateInvoiceLine,
} from '@/utils/invoiceMutations'

function minimalLine(
    overrides: Partial<InvoiceLine> & Pick<InvoiceLine, 'name' | 'sortOrder'>,
): InvoiceLine {
    return {
        lineType: 'custom',
        pricingMode: 'flat',
        quantity: 1,
        unitPriceMinor: 100,
        minutesWorked: null,
        ...overrides,
    }
}

function minimalInvoice(overrides: Partial<Invoice> = {}): Invoice {
    return {
        baseNumber: 1,
        clientId: 1,
        issueDate: '2025-01-01',
        clientSnapshot: {
            name: 'Test',
            companyName: '',
            address: '',
            email: '',
        },
        lines: [],
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

describe('setInvoiceSupplyDate', () => {
    it('stores a trimmed value and clears blank input', () => {
        const inv = minimalInvoice()
        setInvoiceSupplyDate(inv, ' 2026-04-10 ')
        expect(inv.supplyDate).toBe('2026-04-10')

        setInvoiceSupplyDate(inv, '   ')
        expect(inv.supplyDate).toBeUndefined()
    })
})

describe('addInvoiceLine', () => {
    it('assigns incrementing sortOrder', () => {
        const inv = minimalInvoice({
            lines: [minimalLine({ name: 'a', sortOrder: 1 })],
        })
        addInvoiceLine(inv, {
            name: 'b',
            lineType: 'custom',
            pricingMode: 'flat',
            quantity: 1,
            unitPriceMinor: 50,
        })
        expect(inv.lines).toHaveLength(2)
        expect(inv.lines[1]).toMatchObject({ sortOrder: 2, name: 'b' })
    })

    it('merges quantity when same productId and not custom', () => {
        const inv = minimalInvoice({
            lines: [
                minimalLine({
                    name: 'p',
                    sortOrder: 1,
                    lineType: 'sample',
                    productId: 7,
                    quantity: 2,
                    unitPriceMinor: 100,
                }),
            ],
        })
        addInvoiceLine(inv, {
            name: 'p',
            lineType: 'sample',
            pricingMode: 'flat',
            productId: 7,
            quantity: 3,
            unitPriceMinor: 100,
        })
        expect(inv.lines).toHaveLength(1)
        expect(inv.lines[0]).toMatchObject({ quantity: 5 })
    })

    it('does not merge hourly lines when minutes differ', () => {
        const inv = minimalInvoice({
            lines: [
                minimalLine({
                    name: 'hourly sample',
                    sortOrder: 1,
                    lineType: 'sample',
                    pricingMode: 'hourly',
                    productId: 7,
                    quantity: 1,
                    unitPriceMinor: 10_000,
                    minutesWorked: 30,
                }),
            ],
        })

        addInvoiceLine(inv, {
            name: 'hourly sample',
            lineType: 'sample',
            pricingMode: 'hourly',
            productId: 7,
            quantity: 1,
            unitPriceMinor: 10_000,
            minutesWorked: 90,
        })

        expect(inv.lines).toHaveLength(2)
        expect(inv.lines.map((l) => l.minutesWorked)).toEqual([30, 90])
    })

    it('does not merge when the same product has a different rate or edited name', () => {
        const inv = minimalInvoice({
            lines: [
                minimalLine({
                    name: 'Edited line name',
                    sortOrder: 1,
                    lineType: 'sample',
                    pricingMode: 'flat',
                    productId: 7,
                    quantity: 1,
                    unitPriceMinor: 8_000,
                }),
            ],
        })

        addInvoiceLine(inv, {
            name: 'Original product name',
            lineType: 'sample',
            pricingMode: 'flat',
            productId: 7,
            quantity: 1,
            unitPriceMinor: 10_000,
            minutesWorked: null,
        })

        expect(inv.lines).toHaveLength(2)
        expect(inv.lines.map((l) => l.unitPriceMinor)).toEqual([8_000, 10_000])
    })
})

describe('updateInvoiceLine / removeInvoiceLine', () => {
    it('updateInvoiceLine patches by sortOrder', () => {
        const inv = minimalInvoice({
            lines: [minimalLine({ name: 'x', sortOrder: 2, quantity: 1 })],
        })
        updateInvoiceLine(inv, 2, { quantity: 4 })
        expect(inv.lines[0]).toMatchObject({ quantity: 4 })
    })

    it('removeInvoiceLine drops line and reindexes sortOrder', () => {
        const inv = minimalInvoice({
            lines: [
                minimalLine({ name: 'a', sortOrder: 1 }),
                minimalLine({ name: 'b', sortOrder: 2 }),
                minimalLine({ name: 'c', sortOrder: 3 }),
            ],
        })
        removeInvoiceLine(inv, 2)
        expect(inv.lines.map((l) => l.name)).toEqual(['a', 'c'])
        expect(inv.lines.map((l) => l.sortOrder)).toEqual([1, 2])
    })
})

describe('discount mutations', () => {
    it('setInvoiceDiscountFixedGBP', () => {
        const inv = minimalInvoice()
        setInvoiceDiscountFixedGBP(inv, 25.5)
        expect(inv.discountType).toBe('fixed')
        expect(inv.discountMinor).toBe(2550)
        expect(inv.discountRate).toBe(0)
    })

    it('setInvoiceDiscountPercent stores percent times 100 bps', () => {
        const inv = minimalInvoice()
        setInvoiceDiscountPercent(inv, 12.5)
        expect(inv.discountType).toBe('percent')
        expect(inv.discountRate).toBe(1250)
        expect(inv.discountMinor).toBe(0)
    })

    it('clearInvoiceDiscount', () => {
        const inv = minimalInvoice({
            discountType: 'fixed',
            discountMinor: 100,
            discountRate: 0,
        })
        clearInvoiceDiscount(inv)
        expect(inv.discountType).toBe('none')
        expect(inv.discountMinor).toBe(0)
        expect(inv.discountRate).toBe(0)
    })
})

describe('deposit mutations', () => {
    it('setInvoiceDepositFixedGBP', () => {
        const inv = minimalInvoice()
        setInvoiceDepositFixedGBP(inv, 10)
        expect(inv.depositType).toBe('fixed')
        expect(inv.depositMinor).toBe(1000)
        expect(inv.depositRate).toBe(0)
    })

    it('setInvoiceDepositPercent', () => {
        const inv = minimalInvoice()
        setInvoiceDepositPercent(inv, 20)
        expect(inv.depositType).toBe('percent')
        expect(inv.depositRate).toBe(2000)
        expect(inv.depositMinor).toBe(0)
    })

    it('clearInvoiceDeposit', () => {
        const inv = minimalInvoice({
            depositType: 'fixed',
            depositMinor: 500,
            depositRate: 0,
        })
        clearInvoiceDeposit(inv)
        expect(inv.depositType).toBe('none')
        expect(inv.depositMinor).toBe(0)
        expect(inv.depositRate).toBe(0)
    })
})

describe('misc', () => {
    it('setInvoiceNote', () => {
        const inv = minimalInvoice()
        setInvoiceNote(inv, 'hello')
        expect(inv.note).toBe('hello')
    })

    it('setInvoiceVatRateBps clamps', () => {
        const inv = minimalInvoice()
        setInvoiceVatRateBps(inv, 99999)
        expect(inv.vatRate).toBe(10000)
        setInvoiceVatRateBps(inv, -5)
        expect(inv.vatRate).toBe(0)
    })

    it('setInvoiceIssueDate trims incoming value', () => {
        const inv = minimalInvoice()
        setInvoiceIssueDate(inv, ' 2026-03-25 ')
        expect(inv.issueDate).toBe('2026-03-25')
    })

    it('setInvoiceDueByDate clears to undefined when empty', () => {
        const inv = minimalInvoice({ dueByDate: '2026-04-01' })
        setInvoiceDueByDate(inv, '')
        expect(inv.dueByDate).toBeUndefined()
    })
})

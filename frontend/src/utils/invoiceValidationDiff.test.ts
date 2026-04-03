import { describe, expect, it } from 'vitest'
import type { Invoice } from '@/components/invoice/invoiceTypes'
import { cloneInvoice } from '@/utils/cloneInvoice'
import { setInvoiceDepositFixedGBP, setInvoiceDiscountPercent } from '@/utils/invoiceMutations'
import { findNewInvoiceValidationMessage } from '@/utils/invoiceValidationDiff'

function makeInvoice(overrides: Partial<Invoice> = {}): Invoice {
    return {
        baseNumber: 1001,
        clientId: 1,
        status: 'draft',
        issueDate: '2026-03-31',
        dueByDate: '2026-04-14',
        clientSnapshot: {
            name: 'Client Name',
            companyName: 'Company',
            address: 'Address',
            email: 'client@example.com',
        },
        lines: [
            {
                productId: null,
                name: 'Line',
                lineType: 'custom',
                pricingMode: 'flat',
                quantity: 1,
                unitPriceMinor: 10_000,
                minutesWorked: null,
                sortOrder: 1,
            },
        ],
        discountType: 'none',
        discountMinor: 0,
        discountRate: 0,
        vatRate: 2000,
        paidMinor: 0,
        depositType: 'none',
        depositMinor: 0,
        depositRate: 0,
        note: '',
        ...overrides,
    }
}

describe('findNewInvoiceValidationMessage', () => {
    it('ignores unrelated pre-existing draft errors when applying a valid discount', () => {
        const before = makeInvoice({
            note: 'Line one\nLine two',
        })
        const after = cloneInvoice(before)
        setInvoiceDiscountPercent(after, 10)

        const message = findNewInvoiceValidationMessage(before, after, [], [
            'totals.discountRate',
            'totals.discountMinor',
            'totals.paidMinor',
        ])

        expect(message).toBeNull()
    })

    it('returns the new paid error when a deposit would overrun the amount owing', () => {
        const before = makeInvoice({
            paidMinor: 5_000,
        })
        const after = cloneInvoice(before)
        setInvoiceDepositFixedGBP(after, 80)

        const message = findNewInvoiceValidationMessage(before, after, [], [
            'totals.depositRate',
            'totals.depositMinor',
            'totals.paidMinor',
        ])

        expect(message).toBe('Paid amount cannot exceed the balance after deposit.')
    })
})

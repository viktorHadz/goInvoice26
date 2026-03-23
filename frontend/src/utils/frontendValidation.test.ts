import { describe, expect, it } from 'vitest'
import { validateInvoicePayload } from '@/utils/frontendValidation'

function basePayload() {
    return {
        overview: {
            sourceRevisionNo: undefined as number | undefined,
            issueDate: '2026-03-23',
            dueByDate: '2026-04-06',
            clientName: 'Client Name',
            clientCompanyName: 'Company',
            clientAddress: 'Address',
            clientEmail: 'test@example.com',
            note: 'N/A',
        },
        lines: [
            {
                productId: null,
                name: 'Line',
                lineType: 'custom',
                pricingMode: 'flat',
                quantity: 1,
                minutesWorked: null,
                unitPriceMinor: 10000,
                lineTotalMinor: 10000,
                sortOrder: 1,
            },
        ],
        totals: {
            vatRate: 2000,
            vatMinor: 2000,
            depositType: 'none',
            depositRate: 0,
            depositMinor: 0,
            discountType: 'none',
            discountRate: 0,
            discountMinor: 0,
            paidMinor: 5000,
            subtotalAfterDiscountMinor: 10000,
            subtotalMinor: 10000,
            totalMinor: 12000,
            balanceDueMinor: 7000,
        },
        payments: [
            {
                amountMinor: 5000,
                paymentDate: '2026-03-23',
            },
        ],
    }
}

describe('validateInvoicePayload payments', () => {
    it('accepts valid staged payment rows', () => {
        const errors = validateInvoicePayload(basePayload())
        expect(errors).toEqual({})
    })

    it('rejects invalid payment date', () => {
        const payload = basePayload()
        payload.payments[0]!.paymentDate = '23/03/2026'
        const errors = validateInvoicePayload(payload)
        expect(errors['payments[0].paymentDate']).toContain('valid ISO date')
    })

    it('rejects paidMinor lower than staged payment sum', () => {
        const payload = basePayload()
        payload.totals.paidMinor = 4900
        const errors = validateInvoicePayload(payload)
        expect(errors['totals.paidMinor']).toContain('staged payments')
    })

    it('rejects non-positive sourceRevisionNo when present', () => {
        const payload = basePayload()
        payload.overview.sourceRevisionNo = 0
        const errors = validateInvoicePayload(payload)
        expect(errors.sourceRevisionNo).toContain('positive integer')
    })
})

import { describe, expect, it } from 'vitest'
import { validateInvoicePayload } from '@/utils/frontendValidation'

function basePayload() {
    return {
        overview: {
            sourceRevisionNo: undefined as number | undefined,
            issueDate: '2026-03-23',
            supplyDate: undefined as string | undefined,
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
    }
}

describe('validateInvoicePayload invoice rules', () => {
    it('accepts valid invoice payloads without staged payments', () => {
        const errors = validateInvoicePayload(basePayload())
        expect(errors).toEqual({})
    })

    it('rejects invalid supply date', () => {
        const payload = basePayload()
        payload.overview.supplyDate = '23/03/2026'
        const errors = validateInvoicePayload(payload)
        expect(errors.supplyDate).toContain('YYYY-MM-DD')
    })

    it('rejects paidMinor above the invoice total', () => {
        const payload = basePayload()
        payload.totals.paidMinor = 12_100
        const errors = validateInvoicePayload(payload)
        expect(errors['totals.paidMinor']).toContain('invoice total')
    })

    it('rejects non-positive sourceRevisionNo when present', () => {
        const payload = basePayload()
        payload.overview.sourceRevisionNo = 0
        const errors = validateInvoicePayload(payload)
        expect(errors.sourceRevisionNo).toContain('positive integer')
    })

    it('rejects sortOrder values below one', () => {
        const payload = basePayload()
        payload.lines[0]!.sortOrder = 0
        const errors = validateInvoicePayload(payload)
        expect(errors['lines[0].sortOrder']).toBe('Sort order must be 1 or more.')
    })
})

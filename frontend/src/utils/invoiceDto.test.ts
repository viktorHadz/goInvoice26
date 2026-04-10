import { describe, expect, it } from 'vitest'
import type { Invoice } from '@/components/invoice/invoiceTypes'
import { apiDTO } from '@/utils/invoiceDto'

function makeInvoice(overrides: Partial<Invoice> = {}): Invoice {
    return {
        baseNumber: 9001,
        clientId: 7,
        status: 'draft',
        issueDate: '2026-03-24',
        supplyDate: undefined,
        dueByDate: undefined,
        clientSnapshot: {
            name: 'Client',
            companyName: '',
            address: '',
            email: '',
        },
        lines: [
            {
                productId: null,
                name: 'Line',
                lineType: 'custom',
                pricingMode: 'flat',
                quantity: 1,
                unitPriceMinor: 1000,
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

describe('apiDTO date contract', () => {
    it('keeps ISO issueDate values unchanged and carries supplyDate when set', () => {
        const dto = apiDTO(makeInvoice({ supplyDate: '2026-03-25' }))
        expect(dto.overview.issueDate).toBe('2026-03-24')
        expect(dto.overview.supplyDate).toBe('2026-03-25')
    })

    it('omits dueByDate when cleared', () => {
        const dto = apiDTO(makeInvoice({ dueByDate: undefined }))
        expect('dueByDate' in dto.overview).toBe(false)
    })

    it('omits supplyDate when cleared', () => {
        const dto = apiDTO(makeInvoice({ supplyDate: undefined }))
        expect('supplyDate' in dto.overview).toBe(false)
    })
})

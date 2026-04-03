import { describe, expect, it } from 'vitest'
import {
    buildInvoiceStatusContext,
    canDeleteInvoice,
    canEditInvoice,
    reachableStatuses,
    saveModeForStatus,
} from '@/utils/invoiceStatusOptions'
import type { Invoice } from '@/components/invoice/invoiceTypes'

function makeInvoice(overrides: Partial<Invoice> = {}): Invoice {
    return {
        baseNumber: 101,
        clientId: 42,
        status: 'issued',
        issueDate: '2026-03-20',
        dueByDate: '2026-03-30',
        clientSnapshot: {
            name: 'Alex',
            companyName: 'Acme Co',
            address: '1 Test Road',
            email: 'alex@example.com',
        },
        lines: [
            {
                productId: 1,
                name: 'Service line',
                lineType: 'custom',
                pricingMode: 'flat',
                quantity: 1,
                unitPriceMinor: 10000,
                minutesWorked: null,
                sortOrder: 1,
            },
        ],
        discountType: 'none',
        discountMinor: 0,
        discountRate: 0,
        vatRate: 0,
        paidMinor: 0,
        depositType: 'none',
        depositMinor: 0,
        depositRate: 0,
        note: 'Initial',
        ...overrides,
    }
}

describe('invoice status lifecycle helpers', () => {
    it('exposes the lifecycle transition options', () => {
        expect(reachableStatuses('draft')).toEqual(['draft', 'issued'])
        expect(reachableStatuses('issued')).toEqual(['issued', 'paid', 'void'])
        expect(
            reachableStatuses('issued', {
                canReturnIssuedToDraft: true,
            }),
        ).toEqual(['issued', 'draft', 'paid', 'void'])
        expect(reachableStatuses('paid')).toEqual(['paid'])
        expect(
            reachableStatuses('paid', {
                canReopenPaidToIssued: true,
            }),
        ).toEqual(['paid', 'issued'])
        expect(reachableStatuses('void')).toEqual(['void'])
    })

    it('matches edit and delete locks to the lifecycle rules', () => {
        expect(canEditInvoice('draft')).toBe(true)
        expect(canEditInvoice('issued')).toBe(true)
        expect(canEditInvoice('paid')).toBe(false)
        expect(canEditInvoice('void')).toBe(false)

        expect(canDeleteInvoice('draft')).toBe(true)
        expect(canDeleteInvoice('issued')).toBe(true)
        expect(canDeleteInvoice('paid')).toBe(true)
        expect(canDeleteInvoice('void')).toBe(false)
    })

    it('maps save mode by status', () => {
        expect(saveModeForStatus('draft')).toBe('draft')
        expect(saveModeForStatus('issued')).toBe('revision')
        expect(saveModeForStatus('paid')).toBe('locked')
        expect(saveModeForStatus('void')).toBe('locked')
    })

    it('builds contextual lifecycle flags from invoice data', () => {
        expect(buildInvoiceStatusContext(makeInvoice(), 1)).toEqual({
            canReturnIssuedToDraft: true,
            canReopenPaidToIssued: true,
        })

        expect(
            buildInvoiceStatusContext(
                makeInvoice({
                    status: 'paid',
                    paidMinor: 10000,
                }),
                3,
            ),
        ).toEqual({
            canReturnIssuedToDraft: false,
            canReopenPaidToIssued: false,
        })
    })
})

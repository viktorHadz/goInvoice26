import { describe, expect, it } from 'vitest'
import {
    formatActiveEditorNodeLabel,
    formatInvoiceBaseLabel,
    formatInvoiceDisplayLabel,
    formatPaymentReceiptLabel,
    toDisplayRevisionNo,
} from '@/utils/invoiceLabels'

describe('invoiceLabels display mapping', () => {
    it('maps DB revisions to display revisions', () => {
        expect(toDisplayRevisionNo(undefined)).toBeNull()
        expect(toDisplayRevisionNo(1)).toBeNull()
        expect(toDisplayRevisionNo(2)).toBe(1)
        expect(toDisplayRevisionNo(3)).toBe(2)
    })

    it('formats base, revision, and receipt labels with explicit suffixes', () => {
        expect(formatInvoiceBaseLabel('INV-', 7)).toBe('INV-7')
        expect(formatInvoiceDisplayLabel('INV-', 7, 1)).toBe('INV-7')
        expect(formatInvoiceDisplayLabel('INV-', 7, 2)).toBe('INV-7-Rev-1')
        expect(formatInvoiceDisplayLabel('INV-', 7, 3)).toBe('INV-7-Rev-2')
        expect(formatPaymentReceiptLabel('INV-', 7, 1)).toBe('INV-7-PR-1')
    })

    it('formats active editor node labels through shared mapping', () => {
        expect(
            formatActiveEditorNodeLabel('INV-', {
                type: 'invoice',
                clientId: 1,
                id: 10,
                baseNo: 7,
            }),
        ).toBe('INV-7')

        expect(
            formatActiveEditorNodeLabel('INV-', {
                type: 'revision',
                clientId: 1,
                id: 20,
                invoiceId: 10,
                baseNo: 7,
                revisionNo: 2,
            }),
        ).toBe('INV-7-Rev-1')

        expect(
            formatActiveEditorNodeLabel('INV-', {
                type: 'paymentReceipt',
                clientId: 1,
                id: 30,
                invoiceId: 10,
                baseNo: 7,
                receiptNo: 2,
            }),
        ).toBe('INV-7-PR-2')
    })
})

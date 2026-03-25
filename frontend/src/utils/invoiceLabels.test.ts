import { describe, expect, it } from 'vitest'
import {
    formatActiveEditorNodeLabel,
    formatInvoiceBaseLabel,
    formatInvoiceDisplayLabel,
    toDisplayRevisionNo,
} from '@/utils/invoiceLabels'

describe('invoiceLabels display mapping', () => {
    it('maps DB revisions to display revisions', () => {
        expect(toDisplayRevisionNo(undefined)).toBeNull()
        expect(toDisplayRevisionNo(1)).toBeNull()
        expect(toDisplayRevisionNo(2)).toBe(1)
        expect(toDisplayRevisionNo(3)).toBe(2)
    })

    it('formats base and display labels with shifted revision suffix', () => {
        expect(formatInvoiceBaseLabel('INV-', 7)).toBe('INV - 7')
        expect(formatInvoiceDisplayLabel('INV-', 7, 1)).toBe('INV - 7')
        expect(formatInvoiceDisplayLabel('INV-', 7, 2)).toBe('INV - 7.1')
        expect(formatInvoiceDisplayLabel('INV-', 7, 3)).toBe('INV - 7.2')
    })

    it('formats active editor node labels through shared mapping', () => {
        expect(
            formatActiveEditorNodeLabel('INV-', {
                type: 'invoice',
                id: 10,
                baseNo: 7,
            }),
        ).toBe('INV - 7')

        expect(
            formatActiveEditorNodeLabel('INV-', {
                type: 'revision',
                id: 20,
                invoiceId: 10,
                baseNo: 7,
                revisionNo: 2,
            }),
        ).toBe('INV - 7.1')
    })
})

import { describe, expect, it } from 'vitest'
import type { InvoiceLine } from '@/components/invoice/invoiceTypes'
import { editorPreviewLineTotalMinor, formatEditorPreviewLineMeta } from '@/utils/editorPreview'

function line(overrides: Partial<InvoiceLine> = {}): InvoiceLine {
    return {
        name: 'Line item',
        lineType: 'sample',
        pricingMode: 'flat',
        quantity: 1,
        unitPriceMinor: 0,
        minutesWorked: null,
        sortOrder: 1,
        ...overrides,
    }
}

describe('formatEditorPreviewLineMeta', () => {
    it('shows humanized hourly minutes including zero', () => {
        expect(
            formatEditorPreviewLineMeta(
                line({
                    pricingMode: 'hourly',
                    minutesWorked: 0,
                }),
            ),
        ).toBe('hourly · 0m')
    })

    it('falls back to pricing mode for flat lines', () => {
        expect(formatEditorPreviewLineMeta(line())).toBe('flat')
    })
})

describe('editorPreviewLineTotalMinor', () => {
    it('uses hourly line-total semantics for preview rows', () => {
        expect(
            editorPreviewLineTotalMinor(
                line({
                    pricingMode: 'hourly',
                    quantity: 2,
                    unitPriceMinor: 6000,
                    minutesWorked: 90,
                }),
            ),
        ).toBe(18000)
    })
})

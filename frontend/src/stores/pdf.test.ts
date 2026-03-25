import { describe, expect, it } from 'vitest'
import { flattenValidationErrors } from '@/stores/pdf'

describe('flattenValidationErrors', () => {
    it('formats nested fields into readable labels', () => {
        const message = flattenValidationErrors({
            'lines[0].lineTotalMinor': 'Line total does not match quantity and unit price.',
            'payments[1].paymentDate': 'Choose a payment date.',
            'totals.paidMinor': 'Paid amount must include all staged payments.',
        })

        expect(message).toContain('Line 1 lineTotalMinor:')
        expect(message).toContain('Payment 2 paymentDate:')
        expect(message).toContain('totals paidMinor:')
        expect(message).toContain(';')
    })
})

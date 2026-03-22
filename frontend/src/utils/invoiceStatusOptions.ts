import type { InvoiceStatus } from '@/components/invoice/invoiceTypes'

/** Valid one-step status values for the invoice lifecycle dropdown from a given current status. */
export function reachableStatuses(from: InvoiceStatus): InvoiceStatus[] {
    switch (from) {
        case 'draft':
            return ['draft', 'issued', 'paid', 'void']
        case 'issued':
            return ['issued', 'paid', 'void']
        case 'paid':
            return ['paid', 'issued']
        case 'void':
            return ['void', 'issued']
    }
}

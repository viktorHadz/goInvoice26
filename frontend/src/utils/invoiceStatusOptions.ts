import type { Invoice, InvoiceStatus } from '@/components/invoice/invoiceTypes'
import { calcDepositMinor, calcTotals } from '@/utils/money'

export type InvoiceStatusContext = {
    canReturnIssuedToDraft?: boolean
    canReopenPaidToIssued?: boolean
}

export function buildInvoiceStatusContext(
    invoice: Invoice | null | undefined,
    revisionCount = 1,
): InvoiceStatusContext {
    if (!invoice) {
        return {
            canReturnIssuedToDraft: revisionCount <= 1,
            canReopenPaidToIssued: false,
        }
    }

    const totals = calcTotals(invoice)
    const depositMinor = calcDepositMinor(invoice, totals.totalMinor)
    const expectedPaidMinor = Math.max(0, totals.totalMinor - depositMinor)

    return {
        canReturnIssuedToDraft: revisionCount <= 1,
        canReopenPaidToIssued: invoice.paidMinor !== expectedPaidMinor,
    }
}

/** Valid one-step status values for the invoice lifecycle dropdown from a given current status. */
export function reachableStatuses(
    from: InvoiceStatus,
    context: InvoiceStatusContext = {},
): InvoiceStatus[] {
    switch (from) {
        case 'draft':
            return ['draft', 'issued']
        case 'issued':
            return context.canReturnIssuedToDraft
                ? ['issued', 'draft', 'paid', 'void']
                : ['issued', 'paid', 'void']
        case 'paid':
            return context.canReopenPaidToIssued ? ['paid', 'issued'] : ['paid']
        case 'void':
            return ['void']
    }
}

export function canEditInvoice(status: InvoiceStatus): boolean {
    return status === 'draft' || status === 'issued'
}

export function canDeleteInvoice(status: InvoiceStatus): boolean {
    return status !== 'void'
}

export function saveModeForStatus(status: InvoiceStatus): 'draft' | 'revision' | 'locked' {
    if (status === 'draft') return 'draft'
    if (status === 'issued') return 'revision'
    return 'locked'
}

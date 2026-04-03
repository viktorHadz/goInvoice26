import type { Invoice } from '@/components/invoice/invoiceTypes'
import { apiDTO, type DraftPaymentInput } from '@/utils/invoiceDto'
import { validateInvoicePayload } from '@/utils/frontendValidation'

/**
 * Only block an optimistic invoice adjustment when it introduces a new watched error.
 * Existing unrelated draft errors should not cause the adjustment itself to be reverted.
 */
export function findNewInvoiceValidationMessage(
    beforeInvoice: Invoice,
    afterInvoice: Invoice,
    payments: DraftPaymentInput[],
    watchedFields: string[],
): string | null {
    const beforeErrors = validateInvoicePayload(apiDTO(beforeInvoice, payments))
    const afterErrors = validateInvoicePayload(apiDTO(afterInvoice, payments))

    for (const field of watchedFields) {
        const afterMessage = afterErrors[field]
        if (!afterMessage) continue
        if (beforeErrors[field] === afterMessage) continue
        return afterMessage
    }

    return null
}

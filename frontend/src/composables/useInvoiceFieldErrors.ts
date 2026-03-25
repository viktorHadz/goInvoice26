import { computed, type Ref } from 'vue'
import type { Invoice } from '@/components/invoice/invoiceTypes'
import type { InvoicePricing } from '@/composables/useInvoicePricing'
import { validateInvoicePayload } from '@/utils/frontendValidation'
import { apiDTO } from '@/utils/invoiceDto'

/**
 * Derives invoice field validation from a reactive invoice + server errors.
 *
 * This composable is intentionally READ-ONLY:
 * - It does NOT own invoice state
 * - It does NOT mutate anything
 * - It only computes validation based on provided refs
 *
 * Usage:
 * - InvoiceView: pass invoice store refs
 * - EditorSurface: pass activeInvoice refs
 *
 * Precedence:
 * - Live (client-side) validation takes priority
 * - Falls back to serverFieldErrors if no live error exists
 *
 * @param invoice - reactive invoice draft (store or editor)
 * @param serverFieldErrors - reactive server validation errors
 *
 * @returns liveFieldErrors - computed map of current validation errors
 * @returns getFieldError - helper to resolve a single field error
 *
 * !! Keep this composable pure/derived.
 * Do NOT add mutation logic or business workflows here.
 */
export function useInvoiceFieldErrors(
    invoice: Ref<Invoice | null>,
    serverFieldErrors: Ref<Record<string, string>>,
    pricing?: Ref<InvoicePricing | null>,
) {
    const liveFieldErrors = computed<Record<string, string>>(() => {
        const inv = invoice.value
        if (!inv) return {}

        const dto = apiDTO(inv, [], { pricing: pricing?.value ?? undefined })
        return validateInvoicePayload(dto)
    })

    function getFieldError(field: string): string | null {
        return liveFieldErrors.value[field] ?? serverFieldErrors.value[field] ?? null
    }

    return {
        liveFieldErrors,
        getFieldError,
    }
}

import type { Invoice } from '@/components/invoice/invoiceTypes'
import { isApiError } from '@/utils/apiErrors'
import { NetworkError } from '@/utils/fetchHelper'
import { validateInvoicePayload } from '@/utils/frontendValidation'
import { generatePdfHandler } from '@/utils/invoiceHttpHandler'
import { emitToastError, emitToastInfo, emitToastSuccess } from '@/utils/toast'
import { defineStore } from 'pinia'
import { useInvoiceStore } from './invoice'
import { apiDTO } from '@/utils/invoiceDto'

export const usePdfStore = defineStore('pdf', () => {
    const inv = useInvoiceStore()

    async function handlePdfGeneration(
        handler: () => Promise<void>,
        successMessage: string,
        showInfoToast: boolean = false,
    ) {
        try {
            await handler()
            emitToastSuccess(successMessage)
            if (showInfoToast) {
                emitToastInfo('Consider saving draft to enable revisions.')
            }
        } catch (err) {
            if (isApiError(err)) {
                console.error('[invoice pdf api error]', err)
                emitToastError({
                    id: err.id,
                    title: 'Could not generate PDF',
                    message: err.message || 'Please try again.',
                })
                return
            }

            if (err instanceof NetworkError) {
                emitToastError({
                    title: 'Network error',
                    message: 'Could not reach the server. Please check your connection.',
                })
                return
            }

            emitToastError({
                title: 'Could not generate PDF',
                message: 'An unexpected error occurred. Please try again.',
            })
            console.error('[invoice pdf]', err)
        }
    }

    async function generateAndPersistPdf(
        clientId: number,
        baseNumber: number,
        revisionNumber: number = 1,
    ) {
        await handlePdfGeneration(
            () => generatePdfHandler(clientId, baseNumber, 'save', revisionNumber),
            'PDF downloaded successfully.',
        )
    }

    async function quickGeneratePDF(invoice: Invoice, revisionNumber: number = 1) {
        const invo = apiDTO(invoice)

        const errors = validateInvoicePayload(invo)
        if (Object.keys(errors).length > 0) {
            emitToastError({
                title: 'Invalid invoice data',
                message: flattenValidationErrors(errors),
            })
            return
        }

        await handlePdfGeneration(
            () =>
                generatePdfHandler(
                    invo.overview.clientId,
                    invo.overview.baseNumber,
                    'generate',
                    revisionNumber,
                    invo,
                ),
            'Quick PDF generated successfully.',
            true,
        )
    }
    function flattenValidationErrors(
        errors: Record<string, string | string[] | undefined>,
    ): string {
        const parts: string[] = []

        for (const [field, value] of Object.entries(errors)) {
            if (!value) continue

            if (Array.isArray(value)) {
                for (const msg of value) {
                    parts.push(`${field}: ${msg}`)
                }
            } else {
                parts.push(`${field}: ${value}`)
            }
        }

        return parts.join(', ')
    }
    return {
        generateAndPersistPdf,
        quickGeneratePDF,
    }
})

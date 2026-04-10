import type { Invoice } from '@/components/invoice/invoiceTypes'
import { isApiError } from '@/utils/apiErrors'
import { NetworkError } from '@/utils/fetchHelper'
import { validateInvoicePayload } from '@/utils/frontendValidation'
import {
    generateDocxHandler,
    generatePaymentReceiptDownloadHandler,
    generatePdfHandler,
} from '@/utils/invoiceHttpHandler'
import { emitToastError, emitToastSuccess } from '@/utils/toast'
import { defineStore } from 'pinia'
import { apiDTO } from '@/utils/invoiceDto'

export function flattenValidationErrors(
    errors: Record<string, string | string[] | undefined>,
): string {
    const toLabel = (field: string) => {
        const lineMatch = field.match(/^lines\[(\d+)\]\.(.+)$/)
        if (lineMatch) {
            const idx = Number(lineMatch[1]) + 1
            const nested = (lineMatch[2] ?? '').split('.').join(' ')
            return `Line ${idx} ${nested}`
        }

        const paymentMatch = field.match(/^payments\[(\d+)\]\.(.+)$/)
        if (paymentMatch) {
            const idx = Number(paymentMatch[1]) + 1
            const nested = (paymentMatch[2] ?? '').split('.').join(' ')
            return `Payment ${idx} ${nested}`
        }

        return field.split('.').join(' ')
    }

    const parts: string[] = []

    for (const [field, value] of Object.entries(errors)) {
        if (!value) continue

        if (Array.isArray(value)) {
            for (const msg of value) {
                parts.push(`${toLabel(field)}: ${msg}`)
            }
        } else {
            parts.push(`${toLabel(field)}: ${value}`)
        }
    }

    return parts.join('; ')
}

export const usePdfStore = defineStore('pdf', () => {
    async function handleFileGeneration(
        formatLabel: 'PDF' | 'DOCX',
        handler: () => Promise<void>,
        successMessage?: string,
    ) {
        try {
            await handler()
            if (successMessage) {
                emitToastSuccess(successMessage)
            }
        } catch (err) {
            if (isApiError(err)) {
                console.error(`[invoice ${formatLabel.toLowerCase()} api error]`, err)
                emitToastError({
                    id: err.id,
                    title: `Could not generate ${formatLabel}`,
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
                title: `Could not generate ${formatLabel}`,
                message: 'An unexpected error occurred. Please try again.',
            })
            console.error(`[invoice ${formatLabel.toLowerCase()}]`, err)
        }
    }

    async function generateAndPersistPdf(
        clientId: number,
        baseNumber: number,
        revisionNumber: number = 1,
    ) {
        await handleFileGeneration('PDF', () =>
            generatePdfHandler(clientId, baseNumber, 'save', revisionNumber),
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

        await handleFileGeneration(
            'PDF',
            () =>
                generatePdfHandler(
                    invo.overview.clientId,
                    invo.overview.baseNumber,
                    'generate',
                    revisionNumber,
                    invo,
                ),
            'Quick PDF generated successfully.',
        )
    }

    async function generateAndPersistDocx(
        clientId: number,
        baseNumber: number,
        revisionNumber: number = 1,
    ) {
        await handleFileGeneration('DOCX', () =>
            generateDocxHandler(clientId, baseNumber, 'save', revisionNumber),
        )
    }

    async function quickGenerateDocx(invoice: Invoice, revisionNumber: number = 1) {
        const invo = apiDTO(invoice)

        const errors = validateInvoicePayload(invo)
        if (Object.keys(errors).length > 0) {
            emitToastError({
                title: 'Invalid invoice data',
                message: flattenValidationErrors(errors),
            })
            return
        }

        await handleFileGeneration(
            'DOCX',
            () =>
                generateDocxHandler(
                    invo.overview.clientId,
                    invo.overview.baseNumber,
                    'generate',
                    revisionNumber,
                    invo,
                ),
            'Quick DOCX generated successfully.',
        )
    }

    async function generateSavedPaymentReceiptPdf(
        clientId: number,
        baseNumber: number,
        revisionNumber: number,
        receiptNo: number,
        prefix = 'INV',
    ) {
        await handleFileGeneration('PDF', () =>
            generatePaymentReceiptDownloadHandler(
                clientId,
                baseNumber,
                revisionNumber,
                receiptNo,
                'pdf',
                prefix,
            ),
        )
    }

    async function generateSavedPaymentReceiptDocx(
        clientId: number,
        baseNumber: number,
        revisionNumber: number,
        receiptNo: number,
        prefix = 'INV',
    ) {
        await handleFileGeneration('DOCX', () =>
            generatePaymentReceiptDownloadHandler(
                clientId,
                baseNumber,
                revisionNumber,
                receiptNo,
                'docx',
                prefix,
            ),
        )
    }

    return {
        generateAndPersistPdf,
        generateAndPersistDocx,
        quickGeneratePDF,
        quickGenerateDocx,
        generateSavedPaymentReceiptPdf,
        generateSavedPaymentReceiptDocx,
    }
})

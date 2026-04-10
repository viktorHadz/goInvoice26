import { NetworkError, request } from './fetchHelper'
import { parseApiError } from './apiErrors'
import { formatPaymentReceiptLabel, toDisplayRevisionNo } from './invoiceLabels'

export type CreateInvoiceResponse = {
    invoiceId: number
    revisionId: number
}

export type CreateRevisionResponse = CreateInvoiceResponse & {
    revisionNo: number
}

export type CreatePaymentReceiptResponse = {
    invoiceId: number
    receiptId: number
    receiptNo: number
}

export async function getNewInvoiceNumber(clientId: number): Promise<number> {
    const n = await request<number>(`/api/clients/${clientId}/invoice`)
    const out = typeof n === 'number' && Number.isFinite(n) ? Math.round(n) : 0
    return out > 0 ? out : 0
}

export async function newInvoiceHandler(
    clientId: number,
    baseNumber: number,
    invoPayload: unknown,
) {
    const url = `/api/clients/${clientId}/invoice/${baseNumber}`
    const payload = JSON.stringify(invoPayload)

    return await request<CreateInvoiceResponse>(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: payload,
    })
}

export async function updateDraftInvoiceHandler(
    clientId: number,
    baseNumber: number,
    invoPayload: unknown,
) {
    const url = `/api/clients/${clientId}/invoice/${baseNumber}`
    const payload = JSON.stringify(invoPayload)

    return await request<CreateInvoiceResponse>(url, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: payload,
    })
}

export async function newRevisionHandler(
    clientId: number,
    baseNumber: number,
    invoPayload: unknown,
) {
    const url = `/api/clients/${clientId}/invoice/${baseNumber}/revisions`
    const payload = JSON.stringify(invoPayload)
    return await request<CreateRevisionResponse>(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: payload,
    })
}

export type VerifyInvoiceResponse = {
    invoice: unknown
}

export type InvoiceDownloadFormat = 'pdf' | 'docx'
export type InvoiceDownloadMode = 'save' | 'generate'

export async function verifyInvoiceHandler(
    clientId: number,
    baseNumber: number,
    invoPayload: unknown,
    options?: { signal?: AbortSignal },
) {
    const url = `/api/clients/${clientId}/invoice/${baseNumber}/verify`
    const payload = JSON.stringify(invoPayload)

    return await request<VerifyInvoiceResponse>(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: payload,
        signal: options?.signal,
    })
}

function downloadBlob(blob: Blob, filename: string) {
    const blobUrl = URL.createObjectURL(blob)

    try {
        const a = document.createElement('a')
        a.href = blobUrl
        a.download = filename
        a.style.display = 'none'

        document.body.appendChild(a)
        a.click()
        document.body.removeChild(a)
    } finally {
        // slightly safer than immediate revoke in some browsers
        window.setTimeout(() => URL.revokeObjectURL(blobUrl), 1000)
    }
}

function filenameFromContentDisposition(header: string | null): string | null {
    if (!header) return null

    const utf8Match = header.match(/filename\*\s*=\s*UTF-8''([^;]+)/i)
    if (utf8Match?.[1]) {
        const encoded = utf8Match[1].trim()
        try {
            return decodeURIComponent(encoded)
        } catch {
            return encoded
        }
    }

    const plainMatch =
        header.match(/filename\s*=\s*"([^"]+)"/i) ?? header.match(/filename\s*=\s*([^;]+)/i)
    if (!plainMatch?.[1]) return null
    return plainMatch[1].trim()
}

function fallbackInvoiceFilename(
    baseNumber: number,
    revisionNumber: number,
    format: InvoiceDownloadFormat,
): string {
    if (baseNumber < 1) return `Invoice.${format}`

    const displayRevisionNo = toDisplayRevisionNo(revisionNumber)
    if (displayRevisionNo == null) return `Invoice-${baseNumber}.${format}`
    return `Invoice-${baseNumber}-Rev-${displayRevisionNo}.${format}`
}

function fallbackReceiptFilename(
    prefix: string,
    baseNumber: number,
    receiptNo: number,
    format: InvoiceDownloadFormat,
): string {
    const label = formatPaymentReceiptLabel(prefix, baseNumber, receiptNo)
    if (!label) return `Invoice-Receipt.${format}`
    return `${label}.${format}`
}

async function generateInvoiceDownloadHandler(
    clientId: number,
    baseNumber: number,
    format: InvoiceDownloadFormat,
    downloadMode: InvoiceDownloadMode = 'save',
    revisionNumber: number = 1,
    invoicePayload?: unknown,
) {
    let url = `/api/clients/${clientId}/invoice/${baseNumber}/${revisionNumber}/${format}`
    let init: RequestInit | undefined

    if (downloadMode === 'generate') {
        url = `/api/clients/${clientId}/invoice/${baseNumber}/${revisionNumber}/${format}/quick`

        init = {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(invoicePayload),
        }
    }

    let res: Response
    try {
        res = await fetch(url, init)
    } catch (err) {
        const name = (err as any)?.name
        if (name === 'AbortError') {
            throw new NetworkError('Request aborted', err)
        }
        throw new NetworkError('Network request failed', err)
    }

    if (!res.ok) {
        const text = await res.text().catch(() => '')
        throw parseApiError(res.status, text)
    }

    const blob = await res.blob()
    const serverFilename = filenameFromContentDisposition(res.headers.get('content-disposition'))
    downloadBlob(
        blob,
        serverFilename ?? fallbackInvoiceFilename(baseNumber, revisionNumber, format),
    )
}

export async function generatePdfHandler(
    clientId: number,
    baseNumber: number,
    downloadMode: InvoiceDownloadMode = 'save',
    revisionNumber: number = 1,
    invoicePayload?: unknown,
) {
    return generateInvoiceDownloadHandler(
        clientId,
        baseNumber,
        'pdf',
        downloadMode,
        revisionNumber,
        invoicePayload,
    )
}

export async function generateDocxHandler(
    clientId: number,
    baseNumber: number,
    downloadMode: InvoiceDownloadMode = 'save',
    revisionNumber: number = 1,
    invoicePayload?: unknown,
) {
    return generateInvoiceDownloadHandler(
        clientId,
        baseNumber,
        'docx',
        downloadMode,
        revisionNumber,
        invoicePayload,
    )
}

export async function createPaymentReceiptHandler(
    clientId: number,
    baseNumber: number,
    payload: {
        amountMinor: number
        paymentDate: string
        label?: string
    },
) {
    return await request<CreatePaymentReceiptResponse>(
        `/api/clients/${clientId}/invoice/${baseNumber}/receipts`,
        {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
        },
    )
}

export async function updatePaymentReceiptHandler(
    clientId: number,
    baseNumber: number,
    receiptNo: number,
    payload: {
        paymentDate: string
        label?: string
    },
) {
    return await request<CreatePaymentReceiptResponse>(
        `/api/clients/${clientId}/invoice/${baseNumber}/receipts/${receiptNo}`,
        {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
        },
    )
}

export async function generatePaymentReceiptDownloadHandler(
    clientId: number,
    baseNumber: number,
    receiptNo: number,
    format: InvoiceDownloadFormat,
    prefix = 'INV',
) {
    const url = `/api/clients/${clientId}/invoice/${baseNumber}/receipts/${receiptNo}/${format}`

    let res: Response
    try {
        res = await fetch(url)
    } catch (err) {
        const name = (err as any)?.name
        if (name === 'AbortError') {
            throw new NetworkError('Request aborted', err)
        }
        throw new NetworkError('Network request failed', err)
    }

    if (!res.ok) {
        const text = await res.text().catch(() => '')
        throw parseApiError(res.status, text)
    }

    const blob = await res.blob()
    const serverFilename = filenameFromContentDisposition(res.headers.get('content-disposition'))
    downloadBlob(blob, serverFilename ?? fallbackReceiptFilename(prefix, baseNumber, receiptNo, format))
}

import type { Invoice } from '@/components/invoice/invoiceTypes'
import { NetworkError, request } from './fetchHelper'
import { parseApiError } from './apiErrors'

export async function getNewInvoiceNumber(clientId: number): Promise<number> {
    const n = await request<number>(`api/clients/${clientId}/invoice`)
    const out = typeof n === 'number' && Number.isFinite(n) ? Math.round(n) : 0
    return out > 0 ? out : 0
}

export async function newInvoiceHandler(
    clientId: number,
    baseNumber: number,
    invoPayload: unknown,
) {
    const url = `api/clients/${clientId}/invoice/${baseNumber}`
    const payload = JSON.stringify(invoPayload)

    return await request<Invoice>(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: payload,
    })
}

export type VerifyInvoiceResponse = {
    invoice: unknown
}

export async function verifyInvoiceHandler(
    clientId: number,
    baseNumber: number,
    invoPayload: unknown,
    options?: { signal?: AbortSignal },
) {
    const url = `api/clients/${clientId}/invoice/${baseNumber}/verify`
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


export async function generatePdfHandler(
    clientId: number,
    baseNumber: number,
    pdfFetch: 'save' | 'generate' = 'save',
    revisionNumber: number = 1,
    invoicePayload?: unknown,
) {
    let url = `/api/clients/${clientId}/invoice/${baseNumber}/${revisionNumber}/pdf`
    let init: RequestInit | undefined

    if (pdfFetch === 'generate') {
        url = `/api/clients/${clientId}/invoice/${baseNumber}/${revisionNumber}/pdf/quick`

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
    downloadBlob(blob, `invoice-${baseNumber}-rev-${revisionNumber}.pdf`)
}

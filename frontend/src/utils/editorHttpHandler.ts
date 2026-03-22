import { request } from './fetchHelper'
import type { InvoiceBookResponse, InvoiceResponse } from '@/components/editor/invBookTypes'

export async function getInvAndRevNums(
    clientId: number,
    limit: number,
    offset: number,
): Promise<InvoiceBookResponse> {
    const url = `/api/clients/${clientId}/edits?limit=${limit}&offset=${offset}`
    return await request<InvoiceBookResponse>(url)
}

export async function getInvoice(
    clientId: number,
    baseNumber: number,
    revisionNumber: number,
): Promise<InvoiceResponse> {
    const url = `/api/clients/${clientId}/edits/get/${baseNumber}/${revisionNumber}`
    return await request<InvoiceResponse>(url)
}

export async function patchInvoiceStatus(
    clientId: number,
    baseNumber: number,
    status: string,
): Promise<{ status: string }> {
    const url = `/api/clients/${clientId}/invoice/${baseNumber}/status`
    return await request<{ status: string }>(url, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status }),
    })
}

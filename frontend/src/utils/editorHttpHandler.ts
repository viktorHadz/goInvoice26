import { request } from './fetchHelper'
import type { InvoiceBookResponse, InvoiceResponse } from '@/components/editor/invBookTypes'
import type { InvoiceBookFilters } from '@/components/editor/invoiceBookFilters'

export async function getInvAndRevNums(
    limit: number,
    offset: number,
    filters?: InvoiceBookFilters,
    clientId?: number | null,
): Promise<InvoiceBookResponse> {
    const params = new URLSearchParams({
        limit: `${limit}`,
        offset: `${offset}`,
    })

    if (filters) {
        params.set('sortBy', filters.sortBy)
        params.set('sortDirection', filters.sortDirection)
        params.set('paymentState', filters.paymentState)
    }

    if (clientId != null && clientId > 0) {
        params.set('clientId', `${clientId}`)
    }

    const url = `/api/edits?${params.toString()}`
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

export async function getPaymentReceipt(
    clientId: number,
    baseNumber: number,
    receiptNo: number,
): Promise<InvoiceResponse> {
    const url = `/api/clients/${clientId}/edits/get/${baseNumber}/receipts/${receiptNo}`
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

export async function deleteInvoice(clientId: number, baseNumber: number): Promise<void> {
    const url = `/api/clients/${clientId}/invoice/${baseNumber}`
    return await request<void>(url, {
        method: 'DELETE',
    })
}

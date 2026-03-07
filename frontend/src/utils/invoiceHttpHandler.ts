import type { Invoice } from '@/components/invoice/invoiceTypes'
import { request } from './fetchHelper'

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

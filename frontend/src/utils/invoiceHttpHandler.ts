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
    invoPayload: Request,
) {
    console.log(clientId, baseNumber, invoPayload)
    const url = `api/clients/${clientId}/invoice/${baseNumber}`
    return await request<Invoice>(url, { method: 'POST', body: JSON.stringify(invoPayload) })
}

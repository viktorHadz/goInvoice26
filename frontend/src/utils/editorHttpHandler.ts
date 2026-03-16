import type { InvoiceBookResponse } from '@/components/editor/editorTypes'
import { request } from './fetchHelper'

export async function getInvAndRevNums(
    clientId: number,
    limit: number,
    offset: number,
): Promise<InvoiceBookResponse> {
    const url = `/api/clients/${clientId}/edits?limit=${limit}&offset=${offset}`
    return await request<InvoiceBookResponse>(url)
}

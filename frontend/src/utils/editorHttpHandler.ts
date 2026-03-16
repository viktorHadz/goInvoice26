import type { InvoiceBookIn } from '@/components/editor/editorTypes'
import { request } from './fetchHelper'

export async function getInvAndRevNums(clientId: number) {
    const url = `/api/clients/${clientId}/edits`
    const data = await request<InvoiceBookIn | null>(url)
    return Array.isArray(data?.items) ? data.items : []
}

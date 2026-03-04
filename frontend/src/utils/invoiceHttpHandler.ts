import { request } from './fetchHelper'

const base = (clientId: number | null) => `api/clients/${clientId}/invoice`

export async function getNewInvoiceNumber(clientId: number): Promise<number> {
    const bNum = await request<number>(base(clientId))

    if (!Number.isFinite(bNum) || bNum <= 0) {
        throw new Error('Invalid invoice number returned from server')
    }

    return bNum
}

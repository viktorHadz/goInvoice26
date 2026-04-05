import { request } from './fetchHelper'
import type { ProductImportKind, ProductImportResult } from './productImport'

export type ProductType = 'style' | 'sample'
export type PricingMode = 'flat' | 'hourly'

export interface Product {
    id: number
    productType: ProductType
    pricingMode: PricingMode
    productName: string
    flatPriceMinor?: number
    hourlyRateMinor?: number
    minutesWorked?: number
    clientId: number
    created_at: string
    updated_at?: string
}

export type ProductUpsert = {
    productType: ProductType
    pricingMode: PricingMode
    productName: string
    flatPrice?: number
    hourlyRate?: number
    minutesWorked?: number
}

const base = (clientId: number) => `/api/clients/${clientId}/products`

export async function listClientProducts(clientId: number): Promise<Product[]> {
    const data = await request<Product[] | null>(base(clientId))
    return Array.isArray(data) ? data : []
}

export function createProduct(clientId: number, payload: ProductUpsert) {
    return request<Product>(base(clientId), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
    })
}

export function updateProduct(clientId: number, productId: number, payload: ProductUpsert) {
    return request<Product>(`${base(clientId)}/${productId}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
    })
}

export function deleteProduct(clientId: number, productId: number) {
    return request<void>(`${base(clientId)}/${productId}`, {
        method: 'DELETE',
    })
}

export function importProducts(clientId: number, kind: ProductImportKind, file: File) {
    const formData = new FormData()
    formData.append('kind', kind)
    formData.append('file', file)

    return request<ProductImportResult>(`${base(clientId)}/import`, {
        method: 'POST',
        body: formData,
    })
}

export type InvoiceStatus = 'draft' | 'issued' | 'paid' | 'void'
export type DiscountType = 'none' | 'percent' | 'fixed'
export type LineType = 'style' | 'sample' | 'custom'
export type PricingMode = 'flat' | 'hourly'

export type MoneyMinor = number // integer minor units

export type InvoiceLine = {
    id?: number
    productId?: number | null

    name: string
    lineType: LineType
    pricingMode: PricingMode

    quantity: number
    unitPriceMinor: MoneyMinor
    minutesWorked?: number | null

    sortOrder: number
}

export type InvoiceDraft = {
    invoiceId?: number
    baseNumber?: number
    status?: InvoiceStatus

    clientId: number

    issueDate: string
    dueByDate: string

    clientSnapshot: {
        name: string
        companyName: string
        address: string
        email: string
    }

    note: string

    vatRate: number // 2000 => 20.00%
    discountType: DiscountType
    // fixed => minor units; percent => 0..10000 (basis points percent)
    discountValue: number

    lines: InvoiceLine[]

    paidMinor: MoneyMinor
    depositMinor: MoneyMinor
}

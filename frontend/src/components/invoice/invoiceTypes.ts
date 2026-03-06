export type InvoiceStatus = 'draft' | 'issued' | 'paid' | 'void'
export type DiscountType = 'none' | 'percent' | 'fixed'
export type DepositType = 'none' | 'fixed' | 'percent'
export type LineType = 'style' | 'sample' | 'custom'
export type PricingMode = 'flat' | 'hourly'

export type MoneyMinor = number // integer minor units (pence)

export type Totals = {
    subtotalMinor: MoneyMinor
    discountMinor: MoneyMinor
    subtotalAfterDiscountMinor: MoneyMinor
    vatMinor: MoneyMinor
    totalMinor: MoneyMinor
}

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

export type Invoice = {
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

    lines: InvoiceLine[]

    discountType: DiscountType
    // fixed price => minor units | percent => 0..10000 (basis points)
    discountValue: number

    vatRate: number // 2000 => 20.00%

    paidMinor: MoneyMinor

    depositType: DepositType
    // fixed price => minor units | percent => 0..10000 (basis points)
    depositValue: number

    note: string
}

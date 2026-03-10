export type DiscountType = 'none' | 'percent' | 'fixed'
export type DepositType = 'none' | 'percent' | 'fixed'
export type LineType = 'style' | 'sample' | 'custom'
export type PricingMode = 'flat' | 'hourly'

export type MoneyMinor = number // integer minor units (pence)
export type BasisPoints = number // 1000 = 10%, 10000 = 100%

export type Totals = {
    subtotalMinor: MoneyMinor
    discountMinor: MoneyMinor
    subtotalAfterDiscountMinor: MoneyMinor
    vatMinor: MoneyMinor
    totalMinor: MoneyMinor
}

export type InvoiceLine = {
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
    baseNumber: number
    clientId: number

    issueDate: string
    dueByDate?: string

    clientSnapshot: {
        name: string
        companyName: string
        address: string
        email: string
    }

    lines: InvoiceLine[]

    discountType: DiscountType
    discountMinor: MoneyMinor
    discountRate: BasisPoints

    vatRate: BasisPoints

    paidMinor: MoneyMinor

    depositType: DepositType
    depositMinor: MoneyMinor
    depositRate: BasisPoints

    note?: string
}

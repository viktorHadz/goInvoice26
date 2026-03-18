export type InvBookRevision = {
    id: number
    revisionNo: number
    issueDate: string
    dueByDate?: string
    updatedAt?: string
}

export type InvBookInvoice = {
    id: number
    baseNo: number
    status: string
    revisions: InvBookRevision[]
}

export type ActiveEditorNode =
    | {
          type: 'invoice'
          id: number
          baseNo: number
      }
    | {
          type: 'revision'
          id: number
          invoiceId: number
          baseNo: number
          revisionNo: number
      }
    | null

export type InvoiceBookResponse = {
    items: InvBookInvoice[]
    limit: number
    offset: number
    count: number
    total: number
    hasMore: boolean
}

export type InvoiceResponse = {
    totals: {
        baseNumber: number
        revisionNo: number
        issueDate: string
        dueByDate?: string
        clientName: string
        clientCompanyName: string
        clientAddress: string
        clientEmail: string
        note?: string
        vatRate: number
        vatAmountMinor: number
        discountType: string
        discountRate: number
        discountMinor: number
        depositType: string
        depositRate: number
        depositMinor: number
        subtotalMinor: number
        totalMinor: number
        paidMinor: number
    }
    lines: {
        productId?: number | null
        pricingMode?: string | null
        minutesWorked?: number | null
        name: string
        lineType: string
        quantity: number
        unitPriceMinor: number
        lineTotalMinor: number
        sortOrder: number
    }[]
}

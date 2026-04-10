export type InvBookRevision = {
    id: number
    revisionNo: number
    issueDate: string
    dueByDate?: string
    updatedAt?: string
}

export type InvBookHistoryItem = {
    id: number
    type: 'revision' | 'payment_receipt'
    createdAt: string
    revisionNo?: number
    receiptNo?: number
    issueDate?: string
    dueByDate?: string
    paymentDate?: string
    amountMinor?: number
    label?: string
}

export type InvBookInvoice = {
    id: number
    clientId: number
    clientName: string
    clientCompanyName: string
    baseNo: number
    status: string
    latestRevisionNo: number
    issueDate: string
    dueByDate?: string
    totalMinor: number
    depositMinor: number
    paidMinor: number
    balanceDueMinor: number
    revisions: InvBookRevision[]
    history: InvBookHistoryItem[]
}

export type ActiveEditorNode =
    | {
          type: 'invoice'
          clientId: number
          id: number
          baseNo: number
      }
    | {
          type: 'revision'
          clientId: number
          id: number
          invoiceId: number
          baseNo: number
          revisionNo: number
      }
    | {
          type: 'paymentReceipt'
          clientId: number
          id: number
          invoiceId: number
          baseNo: number
          receiptNo: number
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
    status: string
    totals: {
        baseNumber: number
        revisionNo: number
        issueDate: string
        supplyDate?: string
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
    payments: {
        id: number
        amountMinor: number
        paymentDate: string
        paymentType: string
        label?: string
    }[]
    history: InvBookHistoryItem[]
    selectedReceipt?: {
        id: number
        receiptNo: number
        paymentDate: string
        amountMinor: number
        label?: string
        appliedRevisionNo: number
        createdAt: string
    }
}

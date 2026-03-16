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
    | { type: 'invoice'; id: number }
    | { type: 'revision'; id: number }
    | null

export type InvoiceBookResponse = {
    items: InvBookInvoice[]
    limit: number
    offset: number
    count: number
    total: number
    hasMore: boolean
}

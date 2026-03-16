export type InvoiceRevisionNode = {
    id: number
    revisionNo: number
    issueDate: string
    dueByDate?: string
    updatedAt?: string
}

export type InvoiceTreeNode = {
    id: number
    baseNo: number
    status: string
    revisions: InvoiceRevisionNode[]
}

export type ActiveEditorNode =
    | { type: 'invoice'; id: number }
    | { type: 'revision'; id: number }
    | null

// !---
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
export type InvoiceBookIn = {
    items: InvBookInvoice[]
}

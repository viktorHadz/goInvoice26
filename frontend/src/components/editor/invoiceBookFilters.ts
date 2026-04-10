import {
    formatInvoiceBaseLabel,
    formatInvoiceDisplayLabel,
} from '@/utils/invoiceLabels'
import type { InvBookInvoice } from './invBookTypes'

export type InvoiceBookSortBy = 'date' | 'balance'
export type InvoiceBookSortDirection = 'asc' | 'desc'
export type InvoiceBookPaymentState = 'all' | 'paid' | 'unpaid'

export type InvoiceBookFilters = {
    sortBy: InvoiceBookSortBy
    sortDirection: InvoiceBookSortDirection
    paymentState: InvoiceBookPaymentState
    activeClientOnly: boolean
}

const defaultSortDirections: Record<InvoiceBookSortBy, InvoiceBookSortDirection> = {
    date: 'desc',
    balance: 'desc',
}

const paymentStateCycle: InvoiceBookPaymentState[] = ['all', 'unpaid', 'paid']

export function createDefaultInvoiceBookFilters(): InvoiceBookFilters {
    return {
        sortBy: 'date',
        sortDirection: defaultSortDirections.date,
        paymentState: 'all',
        activeClientOnly: false,
    }
}

export function areInvoiceBookFiltersEqual(a: InvoiceBookFilters, b: InvoiceBookFilters): boolean {
    return (
        a.sortBy === b.sortBy &&
        a.sortDirection === b.sortDirection &&
        a.paymentState === b.paymentState &&
        a.activeClientOnly === b.activeClientOnly
    )
}

export function cycleInvoiceBookSort(
    filters: InvoiceBookFilters,
    sortBy: InvoiceBookSortBy,
): InvoiceBookFilters {
    if (filters.sortBy !== sortBy) {
        return {
            ...filters,
            sortBy,
            sortDirection: defaultSortDirections[sortBy],
        }
    }

    return {
        ...filters,
        sortDirection: filters.sortDirection === 'desc' ? 'asc' : 'desc',
    }
}

export function cycleInvoiceBookPaymentState(filters: InvoiceBookFilters): InvoiceBookFilters {
    const currentIndex = paymentStateCycle.indexOf(filters.paymentState)
    const nextIndex = (currentIndex + 1) % paymentStateCycle.length
    const nextPaymentState = paymentStateCycle[nextIndex] ?? 'all'

    return {
        ...filters,
        paymentState: nextPaymentState,
    }
}

export function toggleInvoiceBookActiveClient(filters: InvoiceBookFilters): InvoiceBookFilters {
    return {
        ...filters,
        activeClientOnly: !filters.activeClientOnly,
    }
}

export function isDefaultInvoiceBookFilters(filters: InvoiceBookFilters): boolean {
    return areInvoiceBookFiltersEqual(filters, createDefaultInvoiceBookFilters())
}

export function invoiceBookSortSummary(filters: InvoiceBookFilters): string {
    if (filters.sortBy === 'balance') {
        return filters.sortDirection === 'desc'
            ? 'Outstanding: high to low'
            : 'Outstanding: low to high'
    }

    return filters.sortDirection === 'desc' ? 'Date: newest first' : 'Date: oldest first'
}

export function invoiceBookPaymentSummary(paymentState: InvoiceBookPaymentState): string {
    switch (paymentState) {
        case 'paid':
            return 'Payment: paid only'
        case 'unpaid':
            return 'Payment: unpaid only'
        default:
            return 'Payment: all'
    }
}

export function invoiceBookClientSummary(
    activeClientOnly: boolean,
    activeClientName?: string | null,
): string {
    if (!activeClientOnly) return 'Client: all clients'
    if (activeClientName?.trim()) return `Client: ${activeClientName.trim()}`
    return 'Client: active only'
}

export function filterInvoiceBookByQuery(
    invoices: InvBookInvoice[],
    rawQuery: string,
    prefix: string,
): InvBookInvoice[] {
    const query = rawQuery.trim().toLowerCase()
    if (!query) return invoices

    return invoices
        .map((invoice) => {
            const invoiceLabel = formatInvoiceBaseLabel(prefix, invoice.baseNo).toLowerCase()
            const clientLabel = [invoice.clientCompanyName, invoice.clientName]
                .filter(Boolean)
                .join(' ')
                .toLowerCase()

            if (invoiceLabel.includes(query) || clientLabel.includes(query)) return invoice

            const matchingRevisions = invoice.revisions.filter((entry) =>
                formatInvoiceDisplayLabel(prefix, invoice.baseNo, entry.revisionNo)
                    .toLowerCase()
                    .includes(query),
            )

            if (matchingRevisions.length === 0) return null

            return {
                ...invoice,
                revisions: matchingRevisions,
            }
        })
        .filter((invoice): invoice is InvBookInvoice => invoice !== null)
}

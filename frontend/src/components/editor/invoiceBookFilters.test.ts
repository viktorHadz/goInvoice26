import { describe, expect, it } from 'vitest'
import {
    createDefaultInvoiceBookFilters,
    cycleInvoiceBookPaymentState,
    cycleInvoiceBookSort,
    filterInvoiceBookByQuery,
    invoiceBookClientSummary,
    invoiceBookPaymentSummary,
    invoiceBookSortSummary,
    isDefaultInvoiceBookFilters,
    toggleInvoiceBookActiveClient,
} from './invoiceBookFilters'
import type { InvBookInvoice } from './invBookTypes'

function makeInvoices(): InvBookInvoice[] {
    return [
        {
            id: 1,
            clientId: 99,
            clientName: 'Alex Carter',
            clientCompanyName: 'Acme Co',
            baseNo: 101,
            status: 'issued',
            latestRevisionNo: 2,
            issueDate: '2026-03-21',
            dueByDate: '2026-03-31',
            totalMinor: 12000,
            depositMinor: 0,
            paidMinor: 2000,
            balanceDueMinor: 10000,
            revisions: [
                { id: 10, revisionNo: 1, issueDate: '2026-03-20', dueByDate: '2026-03-30' },
                { id: 11, revisionNo: 2, issueDate: '2026-03-21', dueByDate: '2026-03-31' },
            ],
        },
        {
            id: 2,
            clientId: 100,
            clientName: 'Mia Stone',
            clientCompanyName: 'North Studio',
            baseNo: 102,
            status: 'paid',
            latestRevisionNo: 1,
            issueDate: '2026-03-22',
            dueByDate: '2026-04-01',
            totalMinor: 8000,
            depositMinor: 0,
            paidMinor: 8000,
            balanceDueMinor: 0,
            revisions: [
                { id: 20, revisionNo: 1, issueDate: '2026-03-22', dueByDate: '2026-04-01' },
            ],
        },
    ]
}

describe('invoiceBookFilters', () => {
    it('uses date newest-first as the default view', () => {
        const filters = createDefaultInvoiceBookFilters()

        expect(filters).toEqual({
            sortBy: 'date',
            sortDirection: 'desc',
            paymentState: 'all',
            activeClientOnly: false,
        })
        expect(isDefaultInvoiceBookFilters(filters)).toBe(true)
    })

    it('cycles sort direction for the active sort and resets direction for new sorts', () => {
        const defaults = createDefaultInvoiceBookFilters()

        expect(cycleInvoiceBookSort(defaults, 'date')).toEqual({
            sortBy: 'date',
            sortDirection: 'asc',
            paymentState: 'all',
            activeClientOnly: false,
        })

        expect(cycleInvoiceBookSort(defaults, 'balance')).toEqual({
            sortBy: 'balance',
            sortDirection: 'desc',
            paymentState: 'all',
            activeClientOnly: false,
        })
    })

    it('cycles payment state through all, unpaid, and paid', () => {
        const defaults = createDefaultInvoiceBookFilters()
        const unpaid = cycleInvoiceBookPaymentState(defaults)
        const paid = cycleInvoiceBookPaymentState(unpaid)
        const all = cycleInvoiceBookPaymentState(paid)

        expect(unpaid.paymentState).toBe('unpaid')
        expect(paid.paymentState).toBe('paid')
        expect(all.paymentState).toBe('all')
    })

    it('toggles the active-client filter flag', () => {
        expect(
            toggleInvoiceBookActiveClient(createDefaultInvoiceBookFilters()).activeClientOnly,
        ).toBe(true)
    })

    it('filters invoice rows by invoice and revision labels without mutating the originals', () => {
        const invoices = makeInvoices()

        expect(filterInvoiceBookByQuery(invoices, 'INV - 101', 'INV')).toHaveLength(1)
        expect(filterInvoiceBookByQuery(invoices, 'Acme', 'INV')).toHaveLength(1)

        const revisionMatch = filterInvoiceBookByQuery(invoices, '101.1', 'INV')
        expect(revisionMatch).toHaveLength(1)
        const [matchedInvoice] = revisionMatch
        expect(matchedInvoice).toBeDefined()
        expect(matchedInvoice?.revisions).toHaveLength(1)
        expect(matchedInvoice?.revisions[0]?.revisionNo).toBe(2)

        expect(invoices[0]?.revisions).toHaveLength(2)
    })

    it('formats compact labels for the active sort and payment filter', () => {
        expect(invoiceBookSortSummary(createDefaultInvoiceBookFilters())).toBe('Date: newest first')
        expect(invoiceBookPaymentSummary('unpaid')).toBe('Payment: unpaid only')
        expect(invoiceBookClientSummary(true, 'Acme Co')).toBe('Client: Acme Co')
    })
})

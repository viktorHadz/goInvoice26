import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import type { Invoice } from '@/components/invoice/invoiceTypes'
import { useEditorStore } from '@/stores/editor'

const {
    selectClientByIdMock,
    deleteInvoiceMock,
    getInvAndRevNumsMock,
    getInvoiceMock,
    newRevisionHandlerMock,
    createPaymentReceiptHandlerMock,
    deletePaymentReceiptHandlerMock,
    updateDraftInvoiceHandlerMock,
    updatePaymentReceiptHandlerMock,
    emitToastErrorMock,
    emitToastInfoMock,
    emitToastSuccessMock,
} = vi.hoisted(() => ({
    selectClientByIdMock: vi.fn(),
    deleteInvoiceMock: vi.fn(async () => undefined),
    getInvAndRevNumsMock: vi.fn(async () => ({
        items: [],
        total: 0,
        hasMore: false,
        limit: 10,
        offset: 0,
    })),
    getInvoiceMock: vi.fn(async () => ({
        status: 'issued',
        totals: {
            baseNumber: 101,
            revisionNo: 2,
            issueDate: '2026-03-20',
            dueByDate: '2026-03-30',
            clientName: 'Alex',
            clientCompanyName: 'Acme Co',
            clientAddress: '1 Test Road',
            clientEmail: 'alex@example.com',
            note: 'Saved revision',
            vatRate: 0,
            vatAmountMinor: 0,
            discountType: 'none',
            discountRate: 0,
            discountMinor: 0,
            depositType: 'none',
            depositRate: 0,
            depositMinor: 0,
            subtotalMinor: 10000,
            totalMinor: 10000,
            paidMinor: 0,
        },
        lines: [
            {
                productId: 1,
                pricingMode: 'flat',
                minutesWorked: null,
                name: 'Service line',
                lineType: 'custom',
                quantity: 1,
                unitPriceMinor: 10000,
                lineTotalMinor: 10000,
                sortOrder: 1,
            },
        ],
        receipts: [],
    })),
    newRevisionHandlerMock: vi.fn(async () => ({
        invoiceId: 7,
        revisionId: 70,
        revisionNo: 2,
    })),
    createPaymentReceiptHandlerMock: vi.fn(),
    deletePaymentReceiptHandlerMock: vi.fn(async () => undefined),
    updateDraftInvoiceHandlerMock: vi.fn(async () => ({
        invoiceId: 7,
        revisionId: 7,
    })),
    updatePaymentReceiptHandlerMock: vi.fn(),
    emitToastErrorMock: vi.fn(),
    emitToastInfoMock: vi.fn(),
    emitToastSuccessMock: vi.fn(),
}))

vi.mock('@/stores/clients', () => ({
    useClientStore: () => ({
        selectedClient: { id: 42 },
        selectClientById: selectClientByIdMock,
    }),
}))

vi.mock('@/stores/settings', () => ({
    useSettingsStore: () => ({
        settings: {
            invoicePrefix: 'INV',
            showItemTypeHeaders: true,
            startingInvoiceNumber: 1,
            canEditStartingInvoiceNumber: true,
        },
    }),
}))

vi.mock('@/utils/editorHttpHandler', () => ({
    deleteInvoice: deleteInvoiceMock,
    getInvAndRevNums: getInvAndRevNumsMock,
    getInvoice: getInvoiceMock,
    patchInvoiceStatus: vi.fn(),
}))

vi.mock('@/utils/invoiceHttpHandler', () => ({
    createPaymentReceiptHandler: createPaymentReceiptHandlerMock,
    deletePaymentReceiptHandler: deletePaymentReceiptHandlerMock,
    newRevisionHandler: newRevisionHandlerMock,
    updateDraftInvoiceHandler: updateDraftInvoiceHandlerMock,
    updatePaymentReceiptHandler: updatePaymentReceiptHandlerMock,
}))

vi.mock('@/utils/frontendValidation', () => ({
    validateInvoicePayload: vi.fn(() => ({})),
}))

vi.mock('@/composables/useInvoiceVerification', () => ({
    useInvoiceVerification: () => ({
        verifyStatus: { value: 'idle' },
        lastVerifyAt: { value: null },
        serverCanonicalTotals: { value: null },
        serverCanonicalLineTotals: { value: [] },
        runServerVerify: vi.fn(),
        scheduleServerVerify: vi.fn(),
        clearVerifyState: vi.fn(),
    }),
}))

vi.mock('@/composables/useInvoicePricing', () => ({
    useInvoicePricing: () => ({
        pricing: { value: null },
        totals: { value: null },
        depositMinor: { value: 0 },
        balanceDueMinor: { value: 0 },
    }),
}))

vi.mock('@/composables/useInvoiceFieldErrors', () => ({
    useInvoiceFieldErrors: () => ({
        liveFieldErrors: {},
        getFieldError: vi.fn(() => ''),
    }),
}))

vi.mock('@/utils/toast', () => ({
    emitToastError: emitToastErrorMock,
    emitToastInfo: emitToastInfoMock,
    emitToastSuccess: emitToastSuccessMock,
}))

function makeInvoice(): Invoice {
    return {
        baseNumber: 101,
        clientId: 42,
        status: 'issued',
        issueDate: '2026-03-20',
        dueByDate: '2026-03-30',
        clientSnapshot: {
            name: 'Alex',
            companyName: 'Acme Co',
            address: '1 Test Road',
            email: 'alex@example.com',
        },
        lines: [
            {
                productId: 1,
                name: 'Service line',
                lineType: 'custom',
                pricingMode: 'flat',
                quantity: 1,
                unitPriceMinor: 10000,
                minutesWorked: null,
                sortOrder: 1,
            },
        ],
        discountType: 'none',
        discountMinor: 0,
        discountRate: 0,
        vatRate: 0,
        paidMinor: 0,
        depositType: 'none',
        depositMinor: 0,
        depositRate: 0,
        note: 'Initial',
    }
}

function makeBookInvoice(status: 'draft' | 'issued' | 'paid' | 'void' = 'issued') {
    return {
        id: 7,
        clientId: 42,
        clientName: 'Alex',
        clientCompanyName: 'Acme Co',
        baseNo: 101,
        status,
        latestRevisionNo: 1,
        issueDate: '2026-03-20',
        dueByDate: '2026-03-30',
        totalMinor: 10000,
        depositMinor: 0,
        paidMinor: 0,
        balanceDueMinor: 10000,
        revisions: [],
    }
}

describe('editor store no-change save guard', () => {
    beforeEach(() => {
        setActivePinia(createPinia())
        vi.clearAllMocks()
    })

    it('skips save and shows info toast when payload is unchanged', async () => {
        const store = useEditorStore()
        store.activeInvoice = makeInvoice()
        store.initEdit()

        const result = await store.saveRevision(store.draftInvoice)

        expect(result).toBe(false)
        expect(newRevisionHandlerMock).not.toHaveBeenCalled()
        expect(emitToastInfoMock).toHaveBeenCalledWith('No changes to save.')
        expect(emitToastErrorMock).not.toHaveBeenCalled()
    })

    it('saves revision when payload changes', async () => {
        const store = useEditorStore()
        store.activeNode = { type: 'invoice', clientId: 42, id: 7, baseNo: 101 }
        store.invoiceBook = [makeBookInvoice('issued')]
        store.activeInvoice = makeInvoice()
        store.initEdit()

        store.setNote('Changed note')
        const result = await store.saveRevision(store.draftInvoice)

        expect(result).toBe(true)
        expect(newRevisionHandlerMock).toHaveBeenCalledTimes(1)
        expect(store.activeNode).toEqual({
            type: 'revision',
            clientId: 42,
            id: 70,
            invoiceId: 7,
            baseNo: 101,
            revisionNo: 2,
        })
        expect(emitToastSuccessMock).toHaveBeenCalledTimes(1)
    })

    it('updates draft invoices in place and keeps invoice selection', async () => {
        const store = useEditorStore()
        store.activeNode = { type: 'invoice', clientId: 42, id: 7, baseNo: 101 }
        store.invoiceBook = [makeBookInvoice('draft')]
        store.activeInvoice = { ...makeInvoice(), status: 'draft' }
        getInvoiceMock.mockResolvedValueOnce({
            status: 'draft',
            totals: {
                baseNumber: 101,
                revisionNo: 1,
                issueDate: '2026-03-20',
                dueByDate: '2026-03-30',
                clientName: 'Alex',
                clientCompanyName: 'Acme Co',
                clientAddress: '1 Test Road',
                clientEmail: 'alex@example.com',
                note: 'Changed draft note',
                vatRate: 0,
                vatAmountMinor: 0,
                discountType: 'none',
                discountRate: 0,
                discountMinor: 0,
                depositType: 'none',
                depositRate: 0,
                depositMinor: 0,
                subtotalMinor: 10000,
                totalMinor: 10000,
                paidMinor: 0,
            },
            lines: [
                {
                    productId: 1,
                    pricingMode: 'flat',
                    minutesWorked: null,
                    name: 'Service line',
                    lineType: 'custom',
                    quantity: 1,
                    unitPriceMinor: 10000,
                    lineTotalMinor: 10000,
                    sortOrder: 1,
                },
            ],
            receipts: [],
        })
        store.initEdit()

        store.setNote('Changed draft note')
        const result = await store.saveRevision(store.draftInvoice)

        expect(result).toBe(true)
        expect(updateDraftInvoiceHandlerMock).toHaveBeenCalledTimes(1)
        expect(newRevisionHandlerMock).not.toHaveBeenCalled()
        expect(store.activeNode).toEqual({ type: 'invoice', clientId: 42, id: 7, baseNo: 101 })
        expect(emitToastSuccessMock).toHaveBeenCalledWith('Draft INV-101 saved.')
    })

    it('resets baseline after cancel and reopen edit session', async () => {
        const store = useEditorStore()
        store.activeInvoice = makeInvoice()
        store.initEdit()

        store.setNote('Unsaved draft note')
        expect(store.hasUnsavedChanges).toBe(true)

        store.cancelEdit()
        store.initEdit()

        const result = await store.saveRevision(store.draftInvoice)
        expect(result).toBe(false)
        expect(newRevisionHandlerMock).not.toHaveBeenCalled()
        expect(emitToastInfoMock).toHaveBeenCalledWith('No changes to save.')
    })

    it('deletes the active invoice and clears the editor state', async () => {
        const store = useEditorStore()
        store.activeNode = { type: 'invoice', clientId: 42, id: 7, baseNo: 101 }
        store.activeInvoice = makeInvoice()
        store.invoiceBook = [makeBookInvoice('issued')]

        const result = await store.deleteActiveInvoice()

        expect(result).toBe(true)
        expect(deleteInvoiceMock).toHaveBeenCalledWith(42, 101)
        expect(store.activeNode).toBe(null)
        expect(store.activeInvoice).toBe(null)
        expect(store.draftInvoice).toBe(null)
        expect(getInvAndRevNumsMock).toHaveBeenCalled()
        expect(emitToastSuccessMock).toHaveBeenCalledWith('INV-101 deleted.')
    })

    it('stores invoice-book filters and refetches with them', async () => {
        const store = useEditorStore()
        getInvAndRevNumsMock.mockClear()

        await store.cycleBookSort('balance')

        expect(store.invoiceBookFilters).toEqual({
            sortBy: 'balance',
            sortDirection: 'desc',
            paymentState: 'all',
            activeClientOnly: false,
        })
        expect(getInvAndRevNumsMock).toHaveBeenLastCalledWith(
            10,
            0,
            {
                sortBy: 'balance',
                sortDirection: 'desc',
                paymentState: 'all',
                activeClientOnly: false,
            },
            null,
        )

        await store.cycleBookPaymentState()

        expect(store.invoiceBookFilters).toEqual({
            sortBy: 'balance',
            sortDirection: 'desc',
            paymentState: 'unpaid',
            activeClientOnly: false,
        })
        expect(getInvAndRevNumsMock).toHaveBeenLastCalledWith(
            10,
            0,
            {
                sortBy: 'balance',
                sortDirection: 'desc',
                paymentState: 'unpaid',
                activeClientOnly: false,
            },
            null,
        )
    })

    it('can scope the invoice book to the active client', async () => {
        const store = useEditorStore()
        getInvAndRevNumsMock.mockClear()

        await store.toggleBookActiveClientOnly()

        expect(store.invoiceBookFilters).toEqual({
            sortBy: 'date',
            sortDirection: 'desc',
            paymentState: 'all',
            activeClientOnly: true,
        })
        expect(getInvAndRevNumsMock).toHaveBeenLastCalledWith(
            10,
            0,
            {
                sortBy: 'date',
                sortDirection: 'desc',
                paymentState: 'all',
                activeClientOnly: true,
            },
            42,
        )
    })

    it('queues a client switch when selecting an invoice from another client', async () => {
        const store = useEditorStore()

        await store.selectInvoiceBookNode({ type: 'invoice', clientId: 99, id: 7, baseNo: 101 })

        expect(selectClientByIdMock).toHaveBeenCalledWith(99)
        expect(store.activeNode).toBe(null)
    })
})

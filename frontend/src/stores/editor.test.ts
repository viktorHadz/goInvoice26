import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import type { Invoice } from '@/components/invoice/invoiceTypes'
import { useEditorStore } from '@/stores/editor'

const {
    deleteInvoiceMock,
    getInvAndRevNumsMock,
    newRevisionHandlerMock,
    emitToastErrorMock,
    emitToastInfoMock,
    emitToastSuccessMock,
} = vi.hoisted(() => ({
    deleteInvoiceMock: vi.fn(async () => undefined),
    getInvAndRevNumsMock: vi.fn(async () => ({
        items: [],
        total: 0,
        hasMore: false,
        limit: 10,
        offset: 0,
    })),
    newRevisionHandlerMock: vi.fn(async () => undefined),
    emitToastErrorMock: vi.fn(),
    emitToastInfoMock: vi.fn(),
    emitToastSuccessMock: vi.fn(),
}))

vi.mock('@/stores/clients', () => ({
    useClientStore: () => ({
        selectedClient: { id: 42 },
    }),
}))

vi.mock('@/stores/settings', () => ({
    useSettingsStore: () => ({
        settings: { invoicePrefix: 'INV', showItemTypeHeaders: true },
    }),
}))

vi.mock('@/utils/editorHttpHandler', () => ({
    deleteInvoice: deleteInvoiceMock,
    getInvAndRevNums: getInvAndRevNumsMock,
    getInvoice: vi.fn(),
    patchInvoiceStatus: vi.fn(),
}))

vi.mock('@/utils/invoiceHttpHandler', () => ({
    newRevisionHandler: newRevisionHandlerMock,
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
        store.activeInvoice = makeInvoice()
        store.initEdit()

        store.setNote('Changed note')
        const result = await store.saveRevision(store.draftInvoice)

        expect(result).toBe(true)
        expect(newRevisionHandlerMock).toHaveBeenCalledTimes(1)
        expect(emitToastSuccessMock).toHaveBeenCalledTimes(1)
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
        store.activeNode = { type: 'invoice', id: 7, baseNo: 101 }
        store.activeInvoice = makeInvoice()
        store.invoiceBook = [{ id: 7, baseNo: 101, status: 'issued', revisions: [] }]

        const result = await store.deleteActiveInvoice()

        expect(result).toBe(true)
        expect(deleteInvoiceMock).toHaveBeenCalledWith(42, 101)
        expect(store.activeNode).toBe(null)
        expect(store.activeInvoice).toBe(null)
        expect(store.draftInvoice).toBe(null)
        expect(getInvAndRevNumsMock).toHaveBeenCalled()
        expect(emitToastSuccessMock).toHaveBeenCalledWith('INV - 101 deleted.')
    })
})

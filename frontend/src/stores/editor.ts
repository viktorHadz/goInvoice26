import { defineStore } from 'pinia'
import { computed, ref, watch, type MaybeRefOrGetter } from 'vue'
import { useClientStore } from './clients'
import { getInvAndRevNums, getInvoice } from '@/utils/editorHttpHandler'
import type { InvBookInvoice, InvoiceResponse } from '@/components/editor/invBookTypes'
import { handleActionError } from '@/utils/errors/handleActionError'
import { getApiErrorMessage, isApiError } from '@/utils/apiErrors'
import type {
    DepositType,
    DiscountType,
    Invoice,
    InvoiceLine,
    LineType,
    PricingMode,
} from '@/components/invoice/invoiceTypes'
import { useInvoiceFieldErrors } from '@/composables/useInvoiceFieldErrors'
import { cloneInvoice } from '@/utils/cloneInvoice'
import {
    addInvoiceLine,
    removeInvoiceLine,
    setInvoiceNote,
    updateInvoiceLine,
} from '@/utils/invoiceMutations'
import { fmtPrettyInvoiceNumber } from '@/utils/numbers'
import { useSettingsStore } from './settings'
import { useInvoiceVerification } from '@/composables/useInvoiceVerification'
import { useInvoicePricing } from '@/composables/useInvoicePricing'

// TODO: Need to fix payments for editor and invoice

export const useEditorStore = defineStore('editorStore', () => {
    const clientStore = useClientStore()
    const setsStore = useSettingsStore()

    const invoicePrefix = computed(() => setsStore.settings?.invoicePrefix ?? '')

    const invoiceBook = ref<InvBookInvoice[]>([])
    const activeInvoice = ref<Invoice | null>(null)
    /**
     * draftInvoice is the editable working copy.
     */
    const draftInvoice = ref<Invoice | null>(null)

    const isEditing = ref(false)

    const limit = ref(10)
    const offset = ref(0)
    const total = ref(0)
    const hasMore = ref(false)

    const isLoadingBook = ref(false)
    const isLoadingInvoice = ref(false)

    const bookError = ref('')
    const serverFieldErrors = ref<Record<string, string>>({})

    const prettyBaseNumber = computed(() => {
        return fmtPrettyInvoiceNumber(invoicePrefix.value, draftInvoice.value?.baseNumber)
    })

    /**
     * showAllValidation - whether validation errors are displayed before attempting to submit.
     */
    const showAllValidation = ref(false)
    const canGoPrev = computed(() => offset.value > 0)
    const canGoNext = computed(() => hasMore.value)
    const { totals, depositMinor, balanceDueMinor } = useInvoicePricing(draftInvoice)

    const {
        verifyStatus,
        lastVerifyAt,
        serverCanonicalTotals,
        serverCanonicalLineTotals,
        runServerVerify,
        scheduleServerVerify,
        clearVerifyState,
        abortVerify,
    } = useInvoiceVerification(
        draftInvoice,
        computed(() => clientStore.selectedClient?.id ?? null),
        totals,
        serverFieldErrors,
    )

    function ensureDraft(): Invoice {
        if (!draftInvoice.value) throw new Error('Invoice not initialised')
        return draftInvoice.value
    }

    let lastBookRequestId = 0
    async function fetchInvoiceBook(reset = false) {
        const clientId = clientStore.selectedClient?.id

        if (!clientId) {
            clearInvoiceBook()
            return
        }

        if (reset) {
            offset.value = 0
        }

        const requestId = ++lastBookRequestId

        isLoadingBook.value = true
        bookError.value = ''

        try {
            const data = await getInvAndRevNums(clientId, limit.value, offset.value)

            if (requestId !== lastBookRequestId) return
            if (clientStore.selectedClient?.id !== clientId) return

            invoiceBook.value = data.items
            total.value = data.total
            hasMore.value = data.hasMore
            limit.value = data.limit
            offset.value = data.offset
        } catch (error) {
            if (requestId !== lastBookRequestId) return

            invoiceBook.value = []
            total.value = 0
            hasMore.value = false

            bookError.value = isApiError(error)
                ? getApiErrorMessage(error)
                : error instanceof Error && error.message.trim().length > 0
                  ? error.message
                  : 'Failed to load invoice book'

            handleActionError(error, {
                toastTitle: 'Failed to load invoice book',
                supportMessage: 'Please contact support',
                mapFields: false,
            })
        } finally {
            if (requestId === lastBookRequestId) {
                isLoadingBook.value = false
            }
        }
    }

    let lastInvoiceRequestId = 0

    async function fetchInvoice(baseNumber: number, revisionNumber: number) {
        const clientId = clientStore.selectedClient?.id
        if (!clientId) {
            clearActiveInvoice()
            return
        }

        const requestId = ++lastInvoiceRequestId
        isLoadingInvoice.value = true

        try {
            const data = await getInvoice(clientId, baseNumber, revisionNumber)

            if (requestId !== lastInvoiceRequestId) return
            if (clientStore.selectedClient?.id !== clientId) return

            const formatted = fmtActive(data, clientId)

            activeInvoice.value = formatted
            draftInvoice.value = cloneInvoice(formatted)

            isEditing.value = false
            serverFieldErrors.value = {}
            showAllValidation.value = false
            clearVerifyState()
        } catch (error) {
            if (requestId !== lastInvoiceRequestId) return

            clearActiveInvoice()

            handleActionError(error, {
                toastTitle: 'Failed to fetch invoice',
                supportMessage: 'Please contact support',
                mapFields: false,
            })
        } finally {
            if (requestId === lastInvoiceRequestId) {
                isLoadingInvoice.value = false
            }
        }
    }

    /**
     * fmtActive formats an invoice into a shape accepted by the server.
     */
    function fmtActive(resp: InvoiceResponse, clientId: number): Invoice {
        const t = resp.totals
        return {
            baseNumber: t.baseNumber,
            clientId,
            issueDate: t.issueDate,
            dueByDate: t.dueByDate ?? undefined,
            clientSnapshot: {
                name: t.clientName,
                companyName: t.clientCompanyName,
                address: t.clientAddress,
                email: t.clientEmail,
            },
            lines: resp.lines.map((l) => ({
                productId: l.productId ?? null,
                name: l.name,
                lineType: l.lineType as LineType,
                pricingMode: (l.pricingMode ?? 'flat') as PricingMode,
                quantity: l.quantity,
                unitPriceMinor: l.unitPriceMinor,
                minutesWorked: l.minutesWorked ?? null,
                sortOrder: l.sortOrder,
            })),
            discountType: t.discountType as DiscountType,
            discountMinor: t.discountMinor,
            discountRate: t.discountRate,
            vatRate: t.vatRate,
            paidMinor: t.paidMinor,
            depositType: t.depositType as DepositType,
            depositMinor: t.depositMinor,
            depositRate: t.depositRate,
            note: t.note ?? undefined,
        }
    }

    const { liveFieldErrors, getFieldError } = useInvoiceFieldErrors(
        draftInvoice,
        serverFieldErrors,
    )

    // Helpers
    async function nextPage() {
        if (!hasMore.value || isLoadingBook.value) return
        offset.value += limit.value
        await fetchInvoiceBook()
    }

    async function prevPage() {
        if (offset.value === 0 || isLoadingBook.value) return
        offset.value = Math.max(0, offset.value - limit.value)
        await fetchInvoiceBook()
    }

    async function goToFirstPage() {
        offset.value = 0
        await fetchInvoiceBook()
    }

    function clearInvoiceBook() {
        lastBookRequestId++
        invoiceBook.value = []
        offset.value = 0
        total.value = 0
        hasMore.value = false
        bookError.value = ''
    }

    function clearActiveInvoice() {
        lastInvoiceRequestId++
        activeInvoice.value = null
        draftInvoice.value = null
        isEditing.value = false
        serverFieldErrors.value = {}
        showAllValidation.value = false
        clearVerifyState()
    }

    function initEdit() {
        if (!activeInvoice.value) return
        clearVerifyState()
        draftInvoice.value = cloneInvoice(activeInvoice.value)
        serverFieldErrors.value = {}
        showAllValidation.value = false
        isEditing.value = true
    }

    function cancelEdit() {
        if (!activeInvoice.value) return

        draftInvoice.value = cloneInvoice(activeInvoice.value)
        serverFieldErrors.value = {}
        showAllValidation.value = false
        isEditing.value = false
        clearVerifyState()
    }

    // * CRUD Operations
    function addLine(line: Omit<InvoiceLine, 'sortOrder'>) {
        if (!draftInvoice || draftInvoice.value === null) return
        addInvoiceLine(draftInvoice.value, line)
        scheduleServerVerify()
    }

    function updateLine(sortOrder: number, patch: Partial<InvoiceLine>): void {
        updateInvoiceLine(ensureDraft(), sortOrder, patch)
        scheduleServerVerify()
    }

    function removeLine(sortOrder: number): void {
        removeInvoiceLine(ensureDraft(), sortOrder)
        scheduleServerVerify()
    }
    function setNote(note: string): void {
        setInvoiceNote(ensureDraft(), note)
        scheduleServerVerify()
    }

    watch(
        draftInvoice,
        () => {
            if (Object.keys(serverFieldErrors.value).length > 0) {
                serverFieldErrors.value = {}
            }
        },
        { deep: true },
    )
    return {
        invoiceBook,
        activeInvoice,
        draftInvoice,
        limit,
        offset,
        total,
        hasMore,

        liveFieldErrors,
        showAllValidation,
        isLoadingBook,
        isLoadingInvoice,
        errorMessage: bookError,
        isEditing,
        prettyBaseNumber,

        canGoPrev,
        canGoNext,

        fetchInvoiceBook,
        fetchInvoice,
        nextPage,
        prevPage,
        goToFirstPage,
        clearInvoiceBook,
        clearActiveInvoice,
        fmtActive,
        initEdit,
        cancelEdit,
        getFieldError,
        setNote,

        addLine,
        updateLine,
        removeLine,

        totals,
        depositMinor,
        balanceDueMinor,

        verifyStatus,
        lastVerifyAt,
        serverCanonicalTotals,
        serverCanonicalLineTotals,
        runServerVerify,
        scheduleServerVerify,
    }
})

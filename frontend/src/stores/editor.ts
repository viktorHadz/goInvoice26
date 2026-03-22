import { defineStore } from 'pinia'
import { computed, ref, watch } from 'vue'
import { useClientStore } from './clients'
import { getInvAndRevNums, getInvoice, patchInvoiceStatus } from '@/utils/editorHttpHandler'
import type {
    ActiveEditorNode,
    InvBookInvoice,
    InvoiceResponse,
} from '@/components/editor/invBookTypes'
import { handleActionError } from '@/utils/errors/handleActionError'
import {
    getApiErrorMessage,
    hasFieldErrors,
    isApiError,
    isSupportOnlyApiError,
    toFieldErrorMap,
} from '@/utils/apiErrors'
import type {
    DepositType,
    DiscountType,
    Invoice,
    InvoiceLine,
    InvoiceStatus,
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
import { useSettingsStore } from './settings'
import { useInvoiceVerification } from '@/composables/useInvoiceVerification'
import { useInvoicePricing } from '@/composables/useInvoicePricing'
import { validateInvoicePayload } from '@/utils/frontendValidation'
import { emitToastError, emitToastSuccess } from '@/utils/toast'
import { apiDTO } from '@/utils/invoiceDto'
import { flattenValidationErrors } from './pdf'
import { NetworkError } from '@/utils/fetchHelper'
import { newRevisionHandler } from '@/utils/invoiceHttpHandler'
import { formatInvoiceBaseLabel } from '@/utils/invoiceLabels'

function normalizeInvoiceStatus(s: string | undefined): InvoiceStatus {
    const x = (s ?? 'draft').toLowerCase()
    if (x === 'issued' || x === 'paid' || x === 'void' || x === 'draft') return x
    return 'draft'
}

/* 
  TODO: Need to fix payments for editor and invoice dasdas
  * invoice: payment is fine. But saving logic needs to account it into totals. 

  * editor: 
    * use an array to stick the payments 
    ? add dates (DatePicker?) for each payment 
*/

export const useEditorStore = defineStore('editorStore', () => {
    const clientStore = useClientStore()
    const setsStore = useSettingsStore()

    const invoicePrefix = computed(() => setsStore.settings?.invoicePrefix ?? '')

    const invoiceBook = ref<InvBookInvoice[]>([])
    const activeInvoice = ref<Invoice | null>(null)
    const activeNode = ref<ActiveEditorNode>(null)
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

    const prettyBaseNumber = computed(() =>
        formatInvoiceBaseLabel(invoicePrefix.value, draftInvoice.value?.baseNumber),
    )

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
            status: normalizeInvoiceStatus(resp.status),
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

    async function saveRevision(inv: Invoice | null): Promise<boolean> {
        if (!clientStore.selectedClient) throw new Error('No client selected')
        if (!inv) return false
        if (inv === null) return false
        const st = inv.status ?? 'draft'
        if (st === 'paid' || st === 'void') {
            emitToastError({
                title: 'Cannot save revision',
                message:
                    st === 'void'
                        ? 'This invoice is void.'
                        : 'Reopen the invoice to issued before saving a revision.',
            })
            return false
        }
        const dto = apiDTO(inv)

        showAllValidation.value = true
        serverFieldErrors.value = {}

        const errors = validateInvoicePayload(dto)
        if (Object.keys(errors).length > 0) {
            clearVerifyState()

            emitToastError({
                title: 'Invalid invoice data',
                message: flattenValidationErrors(errors),
            })
            return false
        }

        try {
            await newRevisionHandler(dto.overview.clientId, inv.baseNumber, dto)
            showAllValidation.value = false

            clearVerifyState()
            emitToastSuccess(
                `Revision ${formatInvoiceBaseLabel(invoicePrefix.value, dto.overview.baseNumber)} saved successfully.`,
            )

            lastVerifyAt.value = null
            await fetchInvoiceBook()
            await refreshActiveInvoiceFromServer()
            return true
        } catch (err: unknown) {
            if (isApiError(err) && hasFieldErrors(err)) {
                serverFieldErrors.value = toFieldErrorMap(err.fields)
                return false
            }

            showAllValidation.value = false

            if (isApiError(err) && err.status === 409) {
                emitToastError({
                    title: 'Cannot save revision',
                    message: err.message || 'Invoice status prevents saving.',
                })
                return false
            }

            if (isApiError(err) && isSupportOnlyApiError(err)) {
                emitToastError({
                    id: err.id,
                    title: 'Could not create revision',
                    message: err.id
                        ? `Something went wrong. Please quote error ID: ${err.id}`
                        : 'Something went wrong. Please try again later.',
                })
                console.error('[invoice create]', err)
                return false
            }

            if (isApiError(err)) {
                emitToastError({
                    id: err.id,
                    title: 'Could not create invoice',
                    message: err.message || 'Please check your data and try again.',
                })
                return false
            }

            if (err instanceof NetworkError) {
                emitToastError({
                    title: 'Network error',
                    message: 'Could not reach the server. Please check your connection.',
                })
                return false
            }

            emitToastError({
                title: 'Could not create invoice',
                message: 'An unexpected error occurred. Please try again.',
            })
            console.error('[invoice create]', err)
            return false
        }
    }

    // Helpers
    const { liveFieldErrors, getFieldError } = useInvoiceFieldErrors(
        draftInvoice,
        serverFieldErrors,
    )

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

    async function refreshActiveInvoiceFromServer() {
        const node = activeNode.value
        const clientId = clientStore.selectedClient?.id
        if (!node || !clientId) return
        if (node.type === 'invoice') {
            await fetchInvoice(node.baseNo, 1)
        } else {
            await fetchInvoice(node.baseNo, node.revisionNo)
        }
    }

    async function setInvoiceLifecycleStatus(next: InvoiceStatus): Promise<boolean> {
        const inv = activeInvoice.value
        const clientId = clientStore.selectedClient?.id
        if (!inv || !clientId) return false
        try {
            await patchInvoiceStatus(clientId, inv.baseNumber, next)
            emitToastSuccess('Invoice status updated.')
            await refreshActiveInvoiceFromServer()
            await fetchInvoiceBook()
            return true
        } catch (error) {
            handleActionError(error, {
                toastTitle: 'Could not update status',
                supportMessage: 'Please try again or contact support',
                mapFields: false,
            })
            return false
        }
    }

    function initEdit() {
        if (!activeInvoice.value) return
        const st = activeInvoice.value.status ?? 'draft'
        if (st === 'paid' || st === 'void') return
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

    const quickPayOpen = ref(false)
    const setQuickPayOpen = (open: boolean) => {
        quickPayOpen.value = open
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
    // Reset invoice book on client change
    watch(
        () => clientStore.selectedClient?.id,
        async (newClientId, oldClientId) => {
            if (newClientId === oldClientId) return

            activeNode.value = null
            clearActiveInvoice()
            clearInvoiceBook()

            if (!newClientId) return

            await fetchInvoiceBook(true)
        },
        { immediate: true },
    )

    watch(activeNode, async (node) => {
        if (!node) {
            clearActiveInvoice()
            return
        }

        if (node.type === 'invoice') {
            await fetchInvoice(node.baseNo, 1)
            return
        }

        await fetchInvoice(node.baseNo, node.revisionNo)
    })
    return {
        invoiceBook,
        activeInvoice,
        activeNode,
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
        quickPayOpen,
        canGoPrev,
        canGoNext,

        fetchInvoiceBook,
        fetchInvoice,
        saveRevision,
        nextPage,
        prevPage,
        goToFirstPage,
        clearInvoiceBook,
        clearActiveInvoice,
        fmtActive,
        initEdit,
        setInvoiceLifecycleStatus,
        refreshActiveInvoiceFromServer,
        cancelEdit,
        getFieldError,
        setNote,
        setQuickPayOpen,

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

import { defineStore } from 'pinia'
import { computed, ref, watch } from 'vue'
import { useClientStore } from './clients'
import {
    deleteInvoice,
    getInvAndRevNums,
    getInvoice,
    patchInvoiceStatus,
} from '@/utils/editorHttpHandler'
import type { ActiveEditorNode, InvBookInvoice, InvoiceResponse } from '@/components/editor/invBookTypes'
import {
    areInvoiceBookFiltersEqual,
    createDefaultInvoiceBookFilters,
    cycleInvoiceBookPaymentState as nextInvoiceBookPaymentState,
    cycleInvoiceBookSort as nextInvoiceBookSort,
    toggleInvoiceBookActiveClient as nextInvoiceBookActiveClient,
    type InvoiceBookFilters,
    type InvoiceBookSortBy,
} from '@/components/editor/invoiceBookFilters'
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
    setInvoiceDueByDate,
    setInvoiceIssueDate,
    setInvoiceNote,
    setInvoiceSupplyDate,
    updateInvoiceLine,
} from '@/utils/invoiceMutations'
import { useSettingsStore } from './settings'
import { useInvoiceVerification } from '@/composables/useInvoiceVerification'
import { useInvoicePricing } from '@/composables/useInvoicePricing'
import { validateInvoicePayload } from '@/utils/frontendValidation'
import { emitToastError, emitToastInfo, emitToastSuccess } from '@/utils/toast'
import { apiDTO } from '@/utils/invoiceDto'
import { flattenValidationErrors } from './pdf'
import { NetworkError } from '@/utils/fetchHelper'
import {
    createPaymentReceiptHandler,
    deletePaymentReceiptHandler,
    newRevisionHandler,
    updateDraftInvoiceHandler,
    updatePaymentReceiptHandler,
} from '@/utils/invoiceHttpHandler'
import {
    formatInvoiceBaseLabel,
    formatInvoiceDisplayLabel,
    formatPaymentReceiptLabel,
} from '@/utils/invoiceLabels'
import { requestConfirmation, requestConfirmationChoice } from '@/utils/confirm'

function firstFieldErrorMessage(fields: Record<string, string>): string | null {
    for (const msg of Object.values(fields)) {
        if (msg) return msg
    }
    return null
}

function normalizeInvoiceStatus(s: string | undefined): InvoiceStatus {
    const x = (s ?? 'draft').toLowerCase()
    if (x === 'issued' || x === 'paid' || x === 'void' || x === 'draft') return x
    return 'draft'
}

function invoiceValidationSignal(inv: Invoice | null): string {
    if (!inv) return ''
    const linesSignal = inv.lines
        .map(
            (line) =>
                `${line.sortOrder}:${line.productId ?? ''}:${line.name}:${line.lineType}:${line.pricingMode}:${line.quantity}:${line.unitPriceMinor}:${line.minutesWorked ?? ''}`,
        )
        .join('|')
    return [
        inv.clientId,
        inv.issueDate,
        inv.supplyDate ?? '',
        inv.dueByDate ?? '',
        inv.clientSnapshot.name,
        inv.clientSnapshot.companyName,
        inv.clientSnapshot.address,
        inv.clientSnapshot.email,
        inv.note ?? '',
        inv.vatRate,
        inv.discountType,
        inv.discountRate,
        inv.discountMinor,
        inv.depositType,
        inv.depositRate,
        inv.depositMinor,
        inv.paidMinor,
        linesSignal,
    ].join('~')
}

export const useEditorStore = defineStore('editorStore', () => {
    const clientStore = useClientStore()
    const setsStore = useSettingsStore()

    const invoicePrefix = computed(() => setsStore.settings?.invoicePrefix ?? '')

    const invoiceBook = ref<InvBookInvoice[]>([])
    const invoiceBookFilters = ref<InvoiceBookFilters>(createDefaultInvoiceBookFilters())
    const activeInvoice = ref<Invoice | null>(null)
    const activeRevisionNo = ref<number>(1)
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
    const activeReceipts = ref<InvoiceResponse['receipts']>([])
    const editBaselineSnapshot = ref('')
    const pendingInvoiceBookNode = ref<ActiveEditorNode>(null)

    const prettyBaseNumber = computed(() =>
        formatInvoiceBaseLabel(invoicePrefix.value, draftInvoice.value?.baseNumber),
    )

    /**
     * showAllValidation - whether validation errors are displayed before attempting to submit.
     */
    const showAllValidation = ref(false)
    const canGoPrev = computed(() => offset.value > 0)
    const canGoNext = computed(() => hasMore.value)

    const { pricing, totals, depositMinor, balanceDueMinor } = useInvoicePricing(draftInvoice)

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

    function revisionPayloadSnapshot(inv: Invoice | null): string {
        if (!inv) return ''
        const dto = apiDTO(inv, {
            sourceRevisionNo: activeRevisionNo.value,
        })
        return JSON.stringify(dto)
    }

    function refreshEditBaselineSnapshot() {
        editBaselineSnapshot.value = revisionPayloadSnapshot(draftInvoice.value)
    }

    const hasUnsavedChanges = computed(() => {
        if (!isEditing.value || !draftInvoice.value) return false
        return revisionPayloadSnapshot(draftInvoice.value) !== editBaselineSnapshot.value
    })

    let lastBookRequestId = 0
    async function fetchInvoiceBook(reset = false) {
        if (reset) {
            offset.value = 0
        }

        const requestId = ++lastBookRequestId

        isLoadingBook.value = true
        bookError.value = ''

        try {
            const activeClientId = invoiceBookFilters.value.activeClientOnly
                ? (clientStore.selectedClient?.id ?? null)
                : null
            const data = await getInvAndRevNums(
                limit.value,
                offset.value,
                invoiceBookFilters.value,
                activeClientId,
            )

            if (requestId !== lastBookRequestId) return

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
            activeRevisionNo.value = revisionNumber
            draftInvoice.value = cloneInvoice(formatted)
            activeReceipts.value = data.receipts ?? []
            refreshEditBaselineSnapshot()

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
            supplyDate: t.supplyDate ?? undefined,
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
        const saveLabel = st === 'draft' ? 'draft' : 'revision'
        if (st === 'paid' || st === 'void') {
            emitToastError({
                title: `Cannot save ${saveLabel}`,
                message:
                    st === 'void'
                        ? 'This invoice is void.'
                        : 'Paid invoices are locked. Reopen it to issued first if the recorded payments do not match the balance due.',
            })
            return false
        }
        if (!hasUnsavedChanges.value) {
            emitToastInfo('No changes to save.')
            return false
        }
        const dto = apiDTO(inv, {
            sourceRevisionNo: activeRevisionNo.value,
        })

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
            if (st === 'draft') {
                await updateDraftInvoiceHandler(dto.overview.clientId, inv.baseNumber, dto)
                showAllValidation.value = false

                clearVerifyState()
                emitToastSuccess(
                    `Draft ${formatInvoiceBaseLabel(invoicePrefix.value, dto.overview.baseNumber)} saved.`,
                )

                lastVerifyAt.value = null
                await fetchInvoiceBook()
                await refreshActiveInvoiceFromServer()
                return true
            }

            const created = await newRevisionHandler(dto.overview.clientId, inv.baseNumber, dto)
            showAllValidation.value = false

            clearVerifyState()
            emitToastSuccess(
                `${formatInvoiceDisplayLabel(invoicePrefix.value, dto.overview.baseNumber, created.revisionNo)} saved successfully.`,
            )

            lastVerifyAt.value = null
            await fetchInvoiceBook()
            activeNode.value = {
                type: 'revision',
                clientId: dto.overview.clientId,
                id: created.revisionId,
                invoiceId: created.invoiceId,
                baseNo: dto.overview.baseNumber,
                revisionNo: created.revisionNo,
            }
            return true
        } catch (err: unknown) {
            if (isApiError(err) && hasFieldErrors(err)) {
                serverFieldErrors.value = toFieldErrorMap(err.fields)
                const firstError = firstFieldErrorMessage(serverFieldErrors.value)
                if (firstError) {
                    emitToastError({
                        title: `Cannot save ${saveLabel}`,
                        message: firstError,
                    })
                }
                return false
            }

            showAllValidation.value = false

            if (isApiError(err) && err.status === 409) {
                emitToastError({
                    title: `Cannot save ${saveLabel}`,
                    message: err.message || 'Invoice status prevents saving.',
                })
                return false
            }

            if (isApiError(err) && isSupportOnlyApiError(err)) {
                emitToastError({
                    id: err.id,
                    title: st === 'draft' ? 'Could not save draft' : 'Could not create revision',
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
                    title: st === 'draft' ? 'Could not save draft' : 'Could not save revision',
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
                title: st === 'draft' ? 'Could not save draft' : 'Could not save revision',
                message: 'An unexpected error occurred. Please try again.',
            })
            console.error('[invoice create]', err)
            return false
        }
    }

    async function createPaymentReceipt(payload: {
        amountMinor: number
        paymentDate: string
        label?: string
    }): Promise<boolean> {
        const clientId = clientStore.selectedClient?.id
        const baseNumber = activeInvoice.value?.baseNumber
        const revisionNo = activeRevisionNo.value
        if (!clientId || !baseNumber || revisionNo < 1) return false

        try {
            const created = await createPaymentReceiptHandler(clientId, baseNumber, revisionNo, payload)
            emitToastSuccess(
                `${formatPaymentReceiptLabel(invoicePrefix.value, baseNumber, revisionNo, created.receiptNo)} recorded.`,
            )
            await fetchInvoiceBook()
            await refreshActiveInvoiceFromServer()
            return true
        } catch (error) {
            handleActionError(error, {
                toastTitle: 'Could not record payment receipt',
                supportMessage: 'Please try again or contact support',
                mapFields: false,
            })
            return false
        }
    }

    async function updateSelectedReceiptMetadata(payload: {
        receiptNo: number
        paymentDate: string
        label?: string
    }): Promise<boolean> {
        const clientId = clientStore.selectedClient?.id
        const baseNumber = activeInvoice.value?.baseNumber
        const revisionNo = activeRevisionNo.value
        if (!clientId || !baseNumber || revisionNo < 1 || !payload.receiptNo) return false

        try {
            await updatePaymentReceiptHandler(clientId, baseNumber, revisionNo, payload.receiptNo, {
                paymentDate: payload.paymentDate,
                ...(payload.label ? { label: payload.label } : {}),
            })
            emitToastSuccess(
                `${formatPaymentReceiptLabel(invoicePrefix.value, baseNumber, revisionNo, payload.receiptNo)} updated.`,
            )
            await fetchInvoiceBook()
            await refreshActiveInvoiceFromServer()
            return true
        } catch (error) {
            handleActionError(error, {
                toastTitle: 'Could not update payment receipt',
                supportMessage: 'Please try again or contact support',
                mapFields: false,
            })
            return false
        }
    }

    async function deleteReceipt(receiptNo: number): Promise<boolean> {
        const clientId = clientStore.selectedClient?.id
        const baseNumber = activeInvoice.value?.baseNumber
        const revisionNo = activeRevisionNo.value
        if (!clientId || !baseNumber || revisionNo < 1 || receiptNo < 1) return false

        try {
            await deletePaymentReceiptHandler(clientId, baseNumber, revisionNo, receiptNo)
            emitToastSuccess(
                `${formatPaymentReceiptLabel(invoicePrefix.value, baseNumber, revisionNo, receiptNo)} deleted.`,
            )
            await fetchInvoiceBook()
            await refreshActiveInvoiceFromServer()
            return true
        } catch (error) {
            handleActionError(error, {
                toastTitle: 'Could not delete receipt',
                supportMessage: 'Please try again or contact support',
                mapFields: false,
            })
            return false
        }
    }

    async function resolveUnsavedChangesBeforeProceed(actionLabel: string): Promise<boolean> {
        if (!hasUnsavedChanges.value) return true

        const decision = await requestConfirmationChoice({
            title: 'Save your changes?',
            message: `You have unsaved edits. Save them before you ${actionLabel}?`,
            details:
                'Save keeps your edits, Discard drops them, and Stay keeps you on the current draft.',
            confirmLabel: 'Discard changes',
            cancelLabel: 'Stay here',
            alternateLabel: 'Save changes',
            confirmVariant: 'danger',
            alternateVariant: 'success',
        })

        if (decision === 'cancel') {
            return false
        }

        if (decision === 'alternate') {
            return await saveRevision(draftInvoice.value)
        }

        discardDraftChanges()
        return true
    }

    // Helpers
    const { liveFieldErrors, getFieldError } = useInvoiceFieldErrors(
        draftInvoice,
        serverFieldErrors,
        pricing,
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

    async function setInvoiceBookFilters(nextFilters: InvoiceBookFilters) {
        if (areInvoiceBookFiltersEqual(invoiceBookFilters.value, nextFilters)) return

        invoiceBookFilters.value = nextFilters
        await fetchInvoiceBook(true)
    }

    async function cycleBookSort(sortBy: InvoiceBookSortBy) {
        await setInvoiceBookFilters(nextInvoiceBookSort(invoiceBookFilters.value, sortBy))
    }

    async function cycleBookPaymentState() {
        await setInvoiceBookFilters(nextInvoiceBookPaymentState(invoiceBookFilters.value))
    }

    async function toggleBookActiveClientOnly() {
        await setInvoiceBookFilters(nextInvoiceBookActiveClient(invoiceBookFilters.value))
    }

    async function resetInvoiceBookFilters() {
        await setInvoiceBookFilters(createDefaultInvoiceBookFilters())
    }

    async function selectInvoiceBookNode(node: ActiveEditorNode) {
        if (!(await resolveUnsavedChangesBeforeProceed('switch invoices'))) {
            return
        }

        pendingInvoiceBookNode.value = null

        if (!node) {
            activeNode.value = null
            return
        }

        const currentClientId = clientStore.selectedClient?.id ?? null
        if (node.clientId !== currentClientId) {
            pendingInvoiceBookNode.value = node
            clientStore.selectClientById(node.clientId)
            return
        }

        activeNode.value = node
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
        activeRevisionNo.value = 1
        draftInvoice.value = null
        activeReceipts.value = []
        isEditing.value = false
        editBaselineSnapshot.value = ''
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

    async function requestInvoiceLifecycleStatusChange(next: InvoiceStatus): Promise<boolean> {
        const current = (activeInvoice.value?.status ?? 'draft') as InvoiceStatus
        if (next === current) return true
        if (!(await resolveUnsavedChangesBeforeProceed('change the invoice status'))) {
            return false
        }

        if (next === 'void') {
            const confirmed = await requestConfirmation({
                title: 'Void invoice?',
                message: `Void ${formatInvoiceBaseLabel(invoicePrefix.value, activeInvoice.value?.baseNumber)}?`,
                details:
                    "Voiding makes the invoice and all of it's revisions inactive and final. It cannot be edited, reopened, or deleted afterward.",
                confirmLabel: 'Void invoice',
                cancelLabel: 'Keep current status',
                confirmVariant: 'danger',
            })
            if (!confirmed) return false
        }

        return await setInvoiceLifecycleStatus(next)
    }

    async function deleteActiveInvoice(): Promise<boolean> {
        const clientId = clientStore.selectedClient?.id
        const baseNumber = activeInvoice.value?.baseNumber ?? activeNode.value?.baseNo
        if (!clientId || !baseNumber) return false

        const shouldResetPage = offset.value > 0 && invoiceBook.value.length <= 1
        const invoiceLabel = formatInvoiceBaseLabel(invoicePrefix.value, baseNumber)

        try {
            await deleteInvoice(clientId, baseNumber)

            activeNode.value = null
            clearActiveInvoice()
            await fetchInvoiceBook(shouldResetPage)

            emitToastSuccess(`${invoiceLabel} deleted.`)
            return true
        } catch (error) {
            handleActionError(error, {
                toastTitle: 'Delete invoice failed',
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
        refreshEditBaselineSnapshot()
    }

    function discardDraftChanges() {
        if (!activeInvoice.value) return
        draftInvoice.value = cloneInvoice(activeInvoice.value)
        serverFieldErrors.value = {}
        showAllValidation.value = false
        isEditing.value = false
        refreshEditBaselineSnapshot()
        clearVerifyState()
    }

    async function cancelEdit() {
        if (!activeInvoice.value) return
        if (!(await resolveUnsavedChangesBeforeProceed('cancel editing'))) {
            return
        }
        if (isEditing.value) {
            discardDraftChanges()
        }
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

    function setIssueDate(value: string): void {
        setInvoiceIssueDate(ensureDraft(), value)
        scheduleServerVerify()
    }

    function setSupplyDate(value: string): void {
        setInvoiceSupplyDate(ensureDraft(), value)
        scheduleServerVerify()
    }

    function setDueByDate(value: string): void {
        setInvoiceDueByDate(ensureDraft(), value)
        scheduleServerVerify()
    }

    watch(
        () => invoiceValidationSignal(draftInvoice.value),
        () => {
            if (Object.keys(serverFieldErrors.value).length > 0) {
                serverFieldErrors.value = {}
            }
        },
    )

    let restoringClientSelection = false
    watch(
        () => clientStore.selectedClient?.id,
        async (newClientId, oldClientId) => {
            if (restoringClientSelection) {
                restoringClientSelection = false
                return
            }

            const isInitialLoad = typeof oldClientId === 'undefined'
            const shouldRefreshBook = isInitialLoad || invoiceBookFilters.value.activeClientOnly

            if (newClientId === oldClientId && !shouldRefreshBook) return

            if (
                !isInitialLoad &&
                oldClientId !== newClientId &&
                !(await resolveUnsavedChangesBeforeProceed('switch clients'))
            ) {
                restoringClientSelection = true
                clientStore.selectClientById(oldClientId ?? null)
                return
            }

            activeNode.value = null
            clearActiveInvoice()

            if (shouldRefreshBook) {
                await fetchInvoiceBook(true)
            }

            if (
                pendingInvoiceBookNode.value &&
                pendingInvoiceBookNode.value.clientId === (newClientId ?? null)
            ) {
                activeNode.value = pendingInvoiceBookNode.value
            }

            pendingInvoiceBookNode.value = null
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

    function reset() {
        pendingInvoiceBookNode.value = null
        activeNode.value = null
        invoiceBookFilters.value = createDefaultInvoiceBookFilters()
        clearInvoiceBook()
        clearActiveInvoice()
        isLoadingBook.value = false
        isLoadingInvoice.value = false
    }

    return {
        invoiceBook,
        invoiceBookFilters,
        activeInvoice,
        activeNode,
        activeRevisionNo,
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
        activeReceipts,
        hasUnsavedChanges,

        fetchInvoiceBook,
        fetchInvoice,
        saveRevision,
        createPaymentReceipt,
        updateSelectedReceiptMetadata,
        deleteReceipt,
        nextPage,
        prevPage,
        goToFirstPage,
        cycleBookSort,
        cycleBookPaymentState,
        toggleBookActiveClientOnly,
        resetInvoiceBookFilters,
        selectInvoiceBookNode,
        clearInvoiceBook,
        clearActiveInvoice,
        fmtActive,
        initEdit,
        deleteActiveInvoice,
        setInvoiceLifecycleStatus,
        requestInvoiceLifecycleStatusChange,
        refreshActiveInvoiceFromServer,
        cancelEdit,
        resolveUnsavedChangesBeforeProceed,
        getFieldError,
        setNote,
        setIssueDate,
        setSupplyDate,
        setDueByDate,

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
        reset,
    }
})

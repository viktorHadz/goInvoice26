import { defineStore } from 'pinia'
import { computed, ref, watch } from 'vue'
import type {
    Invoice,
    InvoiceLine,
    Totals,
    MoneyMinor,
    DepositType,
    DiscountType,
} from '@/components/invoice/invoiceTypes'
import {
    newInvoiceHandler,
    getNewInvoiceNumber,
    verifyInvoiceHandler,
} from '@/utils/invoiceHttpHandler'
import { useClientStore } from './clients'
import { useSettingsStore } from './settings'
import {
    isApiError,
    hasFieldErrors,
    isSupportOnlyApiError,
    toFieldErrorMap,
} from '@/utils/apiErrors'
import { validateInvoicePayload } from '@/utils/frontendValidation'
import { emitToastError, emitToastSuccess } from '@/utils/toast'
import { NetworkError } from '@/utils/fetchHelper'
import type { Client } from '@/utils/clientHttpHandler'
import { lineTotalMinor, toMinor } from '@/utils/money'
import { apiDTO } from '@/utils/invoiceDto'
import { flattenValidationErrors } from './pdf'
import { formatInvoiceBaseLabel } from '@/utils/invoiceLabels'
import { useInvoiceFieldErrors } from '@/composables/useInvoiceFieldErrors'
import {
    addInvoiceLine,
    clearInvoiceDeposit,
    clearInvoiceDiscount,
    removeInvoiceLine,
    setInvoiceDepositFixedGBP,
    setInvoiceDepositPercent,
    setInvoiceDepositType,
    setInvoiceDiscountFixedGBP,
    setInvoiceDiscountPercent,
    setInvoiceDiscountType,
    setInvoiceDueByDate,
    setInvoiceIssueDate,
    setInvoiceNote,
    setInvoiceVatRateBps,
    updateInvoiceLine,
} from '@/utils/invoiceMutations'
import { useInvoicePricing } from '@/composables/useInvoicePricing'
import type { DraftPaymentInput } from '@/utils/invoiceDto'

type PendingPayment = DraftPaymentInput & {
    tempId: string
}

function firstFieldErrorMessage(fields: Record<string, string>): string | null {
    for (const msg of Object.values(fields)) {
        if (msg) return msg
    }
    return null
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

// * INVOICE STORE
export const useInvoiceStore = defineStore('invoice', () => {
    const invoice = ref<Invoice | null>(null)
    const pendingPayments = ref<PendingPayment[]>([])
    const serverFieldErrors = ref<Record<string, string>>({})
    /**
     * showAllValidation - whether validation errors are displayed before attempting to submit.
     */
    const showAllValidation = ref(false)
    const setsStore = useSettingsStore()
    const invoicePrefix = computed(() => setsStore.settings?.invoicePrefix ?? '')
    const verifyStatus = ref<'idle' | 'checking' | 'ok' | 'mismatch' | 'invalid' | 'error'>('idle')
    const lastVerifyAt = ref<number | null>(null)
    const serverCanonicalTotals = ref<Totals | null>(null)
    const serverCanonicalLineTotals = ref<Record<number, MoneyMinor>>({})
    const lastVerifyFailureToastedAt = ref<number | null>(null)

    // Called inside functions only never at module scope
    const getClientStore = () => useClientStore()

    function getInvoiceClient(): Client | null {
        const inv = invoice.value
        if (!inv) return null

        const clientStore = getClientStore()
        return clientStore.clients.find((c) => c.id === inv.clientId) ?? null
    }
    function setClientSnapshot(c: Client): void {
        const inv = ensure()

        inv.clientSnapshot = {
            name: c.name ?? '',
            companyName: c.companyName ?? '',
            address: c.address ?? '',
            email: c.email ?? '',
        }
    }

    function buildFreshInvoiceTemplate(c: Client): Omit<Invoice, 'baseNumber'> {
        return {
            clientId: c.id,
            status: 'draft',
            issueDate: '',
            dueByDate: undefined,
            clientSnapshot: {
                name: c.name ?? '',
                companyName: c.companyName ?? '',
                address: c.address ?? '',
                email: c.email ?? '',
            },
            note: '',
            vatRate: 2000,
            discountType: 'none',
            discountRate: 0,
            discountMinor: 0,
            lines: [],
            paidMinor: 0,
            depositType: 'none',
            depositRate: 0,
            depositMinor: 0,
        }
    }

    function ensure(): Invoice {
        if (!invoice.value) throw new Error('Invoice not initialised')
        return invoice.value
    }

    function syncInvoicePaidMinorFromPending() {
        const inv = invoice.value
        if (!inv) return
        const paid = pendingPayments.value.reduce((sum, p) => sum + p.amountMinor, 0)
        inv.paidMinor = Math.max(0, paid) as MoneyMinor
    }

    function stagePendingPayment(payment: DraftPaymentInput): void {
        pendingPayments.value = [
            ...pendingPayments.value,
            {
                tempId: `${Date.now()}-${Math.random().toString(16).slice(2)}`,
                ...payment,
            },
        ]
        syncInvoicePaidMinorFromPending()
        scheduleServerVerify()
    }

    function removePendingPayment(tempId: string): void {
        pendingPayments.value = pendingPayments.value.filter((p) => p.tempId !== tempId)
        syncInvoicePaidMinorFromPending()
        scheduleServerVerify()
    }

    function clearPendingPayments(): void {
        pendingPayments.value = []
        syncInvoicePaidMinorFromPending()
    }

    const prettyBaseNumber = computed(() =>
        formatInvoiceBaseLabel(invoicePrefix.value, invoice.value?.baseNumber),
    )
    // Pricing is a single computed that derives all pricing in one pass
    const { pricing, totals, depositMinor, balanceDueMinor } = useInvoicePricing(invoice)

    let verifyTimer: number | null = null
    let verifyAbort: AbortController | null = null

    function clearVerifyTimer() {
        if (verifyTimer != null) {
            window.clearTimeout(verifyTimer)
            verifyTimer = null
        }
    }

    function abortVerify() {
        if (verifyAbort) {
            verifyAbort.abort()
            verifyAbort = null
        }
    }

    /**
     * ! Verifies server totals - essential for optimistic updates
     */
    async function runServerVerify() {
        const inv = invoice.value
        const { lsClientId } = getClientStore()
        if (!inv || !lsClientId) return

        const dto = apiDTO(inv)
        abortVerify()
        verifyAbort = new AbortController()

        verifyStatus.value = 'checking'
        try {
            const res = await verifyInvoiceHandler(dto.overview.clientId, inv.baseNumber, dto, {
                signal: verifyAbort.signal,
            })

            const canonical = (res as any)?.invoice as any
            const canonicalLines: any[] = Array.isArray(canonical?.lines) ? canonical.lines : []
            const canonicalTotals = canonical?.totals

            const canonicalBySort: Record<number, MoneyMinor> = {}
            for (const ln of canonicalLines) {
                const so = Number(ln?.sortOrder)
                const lt = Number(ln?.lineTotalMinor)
                if (Number.isFinite(so) && Number.isFinite(lt)) {
                    canonicalBySort[so] = Math.round(lt) as MoneyMinor
                }
            }

            const serverTotals: Totals | null =
                canonicalTotals &&
                typeof canonicalTotals === 'object' &&
                Number.isFinite((canonicalTotals as any).subtotalMinor) &&
                Number.isFinite((canonicalTotals as any).discountMinor) &&
                Number.isFinite((canonicalTotals as any).subtotalAfterDiscountMinor) &&
                Number.isFinite((canonicalTotals as any).vatMinor) &&
                Number.isFinite((canonicalTotals as any).totalMinor)
                    ? {
                          subtotalMinor: Math.round(
                              (canonicalTotals as any).subtotalMinor,
                          ) as MoneyMinor,
                          discountMinor: Math.round(
                              (canonicalTotals as any).discountMinor,
                          ) as MoneyMinor,
                          subtotalAfterDiscountMinor: Math.round(
                              (canonicalTotals as any).subtotalAfterDiscountMinor,
                          ) as MoneyMinor,
                          vatMinor: Math.round((canonicalTotals as any).vatMinor) as MoneyMinor,
                          totalMinor: Math.round((canonicalTotals as any).totalMinor) as MoneyMinor,
                      }
                    : null

            serverCanonicalLineTotals.value = canonicalBySort
            serverCanonicalTotals.value = serverTotals
            lastVerifyAt.value = Date.now()
            lastVerifyFailureToastedAt.value = null

            const optimisticTotals = totals.value
            let mismatch = false

            if (optimisticTotals && serverTotals) {
                mismatch =
                    optimisticTotals.subtotalMinor !== serverTotals.subtotalMinor ||
                    optimisticTotals.discountMinor !== serverTotals.discountMinor ||
                    optimisticTotals.subtotalAfterDiscountMinor !==
                        serverTotals.subtotalAfterDiscountMinor ||
                    optimisticTotals.vatMinor !== serverTotals.vatMinor ||
                    optimisticTotals.totalMinor !== serverTotals.totalMinor
            }

            for (const line of inv.lines) {
                const serverLT = canonicalBySort[line.sortOrder]
                if (serverLT == null) continue
                const optimisticLT = lineTotalMinor(line)
                if (optimisticLT !== serverLT) {
                    mismatch = true
                    break
                }
            }

            verifyStatus.value = mismatch ? 'mismatch' : 'ok'
        } catch (err: unknown) {
            if (err instanceof NetworkError) {
                verifyStatus.value = 'error'
                if (lastVerifyFailureToastedAt.value == null) {
                    lastVerifyFailureToastedAt.value = Date.now()
                    emitToastError({
                        title: 'Verification unavailable',
                        message: 'Could not verify totals right now. Check your connection and try again.',
                    })
                }
                return
            }

            if (isApiError(err) && err.code === 'VALIDATION_FAILED') {
                verifyStatus.value = 'invalid'
                if (hasFieldErrors(err)) {
                    serverFieldErrors.value = toFieldErrorMap(err.fields)
                }
                return
            }

            verifyStatus.value = 'error'
            if (
                isApiError(err) &&
                lastVerifyFailureToastedAt.value == null
            ) {
                lastVerifyFailureToastedAt.value = Date.now()
                emitToastError({
                    id: err.id,
                    title: isSupportOnlyApiError(err) ? 'Server error' : 'Verification failed',
                    message: isSupportOnlyApiError(err)
                        ? err.id
                            ? `Something went wrong. Please quote error ID: ${err.id}`
                            : 'Something went wrong. Please try again later.'
                        : err.message || 'Could not verify totals right now. Please try again.',
                })
            }
            console.error('[invoice verify]', err)
        }
    }

    /**
     * Calls server verification on crud ops. Debounced to guard against keystrokes and races
     * - ignored until initial date is set
     */
    function scheduleServerVerify(debounceDur: number = 1000) {
        if (typeof window === 'undefined') return
        if (!invoice.value) return
        if (invoice.value.lines.length <= 0) return // abort on first load
        if (!invoice.value.issueDate) return // abort if missing as server throws

        clearVerifyTimer()
        // debouncer
        verifyTimer = window.setTimeout(() => {
            runServerVerify()
        }, debounceDur)
    }

    watch(() => invoiceValidationSignal(invoice.value), () => {
        if (Object.keys(serverFieldErrors.value).length > 0) {
            serverFieldErrors.value = {}
        }
    })

    async function initInvoiceFromServer(
        newInvoiceData: Omit<Invoice, 'baseNumber'>,
    ): Promise<void> {
        const { lsClientId } = getClientStore()
        if (!lsClientId) throw new Error('No client selected')

        const bNum = await getNewInvoiceNumber(lsClientId)
        invoice.value = { ...newInvoiceData, baseNumber: bNum }
        clearPendingPayments()
    }

    // * -------- LINES CRUD -------- * //
    function addLine(line: Omit<InvoiceLine, 'sortOrder'>) {
        addInvoiceLine(ensure(), line)
        scheduleServerVerify()
    }

    function updateLine(sortOrder: number, patch: Partial<InvoiceLine>): void {
        updateInvoiceLine(ensure(), sortOrder, patch)
        scheduleServerVerify()
    }

    function removeLine(sortOrder: number): void {
        removeInvoiceLine(ensure(), sortOrder)
        scheduleServerVerify()
    }

    // * ----- Setters ----- * //
    function setIssueDate(v: string): void {
        setInvoiceIssueDate(ensure(), v)
        scheduleServerVerify()
    }

    function setDueByDate(v: string): void {
        setInvoiceDueByDate(ensure(), v)
        scheduleServerVerify()
    }

    function setNote(note: string): void {
        setInvoiceNote(ensure(), note)
        scheduleServerVerify()
    }

    function setVatRateBps(v: number): void {
        setInvoiceVatRateBps(ensure(), v)
        scheduleServerVerify()
    }

    function setDiscountType(t: DiscountType): void {
        setInvoiceDiscountType(ensure(), t)
        scheduleServerVerify()
    }

    function setDiscountFixedGBP(gbp: number): void {
        setInvoiceDiscountFixedGBP(ensure(), gbp)
        scheduleServerVerify()
    }

    function setDiscountPercent(percent: number): void {
        setInvoiceDiscountPercent(ensure(), percent)
        scheduleServerVerify()
    }

    function clearDiscount(): void {
        clearInvoiceDiscount(ensure())
        scheduleServerVerify()
    }

    function setDepositType(t: DepositType): void {
        setInvoiceDepositType(ensure(), t)
        scheduleServerVerify()
    }

    function setDepositFixedGBP(gbp: number): void {
        setInvoiceDepositFixedGBP(ensure(), gbp)
        scheduleServerVerify()
    }

    function setDepositPercent(percent: number): void {
        setInvoiceDepositPercent(ensure(), percent)
        scheduleServerVerify()
    }

    function clearDeposit(): void {
        clearInvoiceDeposit(ensure())
        scheduleServerVerify()
    }

    async function newDraftInvoice(inv: Invoice): Promise<boolean> {
        const clientStore = getClientStore()
        const { lsClientId, selectedClient } = clientStore
        if (!lsClientId) throw new Error('No client selected')

        const dto = apiDTO(
            inv,
            pendingPayments.value.map((p) => ({
                amountMinor: p.amountMinor,
                paymentDate: p.paymentDate,
                ...(p.label ? { label: p.label } : {}),
            })),
        )

        showAllValidation.value = true
        serverFieldErrors.value = {}

        const errors = validateInvoicePayload(dto)
        if (Object.keys(errors).length > 0) {
            clearVerifyTimer()
            abortVerify()

            emitToastError({
                title: 'Invalid invoice data',
                message: flattenValidationErrors(errors),
            })
            return false
        }

        try {
            await newInvoiceHandler(dto.overview.clientId, inv.baseNumber, dto)
            showAllValidation.value = false

            clearVerifyTimer()
            abortVerify()
            emitToastSuccess(
                `Invoice ${formatInvoiceBaseLabel(invoicePrefix.value, inv.baseNumber)} saved as draft.`,
            )
            lastVerifyFailureToastedAt.value = null

            if (selectedClient) {
                const template = buildFreshInvoiceTemplate(selectedClient)
                await initInvoiceFromServer(template)
            }

            return true
        } catch (err: unknown) {
            if (isApiError(err) && hasFieldErrors(err)) {
                serverFieldErrors.value = toFieldErrorMap(err.fields)
                const firstError = firstFieldErrorMessage(serverFieldErrors.value)
                if (firstError) {
                    emitToastError({
                        title: 'Cannot create draft',
                        message: firstError,
                    })
                }
                return false
            }

            showAllValidation.value = false

            if (isApiError(err) && isSupportOnlyApiError(err)) {
                emitToastError({
                    id: err.id,
                    title: 'Could not create invoice',
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
    const { liveFieldErrors, getFieldError } = useInvoiceFieldErrors(
        invoice,
        serverFieldErrors,
        pricing,
    )

    function clearServerFieldErrors() {
        serverFieldErrors.value = {}
    }

    return {
        // state
        invoice,
        pendingPayments,
        serverFieldErrors,
        showAllValidation,
        liveFieldErrors,
        getFieldError,
        clearServerFieldErrors,
        prettyBaseNumber,
        buildFreshInvoiceTemplate,

        // init
        initInvoiceFromServer,

        // derived
        totals,
        depositMinor,
        balanceDueMinor,

        // optimistic UI + server verification
        verifyStatus,
        lastVerifyAt,
        serverCanonicalTotals,
        serverCanonicalLineTotals,
        scheduleServerVerify,

        // lines
        addLine,
        updateLine,
        removeLine,

        // setters
        setIssueDate,
        setDueByDate,
        setNote,
        setVatRateBps,
        setDiscountType,
        setDiscountFixedGBP,
        setDiscountPercent,
        clearDiscount,
        setDepositType,
        setDepositFixedGBP,
        setDepositPercent,
        clearDeposit,
        stagePendingPayment,
        removePendingPayment,
        clearPendingPayments,
        getInvoiceClient,
        setClientSnapshot,

        // API
        newDraftInvoice,
    }
})

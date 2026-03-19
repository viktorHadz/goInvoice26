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
import { asNum, round0 } from '@/utils/numbers'
import { clamp } from '@vueuse/core'
import {
    calcBalanceDueMinor,
    calcDepositMinor,
    calcTotals,
    lineTotalMinor,
    toMinor,
} from '@/utils/money'
import { apiDTO } from '@/utils/invoiceDto'
import { flattenValidationErrors } from './pdf'
import { fmtPrettyInvoiceNumber } from '@/utils/numbers'
import { useInvoiceFieldErrors } from '@/composables/getFieldError'
// Exported so components can import directly rather than always going through the store

function assignDefined<T extends object>(target: T, patch: Partial<T>) {
    for (const k in patch) {
        const v = patch[k as keyof T]
        if (v !== undefined) (target as any)[k] = v
    }
    return target
}

// INVOICE STORE
export const useInvoiceStore = defineStore('invoice', () => {
    const invoice = ref<Invoice | null>(null)
    const serverFieldErrors = ref<Record<string, string>>({})
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

    const prettyBaseNumber = computed(() =>
        fmtPrettyInvoiceNumber(invoicePrefix.value, invoice.value?.baseNumber),
    )

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
            if (verifyStatus.value === 'ok') lastVerifyFailureToastedAt.value = null
        } catch (err: unknown) {
            if (err instanceof NetworkError) return

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
                isSupportOnlyApiError(err) &&
                lastVerifyFailureToastedAt.value == null
            ) {
                lastVerifyFailureToastedAt.value = Date.now()
                emitToastError({
                    id: err.id,
                    title: 'Server error',
                    message: err.id
                        ? `Something went wrong. Please quote error ID: ${err.id}`
                        : 'Something went wrong. Please try again later.',
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

    // Single computed that derives all pricing in one pass
    const pricing = computed(() => {
        const inv = invoice.value
        if (!inv) return null

        const totals = calcTotals(inv)
        const deposit = calcDepositMinor(inv, totals.totalMinor)
        const balanceDue = calcBalanceDueMinor(totals.totalMinor, deposit, inv.paidMinor)

        return { totals, depositMinor: deposit, balanceDueMinor: balanceDue }
    })

    const totals = computed<Totals | null>(() => pricing.value?.totals ?? null)
    const depositMinor = computed<MoneyMinor>(
        () => pricing.value?.depositMinor ?? (0 as MoneyMinor),
    )
    const balanceDueMinor = computed<MoneyMinor>(
        () => pricing.value?.balanceDueMinor ?? (0 as MoneyMinor),
    )

    watch(
        invoice,
        () => {
            if (Object.keys(serverFieldErrors.value).length > 0) {
                serverFieldErrors.value = {}
            }
        },
        { deep: true },
    )

    async function initInvoiceFromServer(
        newInvoiceData: Omit<Invoice, 'baseNumber'>,
    ): Promise<void> {
        const { lsClientId } = getClientStore()
        if (!lsClientId) throw new Error('No client selected')

        const bNum = await getNewInvoiceNumber(lsClientId)
        invoice.value = { ...newInvoiceData, baseNumber: bNum }
    }

    // * -------- LINES CRUD -------- * //

    function addLine(line: Omit<InvoiceLine, 'sortOrder'>) {
        const inv = ensure()

        const canMerge = line.lineType !== 'custom' && line.productId != null
        const existingLine = canMerge
            ? inv.lines.find((ln) => ln.productId === line.productId)
            : undefined

        if (existingLine) {
            const qtyToAdd = Number.isFinite(line.quantity) && line.quantity > 0 ? line.quantity : 1
            existingLine.quantity += qtyToAdd
            return
        }

        const maxSort = inv.lines.reduce(
            (m, current) => Math.max(m, asNum(current.sortOrder, 0)),
            0,
        )
        inv.lines.push({ ...line, sortOrder: maxSort + 1 })
        scheduleServerVerify()
    }

    function updateLine(sortOrder: number, patch: Partial<InvoiceLine>): void {
        const inv = ensure()
        const line = inv.lines.find((l) => l.sortOrder === sortOrder)
        if (!line) return
        scheduleServerVerify()

        assignDefined(line, patch)
    }

    function removeLine(sortOrder: number): void {
        const inv = ensure()

        inv.lines = inv.lines
            .filter((l) => l.sortOrder !== sortOrder)
            .sort((a, b) => a.sortOrder - b.sortOrder)
            .map((l, i) => ({ ...l, sortOrder: i + 1 }))

        scheduleServerVerify()
    }

    // Setters — all validated/clamped here so components stay dumb
    function setIssueDate(v: string): void {
        ensure().issueDate = String(v ?? '')
    }

    function setDueByDate(v: string): void {
        ensure().dueByDate = String(v ?? '')
        scheduleServerVerify()
    }

    function setNote(note: string): void {
        ensure().note = String(note ?? '')
        scheduleServerVerify()
    }

    function setVatRateBps(v: number): void {
        ensure().vatRate = clamp(round0(asNum(v, 0)), 0, 10000)
        scheduleServerVerify()
    }

    function setDiscountType(t: DiscountType): void {
        const inv = ensure()
        scheduleServerVerify()

        inv.discountType = t

        if (t === 'none') {
            inv.discountMinor = 0
            inv.discountRate = 0
        }

        if (t === 'percent') {
            inv.discountRate = clamp(round0(asNum(inv.discountRate, 0)), 0, 10000)
            inv.discountMinor = 0
        }

        if (t === 'fixed') {
            inv.discountMinor = Math.max(0, round0(asNum(inv.discountMinor, 0)))
            inv.discountRate = 0
        }
    }

    function setDiscountFixedGBP(gbp: number): void {
        const inv = ensure()

        inv.discountType = 'fixed'
        inv.discountMinor = Math.max(0, toMinor(gbp))
        inv.discountRate = 0

        scheduleServerVerify()
    }

    function setDiscountPercent(percent: number): void {
        const inv = ensure()

        inv.discountType = 'percent'
        inv.discountRate = clamp(round0(asNum(percent, 0) * 100), 0, 10000)
        inv.discountMinor = 0

        scheduleServerVerify()
    }

    function clearDiscount(): void {
        const inv = ensure()

        inv.discountType = 'none'
        inv.discountMinor = 0
        inv.discountRate = 0

        scheduleServerVerify()
    }
    function setDepositType(t: DepositType): void {
        const inv = ensure()

        inv.depositType = t

        if (t === 'none') {
            inv.depositMinor = 0
            inv.depositRate = 0
        }

        if (t === 'percent') {
            inv.depositRate = clamp(round0(asNum(inv.depositRate, 0)), 0, 10000)
            inv.depositMinor = 0
        }

        if (t === 'fixed') {
            inv.depositMinor = Math.max(0, round0(asNum(inv.depositMinor, 0)))
            inv.depositRate = 0
        }
        scheduleServerVerify()
    }

    function setDepositFixedGBP(gbp: number): void {
        const inv = ensure()

        inv.depositType = 'fixed'
        inv.depositMinor = Math.max(0, toMinor(gbp))
        inv.depositRate = 0
        scheduleServerVerify()
    }

    function setDepositPercent(percent: number): void {
        const inv = ensure()

        inv.depositType = 'percent'
        inv.depositRate = clamp(round0(asNum(percent, 0) * 100), 0, 10000)
        inv.depositMinor = 0
        scheduleServerVerify()
    }

    function clearDeposit(): void {
        const inv = ensure()

        inv.depositType = 'none'
        inv.depositMinor = 0
        inv.depositRate = 0
        scheduleServerVerify()
    }

    function setPaidGBP(gbp: number): void {
        ensure().paidMinor = Math.max(0, toMinor(gbp))
        scheduleServerVerify()
    }

    async function newDraftInvoice(inv: Invoice): Promise<boolean> {
        const clientStore = getClientStore()
        const { lsClientId, selectedClient } = clientStore
        if (!lsClientId) throw new Error('No client selected')

        const dto = apiDTO(inv)

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
                `Invoice ${fmtPrettyInvoiceNumber(invoicePrefix.value, inv.baseNumber)} saved as draft.`,
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
    const { liveFieldErrors, getFieldError } = useInvoiceFieldErrors(invoice, serverFieldErrors)

    function clearServerFieldErrors() {
        serverFieldErrors.value = {}
    }

    return {
        // state
        invoice,
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
        setPaidGBP,
        getInvoiceClient,
        setClientSnapshot,

        // helpers exported(again) for components
        lineTotalMinor,
        toMinor,

        // API
        newDraftInvoice,
    }
})

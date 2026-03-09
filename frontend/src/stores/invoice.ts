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
import { isApiError, toFieldErrorMap } from '@/utils/apiErrors'
import { validateInvoicePayload } from '@/utils/frontendValidation'

// Primitives
type Int = number

const isFiniteNum = (n: unknown): n is number => typeof n === 'number' && Number.isFinite(n)
const asNum = (n: unknown, fallback = 0) => (isFiniteNum(n) ? n : fallback)
const clamp = (n: number, min: number, max: number) => Math.min(max, Math.max(min, n))
const round0 = (n: number): Int => Math.round(n)

const mulBpsRound = (baseMinor: MoneyMinor, bps: number): MoneyMinor => {
    const b = clamp(round0(asNum(bps, 0)), 0, 10000)
    return round0((baseMinor * b) / 10000) as MoneyMinor
}

// Exported so components can import directly rather than always going through the store
export const toMinor = (gbp: number): MoneyMinor => round0(asNum(gbp, 0) * 100) as MoneyMinor
export const fromMinor = (minor: MoneyMinor): number => (asNum(minor, 0) as number) / 100
export const fmtGBPMinor = (minor: MoneyMinor): string =>
    new Intl.NumberFormat('en-GB', { style: 'currency', currency: 'GBP' }).format(fromMinor(minor))

function assignDefined<T extends object>(target: T, patch: Partial<T>) {
    for (const k in patch) {
        const v = patch[k as keyof T]
        if (v !== undefined) (target as any)[k] = v
    }
    return target
}

// Pure pricing functions
// Exported to test independently of the store (sometime in the future)
export function lineTotalMinor(l: InvoiceLine): MoneyMinor {
    const qty = asNum(l.quantity, 0)
    const unit = asNum(l.unitPriceMinor, 0)

    if (l.pricingMode === 'hourly') {
        const minutes = asNum(l.minutesWorked, 0)
        // integer-style: round(qty * unit * minutes / 60)
        return round0((qty * unit * minutes) / 60) as MoneyMinor
    }

    return round0(qty * unit) as MoneyMinor
}

export function calcTotals(inv: Invoice): Totals {
    const subtotalMinor = inv.lines.reduce(
        (sum, l) => (sum + lineTotalMinor(l)) as MoneyMinor,
        0 as MoneyMinor,
    )

    const discountMinor: MoneyMinor = (() => {
        if (inv.discountType === 'none') return 0 as MoneyMinor
        if (inv.discountType === 'fixed') {
            return clamp(asNum(inv.discountValue, 0), 0, subtotalMinor) as MoneyMinor
        }
        // percent (basis points)
        return clamp(mulBpsRound(subtotalMinor, inv.discountValue), 0, subtotalMinor) as MoneyMinor
    })()

    const subtotalAfterDiscountMinor = Math.max(0, subtotalMinor - discountMinor) as MoneyMinor
    const vatMinor = mulBpsRound(subtotalAfterDiscountMinor, asNum(inv.vatRate, 0))
    const totalMinor = (subtotalAfterDiscountMinor + vatMinor) as MoneyMinor

    return { subtotalMinor, discountMinor, subtotalAfterDiscountMinor, vatMinor, totalMinor }
}

export function calcDepositMinor(inv: Invoice, totalMinor: MoneyMinor): MoneyMinor {
    if (inv.depositType === 'none') return 0 as MoneyMinor
    if (inv.depositType === 'fixed') {
        return clamp(asNum(inv.depositValue, 0), 0, totalMinor) as MoneyMinor
    }
    // percent (basis points)
    return clamp(mulBpsRound(totalMinor, asNum(inv.depositValue, 0)), 0, totalMinor) as MoneyMinor
}

function calcBalanceDueMinor(
    totalMinor: MoneyMinor,
    depositMinor: MoneyMinor,
    paidMinor: MoneyMinor,
): MoneyMinor {
    return Math.max(0, totalMinor - depositMinor - asNum(paidMinor, 0)) as MoneyMinor
}

function fmtPrettyInvoiceNumber(prefix: string, baseNumber?: number): string {
    if (!baseNumber || baseNumber <= 0) return ''
    return `${prefix} - ${baseNumber}`
}

// Store
export const useInvoiceStore = defineStore('invoice', () => {
    const invoice = ref<Invoice | null>(null)
    const serverFieldErrors = ref<Record<string, string>>({})
    const showAllValidation = ref(false)
    const invoicePrefix = import.meta.env.VITE_INVOICE_PREFIX
    const verifyStatus = ref<'idle' | 'checking' | 'ok' | 'mismatch' | 'invalid' | 'error'>('idle')
    const lastVerifyAt = ref<number | null>(null)
    const serverCanonicalTotals = ref<Totals | null>(null)
    const serverCanonicalLineTotals = ref<Record<number, MoneyMinor>>({})

    // Called inside functions only never at module scope
    const getClientStore = () => useClientStore()

    function ensure(): Invoice {
        if (!invoice.value) throw new Error('Invoice not initialised')
        return invoice.value
    }

    const prettyBaseNumber = computed(() =>
        fmtPrettyInvoiceNumber(invoicePrefix, invoice.value?.baseNumber),
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
        } catch (err: unknown) {
            // background verification should not spam toasts
            if (
                (err as any)?.name === 'NetworkError' &&
                (err as any)?.message === 'Request aborted'
            ) {
                return
            }
            if (isApiError(err) && err.code === 'VALIDATION_FAILED') {
                serverFieldErrors.value = toFieldErrorMap(err.fields)
                verifyStatus.value = 'invalid'
                return
            }

            verifyStatus.value = 'error'
            console.error(err)
        }
    }

    function scheduleServerVerify() {
        if (typeof window === 'undefined') return
        if (!invoice.value) return

        clearVerifyTimer()
        verifyTimer = window.setTimeout(() => {
            runServerVerify()
        }, 450)
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

    const liveFieldErrors = computed<Record<string, string>>(() => {
        const inv = invoice.value
        if (!inv) return {}
        const dto = apiDTO(inv)
        return validateInvoicePayload(dto)
    })

    watch(
        invoice,
        () => {
            if (Object.keys(serverFieldErrors.value).length > 0) {
                serverFieldErrors.value = {}
            }
            showAllValidation.value = false
            scheduleServerVerify()
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

    // Lines CRUD
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
    }

    function updateLine(sortOrder: number, patch: Partial<InvoiceLine>): void {
        const inv = ensure()
        const line = inv.lines.find((l) => l.sortOrder === sortOrder)
        if (!line) return
        assignDefined(line, patch)
    }

    function removeLine(sortOrder: number): void {
        const inv = ensure()
        inv.lines = inv.lines
            .filter((l) => l.sortOrder !== sortOrder)
            .sort((a, b) => a.sortOrder - b.sortOrder)
            .map((l, i) => ({ ...l, sortOrder: i + 1 }))
    }

    // Setters — all validated/clamped here so components stay dumb
    function setIssueDate(v: string): void {
        ensure().issueDate = String(v ?? '')
    }

    function setDueByDate(v: string): void {
        ensure().dueByDate = String(v ?? '')
    }

    function setNote(note: string): void {
        ensure().note = String(note ?? '')
    }

    function setVatRateBps(v: number): void {
        ensure().vatRate = clamp(round0(asNum(v, 0)), 0, 10000)
    }

    function setDiscountType(t: DiscountType): void {
        const inv = ensure()
        inv.discountType = t
        if (t === 'none') inv.discountValue = 0
        if (t === 'percent')
            inv.discountValue = clamp(round0(asNum(inv.discountValue, 0)), 0, 10000)
        if (t === 'fixed') inv.discountValue = Math.max(0, round0(asNum(inv.discountValue, 0)))
    }

    function setDiscountFixedGBP(gbp: number): void {
        const inv = ensure()
        inv.discountType = 'fixed'
        inv.discountValue = Math.max(0, toMinor(gbp))
    }

    function setDiscountPercent(percent: number): void {
        const inv = ensure()
        inv.discountType = 'percent'
        inv.discountValue = clamp(round0(asNum(percent, 0) * 100), 0, 10000)
    }

    function clearDiscount(): void {
        const inv = ensure()
        inv.discountType = 'none'
        inv.discountValue = 0
    }

    function setDepositType(t: DepositType): void {
        const inv = ensure()
        inv.depositType = t
        if (t === 'none') inv.depositValue = 0
        if (t === 'percent') inv.depositValue = clamp(round0(asNum(inv.depositValue, 0)), 0, 10000)
        if (t === 'fixed') inv.depositValue = Math.max(0, round0(asNum(inv.depositValue, 0)))
    }

    function setDepositFixedGBP(gbp: number): void {
        const inv = ensure()
        inv.depositType = 'fixed'
        inv.depositValue = Math.max(0, toMinor(gbp))
    }

    function setDepositPercent(percent: number): void {
        const inv = ensure()
        inv.depositType = 'percent'
        inv.depositValue = clamp(round0(asNum(percent, 0) * 100), 0, 10000)
    }

    function clearDeposit(): void {
        const inv = ensure()
        inv.depositType = 'none'
        inv.depositValue = 0
    }

    function setPaidGBP(gbp: number): void {
        ensure().paidMinor = Math.max(0, toMinor(gbp))
    }

    // API DTO
    function apiDTO(inv: Invoice) {
        const prices = pricing.value

        return {
            overview: {
                clientId: inv.clientId,
                baseNumber: inv.baseNumber,
                clientName: inv.clientSnapshot.name,
                clientCompanyName: inv.clientSnapshot.companyName,
                clientAddress: inv.clientSnapshot.address,
                clientEmail: inv.clientSnapshot.email,
                note: inv.note,
                issueDate: inv.issueDate,
                ...(inv.dueByDate ? { dueByDate: inv.dueByDate } : {}),
            },
            lines: inv.lines.map((l) => ({
                productId: l.productId ?? null,
                name: l.name,
                lineType: l.lineType ?? 'custom',
                pricingMode: l.pricingMode,
                quantity: asNum(l.quantity, 1),
                minutesWorked: l.pricingMode === 'hourly' ? asNum(l.minutesWorked, 0) : null,
                unitPriceMinor: asNum(l.unitPriceMinor, 0),
                lineTotalMinor: lineTotalMinor(l),
                sortOrder: l.sortOrder,
            })),
            totals: {
                depositType: inv.depositType,
                depositMinor: prices?.depositMinor ?? 0,

                discountType: inv.discountType,
                discountMinor: prices?.totals.discountMinor ?? 0,

                vatRate: inv.vatRate,
                vatMinor: prices?.totals.vatMinor ?? 0,

                paidMinor: inv.paidMinor,

                subtotalMinor: prices?.totals.subtotalMinor ?? 0,
                subtotalAfterDiscountMinor: prices?.totals.subtotalAfterDiscountMinor ?? 0,
                totalMinor: prices?.totals.totalMinor ?? 0,
                balanceDueMinor: prices?.balanceDueMinor ?? 0,
            },
        }
    }

    async function newDraftInvoice(inv: Invoice) {
        const { lsClientId } = getClientStore()
        if (!lsClientId) throw new Error('No client selected')

        const dto = apiDTO(inv)
        showAllValidation.value = true
        serverFieldErrors.value = {}
        try {
            const result = await newInvoiceHandler(dto.overview.clientId, inv.baseNumber, dto)
            showAllValidation.value = false
            console.log(result)
            return result
        } catch (err: unknown) {
            if (isApiError(err)) {
                serverFieldErrors.value = toFieldErrorMap(err.fields)
            }
            throw err
        }
    }

    function getFieldError(field: string): string | null {
        return liveFieldErrors.value[field] ?? serverFieldErrors.value[field] ?? null
    }

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

        // helpers exported(again) for components
        lineTotalMinor,
        toMinor,
        fromMinor,
        fmtGBPMinor,

        // API
        apiDTO,
        newDraftInvoice,
    }
})

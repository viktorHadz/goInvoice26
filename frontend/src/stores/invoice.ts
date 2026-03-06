import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import type {
    Invoice,
    InvoiceLine,
    Totals,
    MoneyMinor,
    DepositType,
    DiscountType,
    InvoiceStatus,
} from '@/components/invoice/invoiceTypes'
import { newInvoiceHandler, getNewInvoiceNumber } from '@/utils/invoiceHttpHandler'
import { useClientStore } from './clients'

const clientStore = useClientStore()

/**
 * Goal: 0 drift (as far as is realistic in JS)
 * Rules:
 * - All money is integer minor units (pence)
 * - Percentages are basis points (bps) 0..10000
 * - Time is minutes; hourly lines use integer math: round(qty * unit * minutes / 60)
 * - Clamp anything that can exceed bounds (discount <= subtotal, deposit <= total, etc.)
 */

// -----------------------------
// tiny, readable primitives
// -----------------------------
type Int = number
type Patch<T extends object> = Partial<{ [K in keyof T]: T[K] }>

const isFiniteNum = (n: unknown): n is number => typeof n === 'number' && Number.isFinite(n)
const asNum = (n: unknown, fallback = 0) => (isFiniteNum(n) ? n : fallback)

const clamp = (n: number, min: number, max: number) => Math.min(max, Math.max(min, n))

// Round-half-up to nearest integer (what Go's math.Round does too)
const round0 = (n: number): Int => Math.round(n)

// Integer-safe-ish multiply/divide helper for bps, still number but avoids floats early
const mulBpsRound = (baseMinor: MoneyMinor, bps: number): MoneyMinor => {
    const b = clamp(round0(asNum(bps, 0)), 0, 10000)
    return round0((baseMinor * b) / 10000) as MoneyMinor
}

const toMinor = (gbp: number): MoneyMinor => round0(asNum(gbp, 0) * 100) as MoneyMinor
const fromMinor = (minor: MoneyMinor): number => (asNum(minor, 0) as number) / 100

const fmtGBPMinor = (minor: MoneyMinor): string =>
    new Intl.NumberFormat('en-GB', { style: 'currency', currency: 'GBP' }).format(fromMinor(minor))

function assignDefined<T extends object>(target: T, patch: Patch<T>) {
    for (const k in patch) {
        const v = patch[k as keyof T]
        if (v !== undefined) (target as any)[k] = v
    }
    return target
}

// -----------------------------
// pricing + totals (pure)
// -----------------------------
function lineTotalMinor(l: InvoiceLine): MoneyMinor {
    const qty = asNum(l.quantity, 0)
    const unit = asNum(l.unitPriceMinor, 0)

    if (l.pricingMode === 'hourly') {
        const minutes = asNum(l.minutesWorked, 0)
        // integer-style computation (reduce drift):
        // total = round(qty * unit * minutes / 60)
        return round0((qty * unit * minutes) / 60) as MoneyMinor
    }

    // flat: total = round(qty * unit)
    return round0(qty * unit) as MoneyMinor
}

function calcTotals(inv: Invoice): Totals {
    const subtotalMinor = inv.lines.reduce(
        (sum, l) => (sum + lineTotalMinor(l)) as MoneyMinor,
        0 as MoneyMinor,
    )

    const discountMinor: MoneyMinor = (() => {
        if (inv.discountType === 'none') return 0 as MoneyMinor

        if (inv.discountType === 'fixed') {
            const fixed = asNum(inv.discountValue, 0)
            // discountValue for fixed is already minor units
            return clamp(fixed, 0, subtotalMinor) as MoneyMinor
        }

        // percent (basis points)
        const bps = asNum(inv.discountValue, 0)
        const raw = mulBpsRound(subtotalMinor, bps)
        return clamp(raw, 0, subtotalMinor) as MoneyMinor
    })()

    const subtotalAfterDiscountMinor = Math.max(0, subtotalMinor - discountMinor) as MoneyMinor

    const vatMinor = mulBpsRound(subtotalAfterDiscountMinor, asNum(inv.vatRate, 0))
    const totalMinor = (subtotalAfterDiscountMinor + vatMinor) as MoneyMinor

    return {
        subtotalMinor,
        discountMinor,
        subtotalAfterDiscountMinor,
        vatMinor,
        totalMinor,
    }
}

function calcDepositMinor(inv: Invoice, totalMinor: MoneyMinor): MoneyMinor {
    if (inv.depositType === 'none') return 0 as MoneyMinor

    if (inv.depositType === 'fixed') {
        const fixed = asNum(inv.depositValue, 0) // already minor units
        return clamp(fixed, 0, totalMinor) as MoneyMinor
    }

    // percent (bps)
    const raw = mulBpsRound(totalMinor, asNum(inv.depositValue, 0))
    return clamp(raw, 0, totalMinor) as MoneyMinor
}

function calcBalanceDueMinor(
    totalMinor: MoneyMinor,
    depositMinor: MoneyMinor,
    paidMinor: MoneyMinor,
) {
    const paid = asNum(paidMinor, 0)
    return Math.max(0, totalMinor - depositMinor - paid) as MoneyMinor
}

function fmtPrettyInvoiceNumber(prefix: string, baseNumber?: number) {
    if (!baseNumber || baseNumber <= 0) return ''
    return `${prefix} - ${baseNumber}`
}

// -----------------------------
// Store
// -----------------------------
export const useInvoiceStore = defineStore('invoice', () => {
    console.log(clientStore)
    const invoice = ref<Invoice | null>(null)
    const invoicePrefix = import.meta.env.VITE_INVOICE_PREFIX

    function ensure(): Invoice {
        if (!invoice.value) throw new Error('Invoice not initialised')
        return invoice.value
    }

    const prettyBaseNumber = computed(() =>
        fmtPrettyInvoiceNumber(invoicePrefix, invoice.value?.baseNumber),
    )

    // Computed totals snapshot
    const pricing = computed(() => {
        const inv = invoice.value
        if (!inv) return null

        const totals = calcTotals(inv)
        const depositMinor = calcDepositMinor(inv, totals.totalMinor)
        const balanceDueMinor = calcBalanceDueMinor(totals.totalMinor, depositMinor, inv.paidMinor)

        return { totals, depositMinor, balanceDueMinor }
    })

    const totals = computed<Totals | null>(() => pricing.value?.totals ?? null)

    const depositMinor = computed<MoneyMinor>(
        () => pricing.value?.depositMinor ?? (0 as MoneyMinor),
    )

    const balanceDueMinor = computed<MoneyMinor>(
        () => pricing.value?.balanceDueMinor ?? (0 as MoneyMinor),
    )

    async function setInvoice(newInvoiceData: Invoice) {
        const clientId = clientStore.selectedClientId
        if (!clientId) throw new Error('No client selected')

        const invoiceBaseNumber = await getNewInvoiceNumber(clientId)
        invoice.value = { ...newInvoiceData, baseNumber: invoiceBaseNumber }
    }

    // -----------------------------
    // Lines CRUD
    // -----------------------------
    const linesMap = new Map()
    function addLine(line: Omit<InvoiceLine, 'sortOrder'>) {
        const inv = ensure()
        const maxSort = inv.lines.reduce((m, l) => Math.max(m, asNum(l.sortOrder, 0)), 0)
        // HERE
        // look through all lines and check if line exists
        // use a for quick check
        // need a map
        inv.lines.forEach((line, idx) => {
            linesMap.set(line.id, line.lineType)
        })
        inv.lines.push({ ...line, sortOrder: maxSort + 1 })
    }

    function updateLine(sortOrder: number, patch: Patch<InvoiceLine>) {
        const inv = ensure()
        const line = inv.lines.find((l) => l.sortOrder === sortOrder)
        if (!line) return
        assignDefined(line, patch)
    }

    function removeLine(sortOrder: number) {
        const inv = ensure()

        const kept = inv.lines
            .filter((l) => l.sortOrder !== sortOrder)
            .sort((a, b) => a.sortOrder - b.sortOrder)

        inv.lines = kept.map((l, i) => ({ ...l, sortOrder: i + 1 }))
    }

    // -----------------------------
    // Adjustment setters (clamped)
    // -----------------------------
    function setVatRateBps(v: number) {
        const inv = ensure()
        inv.vatRate = clamp(round0(asNum(v, 0)), 0, 10000)
    }

    function setDiscountType(t: DiscountType) {
        const inv = ensure()
        inv.discountType = t

        if (t === 'none') inv.discountValue = 0
        if (t === 'percent')
            inv.discountValue = clamp(round0(asNum(inv.discountValue, 0)), 0, 10000)
        if (t === 'fixed') inv.discountValue = Math.max(0, round0(asNum(inv.discountValue, 0)))
    }

    function setDiscountFixedGBP(gbp: number) {
        const inv = ensure()
        inv.discountType = 'fixed'
        inv.discountValue = Math.max(0, toMinor(gbp))
    }

    function setDiscountPercent(percent: number) {
        const inv = ensure()
        inv.discountType = 'percent'
        const bps = round0(asNum(percent, 0) * 100) // 12.34% => 1234 bps
        inv.discountValue = clamp(bps, 0, 10000)
    }

    function clearDiscount() {
        const inv = ensure()
        inv.discountType = 'none'
        inv.discountValue = 0
    }

    function setPaidGBP(gbp: number) {
        const inv = ensure()
        inv.paidMinor = Math.max(0, toMinor(gbp))
    }

    function setDepositType(t: DepositType) {
        const inv = ensure()
        inv.depositType = t

        if (t === 'none') inv.depositValue = 0
        if (t === 'percent') inv.depositValue = clamp(round0(asNum(inv.depositValue, 0)), 0, 10000)
        if (t === 'fixed') inv.depositValue = Math.max(0, round0(asNum(inv.depositValue, 0)))
    }

    function setDepositFixedGBP(gbp: number) {
        const inv = ensure()
        inv.depositType = 'fixed'
        inv.depositValue = Math.max(0, toMinor(gbp))
    }

    function setDepositPercent(percent: number) {
        const inv = ensure()
        inv.depositType = 'percent'
        const bps = round0(asNum(percent, 0) * 100)
        inv.depositValue = clamp(bps, 0, 10000)
    }

    function clearDeposit() {
        const inv = ensure()
        inv.depositType = 'none'
        inv.depositValue = 0
    }

    function setIssueDate(v: string) {
        ensure().issueDate = String(v ?? '')
    }

    function setDueByDate(v: string) {
        ensure().dueByDate = String(v ?? '')
    }

    function setNote(note: string) {
        ensure().note = String(note ?? '')
    }

    // API Shape
    function apiDTO(inv: Invoice): any {
        const prices = pricing.value
        const clientId = clientStore.selectedClientId
        const overview = {
            clientId: inv.clientId,
            currentRevisionId: 1,
            baseNumber: inv.baseNumber,
            status: inv.status,
            clientName: inv.clientSnapshot.name,
            clientCompanyName: inv.clientSnapshot.companyName,
            clientAddress: inv.clientSnapshot.address,
            clientEmail: inv.clientSnapshot.email,
            note: inv.note,
            issueDate: inv.issueDate,
            dueByDate: inv.dueByDate,
        }

        const line = inv.lines.map((line) => ({
            revisionId: 1,
            productId: line.productId || null,
            name: line.name,
            lineType: line.lineType || 'custom',
            pricingMode: line.pricingMode,
            quantity: asNum(line.quantity, 1),
            minutesWorked: line.pricingMode === 'hourly' ? asNum(line.minutesWorked, 0) : null,
            unitPriceMinor: asNum(line.unitPriceMinor, 0),
            lineSubtotalMinor: lineTotalMinor(line),
        }))

        const totals = {
            depositType: inv.depositType,
            depositMinor: inv.depositValue,

            discountType: inv.discountType,
            discountMinor: inv.discountValue,

            vatRate: inv.vatRate,
            vatMinor: prices?.totals.vatMinor,

            subtotalAfterDiscountMinor: prices?.totals.subtotalAfterDiscountMinor,
            balanceDueMinor: prices?.balanceDueMinor,
            subtotalMinor: prices?.totals.subtotalMinor,
            totalMinor: prices?.totals.totalMinor,
        }

        const DTO = { overview, line, totals }
        return DTO
    }

    async function newDraftInvoice(inv: Invoice) {
        const dto = apiDTO(inv)
        const clientId = clientStore.selectedClientId
        if (!clientId) throw new Error('No client selected')

        const validatedInvoice = await newInvoiceHandler(dto.clientId, dto.baseNumber, dto)
        console.log(validatedInvoice)

        return validatedInvoice
    }
    return {
        // state
        invoice,
        prettyBaseNumber,

        // init
        setInvoice,

        // derived
        totals,
        depositMinor,
        balanceDueMinor,

        // pricing helpers
        lineTotalMinor,

        // lines
        addLine,
        updateLine,
        removeLine,

        // adjustments
        setVatRateBps,
        setDiscountType,
        setDiscountFixedGBP,
        setDiscountPercent,
        clearDiscount,
        setPaidGBP,
        setDepositType,
        setDepositFixedGBP,
        setDepositPercent,
        clearDeposit,
        setIssueDate,
        setDueByDate,
        setNote,

        // display/helpers
        toMinor,
        fromMinor,
        fmtGBPMinor,

        apiDTO,
        newDraftInvoice,
    }
})

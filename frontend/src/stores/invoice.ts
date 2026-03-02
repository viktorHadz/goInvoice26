import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import type {
    Invoice,
    InvoiceLine,
    Totals,
    MoneyMinor,
    DepositType,
    DiscountType,
} from '@/components/invoice/invoiceTypes'

// Internal helpers centralised here
type Patch<T extends object> = Partial<{ [K in keyof T]: T[K] }>

function assignDefined<T extends object>(target: T, patch: Patch<T>) {
    for (const k in patch) {
        const v = patch[k as keyof T]
        if (v !== undefined) (target as any)[k] = v
    }
    return target
}

function roundInt(n: number): MoneyMinor {
    return Math.round(n) as MoneyMinor
}
function toMinor(gbp: number): MoneyMinor {
    if (!Number.isFinite(gbp)) return 0
    return Math.round(gbp * 100) as MoneyMinor
}
function fromMinor(minor: MoneyMinor): number {
    return (minor ?? 0) / 100
}
function fmtGBPMinor(minor: MoneyMinor): string {
    return new Intl.NumberFormat('en-GB', { style: 'currency', currency: 'GBP' }).format(
        (minor ?? 0) / 100,
    )
}

function lineTotalMinor(l: InvoiceLine): MoneyMinor {
    const qty = Number.isFinite(l.quantity) ? l.quantity : 0
    const unit = Number.isFinite(l.unitPriceMinor) ? l.unitPriceMinor : 0

    if (l.pricingMode === 'hourly') {
        const minutes = Number.isFinite(l.minutesWorked ?? NaN) ? (l.minutesWorked ?? 0) : 0
        const hours = minutes / 60
        return roundInt(qty * unit * hours)
    }

    return roundInt(qty * unit)
}

function calcTotals(inv: Invoice): Totals {
    const subtotalMinor = inv.lines.reduce((sum, l) => sum + lineTotalMinor(l), 0 as MoneyMinor)

    const discountMinor =
        inv.discountType === 'none'
            ? (0 as MoneyMinor)
            : inv.discountType === 'fixed'
              ? (Math.min(inv.discountValue, subtotalMinor) as MoneyMinor)
              : (Math.min(
                    subtotalMinor,
                    roundInt(subtotalMinor * (inv.discountValue / 10000)),
                ) as MoneyMinor)

    const afterDiscountMinor = Math.max(0, subtotalMinor - discountMinor) as MoneyMinor
    const vatMinor = roundInt(afterDiscountMinor * (inv.vatRate / 10000))
    const totalMinor = (afterDiscountMinor + vatMinor) as MoneyMinor

    return { subtotalMinor, discountMinor, afterDiscountMinor, vatMinor, totalMinor }
}

// Store
export const useInvoiceStore = defineStore('invoice', () => {
    const invoice = ref<Invoice | null>(null)

    function setInvoice(next: Invoice) {
        invoice.value = next
    }

    function ensure(): Invoice {
        if (!invoice.value) throw new Error('Invoice not initialised')
        return invoice.value
    }

    // Single source

    const totals = computed<Totals | null>(() => (invoice.value ? calcTotals(invoice.value) : null))

    const depositMinor = computed<MoneyMinor>(() => {
        const inv = invoice.value
        const t = totals.value
        if (!inv || !t) return 0

        if (inv.depositType === 'none') return 0

        if (inv.depositType === 'fixed') {
            const fixed = Number.isFinite(inv.depositValue) ? inv.depositValue : 0
            return Math.max(0, Math.min(fixed, t.totalMinor)) as MoneyMinor
        }

        // percent in basis points
        const bps = Number.isFinite(inv.depositValue) ? inv.depositValue : 0
        const raw = roundInt(t.totalMinor * (bps / 10000))
        return Math.max(0, Math.min(raw, t.totalMinor)) as MoneyMinor
    })

    const balanceDueMinor = computed<MoneyMinor>(() => {
        const inv = invoice.value
        const t = totals.value
        if (!inv || !t) return 0

        const paid = Number.isFinite(inv.paidMinor) ? inv.paidMinor : 0
        return Math.max(0, t.totalMinor - depositMinor.value - paid) as MoneyMinor
    })

    // Lines CRUD

    function addLine(line: Omit<InvoiceLine, 'sortOrder'>) {
        const inv = ensure()
        const max = inv.lines.reduce((m, l) => Math.max(m, l.sortOrder), 0)
        inv.lines.push({ ...line, sortOrder: max + 1 })
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

    // Adjustment setters
    function setVatRateBps(v: number) {
        const inv = ensure()
        const next = Number.isFinite(v) ? Math.round(v) : 0
        inv.vatRate = Math.max(0, Math.min(10000, next))
    }

    function setDiscountType(t: DiscountType) {
        const inv = ensure()
        inv.discountType = t
        if (t === 'none') inv.discountValue = 0 // keep value but clamp based on mode
        if (t === 'percent') inv.discountValue = Math.max(0, Math.min(10000, inv.discountValue))
        if (t === 'fixed') inv.discountValue = Math.max(0, inv.discountValue)
    }

    function setDiscountFixedGBP(gbp: number) {
        const inv = ensure()
        inv.discountType = 'fixed'
        inv.discountValue = Math.max(0, toMinor(gbp))
    }

    function setDiscountPercent(percent: number) {
        const inv = ensure()
        inv.discountType = 'percent'
        const bps = Math.round((Number.isFinite(percent) ? percent : 0) * 100)
        inv.discountValue = Math.max(0, Math.min(10000, bps))
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
        if (t === 'percent') inv.depositValue = Math.max(0, Math.min(10000, inv.depositValue))
        if (t === 'fixed') inv.depositValue = Math.max(0, inv.depositValue)
    }

    function setDepositFixedGBP(gbp: number) {
        const inv = ensure()
        inv.depositType = 'fixed'
        inv.depositValue = Math.max(0, toMinor(gbp))
    }

    function setDepositPercent(percent: number) {
        const inv = ensure()
        inv.depositType = 'percent'
        const bps = Math.round((Number.isFinite(percent) ? percent : 0) * 100)
        inv.depositValue = Math.max(0, Math.min(10000, bps))
    }

    function clearDeposit() {
        const inv = ensure()
        inv.depositType = 'none'
        inv.depositValue = 0
    }

    function setIssueDate(v: string) {
        const inv = ensure()
        inv.issueDate = String(v ?? '')
    }

    function setDueByDate(v: string) {
        const inv = ensure()
        inv.dueByDate = String(v ?? '')
    }

    function setClientEmail(v: string) {
        const inv = ensure()
        inv.clientSnapshot.email = String(v ?? '')
    }

    function setNote(note: string) {
        const inv = ensure()
        inv.note = String(note ?? '')
    }

    return {
        // state
        invoice,
        // init
        setInvoice,
        // derived
        totals,
        depositMinor,
        balanceDueMinor,
        // lines
        addLine,
        lineTotalMinor,
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

        setClientEmail,
        setNote,
        // helpers
        toMinor,
        fromMinor,
        fmtGBPMinor,
    }
})

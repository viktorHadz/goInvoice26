import type {
    Invoice,
    InvoiceLine,
    DepositType,
    DiscountType,
} from '@/components/invoice/invoiceTypes'
import { asNum, round0 } from '@/utils/numbers'
import { clamp } from '@vueuse/core'
import { toMinor } from '@/utils/money'

function assignDefined<T extends object>(target: T, patch: Partial<T>) {
    for (const k in patch) {
        const v = patch[k as keyof T]
        if (v !== undefined) (target as any)[k] = v
    }
    return target
}

//*---------------------------------
// * CRUD Logic
//*---------------------------------
export function addInvoiceLine(inv: Invoice, line: Omit<InvoiceLine, 'sortOrder'>): void {
    const canMerge = line.lineType !== 'custom' && line.productId != null
    const existingLine = canMerge
        ? inv.lines.find((ln) => ln.productId === line.productId)
        : undefined

    if (existingLine) {
        const qtyToAdd = Number.isFinite(line.quantity) && line.quantity > 0 ? line.quantity : 1
        existingLine.quantity += qtyToAdd
        return
    }

    const maxSort = inv.lines.reduce((m, current) => Math.max(m, asNum(current.sortOrder, 0)), 0)
    inv.lines.push({ ...line, sortOrder: maxSort + 1 })
}

export function updateInvoiceLine(
    inv: Invoice,
    sortOrder: number,
    patch: Partial<InvoiceLine>,
): void {
    const line = inv.lines.find((l) => l.sortOrder === sortOrder)
    if (!line) return
    assignDefined(line, patch)
}

export function removeInvoiceLine(inv: Invoice, sortOrder: number): void {
    inv.lines = inv.lines
        .filter((l) => l.sortOrder !== sortOrder)
        .sort((a, b) => a.sortOrder - b.sortOrder)
        .map((l, i) => ({ ...l, sortOrder: i + 1 }))
}

export function setInvoiceIssueDate(inv: Invoice, v: string): void {
    inv.issueDate = String(v ?? '')
}
export function setInvoiceDueByDate(inv: Invoice, v: string): void {
    inv.dueByDate = String(v ?? '')
}

export function setInvoiceNote(inv: Invoice, note: string): void {
    inv.note = String(note ?? '')
}

export function setInvoiceVatRateBps(inv: Invoice, v: number): void {
    inv.vatRate = clamp(round0(asNum(v, 0)), 0, 10000)
}
//*---------------------------------
// * Discount Logic
//*---------------------------------
export function setInvoiceDiscountType(inv: Invoice, t: DiscountType): void {
    inv.discountType = t

    if (t === 'none') {
        inv.discountMinor = 0
        inv.discountRate = 0
    } else if (t === 'percent') {
        inv.discountRate = clamp(round0(asNum(inv.discountRate, 0)), 0, 10000)
        inv.discountMinor = 0
    } else {
        inv.discountMinor = Math.max(0, round0(asNum(inv.discountMinor, 0)))
        inv.discountRate = 0
    }
}
export function setInvoiceDiscountFixedGBP(inv: Invoice, gbp: number): void {
    inv.discountType = 'fixed'
    inv.discountMinor = Math.max(0, toMinor(gbp))
    inv.discountRate = 0
}
export function setInvoiceDiscountPercent(inv: Invoice, percent: number): void {
    inv.discountType = 'percent'
    inv.discountRate = clamp(round0(asNum(percent, 0) * 100), 0, 10000)
    inv.discountMinor = 0
}
export function clearInvoiceDiscount(inv: Invoice): void {
    inv.discountType = 'none'
    inv.discountMinor = 0
    inv.discountRate = 0
}

//*---------------------------------
// * Deposit Logic
//*---------------------------------
export function setInvoiceDepositType(inv: Invoice, t: DepositType): void {
    inv.depositType = t

    if (t === 'none') {
        inv.depositMinor = 0
        inv.depositRate = 0
    } else if (t === 'percent') {
        inv.depositRate = clamp(round0(asNum(inv.depositRate, 0)), 0, 10000)
        inv.depositMinor = 0
    } else {
        inv.depositMinor = Math.max(0, round0(asNum(inv.depositMinor, 0)))
        inv.depositRate = 0
    }
}
export function setInvoiceDepositFixedGBP(inv: Invoice, gbp: number): void {
    inv.depositType = 'fixed'
    inv.depositMinor = Math.max(0, toMinor(gbp))
    inv.depositRate = 0
}
export function setInvoiceDepositPercent(inv: Invoice, percent: number): void {
    inv.depositType = 'percent'
    inv.depositRate = clamp(round0(asNum(percent, 0) * 100), 0, 10000)
    inv.depositMinor = 0
}
export function clearInvoiceDeposit(inv: Invoice): void {
    inv.depositType = 'none'
    inv.depositMinor = 0
    inv.depositRate = 0
}

//*---------------------------------
// * Payments Logic
//*---------------------------------
export function setInvoicePayment(inv: Invoice, amount: number): void {
    inv.paidMinor = Math.max(0, toMinor(amount))
}

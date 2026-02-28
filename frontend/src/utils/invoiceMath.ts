import type { InvoiceDraft, InvoiceLine } from '@/components/invoice/invoiceTypes'
import type { MoneyMinor } from '@/utils/money'

function roundMoney(n: number): MoneyMinor {
    return Math.round(n)
}

export function lineTotalMinor(l: InvoiceLine): MoneyMinor {
    const qty = Number.isFinite(l.quantity) ? l.quantity : 0
    const unit = Number.isFinite(l.unitPriceMinor) ? l.unitPriceMinor : 0

    if (l.pricingMode === 'hourly') {
        const minutes = l.minutesWorked ?? 0
        const hours = minutes / 60
        return roundMoney(qty * unit * hours)
    }

    return roundMoney(qty * unit)
}

export function calcInvoiceTotals(d: InvoiceDraft) {
    const subtotalMinor = d.lines.reduce((sum, l) => sum + lineTotalMinor(l), 0 as MoneyMinor)

    const discountMinor =
        d.discountType === 'none'
            ? 0
            : d.discountType === 'fixed'
              ? Math.min(d.discountValue, subtotalMinor)
              : Math.min(subtotalMinor, roundMoney(subtotalMinor * (d.discountValue / 10000)))

    const afterDiscountMinor = Math.max(0, subtotalMinor - discountMinor)
    const vatMinor = roundMoney(afterDiscountMinor * (d.vatRate / 10000))
    const totalMinor = afterDiscountMinor + vatMinor

    return {
        subtotalMinor,
        discountMinor,
        afterDiscountMinor,
        vatMinor,
        totalMinor,
    }
}

import type { Invoice, InvoiceLine, MoneyMinor, Totals } from '@/components/invoice/invoiceTypes'
import { asNum, multiplyAndRoundBps, round0 } from './numbers'
import { clamp } from '@vueuse/core'

export const toMinor = (gbp: number): MoneyMinor => round0(asNum(gbp, 0) * 100) as MoneyMinor
export const fromMinor = (minor: MoneyMinor): number => (asNum(minor, 0) as number) / 100
export const fmtGBPMinor = (minor: MoneyMinor): string =>
    new Intl.NumberFormat('en-GB', { style: 'currency', currency: 'GBP' }).format(fromMinor(minor))
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
            return clamp(asNum(inv.discountMinor, 0), 0, subtotalMinor) as MoneyMinor
        }

        return clamp(
            multiplyAndRoundBps(subtotalMinor, inv.discountRate),
            0,
            subtotalMinor,
        ) as MoneyMinor
    })()

    const subtotalAfterDiscountMinor = Math.max(0, subtotalMinor - discountMinor) as MoneyMinor
    const vatMinor = multiplyAndRoundBps(subtotalAfterDiscountMinor, asNum(inv.vatRate, 0))
    const totalMinor = (subtotalAfterDiscountMinor + vatMinor) as MoneyMinor

    return {
        subtotalMinor,
        discountMinor,
        subtotalAfterDiscountMinor,
        vatMinor,
        totalMinor,
    }
}

export function calcDepositMinor(inv: Invoice, totalMinor: MoneyMinor): MoneyMinor {
    if (inv.depositType === 'none') return 0 as MoneyMinor
    if (inv.depositType === 'fixed') {
        return clamp(asNum(inv.depositMinor, 0), 0, totalMinor) as MoneyMinor
    }
    return clamp(multiplyAndRoundBps(totalMinor, inv.depositRate), 0, totalMinor) as MoneyMinor
}
export function calcBalanceDueMinor(
    totalMinor: MoneyMinor,
    _depositMinor: MoneyMinor,
    paidMinor: MoneyMinor,
): MoneyMinor {
    return Math.max(0, totalMinor - asNum(paidMinor, 0)) as MoneyMinor
}

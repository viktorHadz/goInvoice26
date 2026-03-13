import type { Invoice } from '@/components/invoice/invoiceTypes'
import { asNum } from '@/utils/numbers'
import { calcBalanceDueMinor, calcDepositMinor, calcTotals, lineTotalMinor } from '@/utils/money'

export function apiDTO(inv: Invoice) {
    const totals = calcTotals(inv)
    const depositMinor = calcDepositMinor(inv, totals.totalMinor)
    const balanceDueMinor = calcBalanceDueMinor(totals.totalMinor, depositMinor, inv.paidMinor)

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
            depositRate: inv.depositRate,
            depositMinor,

            discountType: inv.discountType,
            discountRate: inv.discountRate,
            discountMinor: totals.discountMinor,

            vatRate: inv.vatRate,
            vatMinor: totals.vatMinor,

            paidMinor: inv.paidMinor,

            subtotalMinor: totals.subtotalMinor,
            subtotalAfterDiscountMinor: totals.subtotalAfterDiscountMinor,
            totalMinor: totals.totalMinor,
            balanceDueMinor,
        },
    }
}

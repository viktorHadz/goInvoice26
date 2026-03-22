import type { Invoice } from '@/components/invoice/invoiceTypes'
import { toRaw } from 'vue'

/**
 * Provides a copy of the values inside a type Invoice
 * see usage inside edit stroe - editor.ts
 */
export function cloneInvoice(invoice: Invoice): Invoice {
    const src = toRaw(invoice)

    return {
        baseNumber: src.baseNumber,
        clientId: src.clientId,
        status: src.status,
        issueDate: src.issueDate,
        dueByDate: src.dueByDate,
        clientSnapshot: {
            name: src.clientSnapshot.name,
            companyName: src.clientSnapshot.companyName,
            address: src.clientSnapshot.address,
            email: src.clientSnapshot.email,
        },
        lines: src.lines.map((l) => ({
            productId: l.productId,
            name: l.name,
            lineType: l.lineType,
            pricingMode: l.pricingMode,
            quantity: l.quantity,
            unitPriceMinor: l.unitPriceMinor,
            minutesWorked: l.minutesWorked,
            sortOrder: l.sortOrder,
        })),
        discountType: src.discountType,
        discountMinor: src.discountMinor,
        discountRate: src.discountRate,
        vatRate: src.vatRate,
        paidMinor: src.paidMinor,
        depositType: src.depositType,
        depositMinor: src.depositMinor,
        depositRate: src.depositRate,
        note: src.note,
    }
}

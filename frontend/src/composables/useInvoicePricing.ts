import { computed, type Ref } from 'vue'
import type { Invoice, Totals, MoneyMinor } from '@/components/invoice/invoiceTypes'
import { calcTotals, calcDepositMinor, calcBalanceDueMinor } from '@/utils/money'

export type InvoicePricing = {
    totals: Totals
    depositMinor: MoneyMinor
    balanceDueMinor: MoneyMinor
}

export function useInvoicePricing(invoice: Ref<Invoice | null>) {
    const pricing = computed<InvoicePricing | null>(() => {
        const inv = invoice.value
        if (!inv) return null

        const totals = calcTotals(inv)
        const depositMinor = calcDepositMinor(inv, totals.totalMinor)
        const balanceDueMinor = calcBalanceDueMinor(totals.totalMinor, depositMinor, inv.paidMinor)

        return { totals, depositMinor, balanceDueMinor }
    })

    const totals = computed<Totals | null>(() => pricing.value?.totals ?? null)
    const depositMinor = computed<MoneyMinor>(() => pricing.value?.depositMinor ?? 0)
    const balanceDueMinor = computed<MoneyMinor>(() => pricing.value?.balanceDueMinor ?? 0)

    return {
        pricing,
        totals,
        depositMinor,
        balanceDueMinor,
    }
}

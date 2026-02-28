import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { calcInvoiceTotals } from '@/utils/invoiceMath'
import type { InvoiceDraft, InvoiceLine } from '@/components/invoice/invoiceTypes'

// Patch may omits keys but values are the real types not undefined
export type Patch<T extends object> = Partial<{ [K in keyof T]: T[K] }>

// Runtime helper: ignore keys when their value is undefined
function assignDefined<T extends object>(target: T, patch: Patch<T>) {
    for (const k in patch) {
        const v = patch[k as keyof T]
        if (v !== undefined) {
            ;(target as any)[k] = v
        }
    }
    return target
}

export const useInvoiceDraftStore = defineStore('invoiceDraft', () => {
    const draft = ref<InvoiceDraft | null>(null)

    function setDraft(next: InvoiceDraft) {
        draft.value = next
    }

    function ensure() {
        return draft.value
    }

    function addLine(line: Omit<InvoiceLine, 'sortOrder'>) {
        const d = ensure()
        if (!d) return

        const max = d.lines.reduce((m, l) => Math.max(m, l.sortOrder), 0)
        d.lines.push({ ...line, sortOrder: max + 1 })
    }

    function updateLine(sortOrder: number, patch: Patch<InvoiceLine>) {
        const d = ensure()
        if (!d) return

        const line = d.lines.find((l) => l.sortOrder === sortOrder)
        if (!line) return

        assignDefined(line, patch)
    }

    function removeLine(sortOrder: number) {
        const d = ensure()
        if (!d) return

        const kept = d.lines
            .filter((l) => l.sortOrder !== sortOrder)
            .sort((a, b) => a.sortOrder - b.sortOrder)

        // re-number
        d.lines = kept.map((l, i) => ({ ...l, sortOrder: i + 1 }))
    }

    const totals = computed(() => (draft.value ? calcInvoiceTotals(draft.value) : null))

    const balanceDueMinor = computed(() => {
        const d = draft.value
        const t = totals.value
        if (!d || !t) return 0
        return Math.max(0, t.totalMinor - d.paidMinor - d.depositMinor)
    })

    return {
        draft,
        setDraft,
        addLine,
        updateLine,
        removeLine,
        totals,
        balanceDueMinor,
    }
})

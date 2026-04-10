import type { ActiveEditorNode } from '@/components/editor/invBookTypes'
import { fmtPrettyInvoiceNumber } from '@/utils/numbers'

export function toDisplayRevisionNo(revisionNo?: number | null): number | null {
    if (revisionNo == null || revisionNo <= 1) return null
    return revisionNo
}

export function formatInvoiceBaseLabel(prefix: string, baseNumber?: number | null): string {
    return fmtPrettyInvoiceNumber(prefix, baseNumber ?? undefined)
}

export function formatInvoiceDisplayLabel(
    prefix: string,
    baseNumber: number | null | undefined,
    revisionNo?: number | null,
): string {
    const base = formatInvoiceBaseLabel(prefix, baseNumber)
    if (!base) return ''
    const displayRevisionNo = toDisplayRevisionNo(revisionNo)
    if (displayRevisionNo == null) return base
    return `${base}.${displayRevisionNo}`
}

export function formatPaymentReceiptLabel(
    prefix: string,
    baseNumber: number | null | undefined,
    revisionNo?: number | null,
    receiptNo?: number | null,
): string {
    const referenceLabel = formatInvoiceDisplayLabel(prefix, baseNumber, revisionNo)
    if (!referenceLabel || receiptNo == null || receiptNo < 1) return referenceLabel
    return `${referenceLabel}-PR-${receiptNo}`
}

export function formatActiveEditorNodeLabel(prefix: string, node: ActiveEditorNode | null): string {
    if (!node) return ''
    const revisionNo = node.type === 'revision' ? node.revisionNo : undefined
    return formatInvoiceDisplayLabel(prefix, node.baseNo, revisionNo)
}

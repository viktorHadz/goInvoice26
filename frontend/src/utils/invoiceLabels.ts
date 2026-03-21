import type { ActiveEditorNode } from '@/components/editor/invBookTypes'
import { fmtPrettyInvoiceNumber } from '@/utils/numbers'

export function formatInvoiceBaseLabel(
    prefix: string,
    baseNumber?: number | null,
): string {
    return fmtPrettyInvoiceNumber(prefix, baseNumber ?? undefined)
}

export function formatInvoiceDisplayLabel(
    prefix: string,
    baseNumber: number | null | undefined,
    revisionNo?: number | null,
): string {
    const base = formatInvoiceBaseLabel(prefix, baseNumber)
    if (!base) return ''
    if (revisionNo == null) return base
    return `${base}.${revisionNo}`
}

export function formatActiveEditorNodeLabel(
    prefix: string,
    node: ActiveEditorNode | null,
): string {
    if (!node) return ''
    const revisionNo = node.type === 'revision' ? node.revisionNo : undefined
    return formatInvoiceDisplayLabel(prefix, node.baseNo, revisionNo)
}

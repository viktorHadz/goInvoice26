import type { ActiveEditorNode } from '@/components/editor/invBookTypes'
import { fmtPrettyInvoiceNumber } from '@/utils/numbers'

/**
 * Display mapping contract for user-facing invoice numbering:
 * - DB revision 1 is the base invoice => no dotted suffix.
 * - DB revision N>1 maps to display suffix (N-1): base.1, base.2, ...
 */
export function toDisplayRevisionNo(revisionNo?: number | null): number | null {
    if (revisionNo == null || revisionNo <= 1) return null
    return revisionNo - 1
}

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
    const displayRevisionNo = toDisplayRevisionNo(revisionNo)
    if (displayRevisionNo == null) return base
    return `${base}.${displayRevisionNo}`
}

export function formatActiveEditorNodeLabel(
    prefix: string,
    node: ActiveEditorNode | null,
): string {
    if (!node) return ''
    const revisionNo = node.type === 'revision' ? node.revisionNo : undefined
    return formatInvoiceDisplayLabel(prefix, node.baseNo, revisionNo)
}

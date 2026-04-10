import type { MoneyMinor } from '@/components/invoice/invoiceTypes'

// Primitives
export type Int = number

export const isFiniteNum = (n: unknown): n is number => typeof n === 'number' && Number.isFinite(n)
export const asNum = (n: unknown, fallback = 0) => (isFiniteNum(n) ? n : fallback)
export const clamp = (n: number, min: number, max: number) => Math.min(max, Math.max(min, n))
export const round0 = (n: number): Int => Math.round(n)

export const multiplyAndRoundBps = (baseMinor: MoneyMinor, bps: number): MoneyMinor => {
    const b = clamp(round0(asNum(bps, 0)), 0, 10000)
    return round0((baseMinor * b) / 10000) as MoneyMinor
}
/** Implementation detail for pretty invoice numbers; prefer `@/utils/invoiceLabels` for user-facing UI. */
export function fmtPrettyInvoiceNumber(prefix: string, baseNumber?: number): string {
    if (!baseNumber || baseNumber <= 0) return ''
    const cleanPrefix = prefix.replace(/-\s*$/, '').trim()
    return cleanPrefix ? `${cleanPrefix}-${baseNumber}` : `${baseNumber}`
}

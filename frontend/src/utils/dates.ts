function pad2(v: number): string {
    return String(v).padStart(2, '0')
}

export type DateFormat = 'dd/mm/yyyy' | 'mm/dd/yyyy' | 'yyyy-mm-dd'

type ISODateParts = {
    year: number
    month: number
    day: number
}

function parseISODateParts(value: string): ISODateParts | null {
    if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) return null
    const values = value.split('-').map(Number)
    if (values.length !== 3) return null
    const [y, m, d] = values as [number, number, number]
    if (!Number.isFinite(y) || !Number.isFinite(m) || !Number.isFinite(d)) return null
    const local = new Date(y, m - 1, d)
    if (
        local.getFullYear() !== y ||
        local.getMonth() + 1 !== m ||
        local.getDate() !== d
    ) {
        return null
    }
    return { year: y, month: m, day: d }
}

export function isValidISODate(value: string | null | undefined): value is string {
    if (!value) return false
    return parseISODateParts(value) !== null
}

export function normalizeISODateOrNull(value: string | null | undefined): string | null {
    if (!value) return null
    const trimmed = value.trim()
    if (!trimmed) return null
    return isValidISODate(trimmed) ? trimmed : null
}

function formatParts(parts: ISODateParts, dateFormat: DateFormat): string {
    const year = String(parts.year)
    const month = pad2(parts.month)
    const day = pad2(parts.day)
    switch (dateFormat) {
        case 'mm/dd/yyyy':
            return `${month}/${day}/${year}`
        case 'yyyy-mm-dd':
            return `${year}-${month}-${day}`
        case 'dd/mm/yyyy':
        default:
            return `${day}/${month}/${year}`
    }
}

export function fmtDisplayDate(d: Date, dateFormat: DateFormat = 'dd/mm/yyyy') {
    return formatParts(
        { year: d.getFullYear(), month: d.getMonth() + 1, day: d.getDate() },
        dateFormat,
    )
}

export function fmtStrDate(dateStr: string, dateFormat: DateFormat = 'dd/mm/yyyy') {
    const parts = parseISODateParts(dateStr.trim())
    if (!parts) return dateStr
    return formatParts(parts, dateFormat)
}

export function fromISODate(v: string | null | undefined) {
    const normalized = normalizeISODateOrNull(v)
    if (!normalized) return null
    const values = normalized.split('-').map(Number)
    if (values.length !== 3) return null
    const [y, m, d] = values as [number, number, number]
    return new Date(y, m - 1, d)
}

export function toISODate(d: Date) {
    return `${d.getFullYear()}-${pad2(d.getMonth() + 1)}-${pad2(d.getDate())}`
}

export function todayISO() {
    const now = new Date()
    return `${now.getFullYear()}-${pad2(now.getMonth() + 1)}-${pad2(now.getDate())}`
}

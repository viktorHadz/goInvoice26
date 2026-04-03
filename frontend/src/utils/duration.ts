export function formatInvoiceDurationMinutes(minutes: number): string {
    const safeMinutes = Number.isFinite(minutes) ? Math.max(0, Math.trunc(minutes)) : 0
    const hours = Math.trunc(safeMinutes / 60)
    const remainingMinutes = safeMinutes % 60

    if (hours === 0) return `${remainingMinutes}m`
    if (remainingMinutes === 0) return `${hours}h`
    return `${hours}h ${remainingMinutes}m`
}

const MINUTE_MS = 60_000
const HOUR_MINUTES = 60
const DAY_MINUTES = 24 * HOUR_MINUTES

export function formatTimeRemaining(value: Date | string | number, now = new Date()): string | null {
    const target = value instanceof Date ? value : new Date(value)
    const targetMs = target.getTime()

    if (!Number.isFinite(targetMs)) return null

    const diffMs = targetMs - now.getTime()
    if (diffMs <= 0) return 'less than a minute'

    const totalMinutes = Math.floor(diffMs / MINUTE_MS)
    if (totalMinutes < 1) return 'less than a minute'

    const days = Math.floor(totalMinutes / DAY_MINUTES)
    if (days > 0) {
        const hours = Math.floor((totalMinutes % DAY_MINUTES) / HOUR_MINUTES)
        return hours > 0 ? `${days}d ${hours}h` : `${days}d`
    }

    const hours = Math.floor(totalMinutes / HOUR_MINUTES)
    if (hours > 0) {
        const remainingMinutes = totalMinutes % HOUR_MINUTES
        return remainingMinutes > 0 ? `${hours}h ${remainingMinutes}m` : `${hours}h`
    }

    return `${totalMinutes}m`
}

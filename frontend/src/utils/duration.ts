export function formatInvoiceDurationMinutes(minutes: number): string {
    const safeMinutes = Number.isFinite(minutes) ? Math.max(0, Math.trunc(minutes)) : 0
    const hours = Math.trunc(safeMinutes / 60)
    const remainingMinutes = safeMinutes % 60

    if (hours === 0) return `${remainingMinutes}m`
    if (remainingMinutes === 0) return `${hours}h`
    return `${hours}h ${remainingMinutes}m`
}

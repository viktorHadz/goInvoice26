// Unit helpers
export type MoneyMinor = number

export function toMinor(value: number): MoneyMinor {
    if (!Number.isFinite(value)) return 0
    return Math.round(value * 100)
}

export function fromMinor(minor: MoneyMinor): number {
    return (minor ?? 0) / 100
}

export function fmtGBPMinor(minor: MoneyMinor): string {
    return new Intl.NumberFormat('en-GB', { style: 'currency', currency: 'GBP' }).format(
        (minor ?? 0) / 100,
    )
}

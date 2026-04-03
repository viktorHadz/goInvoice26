function sanitizeMinutes(value: number | null | undefined): number | null {
    if (value == null) return null
    if (!Number.isFinite(value) || value < 0) return null
    return Math.floor(value)
}

export function resolvePickerMinutes(
    overrideMinutes: number | null | undefined,
    productMinutes: number | null | undefined,
    fallbackMinutes = 60,
): number {
    const override = sanitizeMinutes(overrideMinutes)
    if (override != null) return override

    const productDefault = sanitizeMinutes(productMinutes)
    if (productDefault != null) return productDefault

    return sanitizeMinutes(fallbackMinutes) ?? 60
}

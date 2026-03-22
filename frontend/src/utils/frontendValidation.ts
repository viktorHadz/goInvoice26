type TextRules = {
    required?: boolean
    min?: number
    max?: number
    singleLine?: boolean
    trim?: boolean
}

type ProductFormInput = {
    productType: 'style' | 'sample'
    pricingMode: 'flat' | 'hourly'
    productName: string
    flatPrice: number | null
    hourlyRate: number | null
    minutesWorked: number | null
}

type InvoicePayload = {
    overview: {
        issueDate: string
        dueByDate?: string
        clientName: string
        clientCompanyName: string
        clientAddress: string
        clientEmail: string
        note?: string
    }
    lines: Array<{
        productId: number | null
        name: string
        lineType: string
        pricingMode: string
        quantity: number
        minutesWorked: number | null
        unitPriceMinor: number
        lineTotalMinor: number
        sortOrder: number
    }>
    totals: {
        vatRate: number
        vatMinor: number
        depositType: string
        depositRate: number
        depositMinor: number
        discountType: string
        discountRate: number
        discountMinor: number
        paidMinor: number
        subtotalAfterDiscountMinor: number
        subtotalMinor: number
        totalMinor: number
        balanceDueMinor: number
    }
}

const runeLen = (s: string) => Array.from(s).length

function normalizeText(value: unknown): string {
    if (value === null || value === undefined) return ''
    return String(value)
}

function hasControlOrInvalidSeparators(s: string): boolean {
    for (const ch of s) {
        const cp = ch.codePointAt(0) ?? 0
        if (cp < 32 || cp === 127 || cp === 0x2028 || cp === 0x2029) return true
    }
    return false
}

function hasNewlineOrTab(s: string): boolean {
    return /[\n\r\t]/.test(s)
}

function validateText(value: unknown, rules: TextRules): string | null {
    const input = normalizeText(value)
    const out = rules.trim ? input.trim() : input

    if (!out) {
        return rules.required ? 'is required' : null
    }

    const n = runeLen(out)
    if (rules.min && n < rules.min) return 'too short'
    if (rules.max && n > rules.max) return 'too long'

    if (hasControlOrInvalidSeparators(out)) return 'contains invalid characters'
    if (rules.singleLine && hasNewlineOrTab(out)) return 'must be single-line'

    return null
}

function validateEmail(value: unknown, maxRunes: number): string | null {
    const out = normalizeText(value).trim()
    if (!out) return null

    if (maxRunes > 0 && runeLen(out) > maxRunes) return 'too long'
    if (hasControlOrInvalidSeparators(out)) return 'contains invalid characters'

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    if (!emailRegex.test(out)) return 'email format is invalid'

    return null
}

function isValidISODate(value: string): boolean {
    if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) return false
    const parts = value.split('-').map(Number)
    const y = parts[0] ?? Number.NaN
    const m = parts[1] ?? Number.NaN
    const d = parts[2] ?? Number.NaN
    if (!Number.isFinite(y) || !Number.isFinite(m) || !Number.isFinite(d)) return false
    const dt = new Date(Date.UTC(y, m - 1, d))
    return dt.getUTCFullYear() === y && dt.getUTCMonth() === m - 1 && dt.getUTCDate() === d
}

export function validateClientForm(input: {
    name: string | null | undefined
    companyName: string | null | undefined
    address: string | null | undefined
    email: string | null | undefined
}): Record<string, string> {
    const errors: Record<string, string> = {}

    const nameErr = validateText(input.name, {
        required: true,
        min: 2,
        max: 50,
        singleLine: true,
        trim: true,
    })
    if (nameErr) errors.name = nameErr

    const companyErr = validateText(input.companyName, {
        max: 70,
        singleLine: true,
        trim: true,
    })
    if (companyErr) errors.companyName = companyErr

    const addressErr = validateText(input.address, {
        max: 70,
        singleLine: true,
        trim: true,
    })
    if (addressErr) errors.address = addressErr

    const emailErr = validateEmail(input.email, 50)
    if (emailErr) errors.email = emailErr

    return errors
}

export function validateProductForm(input: ProductFormInput): Record<string, string> {
    const errors: Record<string, string> = {}

    const nameErr = validateText(input.productName, {
        required: true,
        min: 2,
        max: 80,
        singleLine: true,
        trim: true,
    })
    if (nameErr) errors.productName = nameErr

    if (input.productType !== 'style' && input.productType !== 'sample') {
        errors.productType = 'invalid value'
    }

    if (input.pricingMode !== 'flat' && input.pricingMode !== 'hourly') {
        errors.pricingMode = 'invalid value'
    }

    if (input.productType === 'style' && input.pricingMode === 'hourly') {
        errors.pricingMode = "must be 'flat' for style"
    }

    if (input.pricingMode === 'flat') {
        if (input.flatPrice == null) {
            errors.flatPrice = 'is required'
        } else if (!Number.isFinite(input.flatPrice) || input.flatPrice < 0) {
            errors.flatPrice = 'value below minimum'
        }
    }

    if (input.pricingMode === 'hourly') {
        if (input.hourlyRate == null) {
            errors.hourlyRate = 'is required'
        } else if (!Number.isFinite(input.hourlyRate) || input.hourlyRate < 0) {
            errors.hourlyRate = 'value below minimum'
        }

        if (input.minutesWorked == null) {
            errors.minutesWorked = 'is required'
        } else if (!Number.isFinite(input.minutesWorked) || input.minutesWorked < 0) {
            errors.minutesWorked = 'value below minimum'
        } else if (!Number.isInteger(input.minutesWorked)) {
            errors.minutesWorked = 'must be an integer'
        }
    }

    return errors
}

function calcExpectedLineTotalMinor(line: InvoicePayload['lines'][number]): number {
    if (line.pricingMode === 'hourly') {
        const minutes = line.minutesWorked ?? 0
        return Math.round((line.quantity * line.unitPriceMinor * minutes) / 60)
    }

    return line.quantity * line.unitPriceMinor
}

export function validateInvoicePayload(payload: InvoicePayload): Record<string, string> {
    const errors: Record<string, string> = {}

    const issueDate = payload.overview.issueDate?.trim() ?? ''
    if (!issueDate) {
        errors.issueDate = 'is required'
    } else if (!isValidISODate(issueDate)) {
        errors.issueDate = 'must be a valid ISO date (YYYY-MM-DD)'
    }

    if (payload.overview.dueByDate?.trim() && !isValidISODate(payload.overview.dueByDate.trim())) {
        errors.dueByDate = 'must be a valid ISO date (YYYY-MM-DD)'
    }

    const note = payload.overview.note ?? ''
    const noteErr = validateText(note, {
        max: 1000,
        singleLine: true,
        trim: true,
    })
    if (noteErr) errors.note = noteErr

    const clientNameErr = validateText(payload.overview.clientName, {
        required: true,
        max: 100,
        singleLine: true,
        trim: true,
    })
    if (clientNameErr) errors.clientName = clientNameErr

    const clientCompanyErr = validateText(payload.overview.clientCompanyName, {
        max: 100,
        singleLine: true,
        trim: true,
    })
    if (clientCompanyErr) errors.clientCompanyName = clientCompanyErr

    const clientAddressErr = validateText(payload.overview.clientAddress, {
        max: 200,
        singleLine: true,
        trim: true,
    })
    if (clientAddressErr) errors.clientAddress = clientAddressErr

    const clientEmailErr = validateEmail(payload.overview.clientEmail, 100)
    if (clientEmailErr) errors.clientEmail = clientEmailErr

    if (payload.lines.length === 0) {
        errors.lines = 'must contain at least one item'
    }

    payload.lines.forEach((line, i) => {
        const prefix = (f: string) => `lines[${i}].${f}`

        if (line.productId != null && line.productId < 1) {
            errors[prefix('productId')] = 'must be greater than 0'
        }

        const lineNameErr = validateText(line.name, {
            required: true,
            min: 1,
            max: 200,
            singleLine: true,
            trim: true,
        })
        if (lineNameErr) errors[prefix('name')] = lineNameErr

        if (!['custom', 'style', 'sample'].includes(line.lineType)) {
            errors[prefix('lineType')] = 'must be one of: custom, style, sample'
        }

        if (!['flat', 'hourly'].includes(line.pricingMode)) {
            errors[prefix('pricingMode')] = 'must be one of: flat, hourly'
        }

        if (line.quantity < 1) {
            errors[prefix('quantity')] = 'must be greater than 0'
        }

        if (line.minutesWorked != null && line.minutesWorked < 0) {
            errors[prefix('minutesWorked')] = 'must be 0 or greater'
        }

        if (line.pricingMode === 'hourly' && line.minutesWorked == null) {
            errors[prefix('minutesWorked')] = 'is required'
        }

        if (line.unitPriceMinor < 0) {
            errors[prefix('unitPriceMinor')] = 'must be 0 or greater'
        }

        if (line.sortOrder < 0) {
            errors[prefix('sortOrder')] = 'must be 0 or greater'
        }

        if (line.lineTotalMinor < 0) {
            errors[prefix('lineTotalMinor')] = 'must be 0 or greater'
        } else {
            const expected = calcExpectedLineTotalMinor(line)

            if (line.lineTotalMinor !== expected) {
                errors[prefix('lineTotalMinor')] =
                    line.pricingMode === 'hourly'
                        ? 'does not match rounded(quantity * unitPriceMinor * minutesWorked / 60)'
                        : 'does not match quantity * unitPriceMinor'
            }
        }
    })

    const totals = payload.totals

    if (totals.vatRate < 0 || totals.vatRate > 10000) {
        errors['totals.vatRate'] = 'must be between 0 and 10000'
    }

    if (totals.vatMinor < 0) errors['totals.vatMinor'] = 'must be 0 or greater'

    if (!['none', 'percent', 'fixed'].includes(totals.depositType)) {
        errors['totals.depositType'] = 'must be one of: none, percent, fixed'
    }

    if (!['none', 'percent', 'fixed'].includes(totals.discountType)) {
        errors['totals.discountType'] = 'must be one of: none, percent, fixed'
    }

    if (totals.depositMinor < 0) errors['totals.depositMinor'] = 'must be 0 or greater'
    if (totals.depositRate < 0 || totals.depositRate > 10000) {
        errors['totals.depositRate'] = 'must be between 0 and 10000'
    }

    if (totals.depositType !== 'percent' && totals.depositRate !== 0) {
        errors['totals.depositRate'] = 'must be 0 unless depositType is percent'
    }
    if (totals.discountType !== 'percent' && totals.discountRate !== 0) {
        errors['totals.discountRate'] = 'must be 0 unless discountType is percent'
    }

    if (totals.discountRate < 0 || totals.discountRate > 10000) {
        errors['totals.discountRate'] = 'must be between 0 and 10000'
    }
    if (totals.discountMinor < 0) errors['totals.discountMinor'] = 'must be 0 or greater'

    if (totals.paidMinor < 0) errors['totals.paidMinor'] = 'must be 0 or greater'

    const maxPaidMinor = Math.max(0, totals.totalMinor - totals.depositMinor)
    if (totals.paidMinor > maxPaidMinor) {
        errors['totals.paidMinor'] = 'cannot exceed amount owing after deposit'
    }

    if (totals.subtotalAfterDiscountMinor < 0) {
        errors['totals.subtotalAfterDiscountMinor'] = 'must be 0 or greater'
    }

    if (totals.subtotalMinor < 0) errors['totals.subtotalMinor'] = 'must be 0 or greater'
    if (totals.totalMinor < 0) errors['totals.totalMinor'] = 'must be 0 or greater'
    if (totals.balanceDueMinor < 0) errors['totals.balanceDueMinor'] = 'must be 0 or greater'

    return errors
}

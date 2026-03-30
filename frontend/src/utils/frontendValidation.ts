import { isValidISODate } from '@/utils/dates'

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
        sourceRevisionNo?: number
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
    payments: Array<{
        amountMinor: number
        paymentDate: string
        label?: string
    }>
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
        return rules.required ? 'Please enter a value.' : null
    }

    const n = runeLen(out)
    if (rules.min && n < rules.min) return `Please enter at least ${rules.min} characters.`
    if (rules.max && n > rules.max) return `Please keep this under ${rules.max} characters.`

    if (hasControlOrInvalidSeparators(out)) return 'Please remove unsupported characters.'
    if (rules.singleLine && hasNewlineOrTab(out)) return 'Use a single line only.'

    return null
}

function validateEmail(value: unknown, maxRunes: number): string | null {
    const out = normalizeText(value).trim()
    if (!out) return null

    if (maxRunes > 0 && runeLen(out) > maxRunes)
        return `Please keep this under ${maxRunes} characters.`
    if (hasControlOrInvalidSeparators(out)) return 'Please remove unsupported characters.'

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    if (!emailRegex.test(out)) return 'Enter a valid email address.'

    return null
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
        errors.productType = 'Choose a valid product type.'
    }

    if (input.pricingMode !== 'flat' && input.pricingMode !== 'hourly') {
        errors.pricingMode = 'Choose a valid pricing mode.'
    }

    if (input.productType === 'style' && input.pricingMode === 'hourly') {
        errors.pricingMode = 'Styles must use flat pricing.'
    }

    if (input.pricingMode === 'flat') {
        if (input.flatPrice == null) {
            errors.flatPrice = 'Enter a flat price.'
        } else if (!Number.isFinite(input.flatPrice) || input.flatPrice < 0) {
            errors.flatPrice = 'Enter a value of 0 or higher.'
        }
    }

    if (input.pricingMode === 'hourly') {
        if (input.hourlyRate == null) {
            errors.hourlyRate = 'Enter an hourly rate.'
        } else if (!Number.isFinite(input.hourlyRate) || input.hourlyRate < 0) {
            errors.hourlyRate = 'Enter a value of 0 or higher.'
        }

        if (input.minutesWorked == null) {
            errors.minutesWorked = 'Enter minutes worked.'
        } else if (!Number.isFinite(input.minutesWorked) || input.minutesWorked < 0) {
            errors.minutesWorked = 'Enter a value of 0 or higher.'
        } else if (!Number.isInteger(input.minutesWorked)) {
            errors.minutesWorked = 'Use a whole number.'
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
    if (
        payload.overview.sourceRevisionNo != null &&
        (!Number.isInteger(payload.overview.sourceRevisionNo) ||
            payload.overview.sourceRevisionNo < 1)
    ) {
        errors.sourceRevisionNo = 'must be a positive integer'
    }
    if (!issueDate) {
        errors.issueDate = 'Choose an issue date.'
    } else if (!isValidISODate(issueDate)) {
        errors.issueDate = 'Use a valid date in YYYY-MM-DD format.'
    }

    if (payload.overview.dueByDate?.trim() && !isValidISODate(payload.overview.dueByDate.trim())) {
        errors.dueByDate = 'Use a valid date in YYYY-MM-DD format.'
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
        errors.lines = 'Add at least one line item.'
    }

    payload.lines.forEach((line, i) => {
        const prefix = (f: string) => `lines[${i}].${f}`

        if (line.productId != null && line.productId < 1) {
            errors[prefix('productId')] = 'Choose a valid product.'
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
            errors[prefix('lineType')] = 'Choose a valid line type.'
        }

        if (!['flat', 'hourly'].includes(line.pricingMode)) {
            errors[prefix('pricingMode')] = 'Choose a valid pricing mode.'
        }

        if (line.quantity < 1) {
            errors[prefix('quantity')] = 'Enter a quantity greater than 0.'
        }

        if (line.minutesWorked != null && line.minutesWorked < 0) {
            errors[prefix('minutesWorked')] = 'Enter minutes of 0 or more.'
        }

        if (line.pricingMode === 'hourly' && line.minutesWorked == null) {
            errors[prefix('minutesWorked')] = 'Enter minutes worked.'
        }

        if (line.unitPriceMinor < 0) {
            errors[prefix('unitPriceMinor')] = 'Enter a unit price of 0 or more.'
        }

        if (line.sortOrder < 0) {
            errors[prefix('sortOrder')] = 'Sort order must be 0 or more.'
        }

        if (line.lineTotalMinor < 0) {
            errors[prefix('lineTotalMinor')] = 'Line total must be 0 or more.'
        } else {
            const expected = calcExpectedLineTotalMinor(line)

            if (line.lineTotalMinor !== expected) {
                errors[prefix('lineTotalMinor')] =
                    line.pricingMode === 'hourly'
                        ? 'Line total does not match the quantity, rate, and time.'
                        : 'Line total does not match quantity and unit price.'
            }
        }
    })

    const totals = payload.totals
    const paymentSumMinor = payload.payments.reduce((sum, p) => sum + Math.max(0, p.amountMinor), 0)

    if (totals.vatRate < 0 || totals.vatRate > 10000) {
        errors['totals.vatRate'] = 'Enter a VAT rate between 0% and 100%.'
    }

    if (totals.vatMinor < 0) errors['totals.vatMinor'] = 'VAT amount must be 0 or more.'

    if (!['none', 'percent', 'fixed'].includes(totals.depositType)) {
        errors['totals.depositType'] = 'Choose a valid deposit type.'
    }

    if (!['none', 'percent', 'fixed'].includes(totals.discountType)) {
        errors['totals.discountType'] = 'Choose a valid discount type.'
    }

    if (totals.depositMinor < 0) errors['totals.depositMinor'] = 'Deposit must be 0 or more.'
    if (totals.depositRate < 0 || totals.depositRate > 10000) {
        errors['totals.depositRate'] = 'Deposit rate must be between 0% and 100%.'
    }

    if (totals.depositType !== 'percent' && totals.depositRate !== 0) {
        errors['totals.depositRate'] = 'Deposit rate must be 0 unless deposit type is percent.'
    }
    if (totals.discountType !== 'percent' && totals.discountRate !== 0) {
        errors['totals.discountRate'] = 'Discount rate must be 0 unless discount type is percent.'
    }

    if (totals.discountRate < 0 || totals.discountRate > 10000) {
        errors['totals.discountRate'] = 'Discount rate must be between 0% and 100%.'
    }
    if (totals.discountMinor < 0) errors['totals.discountMinor'] = 'Discount must be 0 or more.'

    if (totals.paidMinor < 0) errors['totals.paidMinor'] = 'Paid amount must be 0 or more.'

    const maxPaidMinor = Math.max(0, totals.totalMinor - totals.depositMinor)
    if (totals.paidMinor > maxPaidMinor) {
        errors['totals.paidMinor'] = 'Paid amount cannot exceed the balance after deposit.'
    }

    if (totals.subtotalAfterDiscountMinor < 0) {
        errors['totals.subtotalAfterDiscountMinor'] = 'Subtotal after discount must be 0 or more.'
    }

    if (totals.subtotalMinor < 0) errors['totals.subtotalMinor'] = 'Subtotal must be 0 or more.'
    if (totals.totalMinor < 0) errors['totals.totalMinor'] = 'Total must be 0 or more.'
    if (totals.balanceDueMinor < 0)
        errors['totals.balanceDueMinor'] = 'Balance due must be 0 or more.'

    payload.payments.forEach((payment, i) => {
        const prefix = (f: string) => `payments[${i}].${f}`
        if (!Number.isFinite(payment.amountMinor) || payment.amountMinor <= 0) {
            errors[prefix('amountMinor')] = 'Payment amount must be greater than 0.'
        }
        const paymentDate = payment.paymentDate?.trim() ?? ''
        if (!paymentDate) {
            errors[prefix('paymentDate')] = 'Choose a payment date.'
        } else if (!isValidISODate(paymentDate)) {
            errors[prefix('paymentDate')] = 'Use a valid date in YYYY-MM-DD format.'
        }
        if (payment.label != null) {
            const labelErr = validateText(payment.label, {
                max: 120,
                singleLine: true,
                trim: true,
            })
            if (labelErr) errors[prefix('label')] = labelErr
        }
    })

    if (paymentSumMinor > totals.paidMinor) {
        errors['totals.paidMinor'] = 'Paid amount must include all staged payments.'
    }

    return errors
}

export type APIFieldError = {
    field: string
    code: string
    message?: string
    meta?: Record<string, unknown>
}

export type APIErrorPayload = {
    id?: string
    code: string
    message: string
    fields?: APIFieldError[]
}

type APIErrorEnvelope = {
    error?: APIErrorPayload
}

const ERROR_CODE_MESSAGES: Record<string, string> = {
    VALIDATION_FAILED: 'Please review the highlighted fields.',
    BAD_JSON: 'We could not read that request. Please try again.',
    NOT_FOUND: 'The requested resource was not found.',
    DATABASE_ERROR: 'Something went wrong while saving data. Please try again.',
    INTERNAL: 'Something went wrong on our side. Please try again.',
    INVALID_ID: 'The selected item is invalid.',
    INVOICE_DRAFT: 'Issue the draft before saving a revision.',
    INVOICE_ISSUED: 'Issued invoices are locked from deletion and must use revisions for edits.',
    INVOICE_NUMBER_LOCKED:
        'Starting invoice number can only be changed when there are no invoices.',
    DRAFT_HAS_REVISIONS: 'This draft can no longer be updated in place.',
}

export class ApiError extends Error {
    id?: string
    code: string
    status: number
    fields: APIFieldError[]

    constructor(payload: APIErrorPayload, status: number) {
        super(payload.message || translateErrorCode(payload.code))
        this.name = 'ApiError'
        this.id = payload.id
        this.code = payload.code
        this.status = status
        this.fields = payload.fields ?? []
    }
}

export function isApiError(value: unknown): value is ApiError {
    return value instanceof ApiError
}

export function translateErrorCode(code?: string, fallback?: string): string {
    if (!code) return fallback ?? 'Request failed'
    return ERROR_CODE_MESSAGES[code] ?? fallback ?? code
}

export function toFieldErrorMap(fields: APIFieldError[]): Record<string, string> {
    const mapped: Record<string, string> = {}

    for (const fieldError of fields) {
        if (!fieldError.field || mapped[fieldError.field]) continue

        mapped[fieldError.field] =
            typeof fieldError.message === 'string' && fieldError.message.trim().length > 0
                ? fieldError.message
                : 'Invalid value'
    }

    return mapped
}

export function hasFieldErrors(err: ApiError): boolean {
    return err.fields.some(
        (fieldError) => typeof fieldError.field === 'string' && fieldError.field.trim().length > 0,
    )
}

export function getApiErrorMessage(err: ApiError): string {
    const message = err.message?.trim()
    if (message) return message
    return translateErrorCode(err.code, 'Request failed')
}

export function isSupportOnlyApiError(err: ApiError): boolean {
    return err.status >= 500 || err.code === 'INTERNAL' || err.code === 'DATABASE_ERROR'
}

function parseFieldErrors(value: unknown): APIFieldError[] {
    if (!Array.isArray(value)) return []

    return value
        .filter((item): item is APIFieldError => !!item && typeof item === 'object')
        .map((item) => ({
            field: typeof item.field === 'string' ? item.field : '',
            code: typeof item.code === 'string' ? item.code : 'INVALID',
            message: typeof item.message === 'string' ? item.message : undefined,
            meta: item.meta && typeof item.meta === 'object' ? item.meta : undefined,
        }))
}

function parseErrorPayload(input: unknown): APIErrorPayload | null {
    if (!input || typeof input !== 'object') return null

    const data = input as Partial<APIErrorEnvelope & APIErrorPayload>
    const payload = data.error ?? data

    if (!payload || typeof payload !== 'object') return null
    if (typeof payload.code !== 'string') return null

    return {
        id: typeof payload.id === 'string' ? payload.id : undefined,
        code: payload.code,
        message:
            typeof payload.message === 'string' && payload.message.trim().length > 0
                ? payload.message
                : translateErrorCode(payload.code),
        fields: parseFieldErrors(payload.fields),
    }
}

export function parseApiError(status: number, bodyText: string): ApiError {
    if (bodyText) {
        try {
            const data = JSON.parse(bodyText)
            const payload = parseErrorPayload(data)
            if (payload) return new ApiError(payload, status)
        } catch {
            // ignore invalid JSON
        }
    }

    const code = `HTTP_${status}`
    const message = bodyText?.trim() || `Request failed (${status}).`

    return new ApiError({ code, message, fields: [] }, status)
}

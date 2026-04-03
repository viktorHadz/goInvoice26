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
    UNAUTHENTICATED: 'Please sign in to continue.',
    FORBIDDEN: 'You do not have access to that area.',
    INVALID_ID: 'The selected item is invalid.',
    INVOICE_DRAFT: 'Issue the draft before saving a revision.',
    INVOICE_ISSUED: 'Issued invoices are locked from deletion and must use revisions for edits.',
    INVOICE_NUMBER_LOCKED:
        'Starting invoice number can only be changed when there are no invoices.',
    DRAFT_HAS_REVISIONS: 'This draft can no longer be updated in place.',
    TEAM_INVITE_EXISTS: 'That email already has a pending invite.',
    TEAM_MEMBER_EXISTS: 'That teammate already has access.',
    TEAM_CANNOT_REMOVE_SELF: 'Use sign out instead of removing your own account.',
    TEAM_CANNOT_REMOVE_OWNER: 'The owner account cannot be removed here.',
    TEAM_PLAN_REQUIRED: 'Upgrade to the team plan before inviting teammates.',
    TEAM_SEAT_LIMIT_REACHED: 'That team plan is full. Remove someone or upgrade the plan later.',
    SUBSCRIPTION_REQUIRED: 'Active billing or a valid access grant is required to use the workspace.',
    BILLING_OWNER_ONLY: 'Only the workspace admin can manage billing.',
    BILLING_SUBSCRIPTION_NOT_FOUND: 'There is no active subscription to cancel.',
    BILLING_NOT_CONFIGURED: 'Billing is temporarily unavailable. Please get in touch.',
    BILLING_CUSTOMER_NOT_FOUND: 'There is no Stripe customer for this account yet.',
    BILLING_CHECKOUT_PENDING: 'Payment is still being confirmed. Please wait a moment.',
    BILLING_CHECKOUT_INVALID: 'That Stripe checkout session is not linked to this account.',
    BILLING_PLAN_INVALID: 'That billing plan is not supported.',
    BILLING_INTERVAL_INVALID: 'That billing interval is not supported.',
    BILLING_PLAN_UNAVAILABLE: 'That billing selection is not available yet.',
    BILLING_PLAN_ALREADY_ACTIVE: 'That billing selection is already active.',
    BILLING_PLAN_DOWNGRADE_BLOCKED:
        'Remove extra teammates and pending invites before switching to the single-user plan.',
    BILLING_PROVIDER_ERROR: 'Stripe could not complete that request. Please try again.',
    PROMO_CODE_NOT_FOUND: 'That promo code was not found.',
    PROMO_CODE_INACTIVE: 'That promo code is no longer active.',
    PROMO_CODE_ALREADY_REDEEMED: 'That promo code has already been used for this workspace.',
    PROMO_ACCESS_ALREADY_ACTIVE: 'This workspace already has active access.',
    PROMO_CODE_EXISTS: 'That promo code already exists.',
    DIRECT_ACCESS_GRANT_EXISTS: 'That email already has a direct access grant.',
    PLATFORM_ADMIN_ONLY: 'Only the platform admin can manage app access.',
    WORKSPACE_DELETE_BILLING_BLOCKED:
        'Cancel the Stripe subscription before deleting this workspace.',
    SETTINGS_OWNER_ONLY: 'Only the workspace admin can edit settings.',
    CLIENT_HAS_INVOICES:
        "This client can't be deleted because there are saved invoices linked to them.",
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

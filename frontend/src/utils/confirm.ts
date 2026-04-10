export type ConfirmationVariant = 'primary' | 'danger' | 'success'
export type ConfirmationDecision = 'confirm' | 'cancel' | 'alternate'

export type ConfirmationOptions = {
    title: string
    message: string
    details?: string
    confirmLabel?: string
    cancelLabel?: string
    alternateLabel?: string
    confirmVariant?: ConfirmationVariant
    alternateVariant?: ConfirmationVariant
}

export type ConfirmationRequest = {
    id: string
    title: string
    message: string
    details?: string
    confirmLabel: string
    cancelLabel: string
    alternateLabel?: string
    confirmVariant: ConfirmationVariant
    alternateVariant?: ConfirmationVariant
    respond: (decision: ConfirmationDecision) => void
}

const CONFIRM_EVENT = 'app:confirm'

function createConfirmationId() {
    if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
        return crypto.randomUUID()
    }

    return `confirm-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
}

export function requestConfirmationChoice(
    options: ConfirmationOptions,
): Promise<ConfirmationDecision> {
    if (typeof window === 'undefined') return Promise.resolve('cancel')

    return new Promise((resolve) => {
        const payload: ConfirmationRequest = {
            id: createConfirmationId(),
            title: options.title,
            message: options.message,
            details: options.details,
            confirmLabel: options.confirmLabel ?? 'Confirm',
            cancelLabel: options.cancelLabel ?? 'Cancel',
            alternateLabel: options.alternateLabel,
            confirmVariant: options.confirmVariant ?? 'primary',
            alternateVariant: options.alternateVariant ?? 'success',
            respond: resolve,
        }

        window.dispatchEvent(
            new CustomEvent<ConfirmationRequest>(CONFIRM_EVENT, { detail: payload }),
        )
    })
}

export function requestConfirmation(options: ConfirmationOptions): Promise<boolean> {
    return requestConfirmationChoice(options).then((decision) => decision === 'confirm')
}

export function onConfirmationRequest(listener: (request: ConfirmationRequest) => void) {
    if (typeof window === 'undefined') {
        return () => {}
    }

    const handler = (event: Event) => {
        const customEvent = event as CustomEvent<ConfirmationRequest>
        if (customEvent.detail) {
            listener(customEvent.detail)
        }
    }

    window.addEventListener(CONFIRM_EVENT, handler)

    return () => {
        window.removeEventListener(CONFIRM_EVENT, handler)
    }
}

export type ConfirmationVariant = 'primary' | 'danger' | 'success'

export type ConfirmationOptions = {
    title: string
    message: string
    details?: string
    confirmLabel?: string
    cancelLabel?: string
    confirmVariant?: ConfirmationVariant
}

export type ConfirmationRequest = {
    id: string
    title: string
    message: string
    details?: string
    confirmLabel: string
    cancelLabel: string
    confirmVariant: ConfirmationVariant
    respond: (confirmed: boolean) => void
}

const CONFIRM_EVENT = 'app:confirm'

function createConfirmationId() {
    if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
        return crypto.randomUUID()
    }

    return `confirm-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
}

export function requestConfirmation(options: ConfirmationOptions): Promise<boolean> {
    if (typeof window === 'undefined') return Promise.resolve(false)

    return new Promise((resolve) => {
        const payload: ConfirmationRequest = {
            id: createConfirmationId(),
            title: options.title,
            message: options.message,
            details: options.details,
            confirmLabel: options.confirmLabel ?? 'Confirm',
            cancelLabel: options.cancelLabel ?? 'Cancel',
            confirmVariant: options.confirmVariant ?? 'primary',
            respond: resolve,
        }

        window.dispatchEvent(
            new CustomEvent<ConfirmationRequest>(CONFIRM_EVENT, { detail: payload }),
        )
    })
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

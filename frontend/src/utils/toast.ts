export type ToastLevel = 'error' | 'success' | 'info'

export type ToastPayload = {
    id: string
    level: ToastLevel
    code?: string
    title?: string
    message: string
    durationMs?: number
}

const TOAST_EVENT = 'app:toast'

function createToastId() {
    if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
        return crypto.randomUUID()
    }

    return `toast-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
}

function emitToast(toast: Omit<ToastPayload, 'id'> & { id?: string }) {
    if (typeof window === 'undefined') return

    const payload: ToastPayload = {
        id: toast.id ?? createToastId(),
        level: toast.level,
        code: toast.code,
        title: toast.title,
        message: toast.message,
        durationMs: toast.durationMs,
    }

    window.dispatchEvent(new CustomEvent<ToastPayload>(TOAST_EVENT, { detail: payload }))
}

export function emitToastError(error: {
    message: string
    code?: string
    id?: string
    title?: string
}) {
    emitToast({
        id: error.id,
        level: 'error',
        code: error.code,
        title: error.title,
        message: error.message,
    })
}

export function emitToastSuccess(
    message: string,
    options?: { id?: string; code?: string; title?: string; durationMs?: number },
) {
    emitToast({
        id: options?.id,
        level: 'success',
        code: options?.code,
        title: options?.title,
        message,
        durationMs: options?.durationMs,
    })
}

export function emitToastInfo(
    message: string,
    options?: { id?: string; code?: string; title?: string; durationMs?: number },
) {
    emitToast({
        id: options?.id,
        level: 'info',
        code: options?.code,
        title: options?.title,
        message,
        durationMs: options?.durationMs,
    })
}

export function onToast(listener: (toast: ToastPayload) => void) {
    if (typeof window === 'undefined') {
        return () => {}
    }

    const handler = (event: Event) => {
        const customEvent = event as CustomEvent<ToastPayload>
        if (customEvent.detail) {
            listener(customEvent.detail)
        }
    }

    window.addEventListener(TOAST_EVENT, handler)

    return () => {
        window.removeEventListener(TOAST_EVENT, handler)
    }
}

export const onToastError = onToast

import {
    getApiErrorMessage,
    hasFieldErrors,
    isApiError,
    isSupportOnlyApiError,
    toFieldErrorMap,
} from '@/utils/apiErrors'
import { emitToastError } from '@/utils/toast'

export type HandleActionErrorOptions = {
    fieldErrors?: { value: Record<string, string> }
    toastTitle: string
    mapFields?: boolean
    supportMessage?: string
}

/**
 * UI-oriented helper for actions (form submissions, CRUD).
 *
 * Policy:
 * - If the server provided field errors (validation), map them to `fieldErrors` and do not toast.
 * - Otherwise, emit a toast with a user-friendly message. (Callers opt-in by calling this helper.)
 */
export function handleActionError(err: unknown, options: HandleActionErrorOptions) {
    const {
        fieldErrors,
        toastTitle,
        mapFields = true,
        supportMessage = 'Something unexpected went wrong. Please contact support.',
    } = options

    if (fieldErrors) fieldErrors.value = {}

    if (isApiError(err)) {
        if (mapFields && fieldErrors && hasFieldErrors(err)) {
            fieldErrors.value = toFieldErrorMap(err.fields)
            return
        }

        emitToastError({
            id: err.id,
            code: err.code,
            title: toastTitle,
            message:
                isSupportOnlyApiError(err) && err.id
                    ? `Something went wrong. Please contact support and quote error ${err.id}.`
                    : isSupportOnlyApiError(err)
                      ? supportMessage
                      : getApiErrorMessage(err),
        })

        if (isSupportOnlyApiError(err)) console.error(err)
        return
    }

    emitToastError({
        title: toastTitle,
        message: err instanceof Error && err.message.trim() ? err.message : supportMessage,
    })

    console.error(err)
}


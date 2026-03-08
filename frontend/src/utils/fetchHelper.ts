import { parseApiError, translateErrorCode } from './apiErrors'
import { emitToastError } from './toast'

export type RequestOptions = {
    toastOnError?: boolean
}

// API request helper
/**
 * Makes a generic HTTP request and parses the response.
 * @template T - The expected response data type
 * @param input - The URL or RequestInfo object for the request
 * @param init - Optional RequestInit object to configure the request (method, headers, body, etc.)
 * @param options - Request behavior options such as opt-in toasting on error
 * @returns A promise that resolves to the parsed response data of type T
 * @throws Error if the response status is not ok, or if the response body cannot be parsed as JSON
 * @remarks
 * - Handles 204 No Content responses by returning undefined
 * - Attempts to extract error messages from the API response
 * - Returns undefined if the response body is empty but status is ok
 */
export async function request<T>(
    input: RequestInfo,
    init?: RequestInit,
    options?: RequestOptions,
): Promise<T> {
    const res = await fetch(input, init)

    if (!res.ok) {
        const text = await res.text().catch(() => '')
        const apiError = parseApiError(res.status, text)
        const toastMessage = translateErrorCode(apiError.code, apiError.message)
        const hasFieldValidation =
            apiError.code === 'VALIDATION_FAILED' &&
            Array.isArray(apiError.fields) &&
            apiError.fields.length > 0

        if (options?.toastOnError && !hasFieldValidation) {
            emitToastError({ code: apiError.code, message: toastMessage, id: apiError.id })
        }
        throw apiError
    }

    // Handle 204 No Content
    if (res.status === 204) {
        return undefined as unknown as T
    }

    // avoids JSON parse crash if server returns no body but not 204
    const text = await res.text()
    if (!text) {
        return undefined as unknown as T
    }

    return JSON.parse(text) as T
}

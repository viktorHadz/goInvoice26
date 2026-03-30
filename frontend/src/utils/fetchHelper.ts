import { parseApiError } from './apiErrors'

export class NetworkError extends Error {
    cause?: unknown

    constructor(message = 'Network request failed', cause?: unknown) {
        super(message)
        this.name = 'NetworkError'
        this.cause = cause
    }
}

export class ParseError extends Error {
    status: number
    bodyText?: string
    cause?: unknown

    constructor(message: string, status: number, bodyText?: string, cause?: unknown) {
        super(message)
        this.name = 'ParseError'
        this.status = status
        this.bodyText = bodyText
        this.cause = cause
    }
}

export async function request<T>(input: RequestInfo, init?: RequestInit): Promise<T> {
    const normalizedInput =
        typeof input === 'string' && input.startsWith('api/') ? `/${input}` : input

    let res: Response
    try {
        res = await fetch(normalizedInput, init)
    } catch (err) {
        const name = (err as any)?.name
        if (name === 'AbortError') {
            throw new NetworkError('Request aborted', err)
        }
        throw new NetworkError(undefined, err)
    }

    if (!res.ok) {
        const text = await res.text().catch(() => '')
        throw parseApiError(res.status, text)
    }

    if (res.status === 204) {
        return undefined as unknown as T
    }

    const text = await res.text()
    if (!text) {
        return undefined as unknown as T
    }

    try {
        return JSON.parse(text) as T
    } catch (err) {
        throw new ParseError(
            'Received an unreadable response from the server. Please contact support.',
            res.status,
            text,
            err,
        )
    }
}

import { request } from './fetchHelper'

export type BillingLink = {
    url: string
}

export function createCheckoutSession() {
    return request<BillingLink>('/api/billing/checkout-session', {
        method: 'POST',
    })
}

export function createPortalSession() {
    return request<BillingLink>('/api/billing/portal-session', {
        method: 'POST',
    })
}

export function syncCheckoutSession(sessionId: string) {
    return request<void>('/api/billing/checkout/sync', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ sessionId }),
    })
}

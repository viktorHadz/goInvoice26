import { request } from './fetchHelper'
import type { BillingInterval, BillingPlan } from '@/constants/billing'

export type BillingLink = {
    url: string
}

export type PromoCodeRedemption = {
    code: string
    durationDays: number
    expiresAt: string
}

export function createCheckoutSession(
    plan: BillingPlan,
    interval: BillingInterval,
    redirect?: string,
) {
    return request<BillingLink>('/api/billing/checkout-session', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ plan, interval, redirect }),
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

export function cancelSubscription() {
    return request<void>('/api/billing/subscription/cancel', {
        method: 'POST',
    })
}

export function changeSubscriptionPlan(plan: BillingPlan, interval: BillingInterval) {
    return request<void>('/api/billing/subscription/plan', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ plan, interval }),
    })
}

export function redeemPromoCode(code: string) {
    return request<PromoCodeRedemption>('/api/billing/promo-codes/redeem', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ code }),
    })
}

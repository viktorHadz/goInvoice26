import { request } from './fetchHelper'

export type DirectAccessGrant = {
    id: number
    email: string
    plan: 'single' | 'team'
    note?: string
    createdAt: string
}

export type PromoCode = {
    id: number
    code: string
    durationDays: number
    active: boolean
    redemptionCount: number
    createdAt: string
}

export type PlatformAccessOverview = {
    directGrants: DirectAccessGrant[]
    promoCodes: PromoCode[]
}

export function fetchPlatformAccessOverview() {
    return request<PlatformAccessOverview>('/api/admin/access')
}

export function createDirectAccessGrant(email: string, plan: 'single' | 'team', note = '') {
    return request<DirectAccessGrant>('/api/admin/access/grants', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, plan, note }),
    })
}

export function deleteDirectAccessGrant(grantId: number) {
    return request<void>(`/api/admin/access/grants/${grantId}`, {
        method: 'DELETE',
    })
}

export function createPromoCode(code: string, durationDays: number) {
    return request<PromoCode>('/api/admin/access/promo-codes', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ code, durationDays }),
    })
}

export function updatePromoCodeStatus(promoCodeId: number, active: boolean) {
    return request<void>(`/api/admin/access/promo-codes/${promoCodeId}`, {
        method: 'PATCH',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ active }),
    })
}

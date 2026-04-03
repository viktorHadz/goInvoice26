import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { request } from '@/utils/fetchHelper'
import { useClientStore } from '@/stores/clients'
import { useProductStore } from '@/stores/products'
import { useSettingsStore } from '@/stores/settings'
import { useTeamStore } from '@/stores/team'
import { useInvoiceStore } from '@/stores/invoice'
import { useEditorStore } from '@/stores/editor'
import type { BillingInterval, BillingPlan } from '@/constants/billing'

export type AuthMode = 'login' | 'signup'

export type AuthUser = {
    id: number
    name: string
    email: string
    avatarUrl: string
    role: 'owner' | 'member'
}

export type AuthAccount = {
    id: number
    name: string
}

export type AuthBilling = {
    configured: boolean
    status: string
    accessGranted: boolean
    accessSource?: 'subscription' | 'direct' | 'promo'
    accessExpiresAt?: string
    promoCode?: string
    promoExpired: boolean
    canManage: boolean
    portalAvailable: boolean
    trialDays: number
    plan: BillingPlan | ''
    interval: BillingInterval | ''
    seatLimit: number
    singlePlanAvailable: boolean
    teamPlanAvailable: boolean
    singleMonthlyAvailable: boolean
    singleYearlyAvailable: boolean
    teamMonthlyAvailable: boolean
    teamYearlyAvailable: boolean
    currentPeriodEnd?: string
    cancelAtPeriodEnd: boolean
}

export type AuthStatus = {
    authenticated: boolean
    needsSetup: boolean
    canRegister?: boolean
    googleEnabled: boolean
    canManagePlatformAccess: boolean
    user?: AuthUser
    account?: AuthAccount
    billing?: AuthBilling
}

const DEFAULT_REDIRECT = '/app'

export const useAuthStore = defineStore('auth', () => {
    const status = ref<AuthStatus | null>(null)
    const isLoading = ref(false)
    const hasLoaded = ref(false)

    const isAuthenticated = computed(() => status.value?.authenticated === true)
    const isOwner = computed(() => status.value?.user?.role === 'owner')
    const canManagePlatformAccess = computed(() => status.value?.canManagePlatformAccess === true)
    const needsSetup = computed(() => status.value?.needsSetup === true)
    const canRegister = computed(
        () => status.value?.canRegister === true || status.value?.needsSetup === true,
    )
    const googleEnabled = computed(() => status.value?.googleEnabled === true)
    const user = computed(() => status.value?.user ?? null)
    const account = computed(() => status.value?.account ?? null)
    const billing = computed(() => status.value?.billing ?? null)
    const hasBillingAccess = computed(() => status.value?.billing?.accessGranted === true)
    const canManageBilling = computed(() => status.value?.billing?.canManage === true)
    const billingConfigured = computed(() => status.value?.billing?.configured === true)
    const billingPortalAvailable = computed(() => status.value?.billing?.portalAvailable === true)

    async function fetchSession(force = false) {
        if (hasLoaded.value && !force && status.value) {
            return status.value
        }

        isLoading.value = true
        try {
            const data = await request<AuthStatus>('/api/auth/me')
            status.value = data
            hasLoaded.value = true

            if (!data.authenticated || data.billing?.accessGranted !== true) {
                clearWorkspaceState()
            }

            return data
        } finally {
            isLoading.value = false
        }
    }

    function beginGoogleAuth(mode: AuthMode, redirectPath = DEFAULT_REDIRECT) {
        const params = new URLSearchParams({
            mode,
            redirect: sanitizeRedirectPath(redirectPath),
        })

        window.location.assign(`/api/auth/google/start?${params.toString()}`)
    }

    async function logout() {
        await request<void>('/api/auth/logout', { method: 'POST' })
        clearWorkspaceState()
        await fetchSession(true)
    }

    function clearWorkspaceState() {
        useClientStore().reset()
        useSettingsStore().reset()
        useProductStore().reset()
        useTeamStore().reset()
        useInvoiceStore().reset()
        useEditorStore().reset()
    }

    return {
        status,
        isLoading,
        hasLoaded,
        isAuthenticated,
        isOwner,
        canManagePlatformAccess,
        needsSetup,
        canRegister,
        googleEnabled,
        user,
        account,
        billing,
        hasBillingAccess,
        canManageBilling,
        billingConfigured,
        billingPortalAvailable,
        fetchSession,
        beginGoogleAuth,
        logout,
        clearWorkspaceState,
    }
})

function sanitizeRedirectPath(path: string) {
    const normalized = path?.trim() || DEFAULT_REDIRECT
    if (!normalized.startsWith('/') || normalized.startsWith('//')) {
        return DEFAULT_REDIRECT
    }

    return normalized
}

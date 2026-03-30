import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { request } from '@/utils/fetchHelper'
import { useClientStore } from '@/stores/clients'
import { useProductStore } from '@/stores/products'
import { useSettingsStore } from '@/stores/settings'
import { useTeamStore } from '@/stores/team'

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

export type AuthStatus = {
    authenticated: boolean
    needsSetup: boolean
    googleEnabled: boolean
    user?: AuthUser
    account?: AuthAccount
}

const DEFAULT_REDIRECT = '/app'

export const useAuthStore = defineStore('auth', () => {
    const status = ref<AuthStatus | null>(null)
    const isLoading = ref(false)
    const hasLoaded = ref(false)

    const isAuthenticated = computed(() => status.value?.authenticated === true)
    const isOwner = computed(() => status.value?.user?.role === 'owner')
    const needsSetup = computed(() => status.value?.needsSetup === true)
    const googleEnabled = computed(() => status.value?.googleEnabled === true)
    const user = computed(() => status.value?.user ?? null)
    const account = computed(() => status.value?.account ?? null)

    async function fetchSession(force = false) {
        if (hasLoaded.value && !force && status.value) {
            return status.value
        }

        isLoading.value = true
        try {
            const data = await request<AuthStatus>('/api/auth/me')
            status.value = data
            hasLoaded.value = true
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
    }

    return {
        status,
        isLoading,
        hasLoaded,
        isAuthenticated,
        isOwner,
        needsSetup,
        googleEnabled,
        user,
        account,
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

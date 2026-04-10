import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useClientStore } from '@/stores/clients'
import { useEditorStore } from '@/stores/editor'
import { useSettingsStore } from '@/stores/settings'
import { emitToastError } from '@/utils/toast'
import { getApiErrorMessage, isApiError } from '@/utils/apiErrors'

type GuardToastMeta = {
    title: string
    message: string
}

function emitGuardToast(meta: unknown) {
    if (!meta || typeof meta !== 'object') return
    const toast = meta as Partial<GuardToastMeta>
    if (!toast.title || !toast.message) return

    emitToastError({
        title: toast.title,
        message: toast.message,
    })
}

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        {
            path: '/',
            name: 'landing',
            component: () => import('@/views/LandingView.vue'),
        },
        {
            path: '/privacy',
            name: 'privacy',
            component: () => import('@/views/PrivacyView.vue'),
        },
        {
            path: '/signup',
            name: 'signup',
            component: () => import('@/views/AuthEntryView.vue'),
            props: { mode: 'signup' },
        },
        {
            path: '/login',
            name: 'login',
            component: () => import('@/views/AuthEntryView.vue'),
            props: { mode: 'login' },
        },
        {
            path: '/app',
            name: 'app-home',
            component: () => import('@/views/HomeView.vue'),
            meta: { appChrome: true, requiresBilling: true },
        },
        {
            path: '/app/billing',
            name: 'billing',
            component: () => import('@/views/BillingView.vue'),
            meta: { appChrome: true, allowUnpaid: true },
        },
        {
            path: '/app/clients',
            name: 'clients',
            component: () => import('@/views/ClientsView.vue'),
            meta: { appChrome: true, requiresBilling: true, requiresClients: true },
        },
        {
            path: '/app/invoice',
            name: 'invoice',
            component: () => import('@/views/InvoiceView.vue'),
            meta: {
                appChrome: true,
                requiresBilling: true,
                requiresSelectedClient: true,
                guardToast: {
                    title: 'Select a client first',
                    message: 'Pick a client from Quick Menu before opening Invoice.',
                },
            },
        },
        {
            path: '/app/editor',
            name: 'editor',
            component: () => import('@/views/EditorView.vue'),
            meta: {
                appChrome: true,
                requiresBilling: true,
                requiresSelectedClient: true,
                guardToast: {
                    title: 'Select a client first',
                    message: 'Pick a client from Quick Menu before opening Editor.',
                },
            },
        },
        {
            path: '/app/team',
            redirect: { name: 'app-home' },
        },
        {
            path: '/:pathMatch(.*)*',
            name: 'not-found',
            component: () => import('@/views/NotFoundView.vue'),
        },
    ],
    linkActiveClass: 'router-active',
})

router.beforeEach(async (to) => {
    const authStore = useAuthStore()
    const clientStore = useClientStore()
    const editorStore = useEditorStore()
    const settingsStore = useSettingsStore()

    if (to.name !== 'editor' && editorStore.hasUnsavedChanges) {
        const shouldContinue = await editorStore.resolveUnsavedChangesBeforeProceed('leave the editor')
        if (!shouldContinue) {
            return false
        }
    }

    const needsAuthStatus = to.meta.appChrome || to.name === 'login' || to.name === 'signup'

    if (needsAuthStatus) {
        try {
            await authStore.fetchSession(true)
        } catch (err) {
            emitToastError({
                title: 'Could not load session',
                message: isApiError(err)
                    ? getApiErrorMessage(err)
                    : 'Please check your connection and try again.',
            })

            if (to.meta.appChrome) {
                return false
            }
        }
    }

    if (to.meta.appChrome && !authStore.isAuthenticated) {
        authStore.clearWorkspaceState()
        return {
            name: 'login',
            query: { redirect: to.fullPath },
        }
    }

    if (authStore.isAuthenticated && to.meta.requiresBilling && !authStore.hasBillingAccess) {
        return {
            name: 'billing',
            query: { redirect: to.fullPath },
        }
    }

    if ((to.name === 'login' || to.name === 'signup') && authStore.isAuthenticated) {
        const redirect = typeof to.query.redirect === 'string' ? to.query.redirect : '/app'
        if (!authStore.hasBillingAccess) {
            return {
                name: 'billing',
                query: redirect.startsWith('/') ? { redirect } : undefined,
            }
        }
        return redirect.startsWith('/') ? redirect : { name: 'app-home' }
    }

    if (to.meta.appChrome) {
        clientStore.syncClientIdWithLS(authStore.account?.id ?? null)
    }

    if (to.meta.appChrome && authStore.hasBillingAccess && !clientStore.hasLoaded) {
        try {
            await clientStore.load()
        } catch (err) {
            emitToastError({
                title: 'Could not load clients',
                message: isApiError(err)
                    ? getApiErrorMessage(err)
                    : 'Please check your connection and try again.',
            })
        }
    }

    if (to.meta.appChrome && authStore.hasBillingAccess && !settingsStore.hasSettings) {
        try {
            await settingsStore.fetchSettings()
        } catch (err) {
            emitToastError({
                title: 'Could not load settings',
                message: isApiError(err)
                    ? getApiErrorMessage(err)
                    : 'Please check your connection and try again.',
            })
        }
    }

    if (to.meta.requiresClients && !clientStore.hasClients) {
        return { name: 'app-home' }
    }

    if (to.meta.requiresSelectedClient && !clientStore.selectedClient) {
        emitGuardToast(to.meta.guardToast)
        return { name: 'app-home' }
    }

    return true
})

export default router

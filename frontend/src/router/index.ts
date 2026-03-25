import { createRouter, createWebHistory } from 'vue-router'
import { useClientStore } from '@/stores/clients'
import { emitToastError } from '@/utils/toast'
import { getApiErrorMessage, isApiError } from '@/utils/apiErrors'

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        {
            path: '/',
            name: 'home',
            component: () => import('@/views/HomeView.vue'),
        },
        {
            path: '/clients',
            name: 'clients',
            component: () => import('@/views/ClientsView.vue'),
            meta: { requiresClients: true },
        },
        {
            path: '/invoice',
            name: 'invoice',
            component: () => import('@/views/InvoiceView.vue'),
            meta: { requiresSelectedClient: true },
        },
        {
            path: '/editor',
            name: 'editor',
            component: () => import('@/views/EditorView.vue'),
            meta: { requiresSelectedClient: true },
        },
    ],
    linkActiveClass: 'router-active',
})

router.beforeEach(async (to) => {
    const clientStore = useClientStore()

    clientStore.syncClientIdWithLS()

    if ((to.meta.requiresClients || to.meta.requiresSelectedClient) && !clientStore.hasLoaded) {
        try {
            await clientStore.load()
        } catch (err) {
            emitToastError({
                title: 'Could not load clients',
                message: isApiError(err)
                    ? getApiErrorMessage(err)
                    : 'Please check your connection and try again.',
            })
            return { name: 'home' }
        }
    }

    if (to.meta.requiresClients && !clientStore.hasClients) {
        return { name: 'home' }
    }

    if (to.meta.requiresSelectedClient && !clientStore.selectedClient) {
        return { name: 'home' }
    }

    return true
})

export default router

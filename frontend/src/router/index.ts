import { createRouter, createWebHistory } from 'vue-router'
import { useClientStore } from '@/stores/clients'

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

  if ((to.meta.requiresClients || to.meta.requiresSelectedClient) && !clientStore.hasLoaded) {
    await clientStore.load()
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

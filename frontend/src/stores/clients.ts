// stores/clientStore.ts
import { defineStore } from 'pinia'
import { ref, watch, computed } from 'vue'
import { createNewClient, getClients, deleteClient, updateClient } from '@/utils/clients/fetch'
import type { Client, UpdateClientInput } from '@/utils/clients/fetch'

export const useClientStore = defineStore('clients', () => {
  const clients = ref<Client[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  async function load() {
    isLoading.value = true
    error.value = null
    try {
      clients.value = await getClients()
    } catch (e: any) {
      error.value = e?.message ?? 'Failed to load clients'
      clients.value = []
    } finally {
      isLoading.value = false
    }
  }

  // localStorage selection
  const loadSelectedClientId = () => {
    const savedId = localStorage.getItem('selectedClientId')
    return savedId ? parseInt(savedId, 10) : null
  }
  const selectedClientId = ref<number | null>(loadSelectedClientId())

  const selectedClient = computed<Client | null>({
    get() {
      return clients.value.find((c) => c.id === selectedClientId.value) ?? null
    },
    set(client) {
      selectedClientId.value = client ? client.id : null
    },
  })

  watch(selectedClientId, (newValue) => {
    if (newValue != null) localStorage.setItem('selectedClientId', String(newValue))
  })

  async function createNew(client: Omit<Client, 'id'>) {
    error.value = null
    try {
      const created = await createNewClient(client)
      clients.value.push(created)
    } catch (e: any) {
      error.value = e?.message ?? 'Failed to create client'
      throw e
    }
  }
  async function remove(id: number) {
    if (typeof id !== 'number') {
      throw new Error(`remove() expected number id, got ${typeof id}`)
    }

    error.value = null
    await deleteClient(id)
    clients.value = clients.value.filter((c) => c.id !== id)
  }

  async function edit(
    id: number,
    patch: { name: string; company_name: string; email: string; address: string },
  ) {
    error.value = null
    const updated = await updateClient(id, patch)
    clients.value = clients.value.map((c) => (c.id === id ? updated : c))
  }

  const hasClients = computed(() => clients.value.length > 0)

  return {
    clients,
    selectedClient,
    selectedClientId,
    load,
    createNew,
    edit,
    remove,
    hasClients,
    isLoading,
    error,
  }
})

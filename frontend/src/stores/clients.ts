import { defineStore } from 'pinia'
import { ref, watch, computed } from 'vue'
import { createNewClient, getClients, deleteClient, updateClient } from '@/utils/clients/fetch'
import type { Client, UpdateClientInput } from '@/utils/clients/fetch'

export const useClientStore = defineStore('clients', () => {
  const clients = ref<Client[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  const hasLoaded = ref(false)
  async function load() {
    if (hasLoaded.value) return // could cache the req and return if loaded
    isLoading.value = true
    error.value = null
    try {
      const data = await getClients()
      // if no clients and backend returns null
      clients.value = Array.isArray(data) ? data : []
    } catch (e: any) {
      error.value = e?.message ?? 'Failed to load clients'
      clients.value = []
    } finally {
      isLoading.value = false
      hasLoaded.value = true
    }
  }

  /** Used across app  */
  const hasClients = computed(() => clients.value.length > 0)

  // Client Selection Localstorage Sync
  const LS_KEY = 'selectedClientId'

  const loadSelectedClientId = (): number | null => {
    const saved = localStorage.getItem(LS_KEY)
    if (!saved) return null
    const n = Number(saved)
    return Number.isInteger(n) && n > 0 ? n : null
  }

  const selectedClientId = ref<number | null>(loadSelectedClientId())

  const selectedClient = computed<Client | null>({
    get() {
      const id = selectedClientId.value
      if (id == null) return null
      return clients.value.find((c) => c.id === id) ?? null
    },
    set(client) {
      selectedClientId.value = client ? client.id : null
    },
  })

  // Persist selection changes and clear storage when selection cleared
  watch(selectedClientId, (id) => {
    if (id == null) {
      localStorage.removeItem(LS_KEY)
    } else {
      localStorage.setItem(LS_KEY, String(id))
    }
  })

  // After clients load/change validates stored selection
  // If no clients/selected id no longer exists -> clear and remove from localStorage
  watch(
    () => [hasLoaded.value, clients.value] as const,
    ([loaded]) => {
      if (!loaded) return

      // if no clients clear selection
      if (clients.value.length === 0) {
        if (selectedClientId.value != null) selectedClientId.value = null
        localStorage.removeItem(LS_KEY)
        return
      }

      // if selected id doesn't exist anymore (on delete) clear it
      const id = selectedClientId.value
      if (id != null && !clients.value.some((c) => c.id === id)) {
        selectedClientId.value = null
        localStorage.removeItem(LS_KEY)
      }
    },
    { deep: false },
  )
  async function createNew(client: Omit<Client, 'id'>) {
    error.value = null
    try {
      if (client.name.length <= 0) {
        throw new Error('Name cannot be empty')
      }

      const created = await createNewClient(client)

      clients.value.push(created)
    } catch (e: any) {
      error.value = e?.message ?? 'Failed to create client'
      throw e
    }
  }

  async function remove(id: number) {
    error.value = null
    try {
      await deleteClient(id)
      clients.value = clients.value.filter((c) => c.id !== id)
    } catch (e: any) {
      error.value = e?.message ?? 'Failed to delete client'
      throw e
    }
  }

  async function edit(id: number, patch: UpdateClientInput) {
    error.value = null
    try {
      const updated = await updateClient(id, patch)
      clients.value = clients.value.map((c) => (c.id === id ? updated : c))
    } catch (e: any) {
      error.value = e?.message ?? 'Failed to update client'
      throw e
    }
  }

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
    hasLoaded,
    error,
  }
})

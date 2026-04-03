import { defineStore } from 'pinia'
import { ref, watch, computed } from 'vue'
import { createNewClient, getClients, deleteClient, updateClient } from '@/utils/clientHttpHandler'
import type { Client, UpdateClientInput } from '@/utils/clientHttpHandler'

export const useClientStore = defineStore('clients', () => {
    const clients = ref<Client[]>([])
    const isLoading = ref(false)
    const hasLoaded = ref(false)

    const LEGACY_LS_KEY = 'invoicer_selectedClientId'
    const storageKey = ref<string | null>(null)

    function storageKeyForAccount(accountID: number | null | undefined): string | null {
        const normalized = Number(accountID)
        if (!Number.isInteger(normalized) || normalized <= 0) return null
        return `${LEGACY_LS_KEY}:${normalized}`
    }

    async function load() {
        if (hasLoaded.value) return

        isLoading.value = true

        try {
            const data = await getClients()
            clients.value = Array.isArray(data) ? data : []
            hasLoaded.value = true
        } catch (err) {
            clients.value = []
            hasLoaded.value = false
            throw err
        } finally {
            isLoading.value = false
        }
    }

    const hasClients = computed(() => clients.value.length > 0)

    function getClientIdFromLS(key: string | null): number | null {
        if (!key) return null

        const clientLS = localStorage.getItem(key)
        if (!clientLS) return null

        const n = Number(clientLS)
        return Number.isInteger(n) && n > 0 ? n : null
    }

    function persistClientId(id: number | null) {
        const key = storageKey.value
        if (!key) return

        if (id == null) {
            localStorage.removeItem(key)
            return
        }

        localStorage.setItem(key, String(id))
    }

    const lsClientId = ref<number | null>(null)

    /**
        Sets clientId to the value inside LocalStorage 
     */
    function syncClientIdWithLS(accountID: number | null | undefined) {
        storageKey.value = storageKeyForAccount(accountID)
        localStorage.removeItem(LEGACY_LS_KEY)
        lsClientId.value = getClientIdFromLS(storageKey.value)
    }

    const selectedClient = computed<Client | null>({
        get() {
            const id = lsClientId.value
            if (id == null) return null
            return clients.value.find((c) => c.id === id) ?? null
        },
        set(client) {
            lsClientId.value = client ? client.id : null
            persistClientId(lsClientId.value)
        },
    })

    function selectClientById(id: number | null) {
        lsClientId.value = id
        persistClientId(lsClientId.value)
    }

    watch(
        () => [hasLoaded.value, clients.value] as const,
        ([loaded]) => {
            if (!loaded) return

            if (clients.value.length === 0) {
                if (lsClientId.value != null) lsClientId.value = null
                persistClientId(null)
                return
            }

            const id = lsClientId.value
            if (id != null && !clients.value.some((c) => c.id === id)) {
                lsClientId.value = null
                persistClientId(null)
            }
        },
        { deep: false },
    )

    // CRUD

    async function createNew(client: Omit<Client, 'id'>) {
        const created = await createNewClient(client)
        clients.value.push(created)
        return created
    }

    async function remove(id: number) {
        await deleteClient(id)
        clients.value = clients.value.filter((c) => c.id !== id)
    }

    async function edit(id: number, patch: UpdateClientInput) {
        const updated = await updateClient(id, patch)
        clients.value = clients.value.map((c) => (c.id === id ? updated : c))
        return updated
    }

    function reset() {
        clients.value = []
        isLoading.value = false
        hasLoaded.value = false
        persistClientId(null)
        lsClientId.value = null
        storageKey.value = null
        localStorage.removeItem(LEGACY_LS_KEY)
    }

    return {
        clients,
        selectedClient,
        lsClientId,
        load,
        createNew,
        edit,
        remove,
        selectClientById,
        hasClients,
        isLoading,
        hasLoaded,
        syncClientIdWithLS,
        reset,
    }
})

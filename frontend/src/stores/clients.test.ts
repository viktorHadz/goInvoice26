import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useClientStore } from '@/stores/clients'

const { storageState, localStorageMock } = vi.hoisted(() => {
    const storageState: Record<string, string> = {}

    return {
        storageState,
        localStorageMock: {
            getItem: vi.fn((key: string) =>
                Object.prototype.hasOwnProperty.call(storageState, key) ? storageState[key] : null,
            ),
            setItem: vi.fn((key: string, value: string) => {
                storageState[key] = String(value)
            }),
            removeItem: vi.fn((key: string) => {
                delete storageState[key]
            }),
        },
    }
})

vi.stubGlobal('localStorage', localStorageMock)

vi.mock('@/utils/clientHttpHandler', () => ({
    getClients: vi.fn(),
    createNewClient: vi.fn(),
    deleteClient: vi.fn(),
    updateClient: vi.fn(),
}))

describe('clients store', () => {
    beforeEach(() => {
        setActivePinia(createPinia())
        for (const key of Object.keys(storageState)) {
            delete storageState[key]
        }
        vi.clearAllMocks()
    })

    it('keeps selected client ids isolated per workspace account', () => {
        const store = useClientStore()

        store.syncClientIdWithLS(1)
        store.selectClientById(11)

        store.syncClientIdWithLS(2)
        expect(store.lsClientId).toBe(null)

        store.selectClientById(22)

        store.syncClientIdWithLS(1)
        expect(store.lsClientId).toBe(11)

        store.syncClientIdWithLS(2)
        expect(store.lsClientId).toBe(22)
    })

    it('drops the legacy unscoped storage key so it cannot bleed into another workspace', () => {
        localStorage.setItem('invoicer_selectedClientId', '99')

        const store = useClientStore()
        store.syncClientIdWithLS(1)

        expect(store.lsClientId).toBe(null)
        expect(localStorage.getItem('invoicer_selectedClientId')).toBe(null)
    })
})

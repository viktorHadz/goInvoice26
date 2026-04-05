import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useClientStore } from '@/stores/clients'
import { useProductStore } from '@/stores/products'

const { localStorageMock } = vi.hoisted(() => ({
    localStorageMock: {
        getItem: vi.fn(() => null),
        setItem: vi.fn(),
        removeItem: vi.fn(),
    },
}))

vi.stubGlobal('localStorage', localStorageMock)

vi.mock('@/utils/clientHttpHandler', () => ({
    getClients: vi.fn(),
    createNewClient: vi.fn(),
    deleteClient: vi.fn(),
    updateClient: vi.fn(),
}))

vi.mock('@/utils/productHttpHandler', () => ({
    listClientProducts: vi.fn(),
    createProduct: vi.fn(),
    updateProduct: vi.fn(),
    deleteProduct: vi.fn(),
    importProducts: vi.fn(),
}))

describe('products store', () => {
    beforeEach(() => {
        setActivePinia(createPinia())
        vi.clearAllMocks()
    })

    it('rejects import when no client is selected', async () => {
        const store = useProductStore()
        const file = new File(['name,unit price\nHemline,12.50\n'], 'styles.csv', {
            type: 'text/csv',
        })

        await expect(store.importCsv('style', file)).rejects.toThrow('No client selected')
    })

    it('reloads the selected client products after a successful import', async () => {
        const productHttp = await import('@/utils/productHttpHandler')
        vi.mocked(productHttp.importProducts).mockResolvedValue({
            createdCount: 2,
            clientId: 7,
            importKind: 'style',
        })
        vi.mocked(productHttp.listClientProducts).mockResolvedValue([
            {
                id: 99,
                productType: 'style',
                pricingMode: 'flat',
                productName: 'Hemline',
                flatPriceMinor: 1250,
                clientId: 7,
                created_at: '2026-04-05T10:00:00Z',
            },
        ])

        const clientStore = useClientStore()
        clientStore.lsClientId = 7

        const store = useProductStore()
        vi.clearAllMocks()
        vi.mocked(productHttp.importProducts).mockResolvedValue({
            createdCount: 2,
            clientId: 7,
            importKind: 'style',
        })
        vi.mocked(productHttp.listClientProducts).mockResolvedValue([
            {
                id: 99,
                productType: 'style',
                pricingMode: 'flat',
                productName: 'Hemline',
                flatPriceMinor: 1250,
                clientId: 7,
                created_at: '2026-04-05T10:00:00Z',
            },
        ])

        const file = new File(['name,unit price\nHemline,12.50\n'], 'styles.csv', {
            type: 'text/csv',
        })

        const result = await store.importCsv('style', file)

        expect(result.createdCount).toBe(2)
        expect(productHttp.importProducts).toHaveBeenCalledWith(7, 'style', file)
        expect(productHttp.listClientProducts).toHaveBeenCalledWith(7)
        expect(store.products).toHaveLength(1)
        expect(store.products[0]?.productName).toBe('Hemline')
    })
})

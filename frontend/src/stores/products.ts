import { defineStore } from 'pinia'
import { computed, ref, watch } from 'vue'
import { useClientStore } from './clients'
import {
    createProduct,
    deleteProduct,
    listClientProducts,
    updateProduct,
} from '@/utils/productHttpHandler'
import type { Product, ProductType, ProductUpsert } from '@/utils/productHttpHandler'

export const useProductStore = defineStore('products', () => {
    const clientStore = useClientStore()
    const open = ref(false)
    const products = ref<Product[]>([])
    const isLoading = ref(false)
    const loadError = ref<string | null>(null)

    const byType = computed(() => {
        const style: Product[] = []
        const sample: Product[] = []

        for (const p of products.value) {
            ;(p.productType === 'style' ? style : sample).push(p)
        }

        return { style, sample } satisfies Record<ProductType, Product[]>
    })

    let loadToken = 0

    async function reload() {
        const clientId = clientStore.lsClientId
        if (!clientId) {
            products.value = []
            isLoading.value = false
            loadError.value = null
            return
        }

        const token = ++loadToken
        isLoading.value = true
        loadError.value = null

        try {
            const data = await listClientProducts(clientId)
            if (token !== loadToken) return
            products.value = data
        } catch (err) {
            if (token !== loadToken) return
            loadError.value = err instanceof Error ? err.message : 'Could not load products.'
            throw err
        } finally {
            if (token === loadToken) {
                isLoading.value = false
            }
        }
    }

    watch(
        () => clientStore.lsClientId,
        () => void reload().catch(() => {}),
        { immediate: true },
    )

    async function create(input: ProductUpsert) {
        const clientId = clientStore.lsClientId
        if (!clientId) throw new Error('No client selected')

        const created = await createProduct(clientId, input)
        products.value = [...products.value, created]
        return created
    }

    async function update(productId: number, input: ProductUpsert) {
        const clientId = clientStore.lsClientId
        if (!clientId) throw new Error('No client selected')

        const updated = await updateProduct(clientId, productId, input)
        products.value = products.value.map((p) => (p.id === updated.id ? updated : p))
        return updated
    }

    async function remove(productId: number) {
        const clientId = clientStore.lsClientId
        if (!clientId) throw new Error('No client selected')

        await deleteProduct(clientId, productId)
        products.value = products.value.filter((p) => p.id !== productId)
    }

    return {
        products,
        open,
        byType,
        isLoading,
        loadError,
        reload,
        create,
        update,
        remove,
    }
})

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

    const products = ref<Product[]>([])
    const isLoading = ref(false)
    const error = ref<string | null>(null)

    const byType = computed(() => {
        const style: Product[] = []
        const sample: Product[] = []
        for (const p of products.value) (p.productType === 'style' ? style : sample).push(p)
        return { style, sample } satisfies Record<ProductType, Product[]>
    })

    let loadToken = 0
    async function reload() {
        const clientId = clientStore.selectedClientId
        if (!clientId) {
            products.value = []
            error.value = null
            isLoading.value = false
            return
        }

        const token = ++loadToken
        isLoading.value = true
        error.value = null

        try {
            const data = await listClientProducts(clientId)
            if (token !== loadToken) return
            products.value = data
        } catch (e: any) {
            if (token !== loadToken) return
            products.value = []
            error.value = e?.message ?? 'Failed to load products'
        } finally {
            if (token === loadToken) isLoading.value = false
        }
    }

    watch(
        () => clientStore.selectedClientId,
        () => void reload(),
        { immediate: true },
    )

    async function create(input: ProductUpsert) {
        const clientId = clientStore.selectedClientId
        if (!clientId) throw new Error('No client selected')
        const created = await createProduct(clientId, input)
        products.value = [...products.value, created]
        return created
    }

    async function update(productId: number, input: ProductUpsert) {
        const clientId = clientStore.selectedClientId
        if (!clientId) throw new Error('No client selected')
        const updated = await updateProduct(clientId, productId, input)
        products.value = products.value.map((p) => (p.id === updated.id ? updated : p))
        return updated
    }

    async function remove(productId: number) {
        const clientId = clientStore.selectedClientId
        if (!clientId) throw new Error('No client selected')
        await deleteProduct(clientId, productId)
        products.value = products.value.filter((p) => p.id !== productId)
    }

    return {
        products,
        byType,
        isLoading,
        error,
        reload,
        create,
        update,
        remove,
    }
})

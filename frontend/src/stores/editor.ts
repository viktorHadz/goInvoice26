import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { useClientStore } from './clients'
import { getInvAndRevNums } from '@/utils/editorHttpHandler'
import type { InvBookInvoice } from '@/components/editor/editorTypes'

export const useEditStore = defineStore('editStore', () => {
    const clientStore = useClientStore()

    const invoiceBook = ref<InvBookInvoice[]>([])
    const limit = ref(10)
    const offset = ref(0)
    const total = ref(0)
    const hasMore = ref(false)
    const isLoading = ref(false)
    const errorMessage = ref('')

    const canGoPrev = computed(() => offset.value > 0)
    const canGoNext = computed(() => hasMore.value)

    async function fetchInvoiceBook(reset = false) {
        const clientId = clientStore.selectedClient?.id

        if (!clientId) {
            clearInvoiceBook()
            return
        }

        if (reset) {
            offset.value = 0
        }

        isLoading.value = true
        errorMessage.value = ''

        try {
            const data = await getInvAndRevNums(clientId, limit.value, offset.value)

            invoiceBook.value = data.items
            total.value = data.total
            hasMore.value = data.hasMore
            limit.value = data.limit
            offset.value = data.offset
        } catch (error) {
            clearInvoiceBook()
            errorMessage.value =
                error instanceof Error ? error.message : 'Failed to load invoice book'
            throw error
        } finally {
            isLoading.value = false
        }
    }

    async function nextPage() {
        if (!hasMore.value || isLoading.value) return
        offset.value += limit.value
        await fetchInvoiceBook()
    }

    async function prevPage() {
        if (offset.value === 0 || isLoading.value) return
        offset.value = Math.max(0, offset.value - limit.value)
        await fetchInvoiceBook()
    }

    async function goToFirstPage() {
        offset.value = 0
        await fetchInvoiceBook()
    }

    function clearInvoiceBook() {
        invoiceBook.value = []
        offset.value = 0
        total.value = 0
        hasMore.value = false
        errorMessage.value = ''
    }

    return {
        invoiceBook,
        limit,
        offset,
        total,
        hasMore,
        isLoading,
        errorMessage,
        canGoPrev,
        canGoNext,
        fetchInvoiceBook,
        nextPage,
        prevPage,
        goToFirstPage,
        clearInvoiceBook,
    }
})

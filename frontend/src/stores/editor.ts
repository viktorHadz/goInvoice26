import { defineStore } from 'pinia'
import { useClientStore } from './clients'
import { getInvAndRevNums } from '@/utils/editorHttpHandler'
import { ref } from 'vue'
import type { InvBookInvoice } from '@/components/editor/editorTypes'

export const useEditStore = defineStore('editStore', () => {
    const clientStore = useClientStore()
    const invoiceBook = ref<InvBookInvoice[]>([])

    async function getInvoiceBook() {
        try {
            const id = clientStore.selectedClient?.id
            if (!id) {
                invoiceBook.value = []
                return
            }

            invoiceBook.value = await getInvAndRevNums(id)
        } catch (error) {
            invoiceBook.value = []
            throw error
        }
    }

    return { invoiceBook, getInvoiceBook }
})

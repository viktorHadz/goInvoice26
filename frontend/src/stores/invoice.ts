import { defineStore } from 'pinia'
import { reactive } from 'vue'
import { calcPrice } from '@/composables/calcPrice.js'
export const useInvoiceStore = defineStore('invoice', () => {
  const data = reactive({
    client: { id: 0, name: '', company: '', address: '' },
    invoiceNumber: '#S.A.M.-1',
    createdDate: '',
    dueDate: '',
    items: [],
    prices: {},
  })
  const add = (item) => {
    data.items.push(item)
  }
  const remove = () => {}
  const update = () => {}
  const totals = calcPrice()

  return { data, add, remove, update, totals }
})

import { defineStore } from 'pinia'
import { reactive } from 'vue'

export const useItemsStore = defineStore('items', () => {
  const items = reactive({
    style: [],
    sample: [],
  })

  const fetchItems = async (type, url) => {
    // Here I should be passing a client ID when I establish the backend. There should be checks
    // For selected client (should i crash the program if no selected? - to think about -
    // maybe have the select client call again here although i have it in an erarlier stage of the program so this shouldnt be an issue here ])
    try {
      const res = await fetch(url)
      const data = await res.json()
      const transformed = data.products.map((product) => ({
        id: `${type}-${product.id}`,
        name: product.title,
        type,
        price: product.price,
        time: type === 'sample' ? product.rating : 1,
      }))
      items[type].push(...transformed)
    } catch (error) {
      console.error(`Error fetching ${type}s:`, error)
    }
  }

  const createItem = (type, item) => {
    // I will need some real id here
    console.log('item type', type)
    console.log('item received', item)
    const newId = items[type].length - 1 + 1
    const newItem = {
      ...item,
      id: newId,
    }
    items[type].push(newItem)
    console.log('added new item here: ', items[type])
  }
  const updateItem = (type, item) => {
    const index = items[type].findIndex((i) => i.id === item.id)
    if (index !== -1) {
      Object.assign(items[type][index], item)
    }
  }

  const deleteItem = (type, item) => {
    const index = items[type].findIndex((i) => i.id === item.id)
    if (index !== -1) {
      items[type].splice(index, 1)
    }
  }

  return {
    items,
    fetchItems,
    deleteItem,
    updateItem,
    createItem,
  }
})

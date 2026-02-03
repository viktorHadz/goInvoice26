import { defineStore } from 'pinia'
import { reactive, ref, watch, computed } from 'vue'

export const useClientStore = defineStore('clients', () => {
  // Load a hardcoded list of clients
  const clients = reactive([
    {
      id: 1,
      name: 'Viktor',
      company: 'CompanyLTD',
      email: 'vikecah@gmail.com',
      address: '37A Glenton Road, Lewisham, London, SE13 5RS',
    },
    {
      id: 2,
      name: 'Anna',
      company: 'TechCorp',
      email: 'anna.techcorp@gmail.com',
      address: '123 Tech Street, Silicon Valley, CA, 94043',
    },
    {
      id: 3,
      name: 'John',
      company: 'Innovate Ltd',
      email: 'john.innovate@gmail.com',
      address: '456 Innovation Drive, New York, NY, 10001',
    },
    {
      id: 4,
      name: 'Emma',
      company: 'CreativeWorks',
      email: 'emma.creative@gmail.com',
      address: '78 Creative Road, Austin, TX, 73301',
    },
    {
      id: 5,
      name: 'Liam',
      company: 'Proto Inc.',
      email: 'liam.proto@gmail.com',
      address: '90 Proto Street, Boston, MA, 02110',
    },
    {
      id: 6,
      name: 'Olivia',
      company: 'Designify',
      email: 'olivia.designify@gmail.com',
      address: '22 Design Blvd, San Francisco, CA, 94102',
    },
    {
      id: 7,
      name: 'Noah',
      company: 'BuildRight',
      email: 'noah.build@gmail.com',
      address: '300 Build Ave, Seattle, WA, 98101',
    },
    {
      id: 8,
      name: 'Ava',
      company: 'NextGen Solutions',
      email: 'ava.nextgen@gmail.com',
      address: '15 NextGen Street, Denver, CO, 80203',
    },
  ])

  // Load the stored selected client id from localStorage
  const loadSelectedClientId = () => {
    const savedId = localStorage.getItem('selectedClientId')
    return savedId ? parseInt(savedId) : null
  }
  const selectedClientId = ref(loadSelectedClientId())

  // Computed property with a getter and a setter for two-way binding
  const selectedClient = computed({
    get() {
      return clients.find((client) => client.id === selectedClientId.value) || null
    },
    set(client) {
      selectedClientId.value = client ? client.id : null
    },
  })

  // Persist the selected client id to localStorage whenever it changes
  watch(selectedClientId, (newValue) => {
    if (newValue) {
      localStorage.setItem('selectedClientId', newValue.toString())
    }
  })

  // Create a new client
  const createNew = (client) => {
    const newId = clients.length > 0 ? Math.max(...clients.map((c) => c.id)) + 1 : 1
    clients.push({
      id: newId,
      ...client,
    })
  }

  // Update an existing client. If the client is the currently selected one,
  // the computed getter will automatically return the updated data.
  const update = (client) => {
    const index = clients.findIndex((c) => c.id === client.id)
    if (index !== -1) {
      Object.assign(clients[index], client)
    }
  }

  // Delete a client and clear the selection if the deleted client was selected
  const del = (client) => {
    const index = clients.findIndex((c) => c.id === client.id)
    if (index !== -1) {
      clients.splice(index, 1)

      if (selectedClient.value && selectedClient.value.id === client.id) {
        selectedClient.value = null
      }
    }
  }

  const hasClients = ref(true)
  watch(clients, () => {
    if (clients.length > 0) {
      hasClients.value = true
    } else {
      hasClients.value = false
    }
  })
  return { clients, selectedClient, createNew, update, del, hasClients }
})

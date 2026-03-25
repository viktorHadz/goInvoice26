<script setup lang="ts">
import { useClientStore } from '@/stores/clients'
import TheInput from '../UI/TheInput.vue'
import TheButton from '../UI/TheButton.vue'
import { UserPlusIcon } from '@heroicons/vue/24/outline'
import { handleActionError } from '@/utils/errors/handleActionError'

const clientStore = useClientStore()

const resetFormNoClient = () => {
  noClientsFormNew.name = ''
  noClientsFormNew.companyName = ''
  noClientsFormNew.email = ''
  noClientsFormNew.address = ''
}
const noClientsFormNew = { name: '', companyName: '', email: '', address: '' }

async function addNewClient() {
  try {
    await clientStore.createNew(noClientsFormNew)
    resetFormNoClient()
  } catch (err) {
    handleActionError(err, {
      toastTitle: 'Could not create client',
      mapFields: false,
    })
  }
}
</script>
<template>
  <div class="">
    <h2 class="text-2xl font-bold text-sky-600 dark:text-emerald-400">
      Create a client to continue
    </h2>

    <div class="mt-8 flex w-full flex-col gap-8">
      <TheInput
        placeholder="Name"
        id="no-client-add-name-id-1"
        name="client-name"
        type="text"
        autocomplete="name"
        required
        v-model="noClientsFormNew.name"
      />

      <TheInput
        placeholder="Company"
        id="no-client-add-company-1"
        name="client-company"
        type="text"
        autocomplete="organization"
        required
        v-model="noClientsFormNew.companyName"
      />
      <TheInput
        placeholder="Email"
        id="no-client-add-email-1"
        name="client-email"
        type="text"
        autocomplete="email"
        required
        v-model="noClientsFormNew.email"
      />
      <TheInput
        placeholder="Address"
        id="no-client-add-address-1"
        name="client-address"
        type="text"
        autocomplete="address"
        required
        v-model="noClientsFormNew.address"
      />
      <div class="flex w-full">
        <TheButton
          @click="addNewClient()"
          class="flex gap-2"
        >
          <UserPlusIcon class="size-5"></UserPlusIcon>
          <p class="text-sm">Create client</p>
        </TheButton>
      </div>
    </div>
  </div>
</template>

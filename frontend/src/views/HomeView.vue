<script setup lang="ts">
import TheButton from '@/components/UI/TheButton.vue'
import { RouterLink } from 'vue-router'
import { useClientStore } from '@/stores/clients'
import NoClients from '@/components/clients/NoClients.vue'
import heroImg from '@/assets/images/vik_wave.svg'
import TheDropdown from '@/components/UI/TheDropdown.vue'
import { UserIcon } from '@heroicons/vue/24/outline'
// The idea is to disable invoice and editor if no client is selected
const clientStore = useClientStore()
</script>

<template>
  <main class="container mx-auto py-12">
    <div class="min-h-96">
      <div class="relative">
        <div class="absolute top-22 left-55 z-10">
          <h1
            class="text-acc text-2xl font-bold tracking-tighter uppercase sm:text-3xl sm:text-nowrap lg:text-4xl"
          >
            Welcome to Invoicer
          </h1>
          <p class="text-fg mt-4 max-w-xl text-lg">Your one stop shop for everything invoice</p>
        </div>
        <div class="absolute top-20 right-45">
          <img
            :src="heroImg"
            alt="Invoicer dashboard preview"
            class="mx-auto w-full max-w-2xs lg:ml-auto"
          />
        </div>
      </div>
    </div>

    <div class="container-menu mt-6">
      <NoClients v-if="!clientStore.hasClients"></NoClients>
      <div v-else-if="clientStore.hasClients && !clientStore.selectedClient">
        <p class="tracking-wide">Please select a client to continue:</p>
        <TheDropdown
          v-model="clientStore.selectedClient"
          :options="clientStore.clients"
          placeholder="No client selected"
          :left-icon="UserIcon"
          label-key="name"
          value-key="id"
        ></TheDropdown>
      </div>
      <div v-else>
        <p class="text-acc font-bold">Where to captain?</p>
        <div class="mt-4 flex gap-x-4">
          <RouterLink to="/clients">
            <TheButton>clients</TheButton>
          </RouterLink>

          <RouterLink to="/invoice">
            <TheButton>invoice</TheButton>
          </RouterLink>
          <RouterLink to="/editor">
            <TheButton>editor</TheButton>
          </RouterLink>
        </div>
      </div>
    </div>
  </main>
</template>

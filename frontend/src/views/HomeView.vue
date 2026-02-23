<script setup lang="ts">
import TheButton from '@/components/UI/TheButton.vue'
import { RouterLink } from 'vue-router'
import { useClientStore } from '@/stores/clients'
import SelectClient from '@/components/clients/SelectClient.vue'
import NoClients from '@/components/clients/NoClients.vue'
import heroImg from '@/assets/images/vik_wave.svg'
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
        <SelectClient></SelectClient>
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

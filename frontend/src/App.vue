<script setup lang="ts">
import { RouterView } from 'vue-router'
import { RouterLink } from 'vue-router'
import NoSelectedClientModal from './components/clients/NoSelectedClient.vue'
import { useClientStore } from '@/stores/clients'
import LeTopBar from './components/UI/LeTopBar.vue'

const clientStore = useClientStore()
</script>

<template>
  <div class="bg-primary text-fg flex min-h-screen w-full transition duration-500 will-change-transform">
    <header>
      <nav class="navigation" v-if="clientStore.hasClients">
        <RouterLink to="/clients" class="link"> Clients </RouterLink>

        <RouterLink to="/invoice" class="link"> Invoice </RouterLink>

        <RouterLink to="/editor" class="link"> Editor</RouterLink>
      </nav>
    </header>
    <main class="main-content">
      <RouterView />
      <LeTopBar />

      <!-- Modals -- Add parent component to decluter if needed  -->
      <NoSelectedClientModal></NoSelectedClientModal>
    </main>
  </div>
</template>

<style scoped></style>

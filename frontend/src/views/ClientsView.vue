<script setup lang="ts">
import ClientsTable from '@/components/clients/ClientsTable.vue'
import NoClients from '@/components/clients/NoClients.vue'
import { useClientStore } from '@/stores/clients'
const clientStore = useClientStore()
</script>

<template>
    <div class="mt-26">
        <div v-if="clientStore.hasClients === false">
            <NoClients></NoClients>
        </div>
        <div v-else>
            <ClientsTable></ClientsTable>
        </div>
    </div>
</template>

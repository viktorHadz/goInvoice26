<script setup lang="ts">
import ClientsTable from '@/components/clients/ClientsTable.vue'
import NoClients from '@/components/clients/NoClients.vue'
import { useClientStore } from '@/stores/clients'
import { onMounted } from 'vue'
const clientStore = useClientStore()

onMounted(() => {
    clientStore.load()
})
</script>

<template>
    <div class="mt-26">
        <ClientsTable v-if="clientStore.hasClients"></ClientsTable>
        <NoClients v-else></NoClients>
    </div>
</template>

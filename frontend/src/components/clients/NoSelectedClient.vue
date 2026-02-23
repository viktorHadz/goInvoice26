<script setup lang="ts">
import { useClientStore } from '@/stores/clients'
import SelectClient from './SelectClient.vue'
const clientStore = useClientStore()
</script>
<template>
  <Teleport to="body">
    <div v-if="clientStore.selectedClient === null && clientStore.hasClients" class="backdrop">
      <div class="modal-outer rounded bg-transparent!">
        <div class="p-2">
          <SelectClient
            select-title="No client selected"
            select-title-class="text-white text-lg"
          ></SelectClient>
          <slot></slot>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref } from 'vue'
import SelectClient from '../clients/SelectClient.vue'
import LeDarkMode from './LeDarkMode.vue'
import ItemsEditor from '@/components/items/ItemsEditor.vue'
import { useClientStore } from '@/stores/clients'

const clientStore = useClientStore()

const show = ref(false)
function toggleTopBar() {
  show.value = !show.value
  console.log('Show?:', show.value)
}
</script>
<template>
  <div v-if="clientStore.hasClients" class="absolute top-0 right-4">
    <div
      class="bg-sec relative rounded-b-md border-x border-b border-neutral-400 transition-transform delay-150 duration-300 ease-in-out will-change-transform dark:border-neutral-600"
      :class="show === true ? 'translate-y' : '-translate-y-12'"
    >
      <div class="grid grid-cols-5 items-center gap-x-4 gap-y-2">
        <div class="col-span-3 row-start-1 mx-2 pt-2">
          <SelectClient
            title="select a client"
            select-title=""
            select-title-class="text-xs font-normal"
          >
          </SelectClient>
        </div>
        <div class="col-span-1 row-start-1 mt-2 flex w-full justify-center self-center">
          <LeDarkMode></LeDarkMode>
        </div>
        <div class="col-span-1 row-start-1 mt-1.5 pr-2"><ItemsEditor></ItemsEditor></div>
        <button
          @click="toggleTopBar()"
          :title="!show ? 'click to expand' : ''"
          class="hover:bg-acc/70 dark:hover:bg-acc/10 dark:text-fg dark:hover:text-acc col-span-5 grid cursor-pointer grid-cols-subgrid grid-rows-subgrid items-center rounded-b-md text-xs"
        >
          <div class="col-span-3 row-start-2 mb-0.5 tracking-wider">Selected Client</div>
          <div class="col-span-1 row-start-2 mb-0.5">Theme</div>
          <div class="col-span-1 row-start-2 mb-0.5 pr-2">Items</div>
        </button>
      </div>
    </div>
  </div>
</template>

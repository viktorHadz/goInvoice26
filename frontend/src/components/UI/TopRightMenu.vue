<script setup lang="ts">
import { ref, watch } from 'vue'
import DarkMode from './DarkMode.vue'
import ProductsEditor from '@/components/items/ProductsEditor.vue'
import TheDropdown from './TheDropdown.vue'
import { useClientStore } from '@/stores/clients'
import { ChevronUpDownIcon, UserIcon } from '@heroicons/vue/24/outline'

const show = ref(localStorage.getItem('topBarShow') === 'true')

function toggleTopBar() {
  show.value = !show.value
}

watch(show, (newVal) => {
  localStorage.setItem('topBarShow', String(newVal))
})

const clientStore = useClientStore()
</script>
<template>
  <div class="absolute top-0 right-4">
    <div
      class="bg-sec relative min-w-xs rounded-b-md border-x border-b border-neutral-400 transition-transform delay-150 duration-300 ease-in-out will-change-transform dark:border-neutral-600"
      :class="show === true ? 'translate-y' : '-translate-y-12'"
    >
      <div class="grid grid-cols-5 items-center gap-x-4 gap-y-2">
        <!-- Client select dropdown -->
        <div class="col-span-3 row-start-1 mx-2 pt-2">
          <TheDropdown
            v-model="clientStore.selectedClient"
            :options="clientStore.clients"
            placeholder="Select Client"
            :left-icon="UserIcon"
            :right-icon="ChevronUpDownIcon"
            label-key="name"
            value-key="id"
          ></TheDropdown>
        </div>
        <!-- Dark Light Mode: injects data-theme into index html and loads main.css -->
        <div class="col-span-1 row-start-1 mt-2 flex w-full justify-center self-center">
          <DarkMode></DarkMode>
        </div>
        <!-- Products for selected client -->
        <div class="col-span-1 row-start-1 mt-1.5 pr-2"><ProductsEditor></ProductsEditor></div>
        <button
          @click="toggleTopBar()"
          :title="!show ? 'click to expand' : ''"
          class="hover:bg-acc/70 dark:hover:bg-acc/10 dark:text-fg dark:hover:text-acc col-span-5 grid cursor-auto grid-cols-subgrid grid-rows-subgrid items-center rounded-b-md text-xs"
        >
          <div class="col-span-3 row-start-2 mb-0.5 tracking-wider">Client Select</div>
          <div class="col-span-1 row-start-2 mb-0.5">Theme</div>
          <div class="col-span-1 row-start-2 mb-0.5 pr-2">Items</div>
        </button>
      </div>
    </div>
  </div>
</template>

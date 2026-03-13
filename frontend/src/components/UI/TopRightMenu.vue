<script setup lang="ts">
import { ref, watch } from 'vue'
import DarkMode from './DarkMode.vue'
import ProductsEditor from '@/components/items/ProductsEditor.vue'
import TheDropdown from './TheDropdown.vue'
import { useClientStore } from '@/stores/clients'
import { ChevronUpDownIcon, UserIcon } from '@heroicons/vue/24/outline'
import { useShortcuts, type ShortcutDefinition } from '@/composables/keyHandlers'
import { useProductStore } from '@/stores/products'
import { useTheme } from '@/composables/theme'
import TheSettings from './TheSettings.vue'

const show = ref(localStorage.getItem('topBarShow') === 'true')

function toggleTopBar() {
  show.value = !show.value
}

watch(show, (newVal) => {
  localStorage.setItem('topBarShow', String(newVal))
})

const clientStore = useClientStore()
const productStore = useProductStore()
const { mode } = useTheme()

const shortcuts: ShortcutDefinition[] = [
  { key: 'i', modifiers: ['ctrl'], action: () => (productStore.open = true) },
  {
    key: 'm',
    modifiers: ['ctrl', 'shift'],
    action: () => {
      if (mode.value === 'light') {
        mode.value = 'dark'
      } else {
        mode.value = 'light'
      }
    },
  },
]
useShortcuts(shortcuts)
</script>

<template>
  <div class="fixed top-0 right-1/2 z-50 translate-x-1/2 sm:right-4 sm:translate-x-0">
    <div
      class="relative w-72 overflow-visible transition-transform duration-200 ease-out sm:w-80"
      :class="show ? 'translate-y-0' : '-translate-y-[calc(100%-33px)]'"
    >
      <div
        class="rounded-b-2xl border-x border-b border-zinc-200 bg-white shadow-lg dark:border-zinc-800 dark:bg-zinc-900"
      >
        <!-- top row -->
        <div class="grid grid-cols-6 items-center px-2 pt-2">
          <div class="col-span-3 min-w-0 overflow-x-clip px-0.5">
            <TheDropdown
              v-model="clientStore.selectedClient"
              :options="clientStore.clients"
              placeholder="Select Client"
              :left-icon="UserIcon"
              :right-icon="ChevronUpDownIcon"
              label-key="name"
              value-key="id"
              input-class="py-1 font-medium"
            />
          </div>

          <div class="col-span-1 flex justify-center">
            <DarkMode />
          </div>

          <div class="col-span-1 flex justify-center">
            <ProductsEditor />
          </div>

          <div class="col-span-1 flex justify-center">
            <TheSettings />
          </div>
        </div>

        <!-- bottom handle-->
        <button
          type="button"
          @click="toggleTopBar()"
          :aria-expanded="show"
          class="text-mini mt-2 h-8 w-full rounded-b-2xl border-t border-zinc-200 px-2 font-medium text-zinc-600 hover:bg-zinc-50 hover:text-zinc-900 focus-visible:ring-2 focus-visible:ring-zinc-900/10 focus-visible:outline-none dark:border-zinc-800 dark:text-zinc-400 dark:hover:bg-zinc-800/60 dark:hover:text-zinc-100"
        >
          <span class="truncate">Quick Menu</span>
        </button>
      </div>
    </div>
  </div>
</template>

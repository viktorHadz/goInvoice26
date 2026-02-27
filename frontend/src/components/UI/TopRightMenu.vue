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
    <div class="absolute top-0 right-3 z-50 sm:right-4">
        <!-- overflow-visible so dropdown options never clip -->
        <div
            class="relative w-80 overflow-visible transition-transform duration-200 ease-out"
            :class="show ? 'translate-y-0' : '-translate-y-[calc(100%-40px)]'"
        >
            <div
                class="rounded-b-2xl border-x border-b border-zinc-200 bg-white shadow-lg dark:border-zinc-800 dark:bg-zinc-900"
            >
                <!-- top row -->
                <div class="grid grid-cols-5 items-center gap-3 px-2 pt-2">
                    <div class="col-span-3 min-w-0 overflow-x-clip px-0.5">
                        <TheDropdown
                            v-model="clientStore.selectedClient"
                            :options="clientStore.clients"
                            placeholder="Select Client"
                            :left-icon="UserIcon"
                            :right-icon="ChevronUpDownIcon"
                            label-key="name"
                            value-key="id"
                            class="w-full"
                        />
                    </div>

                    <div class="col-span-1 flex justify-center">
                        <DarkMode />
                    </div>

                    <div class="col-span-1 flex justify-center">
                        <ProductsEditor />
                    </div>
                </div>

                <!-- bottom handle-->
                <button
                    @click="toggleTopBar()"
                    :title="show ? 'Collapse' : 'Expand'"
                    :aria-expanded="show"
                    class="mt-2 grid h-10 w-full grid-cols-[3fr_1fr_1fr_auto] items-center rounded-b-2xl border-t border-zinc-200 px-2 text-xs font-medium text-zinc-600 hover:bg-zinc-50 hover:text-zinc-900 focus-visible:ring-2 focus-visible:ring-zinc-900/10 focus-visible:outline-none dark:border-zinc-800 dark:text-zinc-400 dark:hover:bg-zinc-800/60 dark:hover:text-zinc-100"
                >
                    <span class="truncate">Client select</span>
                    <span class="justify-self-center truncate">Theme</span>
                    <span class="justify-self-end truncate pr-2">Items</span>
                </button>
            </div>
        </div>
    </div>
</template>

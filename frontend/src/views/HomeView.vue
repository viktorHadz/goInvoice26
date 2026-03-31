<script setup lang="ts">
import TheButton from '@/components/UI/TheButton.vue'
import { RouterLink } from 'vue-router'
import { useClientStore } from '@/stores/clients'
import NoClients from '@/components/clients/NoClients.vue'
import heroImg from '@/assets/images/vik_wave.svg'
import TheDropdown from '@/components/UI/TheDropdown.vue'
import {
  BriefcaseIcon,
  DocumentTextIcon,
  PencilSquareIcon,
  UserIcon,
  UsersIcon,
} from '@heroicons/vue/24/outline'
import TheTooltip from '@/components/UI/TheTooltip.vue'
import { useProductStore } from '@/stores/products'

// The idea is to disable invoice and editor if no client is selected
const clientStore = useClientStore()
const productStore = useProductStore()

const features = [
  { title: 'Clients', body: 'Create and manage' },
  { title: 'Items', body: 'Build inventory' },
  { title: 'Invoices', body: 'Generate quickly' },
  { title: 'Editor', body: 'Track and audit' },
]
</script>
<template>
  <main class="mx-auto w-full max-w-4xl 2xl:max-w-5xl">
    <!-- Hero -->
    <section class="grid items-center gap-8 lg:grid-cols-2">
      <div>
        <h1
          class="text-3xl font-bold tracking-tight text-sky-600 uppercase sm:text-5xl dark:text-emerald-400"
        >
          Welcome to Invoice And Go
        </h1>

        <p class="mt-3 max-w-xl text-base text-zinc-600 sm:text-xl dark:text-zinc-300">
          Your one stop invoice shop
        </p>

        <!-- quick points -->
        <div class="mt-8 grid gap-3 sm:grid-cols-2">
          <ul
            v-for="feature in features"
            :key="feature.body"
          >
            <li>
              <div
                class="hdr-grid rounded-2xl border border-zinc-300 bg-white px-4 py-3 text-sm text-zinc-700 shadow-sm dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-200"
              >
                <div class="font-semibold text-zinc-900 dark:text-zinc-100">
                  {{ feature.title }}
                </div>
                <div class="mt-1 text-zinc-600 dark:text-zinc-400">
                  {{ feature.body }}
                </div>
              </div>
            </li>
          </ul>
        </div>

        <!-- how to-->
        <div class="mt-8 text-zinc-600 dark:text-zinc-400">
          <div class="font-medium text-zinc-900 dark:text-zinc-200">Quick start:</div>
          <ol class="mt-2 list-decimal space-y-1 pl-5 text-sm">
            <li class="cursor-default hover:text-sky-600 dark:hover:text-emerald-400">
              Select a client from the top-right menu
            </li>
            <li class="cursor-default hover:text-sky-600 dark:hover:text-emerald-400">
              Create some items for the selected client
            </li>
            <li class="cursor-default hover:text-sky-600 dark:hover:text-emerald-400">
              Jump into invoice to generate a PDF
            </li>
            <li class="cursor-default hover:text-sky-600 dark:hover:text-emerald-400">
              Go to editor for your invoice book or to make revisions
            </li>
          </ol>
        </div>
      </div>

      <div class="flex justify-center lg:justify-end">
        <img
          :src="heroImg"
          alt="Invoicer mascot"
          class="w-full max-w-sm select-none"
          draggable="false"
        />
      </div>
    </section>

    <!-- Actions / State -->
    <section class="mt-12">
      <div
        class="hdr-grid rounded-2xl border border-zinc-300 bg-white p-4 shadow-md sm:p-5 dark:border-zinc-800 dark:bg-zinc-900/50"
      >
        <Transition
          name="fade-down-up"
          mode="out-in"
        >
          <NoClients v-if="!clientStore.hasClients"></NoClients>

          <div
            v-else-if="clientStore.hasClients && !clientStore.selectedClient"
            class="space-y-3"
          >
            <p class="text-sm text-sky-600 dark:text-emerald-400">
              Please select a client to continue:
            </p>

            <div class="max-w-xl">
              <TheDropdown
                v-model="clientStore.selectedClient"
                :options="clientStore.clients"
                placeholder="No client selected"
                :left-icon="UserIcon"
                label-key="name"
                value-key="id"
              ></TheDropdown>
            </div>

            <div class="text-xs text-zinc-600 dark:text-zinc-400">
              Tip: create / edit clients in the Clients page.
            </div>

            <RouterLink to="/app/clients">
              <TheButton class="cursor-pointer">clients</TheButton>
            </RouterLink>
          </div>

          <div v-else>
            <div class="mb-3 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <h2 class="text-sm font-semibold text-zinc-900 dark:text-zinc-100">
                Continue where you left off
              </h2>
              <span class="text-sm font-semibold text-sky-600 dark:text-emerald-400">
                client → items → invoice → edit
              </span>
            </div>
            <div class="grid grid-cols-2 gap-3 sm:mt-12 sm:flex sm:flex-wrap">
              <RouterLink to="/app/clients">
                <TheButton class="cursor-pointer">
                  <UsersIcon class="size-4"></UsersIcon>
                  clients
                </TheButton>
              </RouterLink>

              <TheTooltip side="bottom">
                <template #content>
                  <span class="mr-1 text-sky-600 dark:text-emerald-400">Shortcut:</span>
                  <kbd>Ctrl</kbd>
                  +
                  <kbd>i</kbd>
                </template>

                <TheButton
                  @click="productStore.open = true"
                  class="cursor-pointer"
                >
                  <BriefcaseIcon class="size-4"></BriefcaseIcon>
                  items
                </TheButton>
              </TheTooltip>

              <RouterLink to="/app/invoice">
                <TheButton class="cursor-pointer">
                  <DocumentTextIcon class="size-4" />
                  invoice
                </TheButton>
              </RouterLink>

              <RouterLink to="/app/editor">
                <TheButton class="cursor-pointer">
                  <PencilSquareIcon class="size-4" />
                  editor
                </TheButton>
              </RouterLink>
            </div>
          </div>
        </Transition>
      </div>
    </section>
  </main>
</template>

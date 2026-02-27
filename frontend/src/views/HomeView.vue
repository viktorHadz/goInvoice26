<script setup lang="ts">
import TheButton from '@/components/UI/TheButton.vue'
import { RouterLink } from 'vue-router'
import { useClientStore } from '@/stores/clients'
import NoClients from '@/components/clients/NoClients.vue'
import heroImg from '@/assets/images/vik_wave.svg'
import TheDropdown from '@/components/UI/TheDropdown.vue'
import { UserIcon } from '@heroicons/vue/24/outline'

// The idea is to disable invoice and editor if no client is selected
const clientStore = useClientStore()

const features = [
    { title: 'Clients', body: 'Create and manage' },
    { title: 'Items', body: 'Build inventory' },
    { title: 'Invoices', body: 'Generate quickly' },
    { title: 'Editor', body: 'Match demands' },
    { title: 'Invoice Book', body: 'Track and audit' },
]
</script>

<template>
    <main class="sm mx-auto w-full max-w-6xl px-4 py-18 sm:py-28">
        <!-- Hero -->
        <section class="grid items-center gap-8 lg:grid-cols-2">
            <div>
                <h1
                    class="text-3xl font-bold tracking-tight text-sky-600 uppercase sm:text-5xl dark:text-emerald-400"
                >
                    Welcome to Invoicer
                </h1>

                <p class="mt-3 max-w-xl text-base text-zinc-600 sm:text-xl dark:text-zinc-300">
                    Your one stop invoice shop
                </p>

                <!-- quick points -->
                <div class="mt-6 grid gap-3 sm:grid-cols-3">
                    <ul
                        v-for="feature in features"
                        :key="feature.body"
                    >
                        <li>
                            <div
                                class="rounded-2xl border border-zinc-200 bg-white px-4 py-3 text-sm text-zinc-700 shadow-sm dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-200"
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

                <!-- tiny “how to” -->
                <div class="mt-7 text-zinc-600 dark:text-zinc-400">
                    <div class="font-medium text-zinc-900 dark:text-zinc-200">Quick start:</div>
                    <ol class="mt-2 list-decimal space-y-1 pl-5 text-sm">
                        <li class="hover:text-sky-600 dark:hover:text-emerald-400">
                            Create a client
                        </li>
                        <li class="hover:text-sky-600 dark:hover:text-emerald-400">
                            Select them from the top-right menu
                        </li>
                        <li class="hover:text-sky-600 dark:hover:text-emerald-400">
                            Jump into invoice or item editor
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
        <section class="mt-10">
            <div
                class="rounded-2xl border border-zinc-200 bg-white p-4 shadow-lg sm:p-5 dark:border-zinc-800 dark:bg-zinc-900"
            >
                <div class="mb-3 flex items-center justify-between gap-3">
                    <h2 class="text-sm font-semibold text-zinc-900 dark:text-zinc-100">
                        Continue where you left off
                    </h2>
                    <span class="text-sm text-zinc-500 dark:text-zinc-400">
                        client → invoice → export → edit
                    </span>
                </div>

                <NoClients v-if="!clientStore.hasClients"></NoClients>

                <div
                    v-else-if="clientStore.hasClients && !clientStore.selectedClient"
                    class="space-y-3"
                >
                    <p class="text-sm text-zinc-600 dark:text-zinc-300">
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

                    <div class="text-xs text-zinc-500 dark:text-zinc-400">
                        Tip: create / edit clients in the Clients page.
                    </div>

                    <RouterLink to="/clients">
                        <TheButton class="cursor-pointer">clients</TheButton>
                    </RouterLink>
                </div>

                <div v-else>
                    <p class="text-sm text-zinc-600 dark:text-zinc-300">
                        Ready. Pick where you want to go:
                    </p>

                    <div class="mt-4 flex flex-wrap gap-3">
                        <RouterLink to="/clients">
                            <TheButton class="cursor-pointer">clients</TheButton>
                        </RouterLink>

                        <RouterLink to="/invoice">
                            <TheButton class="cursor-pointer">invoice</TheButton>
                        </RouterLink>

                        <RouterLink to="/editor">
                            <TheButton class="cursor-pointer">editor</TheButton>
                        </RouterLink>
                    </div>
                </div>
            </div>
        </section>
    </main>
</template>

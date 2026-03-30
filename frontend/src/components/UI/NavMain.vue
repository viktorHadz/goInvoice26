<script setup lang="ts">
import { computed } from 'vue'

const navLinks = computed(() => [
    { name: 'dashboard', to: '/app' },
    { name: 'clients', to: '/app/clients' },
    { name: 'invoice', to: '/app/invoice' },
    { name: 'editor', to: '/app/editor' },
])
</script>

<template>
    <nav
        class="pointer-events-none fixed bottom-3 left-1/2 z-50 -translate-x-1/2 sm:top-4 sm:left-3 sm:translate-x-0 md:top-3 md:left-4"
    >
        <div
            class="pointer-events-auto flex items-center gap-1 rounded-2xl border border-zinc-300 bg-white/90 p-2 shadow-lg backdrop-blur dark:border-zinc-800 dark:bg-zinc-900/90"
        >
            <RouterLink
                v-for="link in navLinks"
                :key="link.name"
                :to="link.to"
                v-slot="{ isActive, navigate }"
            >
                <a
                    @click="navigate"
                    :href="link.to"
                    class="relative block rounded-xl"
                >
                    <span
                        :class="[
                            'absolute inset-0 rounded-xl transition-all duration-200 ease-out',
                            isActive
                                ? 'scale-100 bg-sky-100 shadow-sm dark:bg-emerald-950/60'
                                : 'scale-75 bg-transparent',
                        ]"
                    />

                    <span
                        :class="[
                            'relative z-10 block rounded-xl px-3 py-1.5 text-sm font-medium transition-colors duration-200',
                            isActive
                                ? 'text-sky-700 dark:text-emerald-200'
                                : 'text-zinc-600 hover:text-zinc-900 dark:text-zinc-300 dark:hover:text-zinc-100',
                        ]"
                    >
                        {{ link.name }}
                    </span>
                </a>
            </RouterLink>
        </div>
    </nav>
</template>

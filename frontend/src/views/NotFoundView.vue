<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { ArrowLeftIcon, HomeIcon } from '@heroicons/vue/24/outline'
import TheButton from '@/components/UI/TheButton.vue'
import dunnoImg from '@/assets/images/vik_dunno.svg'

const route = useRoute()
const currentPath = computed(() => route.fullPath || '/')
</script>

<template>
    <main class="min-h-screen bg-zinc-50 text-zinc-900 dark:bg-zinc-950 dark:text-zinc-100">
        <section class="mx-auto flex min-h-screen w-full max-w-6xl items-center px-5 py-8 sm:px-8">
            <div
                class="grid w-full items-center gap-8 rounded-2xl border border-zinc-300 bg-white p-6 shadow-sm sm:p-8 lg:grid-cols-[1.05fr_0.95fr] dark:border-zinc-800 dark:bg-zinc-900"
            >
                <div>
                    <div
                        class="inline-flex items-center gap-2 rounded-2xl border border-amber-200 bg-amber-50 px-3 py-1 text-xs font-semibold tracking-[0.18em] text-amber-700 uppercase dark:border-amber-400/20 dark:bg-amber-950/30 dark:text-amber-200"
                    >
                        404
                        <span class="opacity-70">Page not found</span>
                    </div>

                    <h1
                        class="mt-5 text-4xl font-semibold tracking-tighter text-zinc-950 sm:text-5xl dark:text-zinc-50"
                    >
                        Looks like this page does not exist.
                    </h1>

                    <p
                        class="mt-4 max-w-xl text-base leading-7 text-zinc-600 sm:text-lg dark:text-zinc-400"
                    >
                        The route you requested is not available in this workspace. You can head
                        back to your app dashboard or return to the public landing page.
                    </p>

                    <p
                        class="mt-3 inline-flex max-w-full items-center rounded-2xl border border-zinc-300 bg-zinc-50 px-3 py-2 text-sm text-zinc-600 dark:border-zinc-800 dark:bg-zinc-950 dark:text-zinc-400"
                    >
                        <span class="truncate">Unknown path: {{ currentPath }}</span>
                    </p>

                    <div class="mt-6 flex flex-wrap gap-3">
                        <RouterLink to="/app">
                            <TheButton>
                                <HomeIcon class="size-4" />
                                Go to app
                            </TheButton>
                        </RouterLink>

                        <RouterLink to="/">
                            <TheButton variant="secondary">
                                <ArrowLeftIcon class="size-4" />
                                Back to landing
                            </TheButton>
                        </RouterLink>
                    </div>
                </div>

                <div class="mx-auto w-full max-w-md">
                    <div
                        class="rounded-2xl border border-zinc-300 bg-zinc-50 p-4 dark:border-zinc-800 dark:bg-zinc-950/60"
                    >
                        <img
                            :src="dunnoImg"
                            alt="Confused mascot illustration"
                            class="h-auto w-full"
                            draggable="false"
                        />
                    </div>
                </div>
            </div>
        </section>
    </main>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import TopRightMenu from './components/UI/TopRightMenu.vue'
import NavMain from './components/UI/NavMain.vue'
import TheToast from './components/UI/TheToast.vue'
import TheConfirmDialog from './components/UI/TheConfirmDialog.vue'

const route = useRoute()
const showAppChrome = computed(() => route.meta.appChrome === true)
</script>

<template>
    <div
        :class="[
            'min-h-screen w-full text-zinc-900 dark:text-zinc-100',
            showAppChrome ? 'flex bg-zinc-50 dark:bg-zinc-950' : 'bg-transparent',
        ]"
    >
        <main class="relative min-h-screen w-full">
            <div :class="showAppChrome ? 'mt-26 px-4 pb-16 sm:py-8 sm:pb-8 md:px-6' : ''">
                <RouterView v-slot="{ Component, route }">
                    <Transition
                        name="page"
                        mode="out-in"
                        appear
                    >
                        <component
                            :is="Component"
                            :key="route.fullPath"
                        />
                    </Transition>
                </RouterView>
            </div>

            <template v-if="showAppChrome">
                <NavMain />
                <TopRightMenu />
            </template>
            <TheToast />
            <TheConfirmDialog />
        </main>
    </div>
</template>

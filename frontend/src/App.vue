<script setup lang="ts">
import { RouterView } from 'vue-router'
import TopRightMenu from './components/UI/TopRightMenu.vue'
import NavMain from './components/UI/NavMain.vue'
</script>

<template>
    <div
        class="flex min-h-screen w-full bg-white text-zinc-900 dark:bg-zinc-950 dark:text-zinc-100"
    >
        <main class="relative min-h-screen w-full">
            <RouterView />
            <NavMain />
            <TopRightMenu />
        </main>
    </div>
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router'
import TopRightMenu from './components/UI/TopRightMenu.vue'
import NavMain from './components/UI/NavMain.vue'
</script>

<template>
  <div
    class="flex min-h-screen w-full bg-zinc-50 text-zinc-900 dark:bg-zinc-950 dark:text-zinc-100"
  >
    <main class="relative min-h-screen w-full">
      <div class="mt-26 px-4 py-4 sm:py-8 md:px-6">
        <RouterView />
      </div>
      <NavMain />
      <TopRightMenu />
    </main>
  </div>
</template>

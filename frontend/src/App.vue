<script setup lang="ts">
import { RouterView } from 'vue-router'
import TopRightMenu from './components/UI/TopRightMenu.vue'
import NavMain from './components/UI/NavMain.vue'
</script>

<template>
  <div
    class="text-fg bg-primary flex min-h-screen w-full transition duration-500 will-change-transform"
  >
    <aside>
      <NavMain></NavMain>
    </aside>
    <main class="main-content">
      <RouterView />
      <TopRightMenu />
    </main>
  </div>
</template>

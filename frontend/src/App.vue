<script setup lang="ts">
import { RouterView } from 'vue-router'
import TopRightMenu from './components/UI/TopRightMenu.vue'
import NavMain from './components/UI/NavMain.vue'
import TheToast from './components/UI/TheToast.vue'
import TheConfirmDialog from './components/UI/TheConfirmDialog.vue'
</script>

<template>
  <div
    class="flex min-h-screen w-full bg-zinc-50 text-zinc-900 dark:bg-zinc-950 dark:text-zinc-100"
  >
    <main class="relative min-h-screen w-full">
      <div class="mt-26 px-4 pb-16 sm:py-8 sm:pb-8 md:px-6">
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

      <NavMain />
      <!-- Shortcuts reside here -->
      <TopRightMenu />
      <TheToast />
      <TheConfirmDialog />
    </main>
  </div>
</template>

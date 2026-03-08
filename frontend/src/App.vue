<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { RouterView } from 'vue-router'
import TopRightMenu from './components/UI/TopRightMenu.vue'
import NavMain from './components/UI/NavMain.vue'
import TheToast, { type ToastNotice } from './components/UI/TheToast.vue'
import { onToastError } from './utils/toast'

type TheToastExpose = {
  onError: (error: ToastNotice) => void
  dismissToast: () => void
  clearErrors: () => void
}

const toastRef = ref<TheToastExpose | null>(null)

let stopToastErrors: (() => void) | null = null

onMounted(() => {
  stopToastErrors = onToastError((error) => {
    toastRef.value?.onError(error)
  })
})

onBeforeUnmount(() => {
  stopToastErrors?.()
  stopToastErrors = null
})
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

      <!-- Shortcuts reside here -->
      <TopRightMenu />
      <TheToast ref="toastRef" />
    </main>
  </div>
</template>

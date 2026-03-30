<script setup lang="ts">
import type { Component } from 'vue'
import { XMarkIcon } from '@heroicons/vue/24/outline'
import DecorGradient from './DecorGradient.vue'

withDefaults(
  defineProps<{
    open: boolean
    title: string
    description?: string
    panelClass?: string
    icon?: Component
  }>(),
  {
    description: '',
    panelClass: 'w-[92vw] max-w-225',
    icon: undefined,
  },
)

const emit = defineEmits<{
  (e: 'close'): void
}>()
</script>

<template>
  <Teleport to="body">
    <transition name="fade">
      <div
        v-if="open"
        class="fixed inset-0 z-100 bg-black/45 backdrop-blur-[1px]"
        @click="emit('close')"
      />
    </transition>

    <transition name="slide">
      <aside
        v-if="open"
        class="fixed top-0 right-0 z-101 flex h-screen flex-col border-l border-zinc-300 bg-white text-zinc-900 dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-100"
        :class="panelClass"
      >
        <header
          class="border-b border-zinc-300 bg-zinc-50 dark:border-zinc-800 dark:bg-zinc-900/70"
        >
          <div class="relative overflow-hidden px-4 py-3">
            <DecorGradient />
            <div class="relative z-10 flex items-center justify-between gap-4">
              <div class="flex min-w-0 items-center gap-3">
                <div
                  v-if="icon"
                  class="grid size-11 shrink-0 place-items-center rounded-2xl border border-zinc-300 bg-white shadow-sm dark:border-zinc-700 dark:bg-zinc-900"
                >
                  <component
                    :is="icon"
                    class="stroke-1.5 size-6 text-sky-700 dark:text-emerald-400"
                  />
                </div>
                <div class="min-w-0">
                  <div class="flex items-center gap-2">
                    <h2
                      class="text-xl font-semibold tracking-tight text-zinc-900 dark:text-zinc-100"
                    >
                      {{ title }}
                    </h2>
                    <slot name="title-extra" />
                  </div>
                  <p
                    v-if="description"
                    class="text-sm tracking-tight text-zinc-600 dark:text-zinc-300"
                  >
                    {{ description }}
                  </p>
                </div>
              </div>

              <slot name="header-actions">
                <button
                  type="button"
                  class="shrink-0 cursor-pointer rounded-lg p-2 text-zinc-600 hover:bg-rose-50 hover:text-rose-400 dark:text-zinc-300 dark:hover:bg-rose-400/15 dark:hover:text-rose-300"
                  @click="emit('close')"
                >
                  <XMarkIcon class="size-5" />
                </button>
              </slot>
            </div>
          </div>

          <slot name="subheader" />
        </header>

        <div class="min-h-0 flex-1">
          <slot />
        </div>
      </aside>
    </transition>
  </Teleport>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
.slide-enter-active,
.slide-leave-active {
  transition: transform 0.25s ease;
}
.slide-enter-from,
.slide-leave-to {
  transform: translateX(100%);
}
</style>

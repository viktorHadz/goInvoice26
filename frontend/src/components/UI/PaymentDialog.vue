<script setup lang="ts">
import { useEditorStore } from '@/stores/editor'
import { XMarkIcon } from '@heroicons/vue/24/outline'
import { onClickOutside } from '@vueuse/core'
import { useTemplateRef } from 'vue'
import TheButton from './TheButton.vue'

const editStore = useEditorStore()
const quickPayEl = useTemplateRef('quickPayEl')

onClickOutside(quickPayEl, (event) => editStore.setQuickPayOpen(false))
</script>
<template>
  <Teleport to="body">
    <div
      v-if="editStore.quickPayOpen"
      class="pointer-events-none fixed top-1/2 left-1/2 z-100 mx-auto mt-4 flex w-full max-w-md -translate-x-1/2 -translate-y-1/2 flex-col px-4"
    >
      <div
        ref="quickPayEl"
        class="pointer-events-auto relative rounded-xl border border-zinc-300 bg-zinc-50 p-4 text-zinc-700 shadow-sm dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-100"
      >
        <div class="flex justify-between">
          <div>
            <h2 class="text-sm font-semibold text-zinc-800 dark:text-zinc-100">
              Add quick payment
            </h2>
            <p class="mt-0.5 text-xs text-zinc-600 dark:text-zinc-400">
              Use when receiving payments from a client
            </p>
          </div>
          <button
            type="button"
            class="h-7 shrink-0 cursor-pointer rounded-lg p-1.5 text-zinc-600 hover:bg-rose-50 hover:text-rose-400 dark:text-zinc-300 dark:hover:bg-rose-400/15 dark:hover:text-rose-300"
            @click="editStore.setQuickPayOpen(false)"
          >
            <XMarkIcon class="size-4" />
          </button>
        </div>
        <div class="mt-6 flex shrink-0 gap-2 py-2">
          <input
            class="input input-accent w-full py-1!"
            placeholder="Payment"
          />
          <TheButton class="py-2!">Apply</TheButton>
        </div>
      </div>
    </div>
  </Teleport>
</template>

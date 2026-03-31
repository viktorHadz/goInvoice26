<script setup lang="ts">
import type { Component } from 'vue'

defineProps<{
  title: string
  icon: Component
}>()
</script>

<template>
  <article class="card-grad rounded-2xl border border-zinc-300 p-5 shadow-sm dark:border-zinc-800">
    <div class="flex items-start justify-between gap-3">
      <div>
        <h3 class="text-sm font-semibold tracking-tight text-zinc-950 dark:text-zinc-50">
          {{ title }}
        </h3>
        <div class="mt-1 text-xs leading-5 text-zinc-600 dark:text-zinc-400">
          <slot name="text" />
        </div>
      </div>

      <div
        class="rounded-xl border border-zinc-300 bg-zinc-50 p-2 text-zinc-700 dark:border-zinc-700 dark:bg-zinc-950/40 dark:text-zinc-200"
      >
        <component
          :is="icon"
          class="size-5"
        />
      </div>
    </div>

    <div class="mt-4">
      <slot />
    </div>
  </article>
</template>

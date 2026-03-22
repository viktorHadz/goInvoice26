<script setup lang="ts">
defineProps<{
  identityLabel: string
  subtitle?: string
}>()
</script>

<template>
  <div class="border-b border-zinc-200 px-3 py-3 sm:px-4 dark:border-zinc-800">
    <div class="flex items-start justify-between sm:items-center">
      <div class="flex justify-between">
        <div class="min-w-0">
          <h2 class="text-base font-semibold text-zinc-800 dark:text-zinc-100">
            {{ identityLabel }}
          </h2>
          <p
            v-if="subtitle"
            class="mt-0.5 text-xs text-zinc-600 dark:text-zinc-400"
          >
            {{ subtitle }}
          </p>
        </div>
      </div>
      <div class="flex flex-col items-end gap-2 sm:flex-row-reverse sm:gap-12">
        <slot name="more-menu"></slot>
        <slot name="actions" />
      </div>
    </div>
  </div>
</template>

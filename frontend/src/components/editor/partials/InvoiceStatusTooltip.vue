<script setup lang="ts">
import TheTooltip from '@/components/UI/TheTooltip.vue'
import { InformationCircleIcon } from '@heroicons/vue/24/outline'
const props = defineProps<{
  iconSize?: string
}>()
</script>

<template>
  <TheTooltip
    side="top"
    align="center"
    max-width-class="max-w-[340px]"
  >
    <template #content>
      <div class="space-y-2 text-left leading-relaxed">
        <div>
          <span class="font-semibold text-sky-600 dark:text-emerald-400">Draft:</span>
          Fully editable and deletable. Saving does not create new revisions.
        </div>

        <div>
          <span class="font-semibold text-sky-600 dark:text-emerald-400">Issued:</span>
          Editing creates new revisions. Adding payment from the table below creates receipts.
        </div>

        <div>
          <span class="font-semibold text-sky-600 dark:text-emerald-400">Paid:</span>
          Can be changed to issued unless due balance reaches 0. Then it auto locks and cannot be
          edited.
        </div>

        <div>
          <span class="font-semibold text-sky-600 dark:text-emerald-400">Void:</span>
          cancelled. Kept for record. No changes allowed.
        </div>
      </div>
    </template>

    <button
      type="button"
      class="inline-flex cursor-help text-zinc-700 transition hover:text-sky-600 dark:text-zinc-400 dark:hover:text-emerald-400"
      aria-label="Invoice status help"
    >
      <InformationCircleIcon
        class="cursor-help"
        :class="iconSize"
      />
    </button>
  </TheTooltip>
</template>

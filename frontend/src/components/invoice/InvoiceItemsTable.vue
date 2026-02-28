<script setup lang="ts">
import { computed } from 'vue'
import { useInvoiceDraftStore } from '@/stores/invoiceDraft'
import InvoiceLineRow from '@/components/invoice/InvoiceLineRow.vue'

const inv = useInvoiceDraftStore()
const lines = computed(() => inv.draft?.lines ?? [])
</script>

<template>
  <div class="overflow-x-auto">
    <div class="min-w-0">
      <div
        class="grid w-full grid-cols-[minmax(220px,1fr)_48px_64px_96px_110px_36px] items-center gap-2 py-2 pr-3 pl-2 text-sm font-semibold text-zinc-600 dark:text-zinc-200"
      >
        <div class="truncate">Product name</div>
        <div class="text-right">Qty</div>
        <div class="text-right">Mins</div>
        <div class="text-right">Unit</div>
        <div class="text-right">Total</div>
        <div class="text-right"></div>
      </div>

      <div class="h-px bg-zinc-200 dark:bg-zinc-800" />

      <div
        v-if="!lines.length"
        class="px-3 py-10 text-base text-zinc-500 dark:text-zinc-400"
      >
        No items yet. Add from the picker above.
      </div>

      <div
        v-else
        class="max-h-136 divide-y divide-zinc-200 overflow-y-auto dark:divide-zinc-800"
      >
        <InvoiceLineRow
          v-for="l in lines"
          :key="l.sortOrder"
          :line="l"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useInvoiceDraftStore } from '@/stores/invoiceDraft'
import InvoiceLineRow from '@/components/invoice/InvoiceLineRow.vue'

const inv = useInvoiceDraftStore()
const lines = computed(() => inv.draft?.lines ?? [])
</script>

<template>
  <div class="min-w-0">
    <div class="min-w-0">
      <div
        class="hidden grid-cols-10 items-center gap-3 py-2 pr-3 pl-2 text-sm font-semibold text-zinc-600 md:grid dark:text-zinc-200"
      >
        <div class="col-span-5">Product name</div>
        <div class="text-right">Qty</div>
        <div class="text-right">Minutes</div>
        <div class="text-right">Unit Price</div>
        <div class="text-right">Total</div>
        <div></div>
      </div>

      <div class="hidden h-px bg-zinc-200 md:block dark:bg-zinc-800" />

      <div
        v-if="!lines.length"
        class="px-3 py-10 text-base text-zinc-500 dark:text-zinc-400"
      >
        No items yet. Add from the picker above.
      </div>

      <div
        v-else
        class="max-h-[38rem] divide-y divide-zinc-200 overflow-y-auto dark:divide-zinc-800"
      >
        <InvoiceLineRow
          v-for="l in lines"
          :key="l.sortOrder"
          :line="l"
          class="items-start"
        />
      </div>
    </div>
  </div>
</template>

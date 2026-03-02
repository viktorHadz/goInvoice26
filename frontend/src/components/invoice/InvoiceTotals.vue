<script setup lang="ts">
import { useInvoiceStore } from '@/stores/invoice'
import { DocumentArrowDownIcon, DocumentIcon } from '@heroicons/vue/24/outline'
import TheButton from '../UI/TheButton.vue'

const inv = useInvoiceStore()
</script>

<template>
  <div
    v-if="!inv.invoice || !inv.totals"
    class="text-base text-zinc-500 dark:text-zinc-400"
  >
    No invoice loaded.
  </div>

  <div
    v-else
    class="min-w-0 space-y-3 text-base"
  >
    <div class="flex min-w-0 items-center justify-between gap-3">
      <div class="min-w-0 truncate text-zinc-600 dark:text-zinc-400">Subtotal</div>
      <div
        class="shrink-0 font-semibold whitespace-nowrap text-zinc-900 tabular-nums dark:text-zinc-100"
      >
        {{ inv.fmtGBPMinor(inv.totals.subtotalMinor) }}
      </div>
    </div>

    <div class="flex min-w-0 items-center justify-between gap-3">
      <div class="min-w-0 truncate text-zinc-600 dark:text-zinc-400">Discount</div>
      <div
        class="shrink-0 font-semibold whitespace-nowrap text-zinc-900 tabular-nums dark:text-zinc-100"
      >
        -{{ inv.fmtGBPMinor(inv.totals.discountMinor) }}
      </div>
    </div>

    <div class="flex min-w-0 items-center justify-between gap-3">
      <div class="min-w-0 truncate text-zinc-600 dark:text-zinc-400">VAT</div>
      <div
        class="shrink-0 font-semibold whitespace-nowrap text-zinc-900 tabular-nums dark:text-zinc-100"
      >
        {{ inv.fmtGBPMinor(inv.totals.vatMinor) }}
      </div>
    </div>

    <div class="h-px bg-zinc-200 dark:bg-zinc-800" />

    <div class="flex min-w-0 items-center justify-between gap-3 text-lg">
      <div class="min-w-0 truncate font-semibold text-zinc-800 dark:text-zinc-100">Total</div>
      <div
        class="shrink-0 font-semibold whitespace-nowrap text-zinc-900 tabular-nums dark:text-zinc-100"
      >
        {{ inv.fmtGBPMinor(inv.totals.totalMinor) }}
      </div>
    </div>

    <div
      class="mt-3 rounded-2xl border border-zinc-200 bg-zinc-50 p-4 dark:border-zinc-800 dark:bg-zinc-900/60"
    >
      <div class="flex min-w-0 items-center justify-between gap-3">
        <div class="min-w-0 truncate text-zinc-600 dark:text-zinc-400">Deposit</div>
        <div
          class="shrink-0 font-semibold whitespace-nowrap text-zinc-900 tabular-nums dark:text-zinc-100"
        >
          -{{ inv.fmtGBPMinor(inv.depositMinor) }}
        </div>
      </div>

      <div class="mt-2 flex min-w-0 items-center justify-between gap-3">
        <div class="min-w-0 truncate text-zinc-600 dark:text-zinc-400">Paid</div>
        <div
          class="shrink-0 font-semibold whitespace-nowrap text-zinc-900 tabular-nums dark:text-zinc-100"
        >
          -{{ inv.fmtGBPMinor(inv.invoice.paidMinor) }}
        </div>
      </div>

      <div class="mt-3 flex min-w-0 items-center justify-between gap-3 text-lg">
        <div class="min-w-0 truncate font-semibold text-zinc-800 dark:text-zinc-100">
          Balance due
        </div>
        <div
          class="shrink-0 font-semibold whitespace-nowrap text-zinc-900 tabular-nums dark:text-zinc-100"
        >
          {{ inv.fmtGBPMinor(inv.balanceDueMinor) }}
        </div>
      </div>
    </div>

    <div class="flex flex-col gap-y-2 sm:flex-row sm:gap-x-4">
      <TheButton
        class="flex w-full items-center gap-2"
        title="Generate PDF"
      >
        <DocumentArrowDownIcon class="size-4" />
        Print / PDF
      </TheButton>

      <TheButton
        class="flex w-full items-center gap-2"
        title="Generate Draft"
      >
        <DocumentIcon class="size-4" />
        Create Draft
      </TheButton>
    </div>
  </div>
</template>

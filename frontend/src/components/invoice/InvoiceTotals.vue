<script setup lang="ts">
import { useInvoiceStore } from '@/stores/invoice'
import { DocumentArrowDownIcon, DocumentIcon } from '@heroicons/vue/24/outline'
import TheButton from '../UI/TheButton.vue'
import { ref } from 'vue'
import { usePdfStore } from '@/stores/pdf'
import { fmtGBPMinor } from '@/utils/money'

const invStore = useInvoiceStore()
const pdfStore = usePdfStore()
const isCreatingDraft = ref(false)
const isGeneratingPdf = ref(false)

async function createDraft() {
  const inv = invStore.invoice
  if (!inv || isCreatingDraft.value) return

  isCreatingDraft.value = true
  try {
    const ok = await invStore.newDraftInvoice(inv)
    if (!ok) return

    await pdfStore.generateAndPersistPdf(inv.clientId, inv.baseNumber, 1)
  } finally {
    isCreatingDraft.value = false
  }
}

async function generatePdfOnly() {
  const inv = invStore.invoice
  if (!inv || isGeneratingPdf.value) return

  isGeneratingPdf.value = true
  try {
    await pdfStore.quickGeneratePDF(inv)
  } finally {
    isGeneratingPdf.value = false
  }
}
</script>

<template>
  <div
    v-if="!invStore.invoice || !invStore.totals"
    class="text-base text-zinc-500 dark:text-zinc-400"
  >
    No invoice loaded.
  </div>

  <div
    v-else
    class="min-w-0 space-y-4 text-sm"
  >
    <div class="grid min-w-0 grid-cols-[1fr_auto] items-center gap-3">
      <div class="min-w-0 truncate text-zinc-600 dark:text-zinc-400">Subtotal</div>
      <div
        class="shrink-0 font-semibold whitespace-nowrap text-zinc-900 tabular-nums dark:text-zinc-100"
      >
        {{ fmtGBPMinor(invStore.totals.subtotalMinor) }}
      </div>
    </div>

    <div class="grid min-w-0 grid-cols-[1fr_auto] items-center gap-3">
      <div class="min-w-0 truncate text-zinc-600 dark:text-zinc-400">Discount</div>
      <div
        class="shrink-0 font-semibold whitespace-nowrap text-zinc-900 tabular-nums dark:text-zinc-100"
      >
        -{{ fmtGBPMinor(invStore.totals.discountMinor) }}
      </div>
    </div>

    <div class="grid min-w-0 grid-cols-[1fr_auto] items-center gap-3">
      <div class="min-w-0 truncate text-zinc-600 dark:text-zinc-400">VAT</div>
      <div
        class="shrink-0 font-semibold whitespace-nowrap text-zinc-900 tabular-nums dark:text-zinc-100"
      >
        {{ fmtGBPMinor(invStore.totals.vatMinor) }}
      </div>
    </div>

    <div class="h-px bg-zinc-200 dark:bg-zinc-800" />

    <div class="grid min-w-0 grid-cols-[1fr_auto] items-center gap-3">
      <div class="min-w-0 truncate font-semibold text-zinc-800 dark:text-zinc-100">Total</div>
      <div
        class="shrink-0 font-semibold whitespace-nowrap text-zinc-900 tabular-nums dark:text-zinc-100"
      >
        {{ fmtGBPMinor(invStore.totals.totalMinor) }}
      </div>
    </div>

    <div class="grid min-w-0 grid-cols-[1fr_auto] items-center gap-3">
      <div class="min-w-0 truncate text-zinc-600 dark:text-zinc-400">Deposit</div>
      <div
        class="shrink-0 font-semibold whitespace-nowrap text-zinc-900 tabular-nums dark:text-zinc-100"
      >
        -{{ fmtGBPMinor(invStore.depositMinor) }}
      </div>
    </div>

    <div class="grid min-w-0 grid-cols-[1fr_auto] items-center gap-3">
      <div class="min-w-0 truncate text-zinc-600 dark:text-zinc-400">Paid</div>
      <div
        class="shrink-0 font-semibold whitespace-nowrap text-zinc-900 tabular-nums dark:text-zinc-100"
      >
        -{{ fmtGBPMinor(invStore.invoice.paidMinor) }}
      </div>
    </div>

    <div class="rounded-xl bg-zinc-50 px-3 py-3 dark:bg-zinc-900/40">
      <div class="grid min-w-0 grid-cols-[1fr_auto] items-center gap-3">
        <div class="min-w-0 truncate font-semibold text-zinc-800 dark:text-zinc-100">
          Balance due
        </div>
        <div
          class="shrink-0 font-semibold whitespace-nowrap text-zinc-900 tabular-nums dark:text-zinc-100"
        >
          {{ fmtGBPMinor(invStore.balanceDueMinor) }}
        </div>
      </div>
    </div>

    <div class="mt-8 flex flex-col gap-y-2 sm:flex-row sm:gap-x-4">
      <TheButton
        class="flex w-full cursor-pointer items-center gap-2"
        :disabled="isGeneratingPdf || isCreatingDraft"
        @click="generatePdfOnly"
      >
        <DocumentArrowDownIcon class="size-4" />
        Generate PDF
      </TheButton>

      <TheButton
        class="flex w-full cursor-pointer items-center gap-2"
        :disabled="isCreatingDraft || isGeneratingPdf"
        @click="createDraft"
      >
        <DocumentIcon class="size-4" />
        Create Draft
      </TheButton>
    </div>
  </div>
</template>

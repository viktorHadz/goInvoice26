<script setup lang="ts">
import { computed, ref } from 'vue'
import { DocumentArrowDownIcon, EnvelopeIcon, PencilSquareIcon } from '@heroicons/vue/24/outline'

import { useEditorStore } from '@/stores/editor'
import { useSettingsStore } from '@/stores/settings'
import { fmtGBPMinor, calcTotals, calcDepositMinor, calcBalanceDueMinor } from '@/utils/money'
import { fmtStrDate } from '@/utils/dates'
import {
  formatActiveEditorNodeLabel,
  formatInvoiceBaseLabel,
} from '@/utils/invoiceLabels'
import TheButton from '@/components/UI/TheButton.vue'
import { usePdfStore } from '@/stores/pdf'

const editStore = useEditorStore()
const setsStore = useSettingsStore()
const pdfStore = usePdfStore()

const isGeneratingPdf = ref(false)

async function generatePdfOnly() {
  const inv = editStore.activeInvoice
  if (!inv || isGeneratingPdf.value) return

  isGeneratingPdf.value = true
  try {
    await pdfStore.quickGeneratePDF(inv)
  } finally {
    isGeneratingPdf.value = false
  }
}

const inv = computed(() => editStore.activeInvoice)

const totals = computed(() => {
  if (!inv.value) return null
  return calcTotals(inv.value)
})

const depositMinor = computed(() => {
  if (!inv.value || !totals.value) return 0
  return calcDepositMinor(inv.value, totals.value.totalMinor)
})

const balanceDueMinor = computed(() => {
  if (!inv.value || !totals.value) return 0
  return calcBalanceDueMinor(totals.value.totalMinor, depositMinor.value, inv.value.paidMinor)
})

const invoicePrefix = computed(() => setsStore.settings?.invoicePrefix ?? '')

const invoiceDisplayLabel = computed(() => {
  const i = inv.value
  if (!i) return ''
  const node = editStore.activeNode
  if (node) return formatActiveEditorNodeLabel(invoicePrefix.value, node)
  return formatInvoiceBaseLabel(invoicePrefix.value, i.baseNumber)
})
</script>

<template>
  <div
    v-if="editStore.isLoadingInvoice"
    class="flex min-h-60 items-center justify-center"
  >
    <div class="text-sm text-zinc-500 dark:text-zinc-400">Loading invoice...</div>
  </div>

  <div
    v-else-if="!inv"
    class="flex min-h-60 items-center justify-center"
  >
    <div class="text-sm text-zinc-500 dark:text-zinc-400">No invoice data available.</div>
  </div>

  <div
    v-else
    class="space-y-4"
  >
    <!-- Header card -->
    <header
      class="overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
    >
      <div
        class="flex flex-col gap-3 border-b border-zinc-200 px-3 py-2.5 sm:flex-row sm:items-start sm:justify-between dark:border-zinc-800"
      >
        <div class="min-w-0">
          <div class="text-base font-semibold text-zinc-800 dark:text-zinc-100">
            Invoice details
          </div>
          <div class="text-xs text-sky-600 dark:text-emerald-400">
            Review invoice information before editing or exporting
          </div>
        </div>

        <div class="flex w-full flex-col gap-2 sm:w-auto sm:flex-row">
          <TheButton
            class="w-full sm:w-auto"
            type="button"
            @click="editStore.initEdit()"
          >
            <PencilSquareIcon class="size-4" />
            Edit invoice
          </TheButton>

          <TheButton
            class="w-full cursor-pointer sm:w-auto"
            type="button"
            @click="generatePdfOnly"
          >
            <DocumentArrowDownIcon class="size-4" />
            Print PDF
          </TheButton>
          <TheButton
            class="w-full cursor-pointer sm:w-auto"
            type="button"
          >
            <EnvelopeIcon class="size-4" />
            Send Email
          </TheButton>
        </div>
      </div>

      <div class="grid gap-4 p-3 md:p-4 lg:grid-cols-2 lg:items-start">
        <div class="min-w-0">
          <div class="mb-4 flex flex-wrap items-center gap-x-4 gap-y-1">
            <span class="text-base font-medium text-zinc-700 dark:text-zinc-300">
              Invoice number:
            </span>
            <span class="text-base font-bold text-sky-600 dark:text-emerald-400">
              {{ invoiceDisplayLabel }}
            </span>
          </div>
          <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <div class="mb-1 text-xs font-medium text-zinc-700 dark:text-zinc-300">
                Issue date
              </div>
              <div
                class="rounded-xl border border-zinc-200 bg-zinc-50 px-3 py-2 text-sm font-medium text-zinc-900 dark:border-zinc-800 dark:bg-zinc-900/40 dark:text-zinc-100"
              >
                {{ fmtStrDate(inv.issueDate) }}
              </div>
            </div>

            <div>
              <div class="mb-1 text-xs font-medium text-zinc-700 dark:text-zinc-300">Due by</div>
              <div
                class="rounded-xl border border-zinc-200 bg-zinc-50 px-3 py-2 text-sm font-medium text-zinc-900 dark:border-zinc-800 dark:bg-zinc-900/40 dark:text-zinc-100"
              >
                {{ inv.dueByDate ? fmtStrDate(inv.dueByDate) : 'N/A' }}
              </div>
            </div>
          </div>
        </div>

        <div
          class="min-w-0 rounded-2xl border border-zinc-200 bg-zinc-50/40 p-3 dark:border-zinc-800 dark:bg-zinc-900/40"
        >
          <div class="mb-2 flex items-center justify-between">
            <div class="font-semibold text-sky-600 dark:text-emerald-400">To</div>
            <div
              class="hidden rounded-full border border-zinc-200 bg-white/90 px-2 py-0.5 text-xs font-medium text-zinc-600 backdrop-blur-sm sm:inline-flex dark:border-zinc-700 dark:bg-zinc-900/70 dark:text-zinc-400"
            >
              client details
            </div>
          </div>

          <div class="space-y-2 text-sm">
            <div class="grid grid-cols-[84px_1fr] items-start gap-2">
              <div class="text-zinc-500 dark:text-zinc-400">Name</div>
              <div class="truncate font-medium text-zinc-900 dark:text-zinc-100">
                {{ inv.clientSnapshot.name || '—' }}
              </div>
            </div>

            <div class="grid grid-cols-[84px_1fr] items-start gap-2">
              <div class="text-zinc-500 dark:text-zinc-400">Company</div>
              <div class="truncate font-medium text-zinc-900 dark:text-zinc-100">
                {{ inv.clientSnapshot.companyName || '—' }}
              </div>
            </div>

            <div class="grid grid-cols-[84px_1fr] items-start gap-2">
              <div class="text-zinc-500 dark:text-zinc-400">Address</div>
              <div class="line-clamp-2 font-medium text-zinc-900 dark:text-zinc-100">
                {{ inv.clientSnapshot.address || '—' }}
              </div>
            </div>

            <div class="grid grid-cols-[84px_1fr] items-start gap-2">
              <div class="text-zinc-500 dark:text-zinc-400">Email</div>
              <div class="line-clamp-2 font-medium text-zinc-900 dark:text-zinc-100">
                {{ inv.clientSnapshot.email || '—' }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </header>

    <!-- Items card -->
    <section
      class="overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
    >
      <div
        class="flex items-start justify-between gap-3 border-b border-zinc-200 px-3 py-2.5 dark:border-zinc-800"
      >
        <div class="min-w-0">
          <div class="text-base font-semibold text-zinc-800 dark:text-zinc-100">Invoice items</div>
          <div class="text-xs text-sky-600 dark:text-emerald-400">
            Saved line items for this invoice
          </div>
        </div>

        <span
          class="hidden rounded-full border border-zinc-200 bg-white/90 px-2 py-0.5 text-xs font-medium text-zinc-600 backdrop-blur-sm sm:inline-flex dark:border-zinc-700 dark:bg-zinc-900/70 dark:text-zinc-400"
        >
          Read only
        </span>
      </div>

      <div class="p-2.5 md:p-3">
        <div class="overflow-x-auto">
          <div class="min-w-180">
            <div
              class="grid grid-cols-[minmax(240px,1fr)_88px_64px_110px_120px] items-center gap-3 px-2 py-2 text-sm font-semibold text-zinc-600 dark:text-zinc-200"
            >
              <div>Product name</div>
              <div>Type</div>
              <div class="text-right">Qty</div>
              <div class="text-right">Unit</div>
              <div class="text-right">Total</div>
            </div>

            <div class="h-px bg-zinc-200 dark:bg-zinc-800" />

            <div
              v-if="!inv.lines?.length"
              class="px-3 py-10 text-base text-zinc-500 dark:text-zinc-400"
            >
              No line items available.
            </div>

            <div
              v-else
              class="max-h-136 divide-y divide-zinc-200 overflow-y-auto dark:divide-zinc-800"
            >
              <div
                v-for="(line, i) in inv.lines"
                :key="i"
                class="grid grid-cols-[minmax(240px,1fr)_88px_64px_110px_120px] items-start gap-3 px-2 py-3 text-sm"
              >
                <div class="min-w-0">
                  <div class="truncate font-medium text-zinc-900 dark:text-zinc-100">
                    {{ line.name }}
                  </div>

                  <div class="mt-0.5 text-xs text-zinc-500 dark:text-zinc-400">
                    {{ line.pricingMode }}
                    <span v-if="line.pricingMode === 'hourly' && line.minutesWorked">
                      · {{ line.minutesWorked }} min
                    </span>
                  </div>
                </div>

                <div class="truncate text-zinc-600 capitalize dark:text-zinc-400">
                  {{ line.lineType }}
                </div>

                <div class="text-right text-zinc-700 tabular-nums dark:text-zinc-300">
                  {{ line.quantity }}
                </div>

                <div class="text-right text-zinc-700 tabular-nums dark:text-zinc-300">
                  {{ fmtGBPMinor(line.unitPriceMinor) }}
                </div>

                <div class="text-right font-medium text-zinc-900 tabular-nums dark:text-zinc-100">
                  {{ fmtGBPMinor(line.quantity * line.unitPriceMinor) }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Totals + Note -->
    <section class="grid gap-4 md:grid-cols-2">
      <section
        class="rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
      >
        <div class="border-b border-zinc-200 px-3 py-2.5 dark:border-zinc-800">
          <div class="text-base font-semibold text-zinc-800 dark:text-zinc-100">Totals</div>
          <div class="text-xs text-sky-600 dark:text-emerald-400">Balance overview</div>
        </div>

        <div
          v-if="totals"
          class="space-y-3 p-3 text-sm md:p-4"
        >
          <div class="grid grid-cols-[1fr_auto] items-center gap-3">
            <div class="text-zinc-600 dark:text-zinc-400">Subtotal</div>
            <div class="font-semibold text-zinc-900 tabular-nums dark:text-zinc-100">
              {{ fmtGBPMinor(totals.subtotalMinor) }}
            </div>
          </div>

          <div
            v-if="inv.discountType !== 'none'"
            class="grid grid-cols-[1fr_auto] items-center gap-3"
          >
            <div class="text-zinc-600 dark:text-zinc-400">Discount</div>
            <div class="font-semibold text-zinc-900 tabular-nums dark:text-zinc-100">
              -{{ fmtGBPMinor(totals.discountMinor) }}
            </div>
          </div>

          <div
            v-if="inv.vatRate > 0"
            class="grid grid-cols-[1fr_auto] items-center gap-3"
          >
            <div class="text-zinc-600 dark:text-zinc-400">VAT</div>
            <div class="font-semibold text-zinc-900 tabular-nums dark:text-zinc-100">
              {{ fmtGBPMinor(totals.vatMinor) }}
            </div>
          </div>

          <div class="h-px bg-zinc-200 dark:bg-zinc-800" />

          <div class="grid grid-cols-[1fr_auto] items-center gap-3">
            <div class="font-semibold text-zinc-800 dark:text-zinc-100">Total</div>
            <div class="font-semibold text-zinc-900 tabular-nums dark:text-zinc-100">
              {{ fmtGBPMinor(totals.totalMinor) }}
            </div>
          </div>

          <div
            v-if="inv.depositType !== 'none'"
            class="grid grid-cols-[1fr_auto] items-center gap-3"
          >
            <div class="text-zinc-600 dark:text-zinc-400">Deposit</div>
            <div class="font-semibold text-zinc-900 tabular-nums dark:text-zinc-100">
              -{{ fmtGBPMinor(depositMinor) }}
            </div>
          </div>

          <div
            v-if="inv.paidMinor > 0"
            class="grid grid-cols-[1fr_auto] items-center gap-3"
          >
            <div class="text-zinc-600 dark:text-zinc-400">Paid</div>
            <div class="font-semibold text-zinc-900 tabular-nums dark:text-zinc-100">
              -{{ fmtGBPMinor(inv.paidMinor) }}
            </div>
          </div>

          <div class="rounded-xl bg-zinc-50 px-3 py-3 dark:bg-zinc-900/40">
            <div class="grid grid-cols-[1fr_auto] items-center gap-3">
              <div class="font-semibold text-zinc-800 dark:text-zinc-100">Balance due</div>
              <div class="font-semibold text-zinc-900 tabular-nums dark:text-zinc-100">
                {{ fmtGBPMinor(balanceDueMinor) }}
              </div>
            </div>
          </div>
        </div>

        <div
          v-else
          class="p-3 text-sm text-zinc-500 dark:text-zinc-400"
        >
          No totals available.
        </div>
      </section>

      <section
        class="overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
      >
        <div class="border-b border-zinc-200 px-3 py-2.5 dark:border-zinc-800">
          <div class="text-base font-semibold text-zinc-800 dark:text-zinc-100">Note</div>
          <div class="text-xs text-sky-600 dark:text-emerald-400">
            Extra text shown on the invoice
          </div>
        </div>

        <div class="p-3 md:p-4">
          <div
            v-if="inv.note"
            class="rounded-xl border border-zinc-200 bg-zinc-50 p-3 text-sm text-zinc-700 italic dark:border-zinc-800 dark:bg-zinc-900/40 dark:text-zinc-300"
          >
            {{ inv.note }}
          </div>

          <div
            v-else
            class="rounded-xl border border-dashed border-zinc-200 bg-zinc-50/60 px-3 py-6 text-sm text-zinc-500 dark:border-zinc-800 dark:bg-zinc-900/30 dark:text-zinc-400"
          >
            No note added to this invoice.
          </div>
        </div>
      </section>
    </section>
  </div>
</template>

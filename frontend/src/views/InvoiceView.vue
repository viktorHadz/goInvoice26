<script setup lang="ts">
import { computed, watch } from 'vue'
import { useClientStore } from '@/stores/clients'
import { useInvoiceDraftStore } from '@/stores/invoiceDraft'
import InvoiceHeader from '@/components/invoice/InvoiceHeader.vue'
import InvoiceItemPicker from '@/components/invoice/InvoiceItemPicker.vue'
import InvoiceItemsTable from '@/components/invoice/InvoiceItemsTable.vue'
import InvoiceAdjustments from '@/components/invoice/InvoiceAdjustments.vue'
import InvoiceTotals from '@/components/invoice/InvoiceTotals.vue'

const clients = useClientStore()
const invStore = useInvoiceDraftStore()

const selected = computed(() => clients.selectedClient)

watch(
  selected,
  (c) => {
    if (!c?.id) return
    invStore.setDraft({
      clientId: c.id,
      issueDate: '',
      dueByDate: '',
      clientSnapshot: {
        name: c.name ?? '',
        companyName: c.companyName ?? '',
        address: c.address ?? '',
        email: c.email ?? '',
      },
      note: '',
      vatRate: 2000,
      discountType: 'none',
      discountValue: 0,
      lines: [],
      paidMinor: 0,
      depositMinor: 0,
    })
  },
  { immediate: true },
)
</script>

<template>
  <main class="mx-auto w-full max-w-4xl px-4">
    <!-- HEADER DOCUMENT -->
    <InvoiceHeader />

    <!-- PRODUCT PICKER -->
    <section
      class="mt-4 rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
    >
      <div class="border-b border-zinc-200 px-3 py-2.5 dark:border-zinc-800">
        <div class="text-base font-semibold text-zinc-800 dark:text-zinc-100">Insert products</div>
        <div class="text-xs text-sky-600 dark:text-emerald-400">
          Select an existing or insert a custom product
        </div>
      </div>

      <div class="p-3 md:p-4">
        <InvoiceItemPicker />
      </div>
    </section>

    <!-- INVOICE LINES  -->
    <section
      class="mt-4 overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
    >
      <div
        class="flex items-start justify-between gap-3 border-b border-zinc-200 px-3 py-2.5 dark:border-zinc-800"
      >
        <div class="min-w-0">
          <div class="text-base font-semibold text-zinc-800 dark:text-zinc-100">Invoice items</div>
          <div class="text-xs text-sky-600 dark:text-emerald-400">
            For #{{ invStore.draft?.baseNumber || '{invoice number}' }}
          </div>
        </div>

        <span
          class="hidden rounded-full border border-zinc-200 bg-white/90 px-2 py-0.5 text-xs font-medium text-zinc-600 backdrop-blur-sm sm:inline-flex dark:border-zinc-700 dark:bg-zinc-900/70 dark:text-zinc-400"
        >
          Tip: items can be modified
        </span>
      </div>

      <div class="p-2.5 md:p-3">
        <InvoiceItemsTable />
      </div>
    </section>

    <!-- FINANCIALS -->
    <section class="mt-4 grid gap-4 md:grid-cols-2">
      <!-- Adjustments -->
      <section
        class="overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
      >
        <div class="border-b border-zinc-200 px-3 py-2.5 dark:border-zinc-800">
          <div class="text-base font-semibold text-zinc-800 dark:text-zinc-100">Adjustments</div>
          <div class="text-xs text-sky-600 dark:text-emerald-400">
            Paid, deposit, discount, VAT and note
          </div>
        </div>

        <div class="p-3 md:p-4">
          <InvoiceAdjustments />
        </div>
      </section>

      <!-- Totals -->
      <section
        class="overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
      >
        <div class="border-b border-zinc-200 px-3 py-2.5 dark:border-zinc-800">
          <div class="text-base font-semibold text-zinc-800 dark:text-zinc-100">Totals</div>
          <div class="text-xs text-sky-600 dark:text-emerald-400">
            Subtotal, discount, VAT, total, balance
          </div>
        </div>

        <div class="p-3 md:p-4">
          <InvoiceTotals />
        </div>
      </section>
    </section>
  </main>
</template>

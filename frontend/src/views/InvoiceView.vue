<script setup lang="ts">
import { computed, watch } from 'vue'
import { useClientStore } from '@/stores/clients'
import { useInvoiceStore } from '@/stores/invoice'

import InvoiceHeader from '@/components/invoice/InvoiceHeader.vue'
import InvoiceItemPicker from '@/components/invoice/InvoiceItemPicker.vue'
import InvoiceItemsTable from '@/components/invoice/InvoiceItemsTable.vue'
import InvoiceAdjustments from '@/components/invoice/InvoiceAdjustments.vue'
import InvoiceTotals from '@/components/invoice/InvoiceTotals.vue'

import TheTooltip from '@/components/UI/TheTooltip.vue'
import { InformationCircleIcon } from '@heroicons/vue/24/outline'

const clients = useClientStore()
const invStore = useInvoiceStore()

const selected = computed(() => clients.selectedClient)

// Initialises or resets the invoice whenever the selected client changes
watch(
  selected,
  async (c, _prev, onCleanup) => {
    if (!c?.id) return

    // do not wipe current invoice if re-selecting same client
    if (invStore.invoice?.clientId === c.id) return

    let cancelled = false
    onCleanup(() => {
      cancelled = true
    })

    try {
      await invStore.initInvoiceFromServer({
        clientId: c.id,
        issueDate: '',
        dueByDate: undefined,

        clientSnapshot: {
          name: c.name ?? '',
          companyName: c.companyName ?? '',
          address: c.address ?? '',
          email: c.email ?? '',
        },

        note: '',

        vatRate: 2000,

        discountType: 'none',
        discountRate: 0,
        discountMinor: 0,

        lines: [],

        paidMinor: 0,

        depositType: 'none',
        depositRate: 0,
        depositMinor: 0,
      })

      if (cancelled) return
    } catch (err) {
      if (cancelled) return
      console.error('Failed to initialise invoice', err)
    }
  },
  { immediate: true },
)

const infoLines = [
  { id: 1, text: 'Amount paid - calculated after VAT' },
  { id: 2, text: 'Discount - calculated before VAT' },
  { id: 3, text: 'Deposit - calculated after VAT' },
  { id: 4, text: 'VAT Rate - set to 0% to exclude VAT' },
]
</script>

<template>
  <main class="mx-auto w-full max-w-4xl px-4 pb-16 sm:pb-0">
    <InvoiceHeader />

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

    <section
      class="mt-4 overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
    >
      <div
        class="flex items-start justify-between gap-3 border-b border-zinc-200 px-3 py-2.5 dark:border-zinc-800"
      >
        <div class="min-w-0">
          <div class="text-base font-semibold text-zinc-800 dark:text-zinc-100">Invoice items</div>
          <div class="text-xs text-sky-600 dark:text-emerald-400">
            {{ invStore.prettyBaseNumber ? 'For ' + invStore.prettyBaseNumber : '' }}
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

    <section class="mt-4 grid gap-4 md:grid-cols-2">
      <section
        class="overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
      >
        <div class="flex justify-between border-b border-zinc-200 px-3 py-2.5 dark:border-zinc-800">
          <div>
            <div class="text-base font-semibold text-zinc-800 dark:text-zinc-100">Adjustments</div>
            <div class="text-xs text-sky-600 dark:text-emerald-400">
              Paid, deposit, discount, VAT and note
            </div>
          </div>
          <TheTooltip
            :icon="InformationCircleIcon"
            :lines="infoLines"
            side="top"
          />
        </div>

        <div class="p-3 md:p-4">
          <InvoiceAdjustments />
        </div>
      </section>

      <section
        class="overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
      >
        <div class="flex justify-between border-b border-zinc-200 px-3 py-2.5 dark:border-zinc-800">
          <div>
            <div class="text-base font-semibold text-zinc-800 dark:text-zinc-100">Totals</div>
            <div class="text-xs text-sky-600 dark:text-emerald-400">Balance overview</div>
          </div>

          <TheTooltip
            :icon="InformationCircleIcon"
            text="Create a draft to save in invoice book. This lets you free edit invoice."
            side="top"
            align="center"
          />
        </div>

        <div class="p-3 md:p-4">
          <InvoiceTotals />
        </div>
      </section>
    </section>
  </main>
</template>

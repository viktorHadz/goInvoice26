<script setup lang="ts">
import { computed, watch } from 'vue'
import { useClientStore } from '@/stores/clients'
import { useInvoiceDraftStore } from '@/stores/invoiceDraft'
import InvoiceHeader from '@/components/invoice/InvoiceHeader.vue'
import InvoiceItemPicker from '@/components/invoice/InvoiceItemPicker.vue'
import InvoiceItemsTable from '@/components/invoice/InvoiceItemsTable.vue'
import InvoiceAdjustments from '@/components/invoice/InvoiceAdjustments.vue'
import InvoiceTotals from '@/components/invoice/InvoiceTotals.vue'
import InvoAddItems from '@/components/invoice/InvoAddItems.vue'

const clients = useClientStore()
const inv = useInvoiceDraftStore()

const selected = computed(() => clients.selectedClient)

watch(
  selected,
  (c) => {
    if (!c?.id) return
    inv.setDraft({
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
  <main class="mx-auto w-full 2xl:max-w-5xl">
    <section
      class="relative rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-900"
    >
      <div class="pointer-events-none absolute inset-0 overflow-hidden rounded-2xl">
        <div
          class="pointer-events-none absolute inset-0 bg-[radial-gradient(900px_circle_at_15%_0%,rgba(56,189,248,0.10),transparent_55%)] dark:bg-[radial-gradient(900px_circle_at_15%_0%,rgba(16,185,129,0.18),transparent_55%)]"
        />
        <div
          class="pointer-events-none absolute inset-0 bg-[linear-gradient(to_right,rgba(255,255,255,0.06)_1px,transparent_1px),linear-gradient(to_bottom,rgba(255,255,255,0.06)_1px,transparent_1px)] bg-size-[36px_36px] opacity-[0.55] dark:opacity-[0.35]"
        />
      </div>

      <div class="relative z-10 p-4 md:p-5">
        <InvoiceHeader />
      </div>
    </section>

    <section
      class="mt-6 grid min-w-0 grid-cols-1 gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(0,18rem)] xl:grid-cols-[minmax(0,1fr)_380px]"
    >
      <div class="min-w-0 space-y-6">
        <section
          class="min-w-0 rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-900"
        >
          <div class="border-b border-zinc-200 px-4 py-3.5 dark:border-zinc-800">
            <div class="text-lg font-semibold text-zinc-800 dark:text-zinc-100">
              Add invoice items
            </div>
            <div class="text-sm text-sky-600 dark:text-emerald-400">
              Search client products and add them as lines
            </div>
          </div>

          <div class="p-4">
            <InvoAddItems />
          </div>
        </section>

        <section
          class="min-w-0 rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-900"
        >
          <div class="border-b border-zinc-200 px-4 py-3.5 dark:border-zinc-800">
            <div class="text-lg font-semibold text-zinc-800 dark:text-zinc-100">Invoice items</div>
            <div class="text-sm text-sky-600 dark:text-emerald-400">Edit products inline</div>
          </div>

          <div class="min-w-0 p-2 md:p-3">
            <InvoiceItemsTable />
          </div>
        </section>
      </div>

      <div class="min-w-0 space-y-6">
        <section
          class="min-w-0 rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-900"
        >
          <div class="border-b border-zinc-200 px-4 py-3.5 dark:border-zinc-800">
            <div class="text-lg font-semibold text-zinc-800 dark:text-zinc-100">Adjustments</div>
            <div class="text-sm text-sky-600 dark:text-emerald-400">
              Paid, deposit, discount, VAT and note
            </div>
          </div>

          <div class="p-4">
            <InvoiceAdjustments />
          </div>
        </section>

        <section
          class="min-w-0 rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-900"
        >
          <div class="border-b border-zinc-200 px-4 py-3.5 dark:border-zinc-800">
            <div class="text-lg font-semibold text-zinc-800 dark:text-zinc-100">Totals</div>
            <div class="text-sm text-sky-600 dark:text-emerald-400">
              Subtotal, discount, VAT, total, balance
            </div>
          </div>

          <div class="p-4">
            <InvoiceTotals />
          </div>
        </section>
      </div>
    </section>
  </main>
</template>

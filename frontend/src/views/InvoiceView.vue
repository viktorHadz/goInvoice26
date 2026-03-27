<script setup lang="ts">
import { computed, onActivated, onMounted, watch } from 'vue'
import { useClientStore } from '@/stores/clients'
import { useInvoiceStore } from '@/stores/invoice'
import { handleActionError } from '@/utils/errors/handleActionError'

import InvoiceHeader from '@/components/invoice/InvoiceHeader.vue'
import InvoiceItemPicker from '@/components/invoice/InvoiceItemPicker.vue'
import InvoiceItemsTable from '@/components/invoice/InvoiceItemsTable.vue'
import InvoiceAdjustments from '@/components/invoice/InvoiceAdjustments.vue'
import InvoiceTotals from '@/components/invoice/InvoiceTotals.vue'

import TheTooltip from '@/components/UI/TheTooltip.vue'
import { InformationCircleIcon } from '@heroicons/vue/24/outline'
import InvoiceNote from '@/components/invoice/InvoiceNote.vue'

const clients = useClientStore()
const invStore = useInvoiceStore()

const selected = computed(() => clients.selectedClient)

function refreshInvoiceClientSnapshot() {
  const c = clients.selectedClient
  const inv = invStore.invoice

  if (!c?.id) return
  if (!inv) return
  if (inv.clientId !== c.id) return

  invStore.setClientSnapshot(c)
}

// Initialises or resets the invoice whenever the selected client changes
watch(
  selected,
  async (c, _prev, onCleanup) => {
    if (!c?.id) return

    if (invStore.invoice?.clientId === c.id) return

    let cancelled = false
    onCleanup(() => {
      cancelled = true
    })

    try {
      const template = invStore.buildFreshInvoiceTemplate(c)
      await invStore.initInvoiceFromServer(template)

      if (cancelled) return
    } catch (err) {
      if (cancelled) return
      handleActionError(err, {
        toastTitle: 'Could not start invoice',
        mapFields: false,
      })
    }
  },
  { immediate: true },
)
onMounted(refreshInvoiceClientSnapshot)
onActivated(refreshInvoiceClientSnapshot)
</script>

<template>
  <main class="mx-auto w-full max-w-4xl pb-16 sm:pb-0">
    <InvoiceHeader />

    <InvoiceItemPicker />

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
          class="hidden rounded-full border border-zinc-200 bg-zinc-50 px-2 py-0.5 text-xs font-medium text-zinc-600 backdrop-blur-sm sm:inline-flex dark:border-zinc-700 dark:bg-zinc-900/70 dark:text-zinc-400"
        >
          Tip: you can edit items anytime
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
        </div>

        <div class="p-3 md:p-4">
          <InvoiceAdjustments />
        </div>
      </section>

      <section
        class="rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
      >
        <div class="flex justify-between border-b border-zinc-200 px-3 py-2.5 dark:border-zinc-800">
          <div>
            <div class="text-base font-semibold text-zinc-800 dark:text-zinc-100">Totals</div>
            <div class="text-xs text-sky-600 dark:text-emerald-400">Balance overview</div>
          </div>

          <div class="flex flex-col items-end gap-y-2">
            <TheTooltip
              text="Save a draft to keep this invoice in your book and continue editing later."
              side="top"
              align="center"
              class="hover:text-sky-600 dark:hover:text-emerald-400"
            >
              <InformationCircleIcon class="size-5" />
            </TheTooltip>
          </div>
        </div>

        <div class="p-3 md:p-4">
          <InvoiceTotals />
        </div>
      </section>
    </section>
    <InvoiceNote />
  </main>
</template>

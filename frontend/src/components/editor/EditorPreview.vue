<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { BanknotesIcon, PencilSquareIcon, TrashIcon } from '@heroicons/vue/24/outline'

import { useEditorStore } from '@/stores/editor'
import { useSettingsStore } from '@/stores/settings'
import { fmtGBPMinor, calcTotals, calcDepositMinor, calcBalanceDueMinor } from '@/utils/money'
import { fmtStrDate } from '@/utils/dates'
import { formatActiveEditorNodeLabel, formatInvoiceBaseLabel } from '@/utils/invoiceLabels'
import TheDropdown from '@/components/UI/TheDropdown.vue'
import DetailsToolbar from '@/components/editor/partials/DetailsToolbar.vue'
import DetailsMenu, { type MenuOption } from '@/components/editor/partials/DetailsMenu.vue'
import { usePdfStore } from '@/stores/pdf'
import type { InvoiceStatus } from '@/components/invoice/invoiceTypes'
import { reachableStatuses } from '@/utils/invoiceStatusOptions'

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

const lifecycleStatus = computed(() => (inv.value?.status ?? 'draft') as InvoiceStatus)

const canStartEdit = computed(
  () => lifecycleStatus.value !== 'paid' && lifecycleStatus.value !== 'void',
)

const statusSelectOptions = computed(() => reachableStatuses(lifecycleStatus.value))

const selectedInvoiceStatus = computed({
  get(): InvoiceStatus {
    return lifecycleStatus.value
  },
  set(next: InvoiceStatus | null) {
    if (next == null || next === lifecycleStatus.value) return
    editStore.setInvoiceLifecycleStatus(next)
  },
})

const menuOpts = computed<MenuOption[]>(() => [
  {
    id: 1,
    name: 'Edit invoice',
    disabled: !canStartEdit.value,
    disabledReason: 'Cannot edit when status is "paid" or "void"',
    effect: () => editStore.initEdit(),
    icon: PencilSquareIcon,
  },
  {
    id: 2,
    name: 'Add payment',
    disabled: !canStartEdit.value,
    disabledReason: 'Payments cannot be added when status is "paid" or "void"',
    effect: () => editStore.setQuickPayOpen(true),
    icon: BanknotesIcon,
  },
  {
    id: 3,
    name: 'Delete invoice',
    disabled: !canStartEdit.value,
    disabledReason: 'Cannot delete when status is "paid" or "void"',
    effect: () => console.log('Deleting invoice'),
    icon: TrashIcon,
  },
])
</script>

<template>
  <Transition
    name="fade-down-up"
    mode="out-in"
  >
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
      <!-- Header -->
      <header
        class="overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
      >
        <DetailsToolbar
          :identity-label="invoiceDisplayLabel"
          subtitle="Read-only preview"
        >
          <template #more-menu>
            <DetailsMenu
              :pdf-disabled="isGeneratingPdf"
              @pdf="generatePdfOnly"
              :options="menuOpts"
            />
          </template>
        </DetailsToolbar>

        <div class="px-3 py-4 md:px-4 md:py-5">
          <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(320px,420px)] lg:items-start">
            <!-- Left: invoice details -->
            <section class="min-w-0">
              <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
                <div
                  class="min-w-0 rounded-2xl border border-zinc-200 bg-zinc-50/40 p-3 dark:border-zinc-800 dark:bg-zinc-900/40"
                >
                  <div class="text-xs font-medium tracking-wide text-zinc-600 dark:text-zinc-400">
                    Issue date
                  </div>
                  <div class="mt-1.5 text-sm font-medium text-zinc-900 dark:text-zinc-100">
                    {{ fmtStrDate(inv.issueDate) }}
                  </div>
                </div>

                <div
                  class="min-w-0 rounded-2xl border border-zinc-200 bg-zinc-50/40 p-3 dark:border-zinc-800 dark:bg-zinc-900/40"
                >
                  <div class="text-xs font-medium tracking-wide text-zinc-600 dark:text-zinc-400">
                    Due by
                  </div>
                  <div class="mt-1.5 text-sm font-medium text-zinc-900 dark:text-zinc-100">
                    {{ inv.dueByDate ? fmtStrDate(inv.dueByDate) : '—' }}
                  </div>
                </div>

                <div
                  class="min-w-0 rounded-2xl border border-zinc-200 bg-zinc-50/40 p-3 sm:col-span-2 xl:col-span-1 dark:border-zinc-800 dark:bg-zinc-900/40"
                >
                  <TheDropdown
                    v-model="selectedInvoiceStatus"
                    select-title="Status"
                    select-title-class="text-xs font-medium tracking-wide text-zinc-600 dark:text-zinc-400"
                    :options="statusSelectOptions"
                    input-class="mt-1.5 py-1.5 capitalize"
                    placeholder="Status"
                  />
                </div>
              </div>
            </section>

            <!-- Right: client card -->
            <section
              class="min-w-0 rounded-2xl border border-zinc-200 bg-zinc-50/40 p-3 dark:border-zinc-800 dark:bg-zinc-900/40"
            >
              <div class="mb-4 flex items-start justify-between gap-3">
                <div>
                  <div class="text-base font-semibold">To</div>
                </div>

                <span
                  class="inline-flex rounded-full border border-zinc-200 bg-zinc-50 px-2 py-0.5 text-xs font-medium text-zinc-600 dark:border-zinc-700 dark:bg-zinc-900/60 dark:text-zinc-400"
                >
                  Client details
                </span>
              </div>

              <div class="space-y-2 text-sm">
                <div class="grid grid-cols-[88px_minmax(0,1fr)] items-start gap-3">
                  <div class="text-zinc-500 dark:text-zinc-400">Name</div>
                  <div class="min-w-0 font-medium text-zinc-900 dark:text-zinc-100">
                    {{ inv.clientSnapshot.name || '—' }}
                  </div>
                </div>

                <div
                  v-if="inv.clientSnapshot.companyName"
                  class="grid grid-cols-[88px_minmax(0,1fr)] items-start gap-3"
                >
                  <div class="text-zinc-500 dark:text-zinc-400">Company</div>
                  <div class="min-w-0 font-medium text-zinc-900 dark:text-zinc-100">
                    {{ inv.clientSnapshot.companyName }}
                  </div>
                </div>

                <div
                  v-if="inv.clientSnapshot.address"
                  class="grid grid-cols-[88px_minmax(0,1fr)] items-start gap-3"
                >
                  <div class="text-zinc-500 dark:text-zinc-400">Address</div>
                  <div class="min-w-0 font-medium text-zinc-900 dark:text-zinc-100">
                    {{ inv.clientSnapshot.address }}
                  </div>
                </div>

                <div
                  v-if="inv.clientSnapshot.email"
                  class="grid grid-cols-[88px_minmax(0,1fr)] items-start gap-3"
                >
                  <div class="text-zinc-500 dark:text-zinc-400">Email</div>
                  <div class="min-w-0 font-medium wrap-break-word text-zinc-900 dark:text-zinc-100">
                    {{ inv.clientSnapshot.email }}
                  </div>
                </div>
              </div>
            </section>
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
            <div class="text-base font-semibold text-zinc-800 dark:text-zinc-100">
              Invoice items
            </div>
            <div class="text-xs text-zinc-600 dark:text-zinc-400">
              Saved line items for this invoice
            </div>
          </div>

          <span
            class="hidden rounded-full border border-zinc-200 bg-zinc-50 px-2 py-0.5 text-xs font-medium text-zinc-600 sm:inline-flex dark:border-zinc-700 dark:bg-zinc-900/50 dark:text-zinc-400"
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

                    <div class="mt-0.5 text-xs text-zinc-600 dark:text-zinc-400">
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

      <!-- Totals  -->
      <section class="grid gap-4 md:grid-cols-2">
        <section
          class="rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
        >
          <div class="border-b border-zinc-200 px-3 py-2.5 dark:border-zinc-800">
            <div class="text-base font-semibold text-zinc-800 dark:text-zinc-100">Totals</div>
            <div class="text-xs text-zinc-600 dark:text-zinc-400">Balance overview</div>
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
            <div class="text-xs text-zinc-600 dark:text-zinc-400">
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
  </Transition>
</template>

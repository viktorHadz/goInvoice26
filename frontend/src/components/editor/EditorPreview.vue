<script setup lang="ts">
import { computed, ref } from 'vue'
import { ChevronUpDownIcon, PencilSquareIcon, TrashIcon } from '@heroicons/vue/24/outline'

import { useEditorStore } from '@/stores/editor'
import { useSettingsStore } from '@/stores/settings'
import { fmtGBPMinor, calcTotals, calcDepositMinor, calcBalanceDueMinor } from '@/utils/money'
import { fmtStrDate } from '@/utils/dates'
import { editorPreviewLineTotalMinor, formatEditorPreviewLineMeta } from '@/utils/editorPreview'
import { formatActiveEditorNodeLabel, formatInvoiceBaseLabel } from '@/utils/invoiceLabels'
import TheDropdown from '@/components/UI/TheDropdown.vue'
import DetailsMenu, { type MenuOption } from '@/components/editor/partials/DetailsMenu.vue'
import { usePdfStore } from '@/stores/pdf'
import type { InvoiceStatus } from '@/components/invoice/invoiceTypes'
import {
  buildInvoiceStatusContext,
  canDeleteInvoice,
  canEditInvoice,
  reachableStatuses,
} from '@/utils/invoiceStatusOptions'
import { requestConfirmation } from '@/utils/confirm'
import InvoiceStatusTooltip from '@/components/editor/partials/InvoiceStatusTooltip.vue'

const editStore = useEditorStore()
const setsStore = useSettingsStore()
const pdfStore = usePdfStore()

const isGeneratingPdf = ref(false)

async function generatePdfOnly() {
  const inv = editStore.activeInvoice
  if (!inv || isGeneratingPdf.value) return

  const selectedRevisionNo =
    editStore.activeNode?.type === 'revision' ? editStore.activeNode.revisionNo : 1

  isGeneratingPdf.value = true
  try {
    await pdfStore.quickGeneratePDF(inv, selectedRevisionNo)
  } finally {
    isGeneratingPdf.value = false
  }
}

/** Draft when present (edit), else last saved invoice — keeps preview aligned with working copy. */
const inv = computed(() => editStore.draftInvoice ?? editStore.activeInvoice)

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
const dateFormat = computed(() => setsStore.settings?.dateFormat ?? 'dd/mm/yyyy')
const showItemTypeHeaders = computed(() => setsStore.settings?.showItemTypeHeaders !== false)

type LineGroup = {
  title: string
  lines: NonNullable<typeof inv.value>['lines']
}

const groupedLines = computed<LineGroup[]>(() => {
  const lines = inv.value?.lines ?? []
  const sorted = [...lines].sort((a, b) => a.sortOrder - b.sortOrder)
  const groups: LineGroup[] = [
    { title: 'Styles', lines: [] },
    { title: 'Samples', lines: [] },
    { title: 'Other Items', lines: [] },
  ]
  const styleGroup = groups[0]
  const sampleGroup = groups[1]
  const otherGroup = groups[2]
  if (!styleGroup || !sampleGroup || !otherGroup) return []

  for (const line of sorted) {
    if (line.lineType === 'style') {
      styleGroup.lines.push(line)
      continue
    }
    if (line.lineType === 'sample') {
      sampleGroup.lines.push(line)
      continue
    }
    otherGroup.lines.push(line)
  }

  return groups.filter((group) => group.lines.length > 0)
})

const invoiceDisplayLabel = computed(() => {
  const i = inv.value
  if (!i) return ''
  const node = editStore.activeNode
  if (node) return formatActiveEditorNodeLabel(invoicePrefix.value, node)
  return formatInvoiceBaseLabel(invoicePrefix.value, i.baseNumber)
})

const lifecycleStatus = computed(
  () => (editStore.activeInvoice?.status ?? inv.value?.status ?? 'draft') as InvoiceStatus,
)

const revisionCount = computed(() => {
  const baseNumber = editStore.activeInvoice?.baseNumber ?? inv.value?.baseNumber
  if (!baseNumber) return 1
  return editStore.invoiceBook.find((entry) => entry.baseNo === baseNumber)?.revisions.length ?? 1
})

const canStartEdit = computed(() => canEditInvoice(lifecycleStatus.value))
const canRemoveInvoice = computed(() => canDeleteInvoice(lifecycleStatus.value))

const statusOptions = computed(() =>
  reachableStatuses(
    lifecycleStatus.value,
    buildInvoiceStatusContext(editStore.activeInvoice, revisionCount.value),
  ),
)

const selectedStatus = computed({
  get(): InvoiceStatus {
    return lifecycleStatus.value
  },
  set(next: InvoiceStatus | null) {
    if (next == null || next === lifecycleStatus.value) return
    void editStore.requestInvoiceLifecycleStatusChange(next)
  },
})

async function confirmDeleteInvoice() {
  const invoice = editStore.activeInvoice
  if (!invoice) return

  const invoiceLabel = formatInvoiceBaseLabel(invoicePrefix.value, invoice.baseNumber)

  const confirmed = await requestConfirmation({
    title: 'Delete invoice?',
    message: `Delete ${invoiceLabel} from the invoice book?`,
    details:
      'This permanently removes the invoice, all saved revisions, and any recorded payments for it.',
    confirmLabel: 'Delete invoice',
    cancelLabel: 'Keep invoice',
    confirmVariant: 'danger',
  })

  if (!confirmed) return

  await editStore.deleteActiveInvoice()
}

const menuOpts = computed<MenuOption[]>(() => [
  {
    id: 1,
    name: 'Edit invoice',
    disabled: !canStartEdit.value,
    disabledReason: 'Only draft and issued invoices can be edited.',
    effect: () => editStore.initEdit(),
    icon: PencilSquareIcon,
  },
  {
    id: 2,
    name: 'Delete invoice',
    disabled: !canRemoveInvoice.value,
    disabledReason: 'Void invoices are final records and cannot be deleted.',
    effect: confirmDeleteInvoice,
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
        class="rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
      >
        <div class="border-b border-zinc-200 px-3 py-3 sm:px-4 dark:border-zinc-800">
          <div class="flex flex-wrap items-center justify-between gap-x-4 gap-y-2">
            <div class="min-w-0">
              <h2 class="text-base font-semibold text-zinc-800 dark:text-zinc-100">
                {{ invoiceDisplayLabel }}
              </h2>
              <p class="mt-0.5 text-xs text-sky-600 dark:text-emerald-400">Read-only preview</p>
            </div>
            <div class="flex items-center gap-2">
              <DetailsMenu
                :pdf-disabled="isGeneratingPdf"
                @pdf="generatePdfOnly"
                :options="menuOpts"
              />
            </div>
          </div>
        </div>

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
                    {{ fmtStrDate(inv.issueDate, dateFormat) }}
                  </div>
                </div>

                <div
                  class="min-w-0 rounded-2xl border border-zinc-200 bg-zinc-50/40 p-3 dark:border-zinc-800 dark:bg-zinc-900/40"
                >
                  <div class="text-xs font-medium tracking-wide text-zinc-600 dark:text-zinc-400">
                    Due by
                  </div>
                  <div class="mt-1.5 text-sm font-medium text-zinc-900 dark:text-zinc-100">
                    {{ inv.dueByDate ? fmtStrDate(inv.dueByDate, dateFormat) : '—' }}
                  </div>
                </div>

                <div
                  class="min-w-0 rounded-2xl border border-zinc-200 bg-zinc-50/40 p-3 sm:col-span-2 xl:col-span-1 dark:border-zinc-800 dark:bg-zinc-900/40"
                >
                  <div
                    class="mb-2 flex items-center justify-between gap-1 text-xs font-medium tracking-wide text-zinc-600 dark:text-zinc-400"
                  >
                    <span>Status</span>
                    <InvoiceStatusTooltip />
                  </div>
                  <TheDropdown
                    v-model="selectedStatus"
                    :right-icon="ChevronUpDownIcon"
                    :options="statusOptions"
                    input-class="py-1.5 capitalize"
                    placeholder="Status"
                    :disabled="statusOptions.length <= 1"
                  />
                </div>
              </div>
            </section>

            <!-- Right: client card -->
            <section
              class="min-w-0 rounded-2xl border border-zinc-200 bg-zinc-50/40 p-3 dark:border-zinc-800 dark:bg-zinc-900/40"
            >
              <div class="mb-2 flex items-start justify-between gap-3">
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
                  <div class="min-w-0 font-medium text-zinc-800 dark:text-zinc-100">
                    {{ inv.clientSnapshot.name || '—' }}
                  </div>
                </div>

                <div
                  v-if="inv.clientSnapshot.companyName"
                  class="grid grid-cols-[88px_minmax(0,1fr)] items-start gap-3"
                >
                  <div class="text-zinc-500 dark:text-zinc-400">Company</div>
                  <div class="min-w-0 font-medium text-zinc-800 dark:text-zinc-100">
                    {{ inv.clientSnapshot.companyName }}
                  </div>
                </div>

                <div
                  v-if="inv.clientSnapshot.address"
                  class="grid grid-cols-[88px_minmax(0,1fr)] items-start gap-3"
                >
                  <div class="text-zinc-500 dark:text-zinc-400">Address</div>
                  <div class="min-w-0 font-medium text-zinc-800 dark:text-zinc-100">
                    {{ inv.clientSnapshot.address }}
                  </div>
                </div>

                <div
                  v-if="inv.clientSnapshot.email"
                  class="grid grid-cols-[88px_minmax(0,1fr)] items-start gap-3"
                >
                  <div class="text-zinc-500 dark:text-zinc-400">Email</div>
                  <div class="min-w-0 font-medium wrap-break-word text-zinc-800 dark:text-zinc-100">
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
            <div class="text-xs text-sky-600 dark:text-emerald-400">
              Saved line items for the current invoice
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
                class="grid grid-cols-[minmax(240px,1fr)_64px_110px_120px] items-center gap-3 px-2 py-2 text-sm font-semibold text-zinc-600 dark:text-zinc-200"
              >
                <div>Product name</div>
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
                <template
                  v-for="group in groupedLines"
                  :key="group.title"
                >
                  <div
                    v-if="showItemTypeHeaders"
                    class="px-2 py-2 text-xs font-semibold tracking-wide text-sky-600 uppercase dark:text-emerald-300"
                  >
                    {{ group.title }}
                  </div>

                  <div
                    v-for="line in group.lines"
                    :key="line.sortOrder"
                    class="grid grid-cols-[minmax(240px,1fr)_64px_110px_120px] items-start gap-3 px-2 py-3 text-sm"
                  >
                    <div class="min-w-0">
                      <div class="truncate font-medium text-zinc-900 dark:text-zinc-100">
                        {{ line.name }}
                      </div>

                      <div class="mt-0.5 text-xs text-zinc-600 dark:text-zinc-400">
                        {{ formatEditorPreviewLineMeta(line) }}
                      </div>
                    </div>

                    <div class="text-right text-zinc-700 tabular-nums dark:text-zinc-300">
                      {{ line.quantity }}
                    </div>

                    <div class="text-right text-zinc-700 tabular-nums dark:text-zinc-300">
                      {{ fmtGBPMinor(line.unitPriceMinor) }}
                    </div>

                    <div
                      class="text-right font-medium text-zinc-900 tabular-nums dark:text-zinc-100"
                    >
                      {{ fmtGBPMinor(editorPreviewLineTotalMinor(line)) }}
                    </div>
                  </div>
                </template>
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
            <div class="text-xs text-sky-600 dark:text-emerald-400">Balance overview</div>
          </div>

          <div
            v-if="totals"
            class="space-y-3 p-3 text-sm md:p-4"
          >
            <div class="grid grid-cols-[1fr_auto] items-center gap-3">
              <div class="text-zinc-600 dark:text-zinc-400">Subtotal</div>
              <div class="font-semibold text-zinc-800 tabular-nums dark:text-zinc-100">
                {{ fmtGBPMinor(totals.subtotalMinor) }}
              </div>
            </div>

            <div
              v-if="inv.discountType !== 'none'"
              class="grid grid-cols-[1fr_auto] items-center gap-3"
            >
              <div class="text-zinc-600 dark:text-zinc-400">Discount</div>
              <div class="font-semibold text-zinc-800 tabular-nums dark:text-zinc-100">
                -{{ fmtGBPMinor(totals.discountMinor) }}
              </div>
            </div>

            <div
              v-if="inv.vatRate > 0"
              class="grid grid-cols-[1fr_auto] items-center gap-3"
            >
              <div class="text-zinc-600 dark:text-zinc-400">VAT</div>
              <div class="font-semibold text-zinc-800 tabular-nums dark:text-zinc-100">
                {{ fmtGBPMinor(totals.vatMinor) }}
              </div>
            </div>

            <div class="h-px bg-zinc-200 dark:bg-zinc-800" />

            <div class="grid grid-cols-[1fr_auto] items-center gap-3">
              <div class="font-semibold text-zinc-800 dark:text-zinc-100">Total</div>
              <div class="font-semibold text-zinc-800 tabular-nums dark:text-zinc-100">
                {{ fmtGBPMinor(totals.totalMinor) }}
              </div>
            </div>

            <div
              v-if="inv.depositType !== 'none'"
              class="grid grid-cols-[1fr_auto] items-center gap-3"
            >
              <div class="text-zinc-600 dark:text-zinc-400">Deposit</div>
              <div class="font-semibold text-zinc-800 tabular-nums dark:text-zinc-100">
                -{{ fmtGBPMinor(depositMinor) }}
              </div>
            </div>

            <div
              v-if="inv.paidMinor > 0"
              class="grid grid-cols-[1fr_auto] items-center gap-3"
            >
              <div class="text-zinc-600 dark:text-zinc-400">Paid</div>
              <div class="font-semibold text-zinc-800 tabular-nums dark:text-zinc-100">
                -{{ fmtGBPMinor(inv.paidMinor) }}
              </div>
            </div>

            <div class="rounded-xl bg-zinc-50 px-3 py-3 dark:bg-zinc-900/40">
              <div class="grid grid-cols-[1fr_auto] items-center gap-3">
                <div class="font-semibold text-zinc-800 dark:text-zinc-100">Balance due</div>
                <div class="font-semibold text-zinc-800 tabular-nums dark:text-zinc-100">
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
  </Transition>
</template>

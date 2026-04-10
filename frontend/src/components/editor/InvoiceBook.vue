<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import {
  ArrowDownIcon,
  ArrowPathIcon,
  ArrowUpIcon,
  BanknotesIcon,
  BookOpenIcon,
  CalendarDaysIcon,
  CheckIcon,
  ChevronRightIcon,
  DocumentCurrencyDollarIcon,
  DocumentDuplicateIcon,
  DocumentIcon,
  FunnelIcon,
  MagnifyingGlassIcon,
  UserCircleIcon,
} from '@heroicons/vue/24/outline'
import type { ActiveEditorNode, InvBookHistoryItem, InvBookInvoice } from './invBookTypes'
import DecorGradient from '@/components/UI/DecorGradient.vue'
import TheButton from '@/components/UI/TheButton.vue'
import TheTooltip from '../UI/TheTooltip.vue'
import { useEscape, useShortcuts, type ShortcutDefinition } from '@/composables/keyHandlers'
import { useSettingsStore } from '@/stores/settings'
import { useEditorStore } from '@/stores/editor'
import { useClientStore } from '@/stores/clients'
import { fmtStrDate } from '@/utils/dates'
import {
  formatInvoiceBaseLabel,
  formatInvoiceDisplayLabel,
  formatPaymentReceiptLabel,
} from '@/utils/invoiceLabels'
import { fmtGBPMinor } from '@/utils/money'
import {
  filterInvoiceBookByQuery,
  invoiceBookPaymentSummary,
  invoiceBookSortSummary,
  isDefaultInvoiceBookFilters,
} from './invoiceBookFilters'
import DetailsMenu, { type MenuOption } from './partials/DetailsMenu.vue'

function historyForBookSublist(history: InvBookHistoryItem[]): InvBookHistoryItem[] {
  return history
}

function bookSublistHistoryBadge(history: InvBookHistoryItem[]): string {
  const revisionCount = history.filter((entry) => entry.type === 'revision').length
  const receiptCount = history.filter((entry) => entry.type === 'payment_receipt').length
  const parts: string[] = []
  if (revisionCount > 0)
    parts.push(revisionCount === 1 ? '1 revision' : `${revisionCount} revisions`)
  if (receiptCount > 0) parts.push(receiptCount === 1 ? '1 receipt' : `${receiptCount} receipts`)
  if (parts.length === 0) return 'No history'
  return parts.join(' • ')
}

const props = defineProps<{
  activeNode: ActiveEditorNode
}>()

const emit = defineEmits<{
  select: [value: ActiveEditorNode]
}>()

const triggerEl = ref<HTMLElement | null>(null)
const panelEl = ref<HTMLElement | null>(null)

const isOpen = ref(false)
const query = ref('')
const openId = ref<number | null>(null)

const panelPos = ref({ top: 0, left: 0, width: 384 })

const setStore = useSettingsStore()
const clientStore = useClientStore()
const bookStore = useEditorStore()

const {
  invoiceBook,
  invoiceBookFilters,
  isLoadingBook,
  canGoPrev,
  canGoNext,
  offset,
  total,
  errorMessage,
} = storeToRefs(bookStore)

const invoices = computed<InvBookInvoice[]>(() => invoiceBook.value)

const bookInvoicePrefix = computed(() => setStore.settings?.invoicePrefix ?? '')
const dateFormat = computed(() => setStore.settings?.dateFormat ?? 'dd/mm/yyyy')
const activeClientLabel = computed(
  () => clientStore.selectedClient?.companyName || clientStore.selectedClient?.name || null,
)
const filteredInvoices = computed(() =>
  filterInvoiceBookByQuery(invoices.value, query.value, bookInvoicePrefix.value),
)

const pageLabel = computed(() => {
  if (!invoiceBook.value.length) return 'Showing 0 of 0'

  const start = offset.value + 1
  const end = Math.min(offset.value + invoiceBook.value.length, total.value)
  return `Showing ${start}-${end} of ${total.value}`
})

const filterSummary = computed(() => {
  const labels = [invoiceBookSortSummary(invoiceBookFilters.value)]

  if (invoiceBookFilters.value.paymentState !== 'all') {
    labels.push(invoiceBookPaymentSummary(invoiceBookFilters.value.paymentState))
  }

  if (invoiceBookFilters.value.activeClientOnly) {
    labels.push(
      activeClientLabel.value ? `Client: ${activeClientLabel.value}` : 'Client: active only',
    )
  }

  return labels
})

const filterMenuTooltip = computed(() =>
  isDefaultInvoiceBookFilters(invoiceBookFilters.value) ? 'Filter invoices' : 'Filters active',
)

function filterDirectionIcon(direction: 'asc' | 'desc') {
  return direction === 'desc' ? ArrowDownIcon : ArrowUpIcon
}

function invoiceStatusLabel(invoice: InvBookInvoice): string {
  const status = invoice.status?.trim()
  if (!status) return 'Unknown'
  return status.charAt(0).toUpperCase() + status.slice(1)
}

function invoiceStatusBadgeClass(invoice: InvBookInvoice): string {
  switch (invoice.status) {
    case 'paid':
      return 'border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-400/20 dark:bg-emerald-950/25 dark:text-emerald-200'
    case 'void':
      return 'border-zinc-300 bg-zinc-100 text-zinc-600 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-400'
    case 'draft':
      return 'border-amber-200 bg-amber-50 text-amber-700 dark:border-amber-400/20 dark:bg-amber-950/25 dark:text-amber-200'
    default:
      return 'border-sky-200 bg-sky-50 text-sky-700 dark:border-sky-400/20 dark:bg-sky-950/25 dark:text-sky-200'
  }
}

function invoiceClientLabel(invoice: InvBookInvoice): string {
  return invoice.clientCompanyName || invoice.clientName || `Client #${invoice.clientId}`
}

function invoiceBalanceLabel(invoice: InvBookInvoice): string {
  if (invoice.status === 'void') return 'Voided'
  if (invoice.balanceDueMinor <= 0) return 'Settled'
  return `${fmtGBPMinor(invoice.balanceDueMinor)} due`
}

function invoiceSummaryBits(invoice: InvBookInvoice): string[] {
  const bits = [invoiceClientLabel(invoice)]

  if (invoice.issueDate) {
    bits.push(fmtStrDate(invoice.issueDate, dateFormat.value))
  }

  bits.push(invoiceBalanceLabel(invoice))

  const historyBadge = bookSublistHistoryBadge(invoice.history)
  if (historyBadge !== 'No history') {
    bits.push(historyBadge)
  }

  return bits
}

// calculates the books position based on window dimensions
function placePanel() {
  if (!triggerEl.value) return

  const r = triggerEl.value.getBoundingClientRect()
  const width = Math.min(384, window.innerWidth - 16)

  const estimatedHeight = Math.min(640, window.innerHeight - 24)
  const preferredTop = r.bottom + 8
  const maxTop = window.innerHeight - estimatedHeight - 12

  panelPos.value = {
    width,
    top: Math.max(8, Math.min(preferredTop, maxTop)),
    left: Math.max(8, Math.min(r.right - width, window.innerWidth - width - 8)),
  }
}

async function openDropdown() {
  isOpen.value = true
  await bookStore.fetchInvoiceBook(true)
  if (props.activeNode?.type === 'revision' || props.activeNode?.type === 'paymentReceipt') {
    openId.value = props.activeNode.invoiceId
  }
  await nextTick()
  placePanel()
}

async function toggleDropdown() {
  if (isOpen.value) {
    closeDropdown()
    return
  }
  await openDropdown()
}

function closeDropdown() {
  isOpen.value = false
}

function toggleInvoice(id: number, hasRevisions: boolean) {
  if (!hasRevisions) return
  openId.value = openId.value === id ? null : id
}

function selectInvoice(invoice: InvBookInvoice) {
  emit('select', {
    type: 'invoice',
    clientId: invoice.clientId,
    id: invoice.id,
    baseNo: invoice.baseNo,
  })
  // closeDropdown() // might use at a later stage
}

function selectRevision(invoice: InvBookInvoice, revision: InvBookHistoryItem) {
  if (revision.revisionNo == null) return
  openId.value = invoice.id
  emit('select', {
    type: 'revision',
    clientId: invoice.clientId,
    id: revision.id,
    invoiceId: invoice.id,
    baseNo: invoice.baseNo,
    revisionNo: revision.revisionNo,
  })
  // closeDropdown() // might use at a later stage
}

function selectReceipt(invoice: InvBookInvoice, receipt: InvBookHistoryItem) {
  if (receipt.receiptNo == null) return
  openId.value = invoice.id
  emit('select', {
    type: 'paymentReceipt',
    clientId: invoice.clientId,
    id: receipt.id,
    invoiceId: invoice.id,
    baseNo: invoice.baseNo,
    receiptNo: receipt.receiptNo,
  })
}

function isActiveInvoice(invoice: InvBookInvoice) {
  return props.activeNode?.type === 'invoice' && props.activeNode.id === invoice.id
}

function isActiveRevision(rev: InvBookHistoryItem) {
  return props.activeNode?.type === 'revision' && props.activeNode.id === rev.id
}

function isActiveReceipt(receipt: InvBookHistoryItem) {
  return props.activeNode?.type === 'paymentReceipt' && props.activeNode.id === receipt.id
}

function rowClass(active: boolean) {
  return [
    'flex w-full items-center gap-3 rounded-xl border px-3 py-2 text-left transition',
    active
      ? 'border-sky-200 bg-sky-50 text-sky-700 dark:border-emerald-900/80 dark:bg-emerald-950/40 dark:text-zinc-100'
      : 'border-transparent text-zinc-700 hover:border-zinc-300 hover:bg-zinc-50 dark:text-zinc-300 dark:hover:border-zinc-700/50 dark:hover:bg-zinc-800/40',
  ]
}

function onClickOutside(e: MouseEvent) {
  const target = e.target as Node
  if (triggerEl.value?.contains(target)) return
  if (panelEl.value?.contains(target)) return
  closeDropdown()
}

function onWindowChange() {
  if (isOpen.value) placePanel()
}

async function handlePrevPage() {
  await bookStore.prevPage()
  openId.value = null
}

async function handleNextPage() {
  await bookStore.nextPage()
  openId.value = null
}

watch(
  () => props.activeNode,
  (node) => {
    if (node?.type === 'revision' || node?.type === 'paymentReceipt') {
      openId.value = node.invoiceId
      return
    }
    if (node?.type === 'invoice') {
      openId.value = null
    }
  },
)

watch(
  () => clientStore.selectedClient?.id,
  () => {
    openId.value = null
    query.value = ''
    closeDropdown()
  },
)

onMounted(() => {
  document.addEventListener('mousedown', onClickOutside)
  window.addEventListener('resize', onWindowChange)
  window.addEventListener('scroll', onWindowChange, true)
})

onBeforeUnmount(() => {
  document.removeEventListener('mousedown', onClickOutside)
  window.removeEventListener('resize', onWindowChange)
  window.removeEventListener('scroll', onWindowChange, true)
})

const shortcuts: ShortcutDefinition[] = [
  { key: 'b', modifiers: ['ctrl'], action: () => toggleDropdown() },
]

const menuOpts = computed<MenuOption[]>(() => [
  {
    id: 1,
    name:
      invoiceBookFilters.value.sortBy === 'date' && invoiceBookFilters.value.sortDirection === 'asc'
        ? 'Date: oldest first'
        : 'Date: newest first',
    effect: () => bookStore.cycleBookSort('date'),
    icon: CalendarDaysIcon,
    rightIcon: filterDirectionIcon(
      invoiceBookFilters.value.sortBy === 'date' ? invoiceBookFilters.value.sortDirection : 'desc',
    ),
    active: invoiceBookFilters.value.sortBy === 'date',
  },
  {
    id: 2,
    name:
      invoiceBookFilters.value.sortBy === 'balance' &&
      invoiceBookFilters.value.sortDirection === 'asc'
        ? 'Outstanding: low to high'
        : 'Outstanding: high to low',
    effect: () => bookStore.cycleBookSort('balance'),
    icon: DocumentCurrencyDollarIcon,
    rightIcon: filterDirectionIcon(
      invoiceBookFilters.value.sortBy === 'balance'
        ? invoiceBookFilters.value.sortDirection
        : 'desc',
    ),
    active: invoiceBookFilters.value.sortBy === 'balance',
  },
  {
    id: 3,
    name: invoiceBookPaymentSummary(invoiceBookFilters.value.paymentState),
    effect: () => bookStore.cycleBookPaymentState(),
    icon: BanknotesIcon,
    rightIcon: invoiceBookFilters.value.paymentState === 'all' ? undefined : CheckIcon,
    active: invoiceBookFilters.value.paymentState !== 'all',
  },
  {
    id: 4,
    name: invoiceBookFilters.value.activeClientOnly ? 'Client: active only' : 'Client: all clients',
    effect: () => bookStore.toggleBookActiveClientOnly(),
    icon: UserCircleIcon,
    rightIcon: invoiceBookFilters.value.activeClientOnly ? CheckIcon : undefined,
    active: invoiceBookFilters.value.activeClientOnly,
    disabled: !clientStore.selectedClient && !invoiceBookFilters.value.activeClientOnly,
    disabledReason: 'Select a client first to filter the invoice book to the active client.',
  },
  {
    id: 5,
    name: 'Reset filters',
    effect: () => bookStore.resetInvoiceBookFilters(),
    icon: ArrowPathIcon,
    disabled: isDefaultInvoiceBookFilters(invoiceBookFilters.value),
    disabledReason: 'Invoice book filters are already using the default view.',
  },
])

useShortcuts(shortcuts)
useEscape(() => closeDropdown())
</script>

<template>
  <div ref="triggerEl">
    <TheTooltip>
      <template #content>
        <span class="text-sky-600 dark:text-emerald-400">Shortcut:</span>
        <kbd>Ctrl</kbd>
        +
        <kbd>B</kbd>
      </template>

      <TheButton
        type="button"
        variant="primary"
        @click="toggleDropdown"
      >
        <BookOpenIcon class="size-5" />
        Open Invoice Book
      </TheButton>
    </TheTooltip>

    <Teleport to="body">
      <Transition name="invoice-book">
        <div
          v-if="isOpen"
          ref="panelEl"
          class="fixed z-100 overflow-hidden rounded-2xl border border-zinc-300 bg-white shadow-lg dark:border-zinc-800 dark:bg-zinc-900"
          :style="{
            top: `${panelPos.top}px`,
            left: `${panelPos.left}px`,
            width: `${panelPos.width}px`,
            maxHeight: '40rem',
          }"
        >
          <div
            class="relative border-b border-zinc-300 bg-white dark:border-zinc-800 dark:bg-zinc-950"
          >
            <DecorGradient />
            <div class="relative p-2 sm:p-4">
              <div class="flex items-center justify-between">
                <div class="min-w-0">
                  <div class="flex items-center gap-2">
                    <h3 class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">
                      Invoice Book
                    </h3>
                    <span
                      class="rounded-full border border-sky-200 bg-sky-50 px-2 py-0.5 text-xs font-medium text-sky-700 sm:inline-flex dark:border-emerald-400/20 dark:bg-emerald-950/25 dark:text-emerald-200"
                    >
                      {{ total }} invoices
                    </span>
                  </div>
                  <p class="mt-0.5 text-xs font-bold text-sky-600 dark:text-emerald-400">
                    Browse saved invoices, revisions, and receipts
                  </p>
                </div>
                <DetailsMenu
                  :menu-icon="FunnelIcon"
                  :options="menuOpts"
                  :tooltip-text="filterMenuTooltip"
                ></DetailsMenu>
              </div>

              <div class="relative mt-4">
                <MagnifyingGlassIcon
                  class="pointer-events-none absolute top-1/2 left-2 size-4 -translate-y-1/2 text-zinc-600 dark:text-zinc-400"
                />
                <input
                  id="invo-book-search"
                  v-model="query"
                  type="text"
                  placeholder="Search invoice, revision, or receipt…"
                  class="input input-accent py-1 pl-9 dark:bg-zinc-900"
                />
              </div>

              <div class="mt-3 flex flex-wrap items-center gap-2">
                <span
                  v-for="label in filterSummary"
                  :key="label"
                  class="rounded-full border border-zinc-300 bg-zinc-50 px-2 py-0.5 text-[11px] font-medium text-zinc-600 dark:border-zinc-700 dark:bg-zinc-900/70 dark:text-zinc-300"
                >
                  {{ label }}
                </span>
              </div>
            </div>
          </div>

          <div
            class="lg relative max-h-[min(26rem,calc(100vh-12rem))] overflow-y-auto p-3 pb-4 sm:max-h-100"
          >
            <Transition
              name="fade-down-up"
              mode="out-in"
            >
              <div
                :key="`${offset}-${query}-${invoiceBookFilters.sortBy}-${invoiceBookFilters.sortDirection}-${invoiceBookFilters.paymentState}-${invoiceBookFilters.activeClientOnly}`"
              >
                <div
                  v-if="isLoadingBook && !invoiceBook.length"
                  class="rounded-xl border border-dashed border-zinc-300 px-3 py-8 text-center text-sm text-zinc-600 dark:border-zinc-800 dark:text-zinc-400"
                >
                  Loading invoices...
                </div>

                <div
                  v-else-if="errorMessage && !invoiceBook.length"
                  class="rounded-xl border border-red-200 bg-red-50 px-3 py-8 text-center text-sm text-red-700 dark:border-red-900/60 dark:bg-red-950/30 dark:text-red-300"
                >
                  {{ errorMessage }}
                </div>

                <div
                  v-else-if="!filteredInvoices.length"
                  class="rounded-xl border border-dashed border-zinc-300 px-3 py-8 text-center text-sm text-zinc-600 dark:border-zinc-800 dark:text-zinc-400"
                >
                  No invoices found.
                </div>

                <ul
                  v-else
                  class="space-y-2"
                >
                  <li
                    v-for="invoice in filteredInvoices"
                    :key="invoice.id"
                  >
                    <div class="flex items-start gap-2">
                      <button
                        type="button"
                        class="mt-1 inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-lg text-zinc-600 transition hover:bg-zinc-50 hover:text-zinc-700 disabled:opacity-40 dark:text-zinc-400 dark:hover:bg-zinc-900 dark:hover:text-zinc-200"
                        :disabled="!historyForBookSublist(invoice.history).length"
                        @click="
                          toggleInvoice(invoice.id, !!historyForBookSublist(invoice.history).length)
                        "
                      >
                        <ChevronRightIcon
                          class="size-4 transition-transform"
                          :class="{
                            'rotate-90':
                              openId === invoice.id &&
                              !!historyForBookSublist(invoice.history).length,
                          }"
                        />
                      </button>

                      <div class="min-w-0 flex-1">
                        <button
                          type="button"
                          :class="rowClass(isActiveInvoice(invoice))"
                          @click="selectInvoice(invoice)"
                        >
                          <DocumentIcon
                            class="size-5 shrink-0"
                            :class="
                              isActiveInvoice(invoice)
                                ? 'text-sky-600 dark:text-emerald-400'
                                : 'text-zinc-400 dark:text-zinc-500'
                            "
                          />

                          <div class="flex min-w-0 flex-1 items-center justify-between gap-2">
                            <div class="min-w-0 flex-1">
                              <div class="truncate text-sm font-medium">
                                {{ formatInvoiceBaseLabel(bookInvoicePrefix, invoice.baseNo) }}
                              </div>
                              <div
                                class="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-zinc-600 dark:text-zinc-400"
                              >
                                <span
                                  v-for="bit in invoiceSummaryBits(invoice)"
                                  :key="`${invoice.id}-${bit}`"
                                >
                                  {{ bit }}
                                </span>
                              </div>
                            </div>

                            <TheTooltip>
                              <template #content>
                                <div class="flex flex-col text-start capitalize">
                                  <div>
                                    <span class="font-bold">Status:</span>
                                    {{ invoice.status || '' }}
                                  </div>
                                  <div>
                                    <span class="font-bold">Client:</span>
                                    {{ invoiceClientLabel(invoice) }}
                                  </div>
                                  <div>
                                    <span class="font-bold">Number:</span>
                                    {{ formatInvoiceBaseLabel(bookInvoicePrefix, invoice.baseNo) }}
                                  </div>
                                  <div>
                                    <span class="font-bold">Current revision:</span>
                                    {{
                                      formatInvoiceDisplayLabel(
                                        bookInvoicePrefix,
                                        invoice.baseNo,
                                        invoice.latestRevisionNo,
                                      )
                                    }}
                                  </div>
                                  <div>
                                    <span class="font-bold">Total:</span>
                                    {{ fmtGBPMinor(invoice.totalMinor) }}
                                  </div>
                                  <div>
                                    <span class="font-bold">Paid:</span>
                                    {{ fmtGBPMinor(invoice.paidMinor) }}
                                  </div>
                                  <div>
                                    <span class="font-bold">Outstanding:</span>
                                    {{ fmtGBPMinor(invoice.balanceDueMinor) }}
                                  </div>
                                </div>
                              </template>

                              <span
                                class="text-tiny shrink-0 rounded-full border px-2 py-0.5 font-medium"
                                :class="invoiceStatusBadgeClass(invoice)"
                              >
                                {{ invoiceStatusLabel(invoice) }}
                              </span>
                            </TheTooltip>
                          </div>
                        </button>

                        <div
                          v-if="
                            openId === invoice.id && historyForBookSublist(invoice.history).length
                          "
                          class="mt-2 space-y-2 pl-8"
                        >
                          <button
                            v-for="entry in historyForBookSublist(invoice.history)"
                            :key="`${entry.type}-${entry.id}`"
                            type="button"
                            :class="
                              rowClass(
                                entry.type === 'revision'
                                  ? isActiveRevision(entry)
                                  : isActiveReceipt(entry),
                              )
                            "
                            @click="
                              entry.type === 'revision'
                                ? selectRevision(invoice, entry)
                                : selectReceipt(invoice, entry)
                            "
                          >
                            <component
                              :is="
                                entry.type === 'revision' ? DocumentDuplicateIcon : BanknotesIcon
                              "
                              class="size-5 shrink-0"
                              :class="
                                entry.type === 'revision'
                                  ? isActiveRevision(entry)
                                    ? 'text-sky-600 dark:text-emerald-400'
                                    : 'text-zinc-400 dark:text-zinc-500'
                                  : isActiveReceipt(entry)
                                    ? 'text-sky-600 dark:text-emerald-400'
                                    : 'text-zinc-400 dark:text-zinc-500'
                              "
                            />

                            <div class="min-w-0 flex-1">
                              <div class="truncate text-sm font-medium">
                                {{
                                  entry.type === 'revision'
                                    ? formatInvoiceDisplayLabel(
                                        bookInvoicePrefix,
                                        invoice.baseNo,
                                        entry.revisionNo,
                                      )
                                    : formatPaymentReceiptLabel(
                                        bookInvoicePrefix,
                                        invoice.baseNo,
                                        entry.receiptNo,
                                      )
                                }}
                              </div>
                              <div
                                class="mt-1 flex flex-wrap gap-x-2 gap-y-1 text-xs text-zinc-600 dark:text-zinc-400"
                              >
                                <span v-if="entry.issueDate">
                                  {{ fmtStrDate(entry.issueDate, dateFormat) }}
                                </span>
                                <span v-else-if="entry.paymentDate">
                                  {{ fmtStrDate(entry.paymentDate, dateFormat) }}
                                </span>
                                <span v-if="entry.amountMinor != null">
                                  {{ fmtGBPMinor(entry.amountMinor) }}
                                </span>
                              </div>
                            </div>

                            <TheTooltip>
                              <template #content>
                                <div class="flex flex-col text-start capitalize">
                                  <div>
                                    <span class="font-bold">Status:</span>
                                    {{ invoice.status || '' }}
                                  </div>
                                  <div>
                                    <span class="font-bold">Number:</span>
                                    {{
                                      entry.type === 'revision'
                                        ? formatInvoiceDisplayLabel(
                                            bookInvoicePrefix,
                                            invoice.baseNo,
                                            entry.revisionNo,
                                          )
                                        : formatPaymentReceiptLabel(
                                            bookInvoicePrefix,
                                            invoice.baseNo,
                                            entry.receiptNo,
                                          )
                                    }}
                                  </div>
                                  <span v-if="entry.issueDate">
                                    <span class="font-bold">Issued at:</span>
                                    {{ fmtStrDate(entry.issueDate, dateFormat) }}
                                  </span>
                                  <span v-if="entry.dueByDate">
                                    <span class="font-bold">Due by:</span>
                                    {{ fmtStrDate(entry.dueByDate, dateFormat) }}
                                  </span>
                                  <span v-if="entry.paymentDate">
                                    <span class="font-bold">Payment date:</span>
                                    {{ fmtStrDate(entry.paymentDate, dateFormat) }}
                                  </span>
                                  <span v-if="entry.amountMinor != null">
                                    <span class="font-bold">Amount:</span>
                                    {{ fmtGBPMinor(entry.amountMinor) }}
                                  </span>
                                  <span v-if="entry.label">
                                    <span class="font-bold">Label:</span>
                                    {{ entry.label }}
                                  </span>
                                </div>
                              </template>

                              <span
                                class="text-tiny shrink-0 rounded-full border border-zinc-300 bg-white/90 px-2 py-0.5 font-medium text-zinc-600 dark:border-zinc-700 dark:bg-zinc-900/70 dark:text-zinc-400"
                              >
                                {{ entry.type === 'revision' ? 'Revision' : 'Receipt' }}
                              </span>
                            </TheTooltip>
                          </button>
                        </div>
                      </div>
                    </div>
                  </li>
                </ul>
              </div>
            </Transition>
          </div>

          <div
            class="flex items-center justify-between gap-2 border-t border-zinc-300 p-3 pb-[max(0.75rem,env(safe-area-inset-bottom))] sm:p-4 dark:border-zinc-800"
          >
            <p class="text-xs text-zinc-600 dark:text-zinc-400">
              {{ pageLabel }}
            </p>

            <div class="flex items-center gap-2">
              <TheButton
                type="button"
                variant="secondary"
                :disabled="!canGoPrev || isLoadingBook"
                @click="handlePrevPage"
              >
                Prev
              </TheButton>

              <TheButton
                type="button"
                variant="secondary"
                :disabled="!canGoNext || isLoadingBook"
                @click="handleNextPage"
              >
                Next
              </TheButton>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
<style scoped>
.invoice-book-enter-active,
.invoice-book-leave-active {
  transition:
    opacity 140ms ease,
    transform 140ms ease;
  transform-origin: top right;
}

.invoice-book-enter-from,
.invoice-book-leave-to {
  opacity: 0;
  transform: translateY(-6px) scale(0.985);
}

.invoice-book-enter-to,
.invoice-book-leave-from {
  opacity: 1;
  transform: translateY(0) scale(1);
}
</style>

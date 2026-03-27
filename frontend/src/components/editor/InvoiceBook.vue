<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import {
  BookOpenIcon,
  ChevronRightIcon,
  DocumentDuplicateIcon,
  DocumentIcon,
  MagnifyingGlassIcon,
} from '@heroicons/vue/24/outline'
import type { ActiveEditorNode, InvBookInvoice, InvBookRevision } from './invBookTypes'
import DecorGradient from '@/components/UI/DecorGradient.vue'
import TheButton from '@/components/UI/TheButton.vue'
import TheTooltip from '../UI/TheTooltip.vue'
import { useEscape, useShortcuts, type ShortcutDefinition } from '@/composables/keyHandlers'
import { useSettingsStore } from '@/stores/settings'
import { useEditorStore } from '@/stores/editor'
import { useClientStore } from '@/stores/clients'
import { fmtStrDate } from '@/utils/dates'
import { formatInvoiceBaseLabel, formatInvoiceDisplayLabel } from '@/utils/invoiceLabels'

/** Revision 1 is represented by the parent row; only show later revisions under it. */
function revisionsForBookSublist(revisions: InvBookRevision[]): InvBookRevision[] {
  return revisions.filter((r) => r.revisionNo > 1)
}

function bookSublistRevisionBadge(revisions: InvBookRevision[]): string {
  const n = revisionsForBookSublist(revisions).length
  if (n === 0) return 'No revisions'
  if (n === 1) return '1 revision'
  return `${n} revisions`
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

const { invoiceBook, isLoadingBook, canGoPrev, canGoNext, offset, total, errorMessage } =
  storeToRefs(bookStore)

const invoices = computed<InvBookInvoice[]>(() => invoiceBook.value)

const bookInvoicePrefix = computed(() => setStore.settings?.invoicePrefix ?? '')
const dateFormat = computed(() => setStore.settings?.dateFormat ?? 'dd/mm/yyyy')

const filteredInvoices = computed(() => {
  const q = query.value.trim().toLowerCase()
  if (!q) return invoices.value

  const prefix = bookInvoicePrefix.value

  return invoices.value
    .map((invoice) => {
      const invoiceLabel = formatInvoiceBaseLabel(prefix, invoice.baseNo).toLowerCase()

      const invoiceMatch = invoiceLabel.includes(q)

      // Match revision labels exactly as users see them in the sublist.
      const matchingRevisions = invoice.revisions.filter((rev) =>
        formatInvoiceDisplayLabel(prefix, invoice.baseNo, rev.revisionNo).toLowerCase().includes(q),
      )

      if (invoiceMatch) return invoice
      if (matchingRevisions.length) return { ...invoice, revisions: matchingRevisions }
      return null
    })
    .filter((x): x is InvBookInvoice => x !== null)
})

const pageLabel = computed(() => {
  if (!invoiceBook.value.length) return 'Showing 0 of 0'

  const start = offset.value + 1
  const end = Math.min(offset.value + invoiceBook.value.length, total.value)
  return `Showing ${start}-${end} of ${total.value}`
})

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
  if (props.activeNode?.type === 'revision') {
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
    id: invoice.id,
    baseNo: invoice.baseNo,
  })
  // closeDropdown() // might use at a later stage
}

function selectRevision(invoice: InvBookInvoice, revision: InvBookRevision) {
  openId.value = invoice.id
  emit('select', {
    type: 'revision',
    id: revision.id,
    invoiceId: invoice.id,
    baseNo: invoice.baseNo,
    revisionNo: revision.revisionNo,
  })
  // closeDropdown() // might use at a later stage
}

function isActiveInvoice(invoice: InvBookInvoice) {
  return props.activeNode?.type === 'invoice' && props.activeNode.id === invoice.id
}

function isActiveRevision(rev: InvBookRevision) {
  return props.activeNode?.type === 'revision' && props.activeNode.id === rev.id
}

function rowClass(active: boolean) {
  return [
    'flex w-full items-center gap-3 rounded-xl border px-3 py-2 text-left transition',
    active
      ? 'border-sky-200 bg-sky-50 text-sky-700 dark:border-emerald-900/80 dark:bg-emerald-950/40 dark:text-zinc-100'
      : 'border-transparent text-zinc-700 hover:border-zinc-200 hover:bg-zinc-50 dark:text-zinc-300 dark:hover:border-zinc-700/50 dark:hover:bg-zinc-800/40',
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
    if (node?.type === 'revision') {
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
          class="fixed z-100 overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-lg dark:border-zinc-800 dark:bg-zinc-900"
          :style="{
            top: `${panelPos.top}px`,
            left: `${panelPos.left}px`,
            width: `${panelPos.width}px`,
            maxHeight: '40rem',
          }"
        >
          <div
            class="relative border-b border-zinc-200 bg-white dark:border-zinc-800 dark:bg-zinc-950"
          >
            <DecorGradient variant="gradientAndGrid" />
            <div class="relative p-4">
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

                <p class="mt-0.5 text-xs text-zinc-600 dark:text-zinc-300">
                  Browse saved invoices and revisions
                </p>
              </div>

              <div class="relative mt-4">
                <MagnifyingGlassIcon
                  class="pointer-events-none absolute top-1/2 left-2 size-4 -translate-y-1/2 text-zinc-500 dark:text-zinc-400"
                />
                <input
                  id="invo-book-search"
                  v-model="query"
                  type="text"
                  placeholder="Search invoice…"
                  class="input input-accent py-1 pl-9"
                />
              </div>
            </div>
          </div>

          <div
            ref="listEl"
            class="lg relative max-h-[min(26rem,calc(100vh-12rem))] overflow-y-auto p-3 pb-4 sm:max-h-100"
          >
            <Transition
              name="fade-down-up"
              mode="out-in"
            >
              <div :key="`${offset}-${query}`">
                <div
                  v-if="isLoadingBook && !invoiceBook.length"
                  class="rounded-xl border border-dashed border-zinc-200 px-3 py-8 text-center text-sm text-zinc-500 dark:border-zinc-800 dark:text-zinc-400"
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
                  class="rounded-xl border border-dashed border-zinc-200 px-3 py-8 text-center text-sm text-zinc-500 dark:border-zinc-800 dark:text-zinc-400"
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
                        class="mt-1 inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-lg text-zinc-500 transition hover:bg-zinc-50 hover:text-zinc-700 disabled:opacity-40 dark:text-zinc-400 dark:hover:bg-zinc-900 dark:hover:text-zinc-200"
                        :disabled="!revisionsForBookSublist(invoice.revisions).length"
                        @click="
                          toggleInvoice(
                            invoice.id,
                            !!revisionsForBookSublist(invoice.revisions).length,
                          )
                        "
                      >
                        <ChevronRightIcon
                          class="size-4 transition-transform"
                          :class="{
                            'rotate-90':
                              openId === invoice.id &&
                              !!revisionsForBookSublist(invoice.revisions).length,
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
                            <div class="truncate text-sm font-medium">
                              {{ formatInvoiceBaseLabel(bookInvoicePrefix, invoice.baseNo) }}
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
                                    {{ formatInvoiceBaseLabel(bookInvoicePrefix, invoice.baseNo) }}
                                  </div>
                                </div>
                              </template>

                              <span
                                class="text-tiny shrink-0 rounded-full border px-2 py-0.5 font-medium"
                                :class="
                                  revisionsForBookSublist(invoice.revisions).length
                                    ? 'border-zinc-200 bg-white/90 text-zinc-600 dark:border-zinc-700 dark:bg-zinc-900/70 dark:text-zinc-400'
                                    : 'border-zinc-200 bg-zinc-50 text-zinc-500 dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-500'
                                "
                              >
                                {{ bookSublistRevisionBadge(invoice.revisions) }}
                              </span>
                            </TheTooltip>
                          </div>
                        </button>

                        <div
                          v-if="
                            openId === invoice.id &&
                            revisionsForBookSublist(invoice.revisions).length
                          "
                          class="mt-2 space-y-2 pl-8"
                        >
                          <button
                            v-for="rev in revisionsForBookSublist(invoice.revisions)"
                            :key="rev.id"
                            type="button"
                            :class="rowClass(isActiveRevision(rev))"
                            @click="selectRevision(invoice, rev)"
                          >
                            <DocumentDuplicateIcon
                              class="size-5 shrink-0"
                              :class="
                                isActiveRevision(rev)
                                  ? 'text-sky-600 dark:text-emerald-400'
                                  : 'text-zinc-400 dark:text-zinc-500'
                              "
                            />

                            <div class="min-w-0 flex-1 truncate text-sm font-medium">
                              {{
                                formatInvoiceDisplayLabel(
                                  bookInvoicePrefix,
                                  invoice.baseNo,
                                  rev.revisionNo,
                                )
                              }}
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
                                      formatInvoiceDisplayLabel(
                                        bookInvoicePrefix,
                                        invoice.baseNo,
                                        rev.revisionNo,
                                      )
                                    }}
                                  </div>
                                  <span>
                                    <span class="font-bold">Issued at:</span>
                                    {{
                                      rev.issueDate
                                        ? fmtStrDate(rev.issueDate, dateFormat)
                                        : 'not set'
                                    }}
                                  </span>
                                  <span>
                                    <span class="font-bold">Due by:</span>
                                    {{
                                      rev.dueByDate
                                        ? fmtStrDate(rev.dueByDate, dateFormat)
                                        : 'not set'
                                    }}
                                  </span>
                                </div>
                              </template>

                              <span
                                class="text-tiny shrink-0 rounded-full border border-zinc-200 bg-white/90 px-2 py-0.5 font-medium text-zinc-500 dark:border-zinc-700 dark:bg-zinc-900/70 dark:text-zinc-400"
                              >
                                Revision
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
            class="flex items-center justify-between gap-2 border-t border-zinc-200 px-4 py-3 pb-[max(0.75rem,env(safe-area-inset-bottom))] dark:border-zinc-800"
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

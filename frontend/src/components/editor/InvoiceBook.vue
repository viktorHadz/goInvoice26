<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import {
  BookOpenIcon,
  ChevronRightIcon,
  DocumentDuplicateIcon,
  DocumentIcon,
  MagnifyingGlassIcon,
} from '@heroicons/vue/24/outline'
import type { ActiveEditorNode, InvoiceRevisionNode, InvoiceTreeNode } from './editorTypes'
import DecorGradient from '@/components/UI/DecorGradient.vue'
import TheButton from '@/components/UI/TheButton.vue'
import TheTooltip from '../UI/TheTooltip.vue'
import { useEscape, useShortcuts, type ShortcutDefinition } from '@/composables/keyHandlers'
import { useSettingsStore } from '@/stores/settings'

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
const invoices = ref<InvoiceTreeNode[]>([
  {
    id: 2,
    baseNo: 2,
    status: 'draft',
    revisions: [{ id: 2, revisionNo: 1, issueDate: '2026-03-22' }],
  },
  {
    id: 1,
    baseNo: 1,
    status: 'draft',
    revisions: [{ id: 1, revisionNo: 1, issueDate: '2026-03-15' }],
  },
])
// const invoices = ref<InvoiceTreeNode[]>([
//   {
//     id: 1,
//     baseNumber: 1,
//     revisions: [
//       { id: 101, revisionNo: 2 },
//       { id: 102, revisionNo: 3 },
//       { id: 103, revisionNo: 4 },
//     ],
//   },
//   {
//     id: 2,
//     baseNumber: 2,
//     revisions: [
//       { id: 201, revisionNo: 2 },
//       { id: 202, revisionNo: 3 },
//     ],
//   },
//   {
//     id: 3,
//     baseNumber: 3,
//     revisions: [
//       { id: 301, revisionNo: 2 },
//       { id: 302, revisionNo: 3 },
//     ],
//   },
//   { id: 4, baseNumber: 4, revisions: [] },
//   { id: 5, baseNumber: 5, revisions: [] },
//   { id: 6, baseNumber: 6, revisions: [] },
//   { id: 7, baseNumber: 7, revisions: [] },
//   { id: 8, baseNumber: 8, revisions: [] },
//   { id: 9, baseNumber: 9, revisions: [] },
//   { id: 10, baseNumber: 10, revisions: [] },
//   { id: 11, baseNumber: 11, revisions: [] },
//   { id: 12, baseNumber: 12, revisions: [] },
//   { id: 13, baseNumber: 13, revisions: [] },
// ])

const setStore = useSettingsStore()
const userPrefix = setStore.settings?.invoicePrefix

const filteredInvoices = computed(() => {
  const q = query.value.trim().toLowerCase()
  if (!q) return invoices.value

  return invoices.value
    .map((invoice) => {
      const invoiceMatch = `${userPrefix}-${invoice.baseNo}`.toLowerCase().includes(q)
      const revisions = invoice.revisions.filter((rev) =>
        `${userPrefix}-${invoice.baseNo}.${rev.revisionNo}`.toLowerCase().includes(q),
      )

      if (invoiceMatch) return invoice
      if (revisions.length) return { ...invoice, revisions }
      return null
    })
    .filter((x): x is InvoiceTreeNode => x !== null)
})

const invoiceCountLabel = computed(() => {
  const n = filteredInvoices.value.length
  return n === 1 ? '1 invoice' : `${n} invoices`
})

function placePanel() {
  if (!triggerEl.value) return

  const r = triggerEl.value.getBoundingClientRect()
  const width = Math.min(384, window.innerWidth - 16)
  panelPos.value = {
    width,
    top: r.bottom + 8,
    left: Math.max(8, r.right - width),
  }
}

async function toggleDropdown() {
  isOpen.value = !isOpen.value
  if (!isOpen.value) return
  await nextTick()
  placePanel()
}

function closeDropdown() {
  isOpen.value = false
}

function toggleInvoice(id: number, hasRevisions: boolean) {
  if (!hasRevisions) return
  openId.value = openId.value === id ? null : id
}

function selectInvoice(invoice: InvoiceTreeNode) {
  emit('select', { type: 'invoice', id: invoice.id })
  closeDropdown()
}

function selectRevision(invoice: InvoiceTreeNode, revision: InvoiceRevisionNode) {
  openId.value = invoice.id
  emit('select', { type: 'revision', id: revision.id })
  closeDropdown()
}

function isActiveInvoice(invoice: InvoiceTreeNode) {
  return props.activeNode?.type === 'invoice' && props.activeNode.id === invoice.id
}

function isActiveRevision(rev: InvoiceRevisionNode) {
  return props.activeNode?.type === 'revision' && props.activeNode.id === rev.id
}

function rowClass(active: boolean) {
  return [
    'flex w-full items-center gap-3 rounded-xl border px-3 py-2 text-left transition',
    active
      ? 'border-sky-200 bg-sky-50/80 text-zinc-900 shadow-sm dark:border-emerald-900/80 dark:bg-emerald-950/40 dark:text-zinc-100'
      : 'border-transparent text-zinc-700 hover:border-zinc-200 hover:bg-zinc-50 dark:text-zinc-300 dark:hover:border-zinc-800 dark:hover:bg-zinc-900/60',
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
      <div
        v-if="isOpen"
        ref="panelEl"
        class="fixed z-100 overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-lg dark:border-zinc-800 dark:bg-zinc-900"
        :style="{
          top: `${panelPos.top}px`,
          left: `${panelPos.left}px`,
          width: `${panelPos.width}px`,
        }"
      >
        <div
          class="relative border-b border-zinc-200 bg-white dark:border-zinc-800 dark:bg-zinc-950"
        >
          <DecorGradient />
          <div class="relative p-4">
            <div class="min-w-0">
              <div class="flex items-center gap-2">
                <h3 class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">Invoice Book</h3>

                <span
                  class="hidden rounded-full border border-sky-200 bg-sky-50 px-2 py-0.5 text-[11px] font-medium text-sky-700 sm:inline-flex dark:border-emerald-400/20 dark:bg-emerald-950/25 dark:text-emerald-200"
                >
                  {{ invoiceCountLabel }}
                </span>
              </div>

              <p class="mt-0.5 text-xs text-zinc-500 dark:text-zinc-300">
                Browse saved invoices and revisions
              </p>
            </div>

            <div class="relative mt-4">
              <MagnifyingGlassIcon
                class="pointer-events-none absolute top-1/2 left-2 size-5 -translate-y-1/2 text-zinc-500 dark:text-zinc-400"
              />
              <input
                v-model="query"
                type="text"
                id="invo-book-search"
                placeholder="Search invoice…"
                class="input input-accent pl-9"
              />
            </div>
          </div>
        </div>

        <div class="max-h-80 overflow-y-auto p-3 sm:max-h-104">
          <div
            v-if="!filteredInvoices.length"
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
                  :disabled="!invoice.revisions.length"
                  @click="toggleInvoice(invoice.id, !!invoice.revisions.length)"
                >
                  <ChevronRightIcon
                    class="size-4 transition-transform"
                    :class="{ 'rotate-90': openId === invoice.id && invoice.revisions.length }"
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
                        {{ userPrefix }}-{{ invoice.baseNo }}
                      </div>

                      <span
                        class="shrink-0 rounded-full border px-2 py-0.5 text-[11px] font-medium"
                        :class="
                          invoice.revisions.length
                            ? 'border-zinc-200 bg-white/90 text-zinc-600 dark:border-zinc-700 dark:bg-zinc-900/70 dark:text-zinc-400'
                            : 'border-zinc-200 bg-zinc-50 text-zinc-500 dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-500'
                        "
                      >
                        {{
                          invoice.revisions.length
                            ? `${invoice.revisions.length} revisions`
                            : 'No revisions'
                        }}
                      </span>
                    </div>
                  </button>

                  <div
                    v-if="openId === invoice.id && invoice.revisions.length"
                    class="mt-2 space-y-2 pl-8"
                  >
                    <button
                      v-for="rev in invoice.revisions"
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
                        {{ userPrefix }}-{{ invoice.baseNo }}.{{ rev.revisionNo }}
                      </div>

                      <span
                        class="shrink-0 rounded-full border border-zinc-200 bg-white/90 px-2 py-0.5 text-[11px] font-medium text-zinc-500 dark:border-zinc-700 dark:bg-zinc-900/70 dark:text-zinc-400"
                      >
                        Revision
                      </span>
                    </button>
                  </div>
                </div>
              </div>
            </li>
          </ul>
        </div>
      </div>
    </Teleport>
  </div>
</template>

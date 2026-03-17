<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import InvoiceBook from '@/components/editor/InvoiceBook.vue'
import type { ActiveEditorNode } from '@/components/editor/invBookTypes'
import DecorGradient from '@/components/UI/DecorGradient.vue'
import { DocumentIcon, PencilSquareIcon } from '@heroicons/vue/24/outline'
import { useEditorStore } from '@/stores/editor'
import { useSettingsStore } from '@/stores/settings'

const setsStore = useSettingsStore()
const activeNode = ref<ActiveEditorNode>(null)
const editStore = useEditorStore()

const selectionLabel = computed(() => {
  const node = activeNode.value
  if (!node) return 'Nothing selected'

  if (node.type === 'invoice') {
    return `${setsStore.settings?.invoicePrefix}-${node.baseNo}`
  }

  return `${setsStore.settings?.invoicePrefix}-${node.baseNo}.${node.revisionNo}`
})
onMounted(() => {
  editStore.fetchInvoiceBook()
  editStore.fetchInvoice(1, 1)
})
</script>

<template>
  <main class="mx-auto w-full max-w-4xl 2xl:max-w-5xl">
    <!-- Header -->
    <section class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div class="flex items-center gap-2">
        <div
          class="grid size-12 shrink-0 place-items-center rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-700 dark:bg-zinc-900"
        >
          <PencilSquareIcon class="stroke-1.5 size-7 text-sky-600 dark:text-emerald-400" />
        </div>

        <div>
          <h2 class="text-2xl font-semibold tracking-tight text-zinc-900 dark:text-zinc-200">
            Editor
          </h2>
          <p class="text-sm tracking-wide text-zinc-500 dark:text-zinc-400">
            Review, search, and edit invoices
          </p>
        </div>
      </div>
    </section>

    <!-- Top action/title panel -->
    <section class="relative mb-4 overflow-visible">
      <div
        class="relative overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
      >
        <div class="relative p-4">
          <div class="flex flex-col items-start justify-between gap-3 sm:flex-row">
            <DecorGradient />

            <div class="flex items-center gap-3">
              <div class="min-w-0">
                <div class="flex items-center gap-2">
                  <h3 class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">
                    Invoice Book
                  </h3>

                  <span
                    class="hidden rounded-full border border-sky-200 bg-sky-50 px-2 py-0.5 text-[11px] font-medium text-sky-700 sm:inline-flex dark:border-emerald-400/20 dark:bg-emerald-950/25 dark:text-emerald-200"
                  >
                    Browse saved invoices
                  </span>
                </div>

                <p class="mt-0.5 text-xs text-zinc-500 dark:text-zinc-300">
                  Open and select an invoice or revision to edit
                </p>
              </div>
            </div>

            <div class="w-full sm:w-auto">
              <InvoiceBook
                :active-node="activeNode"
                @select="activeNode = $event"
              />
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Editor surface -->
    <section
      class="overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
    >
      <div class="border-b border-zinc-200 px-4 py-3 dark:border-zinc-800">
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <div class="text-base font-semibold text-zinc-900 dark:text-zinc-100">
              Invoice Editor
            </div>
            <div class="text-xs text-sky-600 dark:text-emerald-400">
              {{ selectionLabel }}
            </div>
          </div>

          <span
            v-if="activeNode"
            class="inline-flex shrink-0 rounded-full border border-zinc-200 bg-white/90 px-2 py-0.5 text-[11px] font-medium text-zinc-600 backdrop-blur-sm dark:border-zinc-700 dark:bg-zinc-900/70 dark:text-zinc-400"
          >
            {{ activeNode.type === 'invoice' ? 'Base invoice' : 'Revision' }}
          </span>
        </div>
      </div>

      <div class="p-4">
        <div
          v-if="!activeNode"
          class="flex min-h-80 items-center justify-center rounded-2xl border border-dashed border-zinc-200 bg-zinc-50/60 px-6 py-10 dark:border-zinc-800 dark:bg-zinc-900/30"
        >
          <div class="mx-auto max-w-sm text-center">
            <div
              class="mx-auto mb-3 flex h-11 w-11 items-center justify-center rounded-2xl border border-zinc-200 bg-white text-zinc-500 shadow-sm dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-400"
            >
              <DocumentIcon class="size-5" />
            </div>

            <div class="text-sm font-semibold text-zinc-800 dark:text-zinc-100">
              Select an invoice to start editing
            </div>
            <div class="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
              Use the invoice book above to choose a base invoice or revision.
            </div>
          </div>
        </div>

        <div
          v-else
          class="rounded-2xl border border-dashed border-zinc-200 px-4 py-10 dark:border-zinc-800"
        >
          <div class="text-sm font-medium text-zinc-800 dark:text-zinc-100">Editing surface</div>
          <div class="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
            Render the selected invoice or revision here.
          </div>
        </div>
      </div>
    </section>
  </main>
</template>

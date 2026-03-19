<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import InvoiceBook from '@/components/editor/InvoiceBook.vue'
import type { ActiveEditorNode } from '@/components/editor/invBookTypes'
import DecorGradient from '@/components/UI/DecorGradient.vue'
import { DocumentIcon, PencilSquareIcon } from '@heroicons/vue/24/outline'
import { useEditorStore } from '@/stores/editor'
import { useClientStore } from '@/stores/clients'
import EditorPreview from '@/components/editor/EditorPreview.vue'
import EditorSurface from '@/components/editor/EditorSurface.vue'

const clientStore = useClientStore()
const editStore = useEditorStore()

const activeNode = ref<ActiveEditorNode>(null)

watch(
  () => clientStore.selectedClient?.id,
  async (newClientId, oldClientId) => {
    if (newClientId === oldClientId) return

    activeNode.value = null
    editStore.clearActiveInvoice()
    editStore.clearInvoiceBook()

    if (!newClientId) return

    await editStore.fetchInvoiceBook(true)
  },
  { immediate: true },
)

watch(activeNode, async (node) => {
  if (!node) {
    editStore.clearActiveInvoice()
    return
  }

  if (node.type === 'invoice') {
    await editStore.fetchInvoice(node.baseNo, 1)
    return
  }

  await editStore.fetchInvoice(node.baseNo, node.revisionNo)
})
</script>

<template>
  <main class="mx-auto w-full max-w-4xl pb-16 sm:pb-0 2xl:max-w-5xl">
    <!-- Page header -->
    <section class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div class="flex items-center gap-3">
        <div
          class="grid size-12 shrink-0 place-items-center rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-700 dark:bg-zinc-900"
        >
          <PencilSquareIcon class="stroke-1.5 size-7 text-sky-600 dark:text-emerald-400" />
        </div>

        <div class="min-w-0">
          <h2 class="text-2xl font-semibold tracking-tight text-zinc-900 dark:text-zinc-200">
            Editor
          </h2>
          <p class="text-sm tracking-wide text-zinc-500 dark:text-zinc-400">
            Review, search, and edit invoices
          </p>
        </div>
      </div>
    </section>

    <!-- Invoice Book -->
    <section class="relative mb-4 overflow-visible">
      <div
        class="relative overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
      >
        <DecorGradient variant="gradientAndGrid" />

        <div class="relative p-3 md:p-4">
          <div class="flex flex-col items-start justify-between gap-3 sm:flex-row">
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

    <!-- Empty state -->
    <section v-if="!activeNode">
      <div
        class="flex min-h-80 items-center justify-center rounded-2xl border border-zinc-200 bg-white px-6 py-10 shadow-sm dark:border-zinc-800 dark:bg-zinc-900/30"
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
    </section>

    <!-- Invoice Preview -->
    <EditorPreview v-else-if="activeNode && !editStore.isEditing" />
    <EditorSurface v-else-if="activeNode && editStore.isEditing" />
  </main>
</template>

<script setup lang="ts">
// NOTE: To whoever might look at this in the future. I have made the conscious decision to duplicate invoice code to avoid complexity.
import {
  DocumentArrowDownIcon,
  InformationCircleIcon,
  XCircleIcon,
} from '@heroicons/vue/24/outline'
import TheButton from '../UI/TheButton.vue'
import { useEditorStore } from '@/stores/editor'
import EditorHeader from './partials/EditorHeader.vue'
import TheTooltip from '../UI/TheTooltip.vue'
import EditorItemPicker from './partials/EditorItemPicker.vue'
import EditorItemsTable from './partials/EditorItemsTable.vue'
import EditorAdjustments from './partials/EditorAdjustments.vue'
import EditorTotals from './partials/EditorTotals.vue'
import { usePdfStore } from '@/stores/pdf'
import { ref } from 'vue'
import EditorNote from './partials/EditorNote.vue'

const pdfStore = usePdfStore()
const editStore = useEditorStore()

const isGeneratingPdf = ref(false)

async function generatePdfOnly() {
  const inv = editStore.draftInvoice
  if (!inv || isGeneratingPdf.value) return

  isGeneratingPdf.value = true
  try {
    await pdfStore.quickGeneratePDF(inv)
  } finally {
    isGeneratingPdf.value = false
  }
}
</script>
<template>
  <main>
    <section
      v-if="editStore.activeInvoice"
      class="overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
    >
      <div
        class="flex flex-col gap-3 border-b border-zinc-200 px-3 py-2.5 sm:flex-row sm:items-start sm:justify-between dark:border-zinc-800"
      >
        <div class="min-w-0">
          <div class="text-base font-semibold text-zinc-800 dark:text-zinc-100">
            Invoice details
          </div>
          <div class="text-xs text-sky-600 dark:text-emerald-400">
            Edit fields and save as a revision to the current invoice
          </div>
        </div>
        <div class="flex w-full flex-col-reverse gap-2 sm:w-auto sm:flex-row">
          <TheTooltip text="Cancel edit and revert changes.">
            <TheButton
              @click="editStore.cancelEdit"
              class="w-full sm:w-auto"
              variant="danger"
            >
              <XCircleIcon class="size-4" />
              Cancel
            </TheButton>
          </TheTooltip>
          <TheTooltip text="Generate a PDF of the current invoice.">
            <TheButton
              class="flex w-full cursor-pointer items-center gap-2"
              :disabled="isGeneratingPdf"
              @click="generatePdfOnly"
            >
              <DocumentArrowDownIcon class="size-4" />
              PDF
            </TheButton>
          </TheTooltip>
          <TheTooltip text="Saves invoice as a new revision.">
            <TheButton
              type="button"
              class="w-full cursor-pointer truncate sm:w-auto"
              variant="success"
            >
              <DocumentArrowDownIcon class="size-4" />
              Save Invoice
            </TheButton>
          </TheTooltip>
        </div>
      </div>
      <EditorHeader></EditorHeader>
    </section>
    <EditorItemPicker></EditorItemPicker>
    <EditorItemsTable></EditorItemsTable>
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
          <EditorAdjustments />
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
          <TheTooltip
            text="Create a draft to save in invoice book. This lets you free edit invoice later."
            side="top"
            align="center"
            class="hover:text-sky-600 dark:hover:text-emerald-400"
          >
            <InformationCircleIcon class="size-5" />
          </TheTooltip>
        </div>
        <div class="p-3 md:p-4">
          <EditorTotals />
        </div>
      </section>
    </section>
    <EditorNote />
  </main>
</template>

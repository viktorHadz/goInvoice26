<script setup lang="ts">
import { DocumentArrowDownIcon, XCircleIcon } from '@heroicons/vue/24/outline'
import TheButton from '../UI/TheButton.vue'
import TheDropdown from '../UI/TheDropdown.vue'
import DetailsToolbar from '@/components/editor/partials/DetailsToolbar.vue'
import DetailsMenu from '@/components/editor/partials/DetailsMenu.vue'
import { useEditorStore } from '@/stores/editor'
import { useSettingsStore } from '@/stores/settings'
import EditorHeader from './partials/EditorHeader.vue'
import TheTooltip from '../UI/TheTooltip.vue'
import EditorItemPicker from './partials/EditorItemPicker.vue'
import EditorItemsTable from './partials/EditorItemsTable.vue'
import EditorAdjustments from './partials/EditorAdjustments.vue'
import EditorTotals from './partials/EditorTotals.vue'
import { usePdfStore } from '@/stores/pdf'
import { computed, ref } from 'vue'
import EditorNote from './partials/EditorNote.vue'
import { formatActiveEditorNodeLabel } from '@/utils/invoiceLabels'
import type { InvoiceStatus } from '@/components/invoice/invoiceTypes'
import { reachableStatuses } from '@/utils/invoiceStatusOptions'

const pdfStore = usePdfStore()
const editStore = useEditorStore()
const setsStore = useSettingsStore()

const isGeneratingPdf = ref(false)

const invoicePrefix = computed(() => setsStore.settings?.invoicePrefix ?? '')

const invoiceDisplayLabel = computed(() => {
  const i = editStore.draftInvoice
  const node = editStore.activeNode
  if (!i || !node) return ''
  return formatActiveEditorNodeLabel(invoicePrefix.value, node)
})

const lifecycleStatus = computed(() => (editStore.draftInvoice?.status ?? 'draft') as InvoiceStatus)

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

const revisionLocked = computed(() => {
  const st = editStore.draftInvoice?.status ?? 'draft'
  return st === 'paid' || st === 'void'
})

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
      <DetailsToolbar
        :identity-label="invoiceDisplayLabel"
        subtitle="Editing · save adds a new revision"
      >
        <template #actions>
          <div class="w-full max-w-40 sm:w-36 sm:max-w-none">
            <TheDropdown
              v-model="selectedInvoiceStatus"
              select-title="Status"
              select-title-class="text-xs font-medium text-zinc-500 dark:text-zinc-400"
              :options="statusSelectOptions"
              input-class="py-1.5 capitalize"
              placeholder="Status"
            />
          </div>
          <TheTooltip text="Cancel edit and revert changes.">
            <TheButton
              class="w-full sm:w-auto"
              variant="secondary"
              @click="editStore.cancelEdit"
            >
              <XCircleIcon class="size-4" />
              Cancel
            </TheButton>
          </TheTooltip>
          <DetailsMenu
            :pdf-disabled="isGeneratingPdf"
            @pdf="generatePdfOnly"
          />
          <TheTooltip
            :text="
              revisionLocked
                ? 'Cannot save while paid or void. In preview: reopen paid invoices to issued, or restore void invoices to issued.'
                : 'Saves invoice as a new revision.'
            "
          >
            <TheButton
              type="button"
              class="w-full sm:w-auto"
              variant="success"
              :disabled="revisionLocked"
              @click="editStore.saveRevision(editStore.draftInvoice)"
            >
              <DocumentArrowDownIcon class="size-4" />
              Save
            </TheButton>
          </TheTooltip>
        </template>
      </DetailsToolbar>
      <EditorHeader />
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
            <div class="text-xs text-zinc-600 dark:text-zinc-400">
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
            <div class="text-xs text-zinc-600 dark:text-zinc-400">Balance overview</div>
          </div>
        </div>
        <div class="p-3 md:p-4">
          <EditorTotals />
        </div>
      </section>
    </section>
    <EditorNote />
  </main>
</template>

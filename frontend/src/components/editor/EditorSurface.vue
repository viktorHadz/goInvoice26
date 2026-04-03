<script setup lang="ts">
import { DocumentArrowDownIcon, XCircleIcon } from '@heroicons/vue/24/outline'
import TheButton from '../UI/TheButton.vue'
import DetailsMenu, { type MenuOption } from '@/components/editor/partials/DetailsMenu.vue'
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
import { canEditInvoice } from '@/utils/invoiceStatusOptions'
import { resolveEditorExportRevisionNo } from '@/utils/editorExport'

const pdfStore = usePdfStore()
const editStore = useEditorStore()
const setsStore = useSettingsStore()

const isGeneratingExport = ref(false)

const invoicePrefix = computed(() => setsStore.settings?.invoicePrefix ?? '')

const invoiceDisplayLabel = computed(() => {
    const i = editStore.draftInvoice
    const node = editStore.activeNode
    if (!i || !node) return ''
    return formatActiveEditorNodeLabel(invoicePrefix.value, node)
})

const revisionLocked = computed(() => {
    const st = editStore.draftInvoice?.status ?? 'draft'
    return !canEditInvoice(st)
})

const saveTooltipText = computed(() => {
    const status = editStore.draftInvoice?.status ?? 'draft'

    if (status === 'paid') {
        return 'This invoice is marked as paid. It can only be reopened to issued if the recorded payments do not match the balance due.'
    }

    if (status === 'void') {
        return 'This invoice is void. Void invoices are final records and cannot be edited or deleted.'
    }

    if (!editStore.hasUnsavedChanges) {
        return status === 'draft'
            ? 'Make a change to this draft before saving it in place.'
            : 'Make a change to this invoice before saving a new revision.'
    }

    return status === 'draft'
        ? 'Save your edits directly to this draft invoice.'
        : 'Save your edits as a new invoice revision.'
})

async function generatePdfOnly() {
    const inv = editStore.draftInvoice
    if (!inv || isGeneratingExport.value) return

    const selectedRevisionNo = resolveEditorExportRevisionNo(
        editStore.activeNode,
        editStore.activeRevisionNo,
    )

    isGeneratingExport.value = true
    try {
        await pdfStore.quickGeneratePDF(inv, selectedRevisionNo)
    } finally {
        isGeneratingExport.value = false
    }
}

async function generateDocxOnly() {
    const inv = editStore.draftInvoice
    if (!inv || isGeneratingExport.value) return

    const selectedRevisionNo = resolveEditorExportRevisionNo(
        editStore.activeNode,
        editStore.activeRevisionNo,
    )

    isGeneratingExport.value = true
    try {
        await pdfStore.quickGenerateDocx(inv, selectedRevisionNo)
    } finally {
        isGeneratingExport.value = false
    }
}

const menuOpts = computed<MenuOption[]>(() => [
    {
        id: 1,
        name: 'Generate PDF',
        disabled: isGeneratingExport.value,
        disabledReason: 'Processing invoice generation please try again later. ',
        effect: generatePdfOnly,
        icon: DocumentArrowDownIcon,
    },
    {
        id: 2,
        name: 'Generate Docx',
        disabled: isGeneratingExport.value,
        disabledReason: 'Processing invoice generation please try again later. ',
        effect: generateDocxOnly,
        icon: DocumentArrowDownIcon,
    },
])
</script>
<template>
    <div class="space-y-4">
        <section
            v-if="editStore.activeInvoice"
            class="rounded-2xl border border-zinc-300 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
        >
            <div
                class="hdr-grid rounded-t-2xl border-b border-zinc-300 px-4 py-3 dark:border-zinc-800"
            >
                <div class="flex flex-wrap items-center justify-between gap-x-4 gap-y-2">
                    <div class="min-w-0">
                        <h2 class="text-base font-semibold text-zinc-800 dark:text-zinc-100">
                            {{ invoiceDisplayLabel }}
                        </h2>
                        <p class="mt-0.5 text-xs font-medium text-zinc-600 dark:text-zinc-400">
                            <span class="font-bold text-sky-600 dark:text-emerald-400">
                                Editing
                            </span>
                        </p>
                    </div>
                    <div class="flex items-center gap-2">
                        <TheTooltip text="Cancel edit and revert changes">
                            <TheButton
                                variant="secondary"
                                class="cursor-pointer"
                                @click="editStore.cancelEdit"
                            >
                                <XCircleIcon class="size-4" />
                                Cancel
                            </TheButton>
                        </TheTooltip>
                        <TheTooltip :text="saveTooltipText">
                            <TheButton
                                type="button"
                                variant="success"
                                :disabled="revisionLocked || !editStore.hasUnsavedChanges"
                                :class="
                                    revisionLocked || !editStore.hasUnsavedChanges
                                        ? 'cursor-not-allowed'
                                        : 'cursor-pointer'
                                "
                                @click="editStore.saveRevision(editStore.draftInvoice)"
                            >
                                <DocumentArrowDownIcon class="size-4" />
                                Save
                            </TheButton>
                        </TheTooltip>

                        <DetailsMenu :options="menuOpts" />
                    </div>
                </div>
            </div>
            <EditorHeader />
        </section>
        <EditorItemPicker />
        <EditorItemsTable />
        <section class="grid gap-4 md:grid-cols-2">
            <section
                class="overflow-hidden rounded-2xl border border-zinc-300 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
            >
                <div
                    class="hdr-grid flex items-start justify-between gap-3 border-b border-zinc-300 px-3 py-2.5 dark:border-zinc-800"
                >
                    <div class="min-w-0">
                        <div class="text-base font-semibold text-zinc-800 dark:text-zinc-100">
                            Adjustments
                        </div>
                        <div class="text-xs font-bold text-sky-600 dark:text-emerald-400">
                            Paid, deposit, discount, VAT and note
                        </div>
                    </div>
                </div>
                <div class="p-3 md:p-4">
                    <EditorAdjustments />
                </div>
            </section>
            <section
                class="overflow-hidden rounded-2xl border border-zinc-300 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
            >
                <div
                    class="hdr-grid flex items-start justify-between gap-3 border-b border-zinc-300 px-3 py-2.5 dark:border-zinc-800"
                >
                    <div class="min-w-0">
                        <div class="text-base font-semibold text-zinc-800 dark:text-zinc-100">
                            Totals
                        </div>
                        <div class="text-xs font-bold text-sky-600 dark:text-emerald-400">
                            Balance overview
                        </div>
                    </div>
                </div>
                <div class="p-3 md:p-4">
                    <EditorTotals />
                </div>
            </section>
        </section>
        <EditorNote />
    </div>
</template>

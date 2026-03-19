<script setup lang="ts">
// NOTE: To whoever might look at this in the future. I have made the conscious decision to duplicate invoice code to avoid complexity.
import { DocumentArrowDownIcon, XCircleIcon } from '@heroicons/vue/24/outline'
import TheButton from '../UI/TheButton.vue'
import { useEditorStore } from '@/stores/editor'
import EditorHeader from './partials/EditorHeader.vue'
import TheTooltip from '../UI/TheTooltip.vue'
import EditorItemPicker from './partials/EditorItemPicker.vue'
import EditorItemsTable from './partials/EditorItemsTable.vue'

const editStore = useEditorStore()
</script>
<template>
  <div
    v-if="editStore.activeInvoice"
    class="overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
  >
    <div
      class="flex flex-col gap-3 border-b border-zinc-200 px-3 py-2.5 sm:flex-row sm:items-start sm:justify-between dark:border-zinc-800"
    >
      <div class="min-w-0">
        <div class="text-base font-semibold text-zinc-800 dark:text-zinc-100">Invoice details</div>
        <div class="text-xs text-sky-600 dark:text-emerald-400">
          Edit fields and save as a revision to the current invoice
        </div>
      </div>

      <div class="flex w-full flex-col-reverse gap-2 sm:w-auto sm:flex-row">
        <TheTooltip text="Cancel edit and revert changes.">
          <TheButton
            @click="editStore.cancelEdit"
            class="w-full cursor-pointer sm:w-auto"
            variant="danger"
          >
            <XCircleIcon class="size-4" />
            Cancel
          </TheButton>
        </TheTooltip>

        <TheTooltip text="Saves invoice as a new revision.">
          <TheButton
            type="button"
            class="w-full cursor-pointer sm:w-auto"
            variant="success"
          >
            <DocumentArrowDownIcon class="size-4" />
            Save Invoice
          </TheButton>
        </TheTooltip>
      </div>
    </div>

    <EditorHeader></EditorHeader>
  </div>
  <EditorItemPicker></EditorItemPicker>
  <EditorItemsTable></EditorItemsTable>
</template>

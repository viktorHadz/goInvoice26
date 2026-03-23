<script setup lang="ts">
import { computed } from 'vue'
import { useEditorStore } from '@/stores/editor'
import EditorLineRow from './EditorLineRow.vue'

const editStore = useEditorStore()

const lines = computed(() => {
  const list = editStore.draftInvoice?.lines ?? []
  return [...list].sort((a, b) => (b.sortOrder ?? 0) - (a.sortOrder ?? 0))
})
</script>

<template>
  <section
    class="overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
  >
    <div class="flex gap-3 border-b border-zinc-200 px-3 py-2.5 dark:border-zinc-800">
      <div class="min-w-0">
        <div class="text-base font-semibold text-zinc-800 dark:text-zinc-100">Invoice items</div>
        <div class="text-xs text-zinc-600 dark:text-zinc-400">Edit line items for this invoice</div>
      </div>
    </div>

    <div class="p-2.5 md:p-3">
      <div class="overflow-x-auto">
        <div class="min-w-160">
          <p
            v-if="editStore.showAllValidation && editStore.getFieldError('lines')"
            class="px-2 py-2 text-xs text-rose-600 dark:text-rose-300"
          >
            {{ editStore.getFieldError('lines') }}
          </p>

          <div
            class="grid grid-cols-[minmax(220px,1fr)_48px_64px_96px_110px_36px] items-center gap-3 px-2 py-2 text-sm font-semibold text-zinc-600 dark:text-zinc-200"
          >
            <div class="truncate">Product name</div>
            <div class="text-right">Qty</div>
            <div class="text-right">Mins</div>
            <div class="text-right">Unit</div>
            <div class="text-right">Total</div>
            <div></div>
          </div>

          <div class="h-px bg-zinc-200 dark:bg-zinc-800" />

          <div
            v-if="!lines.length"
            class="px-3 py-10 text-base text-zinc-500 dark:text-zinc-400"
          >
            No items yet. Add from the picker above.
          </div>

          <div
            v-else
            class="max-h-136 divide-y divide-zinc-200 overflow-y-auto dark:divide-zinc-800"
          >
            <EditorLineRow
              v-for="(l, idx) in lines"
              :key="l.sortOrder"
              :line="l"
              :line-index="idx"
            />
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

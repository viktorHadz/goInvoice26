<script setup lang="ts">
import { useEditorStore } from '@/stores/editor'
import { useInvoiceStore } from '@/stores/invoice'
import { computed, ref, watch } from 'vue'

const editStore = useEditorStore()
const noteProxy = computed<string>({
  get: () => editStore.draftInvoice?.note ?? '',
  set: (v) => editStore.setNote(String(v ?? '')),
})
const noteTouched = ref(false)

const NOTE_TEXT_LIMIT = 1000

function syncFromInvoice() {
  const v = editStore.draftInvoice
  if (!v) return

  noteTouched.value = false
}

watch(
  () => editStore.draftInvoice,
  () => syncFromInvoice(),
  { immediate: true },
)
</script>
<template>
  <section
    class="mt-4 overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
  >
    <div
      class="flex items-start justify-between gap-3 border-b border-zinc-200 px-3 py-2.5 dark:border-zinc-800"
    >
      <div class="min-w-0">
        <div class="text-base font-semibold text-zinc-800 dark:text-zinc-100">Note</div>
        <div class="text-xs text-sky-600 dark:text-emerald-400">
          Extra text shown on the invoice
        </div>
      </div>
      <div
        class="mt-1 text-right text-xs"
        :class="
          noteProxy.length > NOTE_TEXT_LIMIT * 0.9
            ? 'text-rose-600 dark:text-rose-300'
            : noteProxy.length > NOTE_TEXT_LIMIT * 0.8
              ? 'text-amber-500 dark:text-amber-400'
              : 'text-zinc-500 dark:text-zinc-400'
        "
      >
        {{ noteProxy.length }}/{{ NOTE_TEXT_LIMIT }}
      </div>
    </div>

    <div class="p-2.5 md:p-3">
      <textarea
        id="invoice-adjustments-text-area"
        v-model="noteProxy"
        class="input input-accent w-full resize-y rounded-xl px-3 py-2"
        :disabled="!editStore.draftInvoice"
        placeholder="Add a note to the invoice…"
        @blur.stop="noteTouched = true"
        maxlength="1000"
      />

      <p
        class="mt-1 min-h-5 text-xs"
        :class="
          editStore.getFieldError('note') && (noteTouched || editStore.showAllValidation)
            ? 'text-rose-600 dark:text-rose-300'
            : 'text-transparent'
        "
      >
        {{
          editStore.getFieldError('note') && (noteTouched || editStore.showAllValidation)
            ? editStore.getFieldError('note')
            : '•'
        }}
      </p>
    </div>
  </section>
</template>

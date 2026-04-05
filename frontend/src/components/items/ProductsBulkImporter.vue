<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useProductStore } from '@/stores/products'
import { useClientStore } from '@/stores/clients'
import type { ProductType } from '@/utils/productHttpHandler'
import {
  PRODUCT_IMPORT_ACCEPT,
  PRODUCT_IMPORT_KIND_OPTIONS,
  PRODUCT_IMPORT_MAX_BYTES,
  PRODUCT_IMPORT_MAX_ROWS,
  formatProductImportErrors,
  getProductImportOption,
  validateProductImportFile,
} from '@/utils/productImport'
import type {
  FormattedProductImportError,
  ProductImportKind,
  ProductImportKindOption,
} from '@/utils/productImport'
import { hasFieldErrors, isApiError } from '@/utils/apiErrors'
import type { APIFieldError } from '@/utils/apiErrors'
import { ChevronDownIcon, ArrowUpTrayIcon, ChevronUpIcon } from '@heroicons/vue/24/outline'
import TheDropdown from '../UI/TheDropdown.vue'
import TheButton from '../UI/TheButton.vue'
import TheTooltip from '../UI/TheTooltip.vue'
import { emitToastSuccess } from '@/utils/toast'
import { handleActionError } from '@/utils/errors/handleActionError'

const props = withDefaults(
  defineProps<{
    busy?: boolean
  }>(),
  {
    busy: false,
  },
)

const emit = defineEmits<{
  (e: 'imported', productType: ProductType): void
}>()

const store = useProductStore()
const clientStore = useClientStore()

const open = ref(false)
const importKind = ref<ProductImportKind>('style')
const importFile = ref<File | null>(null)
const importFieldErrors = ref<Record<string, string>>({})
const importErrors = ref<FormattedProductImportError[]>([])
const isImporting = ref(false)
const importFileRef = ref<HTMLInputElement | null>(null)

const selectedClientLabel = computed(
  () => clientStore.selectedClient?.name?.trim() || 'No client selected',
)
const selectedImportOption = computed<ProductImportKindOption | null>({
  get: () => getProductImportOption(importKind.value),
  set: (value) => {
    importKind.value = value?.id ?? 'style'
  },
})
const importSpec = computed(() => getProductImportOption(importKind.value))
const importCanRun = computed(
  () =>
    open.value &&
    !!clientStore.selectedClient &&
    !!importFile.value &&
    !isImporting.value &&
    !props.busy,
)
const importTopError = computed(
  () => importFieldErrors.value.request || importFieldErrors.value.header || null,
)
const importKindError = computed(() => importFieldErrors.value.kind || null)
const importFileError = computed(() => importFieldErrors.value.file || null)
const selectedImportFileName = computed(() => importFile.value?.name?.trim() || '')

watch(importKind, () => {
  importFieldErrors.value = {}
  importErrors.value = []
})

watch(
  () => clientStore.selectedClient?.id ?? null,
  () => {
    clearImportState()
  },
)

function clearImportState() {
  importFile.value = null
  importFieldErrors.value = {}
  importErrors.value = []
  if (importFileRef.value) {
    importFileRef.value.value = ''
  }
}

function applyImportFieldErrors(fields: APIFieldError[]) {
  const next: Record<string, string> = {}
  importErrors.value = formatProductImportErrors(fields)

  for (const fieldError of fields) {
    if (!fieldError.field || next[fieldError.field]) continue
    if (
      fieldError.field !== 'file' &&
      fieldError.field !== 'kind' &&
      fieldError.field !== 'header' &&
      fieldError.field !== 'request'
    ) {
      continue
    }

    next[fieldError.field] =
      typeof fieldError.message === 'string' && fieldError.message.trim().length > 0
        ? fieldError.message
        : 'Please review this CSV upload.'
  }

  importFieldErrors.value = next
}

function onImportFileChange(event: Event) {
  importFieldErrors.value = {}
  importErrors.value = []

  const target = event.target as HTMLInputElement | null
  const files = target?.files

  if (!files || files.length === 0) {
    importFile.value = null
    return
  }

  if (files.length > 1) {
    importFile.value = null
    importFieldErrors.value = { file: 'Select a single CSV file.' }
    if (target) target.value = ''
    return
  }

  try {
    importFile.value = validateProductImportFile(files[0])
  } catch (err: unknown) {
    importFile.value = null
    importFieldErrors.value = {
      file: err instanceof Error && err.message.trim() ? err.message : 'Upload a CSV file.',
    }
    if (target) target.value = ''
  }
}

async function runImport() {
  if (isImporting.value || props.busy) return

  importFieldErrors.value = {}
  importErrors.value = []

  if (!clientStore.selectedClient) {
    importFieldErrors.value = { request: 'Select a client before importing products.' }
    return
  }

  if (!importFile.value) {
    importFieldErrors.value = { file: 'Select a CSV file.' }
    return
  }

  let file: File
  try {
    file = validateProductImportFile(importFile.value)
  } catch (err: unknown) {
    importFieldErrors.value = {
      file: err instanceof Error && err.message.trim() ? err.message : 'Upload a CSV file.',
    }
    return
  }

  isImporting.value = true

  try {
    const result = await store.importCsv(importKind.value, file)
    emit('imported', importSpec.value.productType)
    emitToastSuccess(
      `${result.createdCount} ${result.createdCount === 1 ? 'product was' : 'products were'} imported.`,
      {
        title: selectedClientLabel.value,
      },
    )
    clearImportState()
    open.value = false
  } catch (err: unknown) {
    if (isApiError(err) && hasFieldErrors(err)) {
      applyImportFieldErrors(err.fields)
      open.value = true
      return
    }

    handleActionError(err, {
      toastTitle: 'Import failed',
      mapFields: false,
    })
  } finally {
    isImporting.value = false
  }
}
</script>

<template>
  <section class="pt-2">
    <div class="flex items-start justify-between gap-3">
      <div class="min-w-0">
        <TheTooltip text="Add multiple products to the currently selected client from a CSV file.">
          <button
            type="button"
            class="inline-flex cursor-pointer items-center gap-2 text-sm font-semibold text-zinc-900 transition hover:text-sky-700 focus-visible:ring-0 dark:text-zinc-100 dark:hover:text-emerald-300"
            :aria-expanded="open ? 'true' : 'false'"
            @click="open = !open"
          >
            <span>Bulk import products</span>
            <ChevronDownIcon
              class="size-4 transition-transform"
              :class="open ? 'rotate-180' : ''"
            />
          </button>
        </TheTooltip>
        <div class="mt-1 text-xs text-zinc-600 dark:text-zinc-400">
          Add products to the selected client from a CSV file.
        </div>
      </div>
    </div>

    <transition
      enter-active-class="transition duration-150 ease-out"
      enter-from-class="opacity-0 -translate-y-1"
      enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition duration-120 ease-in"
      leave-from-class="opacity-100 translate-y-0"
      leave-to-class="opacity-0 -translate-y-1"
    >
      <div
        v-if="open"
        class="mt-4 space-y-4"
      >
        <div
          class="rounded-lg border border-amber-300/70 bg-amber-50 px-3 py-3 text-xs text-amber-800 dark:border-amber-400/30 dark:bg-amber-950/25 dark:text-amber-200"
        >
          Import target:
          <span class="font-semibold">{{ selectedClientLabel }}.</span>
          <span class="block">New products will be added to this client only.</span>
        </div>

        <div>
          <TheDropdown
            v-model="selectedImportOption"
            :options="PRODUCT_IMPORT_KIND_OPTIONS"
            select-title="Import type"
            :right-icon="ChevronDownIcon"
            :disabled="isImporting || busy"
          />
          <p class="mt-2 text-xs text-zinc-600 dark:text-zinc-400">
            {{ importSpec.summary }}
          </p>
          <p
            v-if="importKindError"
            class="mt-2 text-xs text-rose-600 dark:text-rose-300"
          >
            {{ importKindError }}
          </p>
        </div>

        <div
          class="rounded-lg border border-zinc-300 bg-white px-3 py-3 dark:border-zinc-800 dark:bg-zinc-950/40"
        >
          <div
            class="text-xs font-semibold tracking-[0.12em] text-zinc-700 uppercase dark:text-emerald-400"
          >
            CSV instructions
          </div>
          <p class="mt-2 text-xs text-zinc-600 dark:text-zinc-400">
            Use comma-separated values only (UTF-8). Column names and order must exactly match the
            header below.
          </p>
          <div
            class="mt-3 rounded-lg border bg-zinc-100 px-3 py-2 text-xs text-zinc-800 dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-100"
          >
            {{ importSpec.csvHeader }}
          </div>
          <div class="mt-3 text-xs text-zinc-600 dark:text-zinc-400">
            Limits: {{ PRODUCT_IMPORT_MAX_ROWS }} rows max, {{ PRODUCT_IMPORT_MAX_BYTES / 1024 }}KB
            max file size, one import every 30 seconds.
          </div>
        </div>

        <div>
          <label
            for="product-import-file"
            class="input-label"
          >
            CSV file
          </label>

          <input
            id="product-import-file"
            ref="importFileRef"
            type="file"
            :accept="PRODUCT_IMPORT_ACCEPT"
            class="input input-accent w-full cursor-pointer file:mr-3 file:cursor-pointer file:border-0 file:bg-transparent file:font-medium"
            :disabled="!clientStore.selectedClient || isImporting || busy"
            @change="onImportFileChange"
          />

          <p
            v-if="selectedImportFileName"
            class="mt-2 text-xs text-zinc-600 dark:text-zinc-400"
          >
            Selected: {{ selectedImportFileName }}
          </p>
          <p
            v-else
            class="mt-2 text-xs text-zinc-600 dark:text-zinc-400"
          >
            Select one CSV file. The upload is processed in memory and not stored afterwards.
          </p>

          <p
            v-if="importFileError"
            class="mt-2 text-xs text-rose-600 dark:text-rose-300"
          >
            {{ importFileError }}
          </p>
        </div>

        <div
          v-if="importTopError"
          class="rounded-xl border border-rose-200 bg-rose-50 px-3 py-2 text-xs text-rose-700 dark:border-rose-400/20 dark:bg-rose-950/25 dark:text-rose-200"
        >
          {{ importTopError }}
        </div>

        <div
          v-if="importErrors.length > 0"
          class="rounded-xl border border-zinc-300 bg-white px-3 py-3 dark:border-zinc-800 dark:bg-zinc-950/40"
        >
          <div
            class="text-xs font-semibold tracking-[0.12em] text-zinc-700 uppercase dark:text-zinc-300"
          >
            Import issues
          </div>
          <div class="mt-2 max-h-48 space-y-2 overflow-y-auto pr-1">
            <div
              v-for="error in importErrors"
              :key="error.id"
              class="rounded-lg border border-zinc-300/80 bg-zinc-50 px-3 py-2 text-xs text-zinc-700 dark:border-zinc-800 dark:bg-zinc-950/30 dark:text-zinc-200"
            >
              {{ error.message }}
            </div>
          </div>
        </div>

        <div class="flex flex-wrap gap-2">
          <TheButton
            :disabled="!importCanRun"
            @click="runImport"
          >
            <ChevronUpIcon class="size-4" />
            {{ isImporting ? 'Importing…' : 'Import CSV' }}
          </TheButton>

          <TheButton
            variant="secondary"
            :disabled="isImporting"
            @click="clearImportState"
          >
            Clear
          </TheButton>
        </div>
      </div>
    </transition>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { TrashIcon } from '@heroicons/vue/24/outline'
import TheInput from '@/components/UI/TheInput.vue'
import type { InvoiceLine } from '@/components/invoice/invoiceTypes'
import { fmtGBPMinor, fromMinor, lineTotalMinor, toMinor } from '@/utils/money'
import { useInvoiceStore } from '@/stores/invoice'

const props = defineProps<{ line: InvoiceLine; lineIndex: number }>()

const invStore = useInvoiceStore()

const totalMinor = computed(() => lineTotalMinor(props.line))
const serverTotalMinor = computed(
  () => invStore.serverCanonicalLineTotals?.[props.line.sortOrder] ?? null,
)
const serverMismatch = computed(
  () => typeof serverTotalMinor.value === 'number' && serverTotalMinor.value !== totalMinor.value,
)
const minutesDisabled = computed(() => props.line.pricingMode !== 'hourly')

const unitPounds = computed(() => fromMinor(props.line.unitPriceMinor))

function fieldError(field: string) {
  return invStore.getFieldError(`lines[${props.lineIndex}].${field}`)
}

function setName(v: unknown) {
  invStore.updateLine(props.line.sortOrder, { name: String(v ?? '') })
}

function setQty(v: unknown) {
  if (v === '' || v === null || v === undefined) {
    invStore.updateLine(props.line.sortOrder, { quantity: 0 })
    return
  }
  const n = Number(v)
  if (!Number.isFinite(n) || n < 0) return
  invStore.updateLine(props.line.sortOrder, { quantity: n })
}

function setMinutes(v: unknown) {
  const n = Number(v)
  if (!Number.isFinite(n) || n < 0) return
  invStore.updateLine(props.line.sortOrder, { minutesWorked: n })
}

function setUnitPounds(v: unknown) {
  const n = Number(v)
  if (!Number.isFinite(n) || n < 0) return
  invStore.updateLine(props.line.sortOrder, { unitPriceMinor: toMinor(n) })
}
</script>

<template>
  <div
    class="grid w-full grid-cols-[minmax(220px,1fr)_48px_64px_96px_110px_36px] items-start gap-2 px-2 py-3"
  >
    <!-- name -->
    <div class="min-w-0">
      <TheInput
        type="text"
        :modelValue="line.name"
        @update:modelValue="setName"
        inputClass="py-1 text-sm"
        placeholder="Product name"
        :error="fieldError('name')"
      />
      <div class="truncate text-sm text-zinc-500 capitalize dark:text-zinc-400">
        {{ line.lineType }} · {{ line.pricingMode }}
      </div>
    </div>

    <!-- qty -->
    <div class="min-w-0 text-right">
      <TheInput
        type="number"
        :modelValue="line.quantity"
        @update:modelValue="setQty"
        inputClass="input-compact text-right tabular-nums"
        :error="fieldError('quantity')"
      />
    </div>

    <!-- minutes -->
    <div class="min-w-0 text-right">
      <TheInput
        type="number"
        :modelValue="line.minutesWorked ?? 0"
        @update:modelValue="setMinutes"
        inputClass="input-compact text-right tabular-nums"
        :disabled="minutesDisabled"
        :placeholder="minutesDisabled ? '—' : '60'"
        :title="minutesDisabled ? 'Only hourly lines use minutes' : 'Minutes worked'"
        :error="fieldError('minutesWorked')"
      />
    </div>

    <!-- unit -->
    <div class="min-w-0 text-right">
      <TheInput
        type="number"
        :modelValue="unitPounds"
        @update:modelValue="setUnitPounds"
        inputClass="input-compact text-right tabular-nums"
        :title="line.pricingMode === 'hourly' ? 'Hourly rate (£)' : 'Unit price (£)'"
        :error="fieldError('unitPriceMinor')"
      />
      <div class="truncate text-sm text-zinc-500 dark:text-zinc-400">
        {{ fmtGBPMinor(line.unitPriceMinor) }}{{ line.pricingMode === 'hourly' ? '/hr' : '' }}
      </div>
    </div>

    <!-- total -->
    <div class="min-w-0 text-right">
      <div
        class="text-base font-semibold tabular-nums"
        :class="{
          'text-zinc-900 dark:text-zinc-100': !serverMismatch,
          'text-amber-700 dark:text-amber-300': serverMismatch,
        }"
        :title="
          serverMismatch && serverTotalMinor != null
            ? `Server: ${fmtGBPMinor(serverTotalMinor)}`
            : ''
        "
      >
        {{ fmtGBPMinor(totalMinor) }}
      </div>
      <div
        v-if="serverMismatch && serverTotalMinor != null"
        class="text-xs text-amber-700 dark:text-amber-300"
      >
        Server: {{ fmtGBPMinor(serverTotalMinor) }}
      </div>
    </div>

    <!-- remove -->
    <div class="flex justify-end">
      <button
        type="button"
        class="cursor-pointer rounded-md border border-transparent p-1 text-zinc-600 hover:border-rose-600/20 hover:bg-rose-50 hover:text-rose-500 dark:text-zinc-300 dark:hover:border-rose-300/20 dark:hover:bg-rose-900/20 dark:hover:text-rose-300"
        @click="invStore.removeLine(line.sortOrder)"
      >
        <TrashIcon class="size-5" />
      </button>
    </div>
  </div>
</template>

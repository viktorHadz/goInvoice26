<script setup lang="ts">
import { computed } from 'vue'
import { TrashIcon } from '@heroicons/vue/24/outline'
import TheInput from '@/components/UI/TheInput.vue'
import TheButton from '@/components/UI/TheButton.vue'
import { useInvoiceDraftStore } from '@/stores/invoiceDraft'
import { fmtGBPMinor, fromMinor, toMinor } from '@/utils/money'
import { lineTotalMinor } from '@/utils/invoiceMath'
import type { InvoiceLine } from './invoiceTypes'

const props = defineProps<{ line: InvoiceLine }>()

const inv = useInvoiceDraftStore()

const totalMinor = computed(() => lineTotalMinor(props.line))
const minutesDisabled = computed(() => props.line.pricingMode !== 'hourly')

const unitPounds = computed(() => fromMinor(props.line.unitPriceMinor))

function setName(v: any) {
  inv.updateLine(props.line.sortOrder, { name: String(v ?? '') })
}

function setQty(v: any) {
  const n = Number(v)
  if (!Number.isFinite(n) || n <= 0) return
  inv.updateLine(props.line.sortOrder, { quantity: n })
}

function setMinutes(v: any) {
  const n = Number(v)
  if (!Number.isFinite(n) || n < 0) return
  inv.updateLine(props.line.sortOrder, { minutesWorked: n })
}

function setUnitPounds(v: any) {
  const n = Number(v)
  if (!Number.isFinite(n) || n < 0) return
  inv.updateLine(props.line.sortOrder, { unitPriceMinor: toMinor(n) })
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
        inputClass="py-1 text-sm "
        placeholder="Product name"
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
      />
      <div class="truncate text-sm text-zinc-500 dark:text-zinc-400">
        {{ fmtGBPMinor(line.unitPriceMinor) }}{{ line.pricingMode === 'hourly' ? '/hr' : '' }}
      </div>
    </div>

    <!-- total -->
    <div
      class="min-w-0 text-right text-base font-semibold text-zinc-900 tabular-nums dark:text-zinc-100"
    >
      {{ fmtGBPMinor(totalMinor) }}
    </div>

    <!-- remove -->
    <div class="flex justify-end">
      <button
        class="cursor-pointer rounded-md border border-transparent p-1 text-zinc-600 hover:border-rose-600/20 hover:bg-rose-50 hover:text-rose-500 dark:text-zinc-300 dark:hover:border-rose-300/20 dark:hover:bg-rose-900/20 dark:hover:text-rose-300"
        @click="inv.removeLine(line.sortOrder)"
      >
        <TrashIcon class="size-5" />
      </button>
    </div>
  </div>
</template>

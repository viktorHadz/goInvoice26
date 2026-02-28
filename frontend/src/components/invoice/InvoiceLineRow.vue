<script setup lang="ts">
import { computed } from 'vue'
import { TrashIcon } from '@heroicons/vue/24/outline'
import TheInput from '@/components/UI/TheInput.vue'
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
  <div class="grid grid-cols-1 gap-3 px-2 py-3 md:grid-cols-10 md:items-center md:gap-3">
    <!-- name -->
    <div class="min-w-0 md:col-span-5">
      <TheInput
        type="text"
        :modelValue="line.name"
        @update:modelValue="setName"
        inputClass="py-1"
        placeholder="Product name"
      />
      <div class="text-sm text-zinc-500 capitalize dark:text-zinc-400">
        {{ line.lineType }} · {{ line.pricingMode }}
      </div>
    </div>

    <div class="grid grid-cols-2 gap-2 sm:grid-cols-4 md:col-span-4 md:grid-cols-4 md:gap-3">
      <!-- qty -->
      <div class="text-left md:text-right">
        <div class="mb-1 text-xs font-medium text-zinc-500 md:hidden dark:text-zinc-400">Qty</div>
        <TheInput
          type="number"
          :modelValue="line.quantity"
          @update:modelValue="setQty"
          inputClass="py-1"
        />
      </div>

      <!-- minutes -->
      <div class="text-left md:text-right">
        <div class="mb-1 text-xs font-medium text-zinc-500 md:hidden dark:text-zinc-400">Minutes</div>
        <TheInput
          type="number"
          :modelValue="line.minutesWorked ?? 0"
          @update:modelValue="setMinutes"
          inputClass="py-1"
          :disabled="minutesDisabled"
          :placeholder="minutesDisabled ? '—' : '60'"
          :title="minutesDisabled ? 'Only hourly lines use minutes' : 'Minutes worked'"
        />
      </div>

      <!-- unit -->
      <div class="text-left md:text-right">
        <div class="mb-1 text-xs font-medium text-zinc-500 md:hidden dark:text-zinc-400">Unit</div>
        <TheInput
          type="number"
          :modelValue="unitPounds"
          inputClass="py-1"
          @update:modelValue="setUnitPounds"
          :title="line.pricingMode === 'hourly' ? 'Hourly rate (£)' : 'Unit price (£)'"
        />
      </div>

      <!-- total -->
      <div class="text-left md:text-right">
        <div class="mb-1 text-xs font-medium text-zinc-500 md:hidden dark:text-zinc-400">Total</div>
        <div class="text-base font-semibold text-zinc-900 dark:text-zinc-100">
          {{ fmtGBPMinor(totalMinor) }}
        </div>
      </div>
    </div>

    <!-- remove -->
    <div class="flex justify-end md:col-span-1">
      <button
        class="cursor-pointer rounded-md border border-transparent p-1 text-zinc-600 hover:border-rose-600/20 hover:bg-rose-50 hover:text-rose-500 dark:text-zinc-300 dark:hover:border-rose-300/20 dark:hover:bg-rose-900/20 dark:hover:text-rose-300"
        @click="inv.removeLine(line.sortOrder)"
      >
        <TrashIcon class="size-5" />
      </button>
    </div>
  </div>
</template>

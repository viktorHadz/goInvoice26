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
  <div class="grid grid-cols-10 items-center gap-6 px-2 py-3">
    <!-- name -->
    <div class="col-span-5 min-w-0">
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

    <!-- qty -->
    <div class="text-right">
      <TheInput
        type="number"
        :modelValue="line.quantity"
        @update:modelValue="setQty"
        inputClass="py-1"
      />
    </div>

    <!-- minutes -->
    <div class="text-right">
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
    <div class="text-right">
      <TheInput
        type="number"
        :modelValue="unitPounds"
        inputClass="py-1"
        @update:modelValue="setUnitPounds"
        :title="line.pricingMode === 'hourly' ? 'Hourly rate (£)' : 'Unit price (£)'"
      />
      <div class="text-sm text-zinc-500 dark:text-zinc-400">
        {{ fmtGBPMinor(line.unitPriceMinor) }}{{ line.pricingMode === 'hourly' ? '/hr' : '' }}
      </div>
    </div>

    <!-- total -->
    <div class="text-right text-base font-semibold text-zinc-900 dark:text-zinc-100">
      {{ fmtGBPMinor(totalMinor) }}
    </div>

    <!-- remove -->
    <div class="flex justify-end">
      <TheButton
        variant="danger"
        title="Remove line"
        class="cursor-pointer"
        @click="inv.removeLine(line.sortOrder)"
      >
        <TrashIcon class="size-4" />
      </TheButton>
    </div>
  </div>
</template>

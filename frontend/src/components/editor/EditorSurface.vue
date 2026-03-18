<script setup lang="ts">
import { computed } from 'vue'
import { useEditorStore } from '@/stores/editor'
import { fmtGBPMinor, calcTotals, calcDepositMinor, calcBalanceDueMinor } from '@/utils/money'
import { fmtStrDate } from '@/utils/dates'
import { useSettingsStore } from '@/stores/settings'

const editStore = useEditorStore()
const setStore = useSettingsStore()

const inv = computed(() => editStore.fmtActiveInv)

const totals = computed(() => {
  if (!inv.value) return null
  return calcTotals(inv.value)
})

const depositMinor = computed(() => {
  if (!inv.value || !totals.value) return 0
  return calcDepositMinor(inv.value, totals.value.totalMinor)
})

const balanceDueMinor = computed(() => {
  if (!inv.value || !totals.value) return 0
  return calcBalanceDueMinor(totals.value.totalMinor, depositMinor.value, inv.value.paidMinor)
})
</script>

<template>
  <!-- Loading -->
  <div
    v-if="editStore.isLoadingInvoice"
    class="flex min-h-60 items-center justify-center"
  >
    <div class="text-sm text-zinc-500 dark:text-zinc-400">Loading invoice...</div>
  </div>

  <!-- No data -->
  <div
    v-else-if="!inv"
    class="flex min-h-60 items-center justify-center"
  >
    <div class="text-sm text-zinc-500 dark:text-zinc-400">No invoice data available.</div>
  </div>

  <!-- Invoice preview -->
  <div
    v-else
    class="space-y-5"
  >
    <!-- Dates -->
    <div class="flex items-start justify-between gap-4">
      <div>
        <div class="text-xs text-zinc-500 dark:text-zinc-400">Issue date:</div>
        <div class="text-sm font-medium text-zinc-900 dark:text-zinc-100">
          {{ fmtStrDate(inv.issueDate) }}
        </div>
      </div>
      <div class="text-right">
        <div class="text-xs text-zinc-500 dark:text-zinc-400">Due date:</div>
        <div class="text-sm font-medium text-zinc-900 dark:text-zinc-100">
          {{ inv.dueByDate ? fmtStrDate(inv.dueByDate) : 'N/A' }}
        </div>
      </div>
    </div>

    <!-- Client snapshot -->
    <div class="flex flex-col justify-between gap-6 sm:flex-row">
      <div
        class="w-full rounded-xl border border-zinc-200 bg-zinc-50 p-3 dark:border-zinc-800 dark:bg-zinc-900/40"
      >
        <div class="mb-1.5 text-xs font-medium text-sky-600 dark:text-emerald-400">From</div>
        <div
          v-if="setStore.settings?.companyName"
          class="text-sm font-medium text-zinc-900 dark:text-zinc-100"
        >
          {{ setStore.settings?.companyName }}
        </div>
        <div
          v-if="setStore.settings?.companyAddress"
          class="mt-1 text-xs text-zinc-500 dark:text-zinc-400"
        >
          {{ setStore.settings?.companyAddress }}
        </div>
        <div
          v-if="setStore.settings?.email"
          class="text-xs text-zinc-500 dark:text-zinc-400"
        >
          {{ setStore.settings?.email }}
        </div>
      </div>
      <div
        class="w-full rounded-xl border border-zinc-200 bg-zinc-50 p-3 dark:border-zinc-800 dark:bg-zinc-900/40"
      >
        <div class="mb-1.5 text-xs font-medium text-sky-600 dark:text-emerald-400">Bill to</div>
        <div class="text-sm font-medium text-zinc-900 dark:text-zinc-100">
          {{ inv.clientSnapshot.name }}
        </div>
        <div
          v-if="inv.clientSnapshot.companyName"
          class="text-sm text-zinc-600 dark:text-zinc-400"
        >
          {{ inv.clientSnapshot.companyName }}
        </div>
        <div
          v-if="inv.clientSnapshot.address"
          class="mt-1 text-xs text-zinc-500 dark:text-zinc-400"
        >
          {{ inv.clientSnapshot.address }}
        </div>
        <div
          v-if="inv.clientSnapshot.email"
          class="text-xs text-zinc-500 dark:text-zinc-400"
        >
          {{ inv.clientSnapshot.email }}
        </div>
      </div>
    </div>

    <!-- Line items -->
    <div class="overflow-x-auto">
      <div class="min-w-150">
        <!-- Header -->
        <div
          class="grid grid-cols-[1fr_72px_56px_96px_96px] gap-x-2 px-1 pb-2 text-sm font-semibold tracking-wider text-zinc-600 capitalize dark:text-zinc-200"
        >
          <div>Item</div>
          <div>Type</div>
          <div class="text-right">Qty</div>
          <div class="text-right">Price</div>
          <div class="text-right">Total</div>
        </div>
        <hr class="text-zinc-200 dark:text-zinc-800" />
        <!-- Rows -->
        <div class="divide-y divide-zinc-200 dark:divide-zinc-800">
          <div
            v-for="(line, i) in inv.lines"
            :key="i"
            class="grid grid-cols-[1fr_72px_56px_96px_96px] items-center gap-x-2 px-1 py-2.5 text-sm"
          >
            <div class="min-w-0">
              <div class="truncate text-zinc-900 dark:text-zinc-100">{{ line.name }}</div>
              <div
                v-if="line.pricingMode === 'hourly' && line.minutesWorked"
                class="mt-0.5 text-xs text-zinc-400 dark:text-zinc-500"
              >
                {{ line.minutesWorked }} min
              </div>
            </div>
            <div class="truncate text-xs text-zinc-500 dark:text-zinc-400">{{ line.lineType }}</div>
            <div class="text-right text-zinc-700 tabular-nums dark:text-zinc-300">
              {{ line.quantity }}
            </div>
            <div class="text-right text-zinc-700 tabular-nums dark:text-zinc-300">
              {{ fmtGBPMinor(line.unitPriceMinor) }}
            </div>
            <div class="text-right font-medium text-zinc-900 tabular-nums dark:text-zinc-100">
              {{ fmtGBPMinor(line.quantity * line.unitPriceMinor) }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Totals -->
    <div
      v-if="totals"
      class="space-y-3 text-sm"
    >
      <hr class="text-zinc-200 dark:text-zinc-800" />

      <div class="flex items-center justify-between gap-3">
        <div class="text-zinc-600 dark:text-zinc-400">Subtotal</div>
        <div class="font-semibold text-zinc-900 tabular-nums dark:text-zinc-100">
          {{ fmtGBPMinor(totals.subtotalMinor) }}
        </div>
      </div>

      <div
        v-if="inv.discountType !== 'none'"
        class="flex items-center justify-between gap-3"
      >
        <div class="text-zinc-600 dark:text-zinc-400">Discount</div>
        <div class="font-semibold text-zinc-900 tabular-nums dark:text-zinc-100">
          -{{ fmtGBPMinor(totals.discountMinor) }}
        </div>
      </div>

      <div
        v-if="inv.vatRate > 0"
        class="flex items-center justify-between gap-3"
      >
        <div class="text-zinc-600 dark:text-zinc-400">VAT</div>
        <div class="font-semibold text-zinc-900 tabular-nums dark:text-zinc-100">
          {{ fmtGBPMinor(totals.vatMinor) }}
        </div>
      </div>

      <hr class="text-zinc-200 dark:text-zinc-800" />

      <div class="flex items-center justify-between gap-3 text-base">
        <div class="font-semibold text-zinc-800 dark:text-zinc-100">Total</div>
        <div class="font-semibold text-zinc-900 tabular-nums dark:text-zinc-100">
          {{ fmtGBPMinor(totals.totalMinor) }}
        </div>
      </div>

      <!-- Deposit / Paid / Balance due -->
      <div
        v-if="inv.depositType !== 'none' || inv.paidMinor > 0"
        class="rounded-xl border border-zinc-200 bg-zinc-50 p-3 dark:border-zinc-800 dark:bg-zinc-900/40"
      >
        <div
          v-if="inv.depositType !== 'none'"
          class="flex items-center justify-between gap-3"
        >
          <div class="text-zinc-600 dark:text-zinc-400">Deposit</div>
          <div class="font-semibold text-zinc-900 tabular-nums dark:text-zinc-100">
            -{{ fmtGBPMinor(depositMinor) }}
          </div>
        </div>

        <div
          v-if="inv.paidMinor > 0"
          class="mt-2 flex items-center justify-between gap-3"
        >
          <div class="text-zinc-600 dark:text-zinc-400">Paid</div>
          <div class="font-semibold text-zinc-900 tabular-nums dark:text-zinc-100">
            -{{ fmtGBPMinor(inv.paidMinor) }}
          </div>
        </div>

        <div class="mt-3 flex items-center justify-between gap-3 text-base">
          <div class="font-semibold text-zinc-800 dark:text-zinc-100">Balance due</div>
          <div class="font-semibold text-zinc-900 tabular-nums dark:text-zinc-100">
            {{ fmtGBPMinor(balanceDueMinor) }}
          </div>
        </div>
      </div>
    </div>

    <!-- Note -->
    <div
      v-if="inv.note"
      class="rounded-xl border border-zinc-200 bg-zinc-50 p-3 dark:border-zinc-800 dark:bg-zinc-900/40"
    >
      <div class="mb-1 text-xs font-medium text-zinc-500 dark:text-zinc-400">Note</div>
      <div class="text-sm text-zinc-700 italic dark:text-zinc-300">{{ inv.note }}</div>
    </div>
  </div>
</template>

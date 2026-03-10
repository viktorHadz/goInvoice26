<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import TheInput from '@/components/UI/TheInput.vue'
import TheButton from '@/components/UI/TheButton.vue'
import TheDropdown from '@/components/UI/TheDropdown.vue'
import { useInvoiceStore } from '@/stores/invoice'
import TheTooltip from '../UI/TheTooltip.vue'
import { InformationCircleIcon } from '@heroicons/vue/24/outline'

const inv = useInvoiceStore()

const paid = ref<number | null>(0)

const depositMode = ref<'none' | 'fixed' | 'percent'>('none')
const deposit = ref<number | null>(0)

const discountMode = ref<'none' | 'fixed' | 'percent'>('none')
const discount = ref<number | null>(0)

const vatPercent = ref<number | null>(20)
const noteTouched = ref(false)

const noteProxy = computed<string>({
  get: () => inv.invoice?.note ?? '',
  set: (v) => inv.setNote(String(v ?? '')),
})

const depositError = computed(() => {
  if (depositMode.value === 'percent') return inv.getFieldError('totals.depositRate')
  return inv.getFieldError('totals.depositMinor')
})

const discountError = computed(() => {
  if (discountMode.value === 'percent') return inv.getFieldError('totals.discountRate')
  return inv.getFieldError('totals.discountMinor')
})

function toNum(v: number | null | undefined) {
  return Number(v || 0)
}

function clamp(n: number, min: number, max: number) {
  return Math.min(max, Math.max(min, n))
}

function syncFromInvoice() {
  const v = inv.invoice
  if (!v) return

  paid.value = inv.fromMinor(v.paidMinor)

  depositMode.value = v.depositType
  deposit.value =
    v.depositType === 'fixed'
      ? inv.fromMinor(v.depositMinor)
      : v.depositType === 'percent'
        ? v.depositRate / 100
        : 0

  discountMode.value = v.discountType
  discount.value =
    v.discountType === 'fixed'
      ? inv.fromMinor(v.discountMinor)
      : v.discountType === 'percent'
        ? v.discountRate / 100
        : 0

  vatPercent.value = (v.vatRate ?? 2000) / 100
  noteTouched.value = false
}

watch(
  () => inv.invoice,
  () => syncFromInvoice(),
  { immediate: true },
)

function applyPaid() {
  if (!inv.invoice) return

  const gbp = Math.max(0, toNum(paid.value))
  inv.setPaidGBP(gbp)
  paid.value = gbp
}

function applyDeposit() {
  if (!inv.invoice) return

  if (depositMode.value === 'none') {
    inv.clearDeposit()
    deposit.value = 0
    return
  }

  if (depositMode.value === 'fixed') {
    const gbp = Math.max(0, toNum(deposit.value))
    inv.setDepositFixedGBP(gbp)
    deposit.value = gbp
    return
  }

  const percent = clamp(toNum(deposit.value), 0, 100)
  inv.setDepositPercent(percent)
  deposit.value = percent
}

function applyDiscount() {
  if (!inv.invoice) return

  if (discountMode.value === 'none') {
    inv.clearDiscount()
    discount.value = 0
    return
  }

  if (discountMode.value === 'fixed') {
    const gbp = Math.max(0, toNum(discount.value))
    inv.setDiscountFixedGBP(gbp)
    discount.value = gbp
    return
  }

  const percent = clamp(toNum(discount.value), 0, 100)
  inv.setDiscountPercent(percent)
  discount.value = percent
}

function applyVat() {
  if (!inv.invoice) return

  const percent = clamp(toNum(vatPercent.value), 0, 100)
  inv.setVatRateBps(Math.round(percent * 100))
  vatPercent.value = percent
}
</script>

<template>
  <div class="min-w-0 space-y-5">
    <!-- Discount -->
    <div class="min-w-0 space-y-2">
      <div class="text-sm font-semibold text-zinc-700 dark:text-zinc-200">Discount</div>

      <div class="grid min-w-0 grid-cols-1 items-end gap-2 sm:grid-cols-[minmax(0,1fr)_6rem_auto]">
        <TheInput
          v-model="discount"
          type="number"
          placeholder="0"
          labelHidden
          :reserveErrorSpace="false"
          :disabled="discountMode === 'none'"
          :title="
            discountMode === 'none' ? 'select discount mode from dropdown first' : 'discount value'
          "
          inputClass="w-full py-1.5"
          :error="discountError"
        />
        <div class="min-w-0 sm:w-24">
          <TheDropdown
            v-model="discountMode"
            input-class="py-1.5"
            :options="['none', 'fixed', 'percent']"
          />
        </div>
        <TheButton
          class="w-full sm:w-auto"
          @click="applyDiscount"
        >
          Apply
        </TheButton>
      </div>
    </div>

    <!-- Deposit -->
    <div class="min-w-0 space-y-2">
      <div class="text-sm font-semibold text-zinc-700 dark:text-zinc-200">Deposit</div>

      <div class="grid min-w-0 grid-cols-1 items-end gap-2 sm:grid-cols-[minmax(0,1fr)_6rem_auto]">
        <TheInput
          v-model="deposit"
          type="number"
          :placeholder="depositMode === 'percent' ? '10' : '0'"
          labelHidden
          :reserveErrorSpace="false"
          :disabled="depositMode === 'none'"
          :title="
            depositMode === 'none' ? 'select deposit mode from dropdown first' : 'deposit value'
          "
          inputClass="w-full py-1.5"
          :error="depositError"
        />
        <div class="min-w-0 sm:w-24">
          <TheDropdown
            v-model="depositMode"
            input-class="py-1.5"
            :options="['none', 'fixed', 'percent']"
          />
        </div>
        <TheButton
          class="w-full sm:w-auto"
          @click="applyDeposit"
        >
          Apply
        </TheButton>
      </div>
    </div>

    <!-- Paid -->
    <div class="min-w-0 space-y-2">
      <div
        class="flex w-full justify-between text-sm font-semibold text-zinc-700 dark:text-zinc-200"
      >
        <p>Amount paid (£)</p>
        <TheTooltip
          :icon="InformationCircleIcon"
          text="For invoice payments made before the deposit"
          side="top"
          max-width-class="w-42"
          align="center"
        />
      </div>

      <div class="grid min-w-0 grid-cols-1 items-start gap-2 sm:grid-cols-[minmax(0,1fr)_auto]">
        <TheInput
          v-model="paid"
          type="number"
          placeholder="Amount paid (£)"
          labelHidden
          :reserveErrorSpace="false"
          inputClass="w-full py-1.5"
          :error="inv.getFieldError('totals.paidMinor')"
        />
        <TheButton
          class="w-full sm:w-auto"
          @click="applyPaid"
        >
          Apply
        </TheButton>
      </div>
    </div>

    <!-- VAT -->
    <div class="min-w-0 space-y-2">
      <div class="text-sm font-semibold text-zinc-700 dark:text-zinc-200">VAT Rate (%)</div>

      <div class="grid min-w-0 grid-cols-1 items-end gap-2 sm:grid-cols-[minmax(0,1fr)_auto]">
        <TheInput
          v-model="vatPercent"
          type="number"
          placeholder="20"
          labelHidden
          :reserveErrorSpace="false"
          inputClass="w-full py-1.5"
          :error="inv.getFieldError('totals.vatRate')"
        />
        <TheButton
          class="w-full sm:w-auto"
          @click="applyVat"
        >
          Apply
        </TheButton>
      </div>
    </div>

    <!-- Note -->
    <div class="min-w-0 space-y-2">
      <div class="text-sm font-semibold text-zinc-700 dark:text-zinc-200">Note</div>

      <textarea
        id="invoice-adjustments-text-area"
        v-model="noteProxy"
        class="input w-full border border-zinc-200 bg-white p-3 text-sm dark:border-zinc-800 dark:bg-zinc-950/40"
        :disabled="!inv.invoice"
        placeholder="Add a note to show on the invoice…"
        @blur.stop="noteTouched = true"
      />

      <p
        class="mt-1 min-h-5 text-xs"
        :class="
          inv.getFieldError('note') && (noteTouched || inv.showAllValidation)
            ? 'text-rose-600 dark:text-rose-300'
            : 'text-transparent'
        "
      >
        {{
          inv.getFieldError('note') && (noteTouched || inv.showAllValidation)
            ? inv.getFieldError('note')
            : '•'
        }}
      </p>
    </div>
  </div>
</template>

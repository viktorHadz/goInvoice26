<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import TheInput from '@/components/UI/TheInput.vue'
import TheButton from '@/components/UI/TheButton.vue'
import TheDropdown from '@/components/UI/TheDropdown.vue'
import { useInvoiceStore } from '@/stores/invoice'
import TheTooltip from '../UI/TheTooltip.vue'
import { InformationCircleIcon } from '@heroicons/vue/24/outline'
import { fromMinor } from '@/utils/money'

const inv = useInvoiceStore()

const paid = ref<number | null>(0)

const depositMode = ref<'none' | 'fixed' | 'percent'>('none')
const deposit = ref<number | null>(0)

const discountMode = ref<'none' | 'fixed' | 'percent'>('none')
const discount = ref<number | null>(0)

const vatPercent = ref<number | null>(20)
const noteTouched = ref(false)

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

  paid.value = fromMinor(v.paidMinor)

  depositMode.value = v.depositType
  deposit.value =
    v.depositType === 'fixed'
      ? fromMinor(v.depositMinor)
      : v.depositType === 'percent'
        ? v.depositRate / 100
        : 0

  discountMode.value = v.discountType
  discount.value =
    v.discountType === 'fixed'
      ? fromMinor(v.discountMinor)
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
  <div class="min-w-0 divide-y divide-zinc-200 dark:divide-zinc-800">
    <!-- Discount -->
    <section class="min-w-0 py-3.5 first:pt-0">
      <div class="mb-2 flex min-w-0 items-center justify-between gap-3">
        <div class="text-sm font-medium text-zinc-800 dark:text-zinc-100">Discount</div>
        <TheTooltip
          side="top"
          align="center"
          class="hover:text-sky-400 dark:hover:text-emerald-400"
          text="Reduce the invoice by a fixed amount or a percentage. Applied before VAT"
        >
          <InformationCircleIcon class="size-5" />
        </TheTooltip>
      </div>

      <div
        class="grid min-w-0 grid-cols-1 items-center gap-2 sm:grid-cols-[minmax(0,1fr)_6.5rem_5.5rem]"
      >
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
          inputClass="w-full py-1"
          :error="discountError"
        />

        <TheDropdown
          v-model="discountMode"
          input-class="py-1"
          :options="['none', 'fixed', 'percent']"
        />

        <TheButton
          class="w-full py-1.5!"
          @click="applyDiscount"
        >
          Apply
        </TheButton>
      </div>
    </section>

    <!-- Deposit -->
    <section class="min-w-0 py-3.5">
      <div class="mb-2 flex min-w-0 items-center justify-between gap-3">
        <div class="text-sm font-medium text-zinc-800 dark:text-zinc-100">Deposit</div>
        <TheTooltip
          text="Take payment upfront as a fixed amount or percentage. Applied after VAT."
          class="hover:text-sky-400 dark:hover:text-emerald-400"
        >
          <InformationCircleIcon class="size-5" />
        </TheTooltip>
      </div>

      <div
        class="grid min-w-0 grid-cols-1 items-center gap-2 sm:grid-cols-[minmax(0,1fr)_6.5rem_5.5rem]"
      >
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
          inputClass="w-full py-1"
          :error="depositError"
        />

        <TheDropdown
          v-model="depositMode"
          input-class="py-1"
          :options="['none', 'fixed', 'percent']"
        />

        <TheButton
          class="w-full py-1.5!"
          @click="applyDeposit"
        >
          Apply
        </TheButton>
      </div>
    </section>

    <!-- Paid -->
    <section class="min-w-0 py-3.5">
      <div class="mb-2 flex min-w-0 items-center justify-between gap-3">
        <div class="text-sm font-medium text-zinc-800 dark:text-zinc-100">Amount paid</div>
        <TheTooltip
          text="Payment already received before the final balance."
          side="top"
          align="center"
          class="hover:text-sky-400 dark:hover:text-emerald-400"
        >
          <InformationCircleIcon class="size-5" />
        </TheTooltip>
      </div>

      <div class="grid min-w-0 grid-cols-1 gap-2 sm:grid-cols-[minmax(0,1fr)_5.5rem]">
        <TheInput
          v-model="paid"
          type="number"
          placeholder="0"
          labelHidden
          :reserveErrorSpace="false"
          inputClass="w-full py-1"
          :error="inv.getFieldError('totals.paidMinor')"
        />

        <TheButton
          class="w-full py-1.5!"
          @click="applyPaid"
        >
          Apply
        </TheButton>
      </div>
    </section>

    <!-- VAT -->
    <section class="min-w-0 py-3.5">
      <div class="mb-2 flex min-w-0 items-center justify-between gap-3">
        <div class="text-sm font-medium text-zinc-800 dark:text-zinc-100">VAT rate</div>
        <TheTooltip
          text="Set to 0% to exclude VAT from the invoice."
          side="top"
          align="center"
          class="hover:text-sky-400 dark:hover:text-emerald-400"
        >
          <InformationCircleIcon class="size-5" />
        </TheTooltip>
      </div>

      <div class="grid min-w-0 grid-cols-1 gap-2 sm:grid-cols-[minmax(0,1fr)_5.5rem]">
        <TheInput
          v-model="vatPercent"
          type="number"
          placeholder="20"
          labelHidden
          :reserveErrorSpace="false"
          inputClass="w-full py-1"
          :error="inv.getFieldError('totals.vatRate')"
        />

        <TheButton
          class="w-full py-1.5!"
          @click="applyVat"
        >
          Apply
        </TheButton>
      </div>
    </section>
  </div>
</template>

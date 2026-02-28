<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import TheInput from '@/components/UI/TheInput.vue'
import TheButton from '@/components/UI/TheButton.vue'
import TheDropdown from '@/components/UI/TheDropdown.vue'
import { useInvoiceDraftStore } from '@/stores/invoiceDraft'
import { toMinor, fromMinor } from '@/utils/money'

const inv = useInvoiceDraftStore()
const draft = computed(() => inv.draft)

const paid = ref<number | null>(0)
const depositMode = ref<'fixed' | 'percent'>('fixed')
const deposit = ref<number | null>(0)

const discountMode = ref<'none' | 'fixed' | 'percent'>('none')
const discount = ref<number | null>(0)

const noteProxy = computed<string>({
  get: () => inv.draft?.note ?? '',
  set: (v) => {
    if (!inv.draft) return
    inv.draft.note = String(v ?? '')
  },
})

watch(
  draft,
  (v) => {
    if (!v) return
    paid.value = fromMinor(v.paidMinor)
    deposit.value = fromMinor(v.depositMinor)
    depositMode.value = 'fixed'

    discountMode.value = v.discountType
    discount.value =
      v.discountType === 'fixed'
        ? fromMinor(v.discountValue)
        : v.discountType === 'percent'
          ? v.discountValue / 100
          : 0
  },
  { immediate: true },
)

function applyPaid() {
  if (!inv.draft) return
  inv.draft.paidMinor = toMinor(Number(paid.value || 0))
}

function applyDeposit() {
  if (!inv.draft || !inv.totals) return
  if (depositMode.value === 'fixed') {
    inv.draft.depositMinor = toMinor(Number(deposit.value || 0))
    return
  }
  const total = inv.totals.totalMinor
  const pct = Math.max(0, Math.min(100, Number(deposit.value || 0)))
  inv.draft.depositMinor = Math.round(total * (pct / 100))
}

function applyDiscount() {
  if (!inv.draft) return

  if (discountMode.value === 'none') {
    inv.draft.discountType = 'none'
    inv.draft.discountValue = 0
    return
  }

  if (discountMode.value === 'fixed') {
    inv.draft.discountType = 'fixed'
    inv.draft.discountValue = toMinor(Number(discount.value || 0))
    return
  }

  const pct = Math.max(0, Math.min(100, Number(discount.value || 0)))
  inv.draft.discountType = 'percent'
  inv.draft.discountValue = Math.round(pct * 100) // bps
}
</script>

<template>
  <div class="min-w-0 space-y-5">
    <!-- Payments -->
    <div class="min-w-0 space-y-2">
      <div class="text-sm font-semibold text-zinc-700 dark:text-zinc-200">Ammount paid (£)</div>
      <div class="grid min-w-0 grid-cols-1 items-start gap-2 sm:grid-cols-[minmax(0,1fr)_auto]">
        <TheInput
          v-model="paid"
          type="number"
          placeholder="Amount paid (£)"
          labelHidden
          :reserveErrorSpace="false"
          inputClass="w-full py-1.5"
        />
        <TheButton
          class="w-full sm:w-auto"
          @click="applyPaid"
        >
          Apply
        </TheButton>
      </div>
      <div class="text-sm font-semibold text-zinc-700 dark:text-zinc-200">Deposit</div>
      <div class="grid min-w-0 grid-cols-1 items-end gap-2 sm:grid-cols-[minmax(0,1fr)_6rem_auto]">
        <TheInput
          v-model="deposit"
          type="number"
          placeholder="Deposit"
          labelHidden
          :reserveErrorSpace="false"
          inputClass="w-full py-1.5"
        />
        <div class="min-w-0 sm:w-24">
          <TheDropdown
            v-model="depositMode"
            :options="['fixed', 'percent']"
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
          inputClass="w-full py-1.5"
        />
        <div class="min-w-0 sm:w-24">
          <TheDropdown
            v-model="discountMode"
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

    <!-- VAT -->
    <div class="min-w-0 space-y-2">
      <div class="text-sm font-semibold text-zinc-700 dark:text-zinc-200">VAT Rate (%)</div>

      <div class="flex min-w-0 items-end gap-2">
        <TheInput
          :modelValue="(draft?.vatRate ?? 2000) / 100"
          @update:modelValue="
            (v) => {
              if (!inv.draft) return
              const pct = Math.max(0, Math.min(100, Number(v) || 0))
              inv.draft.vatRate = Math.round(pct * 100)
            }
          "
          type="number"
          placeholder="20"
          labelHidden
          inputClass="w-full sm:w-32 py-1.5"
        />
      </div>
    </div>

    <!-- Note -->
    <div class="min-w-0 space-y-2">
      <div class="text-sm font-semibold text-zinc-700 dark:text-zinc-200">Note</div>
      <textarea
        class="input w-full border border-zinc-200 bg-white p-3 text-sm dark:border-zinc-800 dark:bg-zinc-950/40"
        :value="noteProxy"
        @input="(e) => (noteProxy = (e.target as HTMLTextAreaElement).value)"
        :disabled="!inv.draft"
        placeholder="Add a note to show on the invoice…"
      />
    </div>
  </div>
</template>

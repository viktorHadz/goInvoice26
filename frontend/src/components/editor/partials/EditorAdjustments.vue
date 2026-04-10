<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import TheInput from '@/components/UI/TheInput.vue'
import TheButton from '@/components/UI/TheButton.vue'
import TheDropdown from '@/components/UI/TheDropdown.vue'
import { ChevronUpDownIcon, InformationCircleIcon } from '@heroicons/vue/24/outline'
import { fromMinor } from '@/utils/money'
import TheTooltip from '@/components/UI/TheTooltip.vue'
import { useEditorStore } from '@/stores/editor'
import { emitToastError } from '@/utils/toast'
import { cloneInvoice } from '@/utils/cloneInvoice'
import { findNewInvoiceValidationMessage } from '@/utils/invoiceValidationDiff'
import {
    clearInvoiceDeposit,
    clearInvoiceDiscount,
    setInvoiceDepositFixedGBP,
    setInvoiceDepositPercent,
    setInvoiceDiscountFixedGBP,
    setInvoiceDiscountPercent,
    setInvoiceVatRateBps,
} from '@/utils/invoiceMutations'

const editStore = useEditorStore()

const depositMode = ref<'none' | 'fixed' | 'percent'>('none')
const deposit = ref<number | null>(0)

const discountMode = ref<'none' | 'fixed' | 'percent'>('none')
const discount = ref<number | null>(0)

const vatPercent = ref<number | null>(20)

const depositError = computed(() => {
    if (depositMode.value === 'percent') return editStore.getFieldError('totals.depositRate')
    return editStore.getFieldError('totals.depositMinor')
})

const discountError = computed(() => {
    if (discountMode.value === 'percent') return editStore.getFieldError('totals.discountRate')
    return editStore.getFieldError('totals.discountMinor')
})

function toNum(v: number | null | undefined) {
    return Number(v || 0)
}

function clamp(n: number, min: number, max: number) {
    return Math.min(max, Math.max(min, n))
}

function syncFromInvoice() {
    const v = editStore.draftInvoice
    if (!v) return

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
}

watch(
    () => editStore.draftInvoice,
    () => syncFromInvoice(),
    { immediate: true },
)

function assertEditorInvoiceValidOrRollback(
    snapshot: ReturnType<typeof cloneInvoice>,
    fieldNames: string[],
    toastTitle: string,
) {
    const current = editStore.draftInvoice
    if (!current) return true
    const blockingMessage = findNewInvoiceValidationMessage(snapshot, current, fieldNames)
    if (!blockingMessage) return true

    editStore.draftInvoice = snapshot
    emitToastError({ title: toastTitle, message: blockingMessage })
    return false
}

function applyDeposit() {
    if (!editStore.draftInvoice) return
    const snapshot = cloneInvoice(editStore.draftInvoice)

    if (depositMode.value === 'none') {
        clearInvoiceDeposit(editStore.draftInvoice)
        deposit.value = 0
        assertEditorInvoiceValidOrRollback(
            snapshot,
            ['totals.depositMinor', 'totals.depositRate'],
            'Deposit reverted',
        )
        return
    }

    if (depositMode.value === 'fixed') {
        const gbp = Math.max(0, toNum(deposit.value))
        setInvoiceDepositFixedGBP(editStore.draftInvoice, gbp)
        deposit.value = gbp
        assertEditorInvoiceValidOrRollback(
            snapshot,
            ['totals.depositMinor', 'totals.depositRate'],
            'Deposit reverted',
        )
        return
    }

    const percent = clamp(toNum(deposit.value), 0, 100)
    setInvoiceDepositPercent(editStore.draftInvoice, percent)
    deposit.value = percent
    assertEditorInvoiceValidOrRollback(
        snapshot,
        ['totals.depositRate', 'totals.depositMinor'],
        'Deposit reverted',
    )
}

function applyDiscount() {
    if (!editStore.draftInvoice) return
    const snapshot = cloneInvoice(editStore.draftInvoice)

    if (discountMode.value === 'none') {
        clearInvoiceDiscount(editStore.draftInvoice)
        discount.value = 0
        assertEditorInvoiceValidOrRollback(
            snapshot,
            ['totals.discountMinor', 'totals.discountRate'],
            'Discount reverted',
        )
        return
    }

    if (discountMode.value === 'fixed') {
        const gbp = Math.max(0, toNum(discount.value))
        setInvoiceDiscountFixedGBP(editStore.draftInvoice, gbp)
        discount.value = gbp
        assertEditorInvoiceValidOrRollback(
            snapshot,
            ['totals.discountMinor', 'totals.discountRate'],
            'Discount reverted',
        )
        return
    }

    const percent = clamp(toNum(discount.value), 0, 100)
    setInvoiceDiscountPercent(editStore.draftInvoice, percent)
    discount.value = percent
    assertEditorInvoiceValidOrRollback(
        snapshot,
        ['totals.discountRate', 'totals.discountMinor'],
        'Discount reverted',
    )
}

function applyVat() {
    if (!editStore.draftInvoice) return
    const snapshot = cloneInvoice(editStore.draftInvoice)

    const percent = clamp(toNum(vatPercent.value), 0, 100)
    setInvoiceVatRateBps(editStore.draftInvoice, Math.round(percent * 100))
    vatPercent.value = percent
    assertEditorInvoiceValidOrRollback(snapshot, ['totals.vatRate'], 'VAT change reverted')
}
</script>

<template>
    <div class="space-y-2 divide-y divide-zinc-200 dark:divide-zinc-800">
        <section class="py-4 first:pt-0">
            <div class="mb-2 flex items-center justify-between gap-3">
                <div class="text-sm font-medium text-zinc-800 dark:text-zinc-100">Discount</div>
                <TheTooltip
                    side="top"
                    align="center"
                    class="hover:text-sky-600 dark:hover:text-emerald-400"
                    text="Reduce the invoice by a fixed amount or a percentage. Applied before VAT"
                >
                    <InformationCircleIcon class="size-5 cursor-help" />
                </TheTooltip>
            </div>

            <div class="flex flex-col gap-2 sm:flex-row">
                <div class="flex justify-between gap-2">
                    <TheInput
                        v-model="discount"
                        type="number"
                        placeholder="0"
                        labelHidden
                        :reserveErrorSpace="false"
                        :disabled="discountMode === 'none'"
                        :title="
                            discountMode === 'none'
                                ? 'select discount mode from dropdown first'
                                : 'discount value'
                        "
                        inputClass="w-full py-1"
                        :error="discountError"
                    />
                    <TheDropdown
                        v-model="discountMode"
                        input-class="py-1 min-w-32 sm:min-w-24"
                        :right-icon="ChevronUpDownIcon"
                        :options="['none', 'fixed', 'percent']"
                    />
                </div>

                <TheButton class="w-full sm:flex-1 py-1.5!" @click="applyDiscount">
                    Apply
                </TheButton>
            </div>
        </section>

        <section class="py-4">
            <div class="mb-2 flex items-center justify-between gap-3">
                <div class="text-sm font-medium text-zinc-800 dark:text-zinc-100">Deposit</div>
                <TheTooltip
                    text="Show the amount you want upfront. Deposits stay visible on the invoice but do not reduce the saved balance due."
                    class="hover:text-sky-600 dark:hover:text-emerald-400"
                >
                    <InformationCircleIcon class="size-5 cursor-help" />
                </TheTooltip>
            </div>

            <div class="flex flex-col gap-2 sm:flex-row">
                <div class="flex justify-between gap-2">
                    <TheInput
                        v-model="deposit"
                        type="number"
                        :placeholder="depositMode === 'percent' ? '10' : '0'"
                        labelHidden
                        :reserveErrorSpace="false"
                        :disabled="depositMode === 'none'"
                        :title="
                            depositMode === 'none'
                                ? 'select deposit mode from dropdown first'
                                : 'deposit value'
                        "
                        inputClass="w-full py-1"
                        :error="depositError"
                    />

                    <TheDropdown
                        v-model="depositMode"
                        input-class="py-1 min-w-32 sm:min-w-24"
                        :right-icon="ChevronUpDownIcon"
                        :options="['none', 'fixed', 'percent']"
                    />
                </div>

                <TheButton class="w-full sm:flex-1 py-1.5!" @click="applyDeposit">
                    Apply
                </TheButton>
            </div>
        </section>

        <section class="min-w-0 py-4">
            <div class="mb-2 flex min-w-0 items-center justify-between gap-3">
                <div class="text-sm font-medium text-zinc-800 dark:text-zinc-100">VAT rate</div>
                <TheTooltip
                    text="Set to 0% to exclude VAT from the invoice."
                    side="top"
                    align="center"
                    class="hover:text-sky-600 dark:hover:text-emerald-400"
                >
                    <InformationCircleIcon class="size-5 cursor-help" />
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
                    :error="editStore.getFieldError('totals.vatRate')"
                />

                <TheButton class="w-full py-1.5!" @click="applyVat">
                    Apply
                </TheButton>
            </div>
        </section>
    </div>
</template>

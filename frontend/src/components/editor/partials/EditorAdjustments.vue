<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import TheInput from '@/components/UI/TheInput.vue'
import TheButton from '@/components/UI/TheButton.vue'
import TheDropdown from '@/components/UI/TheDropdown.vue'
import { InformationCircleIcon } from '@heroicons/vue/24/outline'
import { fromMinor, toMinor, fmtGBPMinor } from '@/utils/money'
import TheTooltip from '@/components/UI/TheTooltip.vue'
import { useEditorStore } from '@/stores/editor'
import { useSettingsStore } from '@/stores/settings'
import { fmtStrDate, todayISO } from '@/utils/dates'
import { emitToastError } from '@/utils/toast'
import { cloneInvoice } from '@/utils/cloneInvoice'
import { validateInvoicePayload } from '@/utils/frontendValidation'
import { apiDTO } from '@/utils/invoiceDto'
import {
    clearInvoiceDeposit,
    clearInvoiceDiscount,
    setInvoiceDepositFixedGBP,
    setInvoiceDepositPercent,
    setInvoiceDiscountFixedGBP,
    setInvoiceDiscountPercent,
    setInvoiceVatRateBps,
} from '@/utils/invoiceMutations'
import DateField from '@/components/invoice/DateField.vue'

const editStore = useEditorStore()
const settingsStore = useSettingsStore()
const dateFormat = computed(() => settingsStore.settings?.dateFormat ?? 'dd/mm/yyyy')

const paymentAmount = ref<number | null>(0)
const paymentDate = ref(todayISO())
const paymentError = ref('')

const depositMode = ref<'none' | 'fixed' | 'percent'>('none')
const deposit = ref<number | null>(0)

const discountMode = ref<'none' | 'fixed' | 'percent'>('none')
const discount = ref<number | null>(0)

const vatPercent = ref<number | null>(20)
const noteTouched = ref(false)

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
    noteTouched.value = false
    paymentAmount.value = 0
    paymentDate.value = todayISO()
    paymentError.value = ''
}

watch(
    () => editStore.draftInvoice,
    () => syncFromInvoice(),
    { immediate: true },
)

const canAddPayment = computed(() => editStore.balanceDueMinor > 0)

function addPendingPayment() {
    paymentError.value = ''
    if (!editStore.draftInvoice) return
    if (!canAddPayment.value) {
        paymentError.value = 'No balance due to apply.'
        paymentAmount.value = 0
        emitToastError({ title: 'Invalid payment', message: paymentError.value })
        return
    }

    const gbp = Math.max(0, toNum(paymentAmount.value))
    const minor = toMinor(gbp)
    if (minor <= 0) {
        paymentError.value = 'Enter an amount greater than zero.'
        paymentAmount.value = 0
        return
    }
    if (minor > editStore.balanceDueMinor) {
        paymentError.value = 'Amount cannot exceed outstanding balance.'
        paymentAmount.value = 0
        return
    }
    if (!paymentDate.value) {
        paymentError.value = 'Payment date is required.'
        paymentAmount.value = 0
        return
    }

    editStore.stagePendingPayment({
        amountMinor: minor,
        paymentDate: paymentDate.value,
    })
    paymentAmount.value = 0
}

function removePendingPayment(tempId: string) {
    editStore.removePendingPayment(tempId)
}

function assertEditorInvoiceValidOrRollback(
    snapshot: ReturnType<typeof cloneInvoice>,
    fieldNames: string[],
    toastTitle: string,
) {
    const current = editStore.draftInvoice
    if (!current) return true
    const dto = apiDTO(
        current,
        editStore.pendingPayments.map((p) => ({
            amountMinor: p.amountMinor,
            paymentDate: p.paymentDate,
            ...(p.label ? { label: p.label } : {}),
        })),
    )
    const errors = validateInvoicePayload(dto)
    if (Object.keys(errors).length === 0) return true

    editStore.draftInvoice = snapshot
    const message =
        fieldNames.map((f) => errors[f]).find((msg) => Boolean(msg)) ??
        errors['totals.paidMinor'] ??
        'This change would make totals invalid and has been reverted.'
    emitToastError({ title: toastTitle, message })
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
    <main class="min-w-0 divide-y divide-zinc-200 text-sm dark:divide-zinc-800">
        <!-- Discount -->
        <section class="min-w-0 py-2 first:pt-0">
            <div class="mb-2 flex min-w-0 items-center justify-between gap-3">
                <div class="font-medium text-zinc-800 dark:text-zinc-100">Discount</div>
                <TheTooltip
                    side="top"
                    align="center"
                    class="hover:text-sky-600 dark:hover:text-emerald-400"
                    text="Reduce the invoice by a fixed amount or a percentage. Applied before VAT"
                >
                    <InformationCircleIcon class="size-5 cursor-help" />
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
                        discountMode === 'none'
                            ? 'select discount mode from dropdown first'
                            : 'discount value'
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
        <section class="min-w-0 py-3">
            <div class="mb-2 flex min-w-0 items-center justify-between gap-3">
                <div class="font-medium text-zinc-800 dark:text-zinc-100">Deposit</div>
                <TheTooltip
                    text="Take payment upfront as a fixed amount or percentage. Applied after VAT."
                    class="hover:text-sky-600 dark:hover:text-emerald-400"
                >
                    <InformationCircleIcon class="size-5 cursor-help" />
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
                        depositMode === 'none'
                            ? 'select deposit mode from dropdown first'
                            : 'deposit value'
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
        <section class="min-w-0 py-3">
            <div class="mb-2 flex min-w-0 items-center justify-between gap-3">
                <div class="font-medium text-zinc-800 dark:text-zinc-100">Payments</div>
                <TheTooltip
                    text="Payments become visible only after saving this revision."
                    side="top"
                    align="center"
                    class="hover:text-sky-600 dark:hover:text-emerald-400"
                >
                    <InformationCircleIcon class="size-5 cursor-help" />
                </TheTooltip>
            </div>

            <div class="space-y-2">
                <div
                    v-if="editStore.existingAppliedPayments.length === 0"
                    class="text-xs text-zinc-600 dark:text-zinc-400"
                >
                    No saved payments on this revision.
                </div>
                <div
                    v-for="payment in editStore.existingAppliedPayments"
                    :key="payment.id"
                    class="flex items-center justify-between rounded-md border border-zinc-300 px-2 py-1.5 text-xs dark:border-zinc-800"
                >
                    <div class="text-zinc-600 dark:text-zinc-300">
                        {{ fmtStrDate(payment.paymentDate, dateFormat) }}
                    </div>
                    <div class="font-medium text-zinc-800 tabular-nums dark:text-zinc-100">
                        {{ fmtGBPMinor(payment.amountMinor) }}
                    </div>
                </div>
            </div>

            <div class="mt-3 space-y-2">
                <div
                    v-for="payment in editStore.pendingPayments"
                    :key="payment.tempId"
                    class="flex items-center justify-between rounded-md border border-dashed border-sky-300 bg-sky-50/60 px-2 py-1.5 text-xs dark:border-emerald-700 dark:bg-emerald-950/20"
                >
                    <div class="text-zinc-700 dark:text-zinc-200">
                        Pending · {{ fmtStrDate(payment.paymentDate, dateFormat) }}
                    </div>
                    <div class="flex items-center gap-2">
                        <span class="font-medium text-zinc-800 tabular-nums dark:text-zinc-100">
                            {{ fmtGBPMinor(payment.amountMinor) }}
                        </span>
                        <button
                            type="button"
                            class="cursor-pointer text-zinc-600 hover:text-rose-600 dark:text-zinc-400 dark:hover:text-rose-400"
                            @click="removePendingPayment(payment.tempId)"
                        >
                            Remove
                        </button>
                    </div>
                </div>
            </div>

            <div
                class="mt-3 grid min-w-0 grid-cols-1 gap-2 sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_5.5rem]"
            >
                <TheInput
                    v-model="paymentAmount"
                    type="number"
                    placeholder="0"
                    labelHidden
                    :reserveErrorSpace="false"
                    :disabled="!canAddPayment"
                    inputClass="w-full py-1"
                />
                <DateField v-model="paymentDate" />

                <div>
                    <TheButton
                        class="w-full py-1.5!"
                        :disabled="!canAddPayment"
                        @click="addPendingPayment"
                    >
                        Add
                    </TheButton>
                </div>
            </div>
            <p
                v-if="paymentError || editStore.getFieldError('totals.paidMinor')"
                class="mt-1 text-xs text-rose-600 dark:text-rose-400"
            >
                {{ paymentError || editStore.getFieldError('totals.paidMinor') }}
            </p>
            <p class="mt-1 text-xs text-sky-600 dark:text-emerald-400">
                Outstanding: {{ fmtGBPMinor(editStore.balanceDueMinor) }}
            </p>
            <div class="mt-1 text-xs text-zinc-600 dark:text-zinc-400">
                Staged payments are saved with the next revision.
            </div>
        </section>

        <!-- VAT -->
        <section class="min-w-0 py-3">
            <div class="mb-2 flex min-w-0 items-center justify-between gap-3">
                <div class="font-medium text-zinc-800 dark:text-zinc-100">VAT rate</div>
                <TheTooltip
                    text="Set to 0% to exclude VAT from the invoice."
                    side="top"
                    align="center"
                    class="hover:text-sky-600 dark:hover:text-emerald-400"
                >
                    <InformationCircleIcon class="size-5 cursor-help" />
                </TheTooltip>
            </div>

            <div
                class="grid min-w-0 grid-cols-1 items-center gap-2 sm:grid-cols-[minmax(0,1fr)_5.5rem]"
            >
                <TheInput
                    v-model="vatPercent"
                    type="number"
                    placeholder="20"
                    labelHidden
                    :reserveErrorSpace="false"
                    inputClass="w-full py-1"
                    :error="editStore.getFieldError('totals.vatRate')"
                />

                <TheButton
                    class="w-full py-1.5!"
                    @click="applyVat"
                >
                    Apply
                </TheButton>
            </div>
        </section>
    </main>
</template>

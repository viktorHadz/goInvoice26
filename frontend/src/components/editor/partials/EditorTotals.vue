<script setup lang="ts">
import { fmtGBPMinor } from '@/utils/money'
import { useEditorStore } from '@/stores/editor'

const editStore = useEditorStore()
</script>

<template>
    <div
        v-if="!editStore.draftInvoice || !editStore.totals"
        class="text-sm text-zinc-600 dark:text-zinc-400"
    >
        No invoice loaded.
    </div>

    <div
        v-else
        class="min-w-0 space-y-4 text-sm"
    >
        <div class="grid min-w-0 grid-cols-[1fr_auto] items-center gap-3">
            <div class="min-w-0 truncate text-zinc-600 dark:text-zinc-400">Subtotal</div>
            <div
                class="shrink-0 font-semibold whitespace-nowrap text-zinc-900 tabular-nums dark:text-zinc-100"
            >
                {{ fmtGBPMinor(editStore.totals.subtotalMinor) }}
            </div>
        </div>

        <div class="grid min-w-0 grid-cols-[1fr_auto] items-center gap-3">
            <div class="min-w-0 truncate text-zinc-600 dark:text-zinc-400">Discount</div>
            <div
                class="shrink-0 font-semibold whitespace-nowrap text-zinc-900 tabular-nums dark:text-zinc-100"
            >
                -{{ fmtGBPMinor(editStore.totals.discountMinor) }}
            </div>
        </div>

        <div class="grid min-w-0 grid-cols-[1fr_auto] items-center gap-3">
            <div class="min-w-0 truncate text-zinc-600 dark:text-zinc-400">VAT</div>
            <div
                class="shrink-0 font-semibold whitespace-nowrap text-zinc-900 tabular-nums dark:text-zinc-100"
            >
                {{ fmtGBPMinor(editStore.totals.vatMinor) }}
            </div>
        </div>

        <div class="h-px bg-zinc-200 dark:bg-zinc-800" />

        <div class="grid min-w-0 grid-cols-[1fr_auto] items-center gap-3">
            <div class="min-w-0 truncate font-semibold text-zinc-800 dark:text-zinc-100">Total</div>
            <div
                class="shrink-0 font-semibold whitespace-nowrap text-zinc-900 tabular-nums dark:text-zinc-100"
            >
                {{ fmtGBPMinor(editStore.totals.totalMinor) }}
            </div>
        </div>

        <div class="grid min-w-0 grid-cols-[1fr_auto] items-center gap-3">
            <div class="min-w-0 truncate text-zinc-600 dark:text-zinc-400">Deposit</div>
            <div
                class="shrink-0 font-semibold whitespace-nowrap text-zinc-900 tabular-nums dark:text-zinc-100"
            >
                -{{ fmtGBPMinor(editStore.depositMinor) }}
            </div>
        </div>

        <div class="grid min-w-0 grid-cols-[1fr_auto] items-center gap-3">
            <div class="min-w-0 truncate text-zinc-600 dark:text-zinc-400">Paid</div>
            <div
                class="shrink-0 font-semibold whitespace-nowrap text-zinc-900 tabular-nums dark:text-zinc-100"
            >
                -{{ fmtGBPMinor(editStore.draftInvoice.paidMinor) }}
            </div>
        </div>

        <div class="rounded-xl bg-zinc-50 px-3 py-3 dark:bg-zinc-900/40">
            <div class="grid min-w-0 grid-cols-[1fr_auto] items-center gap-3">
                <div class="min-w-0 truncate font-semibold text-zinc-800 dark:text-zinc-100">
                    Balance due
                </div>
                <div
                    class="shrink-0 font-semibold whitespace-nowrap text-zinc-900 tabular-nums dark:text-zinc-100"
                >
                    {{ fmtGBPMinor(editStore.balanceDueMinor) }}
                </div>
            </div>
        </div>
    </div>
</template>

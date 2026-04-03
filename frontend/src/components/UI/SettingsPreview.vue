<script setup lang="ts">
import type { Settings } from '@/stores/settings'
import { formatInvoiceBaseLabel } from '@/utils/invoiceLabels'

defineProps<{
  form: Settings
  logoPreview?: string | null
}>()

function lineOrFallback(value: string, fallback: string) {
  return value?.trim() || fallback
}

function invoiceNumberPreview(value: number | null | undefined) {
  const num = Number(value)
  if (!Number.isFinite(num) || num < 1) return 1
  return Math.round(num)
}
</script>

<template>
  <div class="mt-4 rounded-2xl">
    <div
      class="mx-auto max-w-lg overflow-hidden rounded-2xl border border-zinc-300 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-900"
    >
      <div
        class="h-1.5 bg-[linear-gradient(to_right,rgba(14,165,233,0.85),rgba(56,189,248,0.35),transparent)] dark:bg-[linear-gradient(to_right,rgba(16,185,129,0.9),rgba(52,211,153,0.35),transparent)]"
      />

      <div class="p-5">
        <div
          class="flex items-start justify-between gap-4 border-b border-zinc-300 pb-5 dark:border-zinc-800"
        >
          <div class="min-w-0 flex-1">
            <div
              class="text-tiny font-semibold tracking-[0.18em] text-zinc-600 uppercase dark:text-zinc-400"
            >
              Invoice
            </div>

            <div class="mt-2 text-xl font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
              {{ lineOrFallback(form.companyName, 'Company name') }}
            </div>

            <div class="mt-3 space-y-1 text-xs text-zinc-600 dark:text-zinc-400">
              <div>{{ lineOrFallback(form.email, 'name@example.com') }}</div>
              <div>{{ lineOrFallback(form.phone, '+44 1234 567890') }}</div>
              <div class="max-w-[24rem] leading-relaxed whitespace-pre-line">
                {{ lineOrFallback(form.companyAddress, 'Company address goes here') }}
              </div>
              <div>Issue date: {{ lineOrFallback(form.dateFormat, 'DD/MM/YYYY') }}</div>
            </div>
          </div>

          <div class="shrink-0">
            <div
              v-if="logoPreview"
              class="flex h-16 w-24 items-center justify-center overflow-hidden"
            >
              <img
                :src="logoPreview"
                alt="Logo preview"
                class="max-h-full max-w-full object-contain"
              />
            </div>

            <div
              v-else
              class="text-tiny flex h-16 w-24 items-center justify-center rounded-xl border border-dashed border-zinc-300 bg-zinc-50 font-medium tracking-wide text-zinc-400 uppercase dark:border-zinc-700 dark:bg-zinc-950/50 dark:text-zinc-500"
            >
              Logo
            </div>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-3 border-b border-zinc-300 py-4 dark:border-zinc-800">
          <div class="rounded-xl bg-zinc-50 px-3 py-2 dark:bg-zinc-950/50">
            <div
              class="text-tiny font-semibold tracking-[0.14em] text-zinc-600 uppercase dark:text-zinc-400"
            >
              Invoice no.
            </div>
            <div class="mt-1 text-sm font-medium text-zinc-900 dark:text-zinc-100">
              {{
                formatInvoiceBaseLabel(
                  lineOrFallback(form.invoicePrefix, 'INV-'),
                  invoiceNumberPreview(form.startingInvoiceNumber),
                )
              }}
            </div>
          </div>

          <div class="rounded-xl bg-zinc-50 px-3 py-2 dark:bg-zinc-950/50">
            <div
              class="text-tiny font-semibold tracking-[0.14em] text-zinc-600 uppercase dark:text-zinc-400"
            >
              Currency
            </div>
            <div class="mt-1 text-sm font-medium text-zinc-900 dark:text-zinc-100">
              {{ lineOrFallback(form.currency, 'GBP') }}
            </div>
          </div>
        </div>

        <div class="py-4">
          <div
            class="text-tiny grid grid-cols-[1fr_auto] gap-3 border-b border-zinc-300 pb-2 font-semibold tracking-[0.14em] text-zinc-600 uppercase dark:border-zinc-800 dark:text-zinc-400"
          >
            <div>Items</div>
            <div>Total</div>
          </div>

          <div class="space-y-2 pt-3 text-zinc-700 dark:text-zinc-300">
            <template v-if="form.showItemTypeHeaders">
              <div
                class="text-tiny font-semibold tracking-[0.14em] text-zinc-600 uppercase dark:text-zinc-400"
              >
                Styles
              </div>
            </template>
            <div class="text-mini grid grid-cols-[1fr_auto] gap-3">
              <div>Style line item</div>
              <div class="font-medium text-zinc-900 dark:text-zinc-100">75.00</div>
            </div>

            <template v-if="form.showItemTypeHeaders">
              <div
                class="text-tiny pt-1 font-semibold tracking-[0.14em] text-zinc-600 uppercase dark:text-zinc-400"
              >
                Samples
              </div>
            </template>
            <div class="text-mini grid grid-cols-[1fr_auto] gap-3">
              <div>Sample service line</div>
              <div class="font-medium text-zinc-900 dark:text-zinc-100">120.00</div>
            </div>

            <template v-if="form.showItemTypeHeaders">
              <div
                class="text-tiny pt-1 font-semibold tracking-[0.14em] text-zinc-600 uppercase dark:text-zinc-400"
              >
                Other Items
              </div>
            </template>
            <div class="text-mini grid grid-cols-[1fr_auto] gap-3">
              <div>Custom item</div>
              <div class="font-medium text-zinc-900 dark:text-zinc-100">45.00</div>
            </div>
          </div>
        </div>

        <div class="flex justify-end border-b border-zinc-300 pb-4 dark:border-zinc-800">
          <div class="text-mini w-full max-w-56 space-y-2">
            <div class="flex items-center justify-between">
              <span class="text-zinc-500 dark:text-zinc-400">Subtotal</span>
              <span class="text-zinc-900 dark:text-zinc-100">200.00</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-zinc-500 dark:text-zinc-400">VAT</span>
              <span class="text-zinc-900 dark:text-zinc-100">40.00</span>
            </div>
            <div class="text-mini flex items-center justify-between pt-1 font-semibold">
              <span class="text-zinc-900 dark:text-zinc-100">Total</span>
              <span class="text-zinc-900 dark:text-zinc-100">240.00</span>
            </div>
          </div>
        </div>

        <div class="grid gap-3 py-4">
          <div class="rounded-xl bg-zinc-50 p-3 dark:bg-zinc-950/50">
            <div
              class="text-tiny mb-1.5 font-semibold tracking-[0.14em] text-zinc-600 uppercase dark:text-zinc-400"
            >
              Payment terms
            </div>
            <div
              class="text-mini leading-relaxed whitespace-pre-line text-zinc-700 dark:text-zinc-300"
            >
              {{ lineOrFallback(form.paymentTerms, 'Payment terms will appear here.') }}
            </div>
          </div>

          <div class="rounded-xl bg-zinc-50 p-3 dark:bg-zinc-950/50">
            <div
              class="text-tiny mb-1.5 font-semibold tracking-[0.14em] text-zinc-600 uppercase dark:text-zinc-400"
            >
              Payment details
            </div>
            <div
              class="text-sm leading-relaxed whitespace-pre-line text-zinc-700 dark:text-zinc-300"
            >
              {{ lineOrFallback(form.paymentDetails, 'Bank details will appear here.') }}
            </div>
          </div>
        </div>

        <div class="border-t border-zinc-300 pt-4 dark:border-zinc-800">
          <div class="text-tiny leading-relaxed text-zinc-600 dark:text-zinc-400">
            {{ lineOrFallback(form.notesFooter, 'Footer note will appear here on each page.') }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

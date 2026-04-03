<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { useBillingCatalogStore } from '@/stores/billingCatalog'
import {
  ArrowRightIcon,
  BookOpenIcon,
  Cog6ToothIcon,
  PencilSquareIcon,
  PaperAirplaneIcon,
  CheckCircleIcon,
  SparklesIcon,
  BoltIcon,
  BanknotesIcon,
  HandThumbUpIcon,
} from '@heroicons/vue/24/outline'

const billingCatalogStore = useBillingCatalogStore()
void billingCatalogStore.fetchCatalog().catch(() => undefined)

const trialLabel = computed(() => billingCatalogStore.trialLabel)
const soloStartingPriceLabel = computed(() =>
  billingCatalogStore.getPlanStartingPriceLabel('single'),
)

const trustPoints = computed(() => [
  { label: 'Quick set up', icon: BoltIcon },
  { label: 'Easy to use', icon: HandThumbUpIcon },
  { label: `From ${soloStartingPriceLabel.value}`, icon: BanknotesIcon },
])

const invoiceLines = [
  { label: 'Kitchen fitting', qty: 1, unit: 'day', rate: 380.0, amount: 380.0 },
  { label: 'Materials & fixings', qty: 1, unit: 'item', rate: 120.0, amount: 120.0 },
  { label: 'Waste disposal', qty: 1, unit: 'item', rate: 45.0, amount: 45.0 },
]

const subtotal = 545.0
const vatAmount = 109.0
const total = 654.0
const paid = 150.0
const balance = 504.0

const floatingCards = [
  {
    title: 'Quick set up',
    meta: 'Make it yours',
    icon: Cog6ToothIcon,
    className: 'top-18 sm:top-0 -right-3 sm:-right-5 lg:-right-4 rotate-[3deg] float-slow',
    tone: 'default',
  },
  {
    title: 'Invoice book',
    meta: 'Easy records',
    icon: BookOpenIcon,
    className: 'top-[38%] -left-3 sm:-left-6 lg:-left-9 -rotate-[3deg] float-mid',
    tone: 'default',
  },
  {
    title: 'Safe edits',
    meta: 'Revision tracking',
    icon: PencilSquareIcon,
    className: 'bottom-38 -left-3 sm:-left-6 lg:-left-9 rotate-[2deg] float-fast',
    tone: 'success',
  },
]
</script>

<template>
  <section class="relative isolate overflow-visible py-8 select-none sm:py-12 lg:py-16">
    <!-- grid background -->
    <div
      class="hdr-grid pointer-events-none absolute inset-x-0 top-0 bottom-0 mask-radial-from-20% mask-radial-at-center opacity-50 dark:opacity-100"
    />

    <div class="relative grid gap-14 lg:grid-cols-2 lg:items-center lg:gap-20">
      <!-- LEFT -->
      <div class="max-w-xl lg:mb-12">
        <p
          class="text-tiny inline-flex items-center gap-1.5 rounded-full border border-sky-200 bg-sky-50 px-3 py-1 font-semibold tracking-widest text-sky-700 uppercase shadow-sm sm:text-xs dark:border-emerald-400/20 dark:bg-emerald-500/10 dark:text-emerald-300"
        >
          <SparklesIcon class="size-4" />
          Simple invoicing · Built for small business
        </p>

        <h1
          class="mt-5 text-5xl leading-[1.08] font-bold tracking-tight text-zinc-950 lg:text-7xl dark:text-white"
        >
          Set up fast.
          <br />
          <span class="text-zinc-700 dark:text-zinc-300">Send invoices quickly.</span>
          <br />
          <span class="text-zinc-600 dark:text-zinc-400">Get paid.</span>
        </h1>

        <p class="mt-5 text-base leading-7 text-zinc-600 sm:text-lg dark:text-zinc-300">
          Invoice&nbsp;&amp;&nbsp;Go is built for freelancers, trade professions, and small
          businesses who just want invoicing to work. No learning curve, no spreadsheets, no fuss.
        </p>

        <div class="mt-8 flex flex-wrap items-center gap-3">
          <RouterLink
            to="/signup"
            class="inline-flex items-center gap-2 rounded-full bg-sky-600 px-6 py-3 text-sm font-semibold text-white shadow-md transition hover:bg-sky-500 active:scale-95 dark:bg-emerald-500 dark:text-zinc-950 dark:hover:bg-emerald-400"
          >
            Start {{ trialLabel }}
            <ArrowRightIcon class="size-4" />
          </RouterLink>

          <RouterLink
            to="/login"
            class="inline-flex items-center rounded-full border border-zinc-300 bg-white px-6 py-3 text-sm font-medium text-zinc-700 transition hover:border-sky-400 hover:text-zinc-900 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:border-emerald-500/60 dark:hover:text-white"
          >
            Log in
          </RouterLink>
        </div>

        <div class="mt-8 flex flex-wrap gap-2">
          <div
            v-for="point in trustPoints"
            :key="point.label"
            class="flex items-center gap-1.5 rounded-full border border-zinc-200 bg-white px-4 py-2 text-sm font-medium text-zinc-700 shadow-sm dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-300"
          >
            <component
              :is="point.icon"
              class="size-5"
            />
            {{ point.label }}
          </div>
        </div>
      </div>

      <!-- ── RIGHT Invoice card + floating cards ── -->
      <div class="relative mx-auto w-full max-w-176 lg:mx-0">
        <div class="relative pt-6 pb-14 sm:px-10 lg:px-12">
          <!-- Glow -->
          <div
            class="pointer-events-none absolute inset-x-16 top-10 h-64 rounded-full bg-sky-200/30 blur-3xl dark:bg-emerald-500/10"
          />

          <article
            class="relative z-10 overflow-hidden rounded-3xl border border-zinc-200/80 bg-white/95 shadow-2xl backdrop-blur select-none dark:border-zinc-800/80 dark:bg-zinc-950/95 dark:shadow-2xl dark:shadow-black/80"
          >
            <div
              class="pointer-events-none absolute inset-x-0 top-0 h-20 bg-linear-to-b from-white/80 to-transparent dark:from-zinc-950/70"
            />

            <div class="relative p-5 sm:p-7">
              <!-- Invoice header -->
              <div
                class="flex items-start justify-between gap-4 border-b border-zinc-100 pb-5 dark:border-zinc-800/80"
              >
                <div>
                  <p
                    class="text-tiny font-semibold tracking-[0.2em] text-zinc-400 uppercase dark:text-zinc-500"
                  >
                    Invoice
                  </p>
                  <h2
                    class="mt-1.5 text-2xl font-bold tracking-tight text-zinc-950 dark:text-white"
                  >
                    INV-2048
                  </h2>
                  <p class="mt-1 text-sm font-medium text-zinc-500 dark:text-zinc-400">
                    North Vale Studio
                  </p>
                  <p class="mt-0.5 text-xs text-zinc-400 dark:text-zinc-500">
                    Issued 3 Apr 2026 · Due 17 Apr 2026
                  </p>
                </div>

                <span
                  class="inline-flex shrink-0 items-center gap-1.5 rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1.5 text-xs font-semibold text-emerald-700 dark:border-emerald-400/20 dark:bg-emerald-500/10 dark:text-emerald-300"
                >
                  <CheckCircleIcon class="size-3.5" />
                  Ready to send
                </span>
              </div>

              <!-- items -->
              <div
                class="mt-5 overflow-hidden rounded-2xl border border-zinc-100 dark:border-zinc-800/80"
              >
                <div
                  class="text-tiny grid grid-cols-[minmax(0,1fr)_3.5rem_5.5rem] gap-2 border-b border-zinc-100 bg-zinc-50/80 px-4 py-2.5 font-semibold tracking-widest text-zinc-400 uppercase dark:border-zinc-800/80 dark:bg-zinc-900/60 dark:text-zinc-500"
                >
                  <span>Description</span>
                  <span class="text-right">Qty</span>
                  <span class="text-right">Amount</span>
                </div>

                <div
                  v-for="(line, i) in invoiceLines"
                  :key="line.label"
                  class="grid grid-cols-[minmax(0,1fr)_3.5rem_5.5rem] gap-2 px-4 py-3 text-sm"
                  :class="
                    i < invoiceLines.length - 1
                      ? 'border-b border-zinc-100 dark:border-zinc-800/60'
                      : ''
                  "
                >
                  <div class="min-w-0">
                    <p class="truncate font-medium text-zinc-900 dark:text-zinc-100">
                      {{ line.label }}
                    </p>
                    <p class="mt-0.5 text-[11px] text-zinc-400 dark:text-zinc-500">
                      £{{ line.rate.toFixed(2) }} / {{ line.unit }}
                    </p>
                  </div>
                  <span class="self-center text-right text-zinc-500 dark:text-zinc-400">
                    {{ line.qty }}
                  </span>
                  <span
                    class="self-center text-right font-semibold text-zinc-900 dark:text-zinc-100"
                  >
                    £{{ line.amount.toFixed(2) }}
                  </span>
                </div>
              </div>

              <!-- Totals + Send -->
              <div class="mt-4 grid gap-3">
                <!-- Totals -->
                <div
                  class="space-y-2 rounded-2xl border border-zinc-100 bg-zinc-50/70 px-4 py-4 dark:border-zinc-800/80 dark:bg-zinc-900/60"
                >
                  <div
                    v-for="row in [
                      { label: 'Subtotal', value: subtotal },
                      { label: 'VAT (20%)', value: vatAmount },
                    ]"
                    :key="row.label"
                    class="flex items-center justify-between text-sm"
                  >
                    <span class="text-zinc-500 dark:text-zinc-400">{{ row.label }}</span>
                    <span class="font-medium text-zinc-700 dark:text-zinc-300">
                      £{{ row.value.toFixed(2) }}
                    </span>
                  </div>

                  <div class="my-1 border-t border-zinc-200 dark:border-zinc-700/80" />

                  <div class="flex items-center justify-between text-sm">
                    <span class="font-semibold text-zinc-950 dark:text-white">Total</span>
                    <span class="font-bold text-zinc-950 dark:text-white">
                      £{{ total.toFixed(2) }}
                    </span>
                  </div>

                  <div
                    class="flex items-center justify-between text-sm text-emerald-600 dark:text-emerald-400"
                  >
                    <span>Paid</span>
                    <span class="font-medium">−£{{ paid.toFixed(2) }}</span>
                  </div>

                  <div class="rounded-xl bg-white px-3 py-2.5 dark:bg-zinc-950/60">
                    <div class="flex items-center justify-between">
                      <span class="text-sm font-bold text-zinc-950 dark:text-white">
                        Balance due
                      </span>
                      <span class="text-base font-bold text-zinc-950 dark:text-white">
                        £{{ balance.toFixed(2) }}
                      </span>
                    </div>
                  </div>
                </div>

                <!-- Create Button -->
                <button
                  type="button"
                  class="group flex w-full items-center justify-center gap-2.5 rounded-2xl bg-sky-600 px-5 py-4 text-sm font-semibold text-white shadow-lg shadow-sky-600/25 transition hover:bg-sky-500 active:scale-95 dark:bg-emerald-500 dark:text-zinc-950 dark:shadow-emerald-500/20 dark:hover:bg-emerald-400"
                >
                  <PaperAirplaneIcon
                    class="size-5 transition-transform group-hover:translate-x-0.5 group-hover:-translate-y-0.5"
                  />
                  <span>Create Invoice</span>
                </button>
              </div>

              <!-- Notes footer -->
              <p class="mt-3 text-[11px] leading-relaxed text-zinc-400 dark:text-zinc-500">
                Thank you for your business. Payment due within 14 days.
              </p>
            </div>
          </article>

          <!-- Floating Hero Cards -->
          <article
            v-for="card in floatingCards"
            :key="card.title"
            :class="[
              'absolute z-20 w-56 rounded-2xl border p-3.5 shadow-xl',
              card.className,
              card.tone === 'success'
                ? 'border-emerald-200/80 bg-emerald-50/90 dark:border-emerald-400/20 dark:bg-emerald-500/10'
                : 'border-zinc-200/90 bg-white/95 dark:border-zinc-800/90 dark:bg-zinc-900/95',
            ]"
          >
            <div class="flex items-center gap-3">
              <div
                class="flex size-11 shrink-0 items-center justify-center rounded-xl border"
                :class="
                  card.tone === 'success'
                    ? 'border-emerald-300/60 bg-white/80 text-emerald-600 dark:border-emerald-400/25 dark:bg-zinc-950/40 dark:text-emerald-300'
                    : 'border-sky-200 bg-sky-50 text-sky-700 dark:border-emerald-400/20 dark:bg-emerald-500/10 dark:text-emerald-200'
                "
              >
                <component
                  :is="card.icon"
                  class="size-6"
                />
              </div>

              <div class="min-w-0">
                <p class="text-lg leading-tight font-semibold text-zinc-950 dark:text-white">
                  {{ card.title }}
                </p>
                <p class="mt-1 text-base leading-tight text-zinc-800 dark:text-zinc-300">
                  {{ card.meta }}
                </p>
              </div>
            </div>
          </article>
        </div>
      </div>
    </div>
  </section>
</template>

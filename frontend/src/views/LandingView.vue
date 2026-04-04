<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import LandingHero from '@/components/landing/LandingHero.vue'
import LandingAppPannel from '@/components/landing/LandingAppPannel.vue'
import LandingBuiltForRealUse from '@/components/landing/LandingBuiltForRealUse.vue'
import { TEAM_PLAN_SEAT_LIMIT } from '@/constants/billing'
import { useBillingCatalogStore } from '@/stores/billingCatalog'
import { ArrowRightIcon, CheckCircleIcon } from '@heroicons/vue/24/outline'
import LandingTrust from '@/components/landing/LandingTrust.vue'
import LandingHeader from '@/components/landing/LandingHeader.vue'

const billingCatalogStore = useBillingCatalogStore()
void billingCatalogStore.fetchCatalog().catch(() => undefined)

const workspaceTrialLabel = computed(() => billingCatalogStore.trialLabel)
const soloStartingPriceLabel = computed(() =>
  billingCatalogStore.getPlanStartingPriceLabel('single'),
)
const teamStartingPriceLabel = computed(() => billingCatalogStore.getPlanStartingPriceLabel('team'))
const soloYearlyPriceLabel = computed(() => billingCatalogStore.getPriceLabel('single', 'yearly'))
const teamYearlyPriceLabel = computed(() => billingCatalogStore.getPriceLabel('team', 'yearly'))

const pricingCards = computed(() => [
  {
    name: 'Solo',
    price: soloStartingPriceLabel.value,
    yearly: soloYearlyPriceLabel.value,
    helper: 'For freelancers and one-person businesses.',
    featured: true,
    points: [
      workspaceTrialLabel.value,
      'Full invoicing workspace',
      'Simple setup and invoice history',
      'Move to team later if needed',
    ],
  },
  {
    name: 'Team',
    price: teamStartingPriceLabel.value,
    yearly: teamYearlyPriceLabel.value,
    helper: `For shared workspaces with up to ${TEAM_PLAN_SEAT_LIMIT} people.`,
    featured: false,
    points: [
      'Shared clients and invoices',
      'Invite teammates with Google sign-in',
      `Up to ${TEAM_PLAN_SEAT_LIMIT} seats`,
      'One workspace, shared records',
    ],
  },
])

const faqItems = computed(() => [
  {
    question: 'Is there a free trial?',
    answer: `Yes. Every workspace starts with ${workspaceTrialLabel.value.toLowerCase()}.`,
  },
  {
    question: 'Can I start solo and upgrade later?',
    answer: 'Yes. Start alone, then move to the team plan when you need shared access.',
  },
  {
    question: 'Do teammates see the same invoices and clients?',
    answer: 'Yes. The team plan is one shared workspace, not disconnected personal accounts.',
  },
  {
    question: 'Is this built for accountants?',
    answer: 'No. It is built for people who just want invoicing to be clear and quick.',
  },
])
</script>

<template>
  <main
    class="min-h-screen bg-linear-to-b from-sky-50 via-white to-slate-100 text-zinc-900 dark:from-[#06110f] dark:via-[#091312] dark:to-[#0b1715] dark:text-zinc-100"
  >
    <div class="relative">
      <div
        class="pointer-events-none absolute inset-x-0 top-0 h-160 bg-radial from-sky-200/35 via-transparent to-transparent dark:from-emerald-500/10"
      />
      <div
        class="pointer-events-none absolute top-32 left-1/2 h-96 w-96 -translate-x-1/2 rounded-full bg-sky-200/20 blur-3xl dark:bg-emerald-500/10"
      />

      <section class="relative mx-auto w-full pb-6">
        <LandingHeader class="sticky top-0 z-10 mx-auto" />

        <LandingHero
          class="my-10 max-w-7xl place-self-center px-5 sm:my-14 sm:px-8 lg:my-18 lg:px-10"
        />

        <LandingTrust />
        <LandingAppPannel />
        <LandingBuiltForRealUse
          :trial-label="workspaceTrialLabel"
          :team-seat-limit="TEAM_PLAN_SEAT_LIMIT"
        />
        <!-- Pricing -->
        <section class="py-10 sm:py-14 lg:py-18">
          <div class="grid gap-6 lg:grid-cols-[1.05fr_0.95fr] lg:items-start">
            <div>
              <p
                class="text-xs font-semibold tracking-[0.18em] text-sky-700 uppercase sm:text-sm dark:text-emerald-300"
              >
                Pricing
              </p>
              <h2 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl dark:text-white">
                Straight pricing for small businesses
              </h2>
              <p
                class="mt-4 max-w-2xl text-base leading-8 text-zinc-600 sm:text-lg dark:text-zinc-300"
              >
                No giant plan matrix. Just one option for solo work and one for a small shared team.
              </p>
            </div>

            <div
              class="rounded-4xl border border-zinc-200 bg-zinc-50/80 px-5 py-4 text-sm leading-7 text-zinc-600 dark:border-zinc-800 dark:bg-zinc-900/80 dark:text-zinc-300"
            >
              Built to be affordable early, and easy to outgrow later.
            </div>
          </div>

          <div class="mt-8 grid gap-5 lg:grid-cols-2">
            <article
              v-for="card in pricingCards"
              :key="card.name"
              :class="[
                'rounded-4xl border p-6 shadow-sm sm:p-8',
                card.featured
                  ? 'border-sky-300 bg-white dark:border-emerald-400/30 dark:bg-zinc-950/85'
                  : 'border-zinc-200 bg-white/85 dark:border-zinc-800 dark:bg-zinc-950/75',
              ]"
            >
              <div class="flex items-start justify-between gap-4">
                <div>
                  <h3 class="text-2xl font-semibold text-zinc-950 dark:text-white">
                    {{ card.name }}
                  </h3>
                  <p class="mt-2 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                    {{ card.helper }}
                  </p>
                </div>

                <span
                  v-if="card.featured"
                  class="rounded-full border border-sky-200 bg-sky-50 px-3 py-1 text-[11px] font-semibold tracking-[0.14em] text-sky-700 uppercase dark:border-emerald-400/20 dark:bg-emerald-500/10 dark:text-emerald-200"
                >
                  Popular
                </span>
              </div>

              <div class="mt-8 border-t border-zinc-200 pt-6 dark:border-zinc-800">
                <div class="text-4xl font-semibold tracking-tight text-zinc-950 dark:text-white">
                  {{ card.price }}
                </div>
                <p class="mt-2 text-sm text-zinc-500 dark:text-zinc-400">
                  Or {{ card.yearly }} yearly
                </p>
              </div>

              <ul class="mt-8 space-y-3">
                <li
                  v-for="point in card.points"
                  :key="point"
                  class="flex items-start gap-3 text-sm text-zinc-700 dark:text-zinc-200"
                >
                  <CheckCircleIcon
                    class="mt-0.5 size-4 shrink-0 text-sky-600 dark:text-emerald-300"
                  />
                  <span>{{ point }}</span>
                </li>
              </ul>
            </article>
          </div>
        </section>
        <!-- FAQ -->
        <section class="py-10 sm:py-14 lg:py-18">
          <div
            class="rounded-4xl border border-zinc-200 bg-white/90 p-6 shadow-sm sm:p-8 lg:p-10 dark:border-zinc-800 dark:bg-zinc-950/80"
          >
            <div class="max-w-3xl">
              <p
                class="text-xs font-semibold tracking-[0.18em] text-sky-700 uppercase sm:text-sm dark:text-emerald-300"
              >
                Questions
              </p>
              <h2 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl dark:text-white">
                The usual things people ask before signing up
              </h2>
            </div>

            <div class="mt-8 divide-y divide-zinc-200 dark:divide-zinc-800">
              <article
                v-for="item in faqItems"
                :key="item.question"
                class="grid gap-3 py-5 lg:grid-cols-[0.9fr_1.1fr] lg:gap-8"
              >
                <h3 class="text-base font-semibold text-zinc-950 dark:text-white">
                  {{ item.question }}
                </h3>
                <p class="text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                  {{ item.answer }}
                </p>
              </article>
            </div>
          </div>
        </section>

        <section class="pt-8 pb-14 sm:pt-10 sm:pb-20">
          <div
            class="overflow-hidden rounded-4xl border border-sky-200 bg-linear-to-br from-sky-950 via-sky-900 to-sky-700 p-6 text-white shadow-lg sm:p-8 lg:p-10 dark:border-emerald-400/15 dark:from-[#09110e] dark:via-[#0d1d18] dark:to-[#123128]"
          >
            <div class="grid gap-8 lg:grid-cols-[1fr_auto] lg:items-end">
              <div class="max-w-3xl">
                <p
                  class="text-xs font-semibold tracking-[0.18em] text-sky-200 uppercase sm:text-sm dark:text-emerald-200"
                >
                  Ready to start
                </p>
                <h2 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl">
                  Set up the workspace once. Then just invoice.
                </h2>
                <p
                  class="mt-4 max-w-2xl text-base leading-8 text-sky-50/85 sm:text-lg dark:text-emerald-50/85"
                >
                  Start your {{ workspaceTrialLabel.toLowerCase() }}, get the basics in place, and
                  keep invoicing simple from day one.
                </p>
              </div>

              <div class="flex flex-col gap-3 sm:flex-row lg:flex-col">
                <RouterLink
                  to="/signup"
                  class="inline-flex items-center justify-center gap-2 rounded-full bg-white px-5 py-3 text-sm font-semibold text-sky-900 transition hover:bg-sky-50 dark:bg-emerald-500 dark:text-zinc-950 dark:hover:bg-emerald-400"
                >
                  Start free trial
                  <ArrowRightIcon class="size-4" />
                </RouterLink>

                <RouterLink
                  to="/login"
                  class="inline-flex items-center justify-center rounded-full border border-white/20 px-5 py-3 text-sm font-medium text-white transition hover:border-white/40 hover:bg-white/5 dark:border-emerald-400/20 dark:hover:border-emerald-300/40 dark:hover:bg-emerald-500/10"
                >
                  Log in
                </RouterLink>
              </div>
            </div>
          </div>
        </section>
      </section>
    </div>
  </main>
</template>

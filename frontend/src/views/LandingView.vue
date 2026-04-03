<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import DarkMode from '@/components/UI/DarkMode.vue'
import LandingHero from '@/components/landing/LandingHero.vue'
import { TEAM_PLAN_SEAT_LIMIT } from '@/constants/billing'
import { useBillingCatalogStore } from '@/stores/billingCatalog'
import {
  ArrowRightIcon,
  CheckCircleIcon,
  ClockIcon,
  CreditCardIcon,
  DocumentDuplicateIcon,
  FolderOpenIcon,
  PencilSquareIcon,
  ShieldCheckIcon,
  UsersIcon,
} from '@heroicons/vue/24/outline'

const billingCatalogStore = useBillingCatalogStore()
void billingCatalogStore.fetchCatalog().catch(() => undefined)

const workspaceTrialLabel = computed(() => billingCatalogStore.trialLabel)
const soloStartingPriceLabel = computed(() =>
  billingCatalogStore.getPlanStartingPriceLabel('single'),
)
const teamStartingPriceLabel = computed(() => billingCatalogStore.getPlanStartingPriceLabel('team'))
const soloYearlyPriceLabel = computed(() => billingCatalogStore.getPriceLabel('single', 'yearly'))
const teamYearlyPriceLabel = computed(() => billingCatalogStore.getPriceLabel('team', 'yearly'))

const proofPoints = computed(() => [
  workspaceTrialLabel.value,
  `Solo from ${soloStartingPriceLabel.value}`,
  `Teams up to ${TEAM_PLAN_SEAT_LIMIT} people`,
])

const steps = [
  {
    step: '01',
    title: 'Sign in with Google',
    body: 'Get into the app with the account you already use. No passwords to manage.',
  },
  {
    step: '02',
    title: 'Set your details once',
    body: 'Add your business info, payment details, and defaults so invoices start ready.',
  },
  {
    step: '03',
    title: 'Create and send faster',
    body: 'Reuse clients and common line items so repeat work takes less effort.',
  },
]

const featureCards = [
  {
    title: 'Quick set up',
    body: 'Get your business details, invoice defaults, and payment info in place without a long onboarding flow.',
    icon: ClockIcon,
  },
  {
    title: 'Easy records',
    body: 'Keep your clients and invoice history tidy so past work stays easy to find and review.',
    icon: FolderOpenIcon,
  },
  {
    title: 'Safe edits',
    body: 'Make changes properly with revision tracking instead of quietly breaking your records.',
    icon: PencilSquareIcon,
  },
  {
    title: 'Repeat work faster',
    body: 'Reuse familiar invoice lines and client details so regular jobs are quicker to send out.',
    icon: DocumentDuplicateIcon,
  },
  {
    title: 'Built to stay clear',
    body: 'The interface stays focused on invoicing instead of burying simple tasks under clutter.',
    icon: ShieldCheckIcon,
  },
  {
    title: 'Room for a small team',
    body: `Start solo, then move into a shared workspace later with support for up to ${TEAM_PLAN_SEAT_LIMIT} people.`,
    icon: UsersIcon,
  },
]

const pricingCards = computed(() => [
  {
    name: 'Solo',
    price: soloStartingPriceLabel.value,
    yearly: soloYearlyPriceLabel.value,
    helper: 'For one person running the business.',
    featured: true,
    bullets: [
      workspaceTrialLabel.value,
      'Full invoicing workspace',
      'One workspace owner',
      'Upgrade to team later',
    ],
  },
  {
    name: 'Team',
    price: teamStartingPriceLabel.value,
    yearly: teamYearlyPriceLabel.value,
    helper: `Shared workspace for up to ${TEAM_PLAN_SEAT_LIMIT} people.`,
    featured: false,
    bullets: [
      'Shared clients and invoices',
      'Invite teammates with Google sign-in',
      `Up to ${TEAM_PLAN_SEAT_LIMIT} people`,
      'One account, shared workflow',
    ],
  },
])

const faqs = computed(() => [
  {
    question: 'Is there a free trial?',
    answer: `Yes. You can start with a ${workspaceTrialLabel.value.toLowerCase()} before moving onto a paid plan.`,
  },
  {
    question: 'Can I start on my own and add people later?',
    answer: `Yes. You can start on the solo plan and move to team access later when you need a shared workspace.`,
  },
  {
    question: 'Do teammates share the same clients and invoices?',
    answer:
      'Yes. The team plan is built around one shared workspace so everyone works from the same records.',
  },
  {
    question: 'Do I need to learn a complicated system first?',
    answer:
      'No. The whole point is to keep the setup light and the day-to-day invoicing straightforward.',
  },
])
</script>

<template>
  <main
    class="min-h-screen bg-linear-to-b from-sky-50 via-white to-slate-50 text-zinc-900 dark:from-[#07110f] dark:via-[#081312] dark:to-[#091514] dark:text-zinc-100"
  >
    <div class="relative overflow-hidden">
      <div
        class="pointer-events-none absolute inset-x-0 top-0 h-144 bg-radial from-sky-200/35 via-transparent to-transparent dark:from-emerald-500/10"
      />
      <div
        class="pointer-events-none absolute top-40 left-1/2 h-80 w-80 -translate-x-1/2 rounded-full bg-sky-200/25 blur-3xl dark:bg-emerald-500/10"
      />

      <section class="relative mx-auto w-full max-w-7xl px-5 py-6 sm:px-8 lg:px-10">
        <header
          class="flex flex-col gap-4 rounded-4xl border border-white/80 bg-white/85 px-4 py-4 shadow-sm backdrop-blur sm:flex-row sm:items-center sm:justify-between sm:px-5 dark:border-white/10 dark:bg-zinc-950/75"
        >
          <div class="flex items-center gap-4">
            <RouterLink
              to="/"
              class="inline-flex items-center gap-3 text-sm font-semibold tracking-[0.2em] uppercase"
            >
              <span class="size-2.5 rounded-full bg-sky-500 dark:bg-emerald-400" />
              <span class="text-zinc-950 dark:text-zinc-100">Invoice and Go</span>
            </RouterLink>

            <DarkMode variant="pill" />
          </div>

          <div class="flex flex-wrap items-center gap-2">
            <RouterLink
              to="/login"
              class="inline-flex items-center justify-center rounded-full border border-zinc-300 bg-white px-4 py-2 text-sm font-medium text-zinc-700 transition hover:border-sky-400 hover:text-zinc-950 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:border-emerald-400/60 dark:hover:text-white"
            >
              Log in
            </RouterLink>

            <RouterLink
              to="/signup"
              class="inline-flex items-center justify-center gap-2 rounded-full bg-sky-600 px-6 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-sky-500 active:scale-95 dark:bg-emerald-500 dark:text-zinc-950 dark:hover:bg-emerald-400"
            >
              Register
              <ArrowRightIcon class="size-4" />
            </RouterLink>
          </div>
        </header>

        <LandingHero class="my-10 sm:my-16 lg:my-20" />

        <section class="pb-6 sm:pb-8">
          <div
            class="grid gap-3 rounded-4xl border border-zinc-200/80 bg-white/80 p-4 shadow-sm backdrop-blur sm:grid-cols-3 sm:p-5 dark:border-zinc-800 dark:bg-zinc-950/70"
          >
            <div
              v-for="point in proofPoints"
              :key="point"
              class="rounded-2xl border border-zinc-200/80 bg-zinc-50/80 px-4 py-3 text-sm font-medium text-zinc-700 dark:border-zinc-800 dark:bg-zinc-900/80 dark:text-zinc-300"
            >
              {{ point }}
            </div>
          </div>
        </section>

        <section class="py-10 sm:py-14 lg:py-18">
          <div class="max-w-3xl">
            <p
              class="text-xs font-semibold tracking-[0.18em] text-sky-700 uppercase sm:text-sm dark:text-emerald-300"
            >
              How it works
            </p>
            <h2 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl dark:text-white">
              Get from sign-up to sent invoice without the usual drag
            </h2>
            <p
              class="mt-4 max-w-2xl text-base leading-7 text-zinc-600 sm:text-lg dark:text-zinc-300"
            >
              The product should feel fast to pick up. Set your details, create invoices, and keep
              records organised without turning invoicing into admin theatre.
            </p>
          </div>

          <div class="mt-8 grid gap-4 lg:grid-cols-3">
            <article
              v-for="item in steps"
              :key="item.step"
              class="rounded-4xl border border-zinc-200 bg-white p-6 shadow-sm dark:border-zinc-800 dark:bg-zinc-950/80"
            >
              <div
                class="inline-flex rounded-full border border-sky-200 bg-sky-50 px-3 py-1 text-xs font-semibold tracking-[0.14em] text-sky-700 uppercase dark:border-emerald-400/20 dark:bg-emerald-500/10 dark:text-emerald-200"
              >
                {{ item.step }}
              </div>
              <h3 class="mt-5 text-xl font-semibold text-zinc-950 dark:text-white">
                {{ item.title }}
              </h3>
              <p class="mt-3 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                {{ item.body }}
              </p>
            </article>
          </div>
        </section>

        <section class="py-10 sm:py-14 lg:py-18">
          <div
            class="overflow-hidden rounded-4xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/80"
          >
            <div class="grid lg:grid-cols-[0.95fr_1.05fr]">
              <div
                class="border-b border-zinc-200 p-6 sm:p-8 lg:border-r lg:border-b-0 lg:p-10 dark:border-zinc-800"
              >
                <p
                  class="text-xs font-semibold tracking-[0.18em] text-sky-700 uppercase sm:text-sm dark:text-emerald-300"
                >
                  Why it feels better
                </p>
                <h2 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl dark:text-white">
                  A calmer landing for the rest of the product story
                </h2>
                <p class="mt-4 text-base leading-7 text-zinc-600 sm:text-lg dark:text-zinc-300">
                  Instead of piling on noise, this section keeps the benefits obvious: faster setup,
                  clearer records, safer edits, and a workflow that still works once you start
                  repeating jobs.
                </p>

                <div class="mt-8 space-y-4">
                  <div class="flex gap-3">
                    <div
                      class="mt-1 flex size-10 shrink-0 items-center justify-center rounded-2xl border border-sky-200 bg-sky-50 text-sky-700 dark:border-emerald-400/20 dark:bg-emerald-500/10 dark:text-emerald-200"
                    >
                      <CheckCircleIcon class="size-5" />
                    </div>
                    <div>
                      <h3 class="font-semibold text-zinc-950 dark:text-white">Less repetition</h3>
                      <p class="mt-1 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                        Reuse client details and common invoice lines instead of rebuilding work
                        every time.
                      </p>
                    </div>
                  </div>

                  <div class="flex gap-3">
                    <div
                      class="mt-1 flex size-10 shrink-0 items-center justify-center rounded-2xl border border-sky-200 bg-sky-50 text-sky-700 dark:border-emerald-400/20 dark:bg-emerald-500/10 dark:text-emerald-200"
                    >
                      <ShieldCheckIcon class="size-5" />
                    </div>
                    <div>
                      <h3 class="font-semibold text-zinc-950 dark:text-white">Fewer mistakes</h3>
                      <p class="mt-1 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                        Keep records tidy and make edits properly instead of patching over old work.
                      </p>
                    </div>
                  </div>

                  <div class="flex gap-3">
                    <div
                      class="mt-1 flex size-10 shrink-0 items-center justify-center rounded-2xl border border-sky-200 bg-sky-50 text-sky-700 dark:border-emerald-400/20 dark:bg-emerald-500/10 dark:text-emerald-200"
                    >
                      <CreditCardIcon class="size-5" />
                    </div>
                    <div>
                      <h3 class="font-semibold text-zinc-950 dark:text-white">Straight pricing</h3>
                      <p class="mt-1 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                        A small solo plan, a simple team plan, and no weird enterprise fluff.
                      </p>
                    </div>
                  </div>
                </div>
              </div>

              <div class="grid gap-4 p-6 sm:grid-cols-2 sm:p-8 lg:p-10">
                <article
                  v-for="card in featureCards"
                  :key="card.title"
                  class="rounded-[1.75rem] border border-zinc-200 bg-zinc-50/80 p-5 dark:border-zinc-800 dark:bg-zinc-900/80"
                >
                  <div
                    class="inline-flex items-center justify-center rounded-2xl border border-sky-200 bg-white p-3 text-sky-700 dark:border-emerald-400/20 dark:bg-zinc-950 dark:text-emerald-200"
                  >
                    <component
                      :is="card.icon"
                      class="size-5"
                    />
                  </div>
                  <h3 class="mt-4 text-lg font-semibold text-zinc-950 dark:text-white">
                    {{ card.title }}
                  </h3>
                  <p class="mt-2 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                    {{ card.body }}
                  </p>
                </article>
              </div>
            </div>
          </div>
        </section>

        <section class="py-10 sm:py-14 lg:py-18">
          <div class="max-w-3xl">
            <p
              class="text-xs font-semibold tracking-[0.18em] text-sky-700 uppercase sm:text-sm dark:text-emerald-300"
            >
              Pricing
            </p>
            <h2 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl dark:text-white">
              Clear pricing for solo work and small teams
            </h2>
            <p
              class="mt-4 max-w-2xl text-base leading-7 text-zinc-600 sm:text-lg dark:text-zinc-300"
            >
              Start with the free trial, get your account set up properly, then stay on the plan
              that matches how you work.
            </p>
          </div>

          <div class="mt-8 grid gap-5 lg:grid-cols-[1.05fr_0.95fr]">
            <div class="grid gap-5 md:grid-cols-2">
              <article
                v-for="card in pricingCards"
                :key="card.name"
                :class="[
                  'rounded-4xl p-6 shadow-sm',
                  card.featured
                    ? 'border-sky-300 bg-sky-50 dark:border-emerald-400/30 dark:bg-emerald-500/10'
                    : 'border-zinc-200 bg-white dark:border-zinc-800 dark:bg-zinc-950/80',
                ]"
              >
                <div class="flex items-start justify-between gap-4">
                  <div>
                    <h3 class="text-xl font-semibold text-zinc-950 dark:text-white">
                      {{ card.name }}
                    </h3>
                    <p class="mt-2 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                      {{ card.helper }}
                    </p>
                  </div>

                  <span
                    v-if="card.featured"
                    class="rounded-full border border-sky-200 bg-white px-3 py-1 text-[11px] font-semibold tracking-[0.14em] text-sky-700 uppercase dark:border-emerald-400/20 dark:bg-zinc-950 dark:text-emerald-200"
                  >
                    Popular
                  </span>
                </div>

                <div class="mt-6">
                  <div class="text-4xl font-semibold tracking-tight text-zinc-950 dark:text-white">
                    {{ card.price }}
                  </div>
                  <p class="mt-2 text-sm text-zinc-500 dark:text-zinc-400">
                    Also available yearly at {{ card.yearly }}
                  </p>
                </div>

                <ul class="mt-6 space-y-3">
                  <li
                    v-for="bullet in card.bullets"
                    :key="bullet"
                    class="flex items-start gap-2 text-sm text-zinc-700 dark:text-zinc-200"
                  >
                    <CheckCircleIcon
                      class="mt-0.5 size-4 shrink-0 text-sky-600 dark:text-emerald-300"
                    />
                    <span>{{ bullet }}</span>
                  </li>
                </ul>
              </article>
            </div>

            <aside
              class="rounded-4xl border border-zinc-200 bg-white p-6 shadow-sm sm:p-8 dark:border-zinc-800 dark:bg-zinc-950/80"
            >
              <p
                class="text-xs font-semibold tracking-[0.18em] text-sky-700 uppercase sm:text-sm dark:text-emerald-300"
              >
                Included
              </p>
              <h3 class="mt-3 text-2xl font-semibold tracking-tight text-zinc-950 dark:text-white">
                Built for small businesses that want less faff
              </h3>

              <ul class="mt-6 space-y-4">
                <li class="flex gap-3">
                  <CheckCircleIcon
                    class="mt-0.5 size-5 shrink-0 text-sky-600 dark:text-emerald-300"
                  />
                  <div>
                    <p class="font-medium text-zinc-950 dark:text-white">Google sign-in</p>
                    <p class="mt-1 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                      Start quickly without another password flow.
                    </p>
                  </div>
                </li>

                <li class="flex gap-3">
                  <CheckCircleIcon
                    class="mt-0.5 size-5 shrink-0 text-sky-600 dark:text-emerald-300"
                  />
                  <div>
                    <p class="font-medium text-zinc-950 dark:text-white">Invoice history</p>
                    <p class="mt-1 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                      Keep older invoices easy to find and review.
                    </p>
                  </div>
                </li>

                <li class="flex gap-3">
                  <CheckCircleIcon
                    class="mt-0.5 size-5 shrink-0 text-sky-600 dark:text-emerald-300"
                  />
                  <div>
                    <p class="font-medium text-zinc-950 dark:text-white">Revision tracking</p>
                    <p class="mt-1 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                      Update invoices properly without losing the record of what happened.
                    </p>
                  </div>
                </li>

                <li class="flex gap-3">
                  <CheckCircleIcon
                    class="mt-0.5 size-5 shrink-0 text-sky-600 dark:text-emerald-300"
                  />
                  <div>
                    <p class="font-medium text-zinc-950 dark:text-white">Small-team ready</p>
                    <p class="mt-1 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                      Add teammates later when the work stops being just you.
                    </p>
                  </div>
                </li>
              </ul>

              <div
                class="mt-8 rounded-3xl border border-zinc-200 bg-zinc-50/80 p-4 dark:border-zinc-800 dark:bg-zinc-900/80"
              >
                <p class="text-sm font-medium text-zinc-950 dark:text-white">
                  Start with {{ workspaceTrialLabel.toLowerCase() }}
                </p>
                <p class="mt-1 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                  Set things up properly first, then carry on with solo or team pricing.
                </p>
              </div>
            </aside>
          </div>
        </section>

        <section class="py-10 sm:py-14 lg:py-18">
          <div
            class="rounded-4xl border border-zinc-200 bg-white p-6 shadow-sm sm:p-8 lg:p-10 dark:border-zinc-800 dark:bg-zinc-950/80"
          >
            <div class="max-w-3xl">
              <p
                class="text-xs font-semibold tracking-[0.18em] text-sky-700 uppercase sm:text-sm dark:text-emerald-300"
              >
                Questions
              </p>
              <h2 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl dark:text-white">
                The usual things people want to know
              </h2>
            </div>

            <div class="mt-8 grid gap-4 lg:grid-cols-2">
              <article
                v-for="item in faqs"
                :key="item.question"
                class="rounded-[1.75rem] border border-zinc-200 bg-zinc-50/80 p-5 dark:border-zinc-800 dark:bg-zinc-900/80"
              >
                <h3 class="text-base font-semibold text-zinc-950 dark:text-white">
                  {{ item.question }}
                </h3>
                <p class="mt-2 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                  {{ item.answer }}
                </p>
              </article>
            </div>
          </div>
        </section>

        <section class="pt-8 pb-14 sm:pt-10 sm:pb-20">
          <div
            class="rounded-4xl border border-sky-200 bg-linear-to-br from-sky-950 via-sky-900 to-sky-700 p-6 text-white shadow-lg sm:p-8 lg:p-10 dark:border-emerald-400/15 dark:from-[#09110e] dark:via-[#0d1d18] dark:to-[#123128]"
          >
            <div class="grid gap-6 lg:grid-cols-[1fr_auto] lg:items-end">
              <div class="max-w-3xl">
                <p
                  class="text-xs font-semibold tracking-[0.18em] text-sky-200 uppercase sm:text-sm dark:text-emerald-200"
                >
                  Ready to start
                </p>
                <h2 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl">
                  Ready to make invoicing simpler?
                </h2>
                <p
                  class="mt-4 max-w-2xl text-base leading-7 text-sky-50/85 sm:text-lg dark:text-emerald-50/85"
                >
                  Start your {{ workspaceTrialLabel.toLowerCase() }}, by registering an account, and
                  keep invoicing straightforward from day one.
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

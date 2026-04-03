<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import DarkMode from '@/components/UI/DarkMode.vue'
import { TEAM_PLAN_SEAT_LIMIT } from '@/constants/billing'
import { useBillingCatalogStore } from '@/stores/billingCatalog'
import {
  ArrowRightIcon,
  CalendarDaysIcon,
  CheckCircleIcon,
  CreditCardIcon,
  DocumentTextIcon,
  FolderOpenIcon,
  PencilSquareIcon,
  ShieldCheckIcon,
  SparklesIcon,
  SwatchIcon,
  UserIcon,
  UsersIcon,
} from '@heroicons/vue/24/outline'

const billingCatalogStore = useBillingCatalogStore()

void billingCatalogStore.fetchCatalog().catch(() => undefined)

const workspaceTrialLabel = computed(() => billingCatalogStore.trialLabel)
const singleStartingPriceLabel = computed(() =>
  billingCatalogStore.getPlanStartingPriceLabel('single'),
)
const singleMonthlyPriceLabel = computed(() =>
  billingCatalogStore.getPriceLabel('single', 'monthly'),
)
const singleYearlyPriceLabel = computed(() => billingCatalogStore.getPriceLabel('single', 'yearly'))
const teamMonthlyPriceLabel = computed(() => billingCatalogStore.getPriceLabel('team', 'monthly'))
const teamYearlyPriceLabel = computed(() => billingCatalogStore.getPriceLabel('team', 'yearly'))

function selectionVisible(plan: 'single' | 'team', interval: 'monthly' | 'yearly') {
  return !billingCatalogStore.hasLoaded || billingCatalogStore.isSelectionAvailable(plan, interval)
}

const valuePoints = computed(() => [
  `From ${singleStartingPriceLabel.value} for solo use`,
  workspaceTrialLabel.value,
  `Team plan supports up to ${TEAM_PLAN_SEAT_LIMIT} people`,
])

const onboardingSteps = computed(() => [
  {
    step: '1',
    title: 'Create the workspace with Google',
    body: 'No password setup. The owner account is created in one step.',
  },
  {
    step: '2',
    title: `Start the ${workspaceTrialLabel.value}`,
    body: 'Set things up properly before making a billing choice.',
  },
  {
    step: '3',
    title: 'Pick the price that fits',
    body: `Stay solo from ${singleStartingPriceLabel.value} or move to team access from ${billingCatalogStore.getPlanStartingPriceLabel('team')}.`,
  },
])

const featureCards = [
  {
    title: 'Easy to set up',
    body: 'Save company details, payment terms, and branding once so every invoice starts in the right place.',
    icon: SwatchIcon,
  },
  {
    title: 'Fast to use',
    body: 'Reuse saved clients and line items so regular invoices take far less time to put together.',
    icon: DocumentTextIcon,
  },
  {
    title: 'Good for small teams',
    body: `Keep one shared workspace for up to ${TEAM_PLAN_SEAT_LIMIT} people instead of splitting work across separate logins.`,
    icon: UsersIcon,
  },
]

const pricingCards = computed(() =>
  [
    {
      name: 'Single monthly',
      price: singleMonthlyPriceLabel.value,
      summary: 'A very low monthly price for one person running the workspace.',
      note: 'Best for solo use',
      icon: UserIcon,
      bullets: ['1 workspace owner', 'Full invoicing workflow', 'Upgrade to team later'],
      featured: true,
      visible: selectionVisible('single', 'monthly'),
    },
    {
      name: 'Single yearly',
      price: singleYearlyPriceLabel.value,
      summary: 'Annual solo pricing with two months effectively free compared with monthly.',
      note: 'Lower annual cost',
      icon: CalendarDaysIcon,
      bullets: ['1 workspace owner', 'Annual billing', 'Best value for long-term solo use'],
      featured: false,
      visible: selectionVisible('single', 'yearly'),
    },
    {
      name: 'Team monthly',
      price: teamMonthlyPriceLabel.value,
      summary: `Shared workspace access for up to ${TEAM_PLAN_SEAT_LIMIT} people.`,
      note: 'Small-team flexibility',
      icon: UsersIcon,
      bullets: [
        'Owner plus teammates',
        'Shared clients and invoices',
        `Up to ${TEAM_PLAN_SEAT_LIMIT} people total`,
      ],
      featured: false,
      visible: selectionVisible('team', 'monthly'),
    },
    {
      name: 'Team yearly',
      price: teamYearlyPriceLabel.value,
      summary: `Annual team pricing for up to ${TEAM_PLAN_SEAT_LIMIT} people in one workspace.`,
      note: 'Best team value',
      icon: CreditCardIcon,
      bullets: [
        'Owner plus teammates',
        'Annual billing',
        `Up to ${TEAM_PLAN_SEAT_LIMIT} people total`,
      ],
      featured: false,
      visible: selectionVisible('team', 'yearly'),
    },
  ].filter((card) => card.visible),
)

const useCases = [
  {
    title: 'Freelancers and consultants',
    body: 'Keep repeat clients, repeat services, and monthly invoice work in one simple place.',
  },
  {
    title: 'Studios and tiny teams',
    body: 'Start with one owner, then upgrade when you need shared access without moving data around.',
  },
  {
    title: 'Anyone who wants less admin',
    body: 'Write invoices, update them, and export them without juggling separate tools and templates.',
  },
]
</script>

<template>
  <main
    class="min-h-screen bg-[linear-gradient(180deg,#edf5ff_0%,#f6faff_35%,#eef4ff_100%)] text-zinc-900 dark:bg-[linear-gradient(180deg,#0a0f12_0%,#0d1715_35%,#0e1716_100%)] dark:text-zinc-100"
  >
    <div class="relative overflow-hidden">
      <div
        class="absolute inset-x-0 top-0 h-128 bg-[radial-gradient(circle_at_top,#d9eafe_0%,rgba(217,234,254,0.42)_34%,transparent_72%)] dark:bg-[radial-gradient(circle_at_top,rgba(16,185,129,0.18)_0%,rgba(16,185,129,0.06)_34%,transparent_72%)]"
      />
      <div
        class="absolute top-20 -right-24 h-72 w-72 rounded-full bg-sky-200/40 blur-3xl dark:bg-emerald-500/10"
      />
      <div
        class="absolute top-96 -left-16 h-72 w-72 rounded-full bg-blue-200/30 blur-3xl dark:bg-emerald-400/10"
      />

      <section class="relative mx-auto w-full max-w-7xl px-5 py-6 sm:px-8 lg:px-10">
        <header
          class="flex flex-col gap-4 rounded-3xl border border-white/80 bg-white/90 px-4 py-4 shadow-md shadow-sky-100/70 sm:flex-row sm:items-center sm:justify-between sm:px-5 dark:border-white/10 dark:bg-zinc-950/80 dark:shadow-none"
        >
          <div class="flex items-center gap-4">
            <RouterLink
              to="/"
              class="inline-flex items-center gap-3 text-sm font-semibold tracking-[0.2em] text-zinc-900 uppercase dark:text-zinc-100"
            >
              <span class="inline-flex h-2.5 w-2.5 rounded-full bg-sky-500 dark:bg-emerald-400" />
              Invoice and Go
            </RouterLink>
            <DarkMode variant="pill" />
          </div>

          <div class="flex flex-wrap items-center gap-2">
            <RouterLink
              to="/login"
              class="rounded-full border border-zinc-300 bg-white px-4 py-2 text-sm font-medium text-zinc-700 transition hover:border-sky-400 hover:text-zinc-950 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:border-emerald-400/60 dark:hover:text-white"
            >
              Log in
            </RouterLink>
            <RouterLink
              to="/signup"
              class="inline-flex items-center gap-2 rounded-full bg-sky-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-sky-500 dark:bg-emerald-500 dark:hover:bg-emerald-400"
            >
              Create workspace
              <ArrowRightIcon class="size-4" />
            </RouterLink>
          </div>
        </header>

        <section
          class="grid gap-12 py-16 lg:grid-cols-[1.08fr_0.92fr] lg:items-start lg:gap-16 lg:py-24"
        >
          <div class="max-w-3xl">
            <div class="flex flex-wrap gap-2">
              <span
                class="inline-flex items-center rounded-full border border-sky-200 bg-sky-50 px-3 py-1 text-xs font-semibold tracking-[0.18em] text-sky-700 uppercase dark:border-emerald-400/20 dark:bg-emerald-500/10 dark:text-emerald-200"
              >
                {{ workspaceTrialLabel }}
              </span>
              <span
                class="inline-flex items-center rounded-full border border-zinc-200 bg-white px-3 py-1 text-xs font-semibold tracking-[0.18em] text-zinc-700 uppercase dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200"
              >
                From {{ singleStartingPriceLabel }}
              </span>
            </div>

            <h1
              class="mt-6 max-w-4xl text-5xl font-bold tracking-tighter text-zinc-950 sm:text-6xl lg:text-6xl dark:text-white"
            >
              Simple invoicing for solo businesses and small teams.
            </h1>

            <p class="mt-6 max-w-2xl text-lg leading-8 text-zinc-600 sm:text-xl dark:text-zinc-300">
              Create your workspace with Google, start the free trial, and keep clients, invoices,
              exports, and settings in one easy place. It starts at a very small monthly cost and
              works especially well for teams that do not want bloated software.
            </p>

            <div class="mt-8 flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-center">
              <RouterLink
                to="/signup"
                class="inline-flex items-center justify-center gap-2 rounded-full bg-sky-600 px-5 py-3 text-sm font-semibold text-white transition hover:bg-sky-500 dark:bg-emerald-500 dark:hover:bg-emerald-400"
              >
                Start {{ workspaceTrialLabel }}
                <ArrowRightIcon class="size-4" />
              </RouterLink>
              <RouterLink
                to="/login"
                class="inline-flex items-center justify-center rounded-full border border-zinc-300 bg-white px-5 py-3 text-sm font-medium text-zinc-700 transition hover:border-sky-400 hover:text-zinc-950 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:border-emerald-400/60 dark:hover:text-white"
              >
                Log in to your workspace
              </RouterLink>
            </div>

            <div class="mt-10 grid gap-3 sm:grid-cols-3">
              <div
                v-for="point in valuePoints"
                :key="point"
                class="flex items-start gap-3 rounded-2xl border border-sky-100 bg-white p-4 shadow-sm shadow-sky-100/70 dark:border-zinc-800 dark:bg-zinc-900 dark:shadow-none"
              >
                <CheckCircleIcon
                  class="mt-0.5 size-5 shrink-0 text-sky-600 dark:text-emerald-300"
                />
                <p class="text-sm leading-6 text-zinc-700 dark:text-zinc-200">
                  {{ point }}
                </p>
              </div>
            </div>
          </div>

          <aside
            class="rounded-4xl border border-sky-100 bg-white p-6 shadow-xl shadow-sky-100/80 sm:p-8 dark:border-zinc-800 dark:bg-zinc-950/90 dark:shadow-none"
          >
            <div class="flex items-start justify-between gap-4">
              <div>
                <p
                  class="text-xs font-semibold tracking-[0.2em] text-sky-700 uppercase dark:text-emerald-300"
                >
                  Why us
                </p>
                <h2 class="mt-3 text-2xl font-bold text-zinc-950 dark:text-white">
                  Quick setup, easy to use, low cost
                </h2>
              </div>
              <div
                class="rounded-2xl border border-sky-200 bg-sky-50 p-3 text-sky-700 dark:border-emerald-400/20 dark:bg-emerald-500/10 dark:text-emerald-200"
              >
                <ShieldCheckIcon class="size-6" />
              </div>
            </div>

            <div class="mt-8 grid gap-4">
              <article
                v-for="step in onboardingSteps"
                :key="step.title"
                class="rounded-3xl border border-zinc-200 bg-zinc-50 p-5 dark:border-zinc-800 dark:bg-zinc-900"
              >
                <div class="flex items-center gap-3">
                  <div
                    class="flex size-9 items-center justify-center rounded-full bg-sky-100 text-sm font-semibold text-sky-700 dark:bg-emerald-500/15 dark:text-emerald-200"
                  >
                    {{ step.step }}
                  </div>
                  <h3 class="text-base font-semibold text-zinc-950 dark:text-white">
                    {{ step.title }}
                  </h3>
                </div>
                <p class="mt-3 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                  {{ step.body }}
                </p>
              </article>
            </div>

            <div
              class="mt-8 rounded-3xl border border-sky-200 bg-sky-50 p-5 dark:border-emerald-400/20 dark:bg-emerald-500/10"
            >
              <div class="flex items-start gap-3">
                <SparklesIcon class="mt-0.5 size-5 shrink-0 text-sky-700 dark:text-emerald-200" />
                <div>
                  <p class="text-sm font-semibold text-zinc-950 dark:text-white">
                    Especially good for small teams
                  </p>
                  <p class="mt-2 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                    One owner can start alone, then move to team access when the business needs it.
                    No account migration, no complicated setup.
                  </p>
                </div>
              </div>
            </div>
          </aside>
        </section>

        <section
          class="rounded-4xl border border-sky-100 bg-white p-6 shadow-md shadow-sky-100/70 sm:p-8 lg:p-10 dark:border-zinc-800 dark:bg-zinc-950/90 dark:shadow-none"
        >
          <div class="max-w-3xl">
            <p
              class="text-sm font-semibold tracking-[0.18em] text-sky-700 uppercase dark:text-emerald-300"
            >
              What you get
            </p>
            <h2
              class="mt-3 text-3xl font-semibold tracking-tight text-zinc-950 sm:text-4xl dark:text-white"
            >
              Designed to save time, not add more admin
            </h2>
            <p class="mt-4 text-base leading-7 text-zinc-600 sm:text-lg dark:text-zinc-300">
              The workspace is built around the tasks people actually repeat: setting things up,
              building invoices quickly, and keeping a small team in sync.
            </p>
          </div>

          <div class="mt-10 grid gap-5 lg:grid-cols-3">
            <article
              v-for="card in featureCards"
              :key="card.title"
              class="rounded-3xl border border-zinc-200 bg-zinc-50 p-6 dark:border-zinc-800 dark:bg-zinc-900"
            >
              <div
                class="inline-flex items-center justify-center rounded-2xl border border-sky-200 bg-sky-50 p-3 text-sky-700 dark:border-emerald-400/20 dark:bg-emerald-500/10 dark:text-emerald-200"
              >
                <component
                  :is="card.icon"
                  class="size-5"
                />
              </div>
              <h3 class="mt-5 text-xl font-semibold text-zinc-950 dark:text-white">
                {{ card.title }}
              </h3>
              <p class="mt-3 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                {{ card.body }}
              </p>
            </article>
          </div>
        </section>

        <section
          class="mt-14 rounded-4xl border border-sky-100 bg-[linear-gradient(180deg,#ffffff_0%,#f5f9ff_100%)] p-6 shadow-lg shadow-sky-100/70 sm:mt-16 sm:p-8 lg:p-10 dark:border-zinc-800 dark:bg-[linear-gradient(180deg,#0f1615_0%,#101918_100%)] dark:shadow-none"
        >
          <div class="max-w-3xl">
            <p
              class="text-sm font-semibold tracking-[0.18em] text-sky-700 uppercase dark:text-emerald-300"
            >
              Pricing
            </p>
            <h2
              class="mt-3 text-3xl font-semibold tracking-tight text-zinc-950 sm:text-4xl dark:text-white"
            >
              Clear pricing for solo users and small teams
            </h2>
            <p class="mt-4 text-base leading-7 text-zinc-600 sm:text-lg dark:text-zinc-300">
              The monthly fee is intentionally small. Start with the free trial, then pick the
              option that feels right for the way you work.
            </p>
          </div>

          <div class="mt-10 grid gap-5 md:grid-cols-2 xl:grid-cols-4">
            <article
              v-for="card in pricingCards"
              :key="card.name"
              :class="[
                'rounded-3xl border p-6 shadow-sm',
                card.featured
                  ? 'border-sky-300 bg-sky-50 dark:border-emerald-400/30 dark:bg-emerald-500/10'
                  : 'border-zinc-200 bg-white dark:border-zinc-800 dark:bg-zinc-900',
              ]"
            >
              <div class="flex items-start justify-between gap-4">
                <div>
                  <p class="text-sm font-semibold text-zinc-950 dark:text-white">
                    {{ card.name }}
                  </p>
                  <p class="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
                    {{ card.note }}
                  </p>
                </div>
                <div
                  class="rounded-2xl border border-zinc-200 bg-white p-2 text-sky-700 dark:border-zinc-700 dark:bg-zinc-950 dark:text-emerald-200"
                >
                  <component
                    :is="card.icon"
                    class="size-5"
                  />
                </div>
              </div>

              <div class="mt-6 text-3xl font-semibold tracking-tight text-zinc-950 dark:text-white">
                {{ card.price }}
              </div>
              <p class="mt-3 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                {{ card.summary }}
              </p>

              <ul class="mt-6 grid gap-3 text-sm text-zinc-700 dark:text-zinc-200">
                <li
                  v-for="bullet in card.bullets"
                  :key="bullet"
                  class="flex items-start gap-2"
                >
                  <CheckCircleIcon
                    class="mt-0.5 size-4 shrink-0 text-sky-600 dark:text-emerald-300"
                  />
                  <span>{{ bullet }}</span>
                </li>
              </ul>
            </article>
          </div>

          <div
            class="mt-10 flex flex-col gap-6 border-t border-zinc-200 pt-8 sm:flex-row sm:items-center sm:justify-between dark:border-zinc-800"
          >
            <div class="max-w-2xl">
              <p class="text-base font-semibold text-zinc-950 dark:text-white">
                Start with the {{ workspaceTrialLabel }}.
              </p>
              <p class="mt-2 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                Get the workspace set up first, then keep going from
                {{ singleStartingPriceLabel }} for solo use or
                {{ billingCatalogStore.getPlanStartingPriceLabel('team') }} for a team.
              </p>
            </div>

            <div class="flex flex-col gap-3 sm:flex-row">
              <RouterLink
                to="/signup"
                class="inline-flex items-center justify-center gap-2 rounded-full bg-sky-600 px-5 py-3 text-sm font-semibold text-white transition hover:bg-sky-500 dark:bg-emerald-500 dark:hover:bg-emerald-400"
              >
                Create workspace
                <ArrowRightIcon class="size-4" />
              </RouterLink>
              <RouterLink
                to="/login"
                class="inline-flex items-center justify-center rounded-full border border-zinc-300 bg-white px-5 py-3 text-sm font-medium text-zinc-700 transition hover:border-sky-400 hover:text-zinc-950 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:border-emerald-400/60 dark:hover:text-white"
              >
                Log in
              </RouterLink>
            </div>
          </div>
        </section>

        <section class="mt-14 grid gap-6 lg:grid-cols-[0.92fr_1.08fr] lg:gap-8">
          <aside
            class="rounded-4xl border border-zinc-200 bg-zinc-950 p-6 text-white shadow-sm sm:p-8 dark:border-zinc-800 dark:bg-zinc-900"
          >
            <p
              class="text-sm font-semibold tracking-[0.18em] text-sky-200 uppercase dark:text-emerald-200"
            >
              Ease of use
            </p>
            <h2 class="mt-3 text-3xl font-semibold tracking-tight">
              One workspace instead of scattered admin
            </h2>

            <div class="mt-8 grid gap-4">
              <div class="rounded-3xl border border-white/10 bg-white/5 p-5">
                <div class="flex items-center gap-3">
                  <FolderOpenIcon class="size-5 text-sky-200 dark:text-emerald-200" />
                  <p class="font-semibold">Keep clients and items ready</p>
                </div>
                <p class="mt-3 text-sm leading-7 text-zinc-300">
                  Stop rebuilding the same invoice details every time you send a job out.
                </p>
              </div>

              <div class="rounded-3xl border border-white/10 bg-white/5 p-5">
                <div class="flex items-center gap-3">
                  <DocumentTextIcon class="size-5 text-sky-200 dark:text-emerald-200" />
                  <p class="font-semibold">Export without extra steps</p>
                </div>
                <p class="mt-3 text-sm leading-7 text-zinc-300">
                  Generate PDF and DOCX from the same invoice instead of moving data between tools.
                </p>
              </div>

              <div class="rounded-3xl border border-white/10 bg-white/5 p-5">
                <div class="flex items-center gap-3">
                  <PencilSquareIcon class="size-5 text-sky-200 dark:text-emerald-200" />
                  <p class="font-semibold">Update work cleanly</p>
                </div>
                <p class="mt-3 text-sm leading-7 text-zinc-300">
                  Reopen and revise invoices without making the workflow messy.
                </p>
              </div>
            </div>
          </aside>

          <div
            class="rounded-4xl border border-sky-100 bg-white p-6 shadow-md shadow-sky-100/70 sm:p-8 dark:border-zinc-800 dark:bg-zinc-950/90 dark:shadow-none"
          >
            <p
              class="text-sm font-semibold tracking-[0.18em] text-sky-700 uppercase dark:text-emerald-300"
            >
              Good fit for
            </p>
            <h2
              class="mt-3 text-3xl font-semibold tracking-tight text-zinc-950 sm:text-4xl dark:text-white"
            >
              A strong fit for small business and contractors
            </h2>

            <div class="mt-8 grid gap-4">
              <article
                v-for="item in useCases"
                :key="item.title"
                class="rounded-3xl border border-zinc-200 bg-zinc-50 p-5 dark:border-zinc-800 dark:bg-zinc-900"
              >
                <div class="flex items-start gap-3">
                  <div
                    class="rounded-2xl border border-sky-200 bg-sky-50 p-2 text-sky-700 dark:border-emerald-400/20 dark:bg-emerald-500/10 dark:text-emerald-200"
                  >
                    <CheckCircleIcon class="size-5" />
                  </div>
                  <div>
                    <h3 class="text-lg font-semibold text-zinc-950 dark:text-white">
                      {{ item.title }}
                    </h3>
                    <p class="mt-2 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                      {{ item.body }}
                    </p>
                  </div>
                </div>
              </article>
            </div>
          </div>
        </section>

        <section class="py-14 sm:py-18">
          <div
            class="rounded-4xl border border-sky-200 bg-[linear-gradient(135deg,#0f172a_0%,#15304f_56%,#1d4d7b_100%)] p-6 text-white shadow-xl shadow-sky-200/50 sm:p-8 lg:p-10 dark:border-emerald-400/15 dark:bg-[linear-gradient(135deg,#0b1412_0%,#10201a_56%,#163127_100%)] dark:shadow-none"
          >
            <div class="grid gap-6 lg:grid-cols-[1fr_auto] lg:items-end">
              <div class="max-w-3xl">
                <p
                  class="text-sm font-semibold tracking-[0.18em] text-sky-200 uppercase dark:text-emerald-200"
                >
                  Ready to start
                </p>
                <h2 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl">
                  Create the workspace, try it properly, then keep it for a very small monthly fee.
                </h2>
                <p
                  class="mt-4 text-base leading-7 text-sky-50/85 sm:text-lg dark:text-emerald-50/85"
                >
                  The flow is simple: Google sign-in first, trial next, pricing choice after that.
                </p>
              </div>

              <div class="flex flex-col gap-3 sm:flex-row lg:flex-col">
                <RouterLink
                  to="/signup"
                  class="inline-flex items-center justify-center gap-2 rounded-full bg-white px-5 py-3 text-sm font-semibold text-sky-900 transition hover:bg-sky-50 dark:bg-emerald-500 dark:text-zinc-950 dark:hover:bg-emerald-400"
                >
                  Start {{ workspaceTrialLabel }}
                  <ArrowRightIcon class="size-4" />
                </RouterLink>
                <RouterLink
                  to="/privacy"
                  class="inline-flex items-center justify-center rounded-full border border-white/20 px-5 py-3 text-sm font-medium text-white transition hover:border-white/40 hover:bg-white/5 dark:border-emerald-400/20 dark:hover:border-emerald-300/40 dark:hover:bg-emerald-500/10"
                >
                  Privacy and cookies
                </RouterLink>
              </div>
            </div>
          </div>
        </section>
      </section>
    </div>
  </main>
</template>

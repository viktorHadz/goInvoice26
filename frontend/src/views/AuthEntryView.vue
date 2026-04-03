<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { TEAM_PLAN_SEAT_LIMIT } from '@/constants/billing'
import { useBillingCatalogStore } from '@/stores/billingCatalog'
import {
  ArrowLeftIcon,
  CheckCircleIcon,
  CreditCardIcon,
  GlobeAltIcon,
  ShieldCheckIcon,
  SparklesIcon,
  UsersIcon,
} from '@heroicons/vue/24/outline'

const props = defineProps<{
  mode: 'login' | 'signup'
}>()

const route = useRoute()
const authStore = useAuthStore()
const billingCatalogStore = useBillingCatalogStore()

void billingCatalogStore.fetchCatalog().catch(() => undefined)

const workspaceTrialLabel = computed(() => billingCatalogStore.trialLabel)
const singlePricingSummary = computed(() => billingCatalogStore.getPlanPricingSummary('single'))
const teamPricingSummary = computed(() => billingCatalogStore.getPlanPricingSummary('team'))

const signupSteps = computed(() => [
  {
    title: 'Sign in with Google',
    body: 'We create the owner account and the workspace in one step.',
  },
  {
    title: `Start the ${workspaceTrialLabel.value}`,
    body: 'You can set things up and try it before paying.',
  },
  {
    title: 'Choose solo or team pricing',
    body: `Keep going from a very small fee: solo from ${billingCatalogStore.getPlanStartingPriceLabel('single')} or team access for up to ${TEAM_PLAN_SEAT_LIMIT} people from ${billingCatalogStore.getPlanStartingPriceLabel('team')}.`,
  },
])

const loginChecks = [
  'Use the Google account already linked to the workspace.',
  'If you were invited, use that same email with Google.',
  'If you are brand new, create a workspace first.',
]

const signupNotes = computed(() => [
  'No password setup needed.',
  `Solo pricing is ${singlePricingSummary.value}.`,
  `Team pricing is ${teamPricingSummary.value} for up to ${TEAM_PLAN_SEAT_LIMIT} people.`,
  'You can start solo and upgrade later.',
])

const isSignup = computed(() => props.mode === 'signup')
const title = computed(() =>
  isSignup.value ? 'Create your workspace' : 'Log in to your workspace',
)
const subtitle = computed(() =>
  isSignup.value
    ? `Use Google to create your workspace. Start the ${workspaceTrialLabel.value}, then choose solo pricing at ${singlePricingSummary.value} or team pricing at ${teamPricingSummary.value}.`
    : 'Use the Google account that already has access to the workspace.',
)
const redirectPath = computed(() =>
  typeof route.query.redirect === 'string' && route.query.redirect.startsWith('/')
    ? route.query.redirect
    : '/app',
)
const googleDisabled = computed(
  () => !authStore.googleEnabled || (isSignup.value && !authStore.canRegister),
)
const googleButtonLabel = computed(() => {
  if (!authStore.googleEnabled) return 'Google sign-in unavailable'
  return isSignup.value ? 'Create workspace with Google' : 'Log in with Google'
})
const errorCode = computed(() => (typeof route.query.error === 'string' ? route.query.error : ''))
const errorMessage = computed(() => authErrorMessage(errorCode.value))

function startGoogleAuth() {
  if (googleDisabled.value) return
  authStore.beginGoogleAuth(props.mode, redirectPath.value)
}

function authErrorMessage(code: string) {
  const messages: Record<string, string> = {
    google_not_configured: 'Sign-in is temporarily unavailable. Please try again later.',
    invalid_auth_mode: 'That sign-in request could not be started. Please try again.',
    invalid_oauth_state: 'We could not verify the Google sign-in. Please try again.',
    missing_oauth_code: 'Google did not return a sign-in code. Please try again.',
    google_access_denied: 'Google sign-in was cancelled before it completed.',
    google_auth_failed: 'Google sign-in did not complete. Please try again.',
    google_email_not_verified: 'The selected Google account needs a verified email address.',
    account_not_linked:
      'This Google account does not have access yet. Create a workspace or ask your admin for an invite.',
    account_conflict:
      'This email is already linked to a different Google account. Please use the original login or contact support.',
  }

  return code ? (messages[code] ?? 'Authentication could not be completed. Please try again.') : ''
}
</script>

<template>
  <main
    class="min-h-screen bg-[linear-gradient(180deg,#edf5ff_0%,#f7fbff_42%,#eef4ff_100%)] text-zinc-900 dark:bg-[linear-gradient(180deg,#0a0f12_0%,#0d1715_42%,#0e1716_100%)] dark:text-zinc-100"
  >
    <section
      class="mx-auto flex min-h-screen w-full max-w-6xl items-center px-5 py-10 sm:px-8 sm:py-12"
    >
      <div class="grid w-full gap-8 lg:grid-cols-[0.92fr_1.08fr] lg:gap-10">
        <aside
          class="rounded-4xl border border-sky-100 bg-white p-6 shadow-xl shadow-sky-100/80 sm:p-8 dark:border-zinc-800 dark:bg-zinc-950/90 dark:shadow-none"
        >
          <div class="flex flex-wrap items-center justify-between gap-3">
            <RouterLink
              to="/"
              class="inline-flex items-center gap-2 text-sm font-medium text-zinc-600 transition hover:text-zinc-900 dark:text-zinc-300 dark:hover:text-white"
            >
              <ArrowLeftIcon class="size-4" />
              Back to homepage
            </RouterLink>

            <div
              class="inline-flex items-center gap-2 rounded-full border border-sky-200 bg-sky-50 px-3 py-1 text-xs font-semibold text-sky-700 uppercase dark:border-emerald-400/20 dark:bg-emerald-500/10 dark:text-emerald-200"
            >
              <ShieldCheckIcon class="size-4" />
              Secure Google sign-in
            </div>
          </div>

          <div class="mt-6 flex flex-wrap gap-2">
            <span
              v-if="isSignup"
              class="inline-flex items-center rounded-full border border-sky-200 bg-sky-50 px-3 py-1 text-xs font-semibold tracking-[0.18em] text-sky-700 uppercase dark:border-emerald-400/20 dark:bg-emerald-500/10 dark:text-emerald-200"
            >
              {{ workspaceTrialLabel }}
            </span>
            <span
              class="inline-flex items-center rounded-full border border-zinc-200 bg-zinc-50 px-3 py-1 text-xs font-semibold tracking-[0.18em] text-zinc-700 uppercase dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200"
            >
              {{ isSignup ? 'New workspace' : 'Existing workspace access' }}
            </span>
          </div>

          <h1
            class="mt-5 text-4xl leading-tight font-semibold tracking-tight text-zinc-950 sm:text-5xl dark:text-white"
          >
            {{ title }}
          </h1>

          <p class="mt-4 max-w-xl text-base leading-8 text-zinc-600 sm:text-lg dark:text-zinc-300">
            {{ subtitle }}
          </p>

          <div
            v-if="errorMessage"
            class="mt-6 rounded-3xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm leading-6 text-rose-700 dark:border-rose-400/20 dark:bg-rose-500/10 dark:text-rose-200"
          >
            {{ errorMessage }}
          </div>

          <div class="mt-8 grid gap-3">
            <article
              v-for="note in isSignup ? signupNotes : loginChecks"
              :key="note"
              class="rounded-3xl border border-zinc-200 bg-zinc-50 p-4 dark:border-zinc-800 dark:bg-zinc-900"
            >
              <div class="flex items-start gap-3">
                <CheckCircleIcon
                  class="mt-0.5 size-4 shrink-0 text-sky-600 dark:text-emerald-300"
                />
                <span class="text-sm leading-7 text-zinc-700 dark:text-zinc-200">{{ note }}</span>
              </div>
            </article>
          </div>

          <div class="mt-8 flex flex-wrap gap-3">
            <RouterLink
              v-if="isSignup"
              to="/login"
              class="inline-flex items-center justify-center rounded-full border border-zinc-300 bg-white px-4 py-2 text-sm font-medium text-zinc-700 transition hover:border-sky-400 hover:text-zinc-950 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:border-emerald-400/60 dark:hover:text-white"
            >
              Already have access? Log in
            </RouterLink>
            <RouterLink
              v-else-if="authStore.canRegister"
              to="/signup"
              class="inline-flex items-center justify-center rounded-full border border-zinc-300 bg-white px-4 py-2 text-sm font-medium text-zinc-700 transition hover:border-sky-400 hover:text-zinc-950 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:border-emerald-400/60 dark:hover:text-white"
            >
              Need a workspace? Create one
            </RouterLink>
          </div>

          <p class="mt-8 text-sm leading-7 text-zinc-500 dark:text-zinc-400">
            We use a necessary secure cookie to keep you signed in after Google authentication.
            <RouterLink
              to="/privacy"
              class="ml-1 font-semibold text-zinc-700 underline decoration-zinc-400 underline-offset-2 dark:text-zinc-200 dark:decoration-zinc-600"
            >
              Privacy and cookies
            </RouterLink>
          </p>
        </aside>

        <section class="grid gap-5">
          <article
            class="rounded-4xl border border-sky-100 bg-[linear-gradient(180deg,#ffffff_0%,#f4f9ff_100%)] p-6 shadow-xl shadow-sky-100/80 sm:p-8 dark:border-zinc-800 dark:bg-[linear-gradient(180deg,#101817_0%,#121c1a_100%)] dark:shadow-none"
          >
            <div class="flex flex-wrap items-start justify-between gap-4">
              <div>
                <div class="flex flex-wrap gap-2">
                  <span
                    class="inline-flex items-center rounded-full border border-sky-200 bg-sky-50 px-3 py-1 text-xs font-semibold tracking-[0.18em] text-sky-700 uppercase dark:border-emerald-400/20 dark:bg-emerald-500/10 dark:text-emerald-200"
                  >
                    Google sign-in
                  </span>
                  <span
                    v-if="isSignup"
                    class="inline-flex items-center rounded-full border border-zinc-200 bg-white px-3 py-1 text-xs font-semibold tracking-[0.18em] text-zinc-700 uppercase dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200"
                  >
                    {{ workspaceTrialLabel }}
                  </span>
                </div>

                <h2
                  class="mt-4 text-3xl font-semibold tracking-tight text-zinc-950 dark:text-white"
                >
                  {{ isSignup ? 'What happens next' : 'Use the same Google account as before' }}
                </h2>
                <p
                  class="mt-3 max-w-2xl text-sm leading-7 text-zinc-600 sm:text-base dark:text-zinc-300"
                >
                  {{
                    isSignup
                      ? 'Create the workspace first. Then you can try it, set things up, and choose the pricing option that fits.'
                      : 'Logging in should be simple. If Google recognises the same account, we restore your workspace session.'
                  }}
                </p>
              </div>

              <div
                class="rounded-2xl border border-sky-200 bg-sky-50 p-3 text-sky-700 dark:border-emerald-400/20 dark:bg-emerald-500/10 dark:text-emerald-200"
              >
                <GlobeAltIcon class="size-6" />
              </div>
            </div>

            <div
              v-if="isSignup"
              class="mt-8 grid gap-4 sm:grid-cols-3"
            >
              <article
                v-for="step in signupSteps"
                :key="step.title"
                class="rounded-3xl border border-zinc-200 bg-white p-5 dark:border-zinc-800 dark:bg-zinc-900"
              >
                <div
                  class="flex items-center gap-2 text-sm font-semibold text-zinc-950 dark:text-white"
                >
                  <SparklesIcon class="size-4 text-sky-600 dark:text-emerald-300" />
                  {{ step.title }}
                </div>
                <p class="mt-3 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                  {{ step.body }}
                </p>
              </article>
            </div>

            <div
              v-else
              class="mt-8 grid gap-4 sm:grid-cols-3"
            >
              <article
                v-for="item in loginChecks"
                :key="item"
                class="rounded-3xl border border-zinc-200 bg-white p-5 dark:border-zinc-800 dark:bg-zinc-900"
              >
                <div class="flex items-start gap-3">
                  <CheckCircleIcon
                    class="mt-0.5 size-4 shrink-0 text-sky-600 dark:text-emerald-300"
                  />
                  <p class="text-sm leading-7 text-zinc-700 dark:text-zinc-200">
                    {{ item }}
                  </p>
                </div>
              </article>
            </div>

            <div
              v-if="isSignup"
              class="mt-8 grid gap-4 sm:grid-cols-2"
            >
              <article
                class="rounded-3xl border border-sky-200 bg-sky-50 p-5 dark:border-emerald-400/20 dark:bg-emerald-500/10"
              >
                <div class="flex items-center gap-3">
                  <div
                    class="rounded-2xl border border-white/70 bg-white p-2 text-sky-700 dark:border-zinc-700 dark:bg-zinc-950 dark:text-emerald-200"
                  >
                    <CreditCardIcon class="size-5" />
                  </div>
                  <div>
                    <p class="text-sm font-semibold text-zinc-950 dark:text-white">Solo pricing</p>
                    <p class="text-sm text-zinc-600 dark:text-zinc-300">
                      {{ singlePricingSummary }}
                    </p>
                  </div>
                </div>
                <p class="mt-3 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                  A small monthly cost for one person running the workspace.
                </p>
              </article>

              <article
                class="rounded-3xl border border-zinc-200 bg-white p-5 dark:border-zinc-800 dark:bg-zinc-900"
              >
                <div class="flex items-center gap-3">
                  <div
                    class="rounded-2xl border border-zinc-200 bg-zinc-50 p-2 text-zinc-700 dark:border-zinc-700 dark:bg-zinc-950 dark:text-emerald-200"
                  >
                    <UsersIcon class="size-5" />
                  </div>
                  <div>
                    <p class="text-sm font-semibold text-zinc-950 dark:text-white">Team pricing</p>
                    <p class="text-sm text-zinc-600 dark:text-zinc-300">{{ teamPricingSummary }}</p>
                  </div>
                </div>
                <p class="mt-3 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                  Shared workspace access for up to {{ TEAM_PLAN_SEAT_LIMIT }} people.
                </p>
              </article>
            </div>

            <div class="mt-8 flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-center">
              <button
                type="button"
                :disabled="googleDisabled"
                class="inline-flex items-center justify-center rounded-full bg-sky-600 px-5 py-3 text-sm font-semibold text-white transition hover:bg-sky-500 disabled:cursor-not-allowed disabled:opacity-60 dark:bg-emerald-500 dark:text-zinc-950 dark:hover:bg-emerald-400"
                @click="startGoogleAuth"
              >
                {{ googleButtonLabel }}
              </button>

              <span
                class="inline-flex items-center justify-center rounded-full border border-zinc-200 bg-white px-4 py-3 text-xs font-medium text-zinc-600 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-300"
              >
                Secure cookie keeps you signed in after Google authentication
              </span>
            </div>
          </article>

          <article
            class="rounded-4xl border border-sky-100 bg-white p-6 shadow-md shadow-sky-100/70 sm:p-7 dark:border-zinc-800 dark:bg-zinc-950/90 dark:shadow-none"
          >
            <div class="flex items-start gap-3">
              <div
                class="rounded-2xl border border-sky-200 bg-sky-50 p-3 text-sky-700 dark:border-emerald-400/20 dark:bg-emerald-500/10 dark:text-emerald-200"
              >
                <ShieldCheckIcon class="size-5" />
              </div>
              <div>
                <h3 class="text-lg font-semibold text-zinc-950 dark:text-white">
                  {{ isSignup ? 'Simple for small teams' : 'Simple to get back in' }}
                </h3>
                <p class="mt-2 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                  {{
                    isSignup
                      ? `Start on your own, then upgrade later if you need shared access for up to ${TEAM_PLAN_SEAT_LIMIT} people.`
                      : 'If you already belong to a workspace, use that same Google account and we bring you back in.'
                  }}
                </p>
              </div>
            </div>
          </article>
        </section>
      </div>
    </section>
  </main>
</template>

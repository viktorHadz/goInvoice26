<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { TEAM_PLAN_SEAT_LIMIT } from '@/constants/billing'
import { useBillingCatalogStore } from '@/stores/billingCatalog'
import {
  ArrowLeftIcon,
  CheckCircleIcon,
  GlobeAltIcon,
  ShieldCheckIcon,
} from '@heroicons/vue/24/outline'
import LandingHeader from '@/components/landing/LandingHeader.vue'

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
    body: 'Create your workspace with the Google account you want to use as the owner.',
  },
  {
    title: 'Set things up during the trial',
  },
  {
    title: 'Stay solo or add your team later',
    body: `Solo is ${singlePricingSummary.value}. Team access is ${teamPricingSummary.value} for up to ${TEAM_PLAN_SEAT_LIMIT} people.`,
  },
])

const loginChecks = [
  'Use the Google account already linked to the workspace.',
  'If you were invited, sign in with that same email.',
  'If you are new here, create a workspace first.',
]

const isSignup = computed(() => props.mode === 'signup')

const title = computed(() =>
  isSignup.value ? 'Create your workspace' : 'Log in to your workspace',
)

const subtitle = computed(() =>
  isSignup.value
    ? 'Use Google to get started. We will create your workspace and take you into setup.'
    : 'Use the Google account that already has access.',
)

const helperTitle = computed(() => (isSignup.value ? 'What happens next' : 'Before you continue'))

const helperCopy = computed(() =>
  isSignup.value
    ? `After account creation you will be redirected to the billing screen where you can start a trial or chose a payment plan. If anything blocks setup, email invoiceandgo@gmail.com.`
    : 'Make sure you use the same Google account that already belongs to the workspace.',
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
  return isSignup.value ? 'Continue with Google' : 'Log in with Google'
})

const switchLinkLabel = computed(() =>
  isSignup.value ? 'Already have access? Log in' : 'Need a workspace? Create one',
)

const switchLinkTo = computed(() => (isSignup.value ? '/login' : '/signup'))

const showSwitchLink = computed(() => isSignup.value || authStore.canRegister)

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
    class="min-h-screen bg-linear-to-b from-sky-50 via-white to-slate-100 text-zinc-900 dark:from-[#08100f] dark:via-[#0a1211] dark:to-[#0d1715] dark:text-zinc-100">
    <LandingHeader class="sticky top-0 z-10 mx-auto" />

    <div class="relative isolate min-h-screen overflow-hidden">
      <div class="hdr-grid pointer-events-none absolute inset-0 opacity-50 dark:opacity-100" />
      <div
        class="pointer-events-none absolute inset-x-0 top-0 h-72 bg-radial from-sky-200/35 via-transparent to-transparent dark:from-emerald-500/10" />
      <div
        class="pointer-events-none absolute top-20 left-1/2 h-64 w-64 -translate-x-1/2 rounded-full bg-sky-200/25 blur-3xl dark:bg-emerald-500/10" />

      <section class="relative mx-auto flex min-h-screen w-full max-w-6xl items-center px-5 py-8 sm:px-8 sm:py-10">
        <div class="grid w-full gap-6 lg:grid-cols-[0.92fr_1.08fr] lg:gap-8">
          <aside
            class="order-2 rounded-3xl border border-white/80 bg-white/90 p-6 shadow-sm backdrop-blur sm:p-8 lg:order-1 dark:border-white/10 dark:bg-zinc-950/80">
            <div class="flex items-center justify-between gap-3">
              <RouterLink to="/"
                class="inline-flex items-center gap-2 text-sm font-medium text-zinc-600 transition hover:text-zinc-950 dark:text-zinc-300 dark:hover:text-white">
                <ArrowLeftIcon class="size-4" />
                Back to homepage
              </RouterLink>

              <span
                class="hidden items-center gap-2 rounded-full border border-sky-200 bg-sky-50 px-3 py-1 text-xs font-semibold text-sky-700 sm:inline-flex dark:border-emerald-400/20 dark:bg-emerald-500/10 dark:text-emerald-200">
                <ShieldCheckIcon class="size-4" />
                Secure Google sign-in
              </span>
            </div>

            <div class="mt-8">
              <div class="flex flex-wrap gap-2"></div>

              <h1 class="mt-5 max-w-lg text-4xl font-semibold tracking-tight text-zinc-950 dark:text-white">
                {{ title }}
              </h1>

              <p class="mt-4 max-w-xl text-base leading-8 text-zinc-600 dark:text-zinc-300">
                {{ subtitle }}
              </p>
            </div>

            <div v-if="errorMessage"
              class="mt-6 rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm leading-6 text-rose-700 dark:border-rose-400/20 dark:bg-rose-500/10 dark:text-rose-200">
              {{ errorMessage }}
            </div>

            <div class="mt-6 flex flex-wrap gap-3">
              <RouterLink v-if="showSwitchLink" :to="switchLinkTo"
                class="inline-flex items-center justify-center rounded-full border border-zinc-300 bg-white px-4 py-2.5 text-sm font-medium text-zinc-700 transition hover:border-sky-400 hover:text-zinc-950 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:border-emerald-400/60 dark:hover:text-white">
                {{ switchLinkLabel }}
              </RouterLink>
            </div>

            <p class="mt-6 text-sm leading-7 text-zinc-500 dark:text-zinc-400">
              We use a necessary secure cookie to keep you signed in after Google authentication.
              <RouterLink to="/privacy"
                class="ml-1 font-semibold text-zinc-700 underline decoration-zinc-400 underline-offset-2 dark:text-zinc-200 dark:decoration-zinc-600">
                Privacy and cookies
              </RouterLink>
            </p>
          </aside>

          <section class="order-1 lg:order-2">
            <article
              class="rounded-3xl border border-white/80 bg-white/90 p-6 shadow-sm backdrop-blur sm:p-8 dark:border-white/10 dark:bg-zinc-950/80">
              <div class="flex items-start justify-between gap-4">
                <div>
                  <p
                    class="inline-flex items-center rounded-full border border-sky-200 bg-sky-50 px-3 py-1 text-[11px] font-semibold tracking-[0.16em] text-sky-700 uppercase dark:border-emerald-400/20 dark:bg-emerald-500/10 dark:text-emerald-200">
                    {{ isSignup ? 'Google workspace setup' : 'Google workspace access' }}
                  </p>

                  <h2 class="mt-4 text-2xl font-semibold tracking-tight text-zinc-950 sm:text-3xl dark:text-white">
                    {{ isSignup ? 'Continue with Google' : 'Sign in with Google' }}
                  </h2>

                  <p class="mt-3 max-w-xl text-sm leading-7 text-zinc-600 sm:text-base dark:text-zinc-300">
                    {{
                      isSignup
                        ? 'Create the workspace with one Google sign-in.'
                        : 'Use the Google account that already has access.'
                    }}
                  </p>
                </div>

                <div
                  class="hidden rounded-2xl border border-sky-200 bg-sky-50 p-3 text-sky-700 sm:flex dark:border-emerald-400/20 dark:bg-emerald-500/10 dark:text-emerald-200">
                  <GlobeAltIcon class="size-5" />
                </div>
              </div>

              <div class="mt-8 space-y-3">
                <button type="button" :disabled="googleDisabled"
                  class="inline-flex min-h-12 w-full cursor-pointer items-center justify-center rounded-full bg-sky-600 px-5 py-3 text-sm font-semibold text-white transition hover:bg-sky-500 disabled:cursor-not-allowed disabled:opacity-60 dark:bg-emerald-500 dark:text-zinc-950 dark:hover:bg-emerald-400"
                  @click="startGoogleAuth">
                  {{ googleButtonLabel }}
                </button>
              </div>
            </article>
            <div
              class="mt-8 rounded-3xl border border-zinc-200 bg-zinc-50/80 p-4 dark:border-zinc-800 dark:bg-zinc-900/80">
              <h2 class="text-sm font-semibold text-zinc-950 dark:text-white">
                {{ helperTitle }}
              </h2>

              <p class="mt-2 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                {{ helperCopy }}
              </p>

              <div class="mt-4 space-y-3">
                <div v-for="item in isSignup ? signupSteps : loginChecks"
                  :key="typeof item === 'string' ? item : item.title"
                  class="rounded-2xl border border-white/80 bg-white px-4 py-4 dark:border-zinc-800 dark:bg-zinc-950/80">
                  <template v-if="typeof item === 'string'">
                    <div class="flex items-start gap-3">
                      <CheckCircleIcon class="mt-0.5 size-4 shrink-0 text-sky-600 dark:text-emerald-300" />
                      <p class="text-sm leading-7 text-zinc-700 dark:text-zinc-200">
                        {{ item }}
                      </p>
                    </div>
                  </template>

                  <template v-else>
                    <p class="text-sm font-semibold text-zinc-950 dark:text-white">
                      {{ item.title }}
                    </p>
                    <p class="mt-1 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                      {{ item.body }}
                    </p>
                  </template>
                </div>
              </div>

              <p v-if="isSignup" class="mt-4 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                Support email:
                <a href="mailto:invoiceandgo@gmail.com"
                  class="font-semibold text-zinc-800 underline decoration-zinc-400 underline-offset-2 dark:text-zinc-100 dark:decoration-zinc-600">
                  invoiceandgo@gmail.com
                </a>
              </p>
            </div>
          </section>
        </div>
      </section>
    </div>
  </main>
</template>

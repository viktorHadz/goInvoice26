<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  CheckCircleIcon,
  CreditCardIcon,
  ExclamationTriangleIcon,
  SparklesIcon,
  UsersIcon,
} from '@heroicons/vue/24/outline'
import TheButton from '@/components/UI/TheButton.vue'
import TheInput from '@/components/UI/TheInput.vue'
import { useAuthStore } from '@/stores/auth'
import { useBillingCatalogStore } from '@/stores/billingCatalog'
import {
  BILLING_INTERVAL_OPTIONS,
  BILLING_PLAN_OPTIONS,
  DEFAULT_BILLING_INTERVAL,
  DEFAULT_BILLING_PLAN,
  formatBillingSelectionName,
  getBillingPlanOption,
  isBillingInterval,
  isBillingPlan,
  type BillingInterval,
  type BillingPlan,
} from '@/constants/billing'
import {
  changeSubscriptionPlan,
  createCheckoutSession,
  createPortalSession,
  redeemPromoCode,
  syncCheckoutSession,
} from '@/utils/billingHttpHandler'
import { emitToastError, emitToastSuccess } from '@/utils/toast'
import { getApiErrorMessage, isApiError } from '@/utils/apiErrors'
import { handleActionError } from '@/utils/errors/handleActionError'
import { fmtDisplayDate, fmtStrDate } from '@/utils/dates'
import { useSettingsStore } from '@/stores/settings'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const billingCatalogStore = useBillingCatalogStore()
void billingCatalogStore.fetchCatalog().catch(() => undefined)

const isSubmittingPlan = ref(false)
const isOpeningPortal = ref(false)
const isSyncingCheckout = ref(false)
const isRedeemingPromo = ref(false)
const syncedSessionId = ref('')
const selectedPlan = ref<BillingPlan>(DEFAULT_BILLING_PLAN)
const selectedInterval = ref<BillingInterval>(DEFAULT_BILLING_INTERVAL)
const promoCode = ref('')
const promoFieldErrors = ref<Record<string, string>>({})
const CHECKOUT_SYNC_ATTEMPTS = 6
const CHECKOUT_SYNC_DELAY_MS = 1200

const billing = computed(() => authStore.billing)
const isOwner = computed(() => authStore.canManageBilling)
const hasAccess = computed(() => authStore.hasBillingAccess)
const accessSource = computed(() => billing.value?.accessSource ?? '')
const currentPlan = computed(() => billing.value?.plan ?? '')
const currentInterval = computed(() => billing.value?.interval ?? '')
const checkoutState = computed(() =>
  typeof route.query.checkout === 'string' ? route.query.checkout : '',
)
const checkoutSessionId = computed(() =>
  typeof route.query.session_id === 'string' ? route.query.session_id : '',
)
const trialDays = computed(() =>
  Math.max(billing.value?.trialDays ?? billingCatalogStore.trialDays, 0),
)
const redirectPath = computed(() => {
  const candidate = typeof route.query.redirect === 'string' ? route.query.redirect : '/app'
  return candidate.startsWith('/') ? candidate : '/app'
})
const selectedPlanDetails = computed(() => getBillingPlanOption(selectedPlan.value))
const selectedPriceLabel = computed(() =>
  billingCatalogStore.getPriceLabel(selectedPlan.value, selectedInterval.value),
)
const currentPlanDetails = computed(() =>
  isBillingPlan(currentPlan.value) ? getBillingPlanOption(currentPlan.value) : null,
)
const currentPriceLabel = computed(() =>
  isBillingPlan(currentPlan.value) && isBillingInterval(currentInterval.value)
    ? billingCatalogStore.getPriceLabel(currentPlan.value, currentInterval.value)
    : '',
)
const intervalCards = computed(() =>
  BILLING_INTERVAL_OPTIONS.map((option) => ({
    ...option,
    selected: selectedInterval.value === option.id,
  })),
)
const planCards = computed(() =>
  BILLING_PLAN_OPTIONS.map((option) => ({
    ...option,
    available: isBillingSelectionAvailable(option.id, selectedInterval.value),
    isCurrent: currentPlan.value === option.id && currentInterval.value === selectedInterval.value,
    priceLabel: billingCatalogStore.getPriceLabel(option.id, selectedInterval.value),
  })),
)
const selectedPlanAvailable = computed(() =>
  isBillingSelectionAvailable(selectedPlan.value, selectedInterval.value),
)
const isSwitchingSelection = computed(
  () =>
    hasAccess.value &&
    !!currentPlan.value &&
    !!currentInterval.value &&
    (selectedPlan.value !== currentPlan.value || selectedInterval.value !== currentInterval.value),
)
const selectedSelectionLabel = computed(() =>
  formatBillingSelectionName(selectedPlan.value, selectedInterval.value),
)
const selectedOfferLabel = computed(() =>
  trialDays.value > 0
    ? `${trialDays.value}-day free trial, then ${selectedPriceLabel.value}`
    : selectedPriceLabel.value,
)
const selectionActionLabel = computed(() => {
  if (isSwitchingSelection.value) {
    return `Switch to ${selectedSelectionLabel.value}`
  }
  if (trialDays.value > 0) {
    return `Start ${trialDays.value}-day free trial on ${selectedSelectionLabel.value}`
  }
  return `Start ${selectedSelectionLabel.value}`
})
const showPlanAction = computed(
  () =>
    isOwner.value &&
    selectedPlanAvailable.value &&
    (!hasAccess.value || accessSource.value === 'subscription') &&
    (!hasAccess.value ||
      selectedPlan.value !== currentPlan.value ||
      selectedInterval.value !== currentInterval.value),
)
const showPromoForm = computed(() => isOwner.value && !hasAccess.value)
const statusLabel = computed(() => {
  if (billing.value?.promoExpired) {
    return 'Promotional period expired'
  }
  if (accessSource.value === 'promo' && hasAccess.value) {
    return 'Promotional access active'
  }
  if (accessSource.value === 'direct' && hasAccess.value) {
    return 'Direct access active'
  }
  switch (billing.value?.status) {
    case 'active':
      return 'Subscription active'
    case 'trialing':
      return 'Trial active'
    case 'past_due':
      return 'Payment required'
    case 'canceled':
      return 'Subscription canceled'
    case 'unpaid':
      return 'Payment failed'
    case 'incomplete':
    case 'checkout_incomplete':
      return 'Checkout incomplete'
    default:
      return 'Subscription required'
  }
})
const periodEndLabel = computed(() =>
  billing.value?.status === 'trialing' ? 'Trial ends' : 'Current period end',
)

watch(
  [
    () => route.query.plan,
    () => route.query.interval,
    currentPlan,
    currentInterval,
    () => billing.value?.singleMonthlyAvailable,
    () => billing.value?.singleYearlyAvailable,
    () => billing.value?.teamMonthlyAvailable,
    () => billing.value?.teamYearlyAvailable,
  ],
  () => {
    const nextPlan = preferredPlanSelection()
    selectedPlan.value = nextPlan
    selectedInterval.value = preferredIntervalSelection(nextPlan)
  },
  { immediate: true },
)

watch(
  checkoutSessionId,
  async (sessionId) => {
    if (!sessionId || syncedSessionId.value === sessionId || !isOwner.value) return

    syncedSessionId.value = sessionId
    await confirmCheckoutSession(sessionId)
  },
  { immediate: true },
)

async function submitPlanAction() {
  if (!selectedPlanAvailable.value) {
    emitToastError({
      title: 'Selection unavailable',
      message: 'That billing selection is not available yet.',
    })
    return
  }

  isSubmittingPlan.value = true
  try {
    if (isSwitchingSelection.value) {
      await changeSubscriptionPlan(selectedPlan.value, selectedInterval.value)
      await authStore.fetchSession(true)
      emitToastSuccess(`${selectedSelectionLabel.value} activated.`)
      return
    }

    const session = await createCheckoutSession(
      selectedPlan.value,
      selectedInterval.value,
      redirectPath.value,
    )
    window.location.assign(session.url)
  } catch (err) {
    emitToastError({
      title: isSwitchingSelection.value ? 'Could not change billing' : 'Could not start checkout',
      message: isApiError(err)
        ? getApiErrorMessage(err)
        : isSwitchingSelection.value
          ? 'The billing selection could not be changed right now.'
          : 'Stripe checkout could not be started right now.',
    })
  } finally {
    isSubmittingPlan.value = false
  }
}

async function openPortal() {
  isOpeningPortal.value = true
  try {
    const session = await createPortalSession()
    window.location.assign(session.url)
  } catch (err) {
    emitToastError({
      title: 'Could not open billing portal',
      message: isApiError(err)
        ? getApiErrorMessage(err)
        : 'The billing portal is not available right now.',
    })
  } finally {
    isOpeningPortal.value = false
  }
}
async function submitPromoCode() {
  promoFieldErrors.value = {}
  isRedeemingPromo.value = true
  try {
    const result = await redeemPromoCode(promoCode.value)
    promoCode.value = ''
    await authStore.fetchSession(true)
    const expires = 'Promo access is active until ' + fmtDisplayDate(new Date(result.expiresAt))
    emitToastSuccess(`${expires}.`)
    if (authStore.hasBillingAccess) {
      await router.replace(redirectPath.value)
    }
  } catch (err) {
    handleActionError(err, {
      fieldErrors: promoFieldErrors,
      toastTitle: 'Could not redeem promo code',
    })
  } finally {
    isRedeemingPromo.value = false
  }
}

async function confirmCheckoutSession(sessionId: string) {
  isSyncingCheckout.value = true

  try {
    for (let attempt = 1; attempt <= CHECKOUT_SYNC_ATTEMPTS; attempt++) {
      try {
        await syncCheckoutSession(sessionId)
        await authStore.fetchSession(true)
        emitToastSuccess('Subscription confirmed. The workspace is unlocked.')
        if (authStore.hasBillingAccess) {
          await router.replace(redirectPath.value)
        }
        return
      } catch (err) {
        if (
          isApiError(err) &&
          err.code === 'BILLING_CHECKOUT_PENDING' &&
          attempt < CHECKOUT_SYNC_ATTEMPTS
        ) {
          await wait(CHECKOUT_SYNC_DELAY_MS)
          continue
        }

        syncedSessionId.value = ''
        emitToastError({
          title: 'Could not confirm payment',
          message: isApiError(err)
            ? getApiErrorMessage(err)
            : 'Stripe checkout finished, but we could not confirm the subscription yet.',
        })
        return
      }
    }
  } finally {
    isSyncingCheckout.value = false
  }
}

function preferredPlanSelection(): BillingPlan {
  const requestedPlan = isBillingPlan(route.query.plan) ? route.query.plan : null
  if (requestedPlan && isPlanAvailable(requestedPlan)) {
    return requestedPlan
  }
  if (isBillingPlan(currentPlan.value) && isPlanAvailable(currentPlan.value)) {
    return currentPlan.value
  }
  const firstAvailablePlan = BILLING_PLAN_OPTIONS.find((option) => isPlanAvailable(option.id))
  if (firstAvailablePlan) {
    return firstAvailablePlan.id
  }
  return DEFAULT_BILLING_PLAN
}

function preferredIntervalSelection(plan: BillingPlan): BillingInterval {
  const requestedInterval = isBillingInterval(route.query.interval) ? route.query.interval : null
  if (requestedInterval && isBillingSelectionAvailable(plan, requestedInterval)) {
    return requestedInterval
  }
  if (
    isBillingPlan(currentPlan.value) &&
    isBillingInterval(currentInterval.value) &&
    currentPlan.value === plan &&
    isBillingSelectionAvailable(plan, currentInterval.value)
  ) {
    return currentInterval.value
  }
  if (isBillingSelectionAvailable(plan, DEFAULT_BILLING_INTERVAL)) {
    return DEFAULT_BILLING_INTERVAL
  }
  const fallbackInterval = BILLING_INTERVAL_OPTIONS.find((option) =>
    isBillingSelectionAvailable(plan, option.id),
  )
  return fallbackInterval?.id ?? DEFAULT_BILLING_INTERVAL
}

function isBillingSelectionAvailable(plan: BillingPlan, interval: BillingInterval) {
  if (!billing.value) return true

  switch (`${plan}:${interval}`) {
    case 'single:monthly':
      return billing.value.singleMonthlyAvailable
    case 'single:yearly':
      return billing.value.singleYearlyAvailable
    case 'team:monthly':
      return billing.value.teamMonthlyAvailable
    case 'team:yearly':
      return billing.value.teamYearlyAvailable
    default:
      return false
  }
}

function selectPlan(plan: BillingPlan, available: boolean) {
  if (!available) return
  selectedPlan.value = plan
  if (!isBillingSelectionAvailable(plan, selectedInterval.value)) {
    const fallbackInterval = BILLING_INTERVAL_OPTIONS.find((option) =>
      isBillingSelectionAvailable(plan, option.id),
    )
    if (fallbackInterval) {
      selectedInterval.value = fallbackInterval.id
    }
  }
}

function selectInterval(interval: BillingInterval) {
  selectedInterval.value = interval
  if (!isBillingSelectionAvailable(selectedPlan.value, interval)) {
    const firstAvailablePlan = BILLING_PLAN_OPTIONS.find((option) =>
      isBillingSelectionAvailable(option.id, interval),
    )
    if (firstAvailablePlan) {
      selectedPlan.value = firstAvailablePlan.id
    }
  }
}

function wait(ms: number) {
  return new Promise((resolve) => window.setTimeout(resolve, ms))
}

function isPlanAvailable(plan: BillingPlan) {
  if (!billing.value) return true
  return plan === 'team' ? billing.value.teamPlanAvailable : billing.value.singlePlanAvailable
}
</script>

<template>
  <main class="mx-auto flex min-h-[calc(100vh-15rem)] w-full max-w-6xl items-center py-10">
    <section
      class="grid w-full gap-6 rounded-4xl border border-zinc-200 bg-[linear-gradient(160deg,#fffdfa_0%,#f5f7f2_48%,#eef5fb_100%)] p-6 shadow-2xl sm:p-8 lg:grid-cols-[1.15fr_0.85fr] dark:border-zinc-800 dark:bg-[linear-gradient(160deg,#18181b_0%,#0c0c12_48%,#09090b_100%)]"
    >
      <div class="space-y-6">
        <div
          class="inline-flex items-center gap-2 rounded-full border border-zinc-300 bg-white/80 px-3 py-1 text-xs font-semibold tracking-[0.18em] text-zinc-700 uppercase dark:border-zinc-400 dark:bg-zinc-800 dark:text-zinc-100"
        >
          <CreditCardIcon class="size-4" />
          Billing
        </div>

        <div>
          <h1
            class="text-4xl font-semibold tracking-tight text-zinc-950 sm:text-5xl dark:text-zinc-300"
          >
            {{ statusLabel }}
          </h1>
          <p class="mt-4 max-w-2xl text-base leading-8 text-zinc-600 sm:text-lg dark:text-zinc-400">
            Choose how many people need access, then choose monthly or yearly billing. Team access
            is still capped at 5 people total in the same shared workspace.
          </p>
        </div>

        <div class="flex flex-wrap gap-3">
          <button
            v-for="interval in intervalCards"
            :key="interval.id"
            type="button"
            class="rounded-full border px-4 py-2 text-left transition"
            :class="
              interval.selected
                ? 'border-sky-400 bg-sky-50 text-sky-900 dark:border-emerald-400/40 dark:bg-emerald-950/20 dark:text-emerald-100'
                : 'border-zinc-300 bg-white/80 text-zinc-700 hover:border-zinc-400 dark:border-zinc-800 dark:bg-zinc-950/60 dark:text-zinc-200 dark:hover:border-zinc-700'
            "
            @click="selectInterval(interval.id)"
          >
            <span class="block text-sm font-semibold">{{ interval.name }}</span>
            <span class="block text-xs text-zinc-500 dark:text-zinc-400">
              {{ interval.helper }}
            </span>
          </button>
        </div>

        <div class="grid gap-3 sm:grid-cols-2">
          <button
            v-for="plan in planCards"
            :key="plan.id"
            type="button"
            class="rounded-3xl border p-5 text-left transition"
            :class="
              !plan.available
                ? 'cursor-not-allowed border-zinc-200 bg-zinc-100/70 opacity-70 dark:border-zinc-800 dark:bg-zinc-950/40'
                : selectedPlan === plan.id
                  ? 'border-sky-400 bg-sky-50/90 shadow-lg shadow-sky-100 dark:border-emerald-400/40 dark:bg-emerald-950/20 dark:shadow-none'
                  : 'border-zinc-300 bg-white/90 hover:border-zinc-400 dark:border-zinc-800 dark:bg-zinc-950/70 dark:hover:border-zinc-700'
            "
            :disabled="!plan.available"
            @click="selectPlan(plan.id, plan.available)"
          >
            <div class="flex items-start justify-between gap-3">
              <div>
                <p
                  class="text-xs font-semibold tracking-[0.18em] text-zinc-500 uppercase dark:text-zinc-400"
                >
                  {{ plan.name }}
                </p>
                <h2 class="mt-2 text-2xl font-semibold text-zinc-950 dark:text-zinc-100">
                  {{ plan.priceLabel }}
                </h2>
              </div>
              <span
                v-if="plan.isCurrent"
                class="rounded-full bg-emerald-100 px-3 py-1 text-xs font-semibold text-emerald-800 dark:bg-emerald-950/60 dark:text-emerald-200"
              >
                Current
              </span>
              <span
                v-else-if="selectedPlan === plan.id"
                class="rounded-full bg-sky-100 px-3 py-1 text-xs font-semibold text-sky-800 dark:bg-sky-950/60 dark:text-sky-200"
              >
                Selected
              </span>
            </div>

            <p class="mt-3 text-sm leading-6 text-zinc-600 dark:text-zinc-400">
              {{ plan.summary }}
            </p>

            <p
              class="mt-3 text-xs font-semibold tracking-[0.16em] text-zinc-500 uppercase dark:text-zinc-500"
            >
              {{ plan.seatLabel }}
            </p>

            <ul class="mt-4 space-y-2 text-sm text-zinc-700 dark:text-zinc-300">
              <li
                v-for="feature in plan.features"
                :key="feature"
                class="flex items-start gap-2"
              >
                <CheckCircleIcon class="mt-0.5 size-4 shrink-0 text-emerald-600" />
                <span>{{ feature }}</span>
              </li>
            </ul>

            <p
              v-if="!plan.available"
              class="mt-4 text-sm font-medium text-zinc-600 dark:text-zinc-400"
            >
              {{ intervalCards.find((item) => item.selected)?.name }} billing is not available for
              this plan yet.
            </p>
          </button>
        </div>

        <div
          v-if="checkoutState === 'canceled'"
          class="rounded-3xl border border-amber-300 bg-amber-50 px-4 py-3 text-sm leading-6 text-amber-600 dark:border-amber-400/30 dark:bg-amber-300/5 dark:text-amber-400/80"
        >
          Checkout was canceled before payment completed. Your workspace data is still safe here.
        </div>

        <div
          v-else-if="checkoutState === 'success' && isSyncingCheckout"
          class="rounded-3xl border border-sky-200 bg-sky-50 px-4 py-3 text-sm leading-6 text-sky-800"
        >
          Confirming the Stripe checkout and unlocking the workspace.
        </div>

        <div
          v-else-if="billing?.status === 'trialing' && billing.currentPeriodEnd"
          class="rounded-3xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm leading-6 text-emerald-800 dark:border-emerald-400/30 dark:bg-emerald-900/20 dark:text-emerald-400"
        >
          Your free trial is active until {{ billing.currentPeriodEnd }}. Cancel before then if you
          do not want the paid workspace to continue.
        </div>

        <div
          v-else-if="billing?.accessSource === 'promo' && hasAccess && billing.accessExpiresAt"
          class="rounded-3xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm leading-6 text-emerald-800 dark:border-emerald-400/30 dark:bg-emerald-900/20 dark:text-emerald-400"
        >
          Promo access
          <span v-if="billing.promoCode">from {{ billing.promoCode }}</span>
          is active until {{ billing.accessExpiresAt }}. Billing will be required again after that.
        </div>

        <div
          v-else-if="billing?.accessSource === 'direct' && hasAccess"
          class="rounded-3xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm leading-6 text-emerald-800 dark:border-emerald-400/30 dark:bg-emerald-900/20 dark:text-emerald-400"
        >
          Direct {{ billing.plan === 'team' ? 'team' : 'workspace' }} access is active for this
          workspace, so billing is currently bypassed.
        </div>

        <div
          v-else-if="billing?.promoExpired && billing.accessExpiresAt"
          class="rounded-3xl border border-amber-300 bg-amber-50 px-4 py-3 text-sm leading-6 text-amber-700 dark:border-amber-400/30 dark:bg-amber-300/5 dark:text-amber-300"
        >
          Your promotional access ended on {{ billing.accessExpiresAt }}. Enter another promo code
          below or start billing to continue.
        </div>

        <div
          v-else-if="hasAccess"
          class="rounded-3xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm leading-6 text-emerald-800 dark:border-emerald-400/30 dark:bg-emerald-900/20 dark:text-emerald-400"
        >
          {{ currentPlanDetails?.name || 'Current' }}
          {{ currentInterval || '' }} billing is active for this workspace.
        </div>

        <div
          v-else-if="!isOwner"
          class="rounded-3xl border border-zinc-300 bg-white/80 px-4 py-3 text-sm leading-6 text-zinc-700"
        >
          The workspace admin needs to complete billing before teammates can use the app.
        </div>

        <div
          v-else-if="!billing?.configured"
          class="rounded-3xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm leading-6 text-rose-700 dark:border-rose-500/20 dark:bg-rose-400/5 dark:text-rose-400"
        >
          Billing is temporarily unavailable right now. Please get in touch and we will help you
          finish setup.
        </div>

        <div
          class="rounded-3xl border border-zinc-300 bg-white/80 px-4 py-3 text-sm leading-6 text-zinc-700 dark:border-zinc-800 dark:bg-zinc-950/60 dark:text-zinc-300"
        >
          Need setup help? Email
          <a
            href="mailto:invoiceandgo@gmail.com"
            class="font-semibold text-zinc-900 underline decoration-zinc-400 underline-offset-2 dark:text-zinc-100 dark:decoration-zinc-600"
          >
            invoiceandgo@gmail.com
          </a>
          and we will help you finish billing.
        </div>

        <article
          v-if="showPromoForm"
          class="rounded-3xl border border-zinc-300 bg-white/85 p-5 dark:border-zinc-800 dark:bg-zinc-950"
        >
          <div class="flex items-start gap-3">
            <div
              class="rounded-2xl bg-zinc-100 p-3 text-zinc-700 dark:bg-zinc-900 dark:text-zinc-300"
            >
              <SparklesIcon class="size-6" />
            </div>
            <div class="min-w-0 flex-1">
              <h2 class="text-lg font-semibold text-zinc-950 dark:text-zinc-300">Promo code</h2>
              <p class="mt-2 text-sm leading-7 text-zinc-600 dark:text-zinc-500">
                Enter a promo code to unlock the workspace.
              </p>

              <form
                class="mt-4 flex flex-col gap-2 sm:flex-row"
                @submit.prevent="submitPromoCode"
              >
                <TheInput
                  v-model="promoCode"
                  class-names="flex-1"
                  placeholder="EARLYBIRD14"
                  :error="promoFieldErrors.code ?? null"
                />

                <div class="">
                  <TheButton
                    type="submit"
                    :disabled="isRedeemingPromo"
                    class="cursor-pointer py-2.5"
                  >
                    {{ isRedeemingPromo ? 'Applying...' : 'Apply code' }}
                  </TheButton>
                </div>
              </form>
            </div>
          </div>
        </article>

        <div class="flex flex-wrap gap-3">
          <TheButton
            v-if="showPlanAction"
            :disabled="isSubmittingPlan || !billing?.configured || !selectedPlanAvailable"
            @click="submitPlanAction"
          >
            {{ isSubmittingPlan ? 'Saving...' : selectionActionLabel }}
          </TheButton>

          <TheButton
            v-if="isOwner && billing?.portalAvailable"
            variant="secondary"
            :disabled="isOpeningPortal || !billing?.configured"
            @click="openPortal"
          >
            {{ isOpeningPortal ? 'Opening portal...' : 'Manage billing' }}
          </TheButton>

          <TheButton
            v-if="hasAccess"
            variant="secondary"
            @click="router.push(redirectPath)"
          >
            Continue to workspace
          </TheButton>
        </div>
      </div>

      <aside class="grid gap-4">
        <article
          class="rounded-3xl border border-zinc-300 bg-white/85 p-5 dark:border-zinc-800 dark:bg-zinc-950"
        >
          <div class="flex items-start gap-3">
            <div
              class="rounded-2xl bg-zinc-100 p-3 text-zinc-700 dark:bg-zinc-900 dark:text-zinc-300"
            >
              <UsersIcon class="size-6" />
            </div>
            <div>
              <h2 class="text-lg font-semibold text-zinc-950 dark:text-zinc-300">
                Selected billing
              </h2>
              <p class="mt-2 text-sm leading-7 text-zinc-600 dark:text-zinc-500">
                {{ selectedSelectionLabel }}: {{ selectedOfferLabel }}.
                {{ selectedPlanDetails.description }}
              </p>
              <p class="mt-3 text-xs font-medium tracking-[0.16em] text-zinc-500 uppercase">
                {{ selectedPlanDetails.seatLabel }}
              </p>
            </div>
          </div>
        </article>

        <article
          class="rounded-3xl border border-zinc-300 bg-white/85 p-5 dark:border-zinc-800 dark:bg-zinc-950"
        >
          <div class="flex items-start gap-3">
            <div
              class="rounded-2xl bg-zinc-100 p-3 text-zinc-700 dark:bg-zinc-900 dark:text-zinc-300"
            >
              <CheckCircleIcon class="size-6" />
            </div>
            <div>
              <h2 class="text-lg font-semibold text-zinc-950 dark:text-zinc-300">
                Current subscription
              </h2>
              <p class="mt-2 text-sm leading-7 text-zinc-600 dark:text-zinc-500">
                <span v-if="currentPlanDetails && currentPriceLabel">
                  {{ currentPlanDetails.name }} {{ currentInterval }} at {{ currentPriceLabel }}.
                </span>
                <span v-else>No active billing selection yet.</span>
                Upgrades and cadence changes are handled inside the app and still keep the Stripe
                portal for payment method changes and invoices.
              </p>
            </div>
          </div>
        </article>

        <article
          class="rounded-3xl border border-zinc-300 bg-white/85 p-5 dark:border-zinc-800 dark:bg-zinc-950"
        >
          <div class="flex items-start gap-3">
            <div
              class="rounded-2xl bg-zinc-100 p-3 text-zinc-700 dark:bg-zinc-900 dark:text-zinc-300"
            >
              <ExclamationTriangleIcon class="size-6" />
            </div>
            <div>
              <h2 class="text-lg font-semibold text-zinc-950 dark:text-zinc-300">
                Trial and renewal
              </h2>
              <p class="mt-2 text-sm leading-7 text-zinc-600 dark:text-zinc-500">
                Checkout collects a payment method up front. If the trial finishes and no valid
                payment method exists, the subscription is canceled and workspace access locks until
                billing restarts.
              </p>
              <p
                v-if="billing?.currentPeriodEnd"
                class="mt-3 text-xs font-medium tracking-[0.16em] text-zinc-500 uppercase"
              >
                {{ periodEndLabel }}: {{ billing.currentPeriodEnd }}
              </p>
            </div>
          </div>
        </article>
      </aside>
    </section>
  </main>
</template>

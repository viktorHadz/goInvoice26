<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import TheButton from '@/components/UI/TheButton.vue'
import TheInput from '@/components/UI/TheInput.vue'
import UserAvatar from '@/components/UI/UserAvatar.vue'
import TeamCard from '@/components/team/TeamCard.vue'
import { useTeamStore, type TeamInvite, type TeamMember } from '@/stores/team'
import { useAuthStore } from '@/stores/auth'
import { useBillingCatalogStore } from '@/stores/billingCatalog'
import {
  type BillingInterval,
  DEFAULT_BILLING_INTERVAL,
  isBillingInterval,
  TEAM_PLAN_SEAT_LIMIT,
} from '@/constants/billing'
import {
  cancelSubscription,
  changeSubscriptionPlan,
  createCheckoutSession,
  createPortalSession,
} from '@/utils/billingHttpHandler'
import { deleteWorkspace as deleteWorkspaceRequest } from '@/utils/workspaceHttpHandler'
import { handleActionError } from '@/utils/errors/handleActionError'
import { requestConfirmation } from '@/utils/confirm'
import { emitToastError, emitToastSuccess } from '@/utils/toast'
import { getApiErrorMessage, isApiError } from '@/utils/apiErrors'
import {
  ArrowTopRightOnSquareIcon,
  BanknotesIcon,
  BuildingOffice2Icon,
  EnvelopeIcon,
  ExclamationTriangleIcon,
  UserCircleIcon,
  UserGroupIcon,
} from '@heroicons/vue/24/outline'

const teamStore = useTeamStore()
const authStore = useAuthStore()
const billingCatalogStore = useBillingCatalogStore()
const router = useRouter()

void billingCatalogStore.fetchCatalog().catch(() => undefined)

const inviteEmail = ref('')
const fieldErrors = ref<Record<string, string>>({})
const isSubmittingInvite = ref(false)
const revokingInviteId = ref<number | null>(null)
const removingMemberId = ref<number | null>(null)
const isOpeningBillingPortal = ref(false)
const isCancellingSubscription = ref(false)
const isDeletingWorkspace = ref(false)
const isActivatingTeamPlan = ref(false)

const members = computed(() => teamStore.members)
const invites = computed(() => teamStore.invites)
const currentUserId = computed(() => authStore.user?.id ?? 0)
const billing = computed(() => authStore.billing)
const hasBillingAccess = computed(() => authStore.hasBillingAccess)
const isOwner = computed(() => authStore.isOwner)
const isTrialing = computed(() => billing.value?.status === 'trialing')
const currentPlan = computed(() => billing.value?.plan ?? '')
const currentInterval = computed(() => billing.value?.interval ?? '')
const isTeamPlan = computed(() => currentPlan.value === 'team')
const teamPlanAvailable = computed(() => billing.value?.teamPlanAvailable === true)
const targetTeamInterval = computed<BillingInterval>(() => {
  if (isBillingInterval(currentInterval.value)) {
    if (
      (currentInterval.value === 'monthly' && billing.value?.teamMonthlyAvailable) ||
      (currentInterval.value === 'yearly' && billing.value?.teamYearlyAvailable)
    ) {
      return currentInterval.value
    }
  }
  if (billing.value?.teamMonthlyAvailable) return 'monthly'
  if (billing.value?.teamYearlyAvailable) return 'yearly'
  return DEFAULT_BILLING_INTERVAL
})
const targetTeamPriceLabel = computed(() =>
  billingCatalogStore.getPriceLabel('team', targetTeamInterval.value),
)
const canLoadTeam = computed(() => isOwner.value && hasBillingAccess.value)
const canInviteTeammates = computed(() => canLoadTeam.value && isTeamPlan.value)
const seatLimit = computed(() => {
  if (isTeamPlan.value) {
    return Math.max(billing.value?.seatLimit ?? TEAM_PLAN_SEAT_LIMIT, TEAM_PLAN_SEAT_LIMIT)
  }
  return 1
})
const reservedSeats = computed(() => teamStore.memberCount + teamStore.inviteCount)
const remainingSeats = computed(() => Math.max(seatLimit.value - reservedSeats.value, 0))
const billingPlanLabel = computed(() => (isTeamPlan.value ? 'Team plan' : 'Single plan'))
const canOpenBillingPortal = computed(
  () => authStore.billing?.configured === true && authStore.billing?.portalAvailable === true,
)
const canCancelSubscription = computed(() => {
  if (!authStore.canManageBilling || !authStore.billingConfigured) return false
  if (authStore.billing?.cancelAtPeriodEnd) return false

  const status = authStore.billing?.status ?? ''
  return ['active', 'trialing', 'past_due', 'unpaid', 'incomplete'].includes(status)
})
const billingHelpText = computed(() =>
  canOpenBillingPortal.value
    ? 'Open Stripe to update payment methods, invoices, and cancellation details.'
    : 'The Stripe billing portal is unavailable right now. Please try again in a moment.',
)
const workspaceActionHelpText = computed(() => {
  if (authStore.billing?.cancelAtPeriodEnd) {
    if (isTrialing.value) {
      return authStore.billing?.currentPeriodEnd
        ? `The free trial is already scheduled to end on ${authStore.billing.currentPeriodEnd}.`
        : 'The free trial is already scheduled to end without renewal.'
    }
    return authStore.billing?.currentPeriodEnd
      ? `Subscription cancellation is already scheduled for ${authStore.billing.currentPeriodEnd}.`
      : 'Subscription cancellation is already scheduled at the end of the current billing period.'
  }
  if (isTrialing.value && authStore.billing?.currentPeriodEnd) {
    return `The free trial ends on ${authStore.billing.currentPeriodEnd}.`
  }
  if (canCancelSubscription.value && authStore.billing?.currentPeriodEnd) {
    return `Subscription access stays active until ${authStore.billing.currentPeriodEnd}.`
  }
  if (!authStore.billingConfigured) {
    return 'Billing controls are unavailable right now, but you can still remove the workspace if no live subscription needs cancellation.'
  }
  return 'Workspace deletion permanently removes teammates, invoices, settings, and uploaded files.'
})
const teamPlanHelpText = computed(() => {
  if (!teamPlanAvailable.value) {
    return 'The team plan is not available yet.'
  }
  if (hasBillingAccess.value) {
    return `Upgrade to the ${targetTeamPriceLabel.value} team plan to invite up to ${TEAM_PLAN_SEAT_LIMIT} people into this shared workspace.`
  }
  return `Start billing on the ${targetTeamPriceLabel.value} team plan to invite up to ${TEAM_PLAN_SEAT_LIMIT} people.`
})
const activateTeamPlanLabel = computed(() => {
  if (hasBillingAccess.value) {
    return `Switch to ${targetTeamPriceLabel.value} team plan`
  }
  return `Start ${targetTeamPriceLabel.value} team plan`
})

async function loadTeam() {
  if (!canLoadTeam.value) return
  try {
    await teamStore.load()
  } catch (err) {
    handleActionError(err, {
      toastTitle: 'Could not load team',
      mapFields: false,
    })
  }
}

onMounted(async () => {
  await loadTeam()
})

watch(canLoadTeam, async (canLoad, hadAccess) => {
  if (!canLoad) {
    if (hadAccess) teamStore.reset()
    return
  }
  if (!teamStore.hasLoaded) {
    await loadTeam()
  }
})

async function inviteTeammate() {
  fieldErrors.value = {}
  isSubmittingInvite.value = true
  try {
    await teamStore.invite(inviteEmail.value)
    inviteEmail.value = ''
    emitToastSuccess('Invite added. Ask your teammate to log in with Google from the login page.')
  } catch (err) {
    handleActionError(err, {
      fieldErrors,
      toastTitle: 'Could not add invite',
    })
  } finally {
    isSubmittingInvite.value = false
  }
}

async function activateTeamPlan() {
  if (isActivatingTeamPlan.value || !isOwner.value || !teamPlanAvailable.value) return

  isActivatingTeamPlan.value = true
  try {
    if (hasBillingAccess.value) {
      await changeSubscriptionPlan('team', targetTeamInterval.value)
      await authStore.fetchSession(true)
      emitToastSuccess('Team plan activated. You can now invite teammates.')
      await loadTeam()
      return
    }

    const session = await createCheckoutSession(
      'team',
      targetTeamInterval.value,
      router.currentRoute.value.fullPath,
    )
    window.location.assign(session.url)
  } catch (err) {
    emitToastError({
      title: 'Could not activate team plan',
      message: isApiError(err)
        ? getApiErrorMessage(err)
        : 'The team plan could not be activated right now.',
    })
  } finally {
    isActivatingTeamPlan.value = false
  }
}

async function revokeInvite(invite: TeamInvite) {
  const confirmed = await requestConfirmation({
    title: 'Revoke invite',
    message: `Remove the pending invite for ${invite.email}?`,
    details: 'They will no longer be able to join this workspace until you invite them again.',
    confirmLabel: 'Revoke invite',
    confirmVariant: 'danger',
  })
  if (!confirmed) return

  revokingInviteId.value = invite.id
  try {
    await teamStore.revokeInvite(invite.id)
    emitToastSuccess('Invite revoked.')
  } catch (err) {
    handleActionError(err, {
      toastTitle: 'Could not revoke invite',
      mapFields: false,
    })
  } finally {
    revokingInviteId.value = null
  }
}

async function removeMember(member: TeamMember) {
  const confirmed = await requestConfirmation({
    title: 'Remove teammate',
    message: `Remove ${member.email} from the workspace?`,
    details:
      'Their active sessions will be signed out and they will lose access until invited again.',
    confirmLabel: 'Remove teammate',
    confirmVariant: 'danger',
  })
  if (!confirmed) return

  removingMemberId.value = member.id
  try {
    await teamStore.removeMember(member.id)
    emitToastSuccess('Teammate removed.')
  } catch (err) {
    handleActionError(err, {
      toastTitle: 'Could not remove teammate',
      mapFields: false,
    })
  } finally {
    removingMemberId.value = null
  }
}

async function openBillingPortal() {
  if (isOpeningBillingPortal.value || !canOpenBillingPortal.value) return

  isOpeningBillingPortal.value = true
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
    isOpeningBillingPortal.value = false
  }
}

async function scheduleSubscriptionCancellation() {
  if (isCancellingSubscription.value || !canCancelSubscription.value) return

  const confirmed = await requestConfirmation({
    title: isTrialing.value ? 'End free trial' : 'Cancel subscription',
    message: isTrialing.value
      ? 'Let the free trial end without starting the paid workspace subscription?'
      : 'Cancel the workspace subscription at the end of the current billing period?',
    details: authStore.billing?.currentPeriodEnd
      ? `Workspace access stays active until ${authStore.billing.currentPeriodEnd}. After that, the workspace locks until billing is restarted.`
      : isTrialing.value
        ? 'Workspace access stays active until the free trial ends. After that, the workspace locks unless billing is restarted.'
        : 'Workspace access stays active until the current billing period ends. After that, the workspace locks until billing is restarted.',
    confirmLabel: isTrialing.value ? 'End free trial' : 'Cancel subscription',
    confirmVariant: 'danger',
  })
  if (!confirmed) return

  isCancellingSubscription.value = true
  try {
    await cancelSubscription()
    await authStore.fetchSession(true)
    emitToastSuccess(
      isTrialing.value
        ? 'Free trial will end on schedule.'
        : 'Subscription cancellation scheduled.',
    )
  } catch (err) {
    handleActionError(err, {
      toastTitle: 'Could not cancel subscription',
      mapFields: false,
    })
  } finally {
    isCancellingSubscription.value = false
  }
}

async function deleteWorkspace() {
  if (isDeletingWorkspace.value || !isOwner.value) return

  const confirmed = await requestConfirmation({
    title: 'Delete workspace',
    message: `Delete ${authStore.account?.name || 'this workspace'} and all team data?`,
    details:
      'This permanently deletes the owner and teammate accounts, invites, clients, saved items, invoices, settings, and uploaded files. Any active Stripe subscription is canceled immediately before deletion.',
    confirmLabel: 'Delete workspace',
    confirmVariant: 'danger',
  })
  if (!confirmed) return

  isDeletingWorkspace.value = true
  try {
    await deleteWorkspaceRequest()
    authStore.clearWorkspaceState()
    await authStore.fetchSession(true)
    emitToastSuccess('Workspace deleted.')
    await router.replace({ name: 'landing' })
  } catch (err) {
    handleActionError(err, {
      toastTitle: 'Could not delete workspace',
      mapFields: false,
    })
  } finally {
    isDeletingWorkspace.value = false
  }
}

function canRemoveMember(member: TeamMember) {
  return member.role !== 'owner' && member.id !== currentUserId.value
}

function formatDate(value: string) {
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return 'Recently'
  return parsed.toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}
</script>

<template>
  <div class="h-full overflow-y-auto p-4 sm:p-5">
    <section class="grid grid-cols-1 gap-4 sm:grid-cols-2">
      <TeamCard
        v-if="isOwner && !hasBillingAccess"
        title="Workspace locked"
        :icon="ExclamationTriangleIcon"
      >
        <template #text>
          Team member tools are hidden until billing is active again, but owner-only subscription
          and workspace controls remain available below.
        </template>
      </TeamCard>

      <TeamCard
        v-if="hasBillingAccess"
        title="Team access"
        :icon="BuildingOffice2Icon"
      >
        <template #text>
          <span v-if="isTeamPlan">
            Invite teammates into the same shared workspace and remove access when needed.
          </span>
          <span v-else>
            The single plan covers the owner only. Upgrade to the {{ targetTeamPriceLabel }} team
            plan to invite teammates.
          </span>
        </template>

        <div class="grid gap-2 sm:mt-8 sm:grid-cols-3">
          <div
            class="rounded-2xl border border-sky-200 bg-sky-100/50 px-4 py-3 text-sm text-sky-800 dark:border-emerald-400/20 dark:bg-emerald-950/50 dark:text-emerald-100"
          >
            <div class="text-mini font-semibold tracking-[0.18em] uppercase">Members</div>
            <div class="mt-1 text-xl font-semibold">{{ teamStore.memberCount }}</div>
          </div>

          <div
            class="rounded-2xl border border-amber-200 bg-amber-100/40 px-4 py-3 text-sm text-amber-500 dark:border-amber-400/20 dark:bg-amber-900/20 dark:text-amber-100"
          >
            <div class="text-mini font-semibold tracking-[0.18em] uppercase">Pending</div>
            <div class="mt-1 text-xl font-semibold">{{ teamStore.inviteCount }}</div>
          </div>

          <div
            class="rounded-2xl border border-zinc-300 bg-white/80 px-4 py-3 text-sm text-zinc-700 dark:border-zinc-800 dark:bg-zinc-950/40 dark:text-zinc-100"
          >
            <div class="text-mini font-semibold tracking-[0.18em] uppercase">Seats</div>
            <div class="mt-1 text-xl font-semibold">{{ reservedSeats }} / {{ seatLimit }}</div>
          </div>
        </div>

        <p class="mt-4 text-xs leading-6 text-zinc-600 dark:text-zinc-400">
          {{ billingPlanLabel }}. {{ remainingSeats }} seat{{ remainingSeats === 1 ? '' : 's' }}
          remaining.
        </p>
      </TeamCard>

      <TeamCard
        v-if="hasBillingAccess && isOwner && !isTeamPlan"
        title="Team plan required"
        :icon="EnvelopeIcon"
      >
        <template #text>
          {{ teamPlanHelpText }}
        </template>

        <div class="flex flex-wrap items-center gap-3">
          <TheButton
            :disabled="isActivatingTeamPlan || !teamPlanAvailable"
            class="cursor-pointer"
            @click="activateTeamPlan"
          >
            {{ isActivatingTeamPlan ? 'Updating plan...' : activateTeamPlanLabel }}
          </TheButton>

          <span class="text-xs text-zinc-600 dark:text-zinc-400">
            Up to {{ TEAM_PLAN_SEAT_LIMIT }} people total in one workspace.
          </span>
        </div>
      </TeamCard>

      <TeamCard
        v-if="canInviteTeammates"
        title="Invite teammate"
        :icon="EnvelopeIcon"
      >
        <template #text>
          Add the teammate&apos;s email and send them to the
          <RouterLink
            to="/login"
            class="font-medium text-sky-700 hover:text-sky-800 dark:text-emerald-300 dark:hover:text-emerald-200"
          >
            login page
          </RouterLink>
          to sign in.
        </template>

        <form
          class="flex flex-col gap-2"
          @submit.prevent="inviteTeammate"
        >
          <TheInput
            v-model="inviteEmail"
            label="Teammate email"
            type="email"
            placeholder="teammate@company.com"
            :error="fieldErrors.email ?? null"
          />

          <div class="flex flex-wrap items-center gap-3">
            <TheButton
              type="submit"
              :disabled="isSubmittingInvite || remainingSeats <= 0"
              class="cursor-pointer"
            >
              {{ isSubmittingInvite ? 'Adding invite...' : 'Add invite' }}
            </TheButton>

            <span class="text-xs text-zinc-600 dark:text-zinc-400">
              Pending until they sign in with Google.
            </span>
          </div>
        </form>
      </TeamCard>

      <TeamCard
        v-if="hasBillingAccess"
        title="Current members"
        :icon="UserGroupIcon"
      >
        <template #text>
          Everyone here shares the same clients, invoice settings, and logo.
        </template>

        <div
          v-if="teamStore.isLoading && !teamStore.hasLoaded"
          class="text-sm text-zinc-600 dark:text-zinc-400"
        >
          Loading team...
        </div>

        <div
          v-else
          class="space-y-2.5"
        >
          <article
            v-for="member in members"
            :key="member.id"
            class="flex items-center justify-between rounded-2xl border border-zinc-300 bg-zinc-50/70 px-4 py-3 dark:border-zinc-800 dark:bg-zinc-950/40"
          >
            <div class="flex min-w-0 items-center gap-3">
              <UserAvatar
                :name="member.name"
                :email="member.email"
                :avatar-url="member.avatarUrl"
                class="size-10 rounded-2xl"
              />

              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <p class="truncate text-sm font-semibold text-zinc-900 dark:text-zinc-100">
                    {{ member.name || member.email }}
                  </p>

                  <span
                    class="text-mini rounded-lg px-2 py-0.5 font-semibold tracking-[0.16em] uppercase"
                    :class="
                      member.role === 'owner'
                        ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950/50 dark:text-emerald-200'
                        : 'bg-zinc-200 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-200'
                    "
                  >
                    {{ member.role }}
                  </span>
                </div>

                <p class="truncate text-sm text-zinc-600 dark:text-zinc-400">
                  {{ member.email }}
                </p>

                <p class="mt-0.5 text-xs text-zinc-600 dark:text-zinc-500">
                  Joined {{ formatDate(member.createdAt) }}
                </p>
              </div>
            </div>

            <TheButton
              v-if="canRemoveMember(member)"
              variant="danger"
              :disabled="removingMemberId === member.id"
              class="cursor-pointer"
              @click="removeMember(member)"
            >
              {{ removingMemberId === member.id ? 'Removing...' : 'Remove' }}
            </TheButton>
          </article>
        </div>
      </TeamCard>

      <TeamCard
        v-if="hasBillingAccess && (isTeamPlan || invites.length > 0)"
        title="Pending invites"
        :icon="UserCircleIcon"
      >
        <template #text>
          These teammates can join the workspace once they log in with Google.
        </template>

        <div
          v-if="!invites.length"
          class="rounded-xl border border-dashed border-zinc-300 px-4 py-7 text-center text-sm text-zinc-600 dark:border-zinc-700 dark:text-zinc-400"
        >
          No pending invites right now.
        </div>

        <div
          v-else
          class="space-y-2.5"
        >
          <article
            v-for="invite in invites"
            :key="invite.id"
            class="flex items-center justify-between gap-3 rounded-2xl border border-zinc-300 bg-zinc-50/70 px-4 py-3 dark:border-zinc-800 dark:bg-zinc-950/60"
          >
            <div class="min-w-0">
              <p class="truncate text-sm font-semibold text-zinc-900 dark:text-zinc-100">
                {{ invite.email }}
              </p>
              <p class="mt-0.5 text-xs text-zinc-600 dark:text-zinc-500">
                Invited {{ formatDate(invite.createdAt) }}
              </p>
            </div>

            <TheButton
              variant="danger"
              :disabled="revokingInviteId === invite.id"
              class="cursor-pointer"
              @click="revokeInvite(invite)"
            >
              {{ revokingInviteId === invite.id ? 'Revoking...' : 'Revoke' }}
            </TheButton>
          </article>
        </div>
      </TeamCard>

      <TeamCard
        v-if="authStore.canManageBilling"
        title="Workspace billing"
        :icon="BanknotesIcon"
      >
        <template #text>Manage the shared subscription for everyone in this workspace.</template>

        <button
          type="button"
          class="group flex w-full cursor-pointer items-start gap-3 rounded-xl border border-zinc-300 bg-zinc-50/70 p-3 text-left transition hover:border-sky-600/50 hover:bg-sky-50 disabled:cursor-not-allowed disabled:opacity-80 dark:border-zinc-800 dark:bg-zinc-950/40 dark:hover:border-emerald-400/25 dark:hover:bg-emerald-950/20"
          :disabled="isOpeningBillingPortal || !canOpenBillingPortal"
          @click="openBillingPortal"
        >
          <div class="min-w-0 flex-1">
            <div class="flex items-center justify-between gap-2">
              <span
                class="truncate text-sm font-medium text-zinc-900 group-hover:text-sky-700 dark:text-zinc-100 dark:group-hover:text-emerald-300"
              >
                Billing
              </span>

              <div
                class="text-mini flex shrink-0 items-center gap-1.5 font-medium text-zinc-500 transition group-hover:text-sky-700 dark:text-zinc-400 dark:group-hover:text-emerald-300"
              >
                <span>{{ isOpeningBillingPortal ? 'Opening Stripe...' : 'Open Stripe' }}</span>
                <ArrowTopRightOnSquareIcon class="size-4" />
              </div>
            </div>

            <p class="mt-2 text-xs text-zinc-600 dark:text-zinc-400">
              {{ billingHelpText }}
            </p>
          </div>
        </button>
      </TeamCard>

      <TeamCard
        v-if="isOwner"
        title="Workspace management"
        :icon="ExclamationTriangleIcon"
      >
        <template #text>
          Owner-only controls for ending billing and permanently removing the workspace.
        </template>

        <div
          class="rounded-2xl border border-rose-200 bg-rose-50/80 p-4 dark:border-rose-500/20 dark:bg-rose-950/30"
        >
          <p class="text-sm leading-6 text-rose-800 dark:text-rose-200">
            {{ workspaceActionHelpText }}
          </p>

          <p class="mt-3 text-xs leading-5 text-rose-700 dark:text-rose-300/90">
            Deleting the workspace signs out everyone and removes all account data from the database
            along with uploaded assets.
          </p>

          <div class="mt-4 flex flex-wrap gap-3">
            <TheButton
              variant="danger"
              :disabled="isCancellingSubscription || !canCancelSubscription"
              class="cursor-pointer"
              @click="scheduleSubscriptionCancellation"
            >
              {{
                authStore.billing?.cancelAtPeriodEnd
                  ? 'Cancellation scheduled'
                  : isCancellingSubscription
                    ? 'Cancelling subscription...'
                    : 'Cancel subscription'
              }}
            </TheButton>

            <TheButton
              variant="danger"
              :disabled="isDeletingWorkspace"
              class="cursor-pointer"
              @click="deleteWorkspace"
            >
              {{ isDeletingWorkspace ? 'Deleting workspace...' : 'Delete workspace' }}
            </TheButton>
          </div>
        </div>
      </TeamCard>
    </section>
  </div>
</template>

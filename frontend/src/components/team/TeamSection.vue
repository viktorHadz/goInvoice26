<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import TheButton from '@/components/UI/TheButton.vue'
import TheInput from '@/components/UI/TheInput.vue'
import UserAvatar from '@/components/UI/UserAvatar.vue'
import TeamCard from '@/components/team/TeamCard.vue'
import { useTeamStore, type TeamInvite, type TeamMember } from '@/stores/team'
import { useAuthStore } from '@/stores/auth'
import { createPortalSession } from '@/utils/billingHttpHandler'
import { handleActionError } from '@/utils/errors/handleActionError'
import { requestConfirmation } from '@/utils/confirm'
import { emitToastError, emitToastSuccess } from '@/utils/toast'
import { getApiErrorMessage, isApiError } from '@/utils/apiErrors'
import {
  ArrowTopRightOnSquareIcon,
  BanknotesIcon,
  BuildingOffice2Icon,
  EnvelopeIcon,
  UserCircleIcon,
  UserGroupIcon,
} from '@heroicons/vue/24/outline'

const teamStore = useTeamStore()
const authStore = useAuthStore()

const inviteEmail = ref('')
const fieldErrors = ref<Record<string, string>>({})
const isSubmittingInvite = ref(false)
const revokingInviteId = ref<number | null>(null)
const removingMemberId = ref<number | null>(null)
const isOpeningBillingPortal = ref(false)

const members = computed(() => teamStore.members)
const invites = computed(() => teamStore.invites)
const currentUserId = computed(() => authStore.user?.id ?? 0)

const canOpenBillingPortal = computed(
  () => authStore.billing?.configured === true && authStore.billing?.portalAvailable === true,
)

const billingHelpText = computed(() =>
  canOpenBillingPortal.value
    ? 'Open Stripe to update the workspace subscription, payment method, and invoices.'
    : 'The Stripe billing portal is unavailable right now. Please try again in a moment.',
)

onMounted(async () => {
  try {
    await teamStore.load()
  } catch (err) {
    handleActionError(err, {
      toastTitle: 'Could not load team',
      mapFields: false,
    })
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
        title="Team access"
        :icon="BuildingOffice2Icon"
      >
        <template #text>
          Invite teammates into the same workspace and remove access when needed.
        </template>

        <div class="grid gap-2 sm:mt-8 sm:grid-cols-2">
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
        </div>
      </TeamCard>

      <TeamCard
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
              :disabled="isSubmittingInvite"
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
          class="group flex w-full items-start gap-3 rounded-xl border border-zinc-300 bg-zinc-50/70 p-3 text-left transition hover:border-sky-600/50 hover:bg-sky-50 disabled:cursor-not-allowed disabled:opacity-80 dark:border-zinc-800 dark:bg-zinc-950/40 dark:hover:border-emerald-400/25 dark:hover:bg-emerald-950/20"
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
    </section>
  </div>
</template>

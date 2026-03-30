<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import TheButton from '@/components/UI/TheButton.vue'
import TheInput from '@/components/UI/TheInput.vue'
import UserAvatar from '@/components/UI/UserAvatar.vue'
import { useTeamStore, type TeamInvite, type TeamMember } from '@/stores/team'
import { useAuthStore } from '@/stores/auth'
import { handleActionError } from '@/utils/errors/handleActionError'
import { requestConfirmation } from '@/utils/confirm'
import { emitToastSuccess } from '@/utils/toast'
import { EnvelopeIcon, UserCircleIcon, UserGroupIcon } from '@heroicons/vue/24/outline'

const teamStore = useTeamStore()
const authStore = useAuthStore()

const inviteEmail = ref('')
const fieldErrors = ref<Record<string, string>>({})
const isSubmittingInvite = ref(false)
const revokingInviteId = ref<number | null>(null)
const removingMemberId = ref<number | null>(null)

const members = computed(() => teamStore.members)
const invites = computed(() => teamStore.invites)
const currentUserId = computed(() => authStore.user?.id ?? 0)

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
    <section class="grid gap-4">
      <article
        class="rounded-2xl border border-zinc-300 bg-white p-5 shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
      >
        <h3 class="text-xl font-semibold tracking-tight text-zinc-950 dark:text-zinc-50">
          Team access
        </h3>
        <p class="mt-2 text-sm leading-6 text-zinc-600 dark:text-zinc-400">
          Invite teammates into the same workspace and remove access when needed.
        </p>

        <div class="mt-4 grid gap-2 sm:grid-cols-2">
          <div
            class="rounded-2xl border border-sky-200 bg-sky-50 px-4 py-3 text-sm text-sky-800 dark:border-emerald-400/20 dark:bg-emerald-950/30 dark:text-emerald-100"
          >
            <div class="text-xs font-semibold tracking-[0.18em] uppercase opacity-70">Members</div>
            <div class="mt-1.5 text-xl font-semibold">{{ teamStore.memberCount }}</div>
          </div>
          <div
            class="rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800 dark:border-amber-400/20 dark:bg-amber-950/30 dark:text-amber-100"
          >
            <div class="text-xs font-semibold tracking-[0.18em] uppercase opacity-70">Pending</div>
            <div class="mt-1.5 text-xl font-semibold">{{ teamStore.inviteCount }}</div>
          </div>
        </div>
      </article>

      <article
        class="rounded-2xl border border-zinc-300 bg-white p-5 shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
      >
        <div class="flex items-start justify-between gap-3">
          <div>
            <h4 class="text-lg font-semibold text-zinc-950 dark:text-zinc-50">Invite teammate</h4>
            <p class="mt-1.5 text-sm leading-6 text-zinc-600 dark:text-zinc-400">
              Add the teammate&apos;s email and send them to the
              <RouterLink
                to="/login"
                class="font-medium text-sky-700 hover:text-sky-800 dark:text-emerald-300 dark:hover:text-emerald-200"
              >
                login page
              </RouterLink>
              to sign in.
            </p>
          </div>
          <div
            class="rounded-2xl border border-zinc-300 bg-zinc-50 p-2.5 text-zinc-700 dark:border-zinc-700 dark:bg-zinc-950/50 dark:text-zinc-200"
          >
            <EnvelopeIcon class="size-5" />
          </div>
        </div>

        <form
          class="mt-4 flex flex-col gap-2"
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
      </article>

      <article
        class="rounded-2xl border border-zinc-300 bg-white p-5 shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
      >
        <div class="flex items-start justify-between gap-3">
          <div>
            <h4 class="text-lg font-semibold text-zinc-950 dark:text-zinc-50">Current members</h4>
            <p class="mt-1.5 text-sm leading-6 text-zinc-600 dark:text-zinc-400">
              Everyone here shares the same clients, invoice settings, and logo.
            </p>
          </div>
          <div
            class="rounded-2xl border border-zinc-300 bg-zinc-50 p-2.5 text-zinc-700 dark:border-zinc-700 dark:bg-zinc-950/60 dark:text-zinc-200"
          >
            <UserGroupIcon class="size-5" />
          </div>
        </div>

        <div
          v-if="teamStore.isLoading && !teamStore.hasLoaded"
          class="mt-4 text-sm text-zinc-600 dark:text-zinc-400"
        >
          Loading team...
        </div>

        <div
          v-else
          class="mt-4 space-y-2.5"
        >
          <article
            v-for="member in members"
            :key="member.id"
            class="flex items-center justify-between gap-3 rounded-2xl border border-zinc-300 bg-zinc-50/70 px-4 py-3 dark:border-zinc-800 dark:bg-zinc-950/60"
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
                    class="rounded-lg px-2 py-0.5 text-[11px] font-semibold tracking-[0.16em] uppercase"
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
      </article>

      <article
        class="rounded-2xl border border-zinc-300 bg-white p-5 shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
      >
        <div class="flex items-start justify-between gap-3">
          <div>
            <h4 class="text-lg font-semibold text-zinc-950 dark:text-zinc-50">Pending invites</h4>
            <p class="mt-1.5 text-sm leading-6 text-zinc-600 dark:text-zinc-400">
              These teammates can join the workspace once they log in with Google.
            </p>
          </div>
          <div
            class="rounded-xl border border-zinc-300 bg-zinc-50 p-2.5 text-zinc-700 dark:border-zinc-700 dark:bg-zinc-950/60 dark:text-zinc-200"
          >
            <UserCircleIcon class="size-5" />
          </div>
        </div>

        <div
          v-if="!invites.length"
          class="mt-4 rounded-2xl border border-dashed border-zinc-300 px-4 py-7 text-center text-sm text-zinc-600 dark:border-zinc-700 dark:text-zinc-400"
        >
          No pending invites right now.
        </div>

        <div
          v-else
          class="mt-4 space-y-2.5"
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
      </article>
    </section>
  </div>
</template>

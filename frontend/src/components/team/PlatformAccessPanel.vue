<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { KeyIcon, SparklesIcon, TrashIcon } from '@heroicons/vue/24/outline'
import TeamCard from '@/components/team/TeamCard.vue'
import TheButton from '@/components/UI/TheButton.vue'
import TheDropdown from '@/components/UI/TheDropdown.vue'
import TheInput from '@/components/UI/TheInput.vue'
import {
  createDirectAccessGrant,
  createPromoCode,
  deleteDirectAccessGrant,
  fetchPlatformAccessOverview,
  updatePromoCodeStatus,
  type DirectAccessGrant,
  type PromoCode,
} from '@/utils/accessHttpHandler'
import { requestConfirmation } from '@/utils/confirm'
import { handleActionError } from '@/utils/errors/handleActionError'
import { emitToastSuccess } from '@/utils/toast'
import type { BillingPlan } from '@/constants/billing'

const directGrants = ref<DirectAccessGrant[]>([])
const promoCodes = ref<PromoCode[]>([])
const isLoading = ref(false)
const hasLoaded = ref(false)

const grantEmail = ref('')
const grantPlan = ref<BillingPlan>('single')
const grantNote = ref('')
const promoCode = ref('')
const promoDurationDays = ref<number | null>(14)
const grantPlanOptions: Array<{ id: BillingPlan; name: string }> = [
  { id: 'single', name: 'Single access' },
  { id: 'team', name: 'Team access' },
]
const selectedGrantPlan = computed({
  get: () =>
    grantPlanOptions.find((option) => option.id === grantPlan.value) ?? grantPlanOptions[0]!,
  set: (option: { id: BillingPlan; name: string } | null) => {
    if (option) {
      grantPlan.value = option.id
    }
  },
})

const grantFieldErrors = ref<Record<string, string>>({})
const promoFieldErrors = ref<Record<string, string>>({})

const isSubmittingGrant = ref(false)
const isSubmittingPromo = ref(false)
const deletingGrantId = ref<number | null>(null)
const togglingPromoId = ref<number | null>(null)

async function loadOverview(force = false) {
  if (hasLoaded.value && !force) return

  isLoading.value = true
  try {
    const overview = await fetchPlatformAccessOverview()
    directGrants.value = Array.isArray(overview.directGrants) ? overview.directGrants : []
    promoCodes.value = Array.isArray(overview.promoCodes) ? overview.promoCodes : []
    hasLoaded.value = true
  } catch (err) {
    handleActionError(err, {
      toastTitle: 'Could not load access tools',
      mapFields: false,
    })
  } finally {
    isLoading.value = false
  }
}

async function submitDirectGrant() {
  grantFieldErrors.value = {}
  isSubmittingGrant.value = true
  try {
    await createDirectAccessGrant(grantEmail.value, grantPlan.value, grantNote.value)
    grantEmail.value = ''
    grantPlan.value = 'single'
    grantNote.value = ''
    emitToastSuccess('Direct access grant created.')
    await loadOverview(true)
  } catch (err) {
    handleActionError(err, {
      fieldErrors: grantFieldErrors,
      toastTitle: 'Could not create direct access grant',
    })
  } finally {
    isSubmittingGrant.value = false
  }
}

async function removeDirectGrant(grant: DirectAccessGrant) {
  const confirmed = await requestConfirmation({
    title: 'Remove direct access',
    message: `Remove direct access for ${grant.email}?`,
    details:
      'That workspace owner will be asked for billing again unless they already have another access source.',
    confirmLabel: 'Remove access',
    confirmVariant: 'danger',
  })
  if (!confirmed) return

  deletingGrantId.value = grant.id
  try {
    await deleteDirectAccessGrant(grant.id)
    emitToastSuccess('Direct access removed.')
    await loadOverview(true)
  } catch (err) {
    handleActionError(err, {
      toastTitle: 'Could not remove direct access',
      mapFields: false,
    })
  } finally {
    deletingGrantId.value = null
  }
}

async function submitPromoCode() {
  promoFieldErrors.value = {}
  isSubmittingPromo.value = true
  try {
    await createPromoCode(promoCode.value, promoDurationDays.value ?? 0)
    promoCode.value = ''
    promoDurationDays.value = 14
    emitToastSuccess('Promo code created.')
    await loadOverview(true)
  } catch (err) {
    handleActionError(err, {
      fieldErrors: promoFieldErrors,
      toastTitle: 'Could not create promo code',
    })
  } finally {
    isSubmittingPromo.value = false
  }
}

async function togglePromo(promo: PromoCode) {
  togglingPromoId.value = promo.id
  try {
    await updatePromoCodeStatus(promo.id, !promo.active)
    emitToastSuccess(promo.active ? 'Promo code disabled.' : 'Promo code reactivated.')
    await loadOverview(true)
  } catch (err) {
    handleActionError(err, {
      toastTitle: 'Could not update promo code',
      mapFields: false,
    })
  } finally {
    togglingPromoId.value = null
  }
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

onMounted(async () => {
  await loadOverview()
})
</script>

<template>
  <TeamCard
    title="Trusted access"
    :icon="KeyIcon"
  >
    <template #text>Grant permanent workspace access to specific owner emails.</template>

    <form
      class="flex w-full flex-col gap-2 sm:flex-row"
      @submit.prevent="submitDirectGrant"
    >
      <TheInput
        v-model="grantEmail"
        label="Email"
        type="email"
        placeholder="trusted@company.com"
        class="grow"
        :error="grantFieldErrors.email ?? null"
      />

      <TheDropdown
        v-model="selectedGrantPlan"
        :options="grantPlanOptions"
        select-title="Access"
        label-key="name"
        value-key="id"
        placeholder="Select access"
        class="grow"
      />

      <TheInput
        v-model="grantNote"
        label="Note"
        placeholder="Identifier"
        :error="grantFieldErrors.note ?? null"
      />

      <div class="place-self-center">
        <TheButton
          type="submit"
          :disabled="isSubmittingGrant"
          class="cursor-pointer py-2.5"
        >
          {{ isSubmittingGrant ? 'Saving...' : 'Add access' }}
        </TheButton>
      </div>
    </form>

    <div
      v-if="isLoading && !hasLoaded"
      class="mt-4 text-sm text-zinc-600 dark:text-zinc-400"
    >
      Loading direct access grants...
    </div>

    <div
      v-else-if="directGrants.length === 0"
      class="mt-4 rounded-2xl border border-dashed border-zinc-300 px-4 py-3 text-sm text-zinc-600 dark:border-zinc-800 dark:text-zinc-400"
    >
      No direct access grants yet.
    </div>

    <div
      v-else
      class="mt-4 space-y-2.5"
    >
      <article
        v-for="grant in directGrants"
        :key="grant.id"
        class="flex flex-col gap-3 rounded-2xl border border-zinc-300 bg-zinc-50/70 px-4 py-3 sm:flex-row sm:items-center sm:justify-between dark:border-zinc-800 dark:bg-zinc-950/40"
      >
        <div class="min-w-0">
          <p class="truncate text-sm font-semibold text-zinc-900 dark:text-zinc-100">
            {{ grant.email }}
          </p>
          <p
            class="mt-1 text-[11px] font-semibold tracking-[0.16em] text-zinc-500 uppercase dark:text-zinc-500"
          >
            {{ grant.plan === 'team' ? 'Team access' : 'Single access' }}
          </p>
          <p class="mt-1 text-xs leading-6 text-zinc-600 dark:text-zinc-400">
            {{ grant.note || 'No note added.' }} Added {{ formatDate(grant.createdAt) }}.
          </p>
        </div>

        <TheButton
          variant="secondary"
          class="cursor-pointer"
          :disabled="deletingGrantId === grant.id"
          @click="removeDirectGrant(grant)"
        >
          <TrashIcon class="size-4" />
          {{ deletingGrantId === grant.id ? 'Removing...' : 'Remove' }}
        </TheButton>
      </article>
    </div>
  </TeamCard>

  <TeamCard
    title="Promo codes"
    :icon="SparklesIcon"
  >
    <template #text>
      Promo codes skip Stripe checkout for a fixed number of days. Each workspace can use a code
      once.
    </template>

    <form
      class="flex w-full flex-col gap-2 sm:flex-row"
      @submit.prevent="submitPromoCode"
    >
      <TheInput
        v-model="promoCode"
        label="Promo code"
        placeholder="EARLYBIRD14"
        class="grow"
        :error="promoFieldErrors.code ?? null"
      />

      <TheInput
        v-model="promoDurationDays"
        label="Days"
        type="number"
        min="1"
        placeholder="14"
        :error="promoFieldErrors.durationDays ?? null"
      />

      <div class="place-self-center">
        <TheButton
          type="submit"
          :disabled="isSubmittingPromo"
          class="cursor-pointer py-2.5"
        >
          {{ isSubmittingPromo ? 'Saving...' : 'Create code' }}
        </TheButton>
      </div>
    </form>

    <div
      v-if="promoCodes.length === 0 && hasLoaded"
      class="mt-4 rounded-2xl border border-dashed border-zinc-300 px-4 py-3 text-sm text-zinc-600 dark:border-zinc-800 dark:text-zinc-400"
    >
      No promo codes yet.
    </div>

    <div class="mt-4 space-y-2.5">
      <article
        v-for="promo in promoCodes"
        :key="promo.id"
        class="flex flex-col gap-3 rounded-2xl border border-zinc-300 bg-zinc-50/70 px-4 py-3 sm:flex-row sm:items-center sm:justify-between dark:border-zinc-800 dark:bg-zinc-950/40"
      >
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <p class="text-sm font-semibold text-zinc-900 dark:text-zinc-100">
              {{ promo.code }}
            </p>
            <span
              class="rounded-full px-2.5 py-1 text-[11px] font-semibold tracking-[0.16em] uppercase"
              :class="
                promo.active
                  ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950/50 dark:text-emerald-200'
                  : 'bg-zinc-200 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-200'
              "
            >
              {{ promo.active ? 'active' : 'inactive' }}
            </span>
          </div>

          <p class="mt-1 text-xs leading-6 text-zinc-600 dark:text-zinc-400">
            {{ promo.durationDays }} day{{ promo.durationDays === 1 ? '' : 's' }} of access. Used
            {{ promo.redemptionCount }} time{{ promo.redemptionCount === 1 ? '' : 's' }}. Added
            {{ formatDate(promo.createdAt) }}.
          </p>
        </div>

        <TheButton
          variant="secondary"
          class="cursor-pointer"
          :disabled="togglingPromoId === promo.id"
          @click="togglePromo(promo)"
        >
          {{ togglingPromoId === promo.id ? 'Saving...' : promo.active ? 'Disable' : 'Enable' }}
        </TheButton>
      </article>
    </div>
  </TeamCard>
</template>

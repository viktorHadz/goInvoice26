<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
    CheckCircleIcon,
    CreditCardIcon,
    ExclamationTriangleIcon,
    UsersIcon,
} from '@heroicons/vue/24/outline'
import TheButton from '@/components/UI/TheButton.vue'
import { useAuthStore } from '@/stores/auth'
import {
    createCheckoutSession,
    createPortalSession,
    syncCheckoutSession,
} from '@/utils/billingHttpHandler'
import { emitToastError, emitToastSuccess } from '@/utils/toast'
import { getApiErrorMessage, isApiError } from '@/utils/apiErrors'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const isStartingCheckout = ref(false)
const isOpeningPortal = ref(false)
const isSyncingCheckout = ref(false)
const syncedSessionId = ref('')
const CHECKOUT_SYNC_ATTEMPTS = 6
const CHECKOUT_SYNC_DELAY_MS = 1200

const billing = computed(() => authStore.billing)
const isOwner = computed(() => authStore.canManageBilling)
const hasAccess = computed(() => authStore.hasBillingAccess)
const checkoutState = computed(() =>
    typeof route.query.checkout === 'string' ? route.query.checkout : '',
)
const checkoutSessionId = computed(() =>
    typeof route.query.session_id === 'string' ? route.query.session_id : '',
)
const redirectPath = computed(() => {
    const candidate = typeof route.query.redirect === 'string' ? route.query.redirect : '/app'
    return candidate.startsWith('/') ? candidate : '/app'
})
const statusLabel = computed(() => {
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

watch(
    checkoutSessionId,
    async (sessionId) => {
        if (!sessionId || syncedSessionId.value === sessionId || !isOwner.value) return

        syncedSessionId.value = sessionId
        await confirmCheckoutSession(sessionId)
    },
    { immediate: true },
)

async function startCheckout() {
    isStartingCheckout.value = true
    try {
        const session = await createCheckoutSession()
        window.location.assign(session.url)
    } catch (err) {
        emitToastError({
            title: 'Could not start checkout',
            message: isApiError(err)
                ? getApiErrorMessage(err)
                : 'Stripe checkout could not be started right now.',
        })
    } finally {
        isStartingCheckout.value = false
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

function wait(ms: number) {
    return new Promise((resolve) => window.setTimeout(resolve, ms))
}
</script>

<template>
    <main class="mx-auto flex min-h-[calc(100vh-15rem)] w-full max-w-5xl items-center py-10">
        <section
            class="grid w-full gap-6 rounded-4xl border border-zinc-200 bg-[linear-gradient(160deg,#fffdfa_0%,#f5f7f2_48%,#eef5fb_100%)] p-6 shadow-2xl sm:p-8 lg:grid-cols-[1.1fr_0.9fr] dark:border-zinc-800 dark:bg-[linear-gradient(160deg,#18181b_0%,#0c0c12_48%,#09090b_100%)]"
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
                    <p
                        class="mt-4 max-w-2xl text-base leading-8 text-zinc-600 sm:text-lg dark:text-zinc-400"
                    >
                        One workspace subscription covers the account. Teammates use the same
                        account after payment is active.
                    </p>
                </div>

                <div
                    v-if="checkoutState === 'canceled'"
                    class="rounded-3xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm leading-6 text-amber-900 dark:border-amber-200/50 dark:bg-amber-400/20"
                >
                    Checkout was canceled before payment completed. Your workspace data is still
                    safe here.
                </div>

                <div
                    v-else-if="checkoutState === 'success' && isSyncingCheckout"
                    class="rounded-3xl border border-sky-200 bg-sky-50 px-4 py-3 text-sm leading-6 text-sky-800"
                >
                    Confirming the Stripe checkout and unlocking the workspace.
                </div>

                <div
                    v-else-if="hasAccess"
                    class="rounded-3xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm leading-6 text-emerald-800 dark:border-emerald-400/30 dark:bg-emerald-900/20 dark:text-emerald-400"
                >
                    The account can access the full invoicing workspace.
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
                    Billing is temporarily unavailable right now. Please get in touch and we will
                    help you finish setup.
                </div>

                <div class="flex flex-wrap gap-3">
                    <TheButton
                        v-if="isOwner && !hasAccess"
                        :disabled="isStartingCheckout || !billing?.configured"
                        @click="startCheckout"
                    >
                        {{ isStartingCheckout ? 'Opening checkout...' : 'Start £5 / month plan' }}
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
                            class="dark: rounded-2xl bg-zinc-100 p-3 text-zinc-700 dark:bg-zinc-900 dark:text-zinc-300"
                        >
                            <UsersIcon class="size-6" />
                        </div>
                        <div>
                            <h2 class="text-lg font-semibold text-zinc-950 dark:text-zinc-300">
                                Shared workspace
                            </h2>
                            <p class="mt-2 text-sm leading-7 text-zinc-600 dark:text-zinc-500">
                                One subscription unlocks the account for the workspace admin and
                                invited teammates.
                            </p>
                        </div>
                    </div>
                </article>

                <article
                    class="rounded-3xl border border-zinc-300 bg-white/85 p-5 dark:border-zinc-800 dark:bg-zinc-950"
                >
                    <div class="flex items-start gap-3">
                        <div
                            class="dark: rounded-2xl bg-zinc-100 p-3 text-zinc-700 dark:bg-zinc-900 dark:text-zinc-300"
                        >
                            <CheckCircleIcon class="size-6" />
                        </div>
                        <div>
                            <h2 class="text-lg font-semibold text-zinc-950 dark:text-zinc-300">
                                What you get
                            </h2>
                            <p class="mt-2 text-sm leading-7 text-zinc-600 dark:text-zinc-500">
                                Full access to clients, invoices, editor, settings, and document
                                export for this account.
                            </p>
                        </div>
                    </div>
                </article>

                <article
                    class="rounded-3xl border border-zinc-300 bg-white/85 p-5 dark:border-zinc-800 dark:bg-zinc-950"
                >
                    <div class="flex items-start gap-3">
                        <div
                            class="dark: rounded-2xl bg-zinc-100 p-3 text-zinc-700 dark:bg-zinc-900 dark:text-zinc-300"
                        >
                            <ExclamationTriangleIcon class="size-6" />
                        </div>
                        <div>
                            <h2 class="text-lg font-semibold text-zinc-950 dark:text-zinc-300">
                                Past-due handling
                            </h2>
                            <p class="mt-2 text-sm leading-7 text-zinc-600 dark:text-zinc-500">
                                If renewal payment fails, access stays blocked until the workspace
                                admin fixes billing. Data is retained even if the subscription
                                cancels.
                            </p>
                            <p
                                v-if="billing?.currentPeriodEnd"
                                class="mt-3 text-xs font-medium tracking-[0.16em] text-zinc-500 uppercase"
                            >
                                Current period end: {{ billing.currentPeriodEnd }}
                            </p>
                        </div>
                    </div>
                </article>
            </aside>
        </section>
    </main>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import TheButton from '@/components/UI/TheButton.vue'
import { useAuthStore } from '@/stores/auth'
import {
    ArrowLeftIcon,
    EnvelopeIcon,
    GlobeAltIcon,
    ShieldCheckIcon,
} from '@heroicons/vue/24/outline'

const props = defineProps<{
    mode: 'login' | 'signup'
}>()

const route = useRoute()
const authStore = useAuthStore()
const isSignup = computed(() => props.mode === 'signup')
const title = computed(() =>
    isSignup.value ? 'Create the owner account' : 'Sign in to the workspace',
)
const subtitle = computed(() =>
    isSignup.value
        ? 'The first Google signup claims the owner seat and connects it to the shared business workspace.'
        : 'Use the Google account already linked to this workspace. Invite-based teammate access can layer on after this.',
)
const redirectPath = computed(() =>
    typeof route.query.redirect === 'string' && route.query.redirect.startsWith('/')
        ? route.query.redirect
        : '/app',
)
const googleDisabled = computed(
    () => !authStore.googleEnabled || (isSignup.value && !authStore.needsSetup),
)
const setupClosed = computed(() => isSignup.value && authStore.hasLoaded && !authStore.needsSetup)
const googleButtonLabel = computed(() => {
    if (!authStore.googleEnabled) return 'Google auth needs env setup'
    if (setupClosed.value) return 'Owner account already exists'
    return isSignup.value ? 'Register with Google' : 'Log in with Google'
})
const errorCode = computed(() => (typeof route.query.error === 'string' ? route.query.error : ''))
const errorMessage = computed(() => authErrorMessage(errorCode.value))

const options = computed(() => [
    {
        title: isSignup.value ? 'Register with Google' : 'Log in with Google',
        body: isSignup.value
            ? 'Best first path for the owner. It creates the initial user session without adding password reset and verification work up front.'
            : 'Use the Google account that already belongs to this workspace so we can restore the session with a secure server cookie.',
        icon: GlobeAltIcon,
    },
    {
        title: isSignup.value ? 'Register with email' : 'Log in with email',
        body: 'This stays parked for now while we finish the owner flow and protect the app area first.',
        icon: EnvelopeIcon,
    },
])

function startGoogleAuth() {
    if (googleDisabled.value) return
    authStore.beginGoogleAuth(props.mode, redirectPath.value)
}

function authErrorMessage(code: string) {
    const messages: Record<string, string> = {
        google_not_configured:
            'Google auth is not configured on the backend yet. Add the Google env vars before trying again.',
        invalid_auth_mode:
            'This sign-in request was malformed. Please try again from the login page.',
        invalid_oauth_state: 'We could not verify the Google sign-in handshake. Please try again.',
        missing_oauth_code: 'Google did not return a login code. Please try again.',
        google_access_denied: 'Google sign-in was cancelled before it completed.',
        google_auth_failed: 'Google sign-in did not complete. Please try again.',
        google_email_not_verified:
            'The selected Google account must have a verified email address.',
        owner_setup_complete:
            'The owner account has already been created. Use the login flow instead of signup.',
        owner_setup_required: 'Create the owner account first before using the login flow.',
        account_not_linked:
            'This Google account is not linked to the workspace yet. Ask the owner to invite it first.',
        account_conflict:
            'This email is already linked to a different Google login. Use the original account or update the user record first.',
    }

    return code
        ? (messages[code] ?? 'Authentication could not be completed. Please try again.')
        : ''
}
</script>

<template>
    <main
        class="min-h-screen bg-[linear-gradient(180deg,#f7f3ea_0%,#fffdfa_38%,#eef5f2_100%)] text-zinc-900"
    >
        <section class="mx-auto flex min-h-screen w-full max-w-6xl items-center px-5 py-8 sm:px-8">
            <div class="grid w-full gap-8 lg:grid-cols-[0.92fr_1.08fr]">
                <aside
                    class="rounded-2xl border border-stone-200 bg-white/80 p-6 shadow-xl shadow-stone-200/60 backdrop-blur sm:p-8"
                >
                    <RouterLink
                        to="/"
                        class="inline-flex items-center gap-2 text-sm font-medium text-zinc-600 transition hover:text-zinc-900"
                    >
                        <ArrowLeftIcon class="size-4" />
                        Back to homepage
                    </RouterLink>

                    <div
                        class="mt-8 inline-flex items-center gap-2 rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1 text-xs font-semibold tracking-[0.2em] text-emerald-700 uppercase"
                    >
                        <ShieldCheckIcon class="size-4" />
                        Owner auth ready
                    </div>

                    <h1
                        class="mt-5 text-4xl leading-tight font-semibold tracking-tight text-zinc-950 sm:text-5xl"
                    >
                        {{ title }}
                    </h1>

                    <p class="mt-4 max-w-xl text-base leading-8 text-zinc-600 sm:text-lg">
                        {{ subtitle }}
                    </p>

                    <div
                        v-if="errorMessage"
                        class="mt-6 rounded-3xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm leading-6 text-rose-700"
                    >
                        {{ errorMessage }}
                    </div>

                    <div
                        v-else-if="setupClosed"
                        class="mt-6 rounded-3xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm leading-6 text-amber-800"
                    >
                        The owner account already exists for this workspace. Teammates should use
                        the login page once their access is linked.
                    </div>

                    <div class="mt-8 flex flex-wrap gap-3">
                        <RouterLink :to="isSignup ? '/login' : '/'">
                            <TheButton>
                                {{ isSignup ? 'Go to login' : 'Back to homepage' }}
                            </TheButton>
                        </RouterLink>
                        <RouterLink
                            v-if="!isSignup"
                            to="/signup"
                        >
                            <TheButton variant="secondary">Create owner account</TheButton>
                        </RouterLink>
                    </div>
                </aside>

                <section class="grid gap-4">
                    <article
                        v-for="option in options"
                        :key="option.title"
                        class="rounded-2xl border border-zinc-300 bg-white p-6 shadow-lg shadow-zinc-200/50 sm:p-7"
                    >
                        <div class="flex items-start justify-between gap-4">
                            <div>
                                <div
                                    class="inline-flex items-center rounded-full border border-amber-200 bg-amber-50 px-3 py-1 text-xs font-semibold tracking-[0.18em] text-amber-700 uppercase"
                                >
                                    Coming next
                                </div>
                                <h2 class="mt-4 text-2xl font-semibold text-zinc-950">
                                    {{ option.title }}
                                </h2>
                                <p class="mt-3 max-w-xl text-sm leading-7 text-zinc-600">
                                    {{ option.body }}
                                </p>
                            </div>

                            <div
                                class="rounded-2xl border border-zinc-300 bg-zinc-50 p-3 text-zinc-700"
                            >
                                <component
                                    :is="option.icon"
                                    class="size-6"
                                />
                            </div>
                        </div>

                        <div class="mt-6 flex flex-wrap gap-3">
                            <TheButton
                                :variant="option.title.includes('Google') ? 'primary' : 'secondary'"
                                :disabled="option.title.includes('Google') ? googleDisabled : true"
                                @click="
                                    option.title.includes('Google') ? startGoogleAuth() : undefined
                                "
                            >
                                {{
                                    option.title.includes('Google')
                                        ? googleButtonLabel
                                        : option.title
                                }}
                            </TheButton>
                            <span
                                class="inline-flex items-center rounded-full bg-zinc-100 px-3 py-2 text-xs font-medium text-zinc-600"
                            >
                                {{
                                    option.title.includes('Google')
                                        ? authStore.googleEnabled
                                            ? 'Server session cookie after Google callback'
                                            : 'Waiting for Google env vars'
                                        : 'Email flow comes after Google'
                                }}
                            </span>
                        </div>
                    </article>
                </section>
            </div>
        </section>
    </main>
</template>

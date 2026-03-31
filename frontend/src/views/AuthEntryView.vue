<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import TheButton from '@/components/UI/TheButton.vue'
import { useAuthStore } from '@/stores/auth'
import {
    ArrowLeftIcon,
    CheckCircleIcon,
    GlobeAltIcon,
    ShieldCheckIcon,
} from '@heroicons/vue/24/outline'

const props = defineProps<{
    mode: 'login' | 'signup'
}>()

const route = useRoute()
const authStore = useAuthStore()

const isSignup = computed(() => props.mode === 'signup')
const title = computed(() => (isSignup.value ? 'Register' : 'Log in'))
const subtitle = computed(() =>
    isSignup.value
        ? 'Sign in with Google to create your workspace. Billing and team access can be set up after registration.'
        : 'Use the Google account that already has access to your workspace.',
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
    if (!authStore.googleEnabled) return 'Google sign-in unavailable'
    if (setupClosed.value) return 'Use log in instead'
    return isSignup.value ? 'Continue with Google' : 'Log in with Google'
})
const errorCode = computed(() => (typeof route.query.error === 'string' ? route.query.error : ''))
const errorMessage = computed(() => authErrorMessage(errorCode.value))

const quickNotes = computed(() =>
    isSignup.value
        ? [
              'Register once to create the workspace.',
              'Activate billing after sign-up.',
              'Invite teammates later from inside the app.',
          ]
        : [
              'Use the Google account linked to the workspace.',
              'If you were invited, use that same email on Google.',
              'Billing is handled by the workspace admin, not each teammate.',
          ],
)

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
        owner_setup_complete: 'This workspace is already set up. Please log in instead.',
        owner_setup_required: 'No workspace has been created yet. Please register first.',
        account_not_linked:
            "This Google account doesn't have access yet. Ask your workspace admin for an invite.",
        account_conflict:
            'This email is already linked to a different Google account. Please use the original login or contact support.',
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
            <div class="grid w-full gap-8 lg:grid-cols-[0.95fr_1.05fr]">
                <aside
                    class="rounded-2xl border border-stone-200 bg-white/80 p-6 shadow-xl shadow-stone-200/60 backdrop-blur sm:p-8"
                >
                    <div class="flex items-center justify-between">
                        <RouterLink
                            to="/"
                            class="inline-flex items-center gap-2 text-sm font-medium text-zinc-600 transition hover:text-zinc-900"
                        >
                            <ArrowLeftIcon class="size-4" />
                            Back to homepage
                        </RouterLink>
                        <div
                            class="mt-8 flex items-center gap-2 rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1 text-xs font-semibold text-emerald-700 uppercase"
                        >
                            <ShieldCheckIcon class="size-4" />
                            Secure Google sign-in
                        </div>
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
                        This workspace is already set up. Please use the log in page instead.
                    </div>

                    <div class="mt-8 flex flex-wrap gap-3">
                        <RouterLink :to="isSignup ? '/login' : '/signup'">
                            <TheButton variant="secondary">
                                {{
                                    isSignup
                                        ? 'Already have access? Log in'
                                        : 'Need a new workspace? Register'
                                }}
                            </TheButton>
                        </RouterLink>
                    </div>
                </aside>

                <section class="grid gap-4">
                    <article
                        class="rounded-2xl border border-zinc-300 bg-white p-6 shadow-lg shadow-zinc-200/50 sm:p-7"
                    >
                        <div class="flex items-start justify-between gap-4">
                            <div>
                                <div
                                    class="inline-flex items-center rounded-full border border-amber-200 bg-amber-50 px-3 py-1 text-xs font-semibold tracking-[0.18em] text-amber-700 uppercase"
                                >
                                    Google
                                </div>
                                <h2 class="mt-4 text-2xl font-semibold text-zinc-950">
                                    {{ isSignup ? 'Register with Google' : 'Log in with Google' }}
                                </h2>
                                <p class="mt-3 max-w-xl text-sm leading-7 text-zinc-600">
                                    {{
                                        isSignup
                                            ? 'This is the fastest way to create the workspace and start the billing flow.'
                                            : 'This restores your session with the Google account already linked to the workspace.'
                                    }}
                                </p>
                            </div>

                            <div
                                class="rounded-2xl border border-zinc-300 bg-zinc-50 p-3 text-zinc-700"
                            >
                                <GlobeAltIcon class="size-6" />
                            </div>
                        </div>

                        <div class="mt-6 flex flex-wrap gap-3">
                            <TheButton
                                :disabled="googleDisabled"
                                @click="startGoogleAuth"
                            >
                                {{ googleButtonLabel }}
                            </TheButton>
                            <span
                                class="inline-flex items-center rounded-full bg-zinc-100 px-3 py-2 text-xs font-medium text-zinc-600"
                            >
                                Uses a secure session cookie after Google sign-in
                            </span>
                        </div>
                    </article>

                    <article
                        class="rounded-2xl border border-zinc-300 bg-white p-6 shadow-lg shadow-zinc-200/50 sm:p-7"
                    >
                        <div
                            class="text-xs font-semibold tracking-[0.18em] text-zinc-500 uppercase"
                        >
                            Helpful to know
                        </div>

                        <ul class="mt-4 grid gap-3 text-sm text-zinc-700">
                            <li
                                v-for="note in quickNotes"
                                :key="note"
                                class="rounded-2xl border border-zinc-200 bg-zinc-50 px-4 py-3"
                            >
                                <div class="flex items-start gap-3">
                                    <CheckCircleIcon class="mt-0.5 size-4 text-emerald-600" />
                                    <span>{{ note }}</span>
                                </div>
                            </li>
                        </ul>

                        <div
                            class="mt-6 rounded-2xl border border-sky-200 bg-sky-50 px-4 py-3 text-sm leading-6 text-sky-800"
                        >
                            We use a necessary secure cookie to keep you signed in after Google
                            authentication. If you block that cookie, sign-in will not work.
                            <RouterLink
                                to="/privacy"
                                class="ml-1 font-semibold underline decoration-sky-400 underline-offset-2"
                            >
                                Learn more
                            </RouterLink>
                        </div>
                    </article>
                </section>
            </div>
        </section>
    </main>
</template>

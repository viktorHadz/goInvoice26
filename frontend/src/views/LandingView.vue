<script setup lang="ts">
import { RouterLink } from 'vue-router'
import {
    ArrowRightIcon,
    CheckBadgeIcon,
    CreditCardIcon,
    SparklesIcon,
    UserGroupIcon,
} from '@heroicons/vue/24/outline'

const proofPoints = [
    'One business account with shared clients, products, and invoice settings',
    'Owner signup first, then teammates join the same workspace later',
    'Prepared for Google auth, billing, and team invites without reworking the app shell',
]

const workflow = [
    {
        title: 'Start with the owner',
        body: 'Create the business account once, keep ownership clear, and attach billing to the account instead of a single user.',
    },
    {
        title: 'Invite the team',
        body: 'Give employees access to the same shared workspace without duplicating products, clients, or invoice branding.',
    },
    {
        title: 'Scale without rewiring',
        body: 'The app area already sits behind /app, so auth and protected routes can slot in cleanly next.',
    },
]
</script>

<template>
    <main class="relative min-h-screen overflow-hidden bg-stone-950 text-stone-100">
        <div class="absolute inset-0 overflow-hidden">
            <div
                class="absolute top-[-10rem] left-[-6rem] h-72 w-72 rounded-full bg-amber-300/20 blur-3xl"
            />
            <div
                class="absolute right-[-8rem] bottom-[-8rem] h-96 w-96 rounded-full bg-cyan-400/15 blur-3xl"
            />
            <div
                class="absolute inset-x-0 top-0 h-80 bg-[radial-gradient(circle_at_top,rgba(245,158,11,0.18),transparent_50%)]"
            />
            <div
                class="absolute inset-0 [background-image:linear-gradient(rgba(255,255,255,0.06)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.06)_1px,transparent_1px)] [background-size:4.5rem_4.5rem] opacity-20"
            />
        </div>

        <section
            class="relative mx-auto flex min-h-screen w-full max-w-7xl flex-col px-5 py-6 sm:px-8 lg:px-10"
        >
            <header class="flex items-center justify-between gap-4">
                <RouterLink
                    to="/"
                    class="inline-flex items-center gap-3 rounded-full border border-white/12 bg-white/6 px-4 py-2 text-sm font-semibold tracking-[0.24em] uppercase backdrop-blur"
                >
                    <span
                        class="inline-flex h-2.5 w-2.5 rounded-full bg-amber-300 shadow-[0_0_18px_rgba(252,211,77,0.7)]"
                    />
                    Invoicer
                </RouterLink>

                <div class="flex items-center gap-2">
                    <RouterLink
                        to="/login"
                        class="rounded-full border border-white/12 px-4 py-2 text-sm font-medium text-stone-200 transition hover:border-cyan-300/50 hover:bg-white/8 hover:text-white"
                    >
                        Log in
                    </RouterLink>
                    <RouterLink
                        to="/signup"
                        class="rounded-full bg-amber-300 px-4 py-2 text-sm font-semibold text-stone-950 transition hover:bg-amber-200"
                    >
                        Start account
                    </RouterLink>
                </div>
            </header>

            <div class="grid flex-1 items-center gap-12 py-12 lg:grid-cols-[1.1fr_0.9fr] lg:py-16">
                <section class="max-w-3xl">
                    <div
                        class="inline-flex items-center gap-2 rounded-full border border-cyan-300/20 bg-cyan-300/8 px-3 py-1 text-xs font-semibold tracking-[0.2em] text-cyan-100 uppercase"
                    >
                        <SparklesIcon class="size-4" />
                        Sensible rollout for auth and billing
                    </div>

                    <h1
                        class="mt-6 max-w-4xl text-5xl leading-none font-semibold tracking-tight text-stone-50 sm:text-6xl lg:text-7xl"
                    >
                        Build the public front door
                        <span class="block text-amber-300">before the account system.</span>
                    </h1>

                    <p class="mt-6 max-w-2xl text-lg leading-8 text-stone-300 sm:text-xl">
                        The invoicer workspace now lives behind
                        <code class="rounded bg-white/8 px-2 py-0.5 text-base text-stone-100">
                            /app
                        </code>
                        , which gives us a clean path for Google auth, account creation, billing,
                        and teammate invites without twisting the existing UI around later.
                    </p>

                    <div class="mt-8 flex flex-wrap items-center gap-3">
                        <RouterLink
                            to="/signup"
                            class="inline-flex items-center gap-2 rounded-full bg-amber-300 px-5 py-3 text-sm font-semibold text-stone-950 transition hover:bg-amber-200"
                        >
                            Create account
                            <ArrowRightIcon class="size-4" />
                        </RouterLink>
                        <RouterLink
                            to="/login"
                            class="inline-flex items-center gap-2 rounded-full border border-white/12 bg-white/6 px-5 py-3 text-sm font-medium text-stone-100 backdrop-blur transition hover:border-cyan-300/50 hover:bg-white/10"
                        >
                            Explore sign-in paths
                        </RouterLink>
                        <RouterLink
                            to="/app"
                            class="inline-flex items-center gap-2 rounded-full px-4 py-3 text-sm font-medium text-stone-300 transition hover:text-white"
                        >
                            Open current workspace
                        </RouterLink>
                    </div>

                    <ul class="mt-10 grid gap-3 text-sm text-stone-200 sm:grid-cols-3">
                        <li
                            v-for="point in proofPoints"
                            :key="point"
                            class="rounded-3xl border border-white/10 bg-white/6 p-4 backdrop-blur"
                        >
                            <CheckBadgeIcon class="mb-3 size-5 text-amber-300" />
                            {{ point }}
                        </li>
                    </ul>
                </section>

                <aside class="relative">
                    <div
                        class="rounded-[2rem] border border-white/12 bg-white/7 p-5 shadow-2xl shadow-black/30 backdrop-blur-xl"
                    >
                        <div class="rounded-[1.5rem] border border-white/10 bg-stone-900/85 p-5">
                            <div class="flex items-start justify-between gap-4">
                                <div>
                                    <div
                                        class="text-xs font-semibold tracking-[0.22em] text-stone-400 uppercase"
                                    >
                                        Account Blueprint
                                    </div>
                                    <div class="mt-2 text-2xl font-semibold text-stone-50">
                                        Owner first, team second
                                    </div>
                                </div>
                                <div
                                    class="rounded-2xl border border-amber-300/25 bg-amber-300/12 p-3 text-amber-200"
                                >
                                    <UserGroupIcon class="size-6" />
                                </div>
                            </div>

                            <div class="mt-6 space-y-4">
                                <div class="rounded-3xl border border-white/8 bg-white/5 p-4">
                                    <div class="flex items-center justify-between text-sm">
                                        <span class="font-semibold text-stone-100">Step 1</span>
                                        <span class="text-stone-400">Public signup</span>
                                    </div>
                                    <p class="mt-2 text-sm leading-6 text-stone-300">
                                        Let the owner create the business account by email or Google
                                        without mixing billing into the very first save path.
                                    </p>
                                </div>

                                <div class="rounded-3xl border border-white/8 bg-white/5 p-4">
                                    <div class="flex items-center justify-between text-sm">
                                        <span class="font-semibold text-stone-100">Step 2</span>
                                        <span class="text-stone-400">Attach billing</span>
                                    </div>
                                    <p class="mt-2 text-sm leading-6 text-stone-300">
                                        Bill the account, not the user, so plan status and invoice
                                        ownership stay stable when teammates join later.
                                    </p>
                                </div>

                                <div class="rounded-3xl border border-white/8 bg-white/5 p-4">
                                    <div class="flex items-center justify-between text-sm">
                                        <span class="font-semibold text-stone-100">Step 3</span>
                                        <span class="text-stone-400">Invite teammates</span>
                                    </div>
                                    <p class="mt-2 text-sm leading-6 text-stone-300">
                                        Employees log into the same account workspace and share
                                        clients, products, settings, and logos.
                                    </p>
                                </div>
                            </div>

                            <div class="mt-6 grid gap-3 sm:grid-cols-2">
                                <div
                                    class="rounded-3xl border border-cyan-300/15 bg-cyan-300/8 p-4"
                                >
                                    <div
                                        class="flex items-center gap-2 text-sm font-semibold text-cyan-100"
                                    >
                                        <CreditCardIcon class="size-4" />
                                        Billing stays account-scoped
                                    </div>
                                    <p class="mt-2 text-sm leading-6 text-cyan-50/85">
                                        Cleaner subscription logic and fewer edge cases when
                                        ownership changes.
                                    </p>
                                </div>

                                <div
                                    class="rounded-3xl border border-amber-300/15 bg-amber-300/10 p-4"
                                >
                                    <div class="text-sm font-semibold text-amber-100">
                                        Right next step
                                    </div>
                                    <p class="mt-2 text-sm leading-6 text-amber-50/85">
                                        Finish homepage and auth entry points first, then wire the
                                        real session and payment flow.
                                    </p>
                                </div>
                            </div>
                        </div>
                    </div>
                </aside>
            </div>

            <section class="relative mt-4 pb-8">
                <div class="grid gap-4 lg:grid-cols-3">
                    <article
                        v-for="item in workflow"
                        :key="item.title"
                        class="rounded-[1.75rem] border border-white/10 bg-white/6 p-5 backdrop-blur"
                    >
                        <div
                            class="text-xs font-semibold tracking-[0.22em] text-stone-400 uppercase"
                        >
                            Workflow
                        </div>
                        <h2 class="mt-3 text-xl font-semibold text-stone-50">{{ item.title }}</h2>
                        <p class="mt-3 text-sm leading-7 text-stone-300">{{ item.body }}</p>
                    </article>
                </div>
            </section>
        </section>
    </main>
</template>

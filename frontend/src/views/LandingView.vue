<script setup lang="ts">
import { RouterLink } from 'vue-router'
import {
    ArrowRightIcon,
    BanknotesIcon,
    BriefcaseIcon,
    CheckCircleIcon,
    DocumentTextIcon,
    FolderOpenIcon,
    PencilSquareIcon,
    ShieldCheckIcon,
    SparklesIcon,
    SwatchIcon,
    UsersIcon,
} from '@heroicons/vue/24/outline'

const highlights = [
    'Register with Google to create the workspace.',
    'One subscription unlocks the account for the whole team.',
    'Invite teammates into the same shared invoice workspace.',
]

const workspaceCards = [
    {
        title: 'Brand the account once',
        body: 'Save company details, payment terms, and logo once so every invoice starts from the right foundation.',
        icon: SwatchIcon,
        accent: 'from-amber-100 to-white',
    },
    {
        title: 'Build invoices quickly',
        body: 'Reuse saved clients and items, then review totals, discounts, deposits, and due dates without leaving the flow.',
        icon: DocumentTextIcon,
        accent: 'from-emerald-100 to-white',
    },
    {
        title: 'Keep revisions tidy',
        body: 'Open past invoices in the editor, make revisions safely, and keep a cleaner record of what changed.',
        icon: PencilSquareIcon,
        accent: 'from-sky-100 to-white',
    },
]

const featureExamples = [
    {
        title: 'Consultant retainer',
        summary:
            'Save one client, reuse service items monthly, and export a fresh invoice in minutes.',
        eyebrow: 'Example 01',
        rows: [
            'Client: North Pier Studio',
            'Items: Strategy session, edits, travel',
            'Export: PDF or DOCX',
        ],
    },
    {
        title: 'Small studio team',
        summary:
            'One admin subscribes once, teammates sign in with Google, and everyone works inside the same account.',
        eyebrow: 'Example 02',
        rows: [
            'Admin account + subscription',
            'Invite teammate emails',
            'Shared clients, shared records',
        ],
    },
    {
        title: 'Revision-heavy work',
        summary:
            'If a client asks for a change after sending, reopen the invoice, create a revision, and keep the history intact.',
        eyebrow: 'Example 03',
        rows: ['Open invoice book', 'Create revision', 'Regenerate the document'],
    },
]

const billingSteps = [
    {
        title: 'Register the workspace',
        body: 'Start with Google sign-up and create the workspace.',
    },
    {
        title: 'Activate the workspace',
        body: 'Subscribe once for GBP 5 per month to unlock the app for the account.',
    },
    {
        title: 'Invite the team',
        body: 'Add teammate emails so they can log in with Google and join the same workspace.',
    },
    {
        title: 'Run invoicing from one place',
        body: 'Clients, items, invoices, settings, exports, and revisions stay together in one shared account.',
    },
]

const audienceTags = ['Freelancers', 'Consultants', 'Studios', 'Trades', 'Small teams']
</script>

<template>
    <main
        class="min-h-screen bg-[linear-gradient(180deg,#f7f2e8_0%,#fbfaf6_38%,#eef4f1_100%)] text-zinc-900"
    >
        <div class="relative overflow-hidden">
            <div
                class="absolute inset-x-0 top-0 h-140 bg-[radial-gradient(circle_at_top,#f8d9a8_0%,rgba(248,217,168,0.28)_32%,transparent_68%)]"
            />
            <div
                class="absolute top-24 -right-32 h-72 w-72 rounded-full bg-emerald-200/30 blur-3xl"
            />
            <div class="absolute top-64 -left-24 h-72 w-72 rounded-full bg-sky-200/30 blur-3xl" />

            <section class="relative mx-auto w-full max-w-7xl px-5 py-6 sm:px-8 lg:px-10">
                <header class="flex items-center justify-between gap-4">
                    <RouterLink
                        to="/"
                        class="inline-flex items-center gap-3 rounded-full border border-zinc-300 bg-white/90 px-4 py-2 text-sm font-semibold tracking-[0.22em] text-zinc-900 uppercase shadow-sm backdrop-blur"
                    >
                        <span class="inline-flex h-2.5 w-2.5 rounded-full bg-amber-500" />
                        Invoicer
                    </RouterLink>

                    <div class="flex items-center gap-2">
                        <RouterLink
                            to="/login"
                            class="rounded-full border border-zinc-300 bg-white/85 px-4 py-2 text-sm font-medium text-zinc-700 transition hover:border-zinc-400 hover:text-zinc-950"
                        >
                            Log in
                        </RouterLink>
                        <RouterLink
                            to="/signup"
                            class="rounded-full bg-zinc-950 px-4 py-2 text-sm font-semibold text-white transition hover:bg-zinc-800"
                        >
                            Register
                        </RouterLink>
                    </div>
                </header>

                <section
                    class="grid gap-10 py-14 lg:grid-cols-[1.08fr_0.92fr] lg:items-center lg:py-20"
                >
                    <div class="max-w-3xl">
                        <div class="flex flex-wrap gap-2">
                            <span
                                class="inline-flex items-center rounded-full border border-amber-200 bg-amber-50 px-3 py-1 text-xs font-semibold tracking-[0.18em] text-amber-800 uppercase"
                            >
                                One paid workspace per account
                            </span>
                            <span
                                class="inline-flex items-center rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1 text-xs font-semibold tracking-[0.18em] text-emerald-800 uppercase"
                            >
                                Google sign-in plus team invites
                            </span>
                        </div>

                        <h1
                            class="mt-6 max-w-4xl text-5xl font-semibold tracking-tighter text-zinc-950 sm:text-6xl"
                        >
                            Invoicing that fits a
                            <span class="block text-zinc-600">
                                small team, not an accounting department.
                            </span>
                        </h1>

                        <p class="mt-6 max-w-2xl text-lg leading-8 text-zinc-600 sm:text-xl">
                            Register with Google, activate one GBP 5 monthly workspace, invite
                            teammates, and keep clients, invoices, revisions, and exports together
                            in one shared account.
                        </p>

                        <div class="mt-8 flex flex-wrap items-center gap-3">
                            <RouterLink
                                to="/signup"
                                class="inline-flex items-center gap-2 rounded-full bg-zinc-950 px-5 py-3 text-sm font-semibold text-white transition hover:bg-zinc-800"
                            >
                                Register with Google
                                <ArrowRightIcon class="size-4" />
                            </RouterLink>
                            <RouterLink
                                to="/login"
                                class="inline-flex items-center rounded-full border border-zinc-300 bg-white/85 px-5 py-3 text-sm font-medium text-zinc-700 transition hover:border-zinc-400 hover:text-zinc-950"
                            >
                                Log in to your workspace
                            </RouterLink>
                        </div>

                        <ul class="mt-10 grid gap-3 text-sm text-zinc-700 sm:grid-cols-3">
                            <li
                                v-for="highlight in highlights"
                                :key="highlight"
                                class="rounded-3xl border border-zinc-300 bg-white/85 p-4 shadow-sm backdrop-blur"
                            >
                                <CheckCircleIcon class="mb-3 size-5 text-emerald-600" />
                                {{ highlight }}
                            </li>
                        </ul>

                        <div class="mt-10 flex flex-wrap items-center gap-3">
                            <span class="text-sm font-medium text-zinc-500">Best fit for:</span>
                            <span
                                v-for="tag in audienceTags"
                                :key="tag"
                                class="rounded-full border border-zinc-300 bg-white/75 px-4 py-2 text-sm text-zinc-700 shadow-sm"
                            >
                                {{ tag }}
                            </span>
                        </div>
                    </div>

                    <aside class="relative">
                        <div
                            class="rounded-4xl border border-zinc-300 bg-white/85 p-5 shadow-2xl backdrop-blur"
                        >
                            <div
                                class="rounded-[1.75rem] border border-zinc-200 bg-zinc-950 p-5 text-white"
                            >
                                <div class="flex items-start justify-between gap-4">
                                    <div>
                                        <div
                                            class="text-xs font-semibold tracking-[0.2em] text-zinc-400 uppercase"
                                        >
                                            Workspace model
                                        </div>
                                        <div class="mt-2 text-2xl font-semibold">
                                            One admin, one subscription, one shared account
                                        </div>
                                    </div>
                                    <div
                                        class="rounded-2xl border border-white/10 bg-white/10 p-3 text-amber-200"
                                    >
                                        <BanknotesIcon class="size-6" />
                                    </div>
                                </div>

                                <div class="mt-6 grid gap-3">
                                    <div class="rounded-2xl border border-white/10 bg-white/5 p-4">
                                        <div class="flex items-center justify-between gap-3">
                                            <span class="text-sm font-semibold">Admin access</span>
                                            <span class="text-xs text-zinc-400">
                                                Google sign-up
                                            </span>
                                        </div>
                                        <p class="mt-2 text-sm leading-6 text-zinc-300">
                                            Creates the workspace and controls billing.
                                        </p>
                                    </div>

                                    <div class="rounded-2xl border border-white/10 bg-white/5 p-4">
                                        <div class="flex items-center justify-between gap-3">
                                            <span class="text-sm font-semibold">
                                                Workspace billing
                                            </span>
                                            <span class="text-xs text-zinc-400">GBP 5 / month</span>
                                        </div>
                                        <p class="mt-2 text-sm leading-6 text-zinc-300">
                                            Unlocks clients, invoice builder, editor, settings, and
                                            exports for the account.
                                        </p>
                                    </div>

                                    <div class="rounded-2xl border border-white/10 bg-white/5 p-4">
                                        <div class="flex items-center justify-between gap-3">
                                            <span class="text-sm font-semibold">Team access</span>
                                            <span class="text-xs text-zinc-400">
                                                Invite teammates
                                            </span>
                                        </div>
                                        <p class="mt-2 text-sm leading-6 text-zinc-300">
                                            Teammates log in with Google and work inside the same
                                            account after the workspace is paid.
                                        </p>
                                    </div>
                                </div>
                            </div>

                            <div class="mt-4 grid gap-3">
                                <article
                                    v-for="card in workspaceCards"
                                    :key="card.title"
                                    :class="[
                                        'rounded-3xl border border-zinc-200 bg-linear-to-br p-4',
                                        card.accent,
                                    ]"
                                >
                                    <div class="flex items-start gap-3">
                                        <div
                                            class="rounded-2xl border border-zinc-200 bg-white p-3 text-zinc-700"
                                        >
                                            <component
                                                :is="card.icon"
                                                class="size-5"
                                            />
                                        </div>
                                        <div>
                                            <h2 class="text-base font-semibold text-zinc-950">
                                                {{ card.title }}
                                            </h2>
                                            <p class="mt-2 text-sm leading-6 text-zinc-600">
                                                {{ card.body }}
                                            </p>
                                        </div>
                                    </div>
                                </article>
                            </div>
                        </div>
                    </aside>
                </section>

                <section class="border-y border-zinc-200 py-16">
                    <div class="max-w-2xl">
                        <p class="text-sm font-semibold tracking-[0.18em] text-amber-800 uppercase">
                            What the app can do
                        </p>
                        <h2
                            class="mt-3 text-3xl font-semibold tracking-tight text-zinc-950 sm:text-4xl"
                        >
                            Built around the real invoice workflow, not filler features
                        </h2>
                        <p class="mt-4 text-base leading-7 text-zinc-600 sm:text-lg">
                            The core app is focused on the work most small teams actually repeat:
                            saving clients, reusing items, writing invoices, revising them safely,
                            and exporting clean documents.
                        </p>
                    </div>

                    <div class="mt-8 grid gap-4 lg:grid-cols-3">
                        <article
                            v-for="card in workspaceCards"
                            :key="`${card.title}-feature`"
                            class="rounded-[1.75rem] border border-zinc-200 bg-white p-6 shadow-sm"
                        >
                            <div
                                class="inline-flex items-center gap-2 rounded-full border border-zinc-200 bg-zinc-50 px-3 py-1 text-xs font-semibold tracking-[0.18em] text-zinc-600 uppercase"
                            >
                                <component
                                    :is="card.icon"
                                    class="size-4"
                                />
                                Feature
                            </div>
                            <h3 class="mt-4 text-xl font-semibold text-zinc-950">
                                {{ card.title }}
                            </h3>
                            <p class="mt-3 text-sm leading-7 text-zinc-600">
                                {{ card.body }}
                            </p>
                        </article>
                    </div>

                    <div class="mt-4 grid gap-4 lg:grid-cols-3">
                        <article
                            class="rounded-[1.75rem] border border-zinc-200 bg-white p-6 shadow-sm"
                        >
                            <div class="flex items-center gap-3">
                                <div class="rounded-2xl bg-amber-50 p-3 text-amber-700">
                                    <FolderOpenIcon class="size-5" />
                                </div>
                                <h3 class="text-lg font-semibold text-zinc-950">
                                    Clients and items
                                </h3>
                            </div>
                            <p class="mt-3 text-sm leading-7 text-zinc-600">
                                Keep regular clients and reusable invoice items ready instead of
                                rebuilding the same data every time.
                            </p>
                        </article>

                        <article
                            class="rounded-[1.75rem] border border-zinc-200 bg-white p-6 shadow-sm"
                        >
                            <div class="flex items-center gap-3">
                                <div class="rounded-2xl bg-emerald-50 p-3 text-emerald-700">
                                    <SparklesIcon class="size-5" />
                                </div>
                                <h3 class="text-lg font-semibold text-zinc-950">Invoice totals</h3>
                            </div>
                            <p class="mt-3 text-sm leading-7 text-zinc-600">
                                Adjust VAT, discounts, deposits, and payment state while keeping the
                                numbers visible as you work.
                            </p>
                        </article>

                        <article
                            class="rounded-[1.75rem] border border-zinc-200 bg-white p-6 shadow-sm"
                        >
                            <div class="flex items-center gap-3">
                                <div class="rounded-2xl bg-sky-50 p-3 text-sky-700">
                                    <BriefcaseIcon class="size-5" />
                                </div>
                                <h3 class="text-lg font-semibold text-zinc-950">
                                    PDF and DOCX export
                                </h3>
                            </div>
                            <p class="mt-3 text-sm leading-7 text-zinc-600">
                                Generate clean export files from the same invoice data instead of
                                copying it into separate templates.
                            </p>
                        </article>
                    </div>
                </section>

                <section class="py-16">
                    <div class="grid gap-6 lg:grid-cols-[0.92fr_1.08fr]">
                        <aside
                            class="rounded-4xl border border-zinc-200 bg-zinc-950 p-6 text-white shadow-sm sm:p-8"
                        >
                            <p
                                class="text-sm font-semibold tracking-[0.18em] text-amber-200 uppercase"
                            >
                                Billing model
                            </p>
                            <h2 class="mt-3 text-3xl font-semibold tracking-tight">
                                Clear access rules from day one
                            </h2>
                            <div class="mt-6 space-y-4">
                                <article
                                    v-for="step in billingSteps"
                                    :key="step.title"
                                    class="rounded-3xl border border-white/10 bg-white/5 p-4"
                                >
                                    <div
                                        class="flex items-center gap-2 text-sm font-semibold text-white"
                                    >
                                        <ShieldCheckIcon class="size-4 text-amber-200" />
                                        {{ step.title }}
                                    </div>
                                    <p class="mt-2 text-sm leading-6 text-zinc-300">
                                        {{ step.body }}
                                    </p>
                                </article>
                            </div>
                        </aside>

                        <div>
                            <div class="max-w-2xl">
                                <p
                                    class="text-sm font-semibold tracking-[0.18em] text-emerald-700 uppercase"
                                >
                                    Feature examples
                                </p>
                                <h2
                                    class="mt-3 text-3xl font-semibold tracking-tight text-zinc-950 sm:text-4xl"
                                >
                                    Concrete ways the workspace helps
                                </h2>
                                <p class="mt-4 text-base leading-7 text-zinc-600 sm:text-lg">
                                    These are the kinds of jobs the app is designed to make calmer:
                                    repeated monthly invoices, shared small-team admin, and
                                    revision-heavy client work.
                                </p>
                            </div>

                            <div class="mt-8 grid gap-4">
                                <article
                                    v-for="example in featureExamples"
                                    :key="example.title"
                                    class="rounded-[1.75rem] border border-zinc-200 bg-white p-6 shadow-sm"
                                >
                                    <div
                                        class="flex flex-col gap-6 lg:flex-row lg:items-start lg:justify-between"
                                    >
                                        <div class="max-w-xl">
                                            <div
                                                class="text-xs font-semibold tracking-[0.2em] text-zinc-500 uppercase"
                                            >
                                                {{ example.eyebrow }}
                                            </div>
                                            <h3 class="mt-3 text-xl font-semibold text-zinc-950">
                                                {{ example.title }}
                                            </h3>
                                            <p class="mt-3 text-sm leading-7 text-zinc-600">
                                                {{ example.summary }}
                                            </p>
                                        </div>

                                        <div class="grid w-full max-w-sm gap-2">
                                            <div
                                                v-for="row in example.rows"
                                                :key="row"
                                                class="rounded-2xl border border-zinc-200 bg-zinc-50 px-4 py-3 text-sm text-zinc-700"
                                            >
                                                {{ row }}
                                            </div>
                                        </div>
                                    </div>
                                </article>
                            </div>
                        </div>
                    </div>
                </section>

                <section class="pb-16">
                    <div
                        class="rounded-[2.25rem] border border-zinc-200 bg-[linear-gradient(135deg,#fff7e4_0%,#ffffff_48%,#eff6f2_100%)] p-6 shadow-sm sm:p-8"
                    >
                        <div class="grid gap-6 lg:grid-cols-[1fr_auto] lg:items-center">
                            <div class="max-w-2xl">
                                <p
                                    class="text-sm font-semibold tracking-[0.18em] text-amber-800 uppercase"
                                >
                                    Ready to try it
                                </p>
                                <h2
                                    class="mt-3 text-3xl font-semibold tracking-tight text-zinc-950 sm:text-4xl"
                                >
                                    Register the workspace and set it up properly
                                </h2>
                                <p class="mt-4 text-base leading-7 text-zinc-600 sm:text-lg">
                                    Start with Google, activate the workspace, then bring teammates
                                    in when you are ready. The whole model is built around one
                                    shared account instead of disconnected personal logins.
                                </p>
                            </div>

                            <div class="flex flex-wrap gap-3">
                                <RouterLink
                                    to="/signup"
                                    class="inline-flex items-center gap-2 rounded-full bg-zinc-950 px-5 py-3 text-sm font-semibold text-white transition hover:bg-zinc-800"
                                >
                                    Register
                                    <ArrowRightIcon class="size-4" />
                                </RouterLink>
                                <RouterLink
                                    to="/login"
                                    class="inline-flex items-center rounded-full border border-zinc-300 bg-white px-5 py-3 text-sm font-medium text-zinc-700 transition hover:border-zinc-400 hover:text-zinc-950"
                                >
                                    Log in
                                </RouterLink>
                            </div>

                            <div class="text-sm text-zinc-500">
                                By using the app, you agree to the necessary sign-in cookie
                                described on the
                                <RouterLink
                                    to="/privacy"
                                    class="ml-1 font-semibold text-zinc-700 underline decoration-zinc-400 underline-offset-2"
                                >
                                    Privacy and Cookies page
                                </RouterLink>
                                .
                            </div>
                        </div>
                    </div>
                </section>
            </section>
        </div>
    </main>
</template>

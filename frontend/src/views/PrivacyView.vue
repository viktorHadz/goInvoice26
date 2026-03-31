<template>
    <main
        class="min-h-screen bg-[linear-gradient(180deg,#f7f3ea_0%,#fffdfa_38%,#eef5f2_100%)] text-zinc-900"
    >
        <section class="mx-auto w-full max-w-4xl px-5 py-10 sm:px-8 sm:py-14">
            <header class="max-w-2xl">
                <p class="text-sm font-semibold tracking-[0.18em] text-zinc-500 uppercase">
                    Privacy and cookies
                </p>
                <h1 class="mt-3 text-4xl font-semibold tracking-tight text-zinc-950 sm:text-5xl">
                    How Invoicer handles sign-in, billing, and essential cookies
                </h1>
                <p class="mt-5 text-base leading-8 text-zinc-600 sm:text-lg">
                    This page explains the basics of what we store to run the app, keep you signed
                    in, and process billing. It is written as a plain-language product summary, not
                    legal advice.
                </p>
            </header>

            <div class="mt-10 grid gap-4">
                <section class="rounded-3xl border border-zinc-200 bg-white/85 p-6 shadow-sm">
                    <h2 class="text-xl font-semibold text-zinc-950">Necessary sign-in cookie</h2>
                    <p class="mt-3 text-sm leading-7 text-zinc-600">
                        We use a secure session cookie after Google sign-in so the app can recognise
                        you on later requests. This cookie is necessary for authentication and
                        workspace access.
                    </p>
                    <p class="mt-3 text-sm leading-7 text-zinc-600">
                        If you block or delete this cookie, you may be signed out and some parts of
                        the app will stop working until you sign in again.
                    </p>
                </section>

                <section class="rounded-3xl border border-zinc-200 bg-white/85 p-6 shadow-sm">
                    <h2 class="text-xl font-semibold text-zinc-950">Google sign-in data</h2>
                    <p class="mt-3 text-sm leading-7 text-zinc-600">
                        When you sign in with Google, we store the basics needed to run your
                        account, such as your name, email address, avatar, Google account
                        identifier, role in the workspace, and session records.
                    </p>
                </section>

                <section class="rounded-3xl border border-zinc-200 bg-white/85 p-6 shadow-sm">
                    <h2 class="text-xl font-semibold text-zinc-950">Workspace data</h2>
                    <p class="mt-3 text-sm leading-7 text-zinc-600">
                        Inside the app, we store the information you create to use the product:
                        company settings, logo, clients, saved items, invoices, revisions, and team
                        invites.
                    </p>
                </section>

                <section class="rounded-3xl border border-zinc-200 bg-white/85 p-6 shadow-sm">
                    <h2 class="text-xl font-semibold text-zinc-950">Billing and Stripe</h2>
                    <p class="mt-3 text-sm leading-7 text-zinc-600">
                        Billing is handled through Stripe. We store the minimum identifiers needed
                        to link your workspace to Stripe, such as customer ID, subscription ID,
                        billing email, subscription status, and billing period details.
                    </p>
                    <p class="mt-3 text-sm leading-7 text-zinc-600">
                        Card payment details are handled by Stripe and are not stored directly in
                        this app.
                    </p>
                </section>

                <section class="rounded-3xl border border-zinc-200 bg-white/85 p-6 shadow-sm">
                    <h2 class="text-xl font-semibold text-zinc-950">Questions or support</h2>
                    <p class="mt-3 text-sm leading-7 text-zinc-600">
                        If you need a full policy, data export help, or support around billing and
                        access, please get in touch through your normal support contact.
                    </p>
                </section>
            </div>
        </section>
    </main>
</template>

<script setup lang="ts">
import { ClockIcon, FolderOpenIcon, PencilSquareIcon } from '@heroicons/vue/24/outline'

const leadPillar = {
  title: 'Quick to learn',
  subtitle: 'Set up',
  body: 'Built so someone can get up and running quickly without feeling like they need a degree just to produce an invoice.',
  note: 'Less time learning the tool, more time getting paid.',
  icon: ClockIcon,
}

const trustBadges = ['Clear on day one', 'Proper records later', 'Built for busy small businesses']

const sidePillars = [
  {
    title: 'Safe invoice edits',
    subtitle: 'Revise',
    body: 'Revision tracking keeps records proper instead of encouraging messy overwrites.',
    icon: PencilSquareIcon,
  },
  {
    title: 'Easy records',
    subtitle: 'Keep',
    body: 'Clients, invoices, and past work stay easy to revisit when someone asks for an old invoice.',
    icon: FolderOpenIcon,
  },
]
</script>

<template>
  <section class="px-5 py-10 sm:px-8 sm:py-14 lg:px-10 lg:py-18">
    <div class="mx-auto max-w-7xl">
      <div
        class="card-grad relative overflow-hidden rounded-4xl border border-sky-200/70 shadow-[0_28px_80px_-42px_rgba(14,116,144,0.42)] ring-1 ring-white/70 dark:border-zinc-800 dark:shadow-[0_28px_90px_-48px_rgba(0,0,0,0.8)] dark:ring-white/5"
      >
        <div
          class="pointer-events-none absolute -top-16 right-0 h-56 w-56 rounded-full bg-sky-200/55 blur-3xl dark:bg-emerald-500/10"
        />
        <div
          class="pointer-events-none absolute -bottom-20 left-12 h-48 w-48 rounded-full bg-white/70 blur-3xl dark:bg-zinc-900/50"
        />

        <div
          class="relative grid gap-8 px-6 py-8 sm:gap-10 sm:px-8 sm:py-10 lg:grid-cols-[0.92fr_1.08fr] lg:gap-12 lg:px-10 lg:py-12"
        >
          <div class="max-w-xl">
            <p
              class="text-xs font-semibold tracking-[0.18em] text-sky-700 uppercase sm:text-sm dark:text-emerald-300"
            >
              Why trust it
            </p>

            <h2
              class="mt-4 text-3xl font-semibold tracking-tight text-zinc-950 sm:text-4xl lg:text-5xl dark:text-white"
            >
              Simple while you're busy.
              <span class="block text-sky-700 dark:text-emerald-300">
                Organised when it matters.
              </span>
            </h2>

            <p
              class="mt-5 max-w-2xl text-base leading-8 text-zinc-600 sm:text-lg dark:text-zinc-300"
            >
              Invoice-And-Go keeps the day-to-day quick without turning your records into a mess
              later on.
            </p>

            <div class="mt-8 flex flex-wrap gap-3">
              <span
                v-for="badge in trustBadges"
                :key="badge"
                class="inline-flex items-center rounded-full border border-sky-200/80 bg-white/80 px-3.5 py-1.5 text-sm font-medium text-zinc-700 shadow-sm backdrop-blur dark:border-emerald-400/15 dark:bg-white/5 dark:text-zinc-200"
              >
                {{ badge }}
              </span>
            </div>
          </div>

          <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-[1.08fr_0.92fr]">
            <article
              class="card-grad relative flex h-full flex-col overflow-hidden rounded-[1.75rem] border border-sky-200/80 p-6 shadow-sm sm:p-7 lg:row-span-2 dark:border-zinc-800"
            >
              <div
                class="absolute inset-x-0 top-0 h-24 bg-linear-to-b from-sky-100/70 to-transparent dark:from-emerald-500/10"
              />

              <div class="relative flex items-center gap-3">
                <div
                  class="flex size-11 shrink-0 items-center justify-center rounded-2xl border border-sky-200 bg-white text-sky-700 shadow-sm dark:border-emerald-400/20 dark:bg-emerald-500/10 dark:text-emerald-200"
                >
                  <component
                    :is="leadPillar.icon"
                    class="size-5"
                  />
                </div>

                <p
                  class="text-xs font-semibold tracking-[0.18em] text-sky-700 uppercase dark:text-emerald-300"
                >
                  {{ leadPillar.subtitle }}
                </p>
              </div>

              <div class="relative mt-8 max-w-sm">
                <h3
                  class="text-2xl font-semibold tracking-tight text-balance text-zinc-950 dark:text-white"
                >
                  {{ leadPillar.title }}
                </h3>

                <p class="mt-4 text-base leading-8 text-zinc-600 dark:text-zinc-300">
                  {{ leadPillar.body }}
                </p>
              </div>

              <div class="relative mt-auto pt-8">
                <div class="h-px w-full bg-zinc-200/80 dark:bg-zinc-800" />
                <p class="mt-4 text-sm font-medium text-zinc-500 dark:text-zinc-400">
                  {{ leadPillar.note }}
                </p>
              </div>
            </article>

            <article
              v-for="pillar in sidePillars"
              :key="pillar.title"
              class="group flex h-full flex-col rounded-3xl border border-zinc-200/80 bg-white/80 p-5 shadow-sm backdrop-blur transition duration-200 hover:-translate-y-0.5 hover:border-sky-300 hover:shadow-md dark:border-zinc-800 dark:bg-zinc-900/80 dark:hover:border-emerald-400/25"
            >
              <div class="flex items-center gap-3">
                <div
                  class="flex size-10 shrink-0 items-center justify-center rounded-xl border border-sky-200 bg-sky-50 text-sky-700 transition duration-200 group-hover:bg-sky-100 dark:border-emerald-400/20 dark:bg-emerald-500/10 dark:text-emerald-200 dark:group-hover:bg-emerald-500/15"
                >
                  <component
                    :is="pillar.icon"
                    class="size-5"
                  />
                </div>

                <p
                  class="text-xs font-semibold tracking-[0.18em] text-sky-700 uppercase dark:text-emerald-300"
                >
                  {{ pillar.subtitle }}
                </p>
              </div>

              <div class="mt-6 max-w-xs">
                <h3
                  class="text-xl font-semibold tracking-tight text-balance text-zinc-950 dark:text-white"
                >
                  {{ pillar.title }}
                </h3>

                <p class="mt-3 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                  {{ pillar.body }}
                </p>
              </div>
            </article>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

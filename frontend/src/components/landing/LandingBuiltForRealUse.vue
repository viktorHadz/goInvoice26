<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  trialLabel: string
  teamSeatLimit: number
}>()

const operatingNotes = computed(() => [
  {
    label: 'Start with',
    value: props.trialLabel,
    body: 'Add business details, clients, and saved items as you go, then invoice with a clear structure from day one.',
  },
  {
    label: 'Keep continuity',
    value: 'Solo first, team ready',
    body: 'The same workflow works for one person today and shared work later, so your process does not need to be rebuilt.',
  },
  {
    label: 'Scale with control',
    value: `Up to ${props.teamSeatLimit} people`,
    body: 'Invite teammates into one shared workspace where clients, invoices, and revision history stay aligned.',
  },
])

const pillars = computed(() => [
  {
    eyebrow: 'For the person doing the work',
    title: 'Faster admin without cutting corners',
    body: 'Create and revise invoices quickly while keeping records clear enough to trust when questions come back later.',
  },
  {
    eyebrow: 'For the business',
    title: 'Records that hold up as you grow',
    body: 'Consistent invoice history, clear totals, and structured revisions help your business stay professional under real workload.',
  },
  {
    eyebrow: 'For when work stops being solo',
    title: 'Team access without process chaos',
    body: `When collaboration starts, invite up to ${props.teamSeatLimit} people into the same workspace with shared clients, invoices, and settings.`,
  },
])

const transitionSteps = computed(() => [
  {
    title: 'Set foundations',
    body: 'Build business details, clients, and reusable items over time so every invoice starts from a reliable baseline.',
  },
  {
    title: 'Invite when needed',
    body: 'Add teammates when collaboration begins, not before, and keep everyone working in the same system.',
  },
  {
    title: 'Stay in sync',
    body: 'Clients, invoices, and revision history remain shared, so the whole team works from the same source of truth.',
  },
])
</script>

<template>
  <section class="px-5 py-10 sm:px-8 sm:py-14 lg:px-10 lg:py-18">
    <div class="mx-auto max-w-7xl">
      <div class="relative overflow-hidden border-y border-zinc-200/80 py-10 sm:py-12 lg:py-14 dark:border-zinc-800">
        <div
          class="pointer-events-none absolute -top-10 right-10 h-36 w-36 rounded-full bg-sky-200/35 blur-3xl dark:bg-emerald-500/10" />
        <div
          class="pointer-events-none absolute bottom-0 left-1/3 h-28 w-28 rounded-full bg-white/80 blur-3xl dark:bg-zinc-900/60" />

        <div class="relative grid gap-12 lg:grid-cols-[0.84fr_1.16fr] lg:gap-16">
          <div class="max-w-xl">
            <p class="text-xs font-semibold tracking-[0.18em] text-sky-700 uppercase sm:text-sm dark:text-emerald-300">
              Built for real use
            </p>

            <h2 class="mt-3 text-3xl font-semibold tracking-tight text-zinc-950 sm:text-4xl dark:text-white">
              Enough structure to run confidently.
              <span class="block text-sky-700 dark:text-emerald-300">
                Ready when invoicing becomes shared.
              </span>
            </h2>

            <p class="mt-5 max-w-2xl text-base leading-8 text-zinc-600 sm:text-lg dark:text-zinc-300">
              Start with a straightforward solo workflow, then bring teammates into the same workspace
              when needed without changing how work is tracked.
            </p>

            <dl class="mt-8 space-y-5">
              <div v-for="note in operatingNotes" :key="note.label"
                class="border-b border-zinc-200/80 pb-5 last:border-b-0 last:pb-0 dark:border-zinc-800">
                <dt class="text-[11px] font-semibold tracking-[0.18em] text-zinc-500 uppercase dark:text-zinc-400">
                  {{ note.label }}
                </dt>
                <dd class="mt-2 text-2xl font-semibold tracking-tight text-zinc-950 dark:text-white">
                  {{ note.value }}
                </dd>
                <dd class="mt-2 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                  {{ note.body }}
                </dd>
              </div>
            </dl>
          </div>

          <div class="lg:pt-2">
            <div class="relative pl-6.5 sm:pl-8">
              <div
                class="absolute top-2 bottom-2 left-[0.45rem] sm:left-[0.1rem] w-px bg-linear-to-b from-sky-300 via-sky-200 to-zinc-300 dark:from-emerald-400/60 dark:via-zinc-700 dark:to-emerald-500/20 " />

              <article v-for="pillar in pillars" :key="pillar.title" class="relative pb-8 last:pb-0">
                <span
                  class="absolute top-0.5 -left-6 sm:-left-[2.3rem] size-3 rounded-full bg-sky-500 ring-4 ring-white dark:bg-emerald-400 dark:ring-[#091312] " />

                <p class="text-[11px] font-semibold tracking-[0.18em] text-zinc-500 uppercase dark:text-zinc-400">
                  {{ pillar.eyebrow }}
                </p>
                <h3 class="mt-2 text-xl font-semibold tracking-tight text-zinc-950 sm:text-2xl dark:text-white">
                  {{ pillar.title }}
                </h3>
                <p class="mt-3 max-w-2xl text-sm leading-7 text-zinc-600 sm:text-base dark:text-zinc-300">
                  {{ pillar.body }}
                </p>
              </article>
            </div>

            <div class="mt-10 border-t border-zinc-200/80 pt-8 dark:border-zinc-800">
              <div class="relative grid gap-8 md:grid-cols-3 md:gap-10">
                <div
                  class="pointer-events-none absolute top-1.5 right-0 left-0 hidden h-px bg-zinc-200 md:block dark:bg-zinc-800" />

                <article v-for="(step, index) in transitionSteps" :key="step.title" class="relative pt-6">
                  <span
                    class="absolute -top-0.5 left-0 inline-flex text-[11px] font-semibold tracking-[0.18em] text-sky-700 uppercase dark:text-emerald-300">
                    0{{ index + 1 }}
                  </span>

                  <h3 class="text-lg font-semibold tracking-tight text-zinc-950 dark:text-white">
                    {{ step.title }}
                  </h3>
                  <p class="mt-2 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                    {{ step.body }}
                  </p>
                </article>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

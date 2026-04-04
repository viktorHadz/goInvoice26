<script setup lang="ts">
import { computed, ref, type Component } from 'vue'
import {
  BookOpenIcon,
  BriefcaseIcon,
  Cog6ToothIcon,
  DocumentTextIcon,
  FolderOpenIcon,
  PencilSquareIcon,
  QueueListIcon,
  UserGroupIcon,
} from '@heroicons/vue/24/outline'

type FeatureId = 'clients' | 'items' | 'invoice' | 'invoiceBook' | 'editor' | 'settings'

type UsageFeature = {
  id: FeatureId
  menuLabel: string
  title: string
  body: string
  previewTitle: string
  previewBody: string
  badge: string
  metrics?: Array<{ label: string; value: string }>
  rows: Array<{ title: string; caption: string; value: string }>
  icon: Component
  previewSurfaceClass: string
}

const usageFeatures: UsageFeature[] = [
  {
    id: 'clients',
    menuLabel: 'Set up',
    title: '1. Set up',
    body: 'Settings, clients and items.',
    previewTitle: 'Everything is ready before you invoice',
    previewBody: 'Save the basics once so repeat work is quicker and more consistent',
    badge: 'Initial',
    metrics: [
      { label: 'Details', value: 'Saved' },
      { label: 'Clients', value: 'Ready' },
      { label: 'Items', value: 'Priced' },
    ],
    rows: [
      {
        title: 'Business details',
        caption: 'Logo contact info and others',
        value: 'Ready',
      },
      {
        title: 'Client records',
        caption: 'Company name, address and email info stored',
        value: 'Saved',
      },
      {
        title: 'Saved items',
        caption: 'Create items and set prices for reuse',
        value: 'Created',
      },
    ],
    icon: Cog6ToothIcon,
    previewSurfaceClass:
      'bg-linear-to-br from-amber-100 via-white to-yellow-50 dark:from-emerald-500/10 dark:via-zinc-950 dark:to-zinc-950',
  },

  {
    id: 'invoice',
    menuLabel: 'Invoice',
    title: '2. Invoice',
    body: 'Create and total in one place.',
    previewTitle: 'Build invoices without slowing down',
    previewBody: 'Pick a client, add saved items and create clean totals in one flow',
    badge: 'Create',
    metrics: [
      { label: 'Client', value: 'Picked' },
      { label: 'Items', value: 'Added' },
      { label: 'Totals', value: 'Clear' },
    ],
    rows: [
      {
        title: 'Client selection',
        caption: 'Choose a saved client instead of starting from scratch',
        value: 'Selected',
      },
      {
        title: 'Invoice lines',
        caption: 'Add saved items or enter one-off work when needed',
        value: 'Added',
      },
      {
        title: 'Invoice totals',
        caption: 'See subtotals, discounts and final amount together',
        value: 'Ready',
      },
    ],
    icon: DocumentTextIcon,
    previewSurfaceClass:
      'bg-linear-to-br from-sky-200/80 via-white to-cyan-50/80 dark:from-emerald-500/10 dark:via-zinc-950 dark:to-zinc-950',
  },

  {
    id: 'editor',
    menuLabel: 'Editor',
    title: '3. Revise',
    body: 'Make changes without messy overwrites.',
    previewTitle: 'Edit safely when something needs changing',
    previewBody: 'Update invoices properly while keeping earlier versions intact',
    badge: 'Tracked',
    metrics: [
      { label: 'Original', value: 'Kept' },
      { label: 'Changes', value: 'Tracked' },
      { label: 'History', value: 'Clear' },
    ],
    rows: [
      {
        title: 'Revision copy',
        caption: 'Changes are made on a new version instead of replacing the old one',
        value: 'Created',
      },
      {
        title: 'Updated invoice',
        caption: 'Adjust lines dates or details while keeping records proper',
        value: 'Edited',
      },
      {
        title: 'Revision trail',
        caption: 'See what changed and which version is the latest',
        value: 'Visible',
      },
    ],
    icon: PencilSquareIcon,
    previewSurfaceClass:
      'bg-linear-to-br from-violet-100 via-white to-sky-50 dark:from-emerald-500/10 dark:via-zinc-950 dark:to-zinc-950',
  },

  {
    id: 'invoiceBook',
    menuLabel: 'Invoice book',
    title: '4. Find later',
    body: 'Past work stays organised.',
    previewTitle: 'Everything stays easy to find afterwards',
    previewBody: 'Look back through invoices, statuses and revisions without digging around',
    badge: 'History',
    metrics: [
      { label: 'Invoices', value: 'Listed' },
      { label: 'Status', value: 'Visible' },
      { label: 'Revisions', value: 'Linked' },
    ],
    rows: [
      {
        title: 'Invoice list',
        caption: 'See saved invoices together in one organised place',
        value: 'Listed',
      },
      {
        title: 'Status view',
        caption: 'Check whether something is draft issued paid or void',
        value: 'Visible',
      },
      {
        title: 'Revision history',
        caption: 'Open earlier and latest versions from the same invoice record',
        value: 'Linked',
      },
    ],
    icon: BookOpenIcon,
    previewSurfaceClass:
      'bg-linear-to-br from-teal-100 via-white to-sky-50 dark:from-emerald-500/10 dark:via-zinc-950 dark:to-zinc-950',
  },
]
const defaultFeature = usageFeatures[0]!
const activeFeatureId = ref<FeatureId>('clients')

const activeFeature = computed<UsageFeature>(
  () => usageFeatures.find((feature) => feature.id === activeFeatureId.value) ?? defaultFeature,
)
</script>

<template>
  <section class="z-10 px-5 py-10 sm:px-8 sm:py-14 lg:px-10 lg:py-18">
    <div class="mx-auto max-w-7xl">
      <div class="grid gap-10 lg:grid-cols-[1.18fr_0.82fr] lg:items-center lg:gap-12">
        <div class="relative order-2 lg:order-1">
          <div
            class="overflow-hidden rounded-4xl border border-zinc-200/80 bg-white/95 shadow-[0_28px_80px_-42px_rgba(15,23,42,0.35)] ring-1 ring-white/80 dark:border-zinc-800 dark:bg-zinc-950/92 dark:shadow-[0_30px_100px_-50px_rgba(0,0,0,0.9)] dark:ring-white/5"
          >
            <div
              class="flex items-center gap-3 border-b border-zinc-200/80 bg-zinc-50/85 px-4 py-3 dark:border-zinc-800 dark:bg-zinc-900/80"
            >
              <div class="flex shrink-0 items-center gap-1.5">
                <span class="size-2.5 rounded-full bg-rose-400" />
                <span class="size-2.5 rounded-full bg-amber-400" />
                <span class="size-2.5 rounded-full bg-emerald-400" />
              </div>

              <div
                class="min-w-0 flex-1 truncate rounded-full bg-white px-4 py-2 text-[11px] font-medium tracking-[0.12em] text-zinc-500 uppercase shadow-sm dark:bg-zinc-950 dark:text-zinc-400"
              >
                invoiceandgo.app
              </div>
            </div>

            <div
              class="border-b border-zinc-200/80 bg-zinc-50/70 p-4 dark:border-zinc-800 dark:bg-zinc-900/70"
            >
              <div class="flex flex-wrap gap-2">
                <button
                  v-for="feature in usageFeatures"
                  :key="feature.id"
                  type="button"
                  class="inline-flex items-center gap-2 rounded-full px-3.5 py-2 text-sm font-medium transition"
                  :class="
                    feature.id === activeFeature.id
                      ? 'bg-white text-sky-700 shadow-sm dark:bg-zinc-950 dark:text-emerald-200'
                      : 'bg-white/55 text-zinc-500 hover:bg-white hover:text-zinc-900 dark:bg-zinc-950/40 dark:text-zinc-400 dark:hover:bg-zinc-950 dark:hover:text-zinc-200'
                  "
                  @click="activeFeatureId = feature.id"
                >
                  <component
                    :is="feature.icon"
                    class="size-4 shrink-0"
                  />
                  <span>{{ feature.menuLabel }}</span>
                </button>
              </div>
            </div>

            <Transition
              name="usage-preview"
              mode="out-in"
            >
              <div
                :key="activeFeature.id"
                class="grid min-h-152 gap-4 p-4 sm:p-5 lg:p-6"
              >
                <div class="flex flex-col">
                  <div class="flex min-h-29 items-start justify-between gap-4">
                    <div class="max-w-xl">
                      <h3
                        class="text-2xl font-semibold tracking-tight text-zinc-950 dark:text-white"
                      >
                        {{ activeFeature.previewTitle }}
                      </h3>
                      <p class="mt-3 text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                        {{ activeFeature.previewBody }}
                      </p>
                    </div>

                    <span
                      class="inline-flex shrink-0 items-center rounded-full border border-sky-200 bg-white px-3 py-1 text-[11px] font-semibold tracking-[0.16em] text-sky-700 uppercase dark:border-emerald-400/20 dark:bg-zinc-950 dark:text-emerald-300"
                      v-if="activeFeature.badge"
                    >
                      {{ activeFeature.badge }}
                    </span>
                  </div>

                  <div
                    :class="[
                      'mt-4 flex flex-1 flex-col rounded-[1.6rem] p-5 sm:p-6',
                      activeFeature.previewSurfaceClass,
                    ]"
                  >
                    <div class="flex items-center gap-3">
                      <div
                        class="flex size-11 shrink-0 items-center justify-center rounded-2xl bg-white/90 text-zinc-700 shadow-sm dark:bg-zinc-950/85 dark:text-zinc-100"
                      >
                        <component
                          :is="activeFeature.icon"
                          class="size-5"
                        />
                      </div>

                      <p class="text-sm font-semibold text-zinc-950 dark:text-white">
                        {{ activeFeature.menuLabel }}
                      </p>
                    </div>

                    <div class="mt-5 space-y-3">
                      <div
                        v-for="row in activeFeature.rows"
                        :key="row.title"
                        class="flex min-h-21 items-start justify-between gap-4 rounded-[1.15rem] bg-white/85 px-3.5 py-3 shadow-sm dark:bg-zinc-950/80"
                      >
                        <div>
                          <p class="text-sm font-semibold text-zinc-950 dark:text-white">
                            {{ row.title }}
                          </p>
                          <p class="mt-1 text-xs leading-6 text-zinc-500 dark:text-zinc-400">
                            {{ row.caption }}
                          </p>
                        </div>

                        <span
                          class="shrink-0 rounded-full bg-zinc-100 px-2.5 py-1 text-xs font-semibold text-zinc-700 dark:bg-zinc-900 dark:text-zinc-200"
                        >
                          {{ row.value }}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </Transition>
          </div>
        </div>

        <div class="order-1 lg:order-2 lg:pl-2">
          <p
            class="text-xs font-semibold tracking-[0.18em] text-sky-700 uppercase sm:text-sm dark:text-emerald-300"
          >
            See the workflow
          </p>

          <h2
            class="mt-3 text-3xl font-semibold tracking-tight text-zinc-950 sm:text-4xl dark:text-white"
          >
            Explore features
          </h2>

          <p class="mt-4 max-w-xl text-base leading-8 text-zinc-600 sm:text-lg dark:text-zinc-300">
            Each tool exists to cut admin time and speed up the invoicing process to create a
            seamless experience.
          </p>

          <div class="mt-8 space-y-1">
            <button
              v-for="feature in usageFeatures"
              :key="feature.id"
              type="button"
              :aria-pressed="feature.id === activeFeature.id"
              class="group relative block w-full cursor-pointer py-4 pl-6 text-left transition"
              @click="activeFeatureId = feature.id"
            >
              <span
                :class="[
                  'absolute top-3 bottom-3 left-0 w-1 rounded-full transition-all duration-200',
                  feature.id === activeFeature.id
                    ? 'bg-sky-600 opacity-100 dark:bg-emerald-400'
                    : 'bg-sky-500/70 opacity-0 group-hover:opacity-100 dark:bg-emerald-400/80',
                ]"
              />

              <h3
                :class="[
                  'text-xl font-semibold tracking-tight transition-colors sm:text-2xl',
                  feature.id === activeFeature.id
                    ? 'text-zinc-950 dark:text-white'
                    : 'text-zinc-700 group-hover:text-zinc-950 dark:text-zinc-200 dark:group-hover:text-white',
                ]"
              >
                {{ feature.title }}
              </h3>

              <p class="mt-2 max-w-xl text-sm leading-7 text-zinc-600 dark:text-zinc-300">
                {{ feature.body }}
              </p>
            </button>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.usage-preview-enter-active,
.usage-preview-leave-active {
  transition:
    opacity 0.24s ease,
    transform 0.24s ease;
}

.usage-preview-enter-from,
.usage-preview-leave-to {
  opacity: 0;
  transform: translateY(10px);
}
</style>

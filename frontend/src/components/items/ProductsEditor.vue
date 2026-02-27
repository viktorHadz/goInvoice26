<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useProductStore } from '@/stores/products'
import { useClientStore } from '@/stores/clients'
import type { Product, ProductType, PricingMode, ProductUpsert } from '@/utils/productHttpHandler'
import {
  BriefcaseIcon,
  XMarkIcon,
  TrashIcon,
  UserIcon,
  ChevronDownIcon,
  ShieldCheckIcon,
  ArrowPathIcon,
  SquaresPlusIcon,
} from '@heroicons/vue/24/outline'
import TheDropdown from '../UI/TheDropdown.vue'
import TheButton from '../UI/TheButton.vue'
import TheInput from '../UI/TheInput.vue'
import { onKeyStroke } from '@vueuse/core'

const store = useProductStore()
const clientStore = useClientStore()

const open = ref(false)
const tab = ref<ProductType>('style')
const q = ref('')
const selectedId = ref<number | null>(null)

type Form = {
  id: number | null
  productName: string
  pricingMode: PricingMode
  flatPrice: number | null
  hourlyRate: number | null
  minutesWorked: number | null
}

const form = reactive<Form>({
  id: null,
  productName: '',
  pricingMode: 'flat',
  flatPrice: null,
  hourlyRate: null,
  minutesWorked: null,
})

const list = computed(() => store.byType[tab.value] ?? [])
const filtered = computed(() => {
  const s = q.value.trim().toLowerCase()
  if (!s) return list.value
  return list.value.filter((p) => p.productName.toLowerCase().includes(s))
})

watch(open, (v) => v && store.reload())
watch(tab, () => reset())

function reset() {
  selectedId.value = null
  q.value = ''
  form.id = null
  form.productName = ''
  form.pricingMode = 'flat'
  form.flatPrice = null
  form.hourlyRate = null
  form.minutesWorked = null
}

function pick(p: Product) {
  selectedId.value = p.id
  form.id = p.id
  form.productName = p.productName
  form.pricingMode = p.pricingMode
  form.flatPrice = p.flatPriceMinor != null ? p.flatPriceMinor / 100 : null
  form.hourlyRate = p.hourlyRateMinor != null ? p.hourlyRateMinor / 100 : null
  form.minutesWorked = p.minutesWorked ?? null
}

function buildUpsert(): ProductUpsert {
  const productType = tab.value
  const pricingMode: PricingMode = productType === 'style' ? 'flat' : form.pricingMode

  const base = {
    productType,
    pricingMode,
    productName: form.productName.trim(),
  } satisfies ProductUpsert

  if (pricingMode === 'flat') {
    return { ...base, flatPrice: form.flatPrice ?? 0 }
  }
  return {
    ...base,
    hourlyRate: form.hourlyRate ?? 0,
    minutesWorked: form.minutesWorked ?? 0,
  }
}

async function save() {
  const payload = buildUpsert()
  if (form.id == null) {
    await store.create(payload)
  } else {
    const updated = await store.update(form.id, payload)
    pick(updated)
  }
}

async function del() {
  if (!form.id) return
  await store.remove(form.id)
  reset()
}

function money(minor?: number) {
  if (minor == null) return '—'
  return new Intl.NumberFormat('en-GB', { style: 'currency', currency: 'GBP' }).format(minor / 100)
}

// Keys
// Close modal
const escArr = ['esc', 'Escape', 'escape']
function escKeyHandler(e: KeyboardEvent) {
  if (!open.value) {
    console.warn('Tried  to fire outside Products Editor')
    return
  }
  open.value = false
  console.log(e)
}
onKeyStroke(escArr, escKeyHandler, { dedupe: true })
</script>

<template>
  <div
    class="flex flex-col items-center"
    title="products"
  >
    <BriefcaseIcon
      class="size-8 cursor-pointer stroke-1 text-zinc-600 hover:text-sky-600 dark:text-zinc-300 dark:hover:text-emerald-400"
      @click="open = true"
    />
  </div>

  <Teleport to="body">
    <transition name="fade">
      <div
        v-if="open"
        class="fixed inset-0 z-100 bg-black/45 backdrop-blur-[1px]"
        @click="open = false"
      />
    </transition>

    <!-- Top most -->
    <transition name="slide">
      <aside
        v-if="open"
        class="fixed top-0 right-0 z-101 h-screen w-[92vw] max-w-225 border-l border-zinc-200 bg-white text-zinc-900 dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-100"
      >
        <!-- header -->
        <header
          class="border-b border-zinc-200 bg-zinc-50 dark:border-zinc-800 dark:bg-zinc-900/70"
        >
          <div class="relative overflow-hidden px-4 py-3">
            <div
              class="pointer-events-none absolute inset-0 bg-[radial-gradient(900px_circle_at_15%_0%,rgba(56,189,248,0.10),transparent_55%)] opacity-100 dark:bg-[radial-gradient(900px_circle_at_15%_0%,rgba(16,185,129,0.18),transparent_55%)]"
            />
            <div
              class="pointer-events-none absolute inset-0 bg-[linear-gradient(to_right,rgba(255,255,255,0.06)_1px,transparent_1px),linear-gradient(to_bottom,rgba(255,255,255,0.06)_1px,transparent_1px)] bg-size-[36px_36px] opacity-[0.55] dark:bg-[linear-gradient(to_right,rgba(255,255,255,0.04)_1px,transparent_1px),linear-gradient(to_bottom,rgba(255,255,255,0.04)_1px,transparent_1px)]"
            />

            <!-- CONTENT -->
            <div class="relative z-10 flex items-center justify-between gap-4">
              <div class="flex min-w-0 items-center gap-3">
                <!-- icon tile -->
                <div
                  class="grid size-12 shrink-0 place-items-center rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/50"
                >
                  <BriefcaseIcon class="stroke-1.5 size-7 text-sky-700 dark:text-emerald-400" />
                </div>

                <!-- title -->
                <div class="min-w-0">
                  <div class="flex items-center gap-2">
                    <h2
                      class="text-2xl font-semibold tracking-tight text-zinc-900 dark:text-zinc-200"
                    >
                      Products Editor
                    </h2>

                    <span
                      class="rounded-full border border-zinc-200 bg-white/90 px-2 py-0.5 text-xs font-medium text-zinc-600 backdrop-blur-sm dark:border-zinc-700 dark:bg-zinc-900/70 dark:text-zinc-400"
                    >
                      {{ tab === 'style' ? 'Styles' : 'Samples' }}
                    </span>
                  </div>

                  <div class="text-sm tracking-tight text-zinc-500 dark:text-zinc-400">
                    Manage client services, pricing and work units
                  </div>
                </div>
              </div>

              <!-- Close -->
              <button
                class="shrink-0 cursor-pointer rounded-lg p-2 text-zinc-600 hover:bg-zinc-200 hover:text-zinc-900 dark:text-zinc-300 dark:hover:bg-rose-400/15 dark:hover:text-rose-300"
                @click="open = false"
                title="close"
              >
                <XMarkIcon class="size-5" />
              </button>
            </div>
          </div>

          <div class="border-t border-zinc-200/70 dark:border-zinc-800/70"></div>

          <div class="flex items-center gap-3 px-3 py-3">
            <!-- Client Select -->
            <div class="min-w-0 flex-1 pr-6">
              <TheDropdown
                v-model="clientStore.selectedClient"
                :options="clientStore.clients"
                placeholder="Select Client"
                :left-icon="UserIcon"
                :right-icon="ChevronDownIcon"
                label-key="name"
                value-key="id"
              />
            </div>

            <!-- Tabs -->
            <div
              class="flex shrink-0 rounded-full border border-zinc-200 bg-white p-1 shadow-sm dark:border-zinc-800 dark:bg-zinc-900"
            >
              <button
                class="rounded-full px-3 py-1.5 text-sm font-medium transition"
                :class="
                  tab === 'style'
                    ? 'bg-sky-600 text-white shadow-sm dark:bg-emerald-600'
                    : 'text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100'
                "
                @click="tab = 'style'"
              >
                Styles
              </button>

              <button
                class="rounded-full px-3 py-1.5 text-sm font-medium transition"
                :class="
                  tab === 'sample'
                    ? 'bg-sky-600 text-white shadow-sm dark:bg-emerald-600'
                    : 'text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100'
                "
                @click="tab = 'sample'"
              >
                Samples
              </button>
            </div>

            <!-- Search -->
            <div class="hidden w-72 shrink-0 px-3 sm:block">
              <input
                v-model="q"
                class="input input-accent"
                id="product-search"
                :placeholder="`Search ${tab}s…`"
              />
            </div>
          </div>
        </header>

        <div class="grid h-[calc(100%-56px)] grid-cols-1 md:grid-cols-2">
          <!-- list -->
          <section
            class="overflow-y-auto border-b border-zinc-200 px-2 [scrollbar-gutter:stable] md:border-r md:border-b-0 dark:border-zinc-800"
          >
            <div class="flex items-center justify-between gap-3 px-3 py-3">
              <div class="text-sm text-zinc-600 dark:text-zinc-400">
                <span class="font-semibold text-sky-700 dark:text-emerald-400">
                  {{ filtered.length }}
                </span>
                {{ tab }}{{ filtered.length === 1 ? '' : 's' }}
                <span
                  v-if="store.isLoading"
                  class="ml-2 text-xs text-zinc-500 dark:text-zinc-500"
                >
                  loading…
                </span>
              </div>

              <TheButton
                @click="reset"
                class="flex items-center gap-2 px-3 py-2"
              >
                <SquaresPlusIcon class="size-4 stroke-2" />
                New
              </TheButton>
            </div>

            <div class="px-3 pb-2 text-xs text-sky-600 dark:text-emerald-400">
              Select an item to edit
            </div>

            <div class="px-2 pb-3">
              <button
                v-for="p in filtered"
                :key="p.id"
                class="mb-2 w-full rounded-xl border p-2 text-left transition"
                :class="
                  selectedId === p.id
                    ? 'border-sky-500/35 bg-sky-50 ring-1 ring-sky-500/15 dark:border-emerald-400/30 dark:bg-emerald-950/15 dark:ring-emerald-400/15'
                    : 'border-zinc-200 bg-white hover:bg-zinc-50 dark:border-zinc-800 dark:bg-zinc-900/50 dark:hover:bg-zinc-900/80'
                "
                @click="pick(p)"
              >
                <div class="flex items-center justify-between gap-3">
                  <div class="line-clamp-1 text-sm font-semibold text-zinc-900 dark:text-zinc-100">
                    {{ p.productName }}
                  </div>
                  <div class="text-xs whitespace-nowrap text-zinc-500 dark:text-zinc-400">
                    {{
                      p.pricingMode === 'hourly'
                        ? money(p.hourlyRateMinor)
                        : money(p.flatPriceMinor)
                    }}
                  </div>
                </div>

                <div class="mt-1 text-xs text-zinc-500 dark:text-zinc-500">
                  {{ p.productType }} • {{ p.pricingMode }}
                  <span v-if="p.pricingMode === 'hourly'">• {{ p.minutesWorked ?? '—' }} min</span>
                </div>
              </button>

              <div
                v-if="filtered.length === 0"
                class="rounded-xl border border-dashed border-zinc-300 p-4 text-sm text-zinc-600 dark:border-zinc-700 dark:text-zinc-400"
              >
                No {{ tab }}s.
              </div>
            </div>
          </section>

          <!-- form -->
          <section class="overflow-y-auto px-6 py-4 [scrollbar-gutter:stable]">
            <div class="flex items-start justify-between gap-3">
              <div>
                <div class="text-base font-semibold text-zinc-900 dark:text-zinc-100">
                  {{ form.id ? 'Edit' : 'New' }} {{ tab }}
                </div>
                <div class="mt-0.5 text-xs text-sky-600 dark:text-emerald-400">
                  {{ tab === 'style' ? 'Flat price only' : 'Flat or hourly pricing' }}
                </div>
              </div>
            </div>

            <div class="mt-4 space-y-4">
              <TheInput
                v-model="form.productName"
                :label="tab + ' name'"
                class="w-full capitalize"
                placeholder="Name"
                autocomplete="off"
              />

              <div v-if="tab === 'sample'">
                <TheDropdown
                  v-model="form.pricingMode"
                  :options="['flat', 'hourly']"
                  select-title="Sample price type"
                />
              </div>

              <div v-if="tab === 'style' || form.pricingMode === 'flat'">
                <TheInput
                  v-model.number="form.flatPrice"
                  label="Flat price (£)"
                  type="number"
                  :placeholder="`${tab} price`"
                  min="0"
                  step="0.01"
                  class="w-full placeholder:capitalize"
                />
              </div>

              <div
                v-else
                class="grid grid-cols-1 gap-2 md:grid-cols-2"
              >
                <TheInput
                  v-model.number="form.hourlyRate"
                  label="Hourly rate (£/hr)"
                  type="number"
                  min="0"
                  step="0.01"
                  class="w-full"
                />
                <TheInput
                  v-model.number="form.minutesWorked"
                  type="number"
                  label="Minutes"
                  min="0"
                  step="1"
                  class="w-full"
                />
              </div>

              <!-- buttons -->
              <div class="flex flex-wrap gap-2 pt-2">
                <button
                  class="inline-flex cursor-pointer items-center gap-2 rounded-lg border border-sky-200 bg-sky-50 px-3 py-2 text-sm font-medium text-sky-700 transition hover:bg-sky-100 dark:border-emerald-400/20 dark:bg-emerald-950/25 dark:text-emerald-200 dark:hover:bg-emerald-950/40"
                  @click="save"
                >
                  <ShieldCheckIcon class="size-4" />
                  {{ form.id ? 'Save' : 'Create' }}
                </button>

                <button
                  class="inline-flex cursor-pointer items-center gap-2 rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm font-medium text-zinc-700 transition hover:bg-zinc-50 dark:border-zinc-800 dark:bg-zinc-900/50 dark:text-zinc-200 dark:hover:bg-zinc-950/15"
                  @click="reset"
                >
                  <ArrowPathIcon class="size-4" />
                  Reset
                </button>

                <button
                  v-if="form.id"
                  class="inline-flex cursor-pointer items-center gap-2 rounded-lg border border-rose-200 bg-rose-50 px-3 py-2 text-sm font-medium text-rose-700 transition hover:bg-rose-100 dark:border-rose-400/20 dark:bg-rose-950/25 dark:text-rose-200 dark:hover:bg-rose-950/40"
                  @click="del"
                >
                  <TrashIcon class="size-4" />
                  Delete
                </button>
              </div>

              <div
                v-if="store.error"
                class="rounded-lg border border-sky-200 bg-sky-50 p-3 text-sm text-sky-800 dark:border-emerald-400/20 dark:bg-emerald-950/25 dark:text-emerald-200"
              >
                {{ store.error }}
              </div>
            </div>
          </section>
        </div>
      </aside>
    </transition>
  </Teleport>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
.slide-enter-active,
.slide-leave-active {
  transition: transform 0.25s ease;
}
.slide-enter-from,
.slide-leave-to {
  transform: translateX(100%);
}
</style>

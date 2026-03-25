<script setup lang="ts">
import { computed, nextTick, reactive, ref, useTemplateRef, watch } from 'vue'
import { useProductStore } from '@/stores/products'
import { useClientStore } from '@/stores/clients'
import type { Product, ProductType, PricingMode, ProductUpsert } from '@/utils/productHttpHandler'
import {
  XMarkIcon,
  TrashIcon,
  UserIcon,
  ChevronDownIcon,
  ShieldCheckIcon,
  SquaresPlusIcon,
  MagnifyingGlassIcon,
} from '@heroicons/vue/24/outline'
import TheDropdown from '../UI/TheDropdown.vue'
import TheButton from '../UI/TheButton.vue'
import TheInput from '../UI/TheInput.vue'
import { useEscape } from '@/composables/keyHandlers'
import TheTooltip from '../UI/TheTooltip.vue'
import { fmtDisplayDate } from '@/utils/dates'
import { validateProductForm } from '@/utils/frontendValidation'
import { emitToastSuccess } from '@/utils/toast'
import { handleActionError } from '@/utils/errors/handleActionError'
import DecorGradient from '../UI/DecorGradient.vue'

const store = useProductStore()
const clientStore = useClientStore()

withDefaults(
  defineProps<{
    iconOnly?: boolean
  }>(),
  {
    iconOnly: true,
  },
)

const tab = ref<ProductType>('style')
const q = ref('')
const selectedId = ref<number | null>(null)
const fieldErrors = ref<Record<string, string>>({})
const isSaving = ref(false)
const isDeleting = ref(false)
type FocusableInput = {
  focus: () => void
}

const tabNameRef = useTemplateRef<FocusableInput>('tabNameRef')
type Form = {
  id: number | null
  productName: string
  pricingMode: PricingMode
  flatPrice: number | null
  hourlyRate: number | null
  minutesWorked: number | null
  createdAt?: string | null
  updatedAt?: string | null
}

const form = reactive<Form>({
  id: null,
  productName: '',
  pricingMode: 'flat',
  flatPrice: null,
  hourlyRate: null,
  minutesWorked: null,
  createdAt: null,
  updatedAt: null,
})

const liveFieldErrors = computed(() =>
  validateProductForm({
    productType: tab.value,
    pricingMode: tab.value === 'style' ? 'flat' : form.pricingMode,
    productName: form.productName,
    flatPrice: form.flatPrice,
    hourlyRate: form.hourlyRate,
    minutesWorked: form.minutesWorked,
  }),
)

const displayFieldErrors = computed(() => ({
  ...fieldErrors.value,
  ...liveFieldErrors.value,
}))

const canSave = computed(() => Object.keys(liveFieldErrors.value).length === 0)

watch(
  () => [
    tab.value,
    form.productName,
    form.pricingMode,
    form.flatPrice,
    form.hourlyRate,
    form.minutesWorked,
  ],
  () => {
    fieldErrors.value = {}
  },
)

const list = computed(() => store.byType[tab.value] ?? [])

const filtered = computed(() => {
  const s = q.value.trim().toLowerCase()
  if (!s) return list.value
  return list.value.filter((p) => p.productName.toLowerCase().includes(s))
})

watch(
  () => store.open,
  (v) => {
    if (!v) return
    void focusNameInput()
    void store.reload().catch((err) => {
      handleActionError(err, {
        toastTitle: 'Could not load products',
        mapFields: false,
      })
    })
  },
)

watch(tab, () => reset())

async function focusNameInput() {
  await nextTick()
  tabNameRef.value?.focus()
}

function reset() {
  selectedId.value = null
  q.value = ''
  form.id = null
  form.productName = ''
  form.pricingMode = 'flat'
  form.flatPrice = null
  form.hourlyRate = null
  form.minutesWorked = null
  form.createdAt = null
  form.updatedAt = null
  fieldErrors.value = {}
  void focusNameInput()
}

function pick(p: Product) {
  selectedId.value = p.id
  form.id = p.id
  form.productName = p.productName
  form.pricingMode = p.pricingMode
  form.flatPrice = p.flatPriceMinor != null ? p.flatPriceMinor / 100 : null
  form.hourlyRate = p.hourlyRateMinor != null ? p.hourlyRateMinor / 100 : null
  form.minutesWorked = p.minutesWorked ?? null
  form.createdAt = p.created_at ?? null
  form.updatedAt = p.updated_at ?? null
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

function labelForTab(type: ProductType) {
  return type === 'style' ? 'Style' : 'Sample'
}

async function save() {
  if (isSaving.value || isDeleting.value) return

  fieldErrors.value = {}
  if (!canSave.value) return

  isSaving.value = true
  const payload = buildUpsert()
  const isEditing = form.id != null

  try {
    if (!isEditing) {
      const created = await store.create(payload)
      emitToastSuccess('Created successfully.', {
        title: `${labelForTab(tab.value)} ${created.productName}`,
      })
      void focusNameInput()
      // pick(created) - prefer new style menu
    } else {
      const updated = await store.update(form.id!, payload)
      emitToastSuccess('Updated successfully.', {
        title: `${labelForTab(tab.value)} ${updated.productName}`,
      })
      pick(updated)
    }
  } catch (err: unknown) {
    handleActionError(err, {
      toastTitle: isEditing ? 'Update failed' : 'Create failed',
      mapFields: true,
      fieldErrors,
    })
  } finally {
    isSaving.value = false
  }
}

async function del() {
  if (!form.id || isDeleting.value || isSaving.value) return

  isDeleting.value = true
  const name = form.productName

  try {
    await store.remove(form.id)
    emitToastSuccess('Deleted successfully.', {
      title: `${labelForTab(tab.value)} ${name}`,
    })
    reset()
  } catch (err: unknown) {
    handleActionError(err, {
      toastTitle: 'Delete failed',
      mapFields: false,
    })
  } finally {
    isDeleting.value = false
  }
}

function money(minor?: number) {
  if (minor == null) return '—'
  return new Intl.NumberFormat('en-GB', { style: 'currency', currency: 'GBP' }).format(minor / 100)
}

useEscape(
  () => {
    store.open = false
  },
  {
    enabled: () => store.open,
  },
)
</script>

<template>
  <TheTooltip
    side="bottom"
    align="end"
  >
    <template #content>
      <span class="mr-1 text-sky-600 dark:text-emerald-400">Products:</span>
      <br />
      <div class="mt-1">
        <kbd>Ctrl</kbd>
        +
        <kbd>I</kbd>
      </div>
    </template>
    <button
      v-if="iconOnly"
      type="button"
      class="flex cursor-pointer rounded-lg border border-zinc-300 p-1 text-zinc-600 hover:text-sky-600 dark:border-transparent dark:text-zinc-300 dark:hover:bg-zinc-800 dark:hover:text-emerald-400"
    >
      <SquaresPlusIcon
        class="size-6"
        @click="store.open = true"
      />
    </button>
    <TheButton
      v-else
      @click="store.open = true"
      class="cursor-pointer"
    >
      <SquaresPlusIcon class="size-4"></SquaresPlusIcon>
      items
    </TheButton>
  </TheTooltip>

  <Teleport to="body">
    <transition name="fade">
      <div
        v-if="store.open"
        class="fixed inset-0 z-100 bg-black/45 backdrop-blur-[1px]"
        @click="store.open = false"
      />
    </transition>

    <!-- Top most -->
    <transition name="slide">
      <aside
        v-if="store.open"
        ref="editorRef"
        class="fixed top-0 right-0 z-101 h-screen w-[92vw] max-w-225 border-l border-zinc-200 bg-white text-zinc-900 dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-100"
      >
        <!-- header -->
        <header
          class="border-b border-zinc-200 bg-zinc-50 dark:border-zinc-800 dark:bg-zinc-900/70"
        >
          <div class="relative overflow-hidden px-4 py-3">
            <DecorGradient variant="gradientAndGrid"></DecorGradient>

            <!-- CONTENT -->
            <div class="relative z-10 flex items-center justify-between gap-4">
              <div class="flex min-w-0 items-center gap-3">
                <div
                  class="grid size-12 shrink-0 place-items-center rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-700 dark:bg-zinc-900"
                >
                  <SquaresPlusIcon class="stroke-1.5 size-7 text-sky-700 dark:text-emerald-400" />
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
                      class="hidden rounded-full border border-zinc-200 bg-white/90 px-2 py-0.5 text-xs font-medium text-zinc-600 backdrop-blur-sm sm:block dark:border-zinc-700 dark:bg-zinc-900/70 dark:text-zinc-400"
                    >
                      {{ tab === 'style' ? 'Styles' : 'Samples' }}
                    </span>
                  </div>

                  <div class="text-sm tracking-tight text-zinc-500 dark:text-zinc-300">
                    Manage client services, pricing and work units
                  </div>
                </div>
              </div>

              <!-- Close -->
              <TheTooltip side="bottom">
                <template #content>
                  <div class="flex items-center text-start">
                    <span class="mr-1 text-sky-600 dark:text-emerald-400">Shortcut:</span>
                    <kbd>Esc</kbd>
                  </div>
                </template>
                <button
                  type="button"
                  class="shrink-0 cursor-pointer rounded-lg p-2 text-zinc-600 hover:bg-rose-50 hover:text-rose-400 dark:text-zinc-300 dark:hover:bg-rose-400/15 dark:hover:text-rose-300"
                  @click="store.open = false"
                >
                  <XMarkIcon class="size-5" />
                </button>
              </TheTooltip>
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
                input-class="overflow-x-clip"
              />
            </div>

            <!-- Tabs -->
            <div
              class="flex shrink-0 rounded-full border border-zinc-200 bg-white p-1 shadow-sm dark:border-zinc-800 dark:bg-zinc-900"
            >
              <button
                type="button"
                class="rounded-full px-3 py-1.5 text-sm font-medium transition"
                :class="
                  tab === 'style'
                    ? 'bg-sky-100 text-sky-700 shadow-sm dark:bg-emerald-950/60 dark:text-emerald-200'
                    : 'text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100'
                "
                @click="tab = 'style'"
              >
                Styles
              </button>

              <button
                type="button"
                class="rounded-full px-3 py-1.5 text-sm font-medium transition"
                :class="
                  tab === 'sample'
                    ? 'bg-sky-100 text-sky-700 shadow-sm dark:bg-emerald-950/60 dark:text-emerald-200'
                    : 'text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100'
                "
                @click="tab = 'sample'"
              >
                Samples
              </button>
            </div>

            <!-- Search -->
            <div class="hidden w-72 shrink-0 px-3 sm:block">
              <div class="relative shadow-md">
                <MagnifyingGlassIcon
                  class="pointer-events-none absolute top-1/2 left-2 size-5 -translate-y-1/2 text-zinc-500 dark:text-zinc-400"
                />
                <input
                  v-model="q"
                  class="input input-accent pl-9"
                  id="product-search"
                  :placeholder="`Search ${tab}s…`"
                />
              </div>
            </div>
          </div>
        </header>

        <div class="grid h-[calc(100%-150px)] grid-cols-1 md:grid-cols-2">
          <!-- list -->
          <section
            class="overflow-y-auto border-b border-zinc-200 px-2 pb-4 [scrollbar-gutter:stable] md:border-r md:border-b-0 dark:border-zinc-800"
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

            <div
              v-if="store.loadError"
              class="mx-2 mb-2 rounded-lg border border-amber-300 bg-amber-50 px-3 py-2 text-xs text-amber-700 dark:border-amber-400/40 dark:bg-amber-950/30 dark:text-amber-200"
            >
              {{ store.loadError }}
            </div>

            <div class="px-2 pb-3">
              <button
                type="button"
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
                  <div class="text-xs whitespace-nowrap text-zinc-500 dark:text-zinc-300">
                    {{
                      p.pricingMode === 'hourly'
                        ? money(p.hourlyRateMinor) + ' (£/hr)'
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

            <div class="mt-4 space-y-4 pb-12 sm:pb-0">
              <TheInput
                ref="tabNameRef"
                v-model="form.productName"
                :label="tab + ' name'"
                class="w-full capitalize"
                placeholder="Name"
                autocomplete="off"
                :error="displayFieldErrors.productName"
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
                  :error="displayFieldErrors.flatPrice"
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
                  :error="displayFieldErrors.hourlyRate"
                />
                <TheInput
                  v-model.number="form.minutesWorked"
                  type="number"
                  label="Minutes"
                  min="0"
                  step="1"
                  class="w-full"
                  :error="displayFieldErrors.minutesWorked"
                />
              </div>
              <div
                v-if="form.id"
                class="rounded-2xl border border-zinc-200/80 bg-zinc-50/80 p-3 dark:border-zinc-800 dark:bg-zinc-900/60"
              >
                <div
                  class="text-tiny mb-2 font-medium tracking-[0.12em] text-zinc-500 uppercase dark:text-zinc-400"
                >
                  Record details
                </div>

                <div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
                  <div
                    class="rounded-xl border border-zinc-200 bg-white px-3 py-2 dark:border-zinc-800 dark:bg-zinc-950/60"
                  >
                    <div
                      class="text-tiny font-medium tracking-wide text-zinc-500 uppercase dark:text-zinc-400"
                    >
                      Created
                    </div>
                    <div class="mt-1 text-sm font-medium text-zinc-900 dark:text-zinc-100">
                      {{ form.createdAt ? fmtDisplayDate(new Date(form.createdAt)) : 'N/A' }}
                    </div>
                  </div>

                  <div
                    class="rounded-xl border border-zinc-200 bg-white px-3 py-2 dark:border-zinc-800 dark:bg-zinc-950/60"
                  >
                    <div
                      class="text-tiny font-medium tracking-wide text-zinc-500 uppercase dark:text-zinc-400"
                    >
                      Last updated
                    </div>
                    <div class="mt-1 text-sm font-medium text-zinc-900 dark:text-zinc-100">
                      {{ form.updatedAt ? fmtDisplayDate(new Date(form.updatedAt)) : 'N/A' }}
                    </div>
                  </div>
                </div>
              </div>

              <!-- buttons -->
              <div class="flex flex-wrap gap-2 pt-2">
                <button
                  type="button"
                  class="inline-flex cursor-pointer items-center gap-2 rounded-lg border border-sky-200 bg-sky-50 px-3 py-2 text-sm font-medium text-sky-700 transition hover:bg-sky-100 focus-visible:ring-2 focus-visible:ring-sky-500/30 focus-visible:ring-offset-2 focus-visible:ring-offset-white focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-60 dark:border-emerald-400/20 dark:bg-emerald-950/25 dark:text-emerald-200 dark:hover:bg-emerald-950/40 dark:focus-visible:ring-emerald-400/25 dark:focus-visible:ring-offset-zinc-800"
                  :disabled="!canSave || isSaving || isDeleting"
                  @click="save"
                >
                  <ShieldCheckIcon class="size-4" />
                  {{ isSaving ? (form.id ? 'Saving…' : 'Creating…') : form.id ? 'Save' : 'Create' }}
                </button>

                <button
                  type="button"
                  v-if="form.id"
                  class="inline-flex cursor-pointer items-center gap-2 rounded-lg border border-rose-200 bg-rose-50 px-3 py-2 text-sm font-medium text-rose-700 transition hover:bg-rose-100 focus-visible:ring-2 focus-visible:ring-sky-500/30 focus-visible:ring-offset-2 focus-visible:ring-offset-white focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-60 dark:border-rose-400/20 dark:bg-rose-950/25 dark:text-rose-200 dark:hover:bg-rose-950/40 dark:focus-visible:ring-rose-400/25 dark:focus-visible:ring-offset-zinc-800"
                  :disabled="isDeleting || isSaving"
                  @click="del"
                >
                  <TrashIcon class="size-4" />
                  {{ isDeleting ? 'Deleting…' : 'Delete' }}
                </button>
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

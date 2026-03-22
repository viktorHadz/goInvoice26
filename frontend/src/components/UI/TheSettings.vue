<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  Cog6ToothIcon,
  XMarkIcon,
  PhotoIcon,
  SwatchIcon,
  BuildingOffice2Icon,
  DocumentTextIcon,
  BanknotesIcon,
} from '@heroicons/vue/24/outline'

import TheTooltip from './TheTooltip.vue'
import TheButton from './TheButton.vue'
import TheInput from './TheInput.vue'
import TheDropdown from './TheDropdown.vue'
import DecorGradient from '@/components/UI/DecorGradient.vue'
import { useEscape, useShortcuts, type ShortcutDefinition } from '@/composables/keyHandlers'
import SettingsPreview from './SettingsPreview.vue'
import {
  useSettingsStore,
  type Settings,
  type CurrencyCode,
  type DateFormat,
} from '@/stores/settings'
import { emitToastError, emitToastSuccess } from '@/utils/toast'
import { handleImageUpload, readImagePreview } from '@/utils/settingHandlers'

const settingsStore = useSettingsStore()

const settingsOpen = ref(false)
const form = ref<Settings | null>(null)
const logoPreview = ref<string | null>(null)
const logoFile = ref<File | null>(null)
const isSaving = ref(false)

const title = 'Invoice Settings'
const subtitle = 'Manage business identity, invoice defaults and PDF display options'

const currencyOptions: CurrencyCode[] = ['GBP', 'EUR', 'USD']
const dateFormatOptions: DateFormat[] = ['dd/mm/yyyy', 'mm/dd/yyyy', 'yyyy-mm-dd']

const PAYMENT_TERMS_MAX = 1000
const PAYMENT_DETAILS_MAX = 250
const FOOTER_NOTE_MAX = 180

const hasLogo = computed(() => Boolean(logoPreview.value))

function getErrorMessage(err: unknown, fallback: string) {
  return err instanceof Error ? err.message : fallback
}

function cloneSettings(settings: Settings): Settings {
  return { ...settings }
}

async function openSettings() {
  try {
    const settings = settingsStore.settings ?? (await settingsStore.fetchSettings())
    if (!settings) throw new Error('Settings not found')

    form.value = cloneSettings(settings)
    logoPreview.value = form.value.logoUrl || null
    logoFile.value = null
    settingsOpen.value = true
  } catch (err) {
    emitToastError({ message: getErrorMessage(err, 'Failed to load settings.') })
  }
}

function closeSettings() {
  if (isSaving.value) return
  settingsOpen.value = false
}

function forceCloseSettings() {
  settingsOpen.value = false
}

async function onLogoChange(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''

  if (!file || !form.value) return

  try {
    logoPreview.value = await readImagePreview(file)
    logoFile.value = file
  } catch (err) {
    logoFile.value = null
    logoPreview.value = form.value.logoUrl || null
    emitToastError({ message: getErrorMessage(err, 'Failed to read selected image.') })
  }
}

function removeLogo() {
  if (!form.value) return
  logoPreview.value = null
  logoFile.value = null
  form.value.logoUrl = ''
}

async function save() {
  if (isSaving.value || !form.value) return

  isSaving.value = true

  try {
    let logoUrl = form.value.logoUrl

    if (logoFile.value) {
      const uploaded = await handleImageUpload(logoFile.value)
      logoUrl = uploaded.logoUrl
    }

    await settingsStore.saveSettings({
      ...form.value,
      logoUrl,
    })

    logoFile.value = null
    logoPreview.value = logoUrl || null

    emitToastSuccess('Settings saved.')
    forceCloseSettings()
  } catch (err) {
    emitToastError({ message: getErrorMessage(err, 'Failed to save settings.') })
  } finally {
    isSaving.value = false
  }
}

watch(
  () => settingsStore.needsSetup,
  (needsSetup) => {
    if (needsSetup && !settingsOpen.value) {
      openSettings()
    }
  },
  { immediate: true },
)
const shortcuts: ShortcutDefinition[] = [
  { key: 's', modifiers: ['alt', 'shift'], action: () => openSettings() },
]
useShortcuts(shortcuts)

useEscape(closeSettings, {
  enabled: () => settingsOpen.value,
})
</script>
<template>
  <TheTooltip
    side="bottom"
    align="end"
  >
    <template #content>
      <span class="text-sky-600 dark:text-emerald-400">Settings:</span>
      <br />
      <div class="mt-1">
        <kbd>Alt</kbd>
        +
        <kbd>Shift</kbd>
        +
        <kbd>S</kbd>
      </div>
    </template>

    <button
      type="button"
      class="flex cursor-pointer rounded-lg border border-zinc-300 p-1 text-zinc-600 hover:text-sky-600 dark:border-transparent dark:text-zinc-300 dark:hover:bg-zinc-800 dark:hover:text-emerald-400"
      @click="() => openSettings()"
    >
      <Cog6ToothIcon class="size-6 stroke-1" />
    </button>
  </TheTooltip>

  <Teleport to="body">
    <transition name="fade">
      <div
        v-if="settingsOpen"
        class="fixed inset-0 z-100 bg-black/45 backdrop-blur-[2px]"
        @click="() => closeSettings()"
      />
    </transition>

    <transition name="modal">
      <section
        v-if="settingsOpen"
        class="fixed inset-0 z-101 m-auto flex h-[88vh] w-[94vw] max-w-6xl flex-col overflow-hidden rounded-3xl border border-zinc-200 bg-white text-zinc-900 shadow-2xl dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-100"
      >
        <!-- Header -->
        <header
          class="relative overflow-hidden border-b border-zinc-200 bg-zinc-50 dark:border-zinc-800 dark:bg-zinc-900/80"
        >
          <DecorGradient></DecorGradient>
          <!-- More glow -->
          <DecorGradient></DecorGradient>

          <div class="relative z-10 flex items-start justify-between gap-4 px-5 py-4">
            <div class="flex min-w-0 items-center gap-4">
              <div
                class="grid size-12 shrink-0 place-items-center rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-700 dark:bg-zinc-900"
              >
                <Cog6ToothIcon class="stroke-1.5 size-7 text-sky-700 dark:text-emerald-400" />
              </div>

              <div class="min-w-0">
                <h2 class="text-2xl font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
                  {{ title }}
                </h2>

                <p class="mt-1 text-sm tracking-tight text-zinc-500 dark:text-zinc-300">
                  {{ subtitle }}
                </p>
              </div>
            </div>

            <TheTooltip side="bottom">
              <template #content>
                <div class="flex items-center text-start">
                  <span class="mr-1 text-sky-600 dark:text-emerald-400">Shortcut:</span>
                  <kbd>Esc</kbd>
                </div>
              </template>

              <button
                type="button"
                class="shrink-0 cursor-pointer rounded-lg p-2 text-zinc-600 transition hover:bg-rose-50 hover:text-rose-500 dark:text-zinc-300 dark:hover:bg-rose-400/15 dark:hover:text-rose-300"
                @click="() => closeSettings()"
              >
                <XMarkIcon class="size-5" />
              </button>
            </TheTooltip>
          </div>
        </header>

        <!-- Body -->
        <div
          class="grid min-h-0 flex-1 grid-cols-1 gap-0 lg:grid-cols-2"
          v-if="form"
        >
          <!-- Left -->
          <div
            class="min-h-0 overflow-y-auto border-b border-zinc-200 px-5 py-5 lg:border-r lg:border-b-0 dark:border-zinc-800"
          >
            <div class="space-y-5 pb-10">
              <!-- Business identity -->
              <section
                class="rounded-2xl border border-zinc-200 bg-zinc-50/70 p-4 dark:border-zinc-800 dark:bg-zinc-950/30"
              >
                <div class="mb-4 flex items-center gap-2">
                  <BuildingOffice2Icon class="size-5 text-sky-600 dark:text-emerald-400" />
                  <h2 class="font-semibold tracking-wide text-zinc-900 dark:text-zinc-100">
                    Business details
                  </h2>
                </div>

                <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                  <TheInput
                    v-model="form.companyName"
                    label="Name"
                    :input-max-length="50"
                    placeholder="Your business name"
                    autocomplete="name"
                  />

                  <TheInput
                    v-model="form.email"
                    label="Email"
                    :input-max-length="50"
                    placeholder="name@company.com"
                    autocomplete="email"
                  />

                  <TheInput
                    v-model="form.phone"
                    label="Phone"
                    :input-max-length="20"
                    placeholder="+44..."
                    autocomplete="tel"
                  />

                  <TheInput
                    v-model="form.invoicePrefix"
                    :input-max-length="50"
                    label="Invoice prefix"
                    placeholder="INV-"
                  />
                </div>

                <div class="mt-4">
                  <label
                    for="setts-c-addr"
                    class="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-300"
                  >
                    Company address
                  </label>
                  <textarea
                    id="setts-c-addr"
                    v-model="form.companyAddress"
                    rows="4"
                    maxlength="160"
                    class="input input-accent min-h-28 w-full resize-y rounded-xl px-3 py-2"
                    placeholder="Company name&#10;Street&#10;City&#10;Postcode"
                  />
                </div>
              </section>

              <!-- Invoice defaults -->
              <section
                class="rounded-2xl border border-zinc-200 bg-zinc-50/70 p-4 dark:border-zinc-800 dark:bg-zinc-950/30"
              >
                <div class="mb-4 flex items-center gap-2">
                  <DocumentTextIcon class="size-5 text-sky-600 dark:text-emerald-400" />
                  <h2 class="font-semibold tracking-wide text-zinc-900 dark:text-zinc-100">
                    Invoice defaults
                  </h2>
                </div>

                <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                  <TheDropdown
                    v-model="form.currency"
                    :options="currencyOptions"
                    select-title="Currency"
                  />

                  <TheDropdown
                    v-model="form.dateFormat"
                    :options="dateFormatOptions"
                    select-title="Date format"
                  />
                </div>
              </section>

              <!-- Payment -->
              <section
                class="rounded-2xl border border-zinc-200 bg-zinc-50/70 p-4 dark:border-zinc-800 dark:bg-zinc-950/30"
              >
                <div class="mb-4 flex items-center gap-2">
                  <BanknotesIcon class="size-5 text-sky-600 dark:text-emerald-400" />
                  <h2 class="font-semibold tracking-wide text-zinc-900 dark:text-zinc-100">
                    Payment and footer
                  </h2>
                </div>

                <div class="space-y-4">
                  <div>
                    <label
                      for="setts-payment-terms"
                      class="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-300"
                    >
                      Payment terms
                    </label>
                    <textarea
                      v-model="form.paymentTerms"
                      id="setts-payment-terms"
                      rows="3"
                      :maxlength="PAYMENT_TERMS_MAX"
                      class="input input-accent w-full resize-y rounded-xl px-3 py-2"
                      placeholder="Payment terms shown on invoices"
                    />
                    <div
                      class="mt-1 text-right text-xs"
                      :class="
                        form.paymentTerms.length > PAYMENT_TERMS_MAX * 0.9
                          ? 'text-rose-600 dark:text-rose-300'
                          : form.paymentTerms.length > PAYMENT_TERMS_MAX * 0.8
                            ? 'text-amber-600 dark:text-amber-400'
                            : 'text-zinc-500 dark:text-zinc-400'
                      "
                    >
                      {{ form.paymentTerms.length }}/{{ PAYMENT_TERMS_MAX }}
                    </div>
                  </div>

                  <div>
                    <label
                      for="setts-pmnt-details"
                      class="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-300"
                    >
                      Payment details
                    </label>
                    <textarea
                      id="setts-pmnt-details"
                      v-model="form.paymentDetails"
                      rows="3"
                      :maxlength="PAYMENT_DETAILS_MAX"
                      class="input input-accent w-full resize-y rounded-xl px-3 py-2"
                      placeholder="Bank transfer details, sort code, account number, IBAN, etc."
                    />
                    <div
                      class="mt-1 text-right text-xs"
                      :class="
                        form.paymentDetails.length > PAYMENT_DETAILS_MAX * 0.9
                          ? 'text-rose-600 dark:text-rose-300'
                          : form.paymentDetails.length > PAYMENT_DETAILS_MAX * 0.8
                            ? 'text-amber-600 dark:text-amber-400'
                            : 'text-zinc-500 dark:text-zinc-400'
                      "
                    >
                      {{ form.paymentDetails.length }}/{{ PAYMENT_DETAILS_MAX }}
                    </div>
                  </div>

                  <div>
                    <label
                      for="setts-notes-footer"
                      class="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-300"
                    >
                      Footer note
                    </label>
                    <textarea
                      id="setts-notes-footer"
                      v-model="form.notesFooter"
                      rows="3"
                      :maxlength="FOOTER_NOTE_MAX"
                      class="input input-accent w-full resize-y rounded-xl px-3 py-2"
                      placeholder="Optional footer or thank-you note"
                    />

                    <div
                      class="mt-1 text-right text-xs"
                      :class="
                        form.notesFooter.length > FOOTER_NOTE_MAX * 0.9
                          ? 'text-rose-600 dark:text-rose-300'
                          : form.notesFooter.length > FOOTER_NOTE_MAX * 0.8
                            ? 'text-amber-600 dark:text-amber-400'
                            : 'text-zinc-500 dark:text-zinc-400'
                      "
                    >
                      {{ form.notesFooter.length }}/{{ FOOTER_NOTE_MAX }}
                    </div>
                  </div>
                </div>
              </section>
            </div>
          </div>

          <!-- Right -->
          <div class="min-h-0 overflow-y-auto px-5 py-5">
            <div class="space-y-5 pb-10">
              <!-- Logo -->
              <section
                class="rounded-2xl border border-zinc-200 bg-zinc-50/70 p-4 dark:border-zinc-800 dark:bg-zinc-950/30"
              >
                <div class="mb-4 flex items-center gap-2">
                  <PhotoIcon class="size-5 text-sky-600 dark:text-emerald-400" />
                  <h2 class="font-semibold tracking-wide text-zinc-900 dark:text-zinc-100">
                    Invoice logo
                  </h2>
                </div>

                <div
                  class="grid min-h-56 place-items-center rounded-2xl border border-dashed border-zinc-300 bg-white p-4 dark:border-zinc-700 dark:bg-zinc-950/50"
                >
                  <div
                    v-if="hasLogo"
                    class="flex w-full flex-col items-center gap-4"
                  >
                    <img
                      :src="logoPreview!"
                      alt="Invoice logo preview"
                      class="max-h-36 rounded-xl object-contain"
                    />

                    <div class="flex flex-wrap justify-center gap-2">
                      <label
                        class="inline-flex cursor-pointer items-center gap-2 rounded-lg border border-sky-200 bg-sky-50 px-3 py-2 text-sm font-medium text-sky-700 transition hover:bg-sky-100 dark:border-emerald-400/20 dark:bg-emerald-950/25 dark:text-emerald-200 dark:hover:bg-emerald-950/40"
                      >
                        <input
                          type="file"
                          accept="image/png,image/jpeg,image/webp"
                          class="hidden"
                          @change="onLogoChange"
                        />
                        Replace image
                      </label>
                      <TheButton
                        type="button"
                        @click="removeLogo"
                        variant="danger"
                        class="cursor-pointer"
                      >
                        Remove
                      </TheButton>
                    </div>
                  </div>

                  <div
                    v-else
                    class="flex flex-col items-center gap-3 text-center"
                  >
                    <div
                      class="grid size-14 place-items-center rounded-2xl border border-zinc-200 bg-zinc-50 dark:border-zinc-800 dark:bg-zinc-900"
                    >
                      <PhotoIcon class="size-7 text-zinc-600 dark:text-zinc-400" />
                    </div>

                    <div>
                      <div class="text-sm font-medium text-zinc-900 dark:text-zinc-100">
                        Upload logo
                      </div>
                      <div class="mt-1 text-xs text-zinc-600 dark:text-zinc-400">
                        PNG, JPG, or WebP for invoice header
                      </div>
                    </div>

                    <label
                      class="inline-flex cursor-pointer items-center gap-2 rounded-lg border border-sky-200 bg-sky-50 px-3 py-2 text-sm font-medium text-sky-700 transition hover:bg-sky-100 dark:border-emerald-400/20 dark:bg-emerald-950/25 dark:text-emerald-200 dark:hover:bg-emerald-950/40"
                    >
                      <input
                        type="file"
                        accept="image/png,image/jpeg,image/webp"
                        class="hidden"
                        @change="onLogoChange"
                      />
                      Choose image
                    </label>
                  </div>
                </div>
              </section>

              <!-- Preview Pannel -->
              <section
                class="rounded-2xl border border-zinc-200 bg-zinc-50/70 p-4 dark:border-zinc-800 dark:bg-zinc-950/30"
              >
                <div class="mb-4 flex items-center gap-2">
                  <SwatchIcon class="size-5 text-sky-600 dark:text-emerald-400" />
                  <h2 class="font-semibold tracking-wide text-zinc-900 dark:text-zinc-100">
                    Display preview
                  </h2>
                </div>

                <SettingsPreview
                  :form="form"
                  :logo-preview="logoPreview"
                />
              </section>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <footer
          class="border-t border-zinc-200 bg-white/90 px-5 py-4 dark:border-zinc-800 dark:bg-zinc-900/90"
        >
          <div class="flex flex-wrap items-center justify-end gap-2">
            <TheButton
              type="button"
              @click="closeSettings"
              variant="danger"
              class="cursor-pointer"
            >
              Cancel
            </TheButton>
            <TheButton
              type="button"
              @click="save"
              variant="success"
              class="cursor-pointer"
            >
              Save settings
            </TheButton>
          </div>
        </footer>
      </section>
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

.modal-enter-active,
.modal-leave-active {
  transition:
    opacity 0.22s ease,
    transform 0.22s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
  transform: translateY(10px) scale(0.985);
}
</style>

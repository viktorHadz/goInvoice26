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
  InformationCircleIcon,
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
  type SettingsUpdate,
  type CurrencyCode,
  type DateFormat,
} from '@/stores/settings'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import { emitToastError, emitToastInfo, emitToastSuccess } from '@/utils/toast'
import { deleteLogo, readImagePreview, uploadLogo } from '@/utils/settingHandlers'

withDefaults(
  defineProps<{
    showTrigger?: boolean
  }>(),
  {
    showTrigger: true,
  },
)

const settingsStore = useSettingsStore()
const authStore = useAuthStore()
const router = useRouter()

const settingsOpen = ref(false)
const form = ref<Settings | null>(null)
const logoPreview = ref<string | null>(null)
const logoFile = ref<File | null>(null)
const isSaving = ref(false)

const canEditSettings = computed(() => authStore.isOwner)
const title = computed(() => (canEditSettings.value ? 'Invoice Settings' : 'Workspace Settings'))
const subtitle = computed(() =>
  canEditSettings.value
    ? 'Manage business identity, and invoice display'
    : 'View the workspace identity, branding, and invoice display settings',
)
const readOnlyNotice = 'Only the workspace admin can edit these settings'

const currencyOptions: CurrencyCode[] = ['GBP', 'EUR', 'USD']
const dateFormatOptions: DateFormat[] = ['dd/mm/yyyy', 'mm/dd/yyyy', 'yyyy-mm-dd']

const PAYMENT_TERMS_MAX = 1000
const PAYMENT_DETAILS_MAX = 250
const FOOTER_NOTE_MAX = 180

const hasLogo = computed(() => Boolean(logoPreview.value))
const canEditStartingInvoiceNumber = computed(
  () => form.value?.canEditStartingInvoiceNumber === true,
)
const startingInvoiceLockedMessage = computed(() =>
  canEditStartingInvoiceNumber.value
    ? 'You can choose the first invoice number before any invoices exist. Once invoices are created, this field locks until all invoices are deleted.'
    : 'Starting invoice number is locked because invoices already exist. Delete all saved invoices to unlock again.',
)
const startingInvoiceNumberModel = computed<number | null>({
  get() {
    return form.value?.startingInvoiceNumber ?? 1
  },
  set(next) {
    if (!form.value) return
    const n = Number(next ?? 0)
    form.value.startingInvoiceNumber = Number.isFinite(n) ? Math.max(0, Math.round(n)) : 0
  },
})

function getErrorMessage(err: unknown, fallback: string) {
  return err instanceof Error ? err.message : fallback
}

function cloneSettings(settings: Settings): Settings {
  return { ...settings }
}

function syncFormFromSettings(settings: Settings) {
  form.value = cloneSettings(settings)
  logoPreview.value = settings.logoUrl || null
  logoFile.value = null
}

function toSettingsPayload(settings: Settings): SettingsUpdate {
  const {
    logoUrl: _logoUrl,
    canEditStartingInvoiceNumber: _canEditStartingInvoiceNumber,
    ...payload
  } = settings

  return payload
}

async function openSettings() {
  if (!authStore.hasBillingAccess) {
    emitToastInfo(
      authStore.canManageBilling
        ? 'Activate billing to edit invoice settings and branding.'
        : 'The workspace admin needs to reactivate billing before settings are available.',
      { title: 'Workspace locked' },
    )
    void router.push({ name: 'billing' })
    return
  }

  try {
    const settings = await settingsStore.fetchSettings()
    if (!settings) throw new Error('Settings not found')

    syncFormFromSettings(settings)
    settingsOpen.value = true
  } catch (err) {
    emitToastError({
      title: 'Could not load settings',
      message: getErrorMessage(err, 'Failed to load settings.'),
    })
  }
}

function closeSettings() {
  if (isSaving.value) return
  settingsOpen.value = false
}

function forceCloseSettings() {
  settingsOpen.value = false
}

defineExpose({
  openSettings,
  closeSettings,
})

async function onLogoChange(e: Event) {
  if (!canEditSettings.value) return
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
    emitToastError({
      title: 'Could not read image',
      message: getErrorMessage(err, 'Failed to read selected image.'),
    })
  }
}

function removeLogo() {
  if (!canEditSettings.value) return
  if (!form.value) return
  logoPreview.value = null
  logoFile.value = null
  form.value.logoUrl = ''
}

async function save() {
  if (isSaving.value || !form.value || !canEditSettings.value) return

  isSaving.value = true

  try {
    const hadLogoBefore = Boolean(settingsStore.settings?.logoUrl)

    await settingsStore.saveSettings(toSettingsPayload(form.value))

    if (logoFile.value) {
      await uploadLogo(logoFile.value)
    } else if (!logoPreview.value && hadLogoBefore) {
      await deleteLogo()
    }

    const refreshed = await settingsStore.fetchSettings()
    if (!refreshed) throw new Error('Settings not found')

    form.value = cloneSettings(refreshed)
    logoFile.value = null
    logoPreview.value = refreshed.logoUrl || null

    emitToastSuccess('Settings saved.')
    forceCloseSettings()
  } catch (err) {
    emitToastError({
      title: 'Could not save settings',
      message: getErrorMessage(err, 'Failed to save settings.'),
    })
  } finally {
    isSaving.value = false
  }
}

watch(
  () => settingsStore.needsSetup,
  (needsSetup) => {
    if (needsSetup && authStore.hasBillingAccess && canEditSettings.value && !settingsOpen.value) {
      openSettings()
    }
  },
  { immediate: true },
)

watch(
  () => authStore.hasBillingAccess,
  (hasAccess) => {
    if (!hasAccess) {
      forceCloseSettings()
    }
  },
)

watch(
  () => settingsStore.settings,
  (settings) => {
    if (!settingsOpen.value || !settings || canEditSettings.value) return
    syncFormFromSettings(settings)
  },
)

const shortcuts: ShortcutDefinition[] = [
  { key: 's', modifiers: ['alt', 'shift'], action: () => openSettings() },
]
useShortcuts(shortcuts)

useEscape(closeSettings, {
  enabled: () => settingsOpen.value,
})

const cardClass = 'rounded-2xl border border-zinc-300 p-5 shadow-sm dark:border-zinc-800 card-grad'
const iconWrapClass =
  'rounded-xl border border-zinc-300 bg-zinc-50 p-2 text-zinc-700 dark:border-zinc-700 dark:bg-zinc-950/20 dark:text-zinc-200'
</script>
<template>
  <TheTooltip
    v-if="showTrigger"
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
      @click="void openSettings()"
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
        class="fixed inset-0 z-101 m-auto flex h-[88vh] w-[94vw] max-w-6xl flex-col overflow-hidden rounded-3xl border border-zinc-300 bg-white text-zinc-900 shadow-2xl dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-100"
      >
        <!-- Header -->
        <header
          class="relative overflow-hidden border-b border-zinc-300 bg-zinc-50 dark:border-zinc-800 dark:bg-zinc-900/80"
        >
          <DecorGradient />
          <div
            class="relative z-10 flex items-start justify-between gap-4 px-3 py-2 sm:px-5 sm:py-4"
          >
            <div class="flex min-w-0 items-center gap-4">
              <div
                class="grid size-10 shrink-0 place-items-center rounded-2xl border border-zinc-300 bg-white shadow-sm sm:size-12 dark:border-zinc-700 dark:bg-zinc-900"
              >
                <Cog6ToothIcon
                  class="stroke-1.5 size-6 text-sky-700 sm:size-7 dark:text-emerald-400"
                />
              </div>

              <div class="min-w-0">
                <div class="flex items-center gap-2">
                  <h2
                    class="text-xl font-semibold tracking-tight text-zinc-900 sm:text-2xl dark:text-zinc-100"
                  >
                    {{ title }}
                  </h2>
                </div>

                <p class="mt-1 text-sm tracking-tight text-zinc-600 dark:text-zinc-300">
                  {{ subtitle }}
                </p>
                <p
                  v-if="!canEditSettings"
                  class="mt-1 text-xs font-medium text-amber-600 dark:text-amber-300"
                >
                  {{ readOnlyNotice }}
                </p>
              </div>
            </div>

            <button
              type="button"
              class="shrink-0 cursor-pointer rounded-lg p-2 text-zinc-600 transition hover:bg-rose-50 hover:text-rose-500 dark:text-zinc-300 dark:hover:bg-rose-400/15 dark:hover:text-rose-300"
              @click="() => closeSettings()"
            >
              <XMarkIcon class="size-5" />
            </button>
          </div>
        </header>

        <!-- Body -->
        <div
          class="grid min-h-0 flex-1 grid-cols-1 gap-0 lg:grid-cols-2"
          v-if="form"
        >
          <!-- Left -->
          <div
            class="min-h-0 overflow-y-auto border-b border-zinc-300 p-3 sm:p-5 lg:border-r lg:border-b-0 dark:border-zinc-800"
          >
            <div class="space-y-5 pb-10">
              <!-- Business identity -->
              <section :class="cardClass">
                <fieldset :disabled="!canEditSettings">
                  <div class="mb-4 flex items-center gap-4">
                    <div :class="iconWrapClass">
                      <BuildingOffice2Icon class="size-5" />
                    </div>
                    <h2 class="font-semibold tracking-wide text-zinc-900 dark:text-zinc-100">
                      Business details
                    </h2>
                  </div>

                  <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                    <TheInput
                      v-model="form.companyName"
                      label="Company Name"
                      :input-max-length="50"
                      placeholder="Your business name"
                      autocomplete="name"
                      :disabled="!canEditSettings"
                    />

                    <TheInput
                      v-model="form.email"
                      label="Email"
                      :input-max-length="50"
                      placeholder="name@company.com"
                      autocomplete="email"
                      :disabled="!canEditSettings"
                    />

                    <TheInput
                      v-model="form.phone"
                      label="Phone"
                      :input-max-length="20"
                      placeholder="+44..."
                      autocomplete="tel"
                      :disabled="!canEditSettings"
                    />

                    <TheInput
                      v-model="form.invoicePrefix"
                      :input-max-length="50"
                      label="Invoice prefix"
                      placeholder="INV-"
                      :disabled="!canEditSettings"
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
                      rows="3"
                      maxlength="160"
                      class="input input-accent min-h-28 w-full resize-y rounded-lg px-3 py-2"
                      placeholder="Street&#10;City&#10;Postcode"
                      :disabled="!canEditSettings"
                    />
                  </div>
                </fieldset>
              </section>

              <!-- Invoice defaults -->
              <section :class="cardClass">
                <fieldset :disabled="!canEditSettings">
                  <div class="mb-4 flex w-full items-center gap-4">
                    <div :class="iconWrapClass">
                      <DocumentTextIcon class="size-5" />
                    </div>
                    <h2 class="font-semibold tracking-wide text-zinc-900 dark:text-zinc-100">
                      Invoice defaults
                    </h2>
                  </div>

                  <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                    <TheDropdown
                      v-model="form.currency"
                      :options="currencyOptions"
                      select-title="Currency"
                      :disabled="!canEditSettings"
                    />

                    <TheDropdown
                      v-model="form.dateFormat"
                      :options="dateFormatOptions"
                      select-title="Date format"
                      :disabled="!canEditSettings"
                    />
                  </div>

                  <div
                    class="mt-4 rounded-xl border border-zinc-300 bg-white p-4 dark:border-zinc-700 dark:bg-zinc-950/50"
                    :class="!canEditStartingInvoiceNumber ? 'opacity-70' : ''"
                  >
                    <div class="mb-2 flex items-center justify-between gap-2">
                      <div class="text-sm font-medium text-zinc-900 dark:text-zinc-100">
                        Starting invoice number
                      </div>
                      <TheTooltip
                        side="top"
                        max-width-class="max-w-[320px]"
                      >
                        <template #content>
                          <span>{{ startingInvoiceLockedMessage }}</span>
                        </template>
                        <button
                          type="button"
                          class="inline-flex cursor-help text-zinc-600 transition hover:text-sky-600 dark:text-zinc-400 dark:hover:text-emerald-400"
                          aria-label="Starting invoice number help"
                        >
                          <InformationCircleIcon class="size-5 cursor-help" />
                        </button>
                      </TheTooltip>
                    </div>

                    <TheInput
                      v-model="startingInvoiceNumberModel"
                      type="number"
                      labelHidden
                      :reserveErrorSpace="false"
                      placeholder="1"
                      inputClass="w-full"
                      :disabled="!canEditSettings || !canEditStartingInvoiceNumber"
                    />
                  </div>

                  <label
                    class="group mt-4 flex cursor-pointer items-start justify-between gap-4 rounded-xl border border-zinc-300 bg-white px-4 py-3 transition hover:border-sky-600 dark:border-zinc-700 dark:bg-zinc-950/50 dark:hover:border-emerald-400/50 dark:hover:bg-zinc-950/10"
                  >
                    <div class="min-w-0">
                      <div class="flex flex-wrap items-center gap-2">
                        <span
                          class="text-sm font-medium text-zinc-900 transition group-hover:text-sky-600 dark:text-zinc-100 group-hover:dark:text-emerald-400"
                        >
                          Show item type headers
                        </span>
                      </div>

                      <p class="mt-1 text-xs leading-relaxed text-zinc-600 dark:text-zinc-300">
                        Add section labels over Styles, Samples, and Other Items in the invoice.
                      </p>
                    </div>

                    <span class="relative mt-0.5 shrink-0">
                      <input
                        v-model="form.showItemTypeHeaders"
                        type="checkbox"
                        class="peer sr-only"
                        :disabled="!canEditSettings"
                      />
                      <span
                        class="block h-6 w-11 rounded-full border border-zinc-300 bg-zinc-200 transition peer-checked:border-sky-600 peer-checked:bg-sky-600/50 dark:border-zinc-600 dark:bg-zinc-700 dark:peer-checked:border-emerald-400 dark:peer-checked:bg-emerald-400/50"
                      />
                      <span
                        class="pointer-events-none absolute top-0.5 left-0.5 h-5 w-5 rounded-full bg-white shadow-sm transition peer-checked:translate-x-5"
                      />
                    </span>
                  </label>
                </fieldset>
              </section>

              <!-- Payment -->
              <section :class="cardClass">
                <fieldset :disabled="!canEditSettings">
                  <div class="mb-4 flex items-center gap-4">
                    <div :class="iconWrapClass">
                      <BanknotesIcon class="size-5" />
                    </div>
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
                        class="input input-accent w-full resize-y rounded-lg px-3 py-2"
                        placeholder="Payment terms shown on invoices"
                        :disabled="!canEditSettings"
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
                        class="input input-accent w-full resize-y rounded-lg px-3 py-2"
                        placeholder="Bank transfer details, sort code, account number, IBAN, etc."
                        :disabled="!canEditSettings"
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
                        class="input input-accent w-full resize-y rounded-lg px-3 py-2"
                        placeholder="Optional footer or thank-you note"
                        :disabled="!canEditSettings"
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
                </fieldset>
              </section>
            </div>
          </div>

          <!-- Right -->
          <div class="min-h-0 overflow-y-auto p-3 sm:p-5">
            <div class="space-y-5 pb-10">
              <!-- Logo -->
              <section :class="cardClass">
                <div class="mb-4 flex items-center gap-2">
                  <div :class="iconWrapClass">
                    <PhotoIcon class="size-5" />
                  </div>
                  <h2 class="font-semibold tracking-wide text-zinc-900 dark:text-zinc-100">
                    Invoice logo
                  </h2>
                </div>

                <div
                  class="grid min-h-56 place-items-center rounded-2xl border border-dashed border-zinc-300 bg-white p-4 dark:border-zinc-700 dark:bg-zinc-950/40"
                >
                  <div
                    v-if="hasLogo"
                    class="flex w-full flex-col items-center gap-4"
                  >
                    <img
                      :src="logoPreview!"
                      alt="Invoice logo preview"
                      class="max-h-36 rounded-lg object-contain"
                    />

                    <div class="flex flex-wrap justify-center gap-2">
                      <template v-if="canEditSettings">
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
                      </template>
                      <p
                        v-else
                        class="text-xs text-zinc-600 dark:text-zinc-400"
                      >
                        {{ readOnlyNotice }}
                      </p>
                    </div>
                  </div>

                  <div
                    v-else
                    class="flex flex-col items-center gap-3 text-center"
                  >
                    <div
                      class="grid size-14 place-items-center rounded-2xl border border-zinc-300 bg-zinc-50 dark:border-zinc-800 dark:bg-zinc-900"
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
                      v-if="canEditSettings"
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
                    <p
                      v-else
                      class="text-xs text-zinc-600 dark:text-zinc-400"
                    >
                      {{ readOnlyNotice }}
                    </p>
                  </div>
                </div>
              </section>

              <!-- Preview Pannel -->
              <section :class="cardClass">
                <div class="mb-4 flex items-center gap-2">
                  <div :class="iconWrapClass">
                    <SwatchIcon class="size-5" />
                  </div>
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
          class="border-t border-zinc-300 bg-white/90 px-4 py-3 sm:px-10 sm:py-4 dark:border-zinc-800 dark:bg-zinc-900/90"
        >
          <div class="flex flex-wrap items-center justify-end gap-2">
            <TheButton
              type="button"
              @click="closeSettings"
              :variant="canEditSettings ? 'danger' : 'secondary'"
              class="cursor-pointer"
            >
              {{ canEditSettings ? 'Cancel' : 'Close' }}
            </TheButton>
            <TheButton
              v-if="canEditSettings"
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

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { Component } from 'vue'
import ProductsEditor from '@/components/items/ProductsEditor.vue'
import TeamQuickMenu from '@/components/team/TeamQuickMenu.vue'
import PlatformAccessQuickMenu from '@/components/admin/PlatformAccessQuickMenu.vue'
import { useClientStore } from '@/stores/clients'
import { useAuthStore } from '@/stores/auth'
import { useProductStore } from '@/stores/products'
import { useSettingsStore } from '@/stores/settings'
import { useEscape, useShortcuts, type ShortcutDefinition } from '@/composables/keyHandlers'
import { useTheme } from '@/composables/theme'
import { useRouter } from 'vue-router'
import { emitToastError, emitToastInfo } from '@/utils/toast'
import { getApiErrorMessage, isApiError } from '@/utils/apiErrors'
import { formatTimeRemaining } from '@/utils/duration'
import { onClickOutside, useFavicon, useNow } from '@vueuse/core'
import TheDropdown from './TheDropdown.vue'
import TheButton from './TheButton.vue'
import TheSettings from './TheSettings.vue'
import UserAvatar from './UserAvatar.vue'
import lightFavi from '@/assets/lightFavi.svg'
import darkFavi from '@/assets/darkFavi.svg'

import {
  ArrowRightEndOnRectangleIcon,
  BanknotesIcon,
  BriefcaseIcon,
  ChevronDownIcon,
  ChevronUpDownIcon,
  Cog6ToothIcon,
  KeyIcon,
  MoonIcon,
  SunIcon,
  UserIcon,
  UsersIcon,
} from '@heroicons/vue/24/outline'
import { requestConfirmation } from '@/utils/confirm'

type SettingsController = {
  openSettings: () => Promise<void>
}

type TeamMenuController = {
  openMenu: () => void
}

type PlatformAccessMenuController = {
  openMenu: () => void
}

type ActionItem = {
  key: string
  label: string
  detail: string
  shortcut?: string
  icon: Component
  onSelect: () => void
}

const router = useRouter()
const clientStore = useClientStore()
const authStore = useAuthStore()
const productStore = useProductStore()
const settingsStore = useSettingsStore()
const { mode } = useTheme()
const favicon = computed(() => (mode.value === 'light' ? lightFavi : darkFavi))
const now = useNow({ interval: 60_000 })
useFavicon(favicon, {
  rel: 'icon',
})

const open = ref(false)
const menuRef = ref<HTMLElement | null>(null)
const settingsRef = ref<SettingsController | null>(null)
const teamMenuRef = ref<TeamMenuController | null>(null)
const platformAccessMenuRef = ref<PlatformAccessMenuController | null>(null)

const shortcuts: ShortcutDefinition[] = [
  { key: 'i', modifiers: ['ctrl'], action: openProducts },
  {
    key: 'm',
    modifiers: ['ctrl', 'shift'],
    action: toggleTheme,
  },
  {
    key: 'e',
    modifiers: ['ctrl', 'shift'],
    action: () => {
      openTeamMenu()
    },
  },
  {
    key: 's',
    modifiers: ['ctrl', 'shift'],
    action: () => {
      open.value = true
    },
  },
  {
    key: 'x',
    modifiers: ['ctrl', 'shift'],
    action: openPlatformAccessMenu,
  },
]
useShortcuts(shortcuts)

const userName = computed(() => authStore.user?.name?.trim() || authStore.user?.email || 'Account')
const userEmail = computed(() => authStore.user?.email?.trim() || 'Signed in')
const workspaceName = computed(() => authStore.account?.name?.trim() || 'Workspace')
const hasWorkspaceAccess = computed(() => authStore.hasBillingAccess)
const currentClientName = computed(() => {
  if (clientStore.selectedClient?.name) return clientStore.selectedClient.name
  return clientStore.hasClients ? 'Select client' : 'No clients yet'
})
const roleLabel = computed(() => (authStore.user?.role === 'owner' ? 'Admin' : 'Member'))
const promoStatusLabel = computed(() => {
  const billing = authStore.billing
  if (
    billing?.accessSource !== 'promo' ||
    billing.accessGranted !== true ||
    !billing.accessExpiresAt
  ) {
    return ''
  }

  const remaining = formatTimeRemaining(billing.accessExpiresAt, now.value)
  return remaining ? `promo period · ${remaining} remaining` : ''
})
const currentThemeIcon = computed<Component>(() => (mode.value === 'dark' ? MoonIcon : SunIcon))
const currentThemeDetail = computed(() =>
  mode.value === 'dark' ? 'Currently dark' : 'Currently light',
)

const actionItems = computed<ActionItem[]>(() => {
  const items: ActionItem[] = [
    {
      key: 'theme',
      label: 'Theme',
      detail: currentThemeDetail.value,
      shortcut: 'Ctrl+Shift+M',
      icon: currentThemeIcon.value,
      onSelect: toggleTheme,
    },
  ]

  if (!hasWorkspaceAccess.value) {
    items.push({
      key: 'billing',
      label: 'Billing',
      detail: authStore.canManageBilling
        ? 'Activate workspace access'
        : 'Waiting for admin payment',
      icon: BanknotesIcon,
      onSelect: openBilling,
    })
  }

  if (hasWorkspaceAccess.value) {
    items.push(
      {
        key: 'items',
        label: 'Items',
        detail: 'Styles and samples',
        shortcut: 'Ctrl+I',
        icon: BriefcaseIcon,
        onSelect: openProducts,
      },
      {
        key: 'settings',
        label: 'Settings',
        detail: settingsStore.needsSetup
          ? 'Finish workspace setup'
          : 'Invoice defaults and branding',
        shortcut: 'Alt+Shift+S',
        icon: Cog6ToothIcon,
        onSelect: openSettings,
      },
    )
  }

  if (authStore.isOwner) {
    items.splice(hasWorkspaceAccess.value ? 2 : items.length, 0, {
      key: 'team',
      label: 'Team',
      detail: hasWorkspaceAccess.value
        ? authStore.billing?.plan === 'team'
          ? 'Members, invites, and workspace controls'
          : 'Upgrade to the team plan for extra seats'
        : 'Billing and workspace controls',
      shortcut: 'Ctrl+Shift+E',
      icon: UsersIcon,
      onSelect: openTeamMenu,
    })
  }

  if (authStore.canManagePlatformAccess) {
    items.push({
      key: 'platform-access',
      label: 'Platform access',
      detail: 'Trusted access, promo codes, and team-tier test grants',
      shortcut: 'Ctrl+Shift+X',
      icon: KeyIcon,
      onSelect: openPlatformAccessMenu,
    })
  }

  return items
})

function closeMenu() {
  open.value = false
}

function toggleMenu() {
  open.value = !open.value
}

function toggleTheme() {
  mode.value = mode.value === 'light' ? 'dark' : 'light'
  closeMenu()
}

function openProducts() {
  if (!requireWorkspaceAccess()) return
  productStore.open = true
  closeMenu()
}

function openTeamMenu() {
  if (!authStore.isOwner) return
  closeMenu()
  teamMenuRef.value?.openMenu()
}

function openPlatformAccessMenu() {
  if (!authStore.canManagePlatformAccess) return
  closeMenu()
  platformAccessMenuRef.value?.openMenu()
}

function openSettings() {
  if (!requireWorkspaceAccess()) return
  closeMenu()
  void settingsRef.value?.openSettings()
}

function openBilling() {
  closeMenu()
  void router.push({ name: 'billing' })
}

function requireWorkspaceAccess() {
  if (authStore.hasBillingAccess) {
    return true
  }

  closeMenu()
  emitToastInfo(
    authStore.canManageBilling
      ? 'Activate billing to use clients, items, team, and settings.'
      : 'The workspace admin needs to reactivate billing before workspace tools are available.',
    { title: 'Workspace locked' },
  )
  void router.push({ name: 'billing' })
  return false
}

async function signOut() {
  try {
    const confirmed = await requestConfirmation({
      title: 'Sign out?',
      message: 'Are you sure you want to sign out?',
      details: "This action will redirect you to the app's landing page.",
      confirmLabel: 'Sign out',
      cancelLabel: 'Cancel',
      confirmVariant: 'danger',
    })

    if (!confirmed) return
    closeMenu()
    await authStore.logout()
    await router.push({ name: 'login' })
  } catch (err) {
    emitToastError({
      title: 'Could not sign out',
      message: isApiError(err) ? getApiErrorMessage(err) : 'Please try again in a moment.',
    })
  }
}

useEscape(closeMenu, {
  enabled: () => open.value,
})

onClickOutside(menuRef, closeMenu)
</script>

<template>
  <div
    ref="menuRef"
    class="fixed top-3 right-4 z-50"
  >
    <button
      type="button"
      :aria-expanded="open"
      aria-haspopup="menu"
      class="group flex items-center gap-4 rounded-2xl border border-zinc-300 bg-white/90 p-2 font-medium text-zinc-600 shadow-lg transition duration-200 hover:border-sky-600/50 hover:bg-white hover:text-zinc-900 dark:border-zinc-800 dark:bg-zinc-900/90 dark:text-zinc-300 dark:hover:border-emerald-400/25 dark:hover:bg-zinc-900/70 dark:hover:text-zinc-100"
      @click="toggleMenu"
    >
      <UserAvatar
        :name="authStore.user?.name"
        :email="authStore.user?.email"
        :avatar-url="authStore.user?.avatarUrl"
        class="size-8 rounded-xl"
      />
      <div class="text-start text-sm">
        <p>{{ authStore.user?.name }}</p>
      </div>

      <ChevronDownIcon
        class="size-4 shrink-0 text-zinc-500 transition group-hover:text-sky-700 dark:text-zinc-400 dark:group-hover:text-emerald-300"
      />
    </button>

    <transition name="fade-down-up">
      <section
        v-if="open"
        class="absolute right-0 mt-3 w-[min(22rem,calc(100vw-2rem))] rounded-2xl border border-zinc-300 bg-white shadow-2xl dark:border-zinc-800 dark:bg-zinc-900"
      >
        <header class="border-b border-zinc-300 px-4 py-3 dark:border-zinc-800">
          <div class="flex items-start gap-3">
            <UserAvatar
              :name="authStore.user?.name"
              :email="authStore.user?.email"
              :avatar-url="authStore.user?.avatarUrl"
              class="size-11 rounded-2xl"
            />

            <div class="min-w-0 flex-1">
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0 flex-1">
                  <h2 class="truncate text-sm font-semibold text-zinc-950 dark:text-zinc-50">
                    {{ userName }}
                  </h2>
                  <p class="truncate text-sm text-zinc-600 dark:text-zinc-400">
                    {{ userEmail }}
                  </p>
                </div>
                <span
                  class="rounded-full border border-zinc-300 px-2 py-0.5 text-[11px] font-medium text-zinc-600 dark:border-zinc-700 dark:text-zinc-300"
                >
                  {{ roleLabel }}
                </span>
              </div>

              <p
                v-if="promoStatusLabel"
                class="text-mini mt-1 text-emerald-700 dark:text-emerald-400"
              >
                {{ promoStatusLabel }}
              </p>
            </div>
          </div>
        </header>

        <div class="space-y-4 p-4">
          <section>
            <div
              class="mb-2 text-[11px] font-semibold tracking-[0.16em] text-zinc-500 uppercase dark:text-zinc-400"
            >
              Client Picker
            </div>

            <TheDropdown
              v-model="clientStore.selectedClient"
              :options="clientStore.clients"
              :disabled="!hasWorkspaceAccess || !clientStore.hasClients"
              :left-icon="UserIcon"
              :right-icon="ChevronUpDownIcon"
              label-key="name"
              value-key="id"
              :placeholder="
                !hasWorkspaceAccess
                  ? 'Billing required'
                  : clientStore.isLoading
                    ? 'Loading clients...'
                    : clientStore.hasClients
                      ? 'Select client'
                      : 'No clients yet'
              "
              input-class="py-2.5 text-sm"
            />

            <p
              v-if="!hasWorkspaceAccess"
              class="mt-2 text-xs text-zinc-500 dark:text-zinc-400"
            >
              Billing needs to be active before client and workspace tools are available.
            </p>

            <p
              v-else-if="!clientStore.hasClients"
              class="mt-2 text-xs text-zinc-500 dark:text-zinc-400"
            >
              Add a client from the clients screen to start working.
            </p>
          </section>

          <section>
            <div
              class="mb-2 text-[11px] font-semibold tracking-[0.16em] text-zinc-500 uppercase dark:text-zinc-400"
            >
              Quick actions
            </div>

            <div class="grid grid-cols-1 gap-2">
              <button
                v-for="action in actionItems"
                :key="action.key"
                type="button"
                class="group flex items-start gap-3 rounded-xl border border-zinc-300 bg-zinc-50/70 p-3 text-left transition hover:border-sky-600/50 hover:bg-sky-50 dark:border-zinc-800 dark:bg-zinc-950/40 dark:hover:border-emerald-400/25 dark:hover:bg-emerald-950/20"
                @click="action.onSelect"
              >
                <div
                  class="mt-0.5 grid size-9 shrink-0 place-items-center rounded-lg border border-zinc-300 bg-white text-zinc-700 transition group-hover:text-sky-700 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200 dark:group-hover:text-emerald-300"
                >
                  <component
                    :is="action.icon"
                    class="size-4.5"
                  />
                </div>

                <div class="min-w-0 flex-1">
                  <div class="flex items-center justify-between gap-2">
                    <span class="truncate text-sm font-medium text-zinc-900 dark:text-zinc-100">
                      {{ action.label }}
                    </span>
                    <span
                      v-if="action.shortcut"
                      class="text-mini hidden shrink-0 text-zinc-400 sm:block dark:text-zinc-500"
                    >
                      {{ action.shortcut }}
                    </span>
                  </div>
                  <p class="mt-1 text-xs text-zinc-600 dark:text-zinc-400">
                    {{ action.detail }}
                  </p>
                </div>
              </button>
            </div>
          </section>

          <TheButton
            variant="secondary"
            class="w-full justify-center"
            @click="signOut"
          >
            <ArrowRightEndOnRectangleIcon class="size-4" />
            Sign out
          </TheButton>
          <div class="hidden text-xs text-zinc-500 sm:block dark:text-zinc-400">
            Ctr+Shift+S to open this menu
          </div>
        </div>
      </section>
    </transition>

    <ProductsEditor :show-trigger="false" />
    <TeamQuickMenu
      ref="teamMenuRef"
      :show-trigger="false"
    />
    <PlatformAccessQuickMenu
      ref="platformAccessMenuRef"
      :show-trigger="false"
    />
    <TheSettings
      ref="settingsRef"
      :show-trigger="false"
    />
  </div>
</template>

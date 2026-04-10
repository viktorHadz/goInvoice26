<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, watch } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import TopRightMenu from './components/UI/TopRightMenu.vue'
import NavMain from './components/UI/NavMain.vue'
import TheToast from './components/UI/TheToast.vue'
import TheConfirmDialog from './components/UI/TheConfirmDialog.vue'
import { useAuthStore } from './stores/auth'
import { useEditorStore } from './stores/editor'
import { useSettingsStore } from './stores/settings'

const route = useRoute()
const authStore = useAuthStore()
const editorStore = useEditorStore()
const settingsStore = useSettingsStore()
const showAppChrome = computed(() => route.meta.appChrome === true)
const routeTransitionName = computed(() => (showAppChrome.value ? 'page' : 'public-page'))
const shellClassName = computed(() =>
    showAppChrome.value
        ? 'flex bg-zinc-50 dark:bg-zinc-950'
        : 'bg-sky-50 dark:bg-zinc-950',
)

const SETTINGS_SYNC_INTERVAL_MS = 30_000
let settingsSyncTimer: number | null = null

function clearSettingsSyncTimer() {
    if (settingsSyncTimer != null) {
        window.clearInterval(settingsSyncTimer)
        settingsSyncTimer = null
    }
}

async function refreshSettingsInBackground() {
    if (!showAppChrome.value || !authStore.isAuthenticated || !authStore.hasBillingAccess) return
    if (!settingsStore.hasSettings) return

    try {
        await settingsStore.fetchSettings({ background: true })
    } catch {
        // Silent background refresh; foreground routes already handle fetch failures with toasts.
    }
}

function restartSettingsSync() {
    clearSettingsSyncTimer()

    if (!showAppChrome.value || !authStore.isAuthenticated || !authStore.hasBillingAccess) {
        return
    }

    settingsSyncTimer = window.setInterval(() => {
        void refreshSettingsInBackground()
    }, SETTINGS_SYNC_INTERVAL_MS)
}

function handleWindowFocus() {
    void refreshSettingsInBackground()
}

function handleVisibilityChange() {
    if (document.visibilityState === 'visible') {
        void refreshSettingsInBackground()
    }
}

function handleBeforeUnload(event: BeforeUnloadEvent) {
    if (!editorStore.hasUnsavedChanges) return
    event.preventDefault()
    event.returnValue = ''
}

onMounted(() => {
    window.addEventListener('focus', handleWindowFocus)
    window.addEventListener('beforeunload', handleBeforeUnload)
    document.addEventListener('visibilitychange', handleVisibilityChange)
    restartSettingsSync()
})

onBeforeUnmount(() => {
    clearSettingsSyncTimer()
    window.removeEventListener('beforeunload', handleBeforeUnload)
    window.removeEventListener('focus', handleWindowFocus)
    document.removeEventListener('visibilitychange', handleVisibilityChange)
})

watch(
    [showAppChrome, () => authStore.isAuthenticated, () => authStore.hasBillingAccess],
    () => restartSettingsSync(),
    { immediate: true },
)
</script>

<template>
    <div
        :class="[
            'min-h-screen w-full text-zinc-900 dark:text-zinc-100',
            shellClassName,
        ]"
    >
        <main class="relative min-h-screen w-full">
            <div :class="showAppChrome ? 'mt-26 px-4 pb-16 sm:py-8 sm:pb-8 md:px-6' : ''">
                <RouterView v-slot="{ Component, route }">
                    <Transition
                        :name="routeTransitionName"
                        mode="out-in"
                        appear
                    >
                        <component
                            :is="Component"
                            :key="route.fullPath"
                        />
                    </Transition>
                </RouterView>
            </div>

            <template v-if="showAppChrome">
                <NavMain />
                <TopRightMenu />
            </template>
            <TheToast />
            <TheConfirmDialog />
        </main>
    </div>
</template>

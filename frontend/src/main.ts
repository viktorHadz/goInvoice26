import '@/assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { useClientStore } from '@/stores/clients'
import { useSettingsStore } from '@/stores/settings'
import { emitToastError } from '@/utils/toast'
import { getApiErrorMessage, isApiError } from '@/utils/apiErrors'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)

const clientStore = useClientStore()
const settingsStore = useSettingsStore()

const startupErrors: string[] = []
const startupTasks: Array<Promise<unknown>> = [clientStore.load(), settingsStore.fetchSettings()]

const startupResults = await Promise.allSettled(startupTasks)
for (const [idx, result] of startupResults.entries()) {
    if (result.status === 'fulfilled') continue
    const defaultMessage =
        idx === 0
            ? 'Could not load clients during startup.'
            : 'Could not load settings during startup.'
    startupErrors.push(
        isApiError(result.reason) ? getApiErrorMessage(result.reason) : defaultMessage,
    )
}

app.use(router)
app.mount('#app')

for (const message of startupErrors) {
    emitToastError({
        title: 'Startup problem',
        message,
    })
}

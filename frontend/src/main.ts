import '@/assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { useClientStore } from '@/stores/clients'
import { useSettingsStore } from '@/stores/settings'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)

const clientStore = useClientStore()
await clientStore.load()

const settingsStore = useSettingsStore()
await settingsStore.fetchSettings()

app.use(router)
app.mount('#app')

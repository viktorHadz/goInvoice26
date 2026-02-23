import '@/assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { useClientStore } from '@/stores/clients'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
const clientStore = useClientStore()
await clientStore.load()

app.use(router)

app.mount('#app')

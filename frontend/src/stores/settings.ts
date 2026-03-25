import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { request } from '@/utils/fetchHelper'

export type CurrencyCode = 'GBP' | 'EUR' | 'USD'
export type DateFormat = 'dd/mm/yyyy' | 'mm/dd/yyyy' | 'yyyy-mm-dd'
// TODO: add custom coluumn in clients for company number
// TODO: Polylang SUPPORT
export type Settings = {
    companyName: string
    email: string
    phone: string
    companyAddress: string
    invoicePrefix: string
    currency: CurrencyCode
    dateFormat: DateFormat
    paymentTerms: string
    paymentDetails: string
    notesFooter: string
    logoUrl: string
    showItemTypeHeaders: boolean
}

function normalizeSettings(data: Settings): Settings {
    return {
        ...data,
        showItemTypeHeaders: data.showItemTypeHeaders !== false,
    }
}

function isSettingsComplete(s: Settings): boolean {
    return (
        s.companyName.trim().length > 0 &&
        s.invoicePrefix.trim().length > 0 &&
        s.currency.trim().length > 0 &&
        s.dateFormat.trim().length > 0
    )
}

export const useSettingsStore = defineStore('settings', () => {
    const settings = ref<Settings | null>(null)
    const isLoading = ref(false)
    const needsSetup = ref(false)

    const hasSettings = computed(() => settings.value !== null)

    async function fetchSettings() {
        isLoading.value = true
        try {
            const data = await request<Settings>('/api/settings')
            const normalized = normalizeSettings(data)
            settings.value = normalized
            needsSetup.value = !isSettingsComplete(normalized)

            return normalized
        } finally {
            isLoading.value = false
        }
    }

    async function saveSettings(payload: Settings) {
        const data = await request<Settings>('/api/settings', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
        })
        const normalized = normalizeSettings(data)
        settings.value = normalized
        needsSetup.value = !isSettingsComplete(normalized)

        return normalized
    }

    return {
        settings,
        isLoading,
        needsSetup,
        hasSettings,
        fetchSettings,
        saveSettings,
    }
})

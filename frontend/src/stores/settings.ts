import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

export type CurrencyCode = 'GBP' | 'EUR' | 'USD'
export type DateFormat = 'dd/mm/yyyy' | 'mm/dd/yyyy' | 'yyyy-mm-dd'

export type Settings = {
    companyName: string
    email: string
    phone: string
    companyAddress: string
    invoicePrefix: string
    currency: CurrencyCode
    dateFormat: DateFormat
    customItemsPrefix: string
    paymentTerms: string
    paymentDetails: string
    notesFooter: string
    logoUrl: string
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
            const res = await fetch('/api/settings')
            if (!res.ok) throw new Error(`Failed to fetch settings (${res.status})`)

            const data = (await res.json()) as Settings
            settings.value = data
            needsSetup.value = !isSettingsComplete(data)

            return data
        } finally {
            isLoading.value = false
        }
    }

    async function saveSettings(payload: Settings) {
        const res = await fetch('/api/settings', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
        })

        if (!res.ok) {
            throw new Error(`Failed to save settings (${res.status})`)
        }

        const data = (await res.json()) as Settings
        settings.value = data
        needsSetup.value = !isSettingsComplete(data)

        return data
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

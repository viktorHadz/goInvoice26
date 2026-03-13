import { handleImageUpload, type UploadLogoResponse } from '@/utils/imageHandler'
import { defineStore } from 'pinia'
import { reactive } from 'vue'

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

function createDefaultSettings(): Settings {
    return {
        companyName: '',
        email: '',
        phone: '',
        companyAddress: '',
        invoicePrefix: 'INV-',
        currency: 'GBP',
        dateFormat: 'dd/mm/yyyy',
        customItemsPrefix: 'custom',
        paymentTerms: 'Please make payment within 14 days.',
        paymentDetails: '',
        notesFooter: '',
        logoUrl: '',
    }
}

export const useSettingsStore = defineStore('settings', () => {
    const settings = reactive<Settings>(createDefaultSettings())

    function setSettings(payload: Settings) {
        Object.assign(settings, payload)
    }

    function resetSettings() {
        Object.assign(settings, createDefaultSettings())
    }

    function uploadLogo(file: File): Promise<UploadLogoResponse> {
        return handleImageUpload(file)
    }

    return {
        settings,
        setSettings,
        resetSettings,
        uploadLogo,
    }
})

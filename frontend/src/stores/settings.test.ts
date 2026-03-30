import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useSettingsStore } from '@/stores/settings'

const { requestMock } = vi.hoisted(() => ({
    requestMock: vi.fn(),
}))

vi.mock('@/utils/fetchHelper', () => ({
    request: requestMock,
}))

describe('settings store', () => {
    beforeEach(() => {
        setActivePinia(createPinia())
        vi.clearAllMocks()
    })

    it('normalizes starting invoice number state from the server', async () => {
        requestMock.mockResolvedValueOnce({
            companyName: 'Acme',
            email: '',
            phone: '',
            companyAddress: '',
            invoicePrefix: 'INV-',
            currency: 'GBP',
            dateFormat: 'dd/mm/yyyy',
            paymentTerms: '',
            paymentDetails: '',
            notesFooter: '',
            logoUrl: '',
            showItemTypeHeaders: true,
            startingInvoiceNumber: 120,
            canEditStartingInvoiceNumber: false,
        })

        const store = useSettingsStore()
        const settings = await store.fetchSettings()

        expect(settings.startingInvoiceNumber).toBe(120)
        expect(settings.canEditStartingInvoiceNumber).toBe(false)
    })

    it('saves allocator-backed settings fields through the shared payload', async () => {
        requestMock.mockResolvedValueOnce({
            companyName: 'Acme',
            email: '',
            phone: '',
            companyAddress: '',
            invoicePrefix: 'INV-',
            currency: 'GBP',
            dateFormat: 'dd/mm/yyyy',
            paymentTerms: '',
            paymentDetails: '',
            notesFooter: '',
            logoUrl: '',
            showItemTypeHeaders: true,
            startingInvoiceNumber: 200,
            canEditStartingInvoiceNumber: true,
        })

        const store = useSettingsStore()
        const settings = await store.saveSettings({
            companyName: 'Acme',
            email: '',
            phone: '',
            companyAddress: '',
            invoicePrefix: 'INV-',
            currency: 'GBP',
            dateFormat: 'dd/mm/yyyy',
            paymentTerms: '',
            paymentDetails: '',
            notesFooter: '',
            showItemTypeHeaders: true,
            startingInvoiceNumber: 200,
        })

        expect(requestMock).toHaveBeenCalledWith(
            '/api/settings',
            expect.objectContaining({
                method: 'PUT',
                body: JSON.stringify({
                    companyName: 'Acme',
                    email: '',
                    phone: '',
                    companyAddress: '',
                    invoicePrefix: 'INV-',
                    currency: 'GBP',
                    dateFormat: 'dd/mm/yyyy',
                    paymentTerms: '',
                    paymentDetails: '',
                    notesFooter: '',
                    showItemTypeHeaders: true,
                    startingInvoiceNumber: 200,
                }),
            }),
        )
        expect(settings.startingInvoiceNumber).toBe(200)
        expect(settings.canEditStartingInvoiceNumber).toBe(true)
    })
})

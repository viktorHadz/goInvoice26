import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAuthStore } from '@/stores/auth'

const {
    clientResetMock,
    settingsResetMock,
    productResetMock,
    teamResetMock,
    invoiceResetMock,
    editorResetMock,
} = vi.hoisted(() => ({
    clientResetMock: vi.fn(),
    settingsResetMock: vi.fn(),
    productResetMock: vi.fn(),
    teamResetMock: vi.fn(),
    invoiceResetMock: vi.fn(),
    editorResetMock: vi.fn(),
}))

vi.mock('@/utils/fetchHelper', () => ({
    request: vi.fn(),
}))

vi.mock('@/stores/clients', () => ({
    useClientStore: () => ({
        reset: clientResetMock,
    }),
}))

vi.mock('@/stores/settings', () => ({
    useSettingsStore: () => ({
        reset: settingsResetMock,
    }),
}))

vi.mock('@/stores/products', () => ({
    useProductStore: () => ({
        reset: productResetMock,
    }),
}))

vi.mock('@/stores/team', () => ({
    useTeamStore: () => ({
        reset: teamResetMock,
    }),
}))

vi.mock('@/stores/invoice', () => ({
    useInvoiceStore: () => ({
        reset: invoiceResetMock,
    }),
}))

vi.mock('@/stores/editor', () => ({
    useEditorStore: () => ({
        reset: editorResetMock,
    }),
}))

describe('auth store', () => {
    beforeEach(() => {
        setActivePinia(createPinia())
        vi.clearAllMocks()
    })

    it('clears all workspace-scoped stores together', () => {
        const store = useAuthStore()

        store.clearWorkspaceState()

        expect(clientResetMock).toHaveBeenCalledTimes(1)
        expect(settingsResetMock).toHaveBeenCalledTimes(1)
        expect(productResetMock).toHaveBeenCalledTimes(1)
        expect(teamResetMock).toHaveBeenCalledTimes(1)
        expect(invoiceResetMock).toHaveBeenCalledTimes(1)
        expect(editorResetMock).toHaveBeenCalledTimes(1)
    })
})

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useTeamStore } from '@/stores/team'

const { requestMock } = vi.hoisted(() => ({
    requestMock: vi.fn(),
}))

vi.mock('@/utils/fetchHelper', () => ({
    request: requestMock,
}))

describe('team store', () => {
    beforeEach(() => {
        setActivePinia(createPinia())
        vi.clearAllMocks()
    })

    it('loads members and invites from the team endpoint', async () => {
        requestMock.mockResolvedValueOnce({
            members: [
                {
                    id: 1,
                    name: 'Owner Example',
                    email: 'owner@example.com',
                    avatarUrl: '',
                    role: 'owner',
                    createdAt: '2026-03-29T10:00:00Z',
                },
            ],
            invites: [
                {
                    id: 9,
                    email: 'teammate@example.com',
                    createdAt: '2026-03-29T10:00:00Z',
                },
            ],
        })

        const store = useTeamStore()
        await store.load()

        expect(requestMock).toHaveBeenCalledWith('/api/team')
        expect(store.memberCount).toBe(1)
        expect(store.inviteCount).toBe(1)
    })

    it('refreshes team data after adding an invite', async () => {
        requestMock
            .mockResolvedValueOnce({
                id: 9,
                email: 'teammate@example.com',
                createdAt: '2026-03-29T10:00:00Z',
            })
            .mockResolvedValueOnce({
                members: [],
                invites: [
                    {
                        id: 9,
                        email: 'teammate@example.com',
                        createdAt: '2026-03-29T10:00:00Z',
                    },
                ],
            })

        const store = useTeamStore()
        await store.invite('teammate@example.com')

        expect(requestMock).toHaveBeenNthCalledWith(
            1,
            '/api/team/invites',
            expect.objectContaining({
                method: 'POST',
                body: JSON.stringify({ email: 'teammate@example.com' }),
            }),
        )
        expect(requestMock).toHaveBeenNthCalledWith(2, '/api/team')
        expect(store.inviteCount).toBe(1)
    })
})

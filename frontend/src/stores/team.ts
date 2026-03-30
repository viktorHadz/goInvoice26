import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { request } from '@/utils/fetchHelper'

export type TeamRole = 'owner' | 'member'

export type TeamMember = {
    id: number
    name: string
    email: string
    avatarUrl: string
    role: TeamRole
    createdAt: string
}

export type TeamInvite = {
    id: number
    email: string
    createdAt: string
}

type TeamSummary = {
    members: TeamMember[]
    invites: TeamInvite[]
}

export const useTeamStore = defineStore('team', () => {
    const members = ref<TeamMember[]>([])
    const invites = ref<TeamInvite[]>([])
    const isLoading = ref(false)
    const hasLoaded = ref(false)

    const memberCount = computed(() => members.value.length)
    const inviteCount = computed(() => invites.value.length)

    async function load(force = false) {
        if (hasLoaded.value && !force) return

        isLoading.value = true
        try {
            const data = await request<TeamSummary>('/api/team')
            members.value = Array.isArray(data.members) ? data.members : []
            invites.value = Array.isArray(data.invites) ? data.invites : []
            hasLoaded.value = true
        } finally {
            isLoading.value = false
        }
    }

    async function invite(email: string) {
        await request<TeamInvite>('/api/team/invites', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email }),
        })
        await load(true)
    }

    async function revokeInvite(inviteId: number) {
        await request<void>(`/api/team/invites/${inviteId}`, {
            method: 'DELETE',
        })
        await load(true)
    }

    async function removeMember(memberId: number) {
        await request<void>(`/api/team/members/${memberId}`, {
            method: 'DELETE',
        })
        await load(true)
    }

    function reset() {
        members.value = []
        invites.value = []
        isLoading.value = false
        hasLoaded.value = false
    }

    return {
        members,
        invites,
        isLoading,
        hasLoaded,
        memberCount,
        inviteCount,
        load,
        invite,
        revokeInvite,
        removeMember,
        reset,
    }
})

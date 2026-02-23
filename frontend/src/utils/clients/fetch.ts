export type Client = {
  id: number
  name: string
  companyName?: string
  address?: string
  email?: string
  created_at?: string
  updated_at?: string
}

export type UpdateClientInput = Partial<Omit<Client, 'id' | 'created_at' | 'updated_at'>>

// API request helper
async function request<T>(input: RequestInfo, init?: RequestInit): Promise<T> {
  const res = await fetch(input, init)

  if (!res.ok) {
    const text = await res.text().catch(() => '')
    try {
      const data = text ? JSON.parse(text) : null
      const apiMsg = data?.error?.message
      if (apiMsg) throw new Error(apiMsg)
    } catch {
      // fall through
    }
    throw new Error(text || `Response status: ${res.status}`)
  }

  // Handle 204 No Content
  if (res.status === 204) {
    return undefined as unknown as T
  }

  // avoids JSON parse crash if server returns no body but not 204
  const text = await res.text()
  if (!text) {
    return undefined as unknown as T
  }

  return JSON.parse(text) as T
}

// Handlers - consumed by frontend
export async function getClients(): Promise<Client[]> {
  const data = await request<Client[] | null>('/api/clients')
  return Array.isArray(data) ? data : []
}

export function createNewClient(client: Omit<Client, 'id'>): Promise<Client> {
  console.log('Preparing to send client to backend', client)

  return request<Client>('/api/clients', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(client),
  })
}

export function deleteClient(id: number | string): Promise<void> {
  const cleanId = Number(id)
  if (!Number.isFinite(cleanId) || cleanId <= 0) {
    throw new Error(`Invalid client id: ${String(id)}`)
  }

  return request<void>(`/api/clients/${cleanId}`, { method: 'DELETE' })
}

export function updateClient(id: number | string, patch: UpdateClientInput): Promise<Client> {
  const cleanId = Number(id)
  if (!Number.isFinite(cleanId) || cleanId <= 0) {
    throw new Error(`Invalid client id: ${String(id)}`)
  }

  return request<Client>(`/api/clients/${cleanId}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(patch),
  })
}

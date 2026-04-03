import { request } from './fetchHelper'

export function deleteWorkspace() {
    return request<void>('/api/workspace', {
        method: 'DELETE',
    })
}

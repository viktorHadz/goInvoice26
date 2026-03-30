import type { Settings } from '@/stores/settings'
import { request } from '@/utils/fetchHelper'

const MAX_LOGO_SIZE_BYTES = 5 * 1024 * 1024
const ALLOWED_IMAGE_TYPES = new Set(['image/png', 'image/jpeg', 'image/webp'])

export function validateImageFile(file: unknown): File {
    if (!(file instanceof File)) {
        throw new Error('No file selected')
    }

    if (!ALLOWED_IMAGE_TYPES.has(file.type)) {
        throw new Error('Unsupported image type. Please upload PNG, JPG, or WebP.')
    }

    if (file.size <= 0) {
        throw new Error('Selected image is empty.')
    }

    if (file.size > MAX_LOGO_SIZE_BYTES) {
        throw new Error('Image is too large. Maximum size is 5MB.')
    }

    return file
}

export function readImagePreview(file: File): Promise<string> {
    validateImageFile(file)

    return new Promise((resolve, reject) => {
        const reader = new FileReader()

        reader.onload = () => {
            if (typeof reader.result !== 'string') {
                reject(new Error('Failed to read selected image.'))
                return
            }

            resolve(reader.result)
        }

        reader.onerror = () => {
            reject(new Error('Failed to read selected image.'))
        }

        reader.readAsDataURL(file)
    })
}

export async function uploadLogo(file: File): Promise<Settings> {
    const validFile = validateImageFile(file)

    const formData = new FormData()
    formData.append('user_logo', validFile)

    return request<Settings>('/api/settings/logo', {
        method: 'PUT',
        body: formData,
    })
}

export async function deleteLogo(): Promise<Settings> {
    return request<Settings>('/api/settings/logo', {
        method: 'DELETE',
    })
}

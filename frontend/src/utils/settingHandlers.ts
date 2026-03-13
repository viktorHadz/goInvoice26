export type UploadLogoResponse = {
    logoUrl: string
}

const MAX_LOGO_SIZE_BYTES = 5 * 1024 * 1024
const ALLOWED_IMAGE_TYPES = new Set(['image/png', 'image/jpeg', 'image/webp'])

function getErrorMessage(data: unknown, fallback: string): string {
    if (
        typeof data === 'object' &&
        data !== null &&
        'message' in data &&
        typeof (data as { message?: unknown }).message === 'string'
    ) {
        return (data as { message: string }).message
    }

    return fallback
}

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

export async function handleImageUpload(file: File): Promise<UploadLogoResponse> {
    const validFile = validateImageFile(file)

    const formData = new FormData()
    formData.append('user_logo', validFile)

    const response = await fetch('/api/image', {
        method: 'POST',
        body: formData,
    })

    let data: unknown = null

    try {
        data = await response.json()
    } catch {
        // ignore
    }

    if (!response.ok) {
        throw new Error(getErrorMessage(data, `Image upload failed with status ${response.status}`))
    }

    if (
        typeof data !== 'object' ||
        data === null ||
        !('logoUrl' in data) ||
        typeof (data as { logoUrl?: unknown }).logoUrl !== 'string'
    ) {
        throw new Error('Server returned an invalid upload response')
    }

    return data as UploadLogoResponse
}

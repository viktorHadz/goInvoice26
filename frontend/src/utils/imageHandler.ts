export type UploadLogoResponse = {
    logoUrl: string
}

const MAX_LOGO_SIZE_BYTES = 5 * 1024 * 1024
const ALLOWED_IMAGE_TYPES = ['image/png', 'image/jpeg', 'image/webp']

export async function handleImageUpload(file: File): Promise<UploadLogoResponse> {
    if (!(file instanceof File)) {
        throw new Error('No file selected')
    }

    if (!ALLOWED_IMAGE_TYPES.includes(file.type)) {
        throw new Error('Unsupported image type. Please upload PNG, JPG, or WebP.')
    }

    if (file.size > MAX_LOGO_SIZE_BYTES) {
        throw new Error('Image is too large. Maximum size is 5MB.')
    }

    const formData = new FormData()
    formData.append('user_logo', file)

    const response = await fetch('/api/image', {
        method: 'POST',
        body: formData,
    })

    let data: unknown = null

    try {
        data = await response.json()
    } catch {
        // throwing below
    }

    if (!response.ok) {
        const message =
            typeof data === 'object' &&
            data !== null &&
            'message' in data &&
            typeof (data as { message?: unknown }).message === 'string'
                ? (data as { message: string }).message
                : `Image upload failed with status ${response.status}`

        throw new Error(message)
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

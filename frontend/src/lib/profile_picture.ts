export interface ProfilePicture {
    presignedUrl: string
}

export function normalizeProfilePictureURL(value: unknown) {
    if (typeof value !== "string") return null

    try {
        const url = new URL(value)
        if (url.protocol !== "https:" && url.protocol !== "http:") return null
    } catch {
        return null
    }

    return value
}

export function normalizeProfilePicture(value: unknown): ProfilePicture | null {
    if (!value || typeof value !== "object" || !("presignedUrl" in value)) return null
    const presignedUrl = normalizeProfilePictureURL(value.presignedUrl)
    if (!presignedUrl) return null

    return {presignedUrl}
}

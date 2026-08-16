export interface PostAttachment {
    id: string
    presignedUrl: string
    fileName: string
    contentType: string
}

export function normalizePostAttachments(value: unknown): PostAttachment[] {
    if (!Array.isArray(value)) return []

    return value.flatMap((attachment) => {
        if (!attachment || typeof attachment !== "object") return []

        const id = "id" in attachment && typeof attachment.id === "string" ? attachment.id : ""
        const presignedUrl = "presignedUrl" in attachment && typeof attachment.presignedUrl === "string"
            ? attachment.presignedUrl
            : ""
        const fileName = "fileName" in attachment && typeof attachment.fileName === "string"
            ? attachment.fileName.trim()
            : ""
        const contentType = "contentType" in attachment && typeof attachment.contentType === "string"
            ? attachment.contentType.trim().toLocaleLowerCase()
            : ""

        if (!/^\d+$/.test(id) || !fileName || fileName.length > 255 || !contentType || contentType.length > 255) {
            return []
        }

        try {
            const url = new URL(presignedUrl)
            if (url.protocol !== "https:" && url.protocol !== "http:") return []
        } catch {
            return []
        }

        return [{id, presignedUrl, fileName, contentType}]
    })
}

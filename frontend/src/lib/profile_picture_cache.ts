const PROFILE_PICTURE_CACHE_TTL = 5 * 60 * 1000
const SIGNED_URL_EXPIRY_BUFFER = 30 * 1000
const STORAGE_PREFIX = "edulink:profile-picture:"

interface ProfilePictureCacheEntry {
    resource: string
    url: string
    expiresAt: number
}

const memoryCache = new Map<string, ProfilePictureCacheEntry>()

function resourceFromURL(value: string) {
    try {
        const url = new URL(value)
        if (url.protocol !== "https:" && url.protocol !== "http:") return null
        return `${url.origin}${url.pathname}`
    } catch {
        return null
    }
}

function signedURLExpiry(value: string) {
    try {
        const url = new URL(value)
        const signedAt = url.searchParams.get("X-Amz-Date")
        const lifetime = Number(url.searchParams.get("X-Amz-Expires"))
        const match = signedAt?.match(/^(\d{4})(\d{2})(\d{2})T(\d{2})(\d{2})(\d{2})Z$/)
        if (!match || !Number.isFinite(lifetime) || lifetime <= 0) return null

        const [, year, month, day, hour, minute, second] = match
        return Date.UTC(
            Number(year),
            Number(month) - 1,
            Number(day),
            Number(hour),
            Number(minute),
            Number(second),
        ) + lifetime * 1000
    } catch {
        return null
    }
}

function storageKey(cacheKey: string) {
    return `${STORAGE_PREFIX}${cacheKey}`
}

function isCacheEntry(value: unknown): value is ProfilePictureCacheEntry {
    if (!value || typeof value !== "object") return false
    return "resource" in value
        && typeof value.resource === "string"
        && "url" in value
        && typeof value.url === "string"
        && "expiresAt" in value
        && typeof value.expiresAt === "number"
}

function readSessionEntry(cacheKey: string) {
    if (typeof window === "undefined") return null

    try {
        const stored = window.sessionStorage.getItem(storageKey(cacheKey))
        if (!stored) return null
        const value: unknown = JSON.parse(stored)
        return isCacheEntry(value) ? value : null
    } catch {
        return null
    }
}

function removeEntry(cacheKey: string) {
    memoryCache.delete(cacheKey)
    if (typeof window === "undefined") return

    try {
        window.sessionStorage.removeItem(storageKey(cacheKey))
    } catch {
        // The in-memory cache still works when session storage is unavailable.
    }
}

function writeEntry(cacheKey: string, entry: ProfilePictureCacheEntry) {
    memoryCache.set(cacheKey, entry)
    if (typeof window === "undefined") return

    try {
        window.sessionStorage.setItem(storageKey(cacheKey), JSON.stringify(entry))
    } catch {
        // The in-memory cache still works when session storage is unavailable.
    }
}

export function getCachedProfilePictureURL(cacheKey: string, freshURL: string) {
    const resource = resourceFromURL(freshURL)
    if (!resource) return freshURL

    const now = Date.now()
    const cached = memoryCache.get(cacheKey) ?? readSessionEntry(cacheKey)
    if (cached && cached.resource === resource && cached.expiresAt > now) {
        return cached.url
    }

    return freshURL
}

export function cacheProfilePictureURL(cacheKey: string, url: string) {
    const resource = resourceFromURL(url)
    if (!resource) return

    const now = Date.now()
    const cached = memoryCache.get(cacheKey) ?? readSessionEntry(cacheKey)
    if (cached && cached.resource === resource && cached.expiresAt > now) return

    removeEntry(cacheKey)
    const signedExpiry = signedURLExpiry(url)
    const expiresAt = Math.min(
        now + PROFILE_PICTURE_CACHE_TTL,
        signedExpiry ? signedExpiry - SIGNED_URL_EXPIRY_BUFFER : Number.POSITIVE_INFINITY,
    )

    if (expiresAt > now) {
        writeEntry(cacheKey, {resource, url, expiresAt})
    }
}

export function invalidateProfilePictureCache(cacheKey: string) {
    removeEntry(cacheKey)
}

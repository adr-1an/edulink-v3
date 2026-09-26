export interface PortalProfile {
    profilePictureURL: string | null
    name: string
    lastName: string
    email: string
    phone: string
    dateOfBirth: string
    school: {
        name: string
        region: string
        owner: {
            name: string
            email: string
        }
    }
}

function isRecord(value: unknown): value is Record<string, unknown> {
    return typeof value === "object" && value !== null
}

function normalizeProfilePictureURL(value: unknown) {
    if (typeof value !== "string") return null

    try {
        const url = new URL(value)
        if (url.protocol !== "https:" && url.protocol !== "http:") return null
    } catch {
        return null
    }

    return value
}

export function parsePortalProfile(value: unknown): PortalProfile | null {
    if (!isRecord(value)) return null
    const {name, lastName, email, phone, dateOfBirth, pfpUrl, school} = value
    if (
        typeof name !== "string"
        || typeof lastName !== "string"
        || typeof email !== "string"
        || typeof phone !== "string"
        || typeof dateOfBirth !== "string"
    ) {
        return null
    }

    if (!isRecord(school)) return null
    const {name: schoolName, region, owner} = school
    if (typeof schoolName !== "string" || typeof region !== "string" || !isRecord(owner)) return null

    const {name: ownerName, email: ownerEmail} = owner
    if (typeof ownerName !== "string" || typeof ownerEmail !== "string") return null

    return {
        profilePictureURL: normalizeProfilePictureURL(pfpUrl),
        name,
        lastName,
        email,
        phone,
        dateOfBirth,
        school: {
            name: schoolName,
            region,
            owner: {
                name: ownerName,
                email: ownerEmail,
            },
        },
    }
}

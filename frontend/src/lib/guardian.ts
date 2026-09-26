export interface Guardian {
    id: string
    name: string
    lastName: string
    email: string
    phone: string
    notes: string
    dateOfBirth: string
    accountEnabled: boolean
    accountActive: boolean
}

export interface GuardianDraft {
    name: string
    lastName: string
    dateOfBirth: string
    email: string
    phone: string
    notes: string
    accountEnabled: boolean
    password: string
}

export interface GuardianInput {
    name: string
    lastName: string
    dateOfBirth: string
    email: string
    phone: string | null
    notes: string | null
    accountEnabled: boolean
    password: string | null
}

export const emptyGuardianDraft: GuardianDraft = {
    name: "",
    lastName: "",
    dateOfBirth: "",
    email: "",
    phone: "",
    notes: "",
    accountEnabled: false,
    password: "",
}

function isRecord(value: unknown): value is Record<string, unknown> {
    return value !== null && typeof value === "object"
}

function dateOnly(value: string) {
    return value.match(/^\d{4}-\d{2}-\d{2}/)?.[0] ?? null
}

function isValidDateOnly(value: string) {
    if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) return false
    const [year, month, day] = value.split("-").map(Number)
    const date = new Date(Date.UTC(year, month - 1, day))
    return date.getUTCFullYear() === year
        && date.getUTCMonth() === month - 1
        && date.getUTCDate() === day
}

export function normalizeGuardianList(value: unknown): Guardian[] {
    if (!Array.isArray(value)) return []

    return value.flatMap((candidate) => {
        if (!isRecord(candidate)) return []

        const id = candidate.id
        const name = candidate.name
        const lastName = candidate.lastName
        const email = candidate.email
        const phone = candidate.phone
        const notes = candidate.notes
        const rawDateOfBirth = candidate.dateOfBirth
        const accountEnabled = candidate.accountEnabled
        const accountActive = candidate.accountActive
        const normalizedDateOfBirth = typeof rawDateOfBirth === "string" ? dateOnly(rawDateOfBirth) : null

        if (typeof id !== "string" || !/^\d+$/.test(id)
            || typeof name !== "string"
            || typeof lastName !== "string"
            || typeof email !== "string"
            || (phone !== null && typeof phone !== "string")
            || (notes !== null && typeof notes !== "string")
            || normalizedDateOfBirth === null
            || typeof accountEnabled !== "boolean"
            || typeof accountActive !== "boolean") return []

        return [{
            id,
            name,
            lastName,
            email,
            phone: phone ?? "",
            notes: notes ?? "",
            dateOfBirth: normalizedDateOfBirth,
            accountEnabled,
            accountActive,
        }]
    })
}

export function guardianDraftToInput(draft: GuardianDraft): GuardianInput {
    const phone = draft.phone.trim()
    const notes = draft.notes.trim()
    const password = draft.password

    return {
        name: draft.name.trim(),
        lastName: draft.lastName.trim(),
        dateOfBirth: `${draft.dateOfBirth}T00:00:00.000Z`,
        email: draft.email.trim().toLocaleLowerCase("en-US"),
        phone: phone || null,
        notes: notes || null,
        accountEnabled: draft.accountEnabled,
        password: password || null,
    }
}

export function isGuardianDraftValid(
    draft: GuardianDraft,
    mode: "create" | "edit",
    wasAccountEnabled: boolean,
) {
    const name = draft.name.trim()
    const lastName = draft.lastName.trim()
    const email = draft.email.trim()
    const phone = draft.phone.trim()
    const notes = draft.notes.trim()
    const passwordRequired = draft.accountEnabled && (mode === "create" || !wasAccountEnabled)

    return name.length >= 3 && name.length <= 32
        && lastName.length >= 3 && lastName.length <= 32
        && email.length >= 5 && email.length <= 254
        && isValidDateOnly(draft.dateOfBirth)
        && (!phone || (phone.length >= 3 && phone.length <= 32))
        && notes.length <= 2048
        && (!draft.password || draft.password.length >= 8)
        && (!passwordRequired || draft.password.length >= 8)
}

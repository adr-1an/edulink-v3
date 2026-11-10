export interface GuardianStudent {
    id: string
    name: string
    lastName: string
    dateOfBirth: string
}

function isRecord(value: unknown): value is Record<string, unknown> {
    return value !== null && typeof value === "object" && !Array.isArray(value)
}

export function parseGuardianStudentsJSON(body: string): {students: GuardianStudent[]; access: unknown} | null {
    try {
        // The handler currently returns Go structs with numeric int64 IDs.
        // Quote IDs before parsing to avoid rounding them in JavaScript.
        const data: unknown = JSON.parse(body.replace(/("(?:ID|id)"\s*:\s*)(\d+)(?=\s*[,}])/g, '$1"$2"'))
        if (!isRecord(data)) return null
        const rows = data.students === null ? [] : data.students
        if (!Array.isArray(rows)) return null
        const students: GuardianStudent[] = []
        const ids = new Set<string>()
        for (const row of rows) {
            if (!isRecord(row)) return null
            const id = row.id ?? row.ID
            const name = row.name ?? row.Name
            const lastName = row.lastName ?? row.LastName
            const birth = row.dateOfBirth ?? row.doB ?? row.DoB
            if (typeof id !== "string" || !/^\d+$/.test(id) || ids.has(id)
                || typeof name !== "string" || !name.trim()
                || typeof lastName !== "string" || !lastName.trim()
                || typeof birth !== "string") return null
            const dateOfBirth = birth.match(/^\d{4}-\d{2}-\d{2}/)?.[0]
            if (!dateOfBirth) return null
            const date = new Date(`${dateOfBirth}T00:00:00Z`)
            if (!Number.isFinite(date.getTime()) || date.toISOString().slice(0, 10) !== dateOfBirth) return null
            ids.add(id)
            students.push({id, name: name.trim(), lastName: lastName.trim(), dateOfBirth})
        }
        return {students, access: data.access}
    } catch {
        return null
    }
}

export interface StudentGuardian {
    id: string
    name: string
    lastName: string
    email: string
    phone: string
}

// Preserve numeric IDs in legacy guardian records before JSON.parse can round them.
// Current handler records use camelCase fields and string guardian IDs.
export function parseStudentGuardiansJSON(body: string, allowRepeatedRecords = false): StudentGuardian[] | null {
    try {
        const data = JSON.parse(body.replace(/("(?:ID|id)"\s*:\s*)(\d+)(?=\s*[,}])/g, '$1"$2"'))
        if (!Array.isArray(data.guardians)) return data.guardians === null ? [] : null
        const guardians: StudentGuardian[] = []
        for (const row of data.guardians) {
            if (!row || typeof row !== "object") return null
            const id = row.id ?? row.ID
            const name = row.name ?? row.Name
            const lastName = row.lastName ?? row.LastName
            const email = row.email ?? row.Email
            const phone = row.phone ?? row.Phone ?? ""
            if (typeof id !== "string" || !/^\d+$/.test(id)
                || typeof name !== "string" || typeof lastName !== "string"
                || typeof email !== "string" || typeof phone !== "string") return null
            guardians.push({id, name, lastName, email, phone})
        }
        if (allowRepeatedRecords) return deduplicateStudentRecords(guardians)
        if (new Set(guardians.map((guardian) => guardian.id)).size !== guardians.length) return null
        return guardians
    } catch {
        return null
    }
}

export function parseStudentGuardiansResponseJSON(body: string): {
    schoolID: string
    guardians: StudentGuardian[]
    access: unknown
} | null {
    try {
        // schoolId is currently serialized as a Go int64 rather than a string.
        const data = JSON.parse(body.replace(/("schoolI[dD]"\s*:\s*)(\d+)(?=\s*[,}])/g, '$1"$2"'))
        const schoolID = data?.schoolId ?? data?.schoolID
        if (typeof schoolID !== "string" || !/^\d+$/.test(schoolID)) return null
        const guardians = data.student !== undefined
            ? parseStudentGuardiansJSON(JSON.stringify({guardians: data.student?.guardians}), true)
            : parseStudentGuardiansJSON(body)
        if (!guardians) return null
        return {schoolID, guardians, access: data.access}
    } catch {
        return null
    }
}

// Joining submissions with guardians can repeat otherwise identical records.
export function deduplicateStudentRecords<T extends {id: string}>(records: T[]): T[] | null {
    const unique = new Map<string, T>()
    for (const record of records) {
        const existing = unique.get(record.id)
        if (existing && JSON.stringify(existing) !== JSON.stringify(record)) return null
        unique.set(record.id, record)
    }
    return [...unique.values()]
}

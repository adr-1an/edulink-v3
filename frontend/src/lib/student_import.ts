export interface StudentImportRow {
    name: string
    lastName: string
    dateOfBirth: string
    email: string
    phone: string
    notes: string
}

export type StudentImportParseError = "empty" | "invalid_csv" | "missing_headers" | "no_rows" | "too_many_rows"

export type StudentImportParseResult =
    | {ok: true; rows: StudentImportRow[]}
    | {ok: false; error: StudentImportParseError; missingHeaders?: string[]}

type StudentColumn = keyof StudentImportRow

const requiredColumns: StudentColumn[] = ["name", "lastName", "dateOfBirth", "email"]
const headerAliases: Record<StudentColumn, Set<string>> = {
    name: new Set(["name", "firstname", "givenname", "imie"]),
    lastName: new Set(["lastname", "surname", "familyname", "nazwisko"]),
    dateOfBirth: new Set(["dateofbirth", "birthdate", "dob", "dataurodzenia"]),
    email: new Set(["email", "emailaddress", "adresmailowy"]),
    phone: new Set(["phone", "phonenumber", "telephone", "telefon"]),
    notes: new Set(["notes", "note", "comments", "notatki", "uwagi"]),
}

function normalizeHeader(value: string) {
    return value
        .trim()
        .toLocaleLowerCase("en-US")
        .normalize("NFD")
        .replace(/[\u0300-\u036f]/g, "")
        .replace(/[^a-z0-9]/g, "")
}

function columnForHeader(value: string): StudentColumn | null {
    const normalized = normalizeHeader(value)
    const match = (Object.entries(headerAliases) as Array<[StudentColumn, Set<string>]>)
        .find(([, aliases]) => aliases.has(normalized))
    return match?.[0] ?? null
}

function parseDelimitedText(value: string, delimiter: string): string[][] | null {
    const rows: string[][] = []
    let row: string[] = []
    let field = ""
    let quoted = false

    for (let index = 0; index < value.length; index += 1) {
        const character = value[index]

        if (quoted) {
            if (character === '"' && value[index + 1] === '"') {
                field += '"'
                index += 1
            } else if (character === '"') {
                quoted = false
            } else {
                field += character
            }
            continue
        }

        if (character === '"' && field.length === 0) {
            quoted = true
        } else if (character === delimiter) {
            row.push(field)
            field = ""
        } else if (character === "\n" || character === "\r") {
            if (character === "\r" && value[index + 1] === "\n") index += 1
            row.push(field)
            rows.push(row)
            row = []
            field = ""
        } else {
            field += character
        }
    }

    if (quoted) return null
    row.push(field)
    rows.push(row)
    return rows.filter((candidate) => candidate.some((cell) => cell.trim()))
}

function mappedHeaders(row: string[]) {
    const columns = new Map<StudentColumn, number>()
    row.forEach((header, index) => {
        const column = columnForHeader(header)
        if (column && !columns.has(column)) columns.set(column, index)
    })
    return columns
}

export function parseStudentImportFile(value: string): StudentImportParseResult {
    const text = value.replace(/^\uFEFF/, "")
    if (!text.trim()) return {ok: false, error: "empty"}

    const candidates = [",", ";", "\t"].flatMap((delimiter) => {
        const rows = parseDelimitedText(text, delimiter)
        if (!rows?.length) return []
        const columns = mappedHeaders(rows[0])
        return [{rows, columns, score: columns.size}]
    })
    const candidate = candidates.sort((first, second) => second.score - first.score)[0]
    if (!candidate) return {ok: false, error: "invalid_csv"}

    const missingHeaders = requiredColumns.filter((column) => !candidate.columns.has(column))
    if (missingHeaders.length > 0) return {ok: false, error: "missing_headers", missingHeaders}

    const rows = candidate.rows.slice(1)
        .filter((row) => row.some((cell) => cell.trim()))
        .map((row) => ({
            name: row[candidate.columns.get("name") ?? -1]?.trim() ?? "",
            lastName: row[candidate.columns.get("lastName") ?? -1]?.trim() ?? "",
            dateOfBirth: row[candidate.columns.get("dateOfBirth") ?? -1]?.trim() ?? "",
            email: row[candidate.columns.get("email") ?? -1]?.trim().toLocaleLowerCase("en-US") ?? "",
            phone: row[candidate.columns.get("phone") ?? -1]?.trim() ?? "",
            notes: row[candidate.columns.get("notes") ?? -1]?.trim() ?? "",
        }))

    if (rows.length === 0) return {ok: false, error: "no_rows"}
    if (rows.length > 10_000) return {ok: false, error: "too_many_rows"}
    return {ok: true, rows}
}

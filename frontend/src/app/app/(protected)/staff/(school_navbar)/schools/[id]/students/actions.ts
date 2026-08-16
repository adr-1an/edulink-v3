"use server"

import {cookies} from "next/headers"

export type StudentActionErrorCode =
    | "invalid_school" | "invalid_student" | "network" | "invalid_data" | "unauthorized"
    | "forbidden" | "missing" | "conflict" | "validation" | "server" | "save"

function failure<const Code extends StudentActionErrorCode>(code: Code, message: string) {
    return {ok: false as const, code, message}
}

export interface StudentInput {
    name: string
    lastName: string
    dob: string | null
    email: string
    phone: string
    notes: string
    accountEnabled: boolean
    password: string
}

export interface ImportedStudentInput {
    name: string
    lastName: string
    dateOfBirth: string
    email: string
    phone: string | null
    notes: string | null
}

async function studentRequest(path: string, method: "POST" | "PATCH" | "DELETE", input?: StudentInput) {
    const token = (await cookies()).get("token")?.value
    let res: Response

    try {
        res = await fetch(`${process.env.API_URL}${path}`, {
            method,
            headers: {
                ...(input ? {"Content-Type": "application/json"} : {}),
                Authorization: `Bearer ${token}`,
            },
            body: input ? JSON.stringify(input) : undefined,
            cache: "no-store",
        })
    } catch {
        return failure("network", "Network error, please try again.")
    }

    if (res.ok) return {ok: true as const}
    if (res.status === 400) return failure("invalid_data", "Some student information is invalid.")
    if (res.status === 401) return failure("unauthorized", "Unauthorized.")
    if (res.status === 403) return failure("forbidden", "You don't have permission to perform this action.")
    if (res.status === 404) return failure("missing", "This student no longer exists.")
    if (res.status === 409) return failure("conflict", "A student with this email already exists in the school.")
    if (res.status === 422) return failure("validation", "Check the student's details and try again.")
    if (res.status === 500) return failure("server", "Internal server error.")
    return failure("save", "Unable to save the student.")
}

export async function handleCreateStudent(schoolID: string, input: StudentInput) {
    if (!/^\d+$/.test(schoolID)) return failure("invalid_school", "Invalid school.")
    return studentRequest(`/v1/staff/schools/${schoolID}/students`, "POST", input)
}

export async function handleUpdateStudent(studentID: string, input: StudentInput) {
    if (!/^\d+$/.test(studentID)) return failure("invalid_student", "Invalid student.")
    return studentRequest(`/v1/staff/students/${studentID}`, "PATCH", input)
}

export async function handleDeleteStudent(studentID: string) {
    if (!/^\d+$/.test(studentID)) return failure("invalid_student", "Invalid student.")
    return studentRequest(`/v1/staff/students/${studentID}`, "DELETE")
}

export async function handleImportStudents(schoolID: string, students: ImportedStudentInput[], enableAccounts: boolean) {
    if (!/^\d+$/.test(schoolID)) return failure("invalid_school", "Invalid school.")
    if (students.length === 0 || students.length > 10_000) return failure("validation", "The import size is invalid.")

    const token = (await cookies()).get("token")?.value
    let res: Response

    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/schools/${schoolID}/students/import`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({students, enableAccounts}),
            cache: "no-store",
        })
    } catch {
        return failure("network", "Network error, please try again.")
    }

    if (res.ok) return {ok: true as const}
    if (res.status === 400) return failure("invalid_data", "Some student information is invalid.")
    if (res.status === 401) return failure("unauthorized", "Unauthorized.")
    if (res.status === 403) return failure("forbidden", "You don't have permission to import students.")
    if (res.status === 409) {
        let payload: unknown
        try {
            payload = await res.json()
        } catch {
            payload = null
        }

        if (payload && typeof payload === "object" && "conflictingEmails" in payload && Array.isArray(payload.conflictingEmails)) {
            const conflictingEmails = [...new Set(payload.conflictingEmails.flatMap((value) => {
                if (typeof value !== "string") return []
                const email = value.trim().toLocaleLowerCase("en-US")
                return email && email.length <= 254 ? [email] : []
            }))]
            if (conflictingEmails.length > 0) {
                return {
                    ...failure("conflict", "Some students use emails that already exist in the school."),
                    conflictingEmails,
                }
            }
        }
        return failure("conflict", "One of these emails already belongs to a student in the school.")
    }
    if (res.status === 422) {
        let payload: unknown
        try {
            payload = await res.json()
        } catch {
            payload = null
        }

        if (payload && typeof payload === "object" && "invalidRow" in payload && Number.isSafeInteger(payload.invalidRow)) {
            return {
                ...failure("validation", "One imported student is invalid."),
                invalidRow: payload.invalidRow as number,
            }
        }
        return failure("validation", "Check the imported students and try again.")
    }
    if (res.status === 500) return failure("server", "Internal server error.")
    return failure("save", "Unable to import the students.")
}

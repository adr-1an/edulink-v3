"use server"

import {cookies} from "next/headers"

async function changeAssignment(studentID: string, guardianID: string, method: "PUT" | "DELETE") {
    if (!/^\d+$/.test(studentID) || !/^\d+$/.test(guardianID)) return {ok: false as const, code: "invalid" as const}
    const token = (await cookies()).get("token")?.value
    try {
        const response = await fetch(`${process.env.API_URL}/v1/staff/students/${studentID}/guardians/${guardianID}`, {
            method,
            headers: {Authorization: `Bearer ${token}`},
            cache: "no-store",
        })
        if (response.ok) return {ok: true as const}
        if (response.status === 401) return {ok: false as const, code: "unauthorized" as const}
        if (response.status === 403) return {ok: false as const, code: "forbidden" as const}
        if (response.status === 400) return {ok: false as const, code: "invalid" as const}
        return {ok: false as const, code: "server" as const}
    } catch {
        return {ok: false as const, code: "network" as const}
    }
}

export async function assignStudentGuardian(studentID: string, guardianID: string) {
    return changeAssignment(studentID, guardianID, "PUT")
}

export async function unassignStudentGuardian(studentID: string, guardianID: string) {
    return changeAssignment(studentID, guardianID, "DELETE")
}

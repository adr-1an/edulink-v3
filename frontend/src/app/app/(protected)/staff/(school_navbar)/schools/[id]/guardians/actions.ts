"use server"

import {cookies} from "next/headers"
import {type GuardianInput} from "@/lib/guardian"

export type GuardianActionErrorCode =
    | "invalid_school" | "invalid_guardian" | "network" | "invalid_data" | "unauthorized"
    | "forbidden" | "missing" | "conflict" | "validation" | "server" | "save"

function failure<const Code extends GuardianActionErrorCode>(code: Code) {
    return {ok: false as const, code}
}

async function guardianRequest(path: string, method: "POST" | "PATCH" | "DELETE", input?: GuardianInput) {
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
        return failure("network")
    }

    if (res.ok) return {ok: true as const}
    if (res.status === 400) return failure("invalid_data")
    if (res.status === 401) return failure("unauthorized")
    if (res.status === 403) return failure("forbidden")
    if (res.status === 404) return failure("missing")
    if (res.status === 409) return failure("conflict")
    if (res.status === 422) return failure("validation")
    if (res.status === 500) return failure("server")
    return failure("save")
}

export async function handleCreateGuardian(schoolID: string, input: GuardianInput) {
    if (!/^\d+$/.test(schoolID)) return failure("invalid_school")
    return guardianRequest(`/v1/staff/schools/${schoolID}/guardians`, "POST", input)
}

export async function handleUpdateGuardian(guardianID: string, input: GuardianInput) {
    if (!/^\d+$/.test(guardianID)) return failure("invalid_guardian")
    return guardianRequest(`/v1/staff/guardians/${guardianID}`, "PATCH", input)
}

export async function handleDeleteGuardian(guardianID: string) {
    if (!/^\d+$/.test(guardianID)) return failure("invalid_guardian")
    return guardianRequest(`/v1/staff/guardians/${guardianID}`, "DELETE")
}

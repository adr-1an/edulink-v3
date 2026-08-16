"use server"


import {cookies} from "next/headers"

interface UpdateSchoolPayload {
    name: string
    regionCode: string
    activeAcademicYearId?: string
}

export interface SchoolDeletionChallenge {
    token: string
    expiresAt: string
}

type SchoolDeletionResponse = {
    twoFactorChallenge?: {
        token?: unknown
        purpose?: unknown
        expiresAt?: unknown
    }
}

function parseSchoolDeletionChallenge(data: SchoolDeletionResponse): SchoolDeletionChallenge | null {
    const challenge = data.twoFactorChallenge
    if (!challenge
        || challenge.purpose !== "schoolDeletion"
        || typeof challenge.token !== "string"
        || challenge.token.length < 32
        || typeof challenge.expiresAt !== "string"
        || !Number.isFinite(Date.parse(challenge.expiresAt))) {
        return null
    }

    return {token: challenge.token, expiresAt: challenge.expiresAt}
}

export async function handleUpdateSchool(schoolID: string, payload: UpdateSchoolPayload) {
    if (!/^\d+$/.test(schoolID) || (payload.activeAcademicYearId && !/^\d+$/.test(payload.activeAcademicYearId))) {
        return {ok: false, message: "Select a valid academic year."}
    }

    const token = (await cookies()).get("token")?.value
    let res
    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/schools/${schoolID}`, {
            method: "PATCH",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            // Keep the 64-bit ID as numeric JSON without converting it to an unsafe JS number.
            body: `{"name":${JSON.stringify(payload.name)},"regionCode":${JSON.stringify(payload.regionCode)}${payload.activeAcademicYearId ? `,"activeAcademicYearId":${payload.activeAcademicYearId}` : ""}}`,
        })
    } catch {
        return {ok: false, message: "Network error, please try again."}
    }

    if (res.ok) return {ok: true}

    let code: string | undefined
    try { code = (await res.json()).code } catch {}
    if (code === "INVALID_NAME") return {ok: false, message: "Enter a school name between 1 and 64 characters."}
    if (code === "INVALID_REGION_CODE") return {ok: false, message: "Select a valid school region."}
    if (res.status === 401) return {ok: false, message: "Unauthorized."}
    if (res.status === 403) return {ok: false, message: "You don't have permission to update this school."}
    if (res.status === 500) return {ok: false, message: "Internal server error."}
    return {ok: false, message: "Unable to update the school."}
}

async function schoolRequest(schoolID: string, path: string, body: unknown) {
    if (!/^\d+$/.test(schoolID)) return {ok: false, message: "Invalid school."}

    const token = (await cookies()).get("token")?.value
    let res
    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/schools/${schoolID}${path}`, {
            method: "POST",
            headers: {"Content-Type": "application/json", Authorization: `Bearer ${token}`},
            body: JSON.stringify(body),
        })
    } catch {
        return {ok: false, message: "Network error, please try again."}
    }

    if (res.ok) return {ok: true}

    let code: string | undefined
    try { code = (await res.json()).code } catch {}
    if (code === "ACADEMIC_YEAR_CONFLICT") return {ok: false, message: "That academic year already exists."}
    if (code === "NO_ACTIVE_YEAR") return {ok: false, message: "Select an active academic year before promoting."}
    if (res.status === 401) return {ok: false, message: "Unauthorized."}
    if (res.status === 403) return {ok: false, message: "You don't have permission to manage academic years."}
    if (res.status === 422) return {ok: false, message: "Enter a valid academic-year range."}
    if (res.status === 500) return {ok: false, message: "Internal server error."}
    return {ok: false, message: "Unable to update academic years."}
}

export async function handleCreateAcademicYear(schoolID: string, from: number, to: number) {
    return schoolRequest(schoolID, "/academic-years", {academicYear: {from, to}})
}

interface PromotionOptions {
    activateAfterPromotion: boolean
    transferGrades: boolean
    promoteGradeLevels: boolean
}

export async function handlePromoteSchool(schoolID: string, from: number, to: number, options: PromotionOptions) {
    return schoolRequest(schoolID, "/promote", {newAcademicYear: {from, to}, options})
}

export async function handleDeleteAcademicYear(yearID: string) {
    if (!/^\d+$/.test(yearID)) return {ok: false, message: "Invalid academic year."}

    const token = (await cookies()).get("token")?.value
    let res
    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/academic-years/${yearID}`, {
            method: "DELETE",
            headers: {Authorization: `Bearer ${token}`},
        })
    } catch {
        return {ok: false, message: "Network error, please try again."}
    }

    if (res.ok) return {ok: true}
    if (res.status === 401) return {ok: false, message: "Unauthorized."}
    if (res.status === 403) return {ok: false, message: "You don't have permission to delete academic years."}
    if (res.status === 409) return {ok: false, message: "The active academic year cannot be deleted. Activate another year first."}
    if (res.status === 500) return {ok: false, message: "This academic year could not be deleted. Remove its grades first."}
    return {ok: false, message: "Unable to delete the academic year."}
}

export async function handleClearActiveAcademicYear(schoolID: string) {
    if (!/^\d+$/.test(schoolID)) return {ok: false, message: "Invalid school."}

    const token = (await cookies()).get("token")?.value
    let res
    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/schools/${schoolID}/academic-years`, {
            method: "PUT",
            headers: {Authorization: `Bearer ${token}`},
        })
    } catch {
        return {ok: false, message: "Network error, please try again."}
    }

    if (res.ok) return {ok: true}
    if (res.status === 401) return {ok: false, message: "Unauthorized."}
    if (res.status === 403) return {ok: false, message: "You don't have permission to change the active academic year."}
    if (res.status === 500) return {ok: false, message: "Unable to clear the active academic year."}
    return {ok: false, message: "Unable to clear the active academic year."}
}

export async function handleDeleteSchool(schoolID: string) {
    if (!/^\d+$/.test(schoolID)) return {ok: false, message: "Invalid school."}

    const token = (await cookies()).get("token")?.value
    let res
    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/schools/${schoolID}`, {
            method: "DELETE",
            headers: {Authorization: `Bearer ${token}`},
        })
    } catch {
        return {ok: false, message: "Network error, please try again."}
    }

    if (res.status === 204) return {ok: true as const, step: "complete" as const}
    if (res.ok) {
        try {
            const data = await res.json() as SchoolDeletionResponse
            const challenge = parseSchoolDeletionChallenge(data)
            if (!challenge) return {ok: false as const, code: "invalid_response" as const}
            return {ok: true as const, step: "two_factor" as const, challenge}
        } catch {
            return {ok: false as const, code: "invalid_response" as const}
        }
    }
    if (res.status === 401) return {ok: false, message: "Unauthorized."}
    if (res.status === 403) return {ok: false, message: "You don't have permission to delete this school."}
    if (res.status === 500) return {ok: false, message: "Internal server error."}
    return {ok: false, message: "Unable to delete the school."}
}

export async function handleCompleteSchoolDeletion(challengeToken: string, code: string) {
    const normalizedCode = code.trim()
    if (challengeToken.length < 32 || challengeToken.length > 256 || !/^\d{6}$/.test(normalizedCode)) {
        return {ok: false as const, code: "invalid_input" as const}
    }

    let res: Response
    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/auth/two-factor/challenge`, {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify({
                code: normalizedCode,
                challengeToken,
            }),
        })
    } catch {
        return {ok: false as const, code: "network" as const}
    }

    if (res.status === 204) return {ok: true as const}
    if (res.status === 401) return {ok: false as const, code: "invalid_code" as const}
    if (res.status === 404) return {ok: false as const, code: "invalid_challenge" as const}
    if (res.status === 422) {
        const data = await res.json().catch(() => null) as {code?: unknown} | null
        return {
            ok: false as const,
            code: data?.code === "EXPIRED_TOKEN" ? "expired" as const : "invalid_input" as const,
        }
    }
    if (res.status === 403) return {ok: false as const, code: "forbidden" as const}
    if (res.ok) return {ok: false as const, code: "invalid_response" as const}
    return {ok: false as const, code: "server" as const}
}

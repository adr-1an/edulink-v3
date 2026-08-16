"use server"

import {cookies} from "next/headers"
import type {SchoolAccess} from "@/lib/school_access"

function isSchoolAccess(value: unknown): value is SchoolAccess {
    if (!value || typeof value !== "object") return false

    const candidate = value as Partial<SchoolAccess>
    return typeof candidate.owner === "boolean" && Array.isArray(candidate.roles)
}

export async function getSchoolNavigationAccess(schoolID: string): Promise<SchoolAccess | null> {
    if (!/^\d+$/.test(schoolID)) return null

    const token = (await cookies()).get("token")?.value
    const headers = {Authorization: `Bearer ${token}`}
    const paths = [
        `/v1/staff/schools/${schoolID}`,
        `/v1/staff/schools/${schoolID}/grades`,
        `/v1/staff/schools/${schoolID}/academic-years`,
        `/v1/staff/schools/${schoolID}/staff`,
        `/v1/staff/schools/${schoolID}/students`,
        `/v1/staff/schools/${schoolID}/roles`,
    ]

    try {
        const responses = await Promise.all(paths.map((path) => fetch(`${process.env.API_URL}${path}`, {
            headers,
            cache: "no-store",
        })))

        for (const response of responses) {
            if (!response.ok) continue
            const data: unknown = await response.json()
            if (data && typeof data === "object" && "access" in data && isSchoolAccess(data.access)) {
                return data.access
            }
        }

        // A successful school dashboard response proves the user passed school.view,
        // even though that endpoint does not currently include an access object.
        if (responses[0].ok) {
            return {
                owner: false,
                roles: [{position: -1, permissions: ["school.view"]}],
            }
        }
    } catch {
        return null
    }

    return null
}

"use server"

import {cookies} from "next/headers"
import {isSchoolAccess, type SchoolAccess} from "@/lib/school_access"

// Only used on demand by pages whose own API response omits access data.
// The sidebar must not call this to decide which links to display.
export async function getSchoolNavigationAccess(schoolID: string): Promise<SchoolAccess | null> {
    if (!/^\d+$/.test(schoolID)) return null

    const token = (await cookies()).get("token")?.value
    const headers = {Authorization: `Bearer ${token}`}
    const paths = [
        `/v1/staff/schools/${schoolID}/grades`,
        `/v1/staff/schools/${schoolID}/academic-years`,
        `/v1/staff/schools/${schoolID}/staff`,
        `/v1/staff/schools/${schoolID}/students`,
        `/v1/staff/schools/${schoolID}/guardians`,
    ]

    for (const path of paths) {
        try {
            const response = await fetch(`${process.env.API_URL}${path}`, {headers, cache: "no-store"})
            if (!response.ok) continue
            const data: unknown = await response.json()
            if (data && typeof data === "object" && "access" in data && isSchoolAccess(data.access)) {
                return data.access
            }
        } catch {
            // A failed source should not prevent trying another accessible source.
        }
    }
    return null
}

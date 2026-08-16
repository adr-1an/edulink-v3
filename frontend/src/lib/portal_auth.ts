import "server-only"

import {cookies} from "next/headers"

export type PortalAccountType = "student" | "guardian"

interface PortalAuthResponse {
    user?: {
        id?: string | number
        accountType?: string
    }
}

export type PortalSession =
    | {status: "authenticated"; token: string; user: {accountType: PortalAccountType}}
    | {status: "unauthenticated"}
    | {status: "unavailable"}

export function portalHome(accountType: PortalAccountType) {
    return `/app/portal/${accountType}` as const
}

export async function getPortalSession(): Promise<PortalSession> {
    const token = (await cookies()).get("portal_token")?.value
    if (!token) return {status: "unauthenticated"}

    let res: Response
    try {
        res = await fetch(`${process.env.API_URL}/v1/portal/auth`, {
            headers: {Authorization: `Bearer ${token}`},
            cache: "no-store",
        })
    } catch {
        return {status: "unavailable"}
    }

    if (res.status === 401) return {status: "unauthenticated"}
    if (!res.ok) return {status: "unavailable"}

    let data: PortalAuthResponse
    try {
        data = await res.json() as PortalAuthResponse
    } catch {
        return {status: "unavailable"}
    }

    const accountType = data.user?.accountType
    if (accountType !== "student" && accountType !== "guardian") return {status: "unavailable"}

    return {
        status: "authenticated",
        token,
        user: {
            accountType,
        },
    }
}

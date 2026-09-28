"use server"

import {cookies} from "next/headers"

export type PortalUserActivationErrorCode =
    | "invalid_user" | "network" | "unauthorized" | "forbidden" | "already_active" | "server" | "send"

function failure<const Code extends PortalUserActivationErrorCode>(code: Code) {
    return {ok: false as const, code}
}

export async function handleSendPortalUserActivationLink(portalUserID: string) {
    if (!/^\d+$/.test(portalUserID)) return failure("invalid_user")

    const token = (await cookies()).get("token")?.value
    let response: Response

    try {
        response = await fetch(`${process.env.API_URL}/v1/staff/portal-users/${portalUserID}/activation/send`, {
            method: "POST",
            headers: {Authorization: `Bearer ${token}`},
            cache: "no-store",
        })
    } catch {
        return failure("network")
    }

    if (response.ok) return {ok: true as const}
    if (response.status === 401) return failure("unauthorized")
    if (response.status === 403) return failure("forbidden")
    if (response.status === 409) return failure("already_active")
    if (response.status === 500) return failure("server")
    return failure("send")
}

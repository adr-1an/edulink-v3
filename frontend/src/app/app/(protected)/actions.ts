"use server"

import {cookies} from "next/headers"
import {verifyTurnstile} from "@/lib/turnstile"

export async function handleLogout() {
    const cookieStore = await cookies()
    const token = cookieStore.get("token")?.value

    let errorCode = 0
    let res

    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/auth/logout`, {
            method: "DELETE",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
        })

        if (!res.ok) errorCode = res.status
    } catch {
        errorCode = -1
    }

    if (errorCode) {
        if (errorCode === 500) {
            return { ok: false, message: "Internal server error." }
        } else if (errorCode === -1) {
            return { ok: false, message: "Network error, please try again." }
        } else {
            return { ok: false, message: "An unexpected error occurred." }
        }
    }

    return { ok: true }
}

export async function handleFetchStaffInvitations() {
    const cookieStore = await cookies()
    const token = cookieStore.get("token")?.value

    let errorCode = 0
    let res

    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/staff-invitations`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            cache: "no-store",
        })

        if (!res.ok) errorCode = res.status
    } catch {
        errorCode = -1
    }

    if (errorCode) {
        if (errorCode === 500) {
            return { ok: false, message: "Internal server error." }
        } else if (errorCode === 401) {
            return { ok: false, message: "Unauthorized." }
        } else {
            return { ok: false, message: "An unexpected error occurred." }
        }
    }

    let data
    if (res) {
        data = await res.json()
    }

    return { ok: true, data: data.invitations }
}

export async function handleStaffInvitation(
    invitationID: string,
    decision: "accept" | "reject",
) {
    const cookieStore = await cookies()
    const token = cookieStore.get("token")?.value

    let res
    try {
        res = await fetch(
            `${process.env.API_URL}/v1/staff/staff-invitations/by-id/${invitationID}/${decision}`,
            {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    Authorization: `Bearer ${token}`,
                },
            },
        )
    } catch {
        return { ok: false, message: "Network error, please try again." }
    }

    if (res.ok) return { ok: true }
    if (res.status === 401) return { ok: false, message: "Unauthorized." }
    if (res.status === 403) return { ok: false, message: "This invitation belongs to another account." }
    if (res.status === 404) return { ok: false, message: "This invitation is no longer available." }
    if (res.status === 500) return { ok: false, message: "Internal server error." }

    return { ok: false, message: `Unable to ${decision} the invitation.` }
}

export async function handleCreateSchool(formData: FormData) {
    const cookieStore = await cookies()
    const token = cookieStore.get("token")?.value
    const turnstileToken = formData.get("cf-turnstile-response")?.toString() ?? ""

    const turnstile = await verifyTurnstile(turnstileToken)
    if (!turnstile.ok) return turnstile

    let res

    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/schools`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({
                name: formData.get("name")?.toString() ?? "",
                regionCode: formData.get("regionCode")?.toString() ?? "",
            }),
        })
    } catch {
        return { ok: false, message: "Network error, please try again." }
    }

    if (res.ok) return { ok: true }

    let code: string | undefined
    try {
        const data = await res.json()
        code = data.code
    } catch {
        // The API does not return a body for every error response.
    }

    if (code === "INVALID_NAME") {
        return { ok: false, message: "Enter a school name between 1 and 64 characters." }
    }
    if (code === "INVALID_REGION_CODE") {
        return { ok: false, message: "Select a valid school region." }
    }
    if (res.status === 401) {
        return { ok: false, message: "Unauthorized." }
    }
    if (res.status === 500) {
        return { ok: false, message: "Internal server error." }
    }

    return { ok: false, message: "Unable to create the school." }
}

export async function handleLeaveSchool(schoolID: string) {
    if (!/^\d+$/.test(schoolID)) return {ok: false, message: "Invalid school."}

    const token = (await cookies()).get("token")?.value
    let res: Response

    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/schools/${schoolID}/leave`, {
            method: "DELETE",
            headers: {Authorization: `Bearer ${token}`},
        })
    } catch {
        return {ok: false, message: "Network error, please try again."}
    }

    if (res.ok) return {ok: true}
    if (res.status === 401) return {ok: false, message: "Unauthorized."}
    if (res.status === 403) return {ok: false, message: "You can't leave this school."}
    if (res.status === 404) return {ok: false, message: "You are no longer a staff member of this school."}
    if (res.status === 500) return {ok: false, message: "Internal server error."}

    return {ok: false, message: "Unable to leave the school."}
}

"use server"

import {cookies} from "next/headers"
import {refresh} from "next/cache"
import {redirect} from "next/navigation"

export async function handleAccept(invToken: string) {
    const cookieStore = await cookies()
    const token = cookieStore.get("token")?.value

    let errorCode = 0
    let res

    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/staff-invitations/${invToken}/accept`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            }
        })

        if (!res.ok) errorCode = res.status
    } catch {
        return { status: 500, ok: false, message: "Network error, please try again." }
    }

    if (errorCode) {
        if (errorCode === 500) {
            return { status: 500, ok: false, message: "Internal server error." }
        } else {
            return { status: res.status, ok: false, message: "An unexpected error occurred." }
        }
    }

    cookieStore.set("success", "Invitation accepted!")

    redirect(`/app`)
}

export async function handleReject(invToken: string) {
    const cookieStore = await cookies()
    const token = cookieStore.get("token")?.value

    let errorCode = 0
    let res

    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/staff-invitations/${invToken}/reject`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
        })

        if (!res.ok) errorCode = res.status
    } catch {
        return { status: 500, ok: false, message: "Network error, please try again." }
    }

    if (errorCode) {
        if (errorCode === 500) {
            return { status: 500, ok: false, message: "Internal server error." }
        } else {
            return { status: res.status, ok: false, message: "An unexpected error occurred." }
        }
    }

    refresh()
    return { ok: true }
}
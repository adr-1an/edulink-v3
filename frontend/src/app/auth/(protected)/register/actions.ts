"use server"

import {verifyTurnstile} from "@/lib/turnstile"

export async function handleRegistrationLinkSend(formData: FormData) {
    const turnstileToken = formData.get("cf-turnstile-response")?.toString() ?? ""
    const turnstile = await verifyTurnstile(turnstileToken)
    if (!turnstile.ok) return turnstile

    try {
        const res = await fetch(`${process.env.API_URL}/v1/staff/auth/register`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                email: formData.get("email"),
            })
        })

        if (!res.ok) {
            if (res.status === 500) {
                return { status: 500, ok: false, message: "Internal server error." }
            } else {
                return { status: res.status, ok: false, message: "Something went wrong." }
            }
        }

        return { status: res.status, ok: true }
    } catch {
        return { status: 500, ok: false, message: "Network error, please try again." }
    }
}

export async function handleRegistration(formData: FormData, token: string) {
    const turnstileToken = formData.get("cf-turnstile-response")?.toString() ?? ""
    const turnstile = await verifyTurnstile(turnstileToken)
    if (!turnstile.ok) return turnstile

    try {
        const res = await fetch(`${process.env.API_URL}/v1/staff/auth/register/${token}`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                name: formData.get("name"),
                phone: formData.get("phone"),
                password: formData.get("password"),
            })
        })

        if (!res.ok) {
            if (res.status === 500) {
                return { status: 500, ok: false, message: "Internal server error." }
            } else if (res.status === 422) {
                return { status: 422, ok: false, message: "Invalid data." }
            } else if (res.status === 404) {
                return { status: 404, ok: false, message: "Invalid or expired registration token." }
            } else {
                return { status: res.status, ok: false, message: "Something went wrong." }
            }
        }

        return { status: res.status, ok: true }
    } catch {
        return { status: 500, ok: false, message: "Network error, please try again." }
    }
}

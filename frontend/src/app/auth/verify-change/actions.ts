"use server"

export async function handleEmailChangeVerification(token: string) {
    try {
        const res = await fetch(`${process.env.API_URL}/v1/staff/profile/email/${token}`, {
            method: "PUT",
            headers: {
                "Content-Type": "application/json",
            }
        })

        if (!res.ok) {
            if (res.status === 404) {
                return { ok: false, message: "Invalid or expired verification token." }
            } else {
                return { ok: false, message: "Something went wrong." }
            }
        }

        return { ok: true }
    } catch {
        return { ok: false, message: "Network error, please try again." }
    }
}
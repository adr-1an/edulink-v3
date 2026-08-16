"use server"

export async function handleResetPassword(newPassword: string, token: string) {
    try {
        const res = await fetch(`${process.env.API_URL}/v1/staff/auth/reset/${token}`, {
            method: "PUT",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                newPassword: newPassword,
            })
        })

        if (!res.ok) {
            if (res.status === 404) {
                return { status: 404, ok: false, message: "Invalid reset token, try requesting a new reset link." }
            } else {
                return { status: res.status, ok: false, message: "Something went wrong." }
            }
        }

        return { status: res.status, ok: true }
    } catch {
        return { status: 500, ok: false, message: "Network error, please try again." }
    }
}
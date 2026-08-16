"use server"

export async function handleSendPasswordReset(email: string) {
    try {
        const res = await fetch(`${process.env.API_URL}/v1/staff/auth/reset`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                email,
            })
        })

        if (!res.ok) {
            return { status: res.status, ok: false, message: "Something went wrong." }
        }

        return { status: res.status, ok: true }
    } catch {
        return { status: 500, ok: false, message: "Network error, please try again." }
    }
}
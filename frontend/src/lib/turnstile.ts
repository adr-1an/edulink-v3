interface TurnstileVerifyResponse {
    success: boolean
    "error-codes"?: string[]
}

export async function verifyTurnstile(token: string) {
    const secret = process.env.TURNSTILE_SECRET_KEY

    if (!secret) {
        return { ok: false, status: 500, message: "Turnstile secret key is not configured." }
    }
    if (!token) {
        return { ok: false, status: 400, message: "Complete the verification challenge." }
    }

    let res
    try {
        res = await fetch("https://challenges.cloudflare.com/turnstile/v0/siteverify", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                secret,
                response: token,
            }),
        })
    } catch {
        return { ok: false, status: 502, message: "Unable to verify the challenge. Please try again." }
    }

    let data: TurnstileVerifyResponse
    try {
        data = await res.json()
    } catch {
        return { ok: false, status: 502, message: "Unable to verify the challenge. Please try again." }
    }

    if (res.ok && data.success) return { ok: true, status: 200 }

    return { ok: false, status: 400, message: "Verification failed. Please try again." }
}

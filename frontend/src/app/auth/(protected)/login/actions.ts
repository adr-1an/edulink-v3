"use server"

import {cookies} from "next/headers"
import {verifyTurnstile} from "@/lib/turnstile"

type LoginResponse = {
    token?: unknown
    twoFactorChallenge?: TwoFactorChallengeResponse
}

type TwoFactorChallengeResponse = {
    token?: unknown
    purpose?: unknown
    expiresAt?: unknown
}

function setSessionCookie(cookieStore: Awaited<ReturnType<typeof cookies>>, token: string) {
    cookieStore.set("token", token, {
        httpOnly: true,
        sameSite: "lax",
        secure: process.env.NODE_ENV === "production",
        path: "/",
    })
}

function isValidChallenge(data: TwoFactorChallengeResponse | undefined): data is {token: string, purpose: "login", expiresAt: string} {
    if (!data) return false
    return data.purpose === "login"
        && typeof data.token === "string"
        && data.token.length >= 32
        && typeof data.expiresAt === "string"
        && Number.isFinite(Date.parse(data.expiresAt))
}

export async function handleLogin(formData: FormData) {
    const cookieStore = await cookies()
    const turnstileToken = formData.get("cf-turnstile-response")?.toString() ?? ""
    const turnstile = await verifyTurnstile(turnstileToken)
    if (!turnstile.ok) return {status: turnstile.status, ok: false as const, code: "verification" as const}

    try {
        const res = await fetch(`${process.env.API_URL}/v1/staff/auth/login`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                email: formData.get("email"),
                password: formData.get("password"),
                stayLoggedIn: formData.get("stayLoggedIn") === "on",
            })
        })

        if (!res.ok) {
            if (res.status === 403) {
                return { status: 403, ok: false as const, code: "unverified" as const }
            } else if (res.status === 401) {
                return { status: 401, ok: false as const, code: "invalid" as const }
            }  else {
                return { status: res.status, ok: false as const, code: "generic" as const }
            }
        }

        const data = await res.json() as LoginResponse
        if (isValidChallenge(data.twoFactorChallenge)) {
            return {
                status: res.status,
                ok: true as const,
                step: "two_factor" as const,
                challenge: {
                    token: data.twoFactorChallenge.token,
                    expiresAt: data.twoFactorChallenge.expiresAt,
                },
            }
        }

        if (typeof data.token !== "string" || data.token.length === 0) {
            return {status: 502, ok: false as const, code: "invalid_response" as const}
        }

        setSessionCookie(cookieStore, data.token)

        return {status: res.status, ok: true as const, step: "complete" as const}
    } catch {
        return { status: 500, ok: false as const, code: "network" as const }
    }
}

export async function handleCompleteTwoFactorLogin(challengeToken: string, code: string) {
    const normalizedCode = code.trim()
    if (challengeToken.length < 32 || challengeToken.length > 256 || !/^\d{6}$/.test(normalizedCode)) {
        return {status: 422, ok: false as const, code: "invalid_input" as const}
    }

    try {
        const res = await fetch(`${process.env.API_URL}/v1/staff/auth/two-factor/challenge`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                code: normalizedCode,
                challengeToken,
            }),
        })

        if (!res.ok) {
            if (res.status === 401) {
                return {status: res.status, ok: false as const, code: "invalid_code" as const}
            }

            if (res.status === 404) {
                return {status: res.status, ok: false as const, code: "invalid_challenge" as const}
            }

            if (res.status === 422) {
                const data = await res.json().catch(() => null) as {code?: unknown} | null
                return {
                    status: res.status,
                    ok: false as const,
                    code: data?.code === "EXPIRED_TOKEN" ? "expired" as const : "invalid_input" as const,
                }
            }

            return {status: res.status, ok: false as const, code: "generic" as const}
        }

        const data = await res.json() as LoginResponse
        if (typeof data.token !== "string" || data.token.length === 0) {
            return {status: 502, ok: false as const, code: "invalid_response" as const}
        }

        const cookieStore = await cookies()
        setSessionCookie(cookieStore, data.token)
        return {status: res.status, ok: true as const}
    } catch {
        return {status: 500, ok: false as const, code: "network" as const}
    }
}

export async function handleRecoverTwoFactor(recoveryCode: string) {
    const normalizedCode = recoveryCode.trim()
    if (normalizedCode.length === 0 || normalizedCode.length > 256) {
        return {status: 422, ok: false as const, code: "invalid_input" as const}
    }

    try {
        const res = await fetch(`${process.env.API_URL}/v1/staff/auth/two-factor/recovery`, {
            method: "DELETE",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({recoveryCode: normalizedCode}),
        })

        if (res.ok) {
            return {status: res.status, ok: true as const}
        }

        if (res.status === 404) {
            return {status: res.status, ok: false as const, code: "invalid_code" as const}
        }

        if (res.status === 400 || res.status === 422) {
            return {status: res.status, ok: false as const, code: "invalid_input" as const}
        }

        return {status: res.status, ok: false as const, code: "generic" as const}
    } catch {
        return {status: 500, ok: false as const, code: "network" as const}
    }
}

export async function handleResendVerificationLink(email: string) {
    try {
        const res = await fetch(`${process.env.API_URL}/v1/staff/auth/verifications`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                email: email,
            })
        })

        if (!res.ok) {
            return { status: res.status, ok: false as const, code: "generic" as const }
        }

        return { status: res.status, ok: true }
    } catch {
        return { status: 500, ok: false as const, code: "network" as const }
    }
}

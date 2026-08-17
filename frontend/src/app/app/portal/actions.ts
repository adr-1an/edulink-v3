"use server"

import {cookies} from "next/headers"
import {portalHome} from "@/lib/portal_auth"
import {portalActivationErrorFromResponse} from "@/lib/portal_activation"
import {verifyTurnstile} from "@/lib/turnstile"

interface LoginResponse {
    token?: string
}

interface AuthResponse {
    user?: {accountType?: string}
}

async function establishPortalSession(token: string) {
    let authRes: Response
    try {
        authRes = await fetch(`${process.env.API_URL}/v1/portal/auth`, {
            headers: {Authorization: `Bearer ${token}`},
            cache: "no-store",
        })
    } catch {
        return {ok: false as const}
    }

    if (!authRes.ok) return {ok: false as const}

    let authData: AuthResponse
    try {
        authData = await authRes.json() as AuthResponse
    } catch {
        return {ok: false as const}
    }

    const accountType = authData.user?.accountType
    if (accountType !== "student" && accountType !== "guardian") {
        return {ok: false as const}
    }

    const cookieStore = await cookies()
    cookieStore.set("portal_token", token, {
        httpOnly: true,
        sameSite: "lax",
        secure: process.env.NODE_ENV === "production",
        path: "/",
        maxAge: 60 * 60 * 24 * 7,
    })

    return {ok: true as const, destination: portalHome(accountType)}
}

export async function handlePortalLogin(formData: FormData) {
    const turnstileToken = formData.get("cf-turnstile-response")?.toString() ?? ""
    const turnstile = await verifyTurnstile(turnstileToken)
    if (!turnstile.ok) {
        return {ok: false as const, status: turnstile.status, code: "verification" as const}
    }

    const email = formData.get("email")?.toString().trim().toLocaleLowerCase() ?? ""
    const password = formData.get("password")?.toString() ?? ""
    if (!email || !password) return {ok: false as const, status: 422, code: "invalid" as const}

    let loginRes: Response
    try {
        loginRes = await fetch(`${process.env.API_URL}/v1/portal/auth/login`, {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify({email, password}),
            cache: "no-store",
        })
    } catch {
        return {ok: false as const, status: 500, code: "generic" as const}
    }

    if (!loginRes.ok) {
        if (loginRes.status === 401) return {ok: false as const, status: 401, code: "credentials" as const}
        if (loginRes.status === 403) return {ok: false as const, status: 403, code: "disabled" as const}
        if (loginRes.status === 422) return {ok: false as const, status: 422, code: "invalid" as const}
        return {ok: false as const, status: loginRes.status, code: "generic" as const}
    }

    let loginData: LoginResponse
    try {
        loginData = await loginRes.json() as LoginResponse
    } catch {
        return {ok: false as const, status: 500, code: "generic" as const}
    }
    if (!loginData.token) return {ok: false as const, status: 500, code: "generic" as const}

    const session = await establishPortalSession(loginData.token)
    if (!session.ok) return {ok: false as const, status: 500, code: "generic" as const}

    return {ok: true as const, status: 200, destination: session.destination}
}

export async function handlePortalActivation(token: string, newPassword: string) {
    if (!token || newPassword.length < 8) {
        return {ok: false as const, status: 422, code: "invalid_password" as const}
    }

    let activationRes: Response
    try {
        activationRes = await fetch(`${process.env.API_URL}/v1/portal/auth/activate/${encodeURIComponent(token)}`, {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify({newPassword}),
            cache: "no-store",
        })
    } catch {
        return {ok: false as const, status: 500, code: "generic" as const}
    }

    if (!activationRes.ok) {
        let backendCode: unknown
        try {
            const data = await activationRes.json() as {code?: unknown}
            backendCode = data.code
        } catch {
            backendCode = undefined
        }

        return {
            ok: false as const,
            status: activationRes.status,
            code: portalActivationErrorFromResponse(activationRes.status, backendCode),
        }
    }

    let activationData: LoginResponse
    try {
        activationData = await activationRes.json() as LoginResponse
    } catch {
        return {ok: false as const, status: 500, code: "session" as const}
    }
    if (!activationData.token) return {ok: false as const, status: 500, code: "session" as const}

    const session = await establishPortalSession(activationData.token)
    if (!session.ok) return {ok: false as const, status: 500, code: "session" as const}

    return {ok: true as const, status: 200, destination: session.destination}
}

export async function handlePortalLogout() {
    const cookieStore = await cookies()
    const token = cookieStore.get("portal_token")?.value

    if (token) {
        let res: Response
        try {
            res = await fetch(`${process.env.API_URL}/v1/portal/auth`, {
                method: "DELETE",
                headers: {Authorization: `Bearer ${token}`},
                cache: "no-store",
            })
        } catch {
            return {ok: false as const, code: "network" as const}
        }

        if (!res.ok && res.status !== 401) {
            return {ok: false as const, code: "server" as const}
        }
    }

    cookieStore.delete("portal_token")
    return {ok: true as const}
}

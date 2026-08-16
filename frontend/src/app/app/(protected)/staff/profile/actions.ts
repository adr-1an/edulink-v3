"use server"

import {cookies} from "next/headers"
import {refresh} from "next/cache"
import {handleCompleteUpload} from "@/app/app/(protected)/staff/upload_actions"

export interface ProfilePictureUploadInit {
    id: string
    url: string
    completionToken: string
}

const allowedProfilePictureTypes = new Set([
    "image/jpeg",
    "image/png",
    "image/gif",
    "image/webp",
])

interface PrivacySettings {
    publicProfile: boolean
    staffInvitationsDisabled: boolean
}

interface ProfileUpdate extends PrivacySettings {
    name: string
    phone: string
}

async function updateProfile(payload: ProfileUpdate) {
    const cookieStore = await cookies()
    const token = cookieStore.get("token")?.value

    let errorCode = 0
    let res

    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/profile`, {
            method: "PATCH",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify(payload),
        })

        if (!res.ok) errorCode = res.status
    } catch {
        return { ok: false, message: "Network error, please try again."}
    }

    if (errorCode) {
        if (errorCode === 400) {
            return { ok: false, message: "Bad request." }
        } else if (errorCode === 422) {
            return { ok: false, message: "Invalid data." }
        } else if (errorCode === 500) {
            return { ok: false, message: "Internal server error." }
        } else {
            return { ok: false, message: "An unexpected error occurred." }
        }
    }

    refresh()
    return { ok: true }
}

export async function handleUpdateProfile(
    formData: FormData,
    privacySettings: PrivacySettings,
) {
    return updateProfile({
        name: formData.get("name")?.toString() ?? "",
        phone: formData.get("phone")?.toString() ?? "",
        ...privacySettings,
    })
}

export async function handleUpdatePrivacy(payload: ProfileUpdate) {
    return updateProfile(payload)
}

export async function handleInitProfilePictureUpload(file: {
    name: string
    size: number
    type: string
}) {
    if (!file.name || file.name.length > 255) {
        return {ok: false as const, message: "The file name is invalid."}
    }
    if (!Number.isInteger(file.size) || file.size <= 0 || file.size > 5 * 1024 * 1024) {
        return {ok: false as const, message: "Profile pictures can't be larger than 5 MB."}
    }
    if (!allowedProfilePictureTypes.has(file.type)) {
        return {ok: false as const, message: "Choose a JPEG, PNG, GIF, or WebP image."}
    }

    const token = (await cookies()).get("token")?.value
    let res: Response
    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/profile/profile-picture`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({
                name: file.name,
                declaredSize: file.size,
                declaredContentType: file.type,
            }),
        })
    } catch {
        return {ok: false as const, message: "Network error while preparing the profile picture."}
    }

    if (res.status === 401) return {ok: false as const, message: "Unauthorized."}
    if (res.status === 422) return {ok: false as const, message: "The image type or size was rejected."}
    if (!res.ok) {
        return {
            ok: false as const,
            message: res.status === 500
                ? "The server couldn't prepare the profile picture."
                : "Unable to prepare the profile picture.",
        }
    }

    try {
        const data = await res.json() as Partial<ProfilePictureUploadInit> & {id?: string | number}
        const id = typeof data.id === "string" || typeof data.id === "number" ? String(data.id) : ""
        if (
            !/^\d+$/.test(id)
            || typeof data.url !== "string"
            || !/^https?:\/\//.test(data.url)
            || typeof data.completionToken !== "string"
            || !data.completionToken
        ) {
            return {ok: false as const, message: "The server returned invalid upload details."}
        }

        return {
            ok: true as const,
            upload: {id, url: data.url, completionToken: data.completionToken},
        }
    } catch {
        return {ok: false as const, message: "The server returned invalid upload details."}
    }
}

export async function handleCompleteProfilePictureUpload(objectID: string, completionToken: string) {
    return handleCompleteUpload(objectID, completionToken)
}

export async function handleRemoveProfilePicture() {
    const token = (await cookies()).get("token")?.value
    let res: Response
    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/profile/profile-picture`, {
            method: "DELETE",
            headers: {Authorization: `Bearer ${token}`},
        })
    } catch {
        return {ok: false as const, message: "Network error while removing the profile picture."}
    }

    if (res.ok) {
        refresh()
        return {ok: true as const}
    }
    if (res.status === 401) return {ok: false as const, message: "Unauthorized."}
    if (res.status === 404) return {ok: false as const, message: "The profile picture no longer exists."}
    return {
        ok: false as const,
        message: res.status === 500
            ? "The server couldn't remove the profile picture."
            : "Unable to remove the profile picture.",
    }
}

export async function handleSendUpdateEmail(formData: FormData) {
    const cookieStore = await cookies()
    const token = cookieStore.get("token")?.value

    let errorCode = 0
    let res

    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/profile/email`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({
                newEmail: formData.get("email"),
                password: formData.get("password"),
            })
        })

        if (!res.ok) errorCode = res.status
    } catch {
        return { ok: false, message: "Network error, please try again." }
    }

    if (errorCode) {
        if (errorCode === 401) {
            return { ok: false, message: "Incorrect password." }
        } else if (errorCode === 500) {
            return { ok: false, message: "Internal server error." }
        } else if (res.status === 409) {
            return { ok: false, message: "This email is already taken." }
        } else {
            return { ok: false, message: "An unexpected error occurred." }
        }
    }

    refresh()
    return { ok: true }
}

export async function handleChangePassword(formData: FormData) {
    const cookieStore = await cookies()
    const token = cookieStore.get("token")?.value

    let errorCode = 0
    let res

    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/profile/password`, {
            method: "PUT",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({
                password: formData.get("password"),
                newPassword: formData.get("newPassword"),
            }),
        })

        if (!res.ok) errorCode = res.status
    } catch {
        return { ok: false, message: "Network error, please try again." }
    }

    if (errorCode) {
        if (errorCode === 401) {
            return { ok: false, message: "Incorrect password." }
        } else if (res.status === 500) {
            return { ok: false, message: "Internal server error." }
        } else if (res.status === 422) {
            return { ok: false, message: "Invalid data." }
        } else {
            return { ok: false, message: "An unexpected error occurred." }
        }
    }

    refresh()
    return { ok: true }
}

export async function handleStartTwoFactorSetup() {
    const token = (await cookies()).get("token")?.value
    let res: Response

    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/auth/two-factor`, {
            method: "POST",
            headers: {Authorization: `Bearer ${token}`},
            cache: "no-store",
        })
    } catch {
        return {ok: false as const, code: "network" as const}
    }

    if (res.status === 401 || res.status === 403) return {ok: false as const, code: "unauthorized" as const}
    if (res.status === 409) return {ok: false as const, code: "already_enabled" as const}
    if (!res.ok) return {ok: false as const, code: "server" as const}

    try {
        const data = await res.json() as {url?: unknown}
        if (typeof data.url !== "string" || !data.url.startsWith("otpauth://totp/")) {
            return {ok: false as const, code: "invalid_response" as const}
        }
        return {ok: true as const, setupURL: data.url}
    } catch {
        return {ok: false as const, code: "invalid_response" as const}
    }
}

export async function handleVerifyTwoFactorSetup(code: string) {
    const normalizedCode = code.trim()
    if (!/^\d{6}$/.test(normalizedCode)) return {ok: false as const, code: "invalid_input" as const}

    const token = (await cookies()).get("token")?.value
    let res: Response

    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/auth/two-factor`, {
            method: "PUT",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({code: normalizedCode}),
        })
    } catch {
        return {ok: false as const, code: "network" as const}
    }

    if (res.ok) {
        try {
            const data = await res.json() as {codes?: unknown}
            if (!data.codes || typeof data.codes !== "object" || Array.isArray(data.codes)) {
                return {ok: false as const, code: "invalid_response" as const}
            }

            const entries = Object.entries(data.codes)
            if (entries.length !== 8 || entries.some(([key]) => !/^\d+$/.test(key))) {
                return {ok: false as const, code: "invalid_response" as const}
            }

            const recoveryCodes: string[] = []
            for (const [, value] of entries.sort(([left], [right]) => Number(left) - Number(right))) {
                if (typeof value !== "string") return {ok: false as const, code: "invalid_response" as const}
                recoveryCodes.push(value)
            }

            if (new Set(recoveryCodes).size !== recoveryCodes.length || recoveryCodes.some((recoveryCode) => recoveryCode.length === 0)) {
                return {ok: false as const, code: "invalid_response" as const}
            }

            refresh()
            return {ok: true as const, recoveryCodes}
        } catch {
            return {ok: false as const, code: "invalid_response" as const}
        }
    }
    if (res.status === 401) return {ok: false as const, code: "invalid_code" as const}
    if (res.status === 403) return {ok: false as const, code: "setup_expired" as const}
    if (res.status === 400 || res.status === 422) return {ok: false as const, code: "invalid_input" as const}
    return {ok: false as const, code: "server" as const}
}

export async function handleDisableTwoFactor(password: string, code: string) {
    const normalizedCode = code.trim()
    if (!password || !/^\d{6}$/.test(normalizedCode)) {
        return {ok: false as const, code: "invalid_input" as const}
    }

    const token = (await cookies()).get("token")?.value
    let res: Response

    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/auth/two-factor`, {
            method: "DELETE",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({password, code: normalizedCode}),
        })
    } catch {
        return {ok: false as const, code: "network" as const}
    }

    if (res.ok) {
        refresh()
        return {ok: true as const}
    }
    if (res.status === 401) return {ok: false as const, code: "invalid_credentials" as const}
    if (res.status === 403) return {ok: false as const, code: "unavailable" as const}
    if (res.status === 400 || res.status === 422) return {ok: false as const, code: "invalid_input" as const}
    return {ok: false as const, code: "server" as const}
}

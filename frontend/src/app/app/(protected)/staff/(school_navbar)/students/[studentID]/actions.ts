"use server"

import {cookies} from "next/headers"
import {
    handleCompleteUpload, type UploadCompletionErrorCode,
} from "@/app/app/(protected)/staff/upload_actions"

const allowedProfilePictureTypes = new Set([
    "image/jpeg",
    "image/png",
    "image/gif",
    "image/webp",
])

export type StudentProfilePictureErrorCode =
    | UploadCompletionErrorCode
    | "invalid_student" | "invalid_file" | "conflict"

interface ProfilePictureUploadInit {
    id: string
    url: string
    completionToken: string
}

function failure<const Code extends StudentProfilePictureErrorCode>(code: Code) {
    return {ok: false as const, code}
}

export async function handleInitStudentProfilePictureUpload(studentID: string, file: {
    name: string
    size: number
    type: string
}) {
    if (!/^\d+$/.test(studentID)) return failure("invalid_student")
    if (!file.name || file.name.length > 255) return failure("invalid_file")
    if (!Number.isInteger(file.size) || file.size <= 0 || file.size > 5 * 1024 * 1024) {
        return failure("invalid_file")
    }
    if (!allowedProfilePictureTypes.has(file.type)) return failure("invalid_file")

    const token = (await cookies()).get("token")?.value
    let response: Response
    try {
        response = await fetch(`${process.env.API_URL}/v1/staff/students/${studentID}/profile-picture`, {
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
            cache: "no-store",
        })
    } catch {
        return failure("network")
    }

    if (response.status === 401) return failure("unauthorized")
    if (response.status === 403) return failure("forbidden")
    if (response.status === 409) return failure("conflict")
    if (response.status === 422) return failure("validation")
    if (!response.ok) return failure(response.status === 500 ? "server" : "upload")

    try {
        const data = await response.json() as Partial<ProfilePictureUploadInit> & {id?: string | number}
        const id = typeof data.id === "string" || typeof data.id === "number" ? String(data.id) : ""
        if (
            !/^\d+$/.test(id)
            || typeof data.url !== "string"
            || !/^https?:\/\//.test(data.url)
            || typeof data.completionToken !== "string"
            || !data.completionToken
        ) {
            return failure("upload")
        }

        return {
            ok: true as const,
            upload: {id, url: data.url, completionToken: data.completionToken},
        }
    } catch {
        return failure("upload")
    }
}

export async function handleCompleteStudentProfilePictureUpload(objectID: string, completionToken: string) {
    return handleCompleteUpload(objectID, completionToken)
}

export async function handleRemoveStudentProfilePicture(studentID: string) {
    if (!/^\d+$/.test(studentID)) return failure("invalid_student")

    const token = (await cookies()).get("token")?.value
    let response: Response
    try {
        response = await fetch(`${process.env.API_URL}/v1/staff/students/${studentID}/profile-picture`, {
            method: "DELETE",
            headers: {Authorization: `Bearer ${token}`},
            cache: "no-store",
        })
    } catch {
        return failure("network")
    }

    if (response.ok || response.status === 404) return {ok: true as const}
    if (response.status === 401) return failure("unauthorized")
    if (response.status === 403) return failure("forbidden")
    return failure(response.status === 500 ? "server" : "upload")
}

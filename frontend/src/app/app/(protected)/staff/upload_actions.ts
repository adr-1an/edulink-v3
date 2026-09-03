"use server"

import {cookies} from "next/headers"

export type UploadCompletionErrorCode =
    | "invalid_upload" | "network" | "unauthorized" | "forbidden" | "validation" | "server" | "upload"

function failure(code: UploadCompletionErrorCode, message: string) {
    return {ok: false as const, code, message}
}

export async function handleCompleteUpload(objectID: string, completionToken: string) {
    if (!/^\d+$/.test(objectID) || !completionToken) {
        return failure("invalid_upload", "The upload completion details are invalid.")
    }

    const token = (await cookies()).get("token")?.value
    let res: Response
    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/uploads/${objectID}`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({completionToken}),
        })
    } catch {
        return failure("network", "Network error while completing the upload.")
    }

    if (res.ok) return {ok: true as const}
    if (res.status === 401) return failure("unauthorized", "Unauthorized.")
    if (res.status === 403) return failure("forbidden", "The upload could not be verified.")
    if (res.status === 422) {
        return failure("validation", "The uploaded file did not match its declared size or type.")
    }
    return failure(
        res.status === 500 ? "server" : "upload",
        res.status === 500 ? "The server couldn't complete the upload." : "Unable to complete the upload.",
    )
}

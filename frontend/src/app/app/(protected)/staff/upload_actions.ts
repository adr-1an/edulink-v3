"use server"

import {cookies} from "next/headers"

export async function handleCompleteUpload(objectID: string, completionToken: string) {
    if (!/^\d+$/.test(objectID) || !completionToken) {
        return {ok: false as const, message: "The upload completion details are invalid."}
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
        return {ok: false as const, message: "Network error while completing the upload."}
    }

    if (res.ok) return {ok: true as const}
    if (res.status === 401) return {ok: false as const, message: "Unauthorized."}
    if (res.status === 403) return {ok: false as const, message: "The upload could not be verified."}
    if (res.status === 422) {
        return {ok: false as const, message: "The uploaded file did not match its declared size or type."}
    }
    return {
        ok: false as const,
        message: res.status === 500 ? "The server couldn't complete the upload." : "Unable to complete the upload.",
    }
}

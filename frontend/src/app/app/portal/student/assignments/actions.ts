"use server"

import {cookies} from "next/headers"

interface SubmissionResponse {
    submissionId?: string
    status?: string
}

interface UploadResponse {
    id?: string
    attachmentId?: string
    completionToken?: string
    url?: string
}

type SubmissionActionError = "invalid" | "unauthorized" | "forbidden" | "closed" | "not-found" | "network" | "server"

async function portalToken() {
    return (await cookies()).get("portal_token")?.value
}

function failure(status: number): {ok: false; code: SubmissionActionError} {
    if (status === 400 || status === 422) return {ok: false, code: "invalid"}
    if (status === 401) return {ok: false, code: "unauthorized"}
    if (status === 403) return {ok: false, code: "forbidden"}
    if (status === 404) return {ok: false, code: "not-found"}
    if (status === 409) return {ok: false, code: "closed"}
    return {ok: false, code: "server"}
}

export async function handleBeginAssignmentSubmission(assignmentID: string, notes: string) {
    if (!/^\d+$/.test(assignmentID) || notes.length > 2048) return failure(422)

    let res: Response
    try {
        res = await fetch(`${process.env.API_URL}/v1/portal/assignments/${assignmentID}/submissions`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${await portalToken()}`,
            },
            body: JSON.stringify({notes}),
            cache: "no-store",
        })
    } catch {
        return {ok: false as const, code: "network" as const}
    }

    if (!res.ok) return failure(res.status)

    try {
        const data = await res.json() as SubmissionResponse
        if (!data.submissionId || !/^\d+$/.test(data.submissionId)) return failure(500)
        return {ok: true as const, submissionID: data.submissionId}
    } catch {
        return failure(500)
    }
}

export async function handleInitSubmissionAttachment(
    submissionID: string,
    file: {fileName: string; declaredSize: number; declaredContentType: string},
) {
    if (
        !/^\d+$/.test(submissionID)
        || !file.fileName
        || !Number.isSafeInteger(file.declaredSize)
        || file.declaredSize <= 0
        || !file.declaredContentType
    ) return failure(422)

    let res: Response
    try {
        res = await fetch(`${process.env.API_URL}/v1/portal/submissions/${submissionID}/attachments`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${await portalToken()}`,
            },
            body: JSON.stringify(file),
            cache: "no-store",
        })
    } catch {
        return {ok: false as const, code: "network" as const}
    }

    if (!res.ok) return failure(res.status)

    try {
        const data = await res.json() as UploadResponse
        if (!data.id || !data.completionToken || !data.url) return failure(500)
        return {
            ok: true as const,
            upload: {
                objectID: data.id,
                completionToken: data.completionToken,
                url: data.url,
            },
        }
    } catch {
        return failure(500)
    }
}

export async function handleCompleteSubmissionAttachment(objectID: string, completionToken: string) {
    if (!/^\d+$/.test(objectID) || !completionToken) return failure(422)

    let res: Response
    try {
        res = await fetch(`${process.env.API_URL}/v1/portal/submissions/attachments/${objectID}`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${await portalToken()}`,
            },
            body: JSON.stringify({completionToken}),
            cache: "no-store",
        })
    } catch {
        return {ok: false as const, code: "network" as const}
    }

    return res.ok ? {ok: true as const} : failure(res.status)
}

export async function handleSubmitAssignment(submissionID: string) {
    if (!/^\d+$/.test(submissionID)) return failure(422)

    let res: Response
    try {
        res = await fetch(`${process.env.API_URL}/v1/portal/submissions/${submissionID}/submit`, {
            method: "POST",
            headers: {Authorization: `Bearer ${await portalToken()}`},
            cache: "no-store",
        })
    } catch {
        return {ok: false as const, code: "network" as const}
    }

    return res.ok ? {ok: true as const} : failure(res.status)
}

export async function handleDeleteSubmissionAttachment(submissionID: string, attachmentID: string) {
    if (!/^\d+$/.test(submissionID) || !/^\d+$/.test(attachmentID)) return failure(422)

    let res: Response
    try {
        res = await fetch(`${process.env.API_URL}/v1/portal/submissions/${submissionID}/attachments/${attachmentID}`, {
            method: "DELETE",
            headers: {Authorization: `Bearer ${await portalToken()}`},
            cache: "no-store",
        })
    } catch {
        return {ok: false as const, code: "network" as const}
    }

    return res.ok ? {ok: true as const} : failure(res.status)
}

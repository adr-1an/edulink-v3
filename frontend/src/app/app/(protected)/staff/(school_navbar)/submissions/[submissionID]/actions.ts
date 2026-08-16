"use server"

import {cookies} from "next/headers"

export type ReturnSubmissionError = "invalid" | "unauthorized" | "forbidden" | "server" | "network" | "unknown"
export type DeleteSubmissionError = ReturnSubmissionError | "conflict"
export type GradeSubmissionError = ReturnSubmissionError | "conflict" | "validation"

export interface SubmissionGradeInput {
    score: number
    notes: string
}

export async function handleGradeSubmission(submissionID: string, input: SubmissionGradeInput): Promise<
    {ok: true} | {ok: false; error: GradeSubmissionError}
> {
    if (!/^\d+$/.test(submissionID)) return {ok: false, error: "invalid"}
    if (!Number.isInteger(input.score) || input.score < 0 || input.score > 100 || input.notes.trim().length > 2048) {
        return {ok: false, error: "validation"}
    }

    const token = (await cookies()).get("token")?.value
    let response: Response
    try {
        response = await fetch(`${process.env.API_URL}/v1/staff/submissions/${submissionID}/grade`, {
            method: "POST",
            headers: {
                Authorization: `Bearer ${token}`,
                "Content-Type": "application/json",
            },
            body: JSON.stringify({score: input.score, notes: input.notes.trim()}),
        })
    } catch {
        return {ok: false, error: "network"}
    }

    if (response.ok) return {ok: true}
    if (response.status === 400) return {ok: false, error: "invalid"}
    if (response.status === 401) return {ok: false, error: "unauthorized"}
    if (response.status === 403) return {ok: false, error: "forbidden"}
    if (response.status === 409) return {ok: false, error: "conflict"}
    if (response.status === 422) return {ok: false, error: "validation"}
    if (response.status === 500) return {ok: false, error: "server"}
    return {ok: false, error: "unknown"}
}

export async function handleRemoveSubmissionGrade(submissionID: string): Promise<
    {ok: true} | {ok: false; error: ReturnSubmissionError}
> {
    if (!/^\d+$/.test(submissionID)) return {ok: false, error: "invalid"}

    const token = (await cookies()).get("token")?.value
    let response: Response
    try {
        response = await fetch(`${process.env.API_URL}/v1/staff/submissions/${submissionID}/grade`, {
            method: "DELETE",
            headers: {Authorization: `Bearer ${token}`},
        })
    } catch {
        return {ok: false, error: "network"}
    }

    if (response.ok) return {ok: true}
    if (response.status === 400) return {ok: false, error: "invalid"}
    if (response.status === 401) return {ok: false, error: "unauthorized"}
    if (response.status === 403) return {ok: false, error: "forbidden"}
    if (response.status === 500) return {ok: false, error: "server"}
    return {ok: false, error: "unknown"}
}

export async function handleReturnSubmission(submissionID: string): Promise<
    {ok: true} | {ok: false; error: ReturnSubmissionError}
> {
    if (!/^\d+$/.test(submissionID)) return {ok: false, error: "invalid"}

    const token = (await cookies()).get("token")?.value
    let response: Response
    try {
        response = await fetch(`${process.env.API_URL}/v1/staff/submissions/${submissionID}/return`, {
            method: "DELETE",
            headers: {Authorization: `Bearer ${token}`},
        })
    } catch {
        return {ok: false, error: "network"}
    }

    if (response.ok) return {ok: true}
    if (response.status === 400) return {ok: false, error: "invalid"}
    if (response.status === 401) return {ok: false, error: "unauthorized"}
    if (response.status === 403) return {ok: false, error: "forbidden"}
    if (response.status === 500) return {ok: false, error: "server"}
    return {ok: false, error: "unknown"}
}

export async function handleDeleteReturnedSubmission(submissionID: string): Promise<
    {ok: true} | {ok: false; error: DeleteSubmissionError}
> {
    if (!/^\d+$/.test(submissionID)) return {ok: false, error: "invalid"}

    const token = (await cookies()).get("token")?.value
    let response: Response
    try {
        response = await fetch(`${process.env.API_URL}/v1/staff/submissions/${submissionID}`, {
            method: "DELETE",
            headers: {Authorization: `Bearer ${token}`},
        })
    } catch {
        return {ok: false, error: "network"}
    }

    if (response.ok) return {ok: true}
    if (response.status === 400) return {ok: false, error: "invalid"}
    if (response.status === 401) return {ok: false, error: "unauthorized"}
    if (response.status === 403) return {ok: false, error: "forbidden"}
    if (response.status === 409) return {ok: false, error: "conflict"}
    if (response.status === 500) return {ok: false, error: "server"}
    return {ok: false, error: "unknown"}
}

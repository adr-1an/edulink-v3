"use server"


import {cookies} from "next/headers"

async function gradeRequest(
    path: string,
    method: "POST" | "PATCH" | "DELETE",
    payload?: {name: string; level: number},
) {
    const token = (await cookies()).get("token")?.value
    let res

    try {
        res = await fetch(`${process.env.API_URL}${path}`, {
            method,
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: payload ? JSON.stringify(payload) : undefined,
        })
    } catch {
        return {ok: false, message: "Network error, please try again."}
    }

    if (res.ok) return {ok: true}

    let code: string | undefined
    try { code = (await res.json()).code } catch {}

    if (code === "MISSING_LEVEL_VAR") return {ok: false, message: "Choose where the level should appear in the grade name."}
    if (code === "INVALID_NAME") return {ok: false, message: "The grade name is too long."}
    if (code === "LEVEL_OUT_OF_RANGE") return {ok: false, message: "The grade level must be between 0 and 20."}
    if (res.status === 401) return {ok: false, message: "Unauthorized."}
    if (res.status === 403) return {ok: false, message: "You don't have permission to manage grades."}
    if (res.status === 409) return {ok: false, message: "A grade with that name and level already exists."}
    if (res.status === 500) return {ok: false, message: "Internal server error."}
    return {ok: false, message: "Unable to update the grade."}
}

export async function handleCreateGrade(yearID: string, name: string, level: number) {
    if (!/^\d+$/.test(yearID)) return {ok: false, message: "Select a valid academic year."}
    return gradeRequest(`/v1/staff/academic-years/${yearID}/grades`, "POST", {name, level})
}

export async function handleUpdateGrade(gradeID: string, name: string, level: number) {
    if (!/^\d+$/.test(gradeID)) return {ok: false, message: "Invalid grade."}
    return gradeRequest(`/v1/staff/grades/${gradeID}`, "PATCH", {name, level})
}

export async function handleDeleteGrade(gradeID: string) {
    if (!/^\d+$/.test(gradeID)) return {ok: false, message: "Invalid grade."}
    return gradeRequest(`/v1/staff/grades/${gradeID}`, "DELETE")
}

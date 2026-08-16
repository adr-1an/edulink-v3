"use server"

import {cookies} from "next/headers"

interface CourseInput {
    name: string
    description: string
    color: string
}

function validateCourseInput(input: CourseInput) {
    const name = input.name.trim()
    const description = input.description.trim()
    const color = input.color.trim().replace(/^#/, "").toUpperCase()

    if (!name || name.length > 32) return {ok: false as const, message: "Course names must be between 1 and 32 characters."}
    if (description.length > 128) return {ok: false as const, message: "Course descriptions cannot exceed 128 characters."}
    if (!/^[0-9A-F]{6}$/.test(color)) return {ok: false as const, message: "Enter a valid six-character hex color."}
    return {ok: true as const, input: {name, description, color}}
}

async function courseRequest(path: string, method: "POST" | "PATCH" | "DELETE", input?: CourseInput) {
    const token = (await cookies()).get("token")?.value
    let res: Response
    try {
        res = await fetch(`${process.env.API_URL}${path}`, {
            method,
            headers: {
                ...(input ? {"Content-Type": "application/json"} : {}),
                Authorization: `Bearer ${token}`,
            },
            body: input ? JSON.stringify(input) : undefined,
        })
    } catch {
        return {ok: false, message: "Network error, please try again."}
    }

    if (res.ok) return {ok: true}
    if (res.status === 400) return {ok: false, message: "The course request was invalid."}
    if (res.status === 401) return {ok: false, message: "Unauthorized."}
    if (res.status === 403) return {ok: false, message: `You don't have permission to ${method === "POST" ? "create" : method === "PATCH" ? "update" : "delete"} courses.`}
    if (res.status === 409) return {ok: false, message: "A course with this name already exists in the grade."}
    if (res.status === 422) return {ok: false, message: "Check the course name, description, and color."}
    if (res.status === 500) return {ok: false, message: "Internal server error."}
    return {ok: false, message: "Unable to update the course."}
}

export async function handleCreateCourse(gradeID: string, input: CourseInput) {
    if (!/^\d+$/.test(gradeID)) return {ok: false, message: "Invalid grade."}
    const validated = validateCourseInput(input)
    if (!validated.ok) return validated
    return courseRequest(`/v1/staff/grades/${gradeID}/courses`, "POST", validated.input)
}

export async function handleUpdateCourse(courseID: string, input: CourseInput) {
    if (!/^\d+$/.test(courseID)) return {ok: false, message: "Invalid course."}
    const validated = validateCourseInput(input)
    if (!validated.ok) return validated
    return courseRequest(`/v1/staff/courses/${courseID}`, "PATCH", validated.input)
}

export async function handleDeleteCourse(courseID: string) {
    if (!/^\d+$/.test(courseID)) return {ok: false, message: "Invalid course."}
    return courseRequest(`/v1/staff/courses/${courseID}`, "DELETE")
}

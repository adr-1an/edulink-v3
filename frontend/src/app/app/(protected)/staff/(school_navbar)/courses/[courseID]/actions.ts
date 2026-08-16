"use server"

import {cookies} from "next/headers"
import {handleCompleteUpload} from "@/app/app/(protected)/staff/upload_actions"

export interface CoursePostInput {
    title: string
    body: string
    accentColor: string
    showUntil: string | null
}

export interface PostAttachmentUploadInit {
    id: string
    url: string
    completionToken: string
}

const allowedAttachmentTypes = new Set([
    "image/jpeg",
    "image/png",
    "image/gif",
    "image/webp",
    "application/zip",
    "application/pdf",
])

export interface CourseAssignmentInput {
    referencedPostId: string | null
    title: string
    description: string
    dueDate: string | null
    submissionsEnabled: boolean
    submissionsCloseAt: string | null
}

function validatePostInput(input: CoursePostInput) {
    const title = input.title.trim()
    const body = input.body.trim()
    const accentColor = input.accentColor.trim().replace(/^#/, "").toUpperCase()
    const showUntil = input.showUntil?.trim() || null

    if (!title || title.length > 32) return {ok: false as const, message: "Post titles must be between 1 and 32 characters."}
    if (!body || body.length > 2048) return {ok: false as const, message: "Post content must be between 1 and 2048 characters."}
    if (!/^[0-9A-F]{6}$/.test(accentColor)) return {ok: false as const, message: "Enter a valid six-character hex color."}
    if (showUntil && Number.isNaN(Date.parse(showUntil))) return {ok: false as const, message: "Enter a valid expiry date."}

    return {ok: true as const, input: {title, body, accentColor, showUntil}}
}

async function postRequest(path: string, method: "POST" | "PATCH" | "DELETE", input?: CoursePostInput) {
    const token = (await cookies()).get("token")?.value
    let res: Response

    try {
        res = await fetch(`${process.env.API_URL}${path}`, {
            method,
            headers: {
                ...(input ? {"Content-Type": "application/json"} : {}),
                Authorization: `Bearer ${token}`,
            },
            body: input ? JSON.stringify({
                title: input.title,
                body: input.body,
                accentColor: input.accentColor,
                showUntil: input.showUntil
                    ? {Time: input.showUntil, Valid: true}
                    : {Time: "0001-01-01T00:00:00Z", Valid: false},
            }) : undefined,
        })
    } catch {
        return {ok: false as const, message: "Network error, please try again."}
    }

    if (res.ok) {
        if (method !== "POST") return {ok: true as const}

        try {
            const data = await res.json() as {id?: string | number}
            const id = typeof data.id === "string" || typeof data.id === "number" ? String(data.id) : ""
            if (/^\d+$/.test(id)) return {ok: true as const, id}
        } catch {
            // The response error below is more useful than exposing JSON parsing details.
        }

        return {ok: false as const, message: "The post was created, but the server did not return its ID."}
    }
    if (res.status === 400) return {ok: false as const, message: "The post request was invalid."}
    if (res.status === 401) return {ok: false as const, message: "Unauthorized."}
    if (res.status === 403) return {ok: false as const, message: `You don't have permission to ${method === "POST" ? "create" : method === "PATCH" ? "edit" : "delete"} posts.`}
    if (res.status === 422) return {ok: false as const, message: "Check the post title, content, color, and expiry date."}
    if (res.status === 500) return {ok: false as const, message: "Internal server error."}
    return {ok: false as const, message: "Unable to update the post."}
}

export async function handleCreatePost(courseID: string, input: CoursePostInput) {
    if (!/^\d+$/.test(courseID)) return {ok: false as const, message: "Invalid course."}
    const validated = validatePostInput(input)
    if (!validated.ok) return validated
    return postRequest(`/v1/staff/courses/${courseID}/posts`, "POST", validated.input)
}

export async function handleUpdatePost(postID: string, input: CoursePostInput) {
    if (!/^\d+$/.test(postID)) return {ok: false as const, message: "Invalid post."}
    const validated = validatePostInput(input)
    if (!validated.ok) return validated
    return postRequest(`/v1/staff/course-posts/${postID}`, "PATCH", validated.input)
}

export async function handleDeletePost(postID: string) {
    if (!/^\d+$/.test(postID)) return {ok: false as const, message: "Invalid post."}
    return postRequest(`/v1/staff/course-posts/${postID}`, "DELETE")
}

export async function handleInitPostAttachmentUpload(postID: string, file: {
    name: string
    size: number
    type: string
}) {
    if (!/^\d+$/.test(postID)) return {ok: false as const, message: "Invalid post."}
    if (!file.name || file.name.length > 255) return {ok: false as const, message: "The file name is invalid."}
    if (!Number.isInteger(file.size) || file.size <= 0 || file.size > 5 * 1024 * 1024) {
        return {ok: false as const, message: "Files can't be larger than 5 MB."}
    }
    if (!allowedAttachmentTypes.has(file.type)) {
        return {ok: false as const, message: "This file type is not supported."}
    }

    const token = (await cookies()).get("token")?.value
    let res: Response
    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/course-posts/${postID}/upload`, {
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
        return {ok: false as const, message: "Network error while preparing the upload."}
    }

    if (res.status === 401) return {ok: false as const, message: "Unauthorized."}
    if (res.status === 403) return {ok: false as const, message: "You don't have permission to upload post attachments."}
    if (res.status === 422) return {ok: false as const, message: "The file name, size, or type was rejected."}
    if (!res.ok) return {ok: false as const, message: res.status === 500 ? "The server couldn't prepare the upload." : "Unable to prepare the upload."}

    try {
        const data = await res.json() as Partial<PostAttachmentUploadInit> & {id?: string | number}
        const id = typeof data.id === "string" || typeof data.id === "number" ? String(data.id) : ""
        if (!/^\d+$/.test(id) || typeof data.url !== "string" || !/^https?:\/\//.test(data.url) || typeof data.completionToken !== "string" || !data.completionToken) {
            return {ok: false as const, message: "The server returned an invalid upload response."}
        }
        return {ok: true as const, upload: {id, url: data.url, completionToken: data.completionToken}}
    } catch {
        return {ok: false as const, message: "The server returned an invalid upload response."}
    }
}

export async function handleCompletePostAttachmentUpload(objectID: string, completionToken: string) {
    return handleCompleteUpload(objectID, completionToken)
}

export async function handleDeletePostAttachment(postID: string, attachmentID: string) {
    if (!/^\d+$/.test(postID) || !/^\d+$/.test(attachmentID)) {
        return {ok: false as const, message: "The attachment details are invalid."}
    }

    const token = (await cookies()).get("token")?.value
    let res: Response
    try {
        res = await fetch(
            `${process.env.API_URL}/v1/staff/course-posts/attachments/${attachmentID}`,
            {
                method: "DELETE",
                headers: {Authorization: `Bearer ${token}`},
            },
        )
    } catch {
        return {ok: false as const, message: "Network error while deleting the attachment."}
    }

    if (res.ok) return {ok: true as const}
    if (res.status === 400) return {ok: false as const, message: "The attachment ID is invalid."}
    if (res.status === 401) return {ok: false as const, message: "Unauthorized."}
    if (res.status === 403) {
        return {ok: false as const, message: "You don't have permission to delete this attachment, or it no longer exists."}
    }
    return {
        ok: false as const,
        message: res.status === 500
            ? "The server couldn't delete the attachment."
            : "Unable to delete the attachment.",
    }
}

export async function handleCreateAssignment(courseID: string, input: CourseAssignmentInput) {
    if (!/^\d+$/.test(courseID)) return {ok: false as const, message: "Invalid course."}
    const validated = validateAssignmentInput(input, true)
    if (!validated.ok) return validated
    return assignmentRequest(`/v1/staff/courses/${courseID}/assignments`, "POST", validated.input)
}

function validateAssignmentInput(input: CourseAssignmentInput, requireFutureDates: boolean) {
    const normalized = {
        referencedPostId: input.referencedPostId?.trim() || null,
        title: input.title.trim(),
        description: input.description.trim(),
        dueDate: input.dueDate?.trim() || null,
        submissionsEnabled: input.submissionsEnabled,
        submissionsCloseAt: input.submissionsEnabled ? input.submissionsCloseAt?.trim() || null : null,
    }

    if (normalized.referencedPostId !== null && !/^\d+$/.test(normalized.referencedPostId)) {
        return {ok: false as const, message: "Choose a valid referenced post."}
    }
    if (normalized.title.length < 3 || normalized.title.length > 64) {
        return {ok: false as const, message: "Assignment titles must be between 3 and 64 characters."}
    }
    if (normalized.description.length > 4096) {
        return {ok: false as const, message: "Assignment descriptions cannot exceed 4096 characters."}
    }

    const now = Date.now()
    const dueTime = normalized.dueDate ? Date.parse(normalized.dueDate) : null
    const closeTime = normalized.submissionsCloseAt ? Date.parse(normalized.submissionsCloseAt) : null
    if (dueTime !== null && (Number.isNaN(dueTime) || requireFutureDates && dueTime <= now)) {
        return {ok: false as const, message: requireFutureDates ? "The due date must be in the future." : "Enter a valid due date."}
    }
    if (closeTime !== null && (Number.isNaN(closeTime) || requireFutureDates && closeTime <= now)) {
        return {ok: false as const, message: requireFutureDates ? "The submissions closing time must be in the future." : "Enter a valid submissions closing time."}
    }
    if (dueTime !== null && closeTime !== null && closeTime < dueTime) {
        return {ok: false as const, message: "Submissions cannot close before the assignment is due."}
    }

    return {ok: true as const, input: normalized}
}

async function assignmentRequest(path: string, method: "POST" | "PATCH" | "DELETE", input?: CourseAssignmentInput) {
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
        return {ok: false as const, message: "Network error, please try again."}
    }

    if (res.ok) return {ok: true as const}
    if (res.status === 400) return {ok: false as const, message: "The assignment request was invalid."}
    if (res.status === 401) return {ok: false as const, message: "Unauthorized."}
    if (res.status === 403) return {ok: false as const, message: `You don't have permission to ${method === "POST" ? "create" : method === "PATCH" ? "edit" : "delete"} assignments.`}
    if (res.status === 422) return {ok: false as const, message: "Check the assignment details and dates."}
    if (res.status === 500) return {ok: false as const, message: "Internal server error."}
    return {ok: false as const, message: "Unable to update the assignment."}
}

export async function handleUpdateAssignment(assignmentID: string, input: CourseAssignmentInput) {
    if (!/^\d+$/.test(assignmentID)) return {ok: false as const, message: "Invalid assignment."}
    const validated = validateAssignmentInput(input, false)
    if (!validated.ok) return validated
    return assignmentRequest(`/v1/staff/assignments/${assignmentID}`, "PATCH", validated.input)
}

export async function handleDeleteAssignment(assignmentID: string) {
    if (!/^\d+$/.test(assignmentID)) return {ok: false as const, message: "Invalid assignment."}
    return assignmentRequest(`/v1/staff/assignments/${assignmentID}`, "DELETE")
}

async function courseStudentRequest(courseID: string, studentID: string, method: "POST" | "DELETE") {
    if (!/^\d+$/.test(courseID) || !/^\d+$/.test(studentID)) {
        return {ok: false as const, message: "Invalid course or student."}
    }

    const token = (await cookies()).get("token")?.value
    let res: Response
    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/courses/${courseID}/students`, {
            method,
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({studentId: studentID}),
        })
    } catch {
        return {ok: false as const, message: "Network error, please try again."}
    }

    if (res.ok) return {ok: true as const}
    if (res.status === 400 || res.status === 422) return {ok: false as const, message: "The student request was invalid."}
    if (res.status === 401) return {ok: false as const, message: "Unauthorized."}
    if (res.status === 403) return {ok: false as const, message: `You don't have permission to ${method === "POST" ? "assign" : "remove"} course students.`}
    if (res.status === 500) return {ok: false as const, message: "Internal server error."}
    return {ok: false as const, message: "Unable to update the course roster."}
}

export async function handleAssignCourseStudent(courseID: string, studentID: string) {
    return courseStudentRequest(courseID, studentID, "POST")
}

export async function handleRemoveCourseStudent(courseID: string, studentID: string) {
    return courseStudentRequest(courseID, studentID, "DELETE")
}

import {cookies} from "next/headers"
import {redirect} from "next/navigation"
import {CircleX, GlobeX, ServerCrash} from "lucide-react"
import ErrorPage from "@/components/app/error"
import {getTranslations} from "@/i18n/server"
import {type PostAttachment} from "@/lib/post_attachments"
import {emptySchoolAccess, hasSchoolPermission, isSchoolAccess} from "@/lib/school_access"
import SubmissionClientPage, {type StaffSubmissionView} from "./client_page"

interface RawAttachment {
    id: string
    presignedUrl: string
    originalFilename: string
    declaredContentType: string
}

interface RawSubmission extends Omit<StaffSubmissionView, "attachments" | "grade"> {
    attachments: RawAttachment[]
    grade: unknown
}

interface SubmissionViewResponse {
    submission?: unknown
    access?: unknown
}

function isPortalStudent(value: unknown): value is StaffSubmissionView["submittedBy"] {
    if (!value || typeof value !== "object") return false
    const student = value as Partial<StaffSubmissionView["submittedBy"]>
    return typeof student.id === "string"
        && typeof student.email === "string"
        && typeof student.phone === "string"
        && typeof student.name === "string"
        && typeof student.lastName === "string"
}

function normalizeAttachment(value: unknown): PostAttachment | null {
    if (!value || typeof value !== "object") return null
    const attachment = value as Partial<RawAttachment>
    if (typeof attachment.id !== "string" || !/^\d+$/.test(attachment.id)) return null
    if (typeof attachment.presignedUrl !== "string") return null
    if (typeof attachment.originalFilename !== "string" || !attachment.originalFilename.trim() || attachment.originalFilename.length > 255) return null
    if (typeof attachment.declaredContentType !== "string" || !attachment.declaredContentType.trim() || attachment.declaredContentType.length > 255) return null

    try {
        const url = new URL(attachment.presignedUrl)
        if (url.protocol !== "https:" && url.protocol !== "http:") return null
    } catch {
        return null
    }

    return {
        id: attachment.id,
        presignedUrl: attachment.presignedUrl,
        fileName: attachment.originalFilename.trim(),
        contentType: attachment.declaredContentType.trim().toLocaleLowerCase(),
    }
}

function normalizeSubmission(value: unknown): StaffSubmissionView | null {
    if (!value || typeof value !== "object") return null
    const submission = value as Partial<RawSubmission>
    if (typeof submission.id !== "string" || !/^\d+$/.test(submission.id)) return null
    if (submission.status !== "submitted" && submission.status !== "returned") return null
    if (!isPortalStudent(submission.submittedBy)) return null
    if (submission.notes !== null && typeof submission.notes !== "string") return null
    if (typeof submission.submittedAt !== "string" || Number.isNaN(Date.parse(submission.submittedAt))) return null
    if (!Array.isArray(submission.attachments)) return null
    if (!("grade" in submission)) return null

    const attachments = submission.attachments.map(normalizeAttachment)
    if (attachments.some((attachment) => attachment === null)) return null
    const grade = submission.grade === null ? null : normalizeGrade(submission.grade)
    if (submission.grade !== null && !grade) return null

    return {
        id: submission.id,
        status: submission.status,
        submittedBy: submission.submittedBy,
        notes: submission.notes ?? null,
        submittedAt: submission.submittedAt,
        attachments: attachments as PostAttachment[],
        grade,
    }
}

function normalizeGrade(value: unknown): StaffSubmissionView["grade"] {
    if (!value || typeof value !== "object") return null
    const grade = value as Partial<NonNullable<StaffSubmissionView["grade"]>>
    if (typeof grade.score !== "number" || !Number.isInteger(grade.score) || grade.score < 0 || grade.score > 100) return null
    if (grade.notes !== null && typeof grade.notes !== "string") return null
    if (typeof grade.gradedAt !== "string" || Number.isNaN(Date.parse(grade.gradedAt))) return null
    return {score: grade.score, notes: grade.notes ?? null, gradedAt: grade.gradedAt}
}

export async function generateMetadata() {
    const {t} = await getTranslations()
    return {title: t("staff.submission.metaTitle")}
}

export default async function Page({params}: {params: Promise<{submissionID: string}>}) {
    const {t} = await getTranslations()
    const {submissionID} = await params
    if (!/^\d+$/.test(submissionID)) {
        return <ErrorPage message={t("staff.submission.error.invalidId")} icon={CircleX} />
    }

    const token = (await cookies()).get("token")?.value
    let response: Response
    try {
        response = await fetch(`${process.env.API_URL}/v1/staff/submissions/${submissionID}`, {
            headers: {Authorization: `Bearer ${token}`},
            cache: "no-store",
        })
    } catch {
        return <ErrorPage message={t("staff.submission.error.network")} icon={GlobeX} />
    }

    if (response.status === 401) redirect("/auth/login")
    if (response.status === 403) {
        return <ErrorPage message={t("staff.submission.error.forbidden")} icon={CircleX} />
    }
    if (!response.ok) {
        return (
            <ErrorPage
                message={response.status === 500 ? t("staff.submission.error.server") : t("staff.submission.error.load")}
                icon={response.status === 500 ? ServerCrash : CircleX}
            />
        )
    }

    let data: SubmissionViewResponse
    try {
        data = await response.json() as SubmissionViewResponse
    } catch {
        return <ErrorPage message={t("staff.submission.error.invalidResponse")} icon={CircleX} />
    }

    const submission = normalizeSubmission(data.submission)
    if (!submission) {
        return <ErrorPage message={t("staff.submission.error.invalidResponse")} icon={CircleX} />
    }

    const access = isSchoolAccess(data.access) ? data.access : emptySchoolAccess
    return (
        <SubmissionClientPage
            submission={submission}
            canReturn={hasSchoolPermission(access, "submission.return")}
            canDelete={hasSchoolPermission(access, "submission.delete")}
            canGrade={hasSchoolPermission(access, "submission.grade")}
            canRemoveGrade={hasSchoolPermission(access, "submission.removeGrade")}
        />
    )
}

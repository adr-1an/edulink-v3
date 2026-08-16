import {cookies} from "next/headers"
import {CircleX, GlobeX, ServerCrash} from "lucide-react"
import ErrorPage from "@/components/app/error"
import {getTranslations} from "@/i18n/server"
import {emptySchoolAccess, hasSchoolPermission, isSchoolAccess} from "@/lib/school_access"
import SubmissionsClientPage, {type AssignmentSubmission} from "./client_page"

interface SubmissionListResponse {
    submissions?: unknown
    access?: unknown
}

function isSubmission(value: unknown): value is AssignmentSubmission {
    if (!value || typeof value !== "object") return false
    const submission = value as Partial<AssignmentSubmission>
    const user = submission.submittedBy

    return typeof submission.id === "string"
        && (submission.status === "submitted" || submission.status === "returned")
        && typeof submission.submittedAt === "string"
        && !Number.isNaN(Date.parse(submission.submittedAt))
        && (submission.notes === null || typeof submission.notes === "string")
        && Boolean(user)
        && typeof user?.id === "string"
        && typeof user.email === "string"
        && typeof user.phone === "string"
        && typeof user.name === "string"
        && typeof user.lastName === "string"
}

export async function generateMetadata() {
    const {t} = await getTranslations()
    return {title: t("staff.submissions.metaTitle")}
}

export default async function Page({params}: {
    params: Promise<{assignmentID: string}>
}) {
    const {t} = await getTranslations()
    const {assignmentID} = await params

    if (!/^\d+$/.test(assignmentID)) {
        return <ErrorPage message={t("staff.submissions.error.invalidId")} icon={CircleX} />
    }

    const token = (await cookies()).get("token")?.value
    let response: Response

    try {
        response = await fetch(`${process.env.API_URL}/v1/staff/assignments/${assignmentID}/submissions`, {
            headers: {Authorization: `Bearer ${token}`},
            cache: "no-store",
        })
    } catch {
        return <ErrorPage message={t("staff.submissions.error.network")} icon={GlobeX} />
    }

    if (!response.ok) {
        if (response.status === 403) {
            return <ErrorPage message={t("staff.submissions.error.forbidden")} icon={CircleX} />
        }
        if (response.status === 500) {
            return <ErrorPage message={t("staff.submissions.error.server")} icon={ServerCrash} />
        }
        return <ErrorPage message={t("staff.submissions.error.load")} icon={CircleX} />
    }

    let data: SubmissionListResponse
    try {
        data = await response.json() as SubmissionListResponse
    } catch {
        return <ErrorPage message={t("staff.submissions.error.invalidResponse")} icon={CircleX} />
    }

    if (data.submissions !== null && data.submissions !== undefined && !Array.isArray(data.submissions)) {
        return <ErrorPage message={t("staff.submissions.error.invalidResponse")} icon={CircleX} />
    }

    const rawSubmissions = Array.isArray(data.submissions) ? data.submissions : []
    if (!rawSubmissions.every(isSubmission)) {
        return <ErrorPage message={t("staff.submissions.error.invalidResponse")} icon={CircleX} />
    }

    const submissions = [...rawSubmissions]
        .sort((first, second) => new Date(second.submittedAt).getTime() - new Date(first.submittedAt).getTime())
    const access = isSchoolAccess(data.access) ? data.access : emptySchoolAccess

    return (
        <SubmissionsClientPage
            submissions={submissions}
            canViewSubmissions={hasSchoolPermission(access, "submission.view")}
        />
    )
}

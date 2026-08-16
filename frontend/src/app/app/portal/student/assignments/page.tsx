import {cookies} from "next/headers"
import {redirect} from "next/navigation"
import {CalendarClock, CircleX, ClipboardList, GlobeX, GraduationCap, ServerCrash} from "lucide-react"
import ErrorPage from "@/components/app/error"
import AssignmentsSection, {
    type PortalCourseAssignment,
    type PortalSubmissionAttachment,
} from "../courses/[courseID]/assignments_section"
import {getTranslations} from "@/i18n/server"

export async function generateMetadata() {
    const {t} = await getTranslations()
    return {title: t("assignments.metaTitle")}
}

type AssignmentSubmission = NonNullable<PortalCourseAssignment["submission"]>

interface RawAssignmentSubmission extends Omit<AssignmentSubmission, "attachments"> {
    attachments?: unknown
}

interface RawPortalCourseAssignment extends Omit<PortalCourseAssignment, "submission"> {
    submission: RawAssignmentSubmission | null
}

interface AssignmentsResponse {
    assignments?: RawPortalCourseAssignment[] | null
}

function normalizeSubmissionAttachments(value: unknown): PortalSubmissionAttachment[] {
    if (!Array.isArray(value)) return []

    return value.flatMap((attachment) => {
        if (!attachment || typeof attachment !== "object") return []

        const id = "id" in attachment && typeof attachment.id === "string" ? attachment.id : ""
        const fileName = "fileName" in attachment && typeof attachment.fileName === "string" ? attachment.fileName.trim() : ""
        const fileSize = "fileSize" in attachment && typeof attachment.fileSize === "number" ? attachment.fileSize : -1
        const contentType = "contentType" in attachment && typeof attachment.contentType === "string" ? attachment.contentType.trim().toLocaleLowerCase() : ""
        let presignedUrl: string | null = null

        if ("presignedUrl" in attachment && typeof attachment.presignedUrl === "string") {
            try {
                const url = new URL(attachment.presignedUrl)
                if (url.protocol === "https:" || url.protocol === "http:") presignedUrl = attachment.presignedUrl
            } catch {
                // Keep the attachment metadata, but don't expose an invalid link.
            }
        }

        if (!/^\d+$/.test(id) || !fileName || fileName.length > 255 || !Number.isSafeInteger(fileSize) || fileSize < 0 || !contentType) {
            return []
        }

        return [{id, fileName, fileSize, contentType, presignedUrl}]
    })
}

export default async function Page() {
    const {t} = await getTranslations()
    const token = (await cookies()).get("portal_token")?.value
    let res: Response

    try {
        res = await fetch(`${process.env.API_URL}/v1/portal/assignments`, {
            headers: {Authorization: `Bearer ${token}`},
            cache: "no-store",
        })
    } catch {
        return <ErrorPage message={t("assignments.error.network")} icon={GlobeX} />
    }

    if (res.status === 401) redirect("/app/portal")
    if (!res.ok) {
        return (
            <ErrorPage
                message={res.status === 500 ? t("assignments.error.server") : t("assignments.error.load")}
                icon={res.status === 500 ? ServerCrash : CircleX}
            />
        )
    }

    let data: AssignmentsResponse
    try {
        data = await res.json() as AssignmentsResponse
    } catch {
        return <ErrorPage message={t("assignments.error.invalid")} icon={CircleX} />
    }

    const assignments: PortalCourseAssignment[] = (Array.isArray(data.assignments) ? data.assignments : []).map((assignment) => ({
        ...assignment,
        submission: assignment.submission ? {
            ...assignment.submission,
            attachments: normalizeSubmissionAttachments(assignment.submission.attachments),
        } : null,
    }))
    const courseCount = new Set(assignments.flatMap((assignment) => assignment.course ? [assignment.course.id] : [])).size
    const dueDateCount = assignments.filter((assignment) => assignment.dueDate !== null).length
    const referenceTime = new Date().toISOString()
    const referenceDate = referenceTime.slice(0, 10)

    return (
        <div className="space-y-6">
            <section className="overflow-hidden rounded-3xl border bg-card shadow-xs">
                <div className="relative overflow-hidden px-5 py-6 sm:px-7 sm:py-8">
                    <div className="pointer-events-none absolute -right-16 -top-24 size-56 rounded-full bg-primary/8 blur-3xl" />
                    <div className="relative flex items-start gap-4">
                        <div className="flex size-11 shrink-0 items-center justify-center rounded-2xl bg-primary/10 text-primary sm:size-12">
                            <ClipboardList className="size-5 sm:size-6" aria-hidden="true" />
                        </div>
                        <div>
                            <p className="text-sm font-medium text-primary">{t("assignments.workspace")}</p>
                            <h1 className="mt-1 text-2xl font-semibold tracking-tight sm:text-3xl">{t("assignments.allTitle")}</h1>
                            <p className="mt-2 max-w-2xl text-sm leading-6 text-muted-foreground sm:text-base">
                                {t("assignments.allDescription")}
                            </p>
                        </div>
                    </div>
                </div>

                <div className="grid border-t sm:grid-cols-3 sm:divide-x">
                    <Stat icon={ClipboardList} value={assignments.length} label={t("assignments.title")} />
                    <Stat icon={GraduationCap} value={courseCount} label={t("portal.courses")} />
                    <Stat icon={CalendarClock} value={dueDateCount} label={t("assignments.withDueDates")} />
                </div>
            </section>

            <AssignmentsSection assignments={assignments} referenceDate={referenceDate} referenceTime={referenceTime} showCourseFilter showHeading={false} />
        </div>
    )
}

function Stat({icon: Icon, value, label}: {
    icon: typeof ClipboardList
    value: number
    label: string
}) {
    return (
        <div className="flex items-center gap-3 border-t px-5 py-4 first:border-t-0 sm:border-t-0 sm:px-6">
            <Icon className="size-4 text-muted-foreground" />
            <div>
                <p className="text-xl font-semibold tabular-nums">{value}</p>
                <p className="text-xs text-muted-foreground">{label}</p>
            </div>
        </div>
    )
}

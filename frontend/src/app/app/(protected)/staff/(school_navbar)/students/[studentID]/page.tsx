import {cookies} from "next/headers"
import {redirect} from "next/navigation"
import {CircleX, GlobeX, ServerCrash} from "lucide-react"
import ErrorPage from "@/components/app/error"
import {getTranslations} from "@/i18n/server"
import {emptySchoolAccess, isSchoolAccess, type SchoolAccess} from "@/lib/school_access"
import StudentProfileClientPage, {type StaffStudentProfile} from "./client_page"
import {type StudentAssignmentSubmission} from "./client_page"

interface StudentViewResponse {
    student?: unknown
    access?: unknown
    assignmentSubmissions?: unknown
}

function normalizeDateOnly(value: unknown) {
    if (typeof value !== "string") return null
    return value.match(/^\d{4}-\d{2}-\d{2}/)?.[0] ?? null
}

function normalizeStudent(value: unknown): StaffStudentProfile | null {
    if (!value || typeof value !== "object") return null
    const student = value as Partial<StaffStudentProfile>
    const dateOfBirth = normalizeDateOnly(student.dateOfBirth)

    if (typeof student.id !== "string" || !/^\d+$/.test(student.id)) return null
    if (typeof student.name !== "string" || !student.name.trim()) return null
    if (typeof student.lastName !== "string" || !student.lastName.trim()) return null
    if (!dateOfBirth) return null
    if (typeof student.email !== "string" || !student.email.trim()) return null
    if (student.phone !== null && typeof student.phone !== "string") return null
    if (student.notes !== null && typeof student.notes !== "string") return null
    if (typeof student.accountEnabled !== "boolean") return null
    if (typeof student.createdAt !== "string" || Number.isNaN(Date.parse(student.createdAt))) return null

    return {
        id: student.id,
        name: student.name.trim(),
        lastName: student.lastName.trim(),
        dateOfBirth,
        email: student.email.trim(),
        phone: student.phone ?? "",
        notes: student.notes ?? "",
        accountEnabled: student.accountEnabled,
        createdAt: student.createdAt,
    }
}

function isRecord(value: unknown): value is Record<string, unknown> {
    return Boolean(value && typeof value === "object" && !Array.isArray(value))
}

function isNumericID(value: unknown): value is string {
    return typeof value === "string" && /^\d+$/.test(value)
}

function normalizeSubmission(value: unknown): StudentAssignmentSubmission | null {
    if (!isRecord(value) || !isNumericID(value.id)) return null
    if (typeof value.submittedAt !== "string" || Number.isNaN(Date.parse(value.submittedAt))) return null
    if (value.notes !== null && typeof value.notes !== "string") return null
    if (!isRecord(value.assignment) || !isNumericID(value.assignment.id)) return null
    if (typeof value.assignment.title !== "string" || !value.assignment.title.trim()) return null
    if (value.assignment.description !== null && typeof value.assignment.description !== "string") return null
    if (!isRecord(value.assignment.course) || !isNumericID(value.assignment.course.id)) return null
    if (typeof value.assignment.course.name !== "string" || !value.assignment.course.name.trim()) return null
    if (typeof value.assignment.course.accentColor !== "string") return null
    if (!isRecord(value.assignment.course.grade) || !isNumericID(value.assignment.course.grade.id)) return null
    if (typeof value.assignment.course.grade.name !== "string" || !value.assignment.course.grade.name.trim()) return null
    if (typeof value.assignment.course.grade.level !== "number" || !Number.isInteger(value.assignment.course.grade.level)) return null

    let score: StudentAssignmentSubmission["score"] = null
    if (value.score !== null) {
        if (!isRecord(value.score)
            || typeof value.score.scorePercentage !== "number"
            || !Number.isInteger(value.score.scorePercentage)
            || value.score.scorePercentage < 0
            || value.score.scorePercentage > 100) return null
        score = {scorePercentage: value.score.scorePercentage}
    }

    return {
        id: value.id,
        submittedAt: value.submittedAt,
        notes: value.notes ?? null,
        score,
        assignment: {
            id: value.assignment.id,
            title: value.assignment.title.trim(),
            description: value.assignment.description ?? null,
            course: {
                id: value.assignment.course.id,
                name: value.assignment.course.name.trim(),
                accentColor: value.assignment.course.accentColor,
                grade: {
                    id: value.assignment.course.grade.id,
                    name: value.assignment.course.grade.name.trim(),
                    level: value.assignment.course.grade.level,
                },
            },
        },
    }
}

function normalizeSubmissions(value: unknown): StudentAssignmentSubmission[] | null {
    if (value === null) return []
    if (!Array.isArray(value)) return null

    const submissions = value.map(normalizeSubmission)
    if (submissions.some((submission) => submission === null)) return null
    const normalized = submissions as StudentAssignmentSubmission[]
    if (new Set(normalized.map((submission) => submission.id)).size !== normalized.length) return null
    return normalized
}

export async function generateMetadata() {
    const {t} = await getTranslations()
    return {title: t("staff.studentProfile.metaTitle")}
}

export default async function Page({params}: {params: Promise<{studentID: string}>}) {
    const {t} = await getTranslations()
    const {studentID} = await params
    if (!/^\d+$/.test(studentID)) {
        return <ErrorPage message={t("staff.studentProfile.error.invalidId")} icon={CircleX} />
    }

    const token = (await cookies()).get("token")?.value
    let response: Response
    try {
        response = await fetch(`${process.env.API_URL}/v1/staff/students/${studentID}`, {
            headers: {Authorization: `Bearer ${token}`},
            cache: "no-store",
        })
    } catch {
        return <ErrorPage message={t("staff.studentProfile.error.network")} icon={GlobeX} />
    }

    if (response.status === 401) redirect("/auth/login")
    if (response.status === 403) {
        return <ErrorPage message={t("staff.studentProfile.error.forbidden")} icon={CircleX} />
    }
    if (!response.ok) {
        return (
            <ErrorPage
                message={response.status === 500
                    ? t("staff.studentProfile.error.server")
                    : t("staff.studentProfile.error.load")}
                icon={response.status === 500 ? ServerCrash : CircleX}
            />
        )
    }

    let data: StudentViewResponse
    try {
        data = await response.json() as StudentViewResponse
    } catch {
        return <ErrorPage message={t("staff.studentProfile.error.invalidResponse")} icon={CircleX} />
    }

    const student = normalizeStudent(data.student)
    const assignmentSubmissions = normalizeSubmissions(data.assignmentSubmissions)
    if (!student || !assignmentSubmissions) {
        return <ErrorPage message={t("staff.studentProfile.error.invalidResponse")} icon={CircleX} />
    }

    const access: SchoolAccess = isSchoolAccess(data.access) ? data.access : emptySchoolAccess
    return (
        <StudentProfileClientPage
            initialStudent={student}
            assignmentSubmissions={assignmentSubmissions}
            access={access}
        />
    )
}

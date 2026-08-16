import {cookies} from "next/headers"
import {CircleX, GlobeX, ServerCrash} from "lucide-react"
import ErrorPage from "@/components/app/error"
import {emptySchoolAccess, type SchoolAccess} from "@/lib/school_access"
import StudentsClientPage, {type Student} from "./client_page"
import {getTranslations} from "@/i18n/server"
import {type Locale} from "@/i18n/config"

export async function generateMetadata() {
    const {t} = await getTranslations()
    return {title: t("staff.students.metaTitle")}
}

interface RawStudent {
    id: string
    name: string
    last_name: string
    dateOfBirth: string | null
    email: string | null
    phone: string | null
    notes: string | null
    accountEnabled: boolean
    createdAt: string
}

interface StudentListResponse {
    students?: RawStudent[] | null
    access?: SchoolAccess
}

function dateOnly(value: string | null) {
    if (!value) return null
    const match = value.match(/^\d{4}-\d{2}-\d{2}/)
    return match?.[0] ?? null
}

function formatDateOnly(value: string | null, locale: Locale) {
    if (!value) return null
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return null
    return new Intl.DateTimeFormat(locale === "pl" ? "pl-PL" : "en-US", {day: "numeric", month: "short", year: "numeric", timeZone: "UTC"}).format(date)
}

export default async function Page({params}: {params: Promise<{id: string}>}) {
    const {locale, t} = await getTranslations()
    const {id} = await params
    const token = (await cookies()).get("token")?.value
    let res: Response

    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/schools/${id}/students`, {
            headers: {Authorization: `Bearer ${token}`},
            cache: "no-store",
        })
    } catch {
        return <ErrorPage message={t("staff.students.error.network")} icon={GlobeX} />
    }

    if (!res.ok) {
        if (res.status === 403) return <ErrorPage message={t("staff.students.error.pageForbidden")} icon={CircleX} />
        if (res.status === 500) return <ErrorPage message={t("staff.students.error.server")} icon={ServerCrash} />
        return <ErrorPage message={t("staff.students.error.pageLoad")} icon={CircleX} />
    }

    const data = await res.json() as StudentListResponse
    const students: Student[] = (Array.isArray(data.students) ? data.students : []).map((student) => ({
        id: student.id,
        name: student.name,
        lastName: student.last_name,
        dateOfBirth: dateOnly(student.dateOfBirth),
        dateOfBirthLabel: formatDateOnly(student.dateOfBirth, locale),
        email: student.email ?? "",
        phone: student.phone ?? "",
        notes: student.notes ?? "",
        accountEnabled: student.accountEnabled,
        createdAt: student.createdAt,
    }))

    const snapshot = students.map((student) => `${student.id}:${student.name}:${student.lastName}:${student.email}:${student.accountEnabled}`).join("|")
    return <StudentsClientPage key={snapshot} schoolID={id} initialStudents={students} access={data.access ?? emptySchoolAccess} />
}

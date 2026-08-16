import {cookies} from "next/headers"
import ErrorPage from "@/components/app/error"
import {CircleX, GlobeX, ServerCrash} from "lucide-react"
import ClientPage from "@/app/app/(protected)/staff/(school_navbar)/schools/[id]/client_page"
import {emptySchoolAccess, type SchoolAccess} from "@/lib/school_access"
import {getTranslations} from "@/i18n/server"

export async function generateMetadata() {
    const {t} = await getTranslations()
    return {title: t("staff.dashboard.metaTitle")}
}

export default async function Page({ params }: { params: Promise<{ id: string }>}) {
    const {locale, t} = await getTranslations()
    const cookieStore = await cookies()
    const token = cookieStore.get("token")?.value

    const { id } = await params

    // Get the school dashboard
    let errorCode = 0
    let res

    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/schools/${id}`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            }
        })

        if (!res.ok) errorCode = res.status
    } catch {
        return <ErrorPage message={t("staff.dashboard.error.network")} icon={GlobeX} />
    }

    if (errorCode) {
        if (errorCode === 500) {
            return <ErrorPage message={t("staff.dashboard.error.server")} icon={ServerCrash} />
        } else if (errorCode === 403) {
            return <ErrorPage message={t("staff.dashboard.error.forbidden")} icon={CircleX} />
        } else if (errorCode === 404) {
            return <ErrorPage message={t("staff.dashboard.error.notFound")} icon={CircleX} />
        } else {
            return <ErrorPage message={t("staff.dashboard.error.unexpected")} icon={CircleX} />
        }
    }

    const data = await res.json()

    let academicYears = []
    let access: SchoolAccess = data.access ?? emptySchoolAccess
    let canListGrades = false
    let canListAcademicYears = false
    try {
        const headers = {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`,
        }
        const [yearsRes, gradesRes] = await Promise.all([
            fetch(`${process.env.API_URL}/v1/staff/schools/${id}/academic-years`, {method: "GET", headers}),
            fetch(`${process.env.API_URL}/v1/staff/schools/${id}/grades`, {method: "GET", headers}),
        ])

        if (yearsRes.ok) {
            const yearsData = await yearsRes.json()
            academicYears = yearsData.academicYears ?? []
            access = yearsData.access ?? access
            canListAcademicYears = true
        }

        if (gradesRes.ok) {
            const gradesData = await gradesRes.json()
            canListGrades = true
            access = gradesData.access ?? access
        }
    } catch {
        // The dashboard can still render if its permission-aware supporting data is unavailable.
    }
    const displayNames = new Intl.DisplayNames([locale === "pl" ? "pl-PL" : "en-US"], {type: "region"})
    const regionName = /^[A-Z]{2}$/i.test(data.school.regionCode)
        ? displayNames.of(data.school.regionCode.toUpperCase()) ?? data.school.regionCode
        : data.school.regionCode

    return (
        <ClientPage
            school={data.school}
            regionName={regionName}
            academicYears={academicYears}
            access={access}
            canListGrades={canListGrades}
            canListAcademicYears={canListAcademicYears}
        />
    )
}

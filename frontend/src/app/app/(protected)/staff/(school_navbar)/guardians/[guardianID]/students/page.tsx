import {cookies} from "next/headers"
import {redirect} from "next/navigation"
import {CircleX, GlobeX} from "lucide-react"
import ErrorPage from "@/components/app/error"
import {getTranslations} from "@/i18n/server"
import {parseGuardianStudentsJSON} from "@/lib/guardian_students"
import {emptySchoolAccess, isSchoolAccess} from "@/lib/school_access"
import GuardianStudentsPage from "./client_page"

export async function generateMetadata() {
    const {t} = await getTranslations()
    return {title: t("staff.guardianStudents.title")}
}

export default async function Page({params}: {params: Promise<{guardianID: string}>}) {
    const {guardianID} = await params
    const {t} = await getTranslations()
    if (!/^\d+$/.test(guardianID)) return <ErrorPage message={t("staff.guardianStudents.error.invalid")} icon={CircleX} />
    const token = (await cookies()).get("token")?.value
    let response: Response
    try {
        response = await fetch(`${process.env.API_URL}/v1/staff/guardians/${guardianID}`, {
            headers: {Authorization: `Bearer ${token}`},
            cache: "no-store",
        })
    } catch {
        return <ErrorPage message={t("staff.guardianStudents.error.network")} icon={GlobeX} />
    }
    if (response.status === 401) redirect("/auth/login")
    if (!response.ok) {
        const message = response.status === 403 ? "staff.guardianStudents.error.forbidden"
            : response.status === 400 ? "staff.guardianStudents.error.invalid" : "staff.guardianStudents.error.server"
        return <ErrorPage message={t(message)} icon={CircleX} />
    }
    const data = parseGuardianStudentsJSON(await response.text())
    if (!data) return <ErrorPage message={t("staff.guardianStudents.error.invalidResponse")} icon={CircleX} />
    return <GuardianStudentsPage key={guardianID} guardianID={guardianID} students={data.students} access={isSchoolAccess(data.access) ? data.access : emptySchoolAccess} />
}

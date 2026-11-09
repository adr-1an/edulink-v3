import {cookies} from "next/headers"
import {redirect} from "next/navigation"
import {CircleX, GlobeX} from "lucide-react"
import ErrorPage from "@/components/app/error"
import {getTranslations} from "@/i18n/server"
import {parseStudentGuardiansResponseJSON} from "@/lib/student_guardians"
import {normalizeGuardianList} from "@/lib/guardian"
import {emptySchoolAccess, isSchoolAccess} from "@/lib/school_access"
import StudentGuardiansPage from "./client_page"

export async function generateMetadata() {
    const {t} = await getTranslations()
    return {title: t("staff.studentGuardians.metaTitle")}
}

export default async function Page({params}: {
    params: Promise<{studentID: string}>
}) {
    const {studentID} = await params
    const {t} = await getTranslations()
    if (!/^\d+$/.test(studentID)) return <ErrorPage message={t("staff.studentGuardians.error.invalid")} icon={CircleX} />
    const token = (await cookies()).get("token")?.value
    const load = async (path: string) => {
        try {
            return await fetch(`${process.env.API_URL}${path}`, {headers: {Authorization: `Bearer ${token}`}, cache: "no-store"})
        } catch {
            return null
        }
    }
    const studentResponse = await load(`/v1/staff/students/${studentID}`)
    if (!studentResponse) return <ErrorPage message={t("staff.studentGuardians.error.network")} icon={GlobeX} />
    if (studentResponse.status === 401) redirect("/auth/login")
    if (!studentResponse.ok) return <ErrorPage message={t(studentResponse.status === 403 ? "staff.studentGuardians.error.forbidden" : "staff.studentGuardians.error.server")} icon={CircleX} />
    const studentBody = await studentResponse.text()
    const assigned = parseStudentGuardiansResponseJSON(studentBody)
    if (!assigned) return <ErrorPage message={t("staff.studentGuardians.error.invalidResponse")} icon={CircleX} />
    const studentData = JSON.parse(studentBody)
    const read = async (response: Response | null) => {
        if (!response?.ok) return null
        try { return await response.json() } catch { return null }
    }
    const {schoolID} = assigned
    const availableData = await load(`/v1/staff/schools/${schoolID}/guardians`).then(read)
    const access = isSchoolAccess(assigned.access) ? assigned.access
        : isSchoolAccess(studentData?.access) ? studentData.access : emptySchoolAccess
    const available = normalizeGuardianList(availableData?.guardians)
    const candidatesUnavailable = !schoolID || !availableData || !Array.isArray(availableData.guardians)
        || available.length !== availableData.guardians.length
    const studentName = typeof studentData?.student?.name === "string" && typeof studentData?.student?.lastName === "string"
        ? `${studentData.student.name} ${studentData.student.lastName}` : null
    return <StudentGuardiansPage studentID={studentID} schoolID={schoolID} studentName={studentName} initialGuardians={assigned.guardians}
        availableGuardians={candidatesUnavailable ? [] : available} candidatesUnavailable={candidatesUnavailable} access={access} />
}

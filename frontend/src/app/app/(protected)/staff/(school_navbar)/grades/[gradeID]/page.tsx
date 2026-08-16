import {cookies} from "next/headers"
import {CircleX, GlobeX, ServerCrash} from "lucide-react"
import ErrorPage from "@/components/app/error"
import CoursesClientPage from "./client_page"
import {emptySchoolAccess} from "@/lib/school_access"
import {getTranslations} from "@/i18n/server"

export async function generateMetadata() {
    const {t} = await getTranslations()
    return {title: t("staff.courses.metaTitle")}
}

export default async function Page({params}: {
    params: Promise<{gradeID: string}>
}) {
    const {t} = await getTranslations()
    const {gradeID} = await params
    const token = (await cookies()).get("token")?.value
    let coursesRes: Response

    try {
        coursesRes = await fetch(`${process.env.API_URL}/v1/staff/grades/${gradeID}/courses`, {
            headers: {Authorization: `Bearer ${token}`},
        })
    } catch {
        return <ErrorPage message={t("staff.courses.error.network")} icon={GlobeX} />
    }

    if (!coursesRes.ok) {
        if (coursesRes.status === 403) return <ErrorPage message={t("staff.courses.error.forbidden")} icon={CircleX} />
        if (coursesRes.status === 500) return <ErrorPage message={t("staff.courses.error.server")} icon={ServerCrash} />
        return <ErrorPage message={t("staff.courses.error.load")} icon={CircleX} />
    }

    const {courses, access} = await coursesRes.json()
    const courseList = courses ?? []

    return (
        <CoursesClientPage
            key={courseList.map((course: {id: string; name: string; description: string; color: string}) => `${course.id}:${course.name}:${course.description}:${course.color}`).join("|")}
            gradeID={gradeID}
            courses={courseList}
            access={access ?? emptySchoolAccess}
        />
    )
}

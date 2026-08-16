import {cookies} from "next/headers"
import {redirect} from "next/navigation"
import StudentPortalPage, {type PortalCourse} from "./client_page"
import {getTranslations} from "@/i18n/server"

export async function generateMetadata() {
    const {t} = await getTranslations()
    return {title: t("portal.student.metaTitle")}
}

interface CourseResponse {
    courses?: PortalCourse[] | null
}

export default async function Page() {
    const {t} = await getTranslations()
    const token = (await cookies()).get("portal_token")?.value
    let res: Response

    try {
        res = await fetch(`${process.env.API_URL}/v1/portal/courses`, {
            headers: {Authorization: `Bearer ${token}`},
            cache: "no-store",
        })
    } catch {
        return <StudentPortalPage initialCourses={[]} error={t("portal.student.error.network")} />
    }

    if (!res.ok) {
        if (res.status === 401) redirect("/app/portal")
        const message = res.status === 500
            ? t("portal.student.error.server")
            : t("portal.student.error.generic")
        return <StudentPortalPage initialCourses={[]} error={message} />
    }

    let data: CourseResponse
    try {
        data = await res.json() as CourseResponse
    } catch {
        return <StudentPortalPage initialCourses={[]} error={t("portal.student.error.invalid")} />
    }

    return <StudentPortalPage initialCourses={Array.isArray(data.courses) ? data.courses : []} />
}

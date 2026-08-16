import {cookies} from "next/headers"
import ErrorPage from "@/components/app/error"
import {CircleX, GlobeX, ServerCrash} from "lucide-react"
import SettingsClientPage from "./client_page"
import {getSchoolNavigationAccess} from "../../../school_navigation_actions"
import {getTranslations} from "@/i18n/server"
import countries from "world-countries"

export async function generateMetadata() {
    const {t} = await getTranslations()
    return {title: t("staff.settings.metaTitle")}
}

export default async function Page({params}: {params: Promise<{id: string}>}) {
    const {locale, t} = await getTranslations()
    const {id} = await params
    const token = (await cookies()).get("token")?.value

    let schoolRes: Response
    let yearsRes: Response
    try {
        ;[schoolRes, yearsRes] = await Promise.all([
            fetch(`${process.env.API_URL}/v1/staff/schools/${id}`, {headers: {Authorization: `Bearer ${token}`}}),
            fetch(`${process.env.API_URL}/v1/staff/schools/${id}/academic-years`, {headers: {Authorization: `Bearer ${token}`}}),
        ])
    } catch {
        return <ErrorPage message={t("staff.settings.error.network")} icon={GlobeX} />
    }

    if (!schoolRes.ok) {
        const status = schoolRes.status
        if (status === 403) return <ErrorPage message={t("staff.settings.error.forbidden")} icon={CircleX} />
        if (status === 500) return <ErrorPage message={t("staff.settings.error.server")} icon={ServerCrash} />
        return <ErrorPage message={t("staff.settings.error.load")} icon={CircleX} />
    }

    if (!yearsRes.ok && yearsRes.status !== 403) {
        if (yearsRes.status === 500) return <ErrorPage message={t("staff.settings.error.server")} icon={ServerCrash} />
        return <ErrorPage message={t("staff.settings.error.loadYears")} icon={CircleX} />
    }

    const {school, access: schoolAccess} = await schoolRes.json()
    const yearsData = yearsRes.ok ? await yearsRes.json() : {academicYears: []}
    const academicYears = yearsData.academicYears ?? []
    const fallbackAccess = yearsData.access || schoolAccess
        ? null
        : await getSchoolNavigationAccess(id)
    const access = yearsData.access ?? schoolAccess ?? fallbackAccess ?? {
        owner: false,
        roles: [{position: -1, permissions: ["school.view"]}],
    }
    const displayNames = new Intl.DisplayNames([locale === "pl" ? "pl-PL" : "en-US"], {type: "region"})
    const regionNames = Object.fromEntries(countries.map((country) => [
        country.cca2,
        displayNames.of(country.cca2) ?? country.cca2,
    ]))
    return (
        <SettingsClientPage
            school={school}
            academicYears={academicYears}
            canListAcademicYears={yearsRes.ok}
            access={access}
            regionNames={regionNames}
            currentYear={new Date().getUTCFullYear()}
        />
    )
}

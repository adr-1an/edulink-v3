import {cookies} from "next/headers"
import {CircleX, GlobeX, ServerCrash} from "lucide-react"
import ErrorPage from "@/components/app/error"
import {emptySchoolAccess, isSchoolAccess} from "@/lib/school_access"
import {normalizeGuardianList} from "@/lib/guardian"
import {getTranslations} from "@/i18n/server"
import GuardiansClientPage from "./client_page"

export async function generateMetadata() {
    const {t} = await getTranslations()
    return {title: t("staff.guardians.metaTitle")}
}

interface GuardianListResponse {
    guardians?: unknown
    access?: unknown
}

export default async function Page({params}: {params: Promise<{id: string}>}) {
    const {t} = await getTranslations()
    const {id} = await params
    const token = (await cookies()).get("token")?.value
    let res: Response

    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/schools/${id}/guardians`, {
            headers: {Authorization: `Bearer ${token}`},
            cache: "no-store",
        })
    } catch {
        return <ErrorPage message={t("staff.guardians.error.network")} icon={GlobeX} />
    }

    if (!res.ok) {
        if (res.status === 403) return <ErrorPage message={t("staff.guardians.error.pageForbidden")} icon={CircleX} />
        if (res.status === 500) return <ErrorPage message={t("staff.guardians.error.server")} icon={ServerCrash} />
        return <ErrorPage message={t("staff.guardians.error.pageLoad")} icon={CircleX} />
    }

    const data = await res.json() as GuardianListResponse
    const guardians = normalizeGuardianList(data.guardians)
    if (Array.isArray(data.guardians) && guardians.length !== data.guardians.length) {
        return <ErrorPage message={t("staff.guardians.error.invalidResponse")} icon={ServerCrash} />
    }

    const snapshot = guardians.map((guardian) => `${guardian.id}:${guardian.name}:${guardian.lastName}:${guardian.email}:${guardian.accountEnabled}:${guardian.accountActive}`).join("|")
    return <GuardiansClientPage key={snapshot} schoolID={id} initialGuardians={guardians} access={isSchoolAccess(data.access) ? data.access : emptySchoolAccess} />
}

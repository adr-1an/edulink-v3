import {cookies} from "next/headers"
import {CircleX, GlobeX, ServerCrash} from "lucide-react"
import ErrorPage from "@/components/app/error"
import LogsClientPage, {type SchoolLog, type SchoolLogType} from "./client_page"
import {getTranslations} from "@/i18n/server"

export async function generateMetadata() {
    const {t} = await getTranslations()
    return {title: t("staff.logs.metaTitle")}
}

interface RawSchoolLog {
    id: string
    user: {
        id: string
        name: string
        email: string
    }
    action: string
    type: SchoolLogType
    title: string
    message: string
    details: string
    createdAt: string
}

export default async function Page({params}: {params: Promise<{id: string}>}) {
    const {t} = await getTranslations()
    const {id} = await params
    const token = (await cookies()).get("token")?.value
    let logsRes: Response

    try {
        logsRes = await fetch(`${process.env.API_URL}/v1/staff/schools/${id}/logs`, {
            headers: {Authorization: `Bearer ${token}`},
            cache: "no-store",
        })
    } catch {
        return <ErrorPage message={t("staff.logs.error.network")} icon={GlobeX} />
    }

    if (!logsRes.ok) {
        if (logsRes.status === 403) return <ErrorPage message={t("staff.logs.error.forbidden")} icon={CircleX} />
        if (logsRes.status === 500) return <ErrorPage message={t("staff.logs.error.server")} icon={ServerCrash} />
        return <ErrorPage message={t("staff.logs.error.load")} icon={CircleX} />
    }

    const data = await logsRes.json() as {logs?: RawSchoolLog[]}
    const logs: SchoolLog[] = (data.logs ?? [])
        .sort((first, second) => new Date(second.createdAt).getTime() - new Date(first.createdAt).getTime())

    return <LogsClientPage logs={logs} />
}

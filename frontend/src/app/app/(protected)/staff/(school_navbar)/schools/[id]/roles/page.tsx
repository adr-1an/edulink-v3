import {cookies} from "next/headers"
import {CircleX, GlobeX, ServerCrash} from "lucide-react"
import ErrorPage from "@/components/app/error"
import RolesClientPage from "./client_page"
import {emptySchoolAccess} from "@/lib/school_access"
import {getSchoolNavigationAccess} from "../../../school_navigation_actions"
import {getTranslations} from "@/i18n/server"

export async function generateMetadata() {
    const {t} = await getTranslations()
    return {title: t("staff.roles.metaTitle")}
}

export default async function Page({params}: {params: Promise<{id: string}>}) {
    const {t} = await getTranslations()
    const {id} = await params
    const token = (await cookies()).get("token")?.value
    let rolesRes: Response
    let permissionsRes: Response

    try {
        ;[rolesRes, permissionsRes] = await Promise.all([
            fetch(`${process.env.API_URL}/v1/staff/schools/${id}/roles`, {
                headers: {Authorization: `Bearer ${token}`},
                cache: "no-store",
            }),
            fetch(`${process.env.API_URL}/v1/staff/roles/permissions`, {
                headers: {Authorization: `Bearer ${token}`},
                cache: "no-store",
            }),
        ])
    } catch {
        return <ErrorPage message={t("staff.roles.error.network")} icon={GlobeX} />
    }

    if (!rolesRes.ok || !permissionsRes.ok) {
        const status = !rolesRes.ok ? rolesRes.status : permissionsRes.status
        if (status === 403) return <ErrorPage message={t("staff.roles.error.forbidden")} icon={CircleX} />
        if (status === 500) return <ErrorPage message={t("staff.roles.error.server")} icon={ServerCrash} />
        return <ErrorPage message={t("staff.roles.error.load")} icon={CircleX} />
    }

    const [{roles, access}, {permissions}] = await Promise.all([rolesRes.json(), permissionsRes.json()])
    const resolvedAccess = access ?? await getSchoolNavigationAccess(id) ?? emptySchoolAccess
    const roleList = roles ?? []
    const roleVersion = roleList.map((role: {id: string; position: number; name: string; color: string; permissions?: string[]}) =>
        `${role.id}:${role.position}:${role.name}:${role.color}:${(role.permissions ?? []).join(",")}`
    ).join("|")
    return <RolesClientPage key={roleVersion} schoolID={id} roles={roleList} availablePermissions={permissions ?? []} access={resolvedAccess} />
}

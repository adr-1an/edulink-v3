import {cookies} from "next/headers"
import ErrorPage from "@/components/app/error"
import {CircleX, GlobeX, ServerCrash} from "lucide-react"
import StaffClientPage, {type StaffMember, type StaffRole, type UserSummary} from "./client_page"
import {emptySchoolAccess, type SchoolAccess} from "@/lib/school_access"
import {getTranslations} from "@/i18n/server"
import {normalizeProfilePictureURL} from "@/lib/profile_picture"

export async function generateMetadata() {
    const {t} = await getTranslations()
    return {title: t("staff.members.metaTitle")}
}

interface RawUserSummary extends Omit<UserSummary, "profilePictureURL"> {
    profilePictureURL?: unknown
}

interface RawStaffMember extends Omit<StaffMember, "user" | "addedBy"> {
    user: RawUserSummary
    addedBy: RawUserSummary
}

interface StaffListResponse {
    staff?: RawStaffMember[] | null
    access?: SchoolAccess
}

interface RoleListResponse {
    roles?: StaffRole[] | null
}

export default async function Page({params}: {params: Promise<{id: string}>}) {
    const {t} = await getTranslations()
    const {id} = await params
    const token = (await cookies()).get("token")?.value
    let staffRes: Response
    let rolesRes: Response

    try {
        ;[staffRes, rolesRes] = await Promise.all([
            fetch(`${process.env.API_URL}/v1/staff/schools/${id}/staff`, {
                headers: {Authorization: `Bearer ${token}`},
            }),
            fetch(`${process.env.API_URL}/v1/staff/schools/${id}/roles`, {
                headers: {Authorization: `Bearer ${token}`},
            }),
        ])
    } catch {
        return <ErrorPage message={t("staff.members.error.pageNetwork")} icon={GlobeX} />
    }

    if (!staffRes.ok) {
        const status = staffRes.status
        if (status === 403) return <ErrorPage message={t("staff.members.error.pageForbidden")} icon={CircleX} />
        if (status === 500) return <ErrorPage message={t("staff.members.error.server")} icon={ServerCrash} />
        return <ErrorPage message={t("staff.members.error.pageLoad")} icon={CircleX} />
    }

    const data = await staffRes.json() as StaffListResponse
    const roleData = rolesRes.ok ? await rolesRes.json() as RoleListResponse : {roles: []}
    const staff: StaffMember[] = (Array.isArray(data.staff) ? data.staff : []).map((member) => ({
        ...member,
        user: {
            ...member.user,
            profilePictureURL: normalizeProfilePictureURL(member.user.profilePictureURL),
        },
        addedBy: {
            ...member.addedBy,
            profilePictureURL: normalizeProfilePictureURL(member.addedBy.profilePictureURL),
        },
    }))

    const staffSnapshot = staff.map((member) => member.id).join(":")
    return <StaffClientPage key={staffSnapshot} schoolID={id} staff={staff} availableRoles={roleData.roles ?? []} canListRoles={rolesRes.ok} access={data.access ?? emptySchoolAccess} />
}

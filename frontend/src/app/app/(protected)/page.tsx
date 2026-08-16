import { cookies } from "next/headers";
import ErrorPage from "@/components/app/error"
import {CircleX, GlobeX, ServerCrash} from "lucide-react"
import ClientPage, {type StaffInvitation} from "@/app/app/(protected)/client_page"
import {normalizeProfilePicture} from "@/lib/profile_picture"
import {getTranslations} from "@/i18n/server"
import countries from "world-countries"

export async function generateMetadata() {
    const {t} = await getTranslations()
    return {title: t("staff.schools.metaTitle")}
}

export default async function Page() {
    const {locale, t} = await getTranslations()
    const cookieStore = await cookies()
    const token = cookieStore.get("token")?.value

    let profileResponse: Response
    let schoolsResponse: Response
    let invitationsResponse: Response | null

    try {
        ;[profileResponse, schoolsResponse, invitationsResponse] = await Promise.all([
            fetch(`${process.env.API_URL}/v1/staff/profile`, {
                headers: {Authorization: `Bearer ${token}`},
            }),
            fetch(`${process.env.API_URL}/v1/staff/schools`, {
                headers: {Authorization: `Bearer ${token}`},
            }),
            fetch(`${process.env.API_URL}/v1/staff/staff-invitations`, {
                headers: {Authorization: `Bearer ${token}`},
                cache: "no-store",
            }).catch(() => null),
        ])
    } catch {
        return <ErrorPage message={t("staff.schools.error.network")} icon={GlobeX} />
    }

    for (const response of [profileResponse, schoolsResponse]) {
        if (!response.ok) {
            if (response.status === 500) {
                return <ErrorPage message={t("staff.schools.error.server")} icon={ServerCrash} />
            }
            return <ErrorPage message={t("staff.schools.error.unexpected")} icon={CircleX} />
        }
    }

    const profileData = await profileResponse.json() as {
        user?: {
            id: string
            name: string
            profilePicture?: unknown
        }
    }
    const schoolData = await schoolsResponse.json() as {schools?: unknown}
    let invitations: StaffInvitation[] = []
    let invitationsUnavailable = !invitationsResponse?.ok

    if (invitationsResponse?.ok) {
        try {
            const invitationData = await invitationsResponse.json() as {invitations?: StaffInvitation[] | null}
            invitations = Array.isArray(invitationData.invitations) ? invitationData.invitations : []
        } catch {
            invitationsUnavailable = true
        }
    }

    const schools = Array.isArray(schoolData.schools) ? schoolData.schools : []
    if (!profileData.user) return <ErrorPage message={t("staff.schools.error.profile")} icon={CircleX} />
    const invitationSnapshot = invitationsUnavailable ? "unavailable" : invitations.map((invitation) => invitation.id).join(":")
    const displayNames = new Intl.DisplayNames([locale === "pl" ? "pl-PL" : "en-US"], {type: "region"})
    const regionNames = Object.fromEntries(countries.map((country) => [
        country.cca2,
        displayNames.of(country.cca2) ?? country.cca2,
    ]))

    return (
        <ClientPage
            key={invitationSnapshot}
            user={{
                ...profileData.user,
                profilePicture: normalizeProfilePicture(profileData.user.profilePicture),
            }}
            schools={schools}
            regionNames={regionNames}
            initialInvitations={invitations}
            invitationsInitiallyUnavailable={invitationsUnavailable}
        />
    )
}

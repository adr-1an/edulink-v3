import {cookies} from "next/headers"
import ErrorPage from "@/components/app/error"
import {CircleX, GlobeX, ServerCrash} from "lucide-react"
import ClientPage, {type StaffProfileUser} from "@/app/app/(protected)/staff/profile/client_page"
import {type TwoFactorStatus} from "./two_factor_setup"
import {normalizeProfilePicture} from "@/lib/profile_picture"

export const metadata = {title: "Profile"}

export default async function Page() {
    const cookieStore = await cookies()
    const token = cookieStore.get("token")?.value

    let errorCode = 0
    let res

    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/profile`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            }
        })

        if (!res.ok) errorCode = res.status
    } catch {
        return <ErrorPage message="Network error, please try again." icon={GlobeX} />
    }

    if (errorCode) {
        if (errorCode === 500) {
            return <ErrorPage message="Internal server error." icon={ServerCrash} />
        } else {
            return <ErrorPage message="An unexpected error occurred." icon={CircleX} />
        }
    }

    let data: {user?: Omit<StaffProfileUser, "profilePicture" | "twoFactorStatus"> & {
        profilePicture?: unknown
        twoFactorStatus?: unknown
    }}
    try {
        data = await res.json() as typeof data
    } catch {
        return <ErrorPage message="The server returned an invalid profile." icon={CircleX} />
    }
    if (!data.user) return <ErrorPage message="The server returned an incomplete profile." icon={CircleX} />
    const twoFactorStatus: TwoFactorStatus = data.user.twoFactorStatus === "pending" || data.user.twoFactorStatus === "enabled"
        ? data.user.twoFactorStatus
        : "disabled"

    return (
        <ClientPage
            data={{
                user: {
                    ...data.user,
                    twoFactorStatus,
                    profilePicture: normalizeProfilePicture(data.user.profilePicture),
                },
            }}
        />
    )
}

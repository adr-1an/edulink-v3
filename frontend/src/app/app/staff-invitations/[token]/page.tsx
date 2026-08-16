import ErrorPage from "@/components/app/error"
import {CalendarX2, CircleX, FileQuestionMark, GlobeX, ServerCrash} from "lucide-react"
import ClientPage from "@/app/app/staff-invitations/[token]/client_page"
import {cookies} from "next/headers"

export const metadata = {title: "Staff invitation"}

export default async function Page({ params }: { params: Promise<{ token: string }>}) {
    const cookieStore = await cookies()
    const authToken = cookieStore.get("token")?.value

    const { token } = await params

    // Check auth
    let isAuthenticated: boolean
    let authRes

    if (authToken) {
        try {
            authRes = await fetch(`${process.env.API_URL}/v1/staff/profile`, {
                method: "GET",
                headers: {
                    "Content-Type": "application/json",
                    Authorization: `Bearer ${authToken}`,
                },
                cache: "no-store",
            })

            isAuthenticated = authRes.ok
        } catch {
            isAuthenticated = false
        }
    } else {
        isAuthenticated = false
    }

    // Get invitation info
    let errorCode = 0
    let res

    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/staff-invitations/${token}`, {
            method: "GET",
        })

        if (!res.ok) errorCode = res.status
    } catch {
        return <ErrorPage message="Network error, please try again." icon={GlobeX} />
    }

    if (errorCode) {
        if (errorCode === 500) {
            return <ErrorPage message="Internal server error." icon={ServerCrash} />
        } else if (errorCode === 404) {
            return <ErrorPage message="Invalid invite link." icon={FileQuestionMark} />
        } else if (errorCode === 410) {
            return <ErrorPage message="This invite has expired. Request a new one." icon={CalendarX2} />
        } else {
            return <ErrorPage message="An unexpected error occurred." icon={CircleX} />
        }
    }

    const data = await res.json()
    let authData
    if (isAuthenticated && authRes) {
        authData = await authRes.json()
    }

    const canAccept = isAuthenticated
        ? authData?.user?.email === data.invitation.sentToEmail
        : false

    return <ClientPage canAccept={canAccept} authenticated={isAuthenticated} profileName={authData?.user?.name} inv={data.invitation} token={token} />
}

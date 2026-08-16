import ErrorPage from "@/components/app/error"
import {FileQuestionMark, GlobeX, ServerCrash} from "lucide-react"
import ClientPage from "@/app/auth/(protected)/register/[token]/client_page"

export default async function Page({ params }: { params: Promise<{ token: string }>}) {
    const { token } = await params

    let errorCode = 0
    let res

    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/auth/register/${token}`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
            },
        })

        if (!res.ok) errorCode = res.status
    } catch {
        errorCode = -1
    }

    if (errorCode) {
        if (errorCode === 500) {
            return <ErrorPage message="Internal server error." icon={ServerCrash} />
        } else if (errorCode === 404) {
            return <ErrorPage message="Invalid or expired registration token." icon={FileQuestionMark} />
        } else if (errorCode === -1) {
            return <ErrorPage message="Network error, please try again." icon={GlobeX} />
        } else {
            return <ErrorPage message="An unexpected error occurred." icon={ServerCrash} />
        }
    }

    if (!res) return <ErrorPage message="Network error, please try again." icon={GlobeX} />

    const data = await res.json()
    return <ClientPage email={data.email} token={token} />
}
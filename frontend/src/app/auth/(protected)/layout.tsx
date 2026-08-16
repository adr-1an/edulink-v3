import {cookies} from "next/headers"
import {ReactNode} from "react"
import {redirect} from "next/navigation"

export default async function Layout({ children }: { children: ReactNode }) {
    // Check if the user is already logged in
    const cookieStore = await cookies()
    const token = cookieStore.get("token")?.value

    let isAuthenticated = false

    if (token) {
        try {
            const res = await fetch(`${process.env.API_URL}/v1/staff/auth`, {
                method: "GET",
                headers: {
                    "Content-Type": "application/json",
                    Authorization: `Bearer ${token}`,
                },
                cache: "no-store",
            })

            isAuthenticated = res.ok
        } catch {
            return (
                <>
                    {children}
                </>
            )
        }
    }

    if (isAuthenticated) {
        redirect("/app")
    }

    return (
        <>
            {children}
        </>
    )
}
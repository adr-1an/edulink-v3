import {cookies} from "next/headers"
import {redirect} from "next/navigation"
import ClientLayout from "@/app/app/client_layout"
import React from "react"

export default async function Layout({ children } : {children: React.ReactNode}) {
    const cookieStore = await cookies()
    const token = cookieStore.get("token")?.value

    let isAuthenticated: boolean

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
            isAuthenticated = false
        }
    } else {
        isAuthenticated = false
    }

    if (!isAuthenticated) {
        return redirect("/auth/login")
    }

    return <ClientLayout>{children}</ClientLayout>
}
import React from "react"
import AuthPageShell from "@/components/auth/auth-page-shell"

export default function Layout({ children }: { children?: React.ReactNode }) {
    return <AuthPageShell>{children}</AuthPageShell>
}

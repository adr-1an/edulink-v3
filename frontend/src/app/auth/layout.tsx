import {ReactNode} from "react"
import AuthPageShell from "@/components/auth/auth-page-shell"

export default function ClientLayout({ children }: { children: ReactNode }) {
    return <AuthPageShell>{children}</AuthPageShell>
}

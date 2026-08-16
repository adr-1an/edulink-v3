import {redirect} from "next/navigation"
import ErrorPage from "@/components/app/error"
import {ServerCrash} from "lucide-react"
import {getPortalSession, portalHome} from "@/lib/portal_auth"
import PortalShell from "../portal_shell"
import {getTranslations} from "@/i18n/server"

export default async function Layout({children}: {children: React.ReactNode}) {
    const {t} = await getTranslations()
    const session = await getPortalSession()
    if (session.status === "unauthenticated") redirect("/app/portal")
    if (session.status === "unavailable") return <ErrorPage message={t("portal.studentUnavailable")} icon={ServerCrash} />
    if (session.user.accountType !== "student") redirect(portalHome(session.user.accountType))

    return <PortalShell accountType="student">{children}</PortalShell>
}

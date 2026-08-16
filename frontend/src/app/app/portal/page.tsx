import {redirect} from "next/navigation"
import {getPortalSession, portalHome} from "@/lib/portal_auth"
import PortalLoginPage from "./client_page"
import {getTranslations} from "@/i18n/server"

export async function generateMetadata() {
    const {t} = await getTranslations()
    return {title: t("portal.login.metaTitle")}
}

export default async function Page() {
    const session = await getPortalSession()
    if (session.status === "authenticated") redirect(portalHome(session.user.accountType))

    return <PortalLoginPage serviceUnavailable={session.status === "unavailable"} />
}

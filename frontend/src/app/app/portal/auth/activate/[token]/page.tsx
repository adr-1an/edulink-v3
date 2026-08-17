import {redirect} from "next/navigation"
import {getTranslations} from "@/i18n/server"
import {getPortalSession, portalHome} from "@/lib/portal_auth"
import PortalActivationPage from "./client_page"

export async function generateMetadata() {
    const {t} = await getTranslations()
    return {title: t("portal.activation.metaTitle")}
}

export default async function Page({params}: {params: Promise<{token: string}>}) {
    const session = await getPortalSession()
    if (session.status === "authenticated") redirect(portalHome(session.user.accountType))

    const {token} = await params
    return <PortalActivationPage token={token} />
}

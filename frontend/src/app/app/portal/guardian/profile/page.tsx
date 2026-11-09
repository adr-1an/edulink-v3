import {getTranslations} from "@/i18n/server"
import PortalProfilePage from "../../profile_page"

export async function generateMetadata() {
    const {t} = await getTranslations()
    return {title: t("profile.guardian.metaTitle")}
}

export default function Page() {
    return <PortalProfilePage accountType="guardian" />
}

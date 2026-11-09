import {getTranslations} from "@/i18n/server"
import PortalProfilePage from "../../profile_page"

export async function generateMetadata() {
    const {t} = await getTranslations()
    return {title: t("profile.student.metaTitle")}
}

export default function Page() {
    return <PortalProfilePage accountType="student" />
}

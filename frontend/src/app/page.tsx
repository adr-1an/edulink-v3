import type {Metadata} from "next"
import LandingPage from "@/components/landing/landing-page"
import {getTranslations} from "@/i18n/server"

export async function generateMetadata(): Promise<Metadata> {
    const {t} = await getTranslations()
    return {
        title: {absolute: t("landing.metaTitle")},
        description: t("landing.metaDescription"),
    }
}

export default function Page() {
    return <LandingPage />
}

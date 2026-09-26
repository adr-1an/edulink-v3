import {HeartHandshake} from "lucide-react"
import {Card, CardContent} from "@/components/ui/card"
import {getTranslations} from "@/i18n/server"

export async function generateMetadata() {
    const {t} = await getTranslations()
    return {title: t("portal.guardian.metaTitle")}
}

export default async function Page() {
    const {t} = await getTranslations()
    return (
        <div className="mx-auto max-w-2xl py-10 sm:py-16">
            <Card>
                <CardContent className="flex flex-col items-center px-6 py-14 text-center">
                    <span className="mb-5 flex size-14 items-center justify-center rounded-2xl bg-primary/10 text-primary"><HeartHandshake className="size-7" /></span>
                    <h1 className="text-2xl font-semibold tracking-tight">{t("portal.guardian.metaTitle")}</h1>
                    <p className="mt-2 max-w-md text-muted-foreground">{t("portal.guardian.description")}</p>
                </CardContent>
            </Card>
        </div>
    )
}

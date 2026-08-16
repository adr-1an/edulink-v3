import type {LucideIcon} from "lucide-react"
import {
    Card,
    CardContent,
    CardDescription,
    CardFooter,
    CardHeader,
    CardTitle,
} from "@/components/ui/card"
import ErrorActions from "@/components/app/error-actions"
import {getTranslations} from "@/i18n/server"

export default async function ErrorPage({message, icon: Icon}: {message: string, icon?: LucideIcon}) {
    const {t} = await getTranslations()
    return (
        <section
            className="flex min-h-[50vh] w-full items-center justify-center px-4 py-10"
            aria-labelledby="error-page-title"
            role="alert"
        >
            <Card className="w-[min(32rem,calc(100vw-2rem))] overflow-hidden">
                <CardHeader className="items-center text-center">
                    <div className="mb-2 flex size-12 items-center justify-center rounded-full bg-destructive/10 text-destructive ring-1 ring-destructive/15">
                        {Icon ? <Icon className="size-6" aria-hidden="true" /> : null}
                    </div>
                    <CardTitle id="error-page-title">{t("error.title")}</CardTitle>
                    <CardDescription>
                        {t("error.description")}
                    </CardDescription>
                </CardHeader>

                <CardContent>
                    <div className="rounded-xl border border-destructive/15 bg-destructive/5 px-4 py-3 text-center text-sm text-foreground">
                        {message}
                    </div>
                </CardContent>

                <CardFooter className="justify-center gap-2">
                    <ErrorActions />
                </CardFooter>
            </Card>
        </section>
    )
}

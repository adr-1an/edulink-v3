"use client"

import {useState} from "react"
import Link from "next/link"
import {useRouter} from "next/navigation"
import {BookOpen, GraduationCap, ShieldCheck, UsersRound} from "lucide-react"
import {Button} from "@/components/ui/button"
import {Card, CardContent, CardHeader, CardTitle} from "@/components/ui/card"
import {Field, FieldLabel} from "@/components/ui/field"
import {Form} from "@/components/ui/form"
import {Input} from "@/components/ui/input"
import {TurnstileWidget, useTurnstile} from "@/components/turnstile"
import {handlePortalLogin} from "./actions"
import {useLocale} from "@/i18n/provider"

export default function PortalLoginPage({serviceUnavailable}: {serviceUnavailable: boolean}) {
    const router = useRouter()
    const {t} = useLocale()
    const turnstile = useTurnstile()
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(serviceUnavailable ? t("portal.login.unavailable") : "")

    async function submit(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        setLoading(true)
        setError("")

        const formData = new FormData(event.currentTarget)
        formData.set("cf-turnstile-response", turnstile.token)
        const result = await handlePortalLogin(formData)
        setLoading(false)

        if (!result.ok) {
            const errorKey = result.code === "credentials"
                ? "portal.login.error.credentials"
                : result.code === "disabled"
                    ? "portal.login.error.disabled"
                    : result.code === "invalid"
                        ? "portal.login.error.invalid"
                        : result.code === "verification"
                            ? "portal.login.error.verification"
                            : "portal.login.error.generic"
            setError(t(errorKey))
            turnstile.reset()
            return
        }

        router.replace(result.destination)
        router.refresh()
    }

    return (
        <main className="relative min-h-screen overflow-hidden bg-background px-4 py-8 text-foreground sm:px-6 lg:px-8">
            <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_15%_15%,color-mix(in_oklab,var(--primary)_9%,transparent),transparent_28%),radial-gradient(circle_at_85%_80%,color-mix(in_oklab,var(--primary)_7%,transparent),transparent_30%)]" />
            <div className="relative mx-auto grid min-h-[calc(100vh-4rem)] max-w-6xl items-center gap-10 lg:grid-cols-[minmax(0,1.05fr)_minmax(24rem,0.75fr)]">
                <section className="hidden space-y-7 lg:block">
                    <div className="inline-flex items-center gap-2 rounded-full border bg-card/80 px-3 py-1.5 text-sm font-medium shadow-xs backdrop-blur">
                        <GraduationCap className="size-4" /> {t("portal.login.badge")}
                    </div>
                    <div className="max-w-2xl space-y-4">
                        <h1 className="text-5xl font-semibold tracking-tight xl:text-6xl">{t("portal.login.heroTitle")}</h1>
                        <p className="max-w-xl text-lg leading-relaxed text-muted-foreground">{t("portal.login.heroDescription")}</p>
                    </div>
                    <div className="grid max-w-xl grid-cols-3 gap-3">
                        <PortalFeature icon={BookOpen} label={t("portal.login.yourCourses")} />
                        <PortalFeature icon={UsersRound} label={t("portal.login.forFamilies")} />
                        <PortalFeature icon={ShieldCheck} label={t("portal.login.schoolManaged")} />
                    </div>
                </section>

                <Card className="mx-auto w-full max-w-md overflow-hidden border-border/80 shadow-xl shadow-foreground/5">
                    <CardHeader className="space-y-4 px-6 pt-7 sm:px-8">
                        <span className="flex size-11 items-center justify-center rounded-2xl bg-primary text-primary-foreground lg:hidden"><GraduationCap className="size-5" /></span>
                        <div>
                            <CardTitle className="text-2xl">{t("portal.login.title")}</CardTitle>
                            <p className="mt-1.5 text-sm text-muted-foreground">{t("portal.login.description")}</p>
                        </div>
                    </CardHeader>
                    <CardContent className="px-6 pb-7 sm:px-8">
                        <Form className="space-y-4" onSubmit={submit}>
                            {error && <div className="rounded-xl border border-destructive/20 bg-destructive/8 px-3.5 py-3 text-sm text-destructive" role="alert">{error}</div>}
                            <Field>
                                <FieldLabel htmlFor="portal-email">{t("portal.login.email")}</FieldLabel>
                                <Input id="portal-email" name="email" type="email" autoComplete="email" placeholder="you@example.com" required autoFocus />
                            </Field>
                            <Field>
                                <FieldLabel htmlFor="portal-password">{t("portal.login.password")}</FieldLabel>
                                <Input id="portal-password" name="password" type="password" autoComplete="current-password" placeholder={t("portal.login.passwordPlaceholder")} minLength={8} required />
                            </Field>
                            <Field>
                                <FieldLabel>{t("portal.login.verification")}</FieldLabel>
                                {turnstile.configured ? (
                                    <TurnstileWidget key={turnstile.widgetKey} action="portal-login" onTokenChange={turnstile.setToken} />
                                ) : (
                                    <p className="text-sm text-destructive">{t("portal.login.turnstileMissing")}</p>
                                )}
                            </Field>
                            <Button className="w-full" size="lg" type="submit" loading={loading} disabled={!turnstile.configured || !turnstile.token}>{t("portal.login.submit")}</Button>
                        </Form>
                        <div className="mt-6 border-t pt-5 text-center text-sm text-muted-foreground">
                            {t("portal.login.staffQuestion")} <Link className="font-medium text-foreground underline-offset-4 hover:underline" href="/auth/login">{t("portal.login.staffLink")}</Link>
                        </div>
                    </CardContent>
                </Card>
            </div>
        </main>
    )
}

function PortalFeature({icon: Icon, label}: {icon: typeof BookOpen; label: string}) {
    return (
        <div className="rounded-2xl border bg-card/75 p-4 shadow-xs backdrop-blur">
            <Icon className="mb-5 size-5 text-muted-foreground" />
            <p className="text-sm font-medium">{label}</p>
        </div>
    )
}

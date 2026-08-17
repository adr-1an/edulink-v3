"use client"

import {useState} from "react"
import Link from "next/link"
import {useRouter} from "next/navigation"
import {CircleCheck, KeyRound, TriangleAlert} from "lucide-react"
import {handlePortalActivation} from "@/app/app/portal/actions"
import AuthCard from "@/components/auth/auth-card"
import {Alert, AlertDescription, AlertTitle} from "@/components/ui/alert"
import {Button} from "@/components/ui/button"
import {Field, FieldDescription, FieldLabel} from "@/components/ui/field"
import {Input} from "@/components/ui/input"
import {useLocale} from "@/i18n/provider"
import {validatePortalActivationPasswords} from "@/lib/portal_activation"

type TerminalState = "invalid_link" | "expired_link" | "session"

export default function PortalActivationPage({token}: {token: string}) {
    const router = useRouter()
    const {t} = useLocale()
    const [password, setPassword] = useState("")
    const [confirmation, setConfirmation] = useState("")
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState("")
    const [terminalState, setTerminalState] = useState<TerminalState | null>(null)

    const validationError = validatePortalActivationPasswords(password, confirmation)

    async function submit(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        if (loading || validationError) return

        setLoading(true)
        setError("")
        const result = await handlePortalActivation(token, password)
        setLoading(false)

        if (!result.ok) {
            if (result.code === "invalid_link" || result.code === "expired_link" || result.code === "session") {
                setTerminalState(result.code)
                return
            }

            setError(t(result.code === "invalid_password"
                ? "portal.activation.error.invalidPassword"
                : "portal.activation.error.generic"))
            return
        }

        router.replace(result.destination)
        router.refresh()
    }

    if (terminalState) {
        const accountActivated = terminalState === "session"
        const expired = terminalState === "expired_link"
        const titleKey = accountActivated
            ? "portal.activation.sessionTitle"
            : expired
                ? "portal.activation.expiredTitle"
                : "portal.activation.invalidTitle"
        const descriptionKey = accountActivated
            ? "portal.activation.sessionDescription"
            : expired
                ? "portal.activation.expiredDescription"
                : "portal.activation.invalidDescription"

        return (
            <ActivationShell>
                <AuthCard
                    icon={accountActivated ? CircleCheck : TriangleAlert}
                    eyebrow={t("portal.activation.eyebrow")}
                    title={t(titleKey)}
                    description={t(descriptionKey)}
                >
                    <div className="space-y-5">
                        <Alert variant={accountActivated ? "success" : "warning"}>
                            {accountActivated ? <CircleCheck /> : <TriangleAlert />}
                            <AlertTitle>{t(accountActivated ? "portal.activation.sessionAlertTitle" : "portal.activation.linkAlertTitle")}</AlertTitle>
                            <AlertDescription>
                                {t(accountActivated ? "portal.activation.sessionAlertDescription" : "portal.activation.linkAlertDescription")}
                            </AlertDescription>
                        </Alert>
                        <Button className="w-full" render={<Link href="/app/portal" />}>
                            {t(accountActivated ? "portal.activation.continueToLogin" : "portal.activation.backToLogin")}
                        </Button>
                    </div>
                </AuthCard>
            </ActivationShell>
        )
    }

    return (
        <ActivationShell>
            <AuthCard
                icon={KeyRound}
                eyebrow={t("portal.activation.eyebrow")}
                title={t("portal.activation.title")}
                description={t("portal.activation.description")}
            >
                <form className="space-y-5" onSubmit={submit}>
                    {error && (
                        <Alert variant="error">
                            <TriangleAlert />
                            <AlertTitle>{t("portal.activation.errorTitle")}</AlertTitle>
                            <AlertDescription>{error}</AlertDescription>
                        </Alert>
                    )}

                    <Field>
                        <FieldLabel htmlFor="portal-new-password">{t("portal.activation.newPassword")}</FieldLabel>
                        <Input
                            autoFocus
                            id="portal-new-password"
                            name="newPassword"
                            type="password"
                            autoComplete="new-password"
                            placeholder={t("portal.activation.passwordPlaceholder")}
                            minLength={8}
                            required
                            value={password}
                            onChange={(event) => setPassword(event.target.value)}
                        />
                        <FieldDescription className={password && password.length < 8 ? "text-destructive" : undefined}>
                            {t(password && password.length < 8
                                ? "portal.activation.passwordTooShort"
                                : "portal.activation.passwordHelp")}
                        </FieldDescription>
                    </Field>

                    <Field>
                        <FieldLabel htmlFor="portal-confirm-password">{t("portal.activation.confirmPassword")}</FieldLabel>
                        <Input
                            id="portal-confirm-password"
                            name="confirmPassword"
                            type="password"
                            autoComplete="new-password"
                            placeholder={t("portal.activation.confirmPlaceholder")}
                            minLength={8}
                            required
                            value={confirmation}
                            onChange={(event) => setConfirmation(event.target.value)}
                        />
                        {confirmation && password !== confirmation && (
                            <FieldDescription className="text-destructive">{t("portal.activation.passwordMismatch")}</FieldDescription>
                        )}
                    </Field>

                    <Button className="w-full" size="lg" type="submit" loading={loading} disabled={loading || validationError !== null}>
                        {t("portal.activation.submit")}
                    </Button>
                </form>

                <div className="mt-6 border-t pt-5 text-center text-sm text-muted-foreground">
                    {t("portal.activation.alreadyActive")} <Link className="font-medium text-foreground underline-offset-4 hover:underline" href="/app/portal">{t("portal.activation.loginLink")}</Link>
                </div>
            </AuthCard>
        </ActivationShell>
    )
}

function ActivationShell({children}: {children: React.ReactNode}) {
    return (
        <main className="relative flex min-h-screen items-center justify-center overflow-hidden bg-background px-4 py-8 text-foreground sm:px-6">
            <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_15%_15%,color-mix(in_oklab,var(--primary)_9%,transparent),transparent_28%),radial-gradient(circle_at_85%_80%,color-mix(in_oklab,var(--primary)_7%,transparent),transparent_30%)]" />
            <div className="relative w-full max-w-lg">{children}</div>
        </main>
    )
}

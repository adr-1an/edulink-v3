"use client"

import {useState} from "react"
import {KeyRound} from "lucide-react"
import {toast} from "sonner"
import {Button} from "@/components/ui/button"
import {Card, CardContent, CardDescription, CardHeader, CardTitle} from "@/components/ui/card"
import {Field, FieldDescription, FieldLabel} from "@/components/ui/field"
import {Input} from "@/components/ui/input"
import {useLocale} from "@/i18n/provider"
import {validatePortalPasswordUpdate, type PortalPasswordValidationError} from "@/lib/portal_password"
import {handlePortalPasswordUpdate, type PortalPasswordUpdateResult} from "./actions"

export default function PasswordUpdateCard() {
    const {t} = useLocale()
    const [saving, setSaving] = useState(false)
    const [validationError, setValidationError] = useState<PortalPasswordValidationError | null>(null)

    function validationMessage(error: PortalPasswordValidationError) {
        return t(`portal.password.validation.${error === "current_required" ? "currentRequired" : error === "too_short" ? "tooShort" : "mismatch"}`)
    }

    function responseMessage(result: Exclude<PortalPasswordUpdateResult, {ok: true}>) {
        if (result.code === "unauthorized") return t("portal.password.error.unauthorized")
        if (result.code === "invalid") return t("portal.password.error.invalid")
        if (result.code === "network") return t("portal.password.error.network")
        return t("portal.password.error.server")
    }

    async function submit(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        if (saving) return

        const form = event.currentTarget
        const formData = new FormData(form)
        const currentPassword = formData.get("currentPassword")?.toString() ?? ""
        const newPassword = formData.get("newPassword")?.toString() ?? ""
        const confirmation = formData.get("confirmPassword")?.toString() ?? ""
        const error = validatePortalPasswordUpdate(currentPassword, newPassword, confirmation)

        setValidationError(error)
        if (error) return

        setSaving(true)
        const result = await handlePortalPasswordUpdate(currentPassword, newPassword, confirmation)
        setSaving(false)

        if (!result.ok) {
            toast.error(responseMessage(result))
            return
        }

        form.reset()
        setValidationError(null)
        toast.success(t("portal.password.success"))
    }

    return (
        <Card>
            <CardHeader>
                <div className="flex items-start gap-3">
                    <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-muted text-muted-foreground">
                        <KeyRound className="size-4.5" />
                    </span>
                    <div>
                        <CardTitle>{t("portal.password.title")}</CardTitle>
                        <CardDescription className="mt-1">{t("portal.password.description")}</CardDescription>
                    </div>
                </div>
            </CardHeader>
            <CardContent>
                <form className="grid gap-4 sm:grid-cols-2" onSubmit={submit}>
                    <Field className="sm:col-span-2">
                        <FieldLabel htmlFor="portal-current-password">{t("portal.password.current")}</FieldLabel>
                        <Input
                            autoComplete="current-password"
                            disabled={saving}
                            id="portal-current-password"
                            name="currentPassword"
                            placeholder={t("portal.password.currentPlaceholder")}
                            type="password"
                            required
                        />
                    </Field>

                    <Field>
                        <FieldLabel htmlFor="portal-new-password">{t("portal.password.new")}</FieldLabel>
                        <Input
                            aria-invalid={validationError === "too_short" || validationError === "mismatch" || undefined}
                            autoComplete="new-password"
                            disabled={saving}
                            id="portal-new-password"
                            name="newPassword"
                            placeholder={t("portal.password.newPlaceholder")}
                            type="password"
                            minLength={8}
                            required
                        />
                        <FieldDescription>{t("portal.password.help")}</FieldDescription>
                    </Field>

                    <Field>
                        <FieldLabel htmlFor="portal-confirm-password">{t("portal.password.confirm")}</FieldLabel>
                        <Input
                            aria-invalid={validationError === "mismatch" || undefined}
                            autoComplete="new-password"
                            disabled={saving}
                            id="portal-confirm-password"
                            name="confirmPassword"
                            placeholder={t("portal.password.confirmPlaceholder")}
                            type="password"
                            minLength={8}
                            required
                        />
                    </Field>

                    {validationError && (
                        <p className="text-sm text-destructive sm:col-span-2" role="alert">
                            {validationMessage(validationError)}
                        </p>
                    )}

                    <div className="sm:col-span-2">
                        <Button loading={saving} type="submit">{t("portal.password.submit")}</Button>
                    </div>
                </form>
            </CardContent>
        </Card>
    )
}

"use client"

import {useState} from "react"
import {AlertTriangle, Check, Copy, Download, KeyRound, LoaderCircle, RefreshCw, ShieldCheck, ShieldOff, Smartphone} from "lucide-react"
import {QRCodeSVG} from "qrcode.react"
import {toast} from "sonner"
import {handleDisableTwoFactor, handleStartTwoFactorSetup, handleVerifyTwoFactorSetup} from "./actions"
import {Alert, AlertDescription} from "@/components/ui/alert"
import {Badge} from "@/components/ui/badge"
import {Button} from "@/components/ui/button"
import {Card} from "@/components/ui/card"
import {Checkbox} from "@/components/ui/checkbox"
import {
    Dialog, DialogClose, DialogDescription, DialogFooter, DialogHeader,
    DialogPanel, DialogPopup, DialogTitle,
} from "@/components/ui/dialog"
import {Field, FieldDescription, FieldLabel} from "@/components/ui/field"
import {Form} from "@/components/ui/form"
import {Input} from "@/components/ui/input"
import {Label} from "@/components/ui/label"
import {OTPField, OTPFieldInput, OTPFieldSeparator} from "@/components/ui/otp-field"
import {useLocale} from "@/i18n/provider"

export type TwoFactorStatus = "disabled" | "pending" | "enabled"

type SetupErrorCode = "network" | "unauthorized" | "already_enabled" | "server" | "invalid_response"
type VerifyErrorCode = "network" | "invalid_code" | "setup_expired" | "invalid_input" | "invalid_response" | "server"
type DisableErrorCode = "network" | "invalid_credentials" | "unavailable" | "invalid_input" | "server"

function secretFromURL(setupURL: string) {
    try {
        const secret = new URL(setupURL).searchParams.get("secret")?.trim() ?? ""
        return /^[A-Z2-7]+$/i.test(secret) ? secret : null
    } catch {
        return null
    }
}

function groupSecret(secret: string) {
    return secret.match(/.{1,4}/g)?.join(" ") ?? secret
}

export default function TwoFactorSetup({initialStatus}: {initialStatus: TwoFactorStatus}) {
    const {t} = useLocale()
    const [status, setStatus] = useState(initialStatus)
    const [open, setOpen] = useState(false)
    const [starting, setStarting] = useState(false)
    const [verifying, setVerifying] = useState(false)
    const [setupURL, setSetupURL] = useState<string | null>(null)
    const [code, setCode] = useState("")
    const [error, setError] = useState<string | null>(null)
    const [copied, setCopied] = useState(false)
    const [recoveryCodes, setRecoveryCodes] = useState<string[] | null>(null)
    const [recoveryCodesCopied, setRecoveryCodesCopied] = useState(false)
    const [recoveryCodesSaved, setRecoveryCodesSaved] = useState(false)
    const [disableOpen, setDisableOpen] = useState(false)
    const [disabling, setDisabling] = useState(false)
    const [disablePassword, setDisablePassword] = useState("")
    const [disableCode, setDisableCode] = useState("")
    const [disableError, setDisableError] = useState<string | null>(null)
    const secret = setupURL ? secretFromURL(setupURL) : null
    const busy = starting || verifying

    function setupErrorMessage(errorCode: SetupErrorCode) {
        if (errorCode === "network") return t("profile.staff.twoFactor.error.network")
        if (errorCode === "unauthorized") return t("profile.staff.twoFactor.error.unauthorized")
        if (errorCode === "already_enabled") return t("profile.staff.twoFactor.error.alreadyEnabled")
        if (errorCode === "invalid_response") return t("profile.staff.twoFactor.error.invalidResponse")
        return t("profile.staff.twoFactor.error.server")
    }

    function verifyErrorMessage(errorCode: VerifyErrorCode) {
        if (errorCode === "network") return t("profile.staff.twoFactor.error.network")
        if (errorCode === "invalid_code" || errorCode === "invalid_input") return t("profile.staff.twoFactor.error.invalidCode")
        if (errorCode === "setup_expired") return t("profile.staff.twoFactor.error.expired")
        if (errorCode === "invalid_response") return t("profile.staff.twoFactor.error.invalidRecoveryCodes")
        return t("profile.staff.twoFactor.error.server")
    }

    function disableErrorMessage(errorCode: DisableErrorCode) {
        if (errorCode === "network") return t("profile.staff.twoFactor.error.network")
        if (errorCode === "invalid_credentials") return t("profile.staff.twoFactor.disableInvalidCredentials")
        if (errorCode === "unavailable") return t("profile.staff.twoFactor.disableUnavailable")
        if (errorCode === "invalid_input") return t("profile.staff.twoFactor.disableInvalidInput")
        return t("profile.staff.twoFactor.error.server")
    }

    async function startSetup() {
        if (busy) return
        setOpen(true)
        setStarting(true)
        setSetupURL(null)
        setCode("")
        setError(null)
        setCopied(false)
        setRecoveryCodes(null)
        setRecoveryCodesCopied(false)
        setRecoveryCodesSaved(false)

        const result = await handleStartTwoFactorSetup()
        setStarting(false)
        if (!result.ok) {
            if (result.code === "already_enabled") setStatus("enabled")
            setError(setupErrorMessage(result.code))
            return
        }

        if (!secretFromURL(result.setupURL)) {
            setError(t("profile.staff.twoFactor.error.invalidResponse"))
            return
        }
        setStatus("pending")
        setSetupURL(result.setupURL)
    }

    async function verify(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        if (verifying || code.length !== 6) return
        setVerifying(true)
        setError(null)

        const result = await handleVerifyTwoFactorSetup(code)
        setVerifying(false)
        if (!result.ok) {
            setError(verifyErrorMessage(result.code))
            setCode("")
            return
        }

        setStatus("enabled")
        setSetupURL(null)
        setCode("")
        setRecoveryCodes(result.recoveryCodes)
    }

    async function copySecret() {
        if (!secret) return
        try {
            await navigator.clipboard.writeText(secret)
            setCopied(true)
            toast.success(t("profile.staff.twoFactor.copied"))
        } catch {
            toast.error(t("profile.staff.twoFactor.copyFailed"))
        }
    }

    function recoveryCodesText() {
        if (!recoveryCodes) return ""
        return `${t("profile.staff.twoFactor.recoveryCodesFileTitle")}\n\n${recoveryCodes.join("\n")}\n`
    }

    async function copyRecoveryCodes() {
        if (!recoveryCodes) return
        try {
            await navigator.clipboard.writeText(recoveryCodesText())
            setRecoveryCodesCopied(true)
            toast.success(t("profile.staff.twoFactor.recoveryCodesCopied"))
        } catch {
            toast.error(t("profile.staff.twoFactor.recoveryCodesCopyFailed"))
        }
    }

    function downloadRecoveryCodes() {
        if (!recoveryCodes) return
        const blob = new Blob([recoveryCodesText()], {type: "text/plain;charset=utf-8"})
        const url = URL.createObjectURL(blob)
        const link = document.createElement("a")
        link.href = url
        link.download = "edulink-recovery-codes.txt"
        link.click()
        URL.revokeObjectURL(url)
        toast.success(t("profile.staff.twoFactor.recoveryCodesDownloaded"))
    }

    function finishSetup() {
        if (!recoveryCodesSaved) return
        setOpen(false)
        setRecoveryCodes(null)
        setRecoveryCodesCopied(false)
        setRecoveryCodesSaved(false)
        setError(null)
        toast.success(t("profile.staff.twoFactor.enabledToast"))
    }

    async function disableTwoFactor(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        if (disabling || !disablePassword || disableCode.length !== 6) return
        setDisabling(true)
        setDisableError(null)

        const result = await handleDisableTwoFactor(disablePassword, disableCode)
        setDisabling(false)
        if (!result.ok) {
            setDisableError(disableErrorMessage(result.code))
            setDisableCode("")
            return
        }

        setStatus("disabled")
        setDisableOpen(false)
        setDisablePassword("")
        setDisableCode("")
        toast.success(t("profile.staff.twoFactor.disabledToast"))
    }

    function changeOpen(nextOpen: boolean) {
        if (busy || (recoveryCodes && !nextOpen)) return
        setOpen(nextOpen)
        if (!nextOpen) {
            setSetupURL(null)
            setCode("")
            setError(null)
            setCopied(false)
            setRecoveryCodesCopied(false)
            setRecoveryCodesSaved(false)
        }
    }

    function changeDisableOpen(nextOpen: boolean) {
        if (disabling) return
        setDisableOpen(nextOpen)
        if (!nextOpen) {
            setDisablePassword("")
            setDisableCode("")
            setDisableError(null)
        }
    }

    return (
        <>
            <Card className="p-5">
                <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                    <div className="flex min-w-0 items-start gap-3">
                        <span className={`flex size-10 shrink-0 items-center justify-center rounded-xl ${status === "enabled" ? "bg-emerald-500/12 text-emerald-700 dark:text-emerald-400" : "bg-muted text-muted-foreground"}`}>
                            <ShieldCheck className="size-5" />
                        </span>
                        <div className="min-w-0">
                            <div className="flex flex-wrap items-center gap-2">
                                <p className="font-medium">{t("profile.staff.twoFactor.title")}</p>
                                <Badge variant={status === "enabled" ? "success" : status === "pending" ? "warning" : "secondary"}>
                                    {t(`profile.staff.twoFactor.status.${status}`)}
                                </Badge>
                            </div>
                            <p className="mt-1 text-sm leading-5 text-muted-foreground">
                                {t(`profile.staff.twoFactor.description.${status}`)}
                            </p>
                        </div>
                    </div>
                    {status === "enabled" ? (
                        <Button className="shrink-0" type="button" variant="destructive-outline" onClick={() => setDisableOpen(true)}>
                            <ShieldOff /> {t("profile.staff.twoFactor.disable")}
                        </Button>
                    ) : (
                        <Button className="shrink-0" type="button" variant={status === "pending" ? "outline" : "default"} onClick={startSetup}>
                            {status === "pending" ? <RefreshCw /> : <KeyRound />}
                            {t(status === "pending" ? "profile.staff.twoFactor.restart" : "profile.staff.twoFactor.setup")}
                        </Button>
                    )}
                </div>
            </Card>

            <Dialog open={open} onOpenChange={changeOpen}>
                <DialogPopup className="sm:max-w-lg" closeProps={{disabled: busy || Boolean(recoveryCodes)}}>
                    <DialogHeader>
                        <DialogTitle>{t(recoveryCodes ? "profile.staff.twoFactor.recoveryCodesTitle" : "profile.staff.twoFactor.dialogTitle")}</DialogTitle>
                        <DialogDescription>{t(recoveryCodes ? "profile.staff.twoFactor.recoveryCodesDescription" : "profile.staff.twoFactor.dialogDescription")}</DialogDescription>
                    </DialogHeader>

                    <DialogPanel>
                        {recoveryCodes ? (
                            <div className="space-y-5">
                                <Alert variant="warning">
                                    <AlertTriangle />
                                    <AlertDescription>{t("profile.staff.twoFactor.recoveryCodesWarning")}</AlertDescription>
                                </Alert>

                                <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
                                    {recoveryCodes.map((recoveryCode, index) => (
                                        <div className="flex items-center gap-3 rounded-xl border bg-muted/20 px-3 py-2.5" key={recoveryCode}>
                                            <span className="w-4 shrink-0 text-right text-xs tabular-nums text-muted-foreground">{index + 1}</span>
                                            <code className="select-all font-mono text-sm font-semibold tracking-wide">{recoveryCode}</code>
                                        </div>
                                    ))}
                                </div>

                                <div className="grid gap-2 sm:grid-cols-2">
                                    <Button type="button" variant="outline" onClick={copyRecoveryCodes}>
                                        {recoveryCodesCopied ? <Check /> : <Copy />}
                                        {t(recoveryCodesCopied ? "profile.staff.twoFactor.recoveryCodesCopiedButton" : "profile.staff.twoFactor.recoveryCodesCopy")}
                                    </Button>
                                    <Button type="button" variant="outline" onClick={downloadRecoveryCodes}>
                                        <Download /> {t("profile.staff.twoFactor.recoveryCodesDownload")}
                                    </Button>
                                </div>

                                <Label className="flex cursor-pointer items-start gap-2.5 rounded-xl border p-3">
                                    <Checkbox
                                        className="mt-0.5"
                                        checked={recoveryCodesSaved}
                                        onCheckedChange={setRecoveryCodesSaved}
                                    />
                                    <span className="text-sm leading-5">{t("profile.staff.twoFactor.recoveryCodesAcknowledgement")}</span>
                                </Label>

                                <DialogFooter>
                                    <Button type="button" disabled={!recoveryCodesSaved} onClick={finishSetup}>
                                        <ShieldCheck /> {t("profile.staff.twoFactor.recoveryCodesDone")}
                                    </Button>
                                </DialogFooter>
                            </div>
                        ) : starting ? (
                            <div className="flex min-h-64 flex-col items-center justify-center gap-3 text-center">
                                <LoaderCircle className="size-7 animate-spin text-primary" />
                                <div>
                                    <p className="font-medium">{t("profile.staff.twoFactor.preparing")}</p>
                                    <p className="mt-1 text-sm text-muted-foreground">{t("profile.staff.twoFactor.preparingDescription")}</p>
                                </div>
                            </div>
                        ) : setupURL && secret ? (
                            <form className="space-y-5" onSubmit={verify}>
                                <div className="grid gap-5 sm:grid-cols-[auto_1fr] sm:items-center">
                                    <div className="mx-auto rounded-2xl border bg-white p-2 shadow-sm">
                                        <QRCodeSVG
                                            value={setupURL}
                                            size={176}
                                            level="M"
                                            marginSize={2}
                                            title={t("profile.staff.twoFactor.qrTitle")}
                                        />
                                    </div>
                                    <div className="space-y-3">
                                        <div className="flex items-start gap-3">
                                            <span className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-sm font-semibold text-primary">1</span>
                                            <div>
                                                <p className="text-sm font-medium">{t("profile.staff.twoFactor.scanTitle")}</p>
                                                <p className="mt-1 text-xs leading-5 text-muted-foreground">{t("profile.staff.twoFactor.scanDescription")}</p>
                                            </div>
                                        </div>
                                        <div className="flex items-start gap-3">
                                            <span className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-sm font-semibold text-primary">2</span>
                                            <div>
                                                <p className="text-sm font-medium">{t("profile.staff.twoFactor.enterTitle")}</p>
                                                <p className="mt-1 text-xs leading-5 text-muted-foreground">{t("profile.staff.twoFactor.enterDescription")}</p>
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                <details className="group rounded-xl border bg-muted/20 p-3">
                                    <summary className="cursor-pointer text-sm font-medium">{t("profile.staff.twoFactor.manualTitle")}</summary>
                                    <div className="mt-3 flex items-center gap-2">
                                        <code className="min-w-0 flex-1 select-all break-all rounded-lg bg-background px-3 py-2 text-xs font-semibold tracking-wider">{groupSecret(secret)}</code>
                                        <Button type="button" size="icon-sm" variant="outline" aria-label={t("profile.staff.twoFactor.copySecret")} onClick={copySecret}>
                                            {copied ? <Check /> : <Copy />}
                                        </Button>
                                    </div>
                                    <p className="mt-2 text-xs text-muted-foreground">{t("profile.staff.twoFactor.manualDescription")}</p>
                                </details>

                                <Field>
                                    <FieldLabel htmlFor="two-factor-code">{t("profile.staff.twoFactor.codeLabel")}</FieldLabel>
                                    <OTPField
                                        id="two-factor-code"
                                        className="justify-center sm:justify-start"
                                        length={6}
                                        value={code}
                                        disabled={verifying}
                                        onValueChange={(value) => {
                                            setCode(value)
                                            setError(null)
                                        }}
                                    >
                                        <OTPFieldInput />
                                        <OTPFieldInput aria-label={t("profile.staff.twoFactor.digit", {position: 2})} />
                                        <OTPFieldInput aria-label={t("profile.staff.twoFactor.digit", {position: 3})} />
                                        <OTPFieldSeparator />
                                        <OTPFieldInput aria-label={t("profile.staff.twoFactor.digit", {position: 4})} />
                                        <OTPFieldInput aria-label={t("profile.staff.twoFactor.digit", {position: 5})} />
                                        <OTPFieldInput aria-label={t("profile.staff.twoFactor.digit", {position: 6})} />
                                    </OTPField>
                                    <FieldDescription>{t("profile.staff.twoFactor.codeDescription")}</FieldDescription>
                                </Field>

                                {error && <Alert variant="error"><AlertDescription>{error}</AlertDescription></Alert>}

                                <DialogFooter>
                                    <DialogClose render={<Button type="button" variant="outline" />} disabled={verifying}>
                                        {t("common.cancel")}
                                    </DialogClose>
                                    <Button type="submit" loading={verifying} disabled={code.length !== 6 || verifying}>
                                        <Smartphone /> {t("profile.staff.twoFactor.verify")}
                                    </Button>
                                </DialogFooter>
                            </form>
                        ) : (
                            <div className="space-y-4 py-6 text-center">
                                <Alert variant="error"><AlertDescription>{error ?? t("profile.staff.twoFactor.error.server")}</AlertDescription></Alert>
                                <Button type="button" variant="outline" onClick={startSetup}><RefreshCw /> {t("common.tryAgain")}</Button>
                            </div>
                        )}
                    </DialogPanel>
                </DialogPopup>
            </Dialog>

            <Dialog open={disableOpen} onOpenChange={changeDisableOpen}>
                <DialogPopup className="sm:max-w-md" closeProps={{disabled: disabling}}>
                    <DialogHeader>
                        <DialogTitle>{t("profile.staff.twoFactor.disableTitle")}</DialogTitle>
                        <DialogDescription>{t("profile.staff.twoFactor.disableDescription")}</DialogDescription>
                    </DialogHeader>

                    <Form className="contents" onSubmit={disableTwoFactor}>
                        <DialogPanel className="space-y-5">
                            <Alert variant="warning">
                                <AlertTriangle />
                                <AlertDescription>{t("profile.staff.twoFactor.disableWarning")}</AlertDescription>
                            </Alert>

                            <Field>
                                <FieldLabel htmlFor="two-factor-disable-password">{t("profile.staff.twoFactor.passwordLabel")}</FieldLabel>
                                <Input
                                    id="two-factor-disable-password"
                                    type="password"
                                    autoComplete="current-password"
                                    value={disablePassword}
                                    disabled={disabling}
                                    placeholder={t("profile.staff.twoFactor.passwordPlaceholder")}
                                    onChange={(event) => {
                                        setDisablePassword(event.target.value)
                                        setDisableError(null)
                                    }}
                                    required
                                />
                            </Field>

                            <Field>
                                <FieldLabel htmlFor="two-factor-disable-code">{t("profile.staff.twoFactor.codeLabel")}</FieldLabel>
                                <OTPField
                                    id="two-factor-disable-code"
                                    className="justify-center sm:justify-start"
                                    length={6}
                                    value={disableCode}
                                    disabled={disabling}
                                    onValueChange={(value) => {
                                        setDisableCode(value)
                                        setDisableError(null)
                                    }}
                                >
                                    <OTPFieldInput />
                                    <OTPFieldInput aria-label={t("profile.staff.twoFactor.digit", {position: 2})} />
                                    <OTPFieldInput aria-label={t("profile.staff.twoFactor.digit", {position: 3})} />
                                    <OTPFieldSeparator />
                                    <OTPFieldInput aria-label={t("profile.staff.twoFactor.digit", {position: 4})} />
                                    <OTPFieldInput aria-label={t("profile.staff.twoFactor.digit", {position: 5})} />
                                    <OTPFieldInput aria-label={t("profile.staff.twoFactor.digit", {position: 6})} />
                                </OTPField>
                                <FieldDescription>{t("profile.staff.twoFactor.disableCodeDescription")}</FieldDescription>
                            </Field>

                            {disableError && <Alert variant="error"><AlertDescription>{disableError}</AlertDescription></Alert>}
                        </DialogPanel>

                        <DialogFooter>
                            <DialogClose render={<Button type="button" variant="outline" />} disabled={disabling}>
                                {t("common.cancel")}
                            </DialogClose>
                            <Button type="submit" variant="destructive" loading={disabling} disabled={disabling || !disablePassword || disableCode.length !== 6}>
                                <ShieldOff /> {t("profile.staff.twoFactor.disableConfirm")}
                            </Button>
                        </DialogFooter>
                    </Form>
                </DialogPopup>
            </Dialog>
        </>
    )
}

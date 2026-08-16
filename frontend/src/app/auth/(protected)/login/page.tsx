"use client"

import React, {useEffect, useState} from "react"
import Link from "next/link"
import {useRouter, useSearchParams} from "next/navigation"
import {ArrowLeft, CircleCheck, GraduationCap, KeyRound, LifeBuoy, LogIn, MailWarning, ShieldCheck, TriangleAlert} from "lucide-react"
import {
    handleCompleteTwoFactorLogin,
    handleLogin,
    handleRecoverTwoFactor,
    handleResendVerificationLink,
} from "@/app/auth/(protected)/login/actions"
import AuthCard from "@/components/auth/auth-card"
import {useLocalDateTimeFormatter} from "@/components/local-date-time"
import {TurnstileWidget, useTurnstile} from "@/components/turnstile"
import {Alert, AlertDescription, AlertTitle} from "@/components/ui/alert"
import {Button} from "@/components/ui/button"
import {Checkbox} from "@/components/ui/checkbox"
import {Field, FieldDescription, FieldLabel} from "@/components/ui/field"
import {Form} from "@/components/ui/form"
import {Input} from "@/components/ui/input"
import {Label} from "@/components/ui/label"
import {OTPField, OTPFieldInput, OTPFieldSeparator} from "@/components/ui/otp-field"
import {Separator} from "@/components/ui/separator"
import {useLocale} from "@/i18n/provider"

type Errors = Record<string, string | string[]>
type LoginChallenge = {token: string, expiresAt: string}

export default function Page() {
    const router = useRouter()
    const {t} = useLocale()
    const formatDate = useLocalDateTimeFormatter()
    const searchParams = useSearchParams()
    const paramsEmail = searchParams.get("email") ?? ""
    const [email, setEmail] = useState(paramsEmail)
    const [password, setPassword] = useState("")
    const [stayLoggedIn, setStayLoggedIn] = useState(false)
    const [loading, setLoading] = useState(false)
    const [resendLoading, setResendLoading] = useState(false)
    const [resendSuccess, setResendSuccess] = useState(false)
    const [errors, setErrors] = useState<Errors>({})
    const [challenge, setChallenge] = useState<LoginChallenge | null>(null)
    const [twoFactorCode, setTwoFactorCode] = useState("")
    const [twoFactorLoading, setTwoFactorLoading] = useState(false)
    const [twoFactorError, setTwoFactorError] = useState<string | null>(null)
    const [challengeInvalid, setChallengeInvalid] = useState(false)
    const [usingRecoveryCode, setUsingRecoveryCode] = useState(false)
    const [recoveryCode, setRecoveryCode] = useState("")
    const [recoveryLoading, setRecoveryLoading] = useState(false)
    const [recoveryError, setRecoveryError] = useState<string | null>(null)
    const [recoverySuccess, setRecoverySuccess] = useState(false)
    const turnstile = useTurnstile()
    const unverified = errors.root === t("auth.login.unverified")

    useEffect(() => {
        if (paramsEmail) router.replace("/auth/login")
    }, [paramsEmail, router])

    async function handleSubmit(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        setLoading(true)
        setErrors({})
        setResendSuccess(false)
        setRecoverySuccess(false)

        const formData = new FormData(event.currentTarget)
        formData.set("cf-turnstile-response", turnstile.token)
        const result = await handleLogin(formData)

        if (!result.ok) {
            setLoading(false)
            turnstile.reset()
            const message = result.code === "unverified"
                ? t("auth.login.unverified")
                : result.code === "invalid"
                    ? t("auth.login.invalid")
                    : result.code === "verification"
                        ? t("auth.login.verificationError")
                        : result.code === "network"
                            ? t("auth.login.networkError")
                            : t("auth.login.genericError")
            setErrors({root: message})
            return
        }

        if (result.step === "two_factor") {
            setLoading(false)
            setPassword("")
            setChallenge(result.challenge)
            setTwoFactorCode("")
            setTwoFactorError(null)
            setChallengeInvalid(false)
            setUsingRecoveryCode(false)
            setRecoveryCode("")
            setRecoveryError(null)
            setRecoverySuccess(false)
            return
        }

        router.push("/app")
    }

    async function handleRecoverySubmit(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        if (recoveryLoading || recoveryCode.trim().length === 0) return

        setRecoveryLoading(true)
        setRecoveryError(null)
        const result = await handleRecoverTwoFactor(recoveryCode)
        if (!result.ok) {
            setRecoveryLoading(false)

            const message = result.code === "invalid_code"
                ? t("auth.login.recoveryInvalid")
                : result.code === "invalid_input"
                    ? t("auth.login.recoveryInvalidInput")
                    : result.code === "network"
                        ? t("auth.login.twoFactorNetwork")
                        : t("auth.login.recoveryGeneric")
            setRecoveryError(message)
            return
        }

        setChallenge(null)
        setTwoFactorCode("")
        setTwoFactorError(null)
        setChallengeInvalid(false)
        setUsingRecoveryCode(false)
        setRecoveryCode("")
        setRecoveryError(null)
        setRecoverySuccess(true)
        turnstile.reset()
    }

    async function handleTwoFactorSubmit(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        if (!challenge || challengeInvalid || twoFactorCode.length !== 6) return

        setTwoFactorLoading(true)
        setTwoFactorError(null)
        const result = await handleCompleteTwoFactorLogin(challenge.token, twoFactorCode)

        if (!result.ok) {
            setTwoFactorLoading(false)

            if (result.code === "expired" || result.code === "invalid_challenge") {
                setChallengeInvalid(true)
            }

            const message = result.code === "invalid_code"
                ? t("auth.login.twoFactorInvalid")
                : result.code === "expired"
                    ? t("auth.login.twoFactorExpired")
                    : result.code === "invalid_challenge"
                        ? t("auth.login.twoFactorInvalidChallenge")
                        : result.code === "invalid_input"
                            ? t("auth.login.twoFactorInvalidInput")
                            : result.code === "network"
                                ? t("auth.login.twoFactorNetwork")
                                : t("auth.login.twoFactorGeneric")
            setTwoFactorError(message)
            if (result.code === "invalid_code" || result.code === "invalid_input") {
                setTwoFactorCode("")
            }
            return
        }

        router.push("/app")
    }

    function returnToLogin() {
        setChallenge(null)
        setTwoFactorCode("")
        setTwoFactorError(null)
        setChallengeInvalid(false)
        setUsingRecoveryCode(false)
        setRecoveryCode("")
        setRecoveryError(null)
        setErrors({})
        turnstile.reset()
    }

    function showRecoveryForm() {
        setUsingRecoveryCode(true)
        setTwoFactorError(null)
        setRecoveryError(null)
    }

    function showAuthenticatorForm() {
        setUsingRecoveryCode(false)
        setRecoveryCode("")
        setRecoveryError(null)
    }

    async function onResend() {
        setResendLoading(true)
        setResendSuccess(false)
        const result = await handleResendVerificationLink(email)
        setResendLoading(false)

        if (!result.ok) {
            setErrors({root: result.code === "network" ? t("auth.login.networkError") : t("auth.login.resendError")})
            return
        }
        setErrors({})
        setResendSuccess(true)
    }

    if (challenge) {
        return (
            <AuthCard
                icon={usingRecoveryCode ? LifeBuoy : ShieldCheck}
                title={t(usingRecoveryCode ? "auth.login.recoveryTitle" : "auth.login.twoFactorTitle")}
                description={t(usingRecoveryCode ? "auth.login.recoveryDescription" : "auth.login.twoFactorDescription")}
            >
                <form className="space-y-5" onSubmit={usingRecoveryCode ? handleRecoverySubmit : handleTwoFactorSubmit}>
                    <div className="flex items-center gap-3 rounded-2xl border bg-muted/25 p-4">
                        <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
                            <KeyRound className="size-5" />
                        </span>
                        <div className="min-w-0">
                            <p className="text-xs font-medium text-muted-foreground">{t("auth.login.twoFactorAccount")}</p>
                            <p className="truncate text-sm font-semibold">{email}</p>
                        </div>
                    </div>

                    {!usingRecoveryCode && twoFactorError && (
                        <Alert variant="error">
                            <TriangleAlert />
                            <AlertTitle>{t("auth.login.twoFactorFailedTitle")}</AlertTitle>
                            <AlertDescription>{twoFactorError}</AlertDescription>
                        </Alert>
                    )}

                    {usingRecoveryCode ? (
                        <>
                            <Alert variant="warning">
                                <TriangleAlert />
                                <AlertTitle>{t("auth.login.recoveryWarningTitle")}</AlertTitle>
                                <AlertDescription>{t("auth.login.recoveryWarningDescription")}</AlertDescription>
                            </Alert>

                            {recoveryError && (
                                <Alert variant="error">
                                    <TriangleAlert />
                                    <AlertTitle>{t("auth.login.twoFactorFailedTitle")}</AlertTitle>
                                    <AlertDescription>{recoveryError}</AlertDescription>
                                </Alert>
                            )}

                            <Field>
                                <FieldLabel htmlFor="recovery-code">{t("auth.login.recoveryCode")}</FieldLabel>
                                <Input
                                    autoFocus
                                    id="recovery-code"
                                    type="text"
                                    autoComplete="one-time-code"
                                    autoCapitalize="none"
                                    spellCheck={false}
                                    placeholder={t("auth.login.recoveryCodePlaceholder")}
                                    value={recoveryCode}
                                    disabled={recoveryLoading}
                                    onChange={(event) => {
                                        setRecoveryCode(event.target.value)
                                        setRecoveryError(null)
                                    }}
                                />
                            </Field>

                            <Button
                                className="w-full"
                                type="submit"
                                loading={recoveryLoading}
                                disabled={recoveryCode.trim().length === 0}
                            >
                                <LifeBuoy /> {t("auth.login.recoverySubmit")}
                            </Button>
                            <Button className="w-full" type="button" variant="outline" disabled={recoveryLoading} onClick={showAuthenticatorForm}>
                                <ShieldCheck /> {t("auth.login.recoveryBack")}
                            </Button>
                        </>
                    ) : (
                        <>
                            <Field>
                                <FieldLabel htmlFor="two-factor-code">{t("auth.login.twoFactorCode")}</FieldLabel>
                                <OTPField
                                    id="two-factor-code"
                                    className="justify-center sm:justify-start"
                                    length={6}
                                    value={twoFactorCode}
                                    disabled={twoFactorLoading || challengeInvalid}
                                    onValueChange={(value) => {
                                        setTwoFactorCode(value)
                                        setTwoFactorError(null)
                                    }}
                                >
                                    <OTPFieldInput autoFocus aria-label={t("auth.login.twoFactorDigit", {position: 1})} />
                                    <OTPFieldInput aria-label={t("auth.login.twoFactorDigit", {position: 2})} />
                                    <OTPFieldInput aria-label={t("auth.login.twoFactorDigit", {position: 3})} />
                                    <OTPFieldSeparator />
                                    <OTPFieldInput aria-label={t("auth.login.twoFactorDigit", {position: 4})} />
                                    <OTPFieldInput aria-label={t("auth.login.twoFactorDigit", {position: 5})} />
                                    <OTPFieldInput aria-label={t("auth.login.twoFactorDigit", {position: 6})} />
                                </OTPField>
                                <FieldDescription>{t("auth.login.twoFactorCodeDescription")}</FieldDescription>
                            </Field>

                            <p className="text-xs leading-5 text-muted-foreground">
                                {t("auth.login.twoFactorExpires", {date: formatDate(challenge.expiresAt)})}
                            </p>

                            <Button
                                className="w-full"
                                type="submit"
                                loading={twoFactorLoading}
                                disabled={twoFactorCode.length !== 6 || challengeInvalid}
                            >
                                <ShieldCheck /> {t("auth.login.twoFactorSubmit")}
                            </Button>
                            <Button className="w-full" type="button" variant="outline" disabled={twoFactorLoading} onClick={showRecoveryForm}>
                                <LifeBuoy /> {t("auth.login.recoveryUse")}
                            </Button>
                        </>
                    )}

                    <Button className="w-full" type="button" variant="ghost" disabled={twoFactorLoading || recoveryLoading} onClick={returnToLogin}>
                        <ArrowLeft /> {t("auth.login.twoFactorBack")}
                    </Button>
                </form>
            </AuthCard>
        )
    }

    return (
        <AuthCard
            icon={LogIn}
            title={t("auth.login.title")}
            description={t("auth.login.description")}
        >
            <Form errors={errors} onSubmit={handleSubmit} className="space-y-5">
                {errors.root && (
                    <Alert variant={unverified ? "warning" : "error"}>
                        {unverified ? <MailWarning /> : <TriangleAlert />}
                        <AlertTitle>{unverified ? t("auth.login.verifyTitle") : t("auth.login.failedTitle")}</AlertTitle>
                        <AlertDescription>{errors.root}</AlertDescription>
                    </Alert>
                )}
                {resendSuccess && (
                    <Alert variant="success">
                        <CircleCheck />
                        <AlertTitle>{t("auth.login.verificationSent")}</AlertTitle>
                        <AlertDescription>{t("auth.login.verificationSentDescription")}</AlertDescription>
                    </Alert>
                )}
                {recoverySuccess && (
                    <Alert variant="success">
                        <CircleCheck />
                        <AlertTitle>{t("auth.login.recoverySuccessTitle")}</AlertTitle>
                        <AlertDescription>{t("auth.login.recoverySuccessDescription")}</AlertDescription>
                    </Alert>
                )}

                <Field>
                    <FieldLabel htmlFor="email">{t("auth.login.email")}</FieldLabel>
                    <Input
                        autoFocus
                        id="email"
                        name="email"
                        type="email"
                        autoComplete="email"
                        placeholder="you@example.com"
                        required
                        value={email}
                        onChange={(event) => setEmail(event.target.value)}
                    />
                </Field>

                <Field>
                    <div className="flex items-center justify-between gap-3">
                        <FieldLabel htmlFor="password">{t("auth.login.password")}</FieldLabel>
                        <Link className="text-xs font-medium text-muted-foreground hover:text-foreground" href="/auth/forgot">
                            {t("auth.login.forgot")}
                        </Link>
                    </div>
                    <Input
                        id="password"
                        name="password"
                        type="password"
                        autoComplete="current-password"
                        placeholder={t("auth.login.passwordPlaceholder")}
                        required
                        value={password}
                        onChange={(event) => setPassword(event.target.value)}
                    />
                </Field>

                <Label className="flex cursor-pointer items-center gap-2.5">
                    <Checkbox
                        id="stayLoggedIn"
                        name="stayLoggedIn"
                        checked={stayLoggedIn}
                        onCheckedChange={setStayLoggedIn}
                    />
                    <span className="text-sm">{t("auth.login.stayLoggedIn")}</span>
                </Label>

                <Field>
                    <FieldLabel>{t("auth.login.verification")}</FieldLabel>
                    {turnstile.configured ? (
                        <TurnstileWidget
                            key={turnstile.widgetKey}
                            action="login"
                            onTokenChange={turnstile.setToken}
                        />
                    ) : (
                        <Alert variant="error">
                            <TriangleAlert />
                            <AlertDescription>{t("auth.login.turnstileMissing")}</AlertDescription>
                        </Alert>
                    )}
                </Field>

                {unverified && (
                    <Button className="w-full" type="button" variant="outline" loading={resendLoading} onClick={onResend}>
                        {t("auth.login.resend")}
                    </Button>
                )}
                <Button className="w-full" loading={loading} disabled={!turnstile.configured || !turnstile.token} type="submit">
                    {t("auth.login.submit")}
                </Button>

                <div className="relative py-1">
                    <Separator />
                    <span className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 bg-card px-3 text-xs text-muted-foreground">
                        {t("auth.login.new")}
                    </span>
                </div>
                <Button className="w-full" variant="outline" render={<Link href="/auth/register" />}>
                    {t("auth.login.create")}
                </Button>

                <div className="rounded-2xl border bg-muted/25 p-4">
                    <div className="flex gap-3">
                        <GraduationCap className="mt-0.5 size-5 shrink-0 text-muted-foreground" />
                        <div>
                            <p className="text-sm font-medium">{t("auth.login.portalQuestion")}</p>
                            <p className="mt-1 text-xs leading-5 text-muted-foreground">
                                {t("auth.login.portalDescription")}
                            </p>
                            <Link className="mt-2 inline-block text-sm font-medium underline underline-offset-4" href="/app/portal">
                                {t("auth.login.portalLink")}
                            </Link>
                        </div>
                    </div>
                </div>
            </Form>
        </AuthCard>
    )
}

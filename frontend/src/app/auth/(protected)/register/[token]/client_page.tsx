"use client"

import React, {useState} from "react"
import Link from "next/link"
import {CircleCheck, TriangleAlert, UserRoundCheck} from "lucide-react"
import {handleRegistration} from "@/app/auth/(protected)/register/actions"
import AuthCard from "@/components/auth/auth-card"
import {TurnstileWidget, useTurnstile} from "@/components/turnstile"
import {Alert, AlertDescription, AlertTitle} from "@/components/ui/alert"
import {Button} from "@/components/ui/button"
import {Field, FieldDescription, FieldError, FieldLabel} from "@/components/ui/field"
import {Form} from "@/components/ui/form"
import {Input} from "@/components/ui/input"

interface Props {
    email: string
    token: string
}

type Errors = Record<string, string | string[]>

export default function ClientPage({email, token}: Props) {
    const [loading, setLoading] = useState(false)
    const [errors, setErrors] = useState<Errors>({})
    const [complete, setComplete] = useState(false)
    const turnstile = useTurnstile()

    async function handleSubmit(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        setErrors({})
        setLoading(true)

        const formData = new FormData(event.currentTarget)
        formData.set("cf-turnstile-response", turnstile.token)
        const nextErrors: Errors = {}
        const name = formData.get("name")?.toString().trim() ?? ""
        const phone = formData.get("phone")?.toString().trim() ?? ""
        const password = formData.get("password")?.toString() ?? ""
        const confirmation = formData.get("confPassword")?.toString() ?? ""

        if (!name || name.length > 128) nextErrors.name = "Enter a name no longer than 128 characters."
        if (phone && (phone.length < 3 || phone.length > 32)) nextErrors.phone = "Enter a valid phone number."
        if (password.length < 8) nextErrors.password = "Use at least 8 characters."
        if (confirmation !== password) nextErrors.confPassword = "The passwords don’t match."

        if (Object.keys(nextErrors).length > 0) {
            setErrors(nextErrors)
            setLoading(false)
            return
        }

        const result = await handleRegistration(formData, token)
        setLoading(false)
        if (!result.ok) {
            turnstile.reset()
            setErrors({root: result.message || "Unable to finish registration."})
            return
        }
        setComplete(true)
    }

    if (complete) {
        return (
            <AuthCard
                icon={CircleCheck}
                title="Your account is ready"
                description="Registration is complete. You can now sign in to the EduLink staff app."
            >
                <div className="space-y-5">
                    <Alert variant="success">
                        <CircleCheck />
                        <AlertTitle>Welcome to EduLink</AlertTitle>
                        <AlertDescription>Your staff account has been created successfully.</AlertDescription>
                    </Alert>
                    <Button className="w-full" render={<Link href={`/auth/login?email=${encodeURIComponent(email)}`} />}>
                        Continue to staff login
                    </Button>
                </div>
            </AuthCard>
        )
    }

    return (
        <AuthCard
            icon={UserRoundCheck}
            title="Finish your registration"
            description="Create your profile and choose the password you’ll use for the staff app."
        >
            <Form onSubmit={handleSubmit} errors={errors} className="space-y-5">
                {errors.root && (
                    <Alert variant="error">
                        <TriangleAlert />
                        <AlertTitle>Registration failed</AlertTitle>
                        <AlertDescription>{errors.root}</AlertDescription>
                    </Alert>
                )}

                <Field>
                    <FieldLabel htmlFor="email">Email address</FieldLabel>
                    <Input value={email} disabled id="email" type="email" />
                    <FieldDescription>This address was verified by your registration link.</FieldDescription>
                </Field>

                <Field name="name">
                    <FieldLabel htmlFor="name">Name</FieldLabel>
                    <Input autoFocus id="name" name="name" required autoComplete="name" placeholder="Your full name" />
                    <FieldError />
                </Field>

                <Field name="phone">
                    <FieldLabel htmlFor="phone">Phone number <span className="font-normal text-muted-foreground">(optional)</span></FieldLabel>
                    <Input id="phone" name="phone" autoComplete="tel" placeholder="Your phone number" />
                    <FieldError />
                </Field>

                <div className="grid gap-5 sm:grid-cols-2">
                    <Field name="password">
                        <FieldLabel htmlFor="password">Password</FieldLabel>
                        <Input
                            id="password"
                            name="password"
                            type="password"
                            autoComplete="new-password"
                            placeholder="At least 8 characters"
                            required
                        />
                        <FieldError />
                    </Field>
                    <Field name="confPassword">
                        <FieldLabel htmlFor="confPassword">Confirm password</FieldLabel>
                        <Input
                            id="confPassword"
                            name="confPassword"
                            type="password"
                            autoComplete="new-password"
                            placeholder="Enter it again"
                            required
                        />
                        <FieldError />
                    </Field>
                </div>

                <Field>
                    <FieldLabel>Verification</FieldLabel>
                    {turnstile.configured ? (
                        <TurnstileWidget
                            key={turnstile.widgetKey}
                            action="finish-registration"
                            onTokenChange={turnstile.setToken}
                        />
                    ) : (
                        <Alert variant="error">
                            <TriangleAlert />
                            <AlertDescription>Turnstile is not configured.</AlertDescription>
                        </Alert>
                    )}
                </Field>

                <p className="text-xs leading-5 text-muted-foreground">
                    By continuing, you agree to the{" "}
                    <Link className="font-medium text-foreground underline underline-offset-4" href="/legal/terms" target="_blank">
                        Terms of Service
                    </Link>{" "}
                    and acknowledge the{" "}
                    <Link className="font-medium text-foreground underline underline-offset-4" href="/legal/privacy" target="_blank">
                        Privacy Policy
                    </Link>.
                </p>

                <Button className="w-full" loading={loading} disabled={loading || !turnstile.configured || !turnstile.token} type="submit">
                    Create staff account
                </Button>
            </Form>
        </AuthCard>
    )
}

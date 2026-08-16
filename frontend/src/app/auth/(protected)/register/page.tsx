"use client"

import React, {useState} from "react"
import Link from "next/link"
import {CircleCheck, TriangleAlert, UserPlus} from "lucide-react"
import {handleRegistrationLinkSend} from "@/app/auth/(protected)/register/actions"
import AuthCard from "@/components/auth/auth-card"
import {TurnstileWidget, useTurnstile} from "@/components/turnstile"
import {Alert, AlertDescription, AlertTitle} from "@/components/ui/alert"
import {Button} from "@/components/ui/button"
import {Field, FieldDescription, FieldLabel} from "@/components/ui/field"
import {Form} from "@/components/ui/form"
import {Input} from "@/components/ui/input"

type Errors = Record<string, string | string[]>

export default function Page() {
    const [loading, setLoading] = useState(false)
    const [errors, setErrors] = useState<Errors>({})
    const [sentEmail, setSentEmail] = useState("")
    const turnstile = useTurnstile()

    async function handleSubmit(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        setErrors({})
        setLoading(true)

        const formData = new FormData(event.currentTarget)
        formData.set("cf-turnstile-response", turnstile.token)
        const email = formData.get("email")?.toString().trim().toLocaleLowerCase() ?? ""
        if (email.length < 5 || email.length > 254 || !email.includes("@")) {
            setErrors({email: "Enter a valid email address."})
            setLoading(false)
            return
        }

        const result = await handleRegistrationLinkSend(formData)
        setLoading(false)
        if (!result.ok) {
            turnstile.reset()
            setErrors({root: result.message || "Unable to start registration."})
            return
        }
        setSentEmail(email)
    }

    if (sentEmail) {
        return (
            <AuthCard
                icon={CircleCheck}
                title="Check your inbox"
                description="Use the secure link we sent to finish creating your staff account."
            >
                <div className="space-y-5">
                    <Alert variant="success">
                        <CircleCheck />
                        <AlertTitle>Registration link sent</AlertTitle>
                        <AlertDescription>
                            We sent the next step to {sentEmail}. The link may take a minute to arrive.
                        </AlertDescription>
                    </Alert>
                    <Button className="w-full" render={<Link href={`/auth/login?email=${encodeURIComponent(sentEmail)}`} />}>
                        Return to staff login
                    </Button>
                    <Button className="w-full" variant="ghost" onClick={() => setSentEmail("")}>
                        Use another email
                    </Button>
                </div>
            </AuthCard>
        )
    }

    return (
        <AuthCard
            icon={UserPlus}
            title="Create a staff account"
            description="Start with your work email. We’ll send you a secure link to finish setting up your account."
        >
            <Form onSubmit={handleSubmit} errors={errors} className="space-y-5">
                {errors.root && (
                    <Alert variant="error">
                        <TriangleAlert />
                        <AlertTitle>Registration couldn&apos;t start</AlertTitle>
                        <AlertDescription>{errors.root}</AlertDescription>
                    </Alert>
                )}

                <Field name="email">
                    <FieldLabel htmlFor="email">Work email</FieldLabel>
                    <Input
                        autoFocus
                        id="email"
                        name="email"
                        type="email"
                        autoComplete="email"
                        placeholder="you@school.edu"
                        required
                    />
                    <FieldDescription>Students and parents receive portal accounts from their school.</FieldDescription>
                </Field>

                <Field>
                    <FieldLabel>Verification</FieldLabel>
                    {turnstile.configured ? (
                        <TurnstileWidget
                            key={turnstile.widgetKey}
                            action="register-link"
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
                    Send registration link
                </Button>
                <p className="text-center text-sm text-muted-foreground">
                    Already have a staff account?{" "}
                    <Link className="font-medium text-foreground underline underline-offset-4" href="/auth/login">Sign in</Link>
                </p>
            </Form>
        </AuthCard>
    )
}

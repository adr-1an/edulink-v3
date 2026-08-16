"use client"

import React, {useState} from "react"
import Link from "next/link"
import {ArrowLeft, CircleCheck, Mail, TriangleAlert} from "lucide-react"
import {handleSendPasswordReset} from "@/app/auth/(protected)/forgot/actions"
import AuthCard from "@/components/auth/auth-card"
import {Alert, AlertDescription, AlertTitle} from "@/components/ui/alert"
import {Button} from "@/components/ui/button"
import {Field, FieldDescription, FieldLabel} from "@/components/ui/field"
import {Input} from "@/components/ui/input"

export default function Page() {
    const [email, setEmail] = useState("")
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState("")
    const [sent, setSent] = useState(false)

    const normalizedEmail = email.trim().toLocaleLowerCase()
    const isValid = normalizedEmail.length >= 5
        && normalizedEmail.length <= 254
        && normalizedEmail.includes("@")

    async function handleSubmit(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        if (!isValid || loading) return

        setLoading(true)
        setError("")
        const result = await handleSendPasswordReset(normalizedEmail)
        setLoading(false)

        if (!result.ok) return setError(result.message || "Unable to send a reset link.")
        setSent(true)
    }

    if (sent) {
        return (
            <AuthCard
                icon={CircleCheck}
                title="Check your inbox"
                description="If a staff account exists for that email, its reset link is on the way."
            >
                <div className="space-y-5">
                    <Alert variant="success">
                        <CircleCheck />
                        <AlertTitle>Reset link requested</AlertTitle>
                        <AlertDescription>
                            We sent password-reset instructions to {normalizedEmail}. The link may take a minute to arrive.
                        </AlertDescription>
                    </Alert>
                    <Button className="w-full" render={<Link href={`/auth/login?email=${encodeURIComponent(normalizedEmail)}`} />}>
                        Continue to staff login
                    </Button>
                    <Button className="w-full" variant="ghost" onClick={() => setSent(false)}>
                        Use another email
                    </Button>
                </div>
            </AuthCard>
        )
    }

    return (
        <AuthCard
            icon={Mail}
            title="Reset your password"
            description="Enter the email address connected to your staff account and we’ll send you a secure reset link."
        >
            <form className="space-y-5" onSubmit={handleSubmit}>
                {error && (
                    <Alert variant="error">
                        <TriangleAlert />
                        <AlertTitle>Couldn&apos;t send the link</AlertTitle>
                        <AlertDescription>{error}</AlertDescription>
                    </Alert>
                )}

                <Field>
                    <FieldLabel htmlFor="email">Email address</FieldLabel>
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
                    <FieldDescription>Use the email address you use for the staff app.</FieldDescription>
                </Field>

                <Button className="w-full" loading={loading} disabled={!isValid || loading} type="submit">
                    Send reset link
                </Button>
                <Button className="w-full" variant="ghost" render={<Link href="/auth/login" />}>
                    <ArrowLeft /> Back to staff login
                </Button>
            </form>
        </AuthCard>
    )
}

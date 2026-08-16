"use client"

import React, {use, useState} from "react"
import Link from "next/link"
import {CircleCheck, KeyRound, TriangleAlert} from "lucide-react"
import {handleResetPassword} from "@/app/auth/(protected)/reset/[token]/actions"
import AuthCard from "@/components/auth/auth-card"
import {Alert, AlertDescription, AlertTitle} from "@/components/ui/alert"
import {Button} from "@/components/ui/button"
import {Field, FieldDescription, FieldLabel} from "@/components/ui/field"
import {Input} from "@/components/ui/input"

export default function Page({params}: {params: Promise<{token: string}>}) {
    const {token} = use(params)
    const [newPassword, setNewPassword] = useState("")
    const [confirmPassword, setConfirmPassword] = useState("")
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState("")
    const [complete, setComplete] = useState(false)

    const passwordsMatch = confirmPassword === newPassword
    const isValid = newPassword.length >= 8 && passwordsMatch

    async function onSubmit(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        if (!isValid || loading) return

        setLoading(true)
        setError("")
        const result = await handleResetPassword(newPassword, token)
        setLoading(false)

        if (!result.ok) return setError(result.message || "Unable to reset your password.")
        setComplete(true)
    }

    if (complete) {
        return (
            <AuthCard
                icon={CircleCheck}
                title="Password updated"
                description="Your new password is ready to use."
            >
                <div className="space-y-5">
                    <Alert variant="success">
                        <CircleCheck />
                        <AlertTitle>Reset complete</AlertTitle>
                        <AlertDescription>You can now sign in to your staff account with the new password.</AlertDescription>
                    </Alert>
                    <Button className="w-full" render={<Link href="/auth/login" />}>Continue to staff login</Button>
                </div>
            </AuthCard>
        )
    }

    return (
        <AuthCard
            icon={KeyRound}
            title="Choose a new password"
            description="Use at least eight characters and avoid reusing a password from another service."
        >
            <form className="space-y-5" onSubmit={onSubmit}>
                {error && (
                    <Alert variant="error">
                        <TriangleAlert />
                        <AlertTitle>Couldn&apos;t reset your password</AlertTitle>
                        <AlertDescription>{error}</AlertDescription>
                    </Alert>
                )}

                <Field>
                    <FieldLabel htmlFor="newPassword">New password</FieldLabel>
                    <Input
                        autoFocus
                        id="newPassword"
                        name="newPassword"
                        type="password"
                        autoComplete="new-password"
                        placeholder="At least 8 characters"
                        required
                        value={newPassword}
                        onChange={(event) => setNewPassword(event.target.value)}
                    />
                    <FieldDescription>
                        {newPassword && newPassword.length < 8
                            ? "Your password needs at least 8 characters."
                            : "Use at least 8 characters."}
                    </FieldDescription>
                </Field>

                <Field>
                    <FieldLabel htmlFor="confirmPassword">Confirm new password</FieldLabel>
                    <Input
                        id="confirmPassword"
                        type="password"
                        autoComplete="new-password"
                        placeholder="Enter it again"
                        required
                        value={confirmPassword}
                        onChange={(event) => setConfirmPassword(event.target.value)}
                    />
                    {confirmPassword && !passwordsMatch && (
                        <FieldDescription className="text-destructive">The passwords don&apos;t match.</FieldDescription>
                    )}
                </Field>

                <Button className="w-full" loading={loading} type="submit" disabled={loading || !isValid}>
                    Update password
                </Button>
            </form>
        </AuthCard>
    )
}

"use client"

import {use, useEffect, useState} from "react"
import Link from "next/link"
import {CircleCheck, LoaderCircle, MailCheck, TriangleAlert} from "lucide-react"
import {handleEmailChangeVerification} from "@/app/auth/verify-change/actions"
import AuthCard from "@/components/auth/auth-card"
import {Alert, AlertDescription, AlertTitle} from "@/components/ui/alert"
import {Button} from "@/components/ui/button"

type VerificationState =
    | {status: "verifying"}
    | {status: "success"}
    | {status: "error"; message: string}

export default function Page({params}: {params: Promise<{token: string}>}) {
    const {token} = use(params)
    const [state, setState] = useState<VerificationState>({status: "verifying"})

    useEffect(() => {
        let active = true

        void handleEmailChangeVerification(token)
            .then((result) => {
                if (!active) return
                setState(result.ok
                    ? {status: "success"}
                    : {status: "error", message: result.message || "Unable to verify this email change."})
            })
            .catch(() => {
                if (active) setState({status: "error", message: "Network error, please try again."})
            })

        return () => {
            active = false
        }
    }, [token])

    return (
        <AuthCard
            icon={MailCheck}
            title={state.status === "verifying"
                ? "Verifying your email"
                : state.status === "success" ? "Email updated" : "Verification failed"}
            description={state.status === "verifying"
                ? "Keep this page open while we apply the change."
                : state.status === "success"
                    ? "Your new email address has been applied to your staff account."
                    : "We couldn’t apply this email change."}
        >
            {state.status === "verifying" ? (
                <div className="flex items-center gap-3 rounded-xl border bg-muted/25 p-4 text-sm text-muted-foreground">
                    <LoaderCircle className="size-5 animate-spin" /> Verifying secure link…
                </div>
            ) : state.status === "success" ? (
                <div className="space-y-5">
                    <Alert variant="success">
                        <CircleCheck />
                        <AlertTitle>Email change complete</AlertTitle>
                        <AlertDescription>You can return to EduLink and continue using your account.</AlertDescription>
                    </Alert>
                    <Button className="w-full" render={<Link href="/app/staff/profile" />}>Return to profile</Button>
                </div>
            ) : (
                <div className="space-y-5">
                    <Alert variant="error">
                        <TriangleAlert />
                        <AlertTitle>Link not accepted</AlertTitle>
                        <AlertDescription>{state.message}</AlertDescription>
                    </Alert>
                    <Button className="w-full" variant="outline" render={<Link href="/app/staff/profile" />}>
                        Return to profile
                    </Button>
                </div>
            )}
        </AuthCard>
    )
}

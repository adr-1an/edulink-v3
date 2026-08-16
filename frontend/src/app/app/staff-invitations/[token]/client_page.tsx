"use client"

import {useState} from "react"
import Link from "next/link"
import {useRouter} from "next/navigation"
import Cookies from "js-cookie"
import {CalendarClock, CircleCheck, Mail, MailPlus, TriangleAlert, UserRound} from "lucide-react"
import {handleAccept, handleReject} from "@/app/app/staff-invitations/[token]/actions"
import AuthCard from "@/components/auth/auth-card"
import LocalDateTime from "@/components/local-date-time"
import {
    AlertDialog, AlertDialogClose, AlertDialogDescription, AlertDialogFooter,
    AlertDialogHeader, AlertDialogPopup, AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import {Alert, AlertDescription, AlertTitle} from "@/components/ui/alert"
import {Badge} from "@/components/ui/badge"
import {Button} from "@/components/ui/button"

interface Invitation {
    id: string
    sentToEmail: string
    schoolName: string
    sentByName: string
    sentByEmail: string
    status: string
    createdAt: string
    expiresAt: string
}

interface Props {
    canAccept: boolean
    authenticated: boolean
    profileName?: string
    inv: Invitation
    token: string
}

function statusVariant(status: string) {
    if (status === "pending") return "warning" as const
    if (status === "rejected") return "error" as const
    if (status === "accepted") return "success" as const
    return "secondary" as const
}

export default function ClientPage({canAccept, authenticated, profileName, inv, token}: Props) {
    const router = useRouter()
    const [status, setStatus] = useState(inv.status)
    const [rejecting, setRejecting] = useState(false)
    const [accepting, setAccepting] = useState(false)
    const [rejectLoading, setRejectLoading] = useState(false)
    const [error, setError] = useState("")
    const available = status === "pending"

    async function accept() {
        if (!available || accepting) return
        setAccepting(true)
        setError("")
        const result = await handleAccept(token)
        setAccepting(false)
        if (!result.ok) setError(result.message || "Unable to accept this invitation.")
    }

    async function reject() {
        if (!available || rejectLoading) return
        setRejectLoading(true)
        setError("")
        const result = await handleReject(token)
        setRejectLoading(false)
        if (!result.ok) {
            setRejecting(false)
            setError(result.message || "Unable to reject this invitation.")
            return
        }

        setStatus("rejected")
        setRejecting(false)
        router.refresh()
    }

    function continueTo(path: "/auth/login" | "/auth/register") {
        Cookies.set("pendingRedirect", `/app/staff-invitations/${token}`)
        router.push(path)
    }

    return (
        <>
            <AuthCard
                icon={MailPlus}
                eyebrow="Staff invitation"
                title={`Join ${inv.schoolName}`}
                description={`${inv.sentByName} invited you to join this school’s staff workspace.`}
                className="max-w-xl"
            >
                <div className="space-y-5">
                    <div className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border bg-muted/25 p-4">
                        <div>
                            <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">Invitation status</p>
                            <Badge className="mt-2 capitalize" variant={statusVariant(status)}>{status}</Badge>
                        </div>
                        <div className="text-right text-xs text-muted-foreground">
                            <p className="flex items-center justify-end gap-1.5">
                                <CalendarClock className="size-3.5" />
                                Expires <LocalDateTime value={inv.expiresAt} />
                            </p>
                        </div>
                    </div>

                    {error && (
                        <Alert variant="error">
                            <TriangleAlert />
                            <AlertTitle>Invitation couldn&apos;t be updated</AlertTitle>
                            <AlertDescription>{error}</AlertDescription>
                        </Alert>
                    )}

                    {status === "accepted" && (
                        <Alert variant="success">
                            <CircleCheck />
                            <AlertTitle>Invitation accepted</AlertTitle>
                            <AlertDescription>You already belong to this school.</AlertDescription>
                        </Alert>
                    )}
                    {status === "rejected" && (
                        <Alert variant="warning">
                            <TriangleAlert />
                            <AlertTitle>Invitation rejected</AlertTitle>
                            <AlertDescription>This invitation is no longer available.</AlertDescription>
                        </Alert>
                    )}

                    <div className="rounded-2xl border p-4">
                        <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">Sent by</p>
                        <div className="mt-3 flex items-center gap-3">
                            <span className="flex size-10 shrink-0 items-center justify-center rounded-full bg-muted">
                                <UserRound className="size-4 text-muted-foreground" />
                            </span>
                            <div className="min-w-0">
                                <p className="truncate text-sm font-medium">{inv.sentByName}</p>
                                <a className="flex items-center gap-1.5 truncate text-xs text-muted-foreground hover:text-foreground hover:underline" href={`mailto:${inv.sentByEmail}`}>
                                    <Mail className="size-3.5 shrink-0" /> {inv.sentByEmail}
                                </a>
                            </div>
                        </div>
                    </div>

                    {authenticated ? canAccept ? (
                        <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
                            <Button
                                variant="destructive-outline"
                                disabled={!available || accepting}
                                onClick={() => setRejecting(true)}
                            >
                                Reject
                            </Button>
                            <Button loading={accepting} disabled={!available || accepting} onClick={accept}>
                                Accept as {profileName}
                            </Button>
                        </div>
                    ) : (
                        <Alert variant="error">
                            <TriangleAlert />
                            <AlertTitle>This invitation belongs to another email</AlertTitle>
                            <AlertDescription>
                                It was sent to {inv.sentToEmail}. Sign in with that account, change your current account’s email, or request a new invitation.
                                <Link className="mt-1 font-medium text-foreground underline underline-offset-4" href="/app/staff/profile">
                                    Open profile settings
                                </Link>
                            </AlertDescription>
                        </Alert>
                    ) : (
                        <div className="space-y-3">
                            <p className="text-sm text-muted-foreground">
                                Sign in with {inv.sentToEmail}, or create a staff account with that address.
                            </p>
                            <div className="grid gap-2 sm:grid-cols-2">
                                <Button variant="outline" onClick={() => continueTo("/auth/login")}>Staff login</Button>
                                <Button onClick={() => continueTo("/auth/register")}>Create staff account</Button>
                            </div>
                        </div>
                    )}

                    <p className="text-center text-xs text-muted-foreground">
                        Sent <LocalDateTime value={inv.createdAt} /> · Need help?{" "}
                        <a className="font-medium text-foreground underline underline-offset-4" href="mailto:support@vertexapp.net">
                            Contact support
                        </a>
                    </p>
                </div>
            </AuthCard>

            <AlertDialog open={rejecting} onOpenChange={(open) => {
                if (!open && !rejectLoading) setRejecting(false)
            }}>
                <AlertDialogPopup>
                    <AlertDialogHeader>
                        <AlertDialogTitle>Reject this invitation?</AlertDialogTitle>
                        <AlertDialogDescription>
                            You won&apos;t be able to use this link afterward. The school will need to send you another invitation if you change your mind.
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogClose render={<Button variant="outline" disabled={rejectLoading} />}>Cancel</AlertDialogClose>
                        <Button variant="destructive" loading={rejectLoading} disabled={rejectLoading} onClick={reject}>
                            Reject invitation
                        </Button>
                    </AlertDialogFooter>
                </AlertDialogPopup>
            </AlertDialog>
        </>
    )
}

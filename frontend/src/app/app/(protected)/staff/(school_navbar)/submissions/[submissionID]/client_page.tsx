"use client"

import {useState} from "react"
import {useRouter} from "next/navigation"
import {
    ArrowLeft, Award, CheckCircle2, Clock3, Mail, Paperclip, Pencil, RotateCcw, StickyNote,
    Trash2, TriangleAlert, UserRound,
} from "lucide-react"
import {toast} from "sonner"
import LocalDateTime from "@/components/local-date-time"
import PostAttachments from "@/components/app/post-attachments"
import StudentProfilePopover, {portalStudentDisplayName, type PortalStudentSummary} from "@/components/app/student-profile-popover"
import {Badge} from "@/components/ui/badge"
import {Button} from "@/components/ui/button"
import {Card, CardContent, CardDescription, CardHeader, CardTitle} from "@/components/ui/card"
import {
    Dialog, DialogClose, DialogDescription, DialogFooter, DialogHeader,
    DialogPanel, DialogPopup, DialogTitle,
} from "@/components/ui/dialog"
import {
    AlertDialog, AlertDialogClose, AlertDialogDescription, AlertDialogFooter,
    AlertDialogHeader, AlertDialogPopup, AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import {Field, FieldDescription, FieldLabel} from "@/components/ui/field"
import {Form} from "@/components/ui/form"
import {Input} from "@/components/ui/input"
import {Progress, ProgressIndicator, ProgressTrack} from "@/components/ui/progress"
import {Textarea} from "@/components/ui/textarea"
import {useLocale} from "@/i18n/provider"
import {type PostAttachment} from "@/lib/post_attachments"
import {scoreAccent} from "@/lib/score"
import {
    handleDeleteReturnedSubmission, handleGradeSubmission, handleRemoveSubmissionGrade, handleReturnSubmission,
    type DeleteSubmissionError, type GradeSubmissionError, type ReturnSubmissionError,
} from "./actions"

export interface SubmissionGrade {
    score: number
    notes: string | null
    gradedAt: string
}

export interface StaffSubmissionView {
    id: string
    status: "submitted" | "returned"
    submittedBy: PortalStudentSummary
    attachments: PostAttachment[]
    notes: string | null
    submittedAt: string
    grade: SubmissionGrade | null
}

export default function SubmissionClientPage({submission, canReturn, canDelete, canGrade, canRemoveGrade}: {
    submission: StaffSubmissionView
    canReturn: boolean
    canDelete: boolean
    canGrade: boolean
    canRemoveGrade: boolean
}) {
    const router = useRouter()
    const {t} = useLocale()
    const [confirmingReturn, setConfirmingReturn] = useState(false)
    const [confirmingDelete, setConfirmingDelete] = useState(false)
    const [returning, setReturning] = useState(false)
    const [deleting, setDeleting] = useState(false)
    const [grade, setGrade] = useState(submission.grade)
    const [gradeOpen, setGradeOpen] = useState(false)
    const [gradeScore, setGradeScore] = useState("")
    const [gradeNotes, setGradeNotes] = useState("")
    const [savingGrade, setSavingGrade] = useState(false)
    const [removeGradeOpen, setRemoveGradeOpen] = useState(false)
    const [removingGrade, setRemovingGrade] = useState(false)
    const studentName = portalStudentDisplayName(submission.submittedBy)
    const gradeAccent = grade ? scoreAccent(grade.score) : undefined
    const parsedGradeScore = Number(gradeScore)
    const gradeValid = gradeScore.trim() !== ""
        && Number.isInteger(parsedGradeScore)
        && parsedGradeScore >= 0
        && parsedGradeScore <= 100
        && gradeNotes.trim().length <= 2048

    function gradingErrorMessage(error: GradeSubmissionError) {
        switch (error) {
            case "invalid": return t("staff.submission.grade.error.invalid")
            case "unauthorized": return t("staff.submission.grade.error.unauthorized")
            case "forbidden": return t("staff.submission.grade.error.forbidden")
            case "conflict": return t("staff.submission.grade.error.conflict")
            case "validation": return t("staff.submission.grade.error.validation")
            case "server": return t("staff.submission.grade.error.server")
            case "network": return t("staff.submission.grade.error.network")
            default: return t("staff.submission.grade.error.unknown")
        }
    }

    function openGradeEditor() {
        setGradeScore(grade ? String(grade.score) : "")
        setGradeNotes(grade?.notes ?? "")
        setGradeOpen(true)
    }

    async function saveGrade(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        if (!canGrade || submission.status !== "submitted" || !gradeValid || savingGrade) return

        setSavingGrade(true)
        const result = await handleGradeSubmission(submission.id, {
            score: parsedGradeScore,
            notes: gradeNotes,
        })
        setSavingGrade(false)
        if (!result.ok) return toast.error(gradingErrorMessage(result.error))

        setGrade({
            score: parsedGradeScore,
            notes: gradeNotes.trim() || null,
            gradedAt: new Date().toISOString(),
        })
        setGradeOpen(false)
        toast.success(t(grade ? "staff.submission.grade.updated" : "staff.submission.grade.created"))
        router.refresh()
    }

    async function removeGrade() {
        if (!canRemoveGrade || !grade || removingGrade) return

        setRemovingGrade(true)
        const result = await handleRemoveSubmissionGrade(submission.id)
        setRemovingGrade(false)
        if (!result.ok) return toast.error(gradingErrorMessage(result.error))

        setGrade(null)
        setRemoveGradeOpen(false)
        toast.success(t("staff.submission.grade.removed"))
        router.refresh()
    }

    function returnErrorMessage(error: ReturnSubmissionError) {
        switch (error) {
            case "invalid": return t("staff.submission.return.error.invalid")
            case "unauthorized": return t("staff.submission.return.error.unauthorized")
            case "forbidden": return t("staff.submission.return.error.forbidden")
            case "server": return t("staff.submission.return.error.server")
            case "network": return t("staff.submission.return.error.network")
            default: return t("staff.submission.return.error.unknown")
        }
    }

    async function returnSubmission() {
        if (returning) return
        setReturning(true)
        const result = await handleReturnSubmission(submission.id)
        setReturning(false)

        if (!result.ok) {
            toast.error(returnErrorMessage(result.error))
            return
        }

        setConfirmingReturn(false)
        toast.success(t("staff.submission.return.success", {name: studentName}))
        router.back()
        router.refresh()
    }

    function deleteErrorMessage(error: DeleteSubmissionError) {
        switch (error) {
            case "invalid": return t("staff.submission.delete.error.invalid")
            case "unauthorized": return t("staff.submission.delete.error.unauthorized")
            case "forbidden": return t("staff.submission.delete.error.forbidden")
            case "conflict": return t("staff.submission.delete.error.conflict")
            case "server": return t("staff.submission.delete.error.server")
            case "network": return t("staff.submission.delete.error.network")
            default: return t("staff.submission.delete.error.unknown")
        }
    }

    async function deleteSubmission() {
        if (deleting || submission.status !== "returned") return
        setDeleting(true)
        const result = await handleDeleteReturnedSubmission(submission.id)
        setDeleting(false)

        if (!result.ok) {
            toast.error(deleteErrorMessage(result.error))
            return
        }

        setConfirmingDelete(false)
        toast.success(t("staff.submission.delete.success", {name: studentName}))
        router.back()
        router.refresh()
    }

    return (
        <div className="mx-auto w-full max-w-6xl space-y-6">
            <Button variant="ghost" className="-ml-3" onClick={() => router.back()}>
                <ArrowLeft /> {t("staff.submission.back")}
            </Button>

            <header className="flex flex-wrap items-start justify-between gap-4">
                <div className="space-y-3">
                    <Badge variant={submission.status === "returned" ? "warning" : "success"}>
                        {submission.status === "returned" ? <RotateCcw /> : <CheckCircle2 />}
                        {t(submission.status === "returned" ? "staff.submission.returned" : "staff.submission.submitted")}
                    </Badge>
                    <div>
                        <h1 className="wrap-break-word text-3xl font-semibold tracking-tight sm:text-4xl">
                            {t("staff.submission.title", {name: studentName})}
                        </h1>
                        <p className="mt-2 flex flex-wrap items-center gap-1.5 text-sm text-muted-foreground">
                            <Clock3 className="size-4" />
                            {t("staff.submission.received")} <LocalDateTime value={submission.submittedAt} />
                        </p>
                    </div>
                </div>
                <div className="flex flex-wrap items-center gap-2">
                    {canReturn && submission.status === "submitted" && (
                        <Button variant="outline" onClick={() => setConfirmingReturn(true)}>
                            <RotateCcw /> {t("staff.submission.return.action")}
                        </Button>
                    )}
                    {canDelete && submission.status === "returned" && (
                        <Button variant="destructive-outline" onClick={() => setConfirmingDelete(true)}>
                            <Trash2 /> {t("staff.submission.delete.action")}
                        </Button>
                    )}
                </div>
            </header>

            <div className="grid items-start gap-5 lg:grid-cols-[minmax(0,1fr)_20rem]">
                <div className="space-y-5">
                    <Card style={gradeAccent ? {borderColor: `color-mix(in oklab, ${gradeAccent} 28%, var(--border))`} : undefined}>
                        <CardHeader className="flex-row items-start justify-between gap-4">
                            <div>
                                <CardTitle className="flex items-center gap-2"><Award /> {t("staff.submission.grade.title")}</CardTitle>
                                <CardDescription>{t("staff.submission.grade.description")}</CardDescription>
                            </div>
                            {grade && <Badge variant="outline">{t("staff.submission.grade.graded")}</Badge>}
                        </CardHeader>
                        <CardContent className="space-y-5">
                            {grade ? (
                                <>
                                    <div
                                        className="rounded-2xl border p-5 transition-colors"
                                        style={gradeAccent ? {
                                            backgroundColor: `color-mix(in oklab, ${gradeAccent} 8%, transparent)`,
                                            borderColor: `color-mix(in oklab, ${gradeAccent} 24%, var(--border))`,
                                        } : undefined}
                                    >
                                        <div className="flex flex-wrap items-end justify-between gap-3">
                                            <div>
                                                <p className="text-sm font-medium text-muted-foreground">{t("staff.submission.grade.score")}</p>
                                                <p className="mt-1 text-4xl font-semibold tabular-nums tracking-tight" style={{color: gradeAccent}}>
                                                    {grade.score}<span className="text-xl opacity-65">%</span>
                                                </p>
                                            </div>
                                            <p className="text-xs text-muted-foreground">
                                                {t("staff.submission.grade.gradedAt")} <LocalDateTime value={grade.gradedAt} />
                                            </p>
                                        </div>
                                        <Progress className="mt-4" value={grade.score}>
                                            <ProgressTrack style={{backgroundColor: `color-mix(in oklab, ${gradeAccent} 14%, var(--input))`}}>
                                                <ProgressIndicator style={{backgroundColor: gradeAccent}} />
                                            </ProgressTrack>
                                        </Progress>
                                    </div>
                                    <div>
                                        <p className="text-sm font-medium">{t("staff.submission.grade.feedback")}</p>
                                        {grade.notes ? (
                                            <p className="mt-2 whitespace-pre-wrap wrap-break-word rounded-xl border bg-muted/20 p-4 text-sm leading-7">{grade.notes}</p>
                                        ) : (
                                            <p className="mt-2 rounded-xl border border-dashed p-4 text-sm italic text-muted-foreground">{t("staff.submission.grade.noFeedback")}</p>
                                        )}
                                    </div>
                                    {(canGrade || canRemoveGrade) && (
                                        <div className="flex flex-wrap gap-2 border-t pt-4">
                                            {canGrade && submission.status === "submitted" && <Button variant="outline" onClick={openGradeEditor}><Pencil /> {t("staff.submission.grade.edit")}</Button>}
                                            {canRemoveGrade && <Button variant="destructive-outline" onClick={() => setRemoveGradeOpen(true)}><Trash2 /> {t("staff.submission.grade.remove")}</Button>}
                                        </div>
                                    )}
                                </>
                            ) : (
                                <div className="flex flex-col items-center justify-center rounded-xl border border-dashed px-5 py-10 text-center">
                                    <Award className="mb-3 size-8 text-muted-foreground" />
                                    <p className="font-medium">{t("staff.submission.grade.ungraded")}</p>
                                    <p className="mt-1 max-w-md text-sm text-muted-foreground">{t("staff.submission.grade.ungradedDescription")}</p>
                                    {canGrade && submission.status === "submitted" && <Button className="mt-5" onClick={openGradeEditor}><Award /> {t("staff.submission.grade.action")}</Button>}
                                </div>
                            )}
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2"><StickyNote /> {t("staff.submission.notes")}</CardTitle>
                            <CardDescription>{t("staff.submission.notesDescription")}</CardDescription>
                        </CardHeader>
                        <CardContent>
                            {submission.notes ? (
                                <p className="whitespace-pre-wrap wrap-break-word rounded-xl border bg-muted/20 p-4 text-sm leading-7">
                                    {submission.notes}
                                </p>
                            ) : (
                                <p className="rounded-xl border border-dashed p-5 text-sm italic text-muted-foreground">
                                    {t("staff.submission.noNotes")}
                                </p>
                            )}
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2"><Paperclip /> {t("staff.submission.attachments")}</CardTitle>
                            <CardDescription>{t("staff.submission.attachmentsDescription")}</CardDescription>
                        </CardHeader>
                        <CardContent>
                            {submission.attachments.length > 0 ? (
                                <PostAttachments
                                    attachments={submission.attachments}
                                    postTitle={studentName}
                                    viewerTitle={t("staff.submission.attachments")}
                                    viewerDescription={t("staff.submission.attachmentCount", {count: submission.attachments.length})}
                                />
                            ) : (
                                <div className="flex flex-col items-center justify-center rounded-xl border border-dashed px-5 py-10 text-center text-muted-foreground">
                                    <Paperclip className="mb-3 size-7" />
                                    <p className="text-sm font-medium text-foreground">{t("staff.submission.noAttachments")}</p>
                                    <p className="mt-1 text-xs">{t("staff.submission.noAttachmentsDescription")}</p>
                                </div>
                            )}
                        </CardContent>
                    </Card>
                </div>

                <Card className="lg:sticky lg:top-6">
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2"><UserRound /> {t("staff.submission.student")}</CardTitle>
                        <CardDescription>{t("staff.submission.studentDescription")}</CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <StudentProfilePopover student={submission.submittedBy} showEmail />
                        <div className="space-y-3 border-t pt-4 text-sm">
                            <a className="flex min-w-0 items-center gap-2 text-muted-foreground underline-offset-4 hover:text-foreground hover:underline" href={`mailto:${submission.submittedBy.email}`}>
                                <Mail className="size-4 shrink-0" />
                                <span className="truncate">{submission.submittedBy.email}</span>
                            </a>
                            <div className="flex items-start gap-2 text-muted-foreground">
                                <Clock3 className="mt-0.5 size-4 shrink-0" />
                                <span>{t("staff.submission.submittedAt")}<br /><LocalDateTime value={submission.submittedAt} /></span>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            </div>

            <Dialog open={gradeOpen} onOpenChange={(open) => {
                if (!savingGrade) setGradeOpen(open)
            }}>
                <DialogPopup className="sm:max-w-lg">
                    <DialogHeader>
                        <DialogTitle>{t(grade ? "staff.submission.grade.editTitle" : "staff.submission.grade.createTitle")}</DialogTitle>
                        <DialogDescription>{t("staff.submission.grade.formDescription", {name: studentName})}</DialogDescription>
                    </DialogHeader>
                    <Form className="contents" onSubmit={saveGrade}>
                        <DialogPanel className="space-y-5">
                            <Field>
                                <FieldLabel htmlFor="submission-grade-score">{t("staff.submission.grade.scorePercentage")}</FieldLabel>
                                <div className="relative">
                                    <Input
                                        id="submission-grade-score"
                                        className="pr-10 text-lg font-semibold tabular-nums"
                                        type="number"
                                        min={0}
                                        max={100}
                                        step={1}
                                        value={gradeScore}
                                        onChange={(event) => setGradeScore(event.target.value)}
                                        disabled={savingGrade}
                                        autoFocus
                                        required
                                    />
                                    <span className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">%</span>
                                </div>
                                <FieldDescription>{t("staff.submission.grade.scoreHelp")}</FieldDescription>
                            </Field>
                            <Field>
                                <div className="flex items-center justify-between gap-3">
                                    <FieldLabel htmlFor="submission-grade-notes">{t("staff.submission.grade.feedback")}</FieldLabel>
                                    <span className="text-xs tabular-nums text-muted-foreground">{gradeNotes.length}/2048</span>
                                </div>
                                <Textarea
                                    id="submission-grade-notes"
                                    className="min-h-32 resize-y"
                                    value={gradeNotes}
                                    onChange={(event) => setGradeNotes(event.target.value)}
                                    maxLength={2048}
                                    placeholder={t("staff.submission.grade.feedbackPlaceholder")}
                                    disabled={savingGrade}
                                />
                                <FieldDescription>{t("staff.submission.grade.feedbackHelp")}</FieldDescription>
                            </Field>
                        </DialogPanel>
                        <DialogFooter>
                            <DialogClose render={<Button variant="outline" disabled={savingGrade} />}>{t("common.cancel")}</DialogClose>
                            <Button type="submit" loading={savingGrade} disabled={!gradeValid || savingGrade}>
                                {t(grade ? "staff.submission.grade.save" : "staff.submission.grade.publish")}
                            </Button>
                        </DialogFooter>
                    </Form>
                </DialogPopup>
            </Dialog>

            <AlertDialog open={removeGradeOpen} onOpenChange={(open) => {
                if (!removingGrade) setRemoveGradeOpen(open)
            }}>
                <AlertDialogPopup>
                    <AlertDialogHeader>
                        <AlertDialogTitle>{t("staff.submission.grade.removeTitle")}</AlertDialogTitle>
                        <AlertDialogDescription>{t("staff.submission.grade.removeDescription", {name: studentName})}</AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogClose render={<Button variant="outline" disabled={removingGrade} />}>{t("common.cancel")}</AlertDialogClose>
                        <Button variant="destructive" loading={removingGrade} disabled={removingGrade} onClick={removeGrade}>
                            {t("staff.submission.grade.removeConfirm")}
                        </Button>
                    </AlertDialogFooter>
                </AlertDialogPopup>
            </AlertDialog>

            <AlertDialog open={confirmingReturn} onOpenChange={(open) => {
                if (!returning) setConfirmingReturn(open)
            }}>
                <AlertDialogPopup>
                    <AlertDialogHeader>
                        <div className="mb-2 flex size-11 items-center justify-center rounded-xl bg-amber-500/10 text-amber-700 dark:text-amber-400">
                            <RotateCcw className="size-5" />
                        </div>
                        <AlertDialogTitle>{t("staff.submission.return.title")}</AlertDialogTitle>
                        <AlertDialogDescription className="space-y-2 text-left">
                            <span className="block">{t("staff.submission.return.description", {name: studentName})}</span>
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogClose render={<Button variant="outline" disabled={returning} />}>
                            {t("common.cancel")}
                        </AlertDialogClose>
                        <Button loading={returning} disabled={returning} onClick={returnSubmission}>
                            <RotateCcw /> {t("staff.submission.return.confirm")}
                        </Button>
                    </AlertDialogFooter>
                </AlertDialogPopup>
            </AlertDialog>

            <AlertDialog open={confirmingDelete} onOpenChange={(open) => {
                if (!deleting) setConfirmingDelete(open)
            }}>
                <AlertDialogPopup className="border-destructive/30 sm:max-w-lg">
                    <AlertDialogHeader>
                        <div className="mb-2 flex size-11 items-center justify-center rounded-xl bg-destructive/10 text-destructive">
                            <TriangleAlert className="size-5" />
                        </div>
                        <AlertDialogTitle className="text-destructive">{t("staff.submission.delete.title")}</AlertDialogTitle>
                        <AlertDialogDescription className="space-y-2 text-left">
                            <span className="block">{t("staff.submission.delete.description", {name: studentName})}</span>
                            <span className="block font-medium text-destructive">{t("staff.submission.delete.warning")}</span>
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogClose render={<Button variant="outline" disabled={deleting} />}>
                            {t("common.cancel")}
                        </AlertDialogClose>
                        <Button variant="destructive" loading={deleting} disabled={deleting} onClick={deleteSubmission}>
                            <Trash2 /> {t("staff.submission.delete.confirm")}
                        </Button>
                    </AlertDialogFooter>
                </AlertDialogPopup>
            </AlertDialog>
        </div>
    )
}

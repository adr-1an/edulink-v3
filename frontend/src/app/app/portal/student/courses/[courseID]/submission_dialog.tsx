"use client"

import {useId, useState} from "react"
import {useRouter} from "next/navigation"
import {
    AlertTriangle, Award, CheckCircle2, CircleX, Clock3, ExternalLink, File, LoaderCircle, Mail,
    Paperclip, Phone, RotateCcw, Send, Upload, X,
} from "lucide-react"
import {toast} from "sonner"
import PostAttachments from "@/components/app/post-attachments"
import {Avatar, AvatarFallback} from "@/components/ui/avatar"
import {Button} from "@/components/ui/button"
import LocalDateTime from "@/components/local-date-time"
import {
    Dialog, DialogClose, DialogDescription, DialogFooter, DialogHeader,
    DialogPanel, DialogPopup, DialogTitle,
} from "@/components/ui/dialog"
import {Field, FieldDescription, FieldLabel} from "@/components/ui/field"
import {Popover, PopoverDescription, PopoverPopup, PopoverTitle, PopoverTrigger} from "@/components/ui/popover"
import {Progress} from "@/components/ui/progress"
import {Textarea} from "@/components/ui/textarea"
import {scoreAccent} from "@/lib/score"
import type {PostAttachment} from "@/lib/post_attachments"
import {uploadToPresignedURL} from "@/lib/upload_to_presigned_url"
import {useLocale} from "@/i18n/provider"
import {
    handleBeginAssignmentSubmission,
    handleCompleteSubmissionAttachment,
    handleDeleteSubmissionAttachment,
    handleInitSubmissionAttachment,
    handleSubmitAssignment,
} from "../../assignments/actions"
import type {PortalCourseAssignment, PortalSubmissionAttachment} from "./assignments_section"

type UploadStatus = "ready" | "preparing" | "uploading" | "verifying" | "done" | "error"

interface PendingAttachment {
    id: string
    file: File
    status: UploadStatus
    progress: number
    error?: string
}

const maxFileSize = 50 * 1024 * 1024
const allowedTypes = new Set([
    "image/jpeg",
    "image/png",
    "image/gif",
    "image/webp",
    "application/zip",
    "application/x-zip-compressed",
    "application/pdf",
    "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
    "application/msword",
    "application/gzip",
    "application/x-gzip",
    "audio/mp4",
    "audio/wav",
    "audio/mpeg",
    "video/mpeg",
    "video/webm",
    "application/x-7z-compressed",
])
const acceptedFiles = ".jpg,.jpeg,.png,.gif,.webp,.zip,.pdf,.doc,.docx,.gz,.7z,.mp3,.mp4,.wav,.mpeg,.webm"

function formatBytes(bytes: number) {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export default function SubmissionDialog({assignment, onOpenChange}: {
    assignment: PortalCourseAssignment | null
    onOpenChange: (open: boolean) => void
}) {
    const [submitting, setSubmitting] = useState(false)

    return (
        <Dialog open={assignment !== null} onOpenChange={(open) => {
            if (!submitting) onOpenChange(open)
        }}>
            {assignment && (
                <SubmissionDialogContent
                    assignment={assignment}
                    submitting={submitting}
                    setSubmitting={setSubmitting}
                    onClose={() => onOpenChange(false)}
                    key={assignment.id}
                />
            )}
        </Dialog>
    )
}

function SubmissionDialogContent({assignment, submitting, setSubmitting, onClose}: {
    assignment: PortalCourseAssignment
    submitting: boolean
    setSubmitting: (submitting: boolean) => void
    onClose: () => void
}) {
    const router = useRouter()
    const {t} = useLocale()
    const returned = assignment.submission?.status === "returned"
    const [notes, setNotes] = useState(returned ? "" : assignment.submission?.notes ?? "")
    const [attachments, setAttachments] = useState<PendingAttachment[]>([])
    const [confirming, setConfirming] = useState(false)
    const [removingAttachment, setRemovingAttachment] = useState<string | null>(null)
    const [existingAttachments, setExistingAttachments] = useState(returned ? [] : assignment.submission?.attachments ?? [])
    const returnedAttachments = returned ? assignment.submission?.attachments ?? [] : []
    const submitted = assignment.submission?.status === "submitted"
    const grade = assignment.submission?.status === "submitted" ? assignment.submission.grade : null

    function updateAttachment(id: string, update: Partial<PendingAttachment>) {
        setAttachments((current) => current.map((attachment) => attachment.id === id ? {...attachment, ...update} : attachment))
    }

    async function submit() {
        if (submitting) return
        setSubmitting(true)

        const begun = await handleBeginAssignmentSubmission(assignment.id, notes)
        if (!begun.ok) {
            toast.error(errorMessage(begun.code, t))
            setSubmitting(false)
            return
        }
        const submissionID = begun.submissionID

        const pendingAttachments = attachments.filter(({status}) => status !== "done")
        let nextAttachmentIndex = 0
        let failedUploads = 0

        async function uploadAttachment(attachment: PendingAttachment) {
            updateAttachment(attachment.id, {status: "preparing", progress: 0, error: undefined})
            const initialized = await handleInitSubmissionAttachment(submissionID, {
                fileName: attachment.file.name,
                declaredSize: attachment.file.size,
                declaredContentType: attachment.file.type,
            })

            if (!initialized.ok) {
                const message = errorMessage(initialized.code, t)
                updateAttachment(attachment.id, {status: "error", error: message})
                return false
            }

            try {
                updateAttachment(attachment.id, {status: "uploading", progress: 0})
                await uploadToPresignedURL(attachment.file, initialized.upload.url, (progress) => {
                    updateAttachment(attachment.id, {progress})
                })
                updateAttachment(attachment.id, {status: "verifying", progress: 100})
            } catch {
                // Calling completion after a failed storage upload lets the backend
                // mark the pending object as failed and clean up any partial data.
                await handleCompleteSubmissionAttachment(initialized.upload.objectID, initialized.upload.completionToken)
                updateAttachment(attachment.id, {status: "error", error: t("assignments.submission.storageError")})
                return false
            }

            const completed = await handleCompleteSubmissionAttachment(initialized.upload.objectID, initialized.upload.completionToken)
            if (!completed.ok) {
                const message = errorMessage(completed.code, t)
                updateAttachment(attachment.id, {status: "error", error: message})
                return false
            }
            updateAttachment(attachment.id, {status: "done", progress: 100})
            return true
        }

        async function uploadWorker() {
            while (nextAttachmentIndex < pendingAttachments.length) {
                const attachment = pendingAttachments[nextAttachmentIndex]
                nextAttachmentIndex += 1
                if (!await uploadAttachment(attachment)) failedUploads += 1
            }
        }

        const uploadConcurrency = Math.min(3, pendingAttachments.length)
        await Promise.all(Array.from({length: uploadConcurrency}, () => uploadWorker()))

        if (failedUploads > 0) {
            toast.error(t("assignments.submission.uploadsFailed", {count: failedUploads}), {
                description: t("assignments.submission.uploadsFailedDescription"),
            })
            setSubmitting(false)
            setConfirming(false)
            return
        }

        const submitted = await handleSubmitAssignment(submissionID)
        if (!submitted.ok) {
            toast.error(errorMessage(submitted.code, t))
            setSubmitting(false)
            setConfirming(false)
            return
        }

        toast.success(t("assignments.submission.success"))
        setSubmitting(false)
        onClose()
        router.refresh()
    }

    async function removeExistingAttachment(attachmentID: string) {
        if (!assignment.submission || removingAttachment) return
        setRemovingAttachment(attachmentID)
        const result = await handleDeleteSubmissionAttachment(assignment.submission.id, attachmentID)
        if (result.ok) {
            setExistingAttachments((current) => current.filter(({id}) => id !== attachmentID))
            toast.success(t("assignments.submission.fileRemoved"))
        } else {
            toast.error(errorMessage(result.code, t))
        }
        setRemovingAttachment(null)
    }

    if (submitted) {
        const accent = grade ? scoreAccent(grade.score) : null
        return (
            <DialogPopup className="sm:max-w-2xl">
                    <DialogHeader>
                        <DialogTitle>{t(grade ? "assignments.grade.detailsTitle" : "assignments.submission.submittedTitle")}</DialogTitle>
                        <DialogDescription>{t(grade ? "assignments.grade.detailsDescription" : "assignments.submission.submittedDescription", {title: assignment.title})}</DialogDescription>
                    </DialogHeader>
                    <DialogPanel className="space-y-5">
                        {grade && accent ? (
                            <section className="space-y-4 rounded-2xl border p-4 sm:p-5" style={{
                                backgroundColor: `color-mix(in oklab, ${accent} 7%, transparent)`,
                                borderColor: `color-mix(in oklab, ${accent} 28%, var(--border))`,
                            }} aria-labelledby="assignment-grade-heading">
                                <div className="flex items-center justify-between gap-4">
                                    <div className="flex items-center gap-3">
                                        <span className="flex size-11 shrink-0 items-center justify-center rounded-xl" style={{backgroundColor: `color-mix(in oklab, ${accent} 16%, transparent)`, color: accent}}>
                                            <Award className="size-5" />
                                        </span>
                                        <div>
                                            <p className="text-sm font-semibold" id="assignment-grade-heading">{t("assignments.grade.graded")}</p>
                                            <p className="text-xs text-muted-foreground">{t("assignments.grade.resultDescription")}</p>
                                        </div>
                                    </div>
                                    <p className="text-3xl font-semibold tabular-nums" style={{color: accent}}>{t("assignments.grade.score", {score: grade.score})}</p>
                                </div>

                                <div>
                                    <p className="text-sm font-medium">{t("assignments.grade.teacherNotes")}</p>
                                    <p className="mt-2 whitespace-pre-wrap wrap-break-word rounded-xl border bg-background/70 p-3 text-sm leading-6">
                                        {grade.notes?.trim() || t("assignments.grade.noTeacherNotes")}
                                    </p>
                                </div>

                                <div className="flex flex-wrap items-center justify-between gap-3 border-t pt-4">
                                    <div className="flex items-center gap-2 text-sm">
                                        <span className="text-muted-foreground">{t("assignments.grade.gradedBy")}</span>
                                        <GraderProfilePopover grader={grade.gradedBy} />
                                    </div>
                                    <span className="flex items-center gap-1.5 text-xs text-muted-foreground">
                                        <Clock3 className="size-3.5" />
                                        <LocalDateTime value={grade.gradedAt} />
                                    </span>
                                </div>
                            </section>
                        ) : (
                            <div className="flex items-center gap-3 rounded-2xl border border-emerald-200 bg-emerald-50 p-4 text-emerald-900 dark:border-emerald-900 dark:bg-emerald-950/30 dark:text-emerald-100">
                                <CheckCircle2 className="size-5 shrink-0" />
                                <p className="text-sm font-medium">{t("assignments.submission.submitted")}</p>
                            </div>
                        )}

                        <section className="space-y-4" aria-labelledby="student-submission-heading">
                            <div>
                                <h3 className="font-semibold" id="student-submission-heading">{t("assignments.submission.workTitle")}</h3>
                                <p className="text-sm text-muted-foreground">{t("assignments.submission.workDescription")}</p>
                            </div>
                            <div>
                            <p className="text-sm font-medium">{t("assignments.submission.notes")}</p>
                            <p className="mt-2 whitespace-pre-wrap rounded-xl border bg-muted/20 p-3 text-sm leading-6">
                                {assignment.submission?.notes.trim() || t("assignments.submission.noNotes")}
                            </p>
                            </div>
                            <div className="space-y-2">
                            <p className="text-sm font-medium">{t("assignments.submission.attachments")}</p>
                            {existingAttachments.length ? (
                                <SubmissionAttachments attachments={existingAttachments} assignmentTitle={assignment.title} />
                            ) : (
                                <p className="text-sm text-muted-foreground">{t("assignments.submission.noAttachments")}</p>
                            )}
                            </div>
                        </section>
                    </DialogPanel>
                    <DialogFooter>
                        <DialogClose render={<Button />}>{t("assignments.submission.close")}</DialogClose>
                    </DialogFooter>
            </DialogPopup>
        )
    }

    return (
        <DialogPopup className="sm:max-w-xl" closeProps={{disabled: submitting}}>
                <DialogHeader>
                    <DialogTitle>{confirming ? t("assignments.submission.confirmTitle") : t("assignments.submission.title")}</DialogTitle>
                    <DialogDescription>
                        {confirming
                            ? t("assignments.submission.confirmDescription")
                            : t("assignments.submission.description", {title: assignment.title})}
                    </DialogDescription>
                </DialogHeader>

                <DialogPanel className="space-y-5">
                    {confirming ? (
                        <>
                            <div className="rounded-2xl border border-amber-300/60 bg-amber-50 p-4 text-amber-950 dark:border-amber-800 dark:bg-amber-950/30 dark:text-amber-100">
                                <div className="flex gap-3">
                                    <AlertTriangle className="mt-0.5 size-5 shrink-0" />
                                    <div>
                                        <p className="font-semibold">{t("assignments.submission.finalWarning")}</p>
                                        <p className="mt-1 text-sm leading-6 opacity-80">{t("assignments.submission.finalWarningDescription")}</p>
                                    </div>
                                </div>
                                <div className="mt-4 grid gap-2 text-sm sm:grid-cols-2">
                                    <Summary label={t("assignments.submission.notes")} value={notes.trim() ? t("assignments.submission.included") : t("assignments.submission.notIncluded")} />
                                    <Summary
                                        label={t("assignments.submission.attachments")}
                                        value={t("assignments.submission.fileCount", {
                                            count: existingAttachments.length + attachments.length,
                                        })}
                                    />
                                </div>
                            </div>
                            {submitting && attachments.length > 0 && <SubmissionUploadProgress attachments={attachments} />}
                        </>
                    ) : (
                        <>
                            {returned && (
                                <div className="space-y-3">
                                    <div className="flex gap-3 rounded-2xl border border-warning/35 bg-warning/[0.07] p-4 text-warning-foreground">
                                        <RotateCcw className="mt-0.5 size-5 shrink-0" />
                                        <div>
                                            <p className="font-semibold">{t("assignments.submission.returnedTitle")}</p>
                                            <p className="mt-1 text-sm leading-6 opacity-80">{t("assignments.submission.returnedDescription")}</p>
                                        </div>
                                    </div>
                                    {returnedAttachments.length > 0 && (
                                        <div className="space-y-2">
                                            <p className="text-sm font-medium">{t("assignments.submission.previousFiles")}</p>
                                            <SubmissionAttachments attachments={returnedAttachments} assignmentTitle={assignment.title} />
                                        </div>
                                    )}
                                </div>
                            )}
                            {existingAttachments.length ? (
                                <div className="space-y-2">
                                    <p className="text-sm font-medium">{t("assignments.submission.uploadedFiles")}</p>
                                    <SubmissionAttachments
                                        attachments={existingAttachments}
                                        assignmentTitle={assignment.title}
                                        removingAttachmentId={removingAttachment}
                                        removeDisabled={removingAttachment !== null || submitting}
                                        onRemove={removeExistingAttachment}
                                    />
                                </div>
                            ) : null}

                            <Field>
                                <FieldLabel htmlFor="submission-notes">{t("assignments.submission.notes")}</FieldLabel>
                                <Textarea
                                    className="min-h-28 resize-y"
                                    id="submission-notes"
                                    value={notes}
                                    maxLength={2048}
                                    disabled={submitting}
                                    placeholder={t("assignments.submission.notesPlaceholder")}
                                    onChange={(event) => setNotes(event.target.value)}
                                />
                                <FieldDescription>{t("assignments.submission.notesDescription", {count: notes.length})}</FieldDescription>
                            </Field>

                            <AttachmentPicker attachments={attachments} disabled={submitting} onChange={setAttachments} />
                        </>
                    )}
                </DialogPanel>

                <DialogFooter>
                    {confirming ? (
                        <>
                            <Button variant="outline" disabled={submitting} onClick={() => setConfirming(false)}>
                                {t("common.back")}
                            </Button>
                            <Button disabled={submitting} onClick={submit}>
                                {submitting ? <LoaderCircle className="animate-spin" /> : <Send />}
                                {submitting ? t("assignments.submission.submitting") : t("assignments.submission.submitFinal")}
                            </Button>
                        </>
                    ) : (
                        <>
                            <DialogClose render={<Button variant="outline" />} disabled={submitting}>{t("common.cancel")}</DialogClose>
                            <Button disabled={submitting} onClick={() => setConfirming(true)}>
                                {assignment.submission?.status === "pending"
                                    ? t("assignments.submission.continue")
                                    : t("assignments.submission.review")}
                                <Send />
                            </Button>
                        </>
                    )}
                </DialogFooter>
        </DialogPopup>
    )
}

function toPreviewAttachment(attachment: PortalSubmissionAttachment): PostAttachment | null {
    if (!attachment.presignedUrl) return null

    try {
        const url = new URL(attachment.presignedUrl)
        if (url.protocol !== "https:" && url.protocol !== "http:") return null
    } catch {
        return null
    }

    return {
        id: attachment.id,
        fileName: attachment.fileName,
        contentType: attachment.contentType,
        presignedUrl: attachment.presignedUrl,
    }
}

function SubmissionAttachments({
    attachments,
    assignmentTitle,
    removingAttachmentId = null,
    removeDisabled = false,
    onRemove,
}: {
    attachments: PortalSubmissionAttachment[]
    assignmentTitle: string
    removingAttachmentId?: string | null
    removeDisabled?: boolean
    onRemove?: (attachmentID: string) => void
}) {
    const {t} = useLocale()
    const previewableAttachments = attachments.flatMap((attachment) => {
        const previewAttachment = toPreviewAttachment(attachment)
        return previewAttachment ? [previewAttachment] : []
    })
    const unavailableAttachments = attachments.filter((attachment) => !toPreviewAttachment(attachment))

    return (
        <>
            <PostAttachments
                attachments={previewableAttachments}
                postTitle={assignmentTitle}
                viewerTitle={t("assignments.submission.attachments")}
                viewerDescription={t("assignments.submission.fileCount", {count: previewableAttachments.length})}
                onRemove={onRemove ? (attachment) => onRemove(attachment.id) : undefined}
                removingAttachmentId={removingAttachmentId}
                removeDisabled={removeDisabled}
            />
            {unavailableAttachments.map((attachment) => (
                <SubmissionAttachmentRow
                    attachment={attachment}
                    removing={removingAttachmentId === attachment.id}
                    removeDisabled={removeDisabled}
                    onRemove={onRemove ? () => onRemove(attachment.id) : undefined}
                    key={attachment.id}
                />
            ))}
        </>
    )
}

function SubmissionAttachmentRow({
    attachment,
    removing = false,
    removeDisabled = false,
    onRemove,
}: {
    attachment: PortalSubmissionAttachment
    removing?: boolean
    removeDisabled?: boolean
    onRemove?: () => void
}) {
    const {t} = useLocale()
    const openURL = attachment.presignedUrl

    return (
        <div className="flex min-w-0 items-center gap-3 rounded-xl border bg-muted/20 p-3">
            <File className="size-4 shrink-0 text-muted-foreground" />
            <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium" title={attachment.fileName}>{attachment.fileName}</p>
                <p className="text-xs text-muted-foreground">{formatBytes(attachment.fileSize)}</p>
            </div>
            {openURL
                ? <CheckCircle2 className="size-4 shrink-0 text-emerald-600" />
                : <AlertTriangle className="size-4 shrink-0 text-amber-600" />}
            {openURL && (
                <Button
                    size="icon-xs"
                    variant="ghost"
                    aria-label={t("attachments.open", {name: attachment.fileName})}
                    render={<a href={openURL} target="_blank" rel="noreferrer" />}
                >
                    <ExternalLink />
                </Button>
            )}
            {onRemove && (
                <Button
                    size="icon-xs"
                    variant="ghost"
                    disabled={removeDisabled}
                    aria-label={t("assignments.submission.removeFile", {name: attachment.fileName})}
                    onClick={onRemove}
                >
                    {removing ? <LoaderCircle className="animate-spin" /> : <X />}
                </Button>
            )}
        </div>
    )
}

function initials(name: string) {
    return name.trim().split(/\s+/).slice(0, 2).map((part) => part[0]).join("").toUpperCase() || "?"
}

function GraderProfilePopover({grader}: {
    grader: NonNullable<NonNullable<PortalCourseAssignment["submission"]>["grade"]>["gradedBy"]
}) {
    const {t} = useLocale()
    return (
        <Popover>
            <PopoverTrigger className="group flex cursor-pointer items-center gap-2 rounded-lg text-left outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2" type="button">
                <Avatar className="size-7 border transition-opacity group-hover:opacity-80">
                    <AvatarFallback className="text-[0.65rem]">{initials(grader.name)}</AvatarFallback>
                </Avatar>
                <span className="font-medium underline-offset-2 group-hover:underline group-focus-visible:underline">{grader.name}</span>
            </PopoverTrigger>
            <PopoverPopup className="w-72" align="start">
                <div className="flex min-w-0 items-start gap-3">
                    <Avatar className="size-10 border"><AvatarFallback>{initials(grader.name)}</AvatarFallback></Avatar>
                    <div className="min-w-0 flex-1">
                        <PopoverTitle className="truncate text-base">{grader.name}</PopoverTitle>
                        <PopoverDescription className="mt-2 space-y-1.5">
                            <a className="flex min-w-0 items-center gap-1.5 underline-offset-4 hover:text-foreground hover:underline" href={`mailto:${grader.email}`}>
                                <Mail className="size-3.5 shrink-0" />
                                <span className="truncate">{grader.email}</span>
                            </a>
                            {grader.phone && (
                                <a className="flex min-w-0 items-center gap-1.5 underline-offset-4 hover:text-foreground hover:underline" href={`tel:${grader.phone}`}>
                                    <Phone className="size-3.5 shrink-0" />
                                    <span className="truncate">{grader.phone}</span>
                                </a>
                            )}
                            {!grader.phone && <p className="text-xs">{t("assignments.grade.noPhone")}</p>}
                        </PopoverDescription>
                    </div>
                </div>
            </PopoverPopup>
        </Popover>
    )
}

function AttachmentPicker({attachments, disabled, onChange}: {
    attachments: PendingAttachment[]
    disabled: boolean
    onChange: (attachments: PendingAttachment[]) => void
}) {
    const inputID = useId()
    const {t} = useLocale()

    function addFiles(files: FileList | null) {
        if (!files) return

        const existing = new Set(attachments.map(({file}) => `${file.name}:${file.size}:${file.lastModified}`))
        const accepted: PendingAttachment[] = []
        const rejected: string[] = []

        for (const file of Array.from(files)) {
            const key = `${file.name}:${file.size}:${file.lastModified}`
            if (existing.has(key)) continue
            if (!file.name || file.name.length > 255 || file.size <= 0 || file.size > maxFileSize || !allowedTypes.has(file.type)) {
                rejected.push(file.name || t("assignments.submission.unnamedFile"))
                continue
            }
            existing.add(key)
            accepted.push({id: crypto.randomUUID(), file, status: "ready", progress: 0})
        }

        if (rejected.length) {
            toast.error(t("assignments.submission.filesRejected", {count: rejected.length}), {
                description: rejected.slice(0, 3).join(", "),
            })
        }
        if (accepted.length) onChange([...attachments, ...accepted])
    }

    return (
        <Field>
            <div className="flex w-full items-center justify-between gap-3">
                <FieldLabel htmlFor={inputID}>{t("assignments.submission.attachments")}</FieldLabel>
                {attachments.length > 0 && <span className="text-xs tabular-nums text-muted-foreground">{t("assignments.submission.selected", {count: attachments.length})}</span>}
            </div>
            <label className={`flex items-center justify-center gap-2 rounded-xl border border-dashed px-4 py-4 text-sm font-medium transition-colors ${disabled ? "pointer-events-none opacity-60" : "cursor-pointer hover:border-primary/40 hover:bg-primary/5"}`} htmlFor={inputID}>
                <Upload className="size-4" /> {t("assignments.submission.addFiles")}
            </label>
            <input
                className="sr-only"
                id={inputID}
                type="file"
                accept={acceptedFiles}
                multiple
                disabled={disabled}
                onChange={(event) => {
                    addFiles(event.currentTarget.files)
                    event.currentTarget.value = ""
                }}
            />
            <FieldDescription>{t("assignments.submission.fileDescription")}</FieldDescription>

            {attachments.length > 0 && (
                <div className="max-h-52 w-full space-y-2 overflow-y-auto pr-1">
                    {attachments.map((attachment) => (
                        <div className="rounded-xl border bg-muted/20 p-3" key={attachment.id}>
                            <div className="flex min-w-0 items-center gap-3">
                                <File className="size-4 shrink-0 text-muted-foreground" />
                                <div className="min-w-0 flex-1">
                                    <p className="truncate text-sm font-medium">{attachment.file.name}</p>
                                    <p className={`mt-0.5 flex items-center gap-1.5 text-xs ${attachment.status === "error" ? "text-destructive" : "text-muted-foreground"}`}>
                                        {["preparing", "uploading", "verifying"].includes(attachment.status) && <LoaderCircle className="size-3.5 animate-spin" />}
                                        {attachment.status === "done" && <CheckCircle2 className="size-3.5 text-emerald-600" />}
                                        {attachment.status === "error" && <CircleX className="size-3.5" />}
                                        <span className="truncate">{formatBytes(attachment.file.size)} · {statusLabel(attachment, t)}</span>
                                    </p>
                                </div>
                                {attachment.status === "ready" && (
                                    <Button size="icon-xs" variant="ghost" disabled={disabled} aria-label={t("assignments.submission.removeFile", {name: attachment.file.name})} onClick={() => onChange(attachments.filter(({id}) => id !== attachment.id))}>
                                        <X />
                                    </Button>
                                )}
                            </div>
                            {attachment.status === "uploading" && <Progress className="mt-2" value={attachment.progress} />}
                        </div>
                    ))}
                </div>
            )}

            {!attachments.length && (
                <p className="flex items-center gap-2 text-xs text-muted-foreground"><Paperclip className="size-3.5" /> {t("assignments.submission.optionalFiles")}</p>
            )}
        </Field>
    )
}

function Summary({label, value}: {label: string; value: string}) {
    return <div className="rounded-xl bg-white/60 p-3 dark:bg-black/15"><p className="text-xs opacity-70">{label}</p><p className="mt-1 font-medium">{value}</p></div>
}

function SubmissionUploadProgress({attachments}: {attachments: PendingAttachment[]}) {
    const {t} = useLocale()
    const completed = attachments.filter(({status}) => status === "done").length
    const allUploaded = completed === attachments.length
    const overallProgress = Math.round(attachments.reduce((total, attachment) => {
        if (attachment.status === "done" || attachment.status === "verifying") return total + 100
        if (attachment.status === "uploading") return total + attachment.progress
        return total
    }, 0) / attachments.length)

    return (
        <div className="space-y-4 rounded-2xl border bg-muted/15 p-4" aria-live="polite" aria-busy={!allUploaded}>
            <div className="flex items-start gap-3">
                <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
                    {allUploaded ? <CheckCircle2 className="size-5" /> : <Upload className="size-5" />}
                </span>
                <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center justify-between gap-2">
                        <p className="font-semibold">{t(allUploaded ? "assignments.submission.finalizing" : "assignments.submission.uploadingFiles")}</p>
                        <span className="text-sm font-medium tabular-nums text-muted-foreground">{overallProgress}%</span>
                    </div>
                    <p className="mt-1 text-sm text-muted-foreground">
                        {allUploaded
                            ? t("assignments.submission.finalizingDescription")
                            : t("assignments.submission.filesUploadedProgress", {completed, total: attachments.length})}
                    </p>
                    <Progress className="mt-3" value={overallProgress} />
                </div>
            </div>

            <div className="max-h-44 space-y-2 overflow-y-auto pr-1">
                {attachments.map((attachment) => (
                    <div className="flex min-w-0 items-center gap-3 rounded-xl border bg-background/80 p-3" key={attachment.id}>
                        <File className="size-4 shrink-0 text-muted-foreground" />
                        <div className="min-w-0 flex-1">
                            <p className="truncate text-sm font-medium">{attachment.file.name}</p>
                            <p className={`mt-0.5 flex items-center gap-1.5 text-xs ${attachment.status === "error" ? "text-destructive" : "text-muted-foreground"}`}>
                                {["preparing", "uploading", "verifying"].includes(attachment.status) && <LoaderCircle className="size-3.5 animate-spin" />}
                                {attachment.status === "done" && <CheckCircle2 className="size-3.5 text-emerald-600" />}
                                {attachment.status === "error" && <CircleX className="size-3.5" />}
                                <span className="truncate">{formatBytes(attachment.file.size)} · {statusLabel(attachment, t)}</span>
                            </p>
                        </div>
                    </div>
                ))}
            </div>

            <p className="flex items-center gap-2 text-xs font-medium text-muted-foreground">
                <LoaderCircle className="size-3.5 animate-spin" /> {t("assignments.submission.keepOpen")}
            </p>
        </div>
    )
}

function statusLabel(attachment: PendingAttachment, t: ReturnType<typeof useLocale>["t"]) {
    if (attachment.status === "preparing") return t("assignments.submission.preparing")
    if (attachment.status === "uploading") return t("assignments.submission.uploading", {count: attachment.progress})
    if (attachment.status === "verifying") return t("assignments.submission.verifying")
    if (attachment.status === "done") return t("assignments.submission.uploaded")
    if (attachment.status === "error") return attachment.error ?? t("assignments.submission.uploadFailed")
    return t("assignments.submission.ready")
}

function errorMessage(code: string, t: ReturnType<typeof useLocale>["t"]) {
    if (code === "network") return t("assignments.submission.error.network")
    if (code === "unauthorized") return t("assignments.submission.error.unauthorized")
    if (code === "forbidden") return t("assignments.submission.error.forbidden")
    if (code === "closed") return t("assignments.submission.error.closed")
    if (code === "not-found") return t("assignments.submission.error.notFound")
    if (code === "invalid") return t("assignments.submission.error.invalid")
    return t("assignments.submission.error.server")
}

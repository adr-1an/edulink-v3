"use client"

import {useId} from "react"
import {CheckCircle2, CircleX, File, LoaderCircle, Paperclip, Upload, X} from "lucide-react"
import {toast} from "sonner"
import {Button} from "@/components/ui/button"
import {Field, FieldDescription, FieldLabel} from "@/components/ui/field"
import {Progress} from "@/components/ui/progress"

export type AttachmentUploadStatus = "ready" | "preparing" | "uploading" | "finalizing" | "done" | "error"

export interface PendingPostAttachment {
    id: string
    file: File
    status: AttachmentUploadStatus
    progress: number
    error?: string
}

const allowedTypes = new Set([
    "image/jpeg",
    "image/png",
    "image/gif",
    "image/webp",
    "application/zip",
    "application/pdf",
])

const acceptedFiles = ".jpg,.jpeg,.png,.gif,.webp,.zip,.pdf"
const maxFileSize = 5 * 1024 * 1024

function formatBytes(bytes: number) {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function statusLabel(attachment: PendingPostAttachment) {
    switch (attachment.status) {
    case "preparing":
        return "Preparing"
    case "uploading":
        return `Uploading ${attachment.progress}%`
    case "finalizing":
        return "Verifying"
    case "done":
        return "Uploaded"
    case "error":
        return attachment.error ?? "Upload failed"
    default:
        return "Ready"
    }
}

export default function PostAttachmentPicker({
    attachments,
    disabled,
    label = "Attachments",
    onChange,
}: {
    attachments: PendingPostAttachment[]
    disabled: boolean
    label?: string
    onChange: (attachments: PendingPostAttachment[]) => void
}) {
    const inputID = useId()

    function addFiles(files: FileList | null) {
        if (!files) return

        const existing = new Set(attachments.map(({file}) => `${file.name}:${file.size}:${file.lastModified}`))
        const accepted: PendingPostAttachment[] = []
        const rejected: string[] = []

        for (const file of Array.from(files)) {
            const key = `${file.name}:${file.size}:${file.lastModified}`
            if (existing.has(key)) continue
            if (!file.name || file.name.length > 255) {
                rejected.push(`${file.name || "Unnamed file"} has an invalid file name.`)
                continue
            }
            if (file.size <= 0 || file.size > maxFileSize) {
                rejected.push(`${file.name} can't be larger than 5 MB.`)
                continue
            }
            if (!allowedTypes.has(file.type)) {
                rejected.push(`${file.name} is not a supported file type.`)
                continue
            }

            existing.add(key)
            accepted.push({
                id: crypto.randomUUID(),
                file,
                status: "ready",
                progress: 0,
            })
        }

        if (rejected.length > 0) {
            toast.error(rejected[0], {
                description: rejected.length > 1 ? `${rejected.length - 1} more files were rejected.` : undefined,
            })
        }
        if (accepted.length > 0) onChange([...attachments, ...accepted])
    }

    return (
        <Field>
            <div className="flex items-center justify-between gap-3">
                <FieldLabel htmlFor={inputID}>{label}</FieldLabel>
                {attachments.length > 0 && <span className="text-xs tabular-nums text-muted-foreground">{attachments.length} selected</span>}
            </div>

            <label
                className={`flex cursor-pointer items-center justify-center gap-2 rounded-xl border border-dashed px-4 py-4 text-sm font-medium transition-colors ${
                    disabled ? "pointer-events-none opacity-60" : "hover:border-primary/40 hover:bg-primary/5"
                }`}
                htmlFor={inputID}
            >
                <Upload className="size-4 text-muted-foreground" />
                Add files
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
            <FieldDescription>Optional. JPEG, PNG, GIF, WebP, ZIP, or PDF files up to 5 MB each.</FieldDescription>

            {attachments.length > 0 && (
                <div className="max-h-48 space-y-2 overflow-y-auto pr-1" aria-label="Selected attachments">
                    {attachments.map((attachment) => (
                        <div className="rounded-xl border bg-muted/20 p-3" key={attachment.id}>
                            <div className="flex min-w-0 items-center gap-3">
                                <span className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-background text-muted-foreground">
                                    <File className="size-4" />
                                </span>
                                <div className="min-w-0 flex-1">
                                    <p className="truncate text-sm font-medium">{attachment.file.name}</p>
                                    <p className={`mt-0.5 flex items-center gap-1.5 truncate text-xs ${
                                        attachment.status === "error" ? "text-destructive" : "text-muted-foreground"
                                    }`}>
                                        {attachment.status === "done" && <CheckCircle2 className="size-3.5 shrink-0" />}
                                        {attachment.status === "error" && <CircleX className="size-3.5 shrink-0" />}
                                        {["preparing", "uploading", "finalizing"].includes(attachment.status) && <LoaderCircle className="size-3.5 shrink-0 animate-spin" />}
                                        <span className="truncate">{formatBytes(attachment.file.size)} · {statusLabel(attachment)}</span>
                                    </p>
                                </div>
                                {attachment.status === "ready" && (
                                    <Button
                                        size="icon-xs"
                                        variant="ghost"
                                        disabled={disabled}
                                        aria-label={`Remove ${attachment.file.name}`}
                                        onClick={() => onChange(attachments.filter((item) => item.id !== attachment.id))}
                                    >
                                        <X />
                                    </Button>
                                )}
                            </div>
                            {attachment.status === "uploading" && (
                                <Progress className="mt-2" value={attachment.progress} aria-label={`Uploading ${attachment.file.name}`} />
                            )}
                        </div>
                    ))}
                </div>
            )}

            {attachments.length === 0 && (
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                    <Paperclip className="size-3.5" /> The post can be created without attachments.
                </div>
            )}
        </Field>
    )
}

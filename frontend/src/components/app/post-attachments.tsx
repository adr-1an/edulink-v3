"use client"

import {useState} from "react"
import {Download, ExternalLink, Eye, FileArchive, FileText, ImageIcon, ImageOff, LoaderCircle, Paperclip, X} from "lucide-react"
import {toast} from "sonner"
import {Button} from "@/components/ui/button"
import {
    Dialog, DialogClose, DialogDescription, DialogFooter, DialogHeader,
    DialogPanel, DialogPopup, DialogTitle, DialogTrigger,
} from "@/components/ui/dialog"
import {ScrollArea} from "@/components/ui/scroll-area"
import {type PostAttachment} from "@/lib/post_attachments"
import {cn} from "@/lib/utils"
import {useLocale} from "@/i18n/provider"
import {pluralCategory} from "@/i18n/config"

const previewableImages = new Set(["image/jpeg", "image/png", "image/gif", "image/webp"])

function safeDownloadName(value: string) {
    return value.replace(/[\\/\0]/g, "_").trim() || "attachment"
}

function saveBlob(blob: Blob, fileName: string) {
    const url = URL.createObjectURL(blob)
    const anchor = document.createElement("a")
    anchor.href = url
    anchor.download = safeDownloadName(fileName)
    anchor.hidden = true
    document.body.append(anchor)
    anchor.click()
    anchor.remove()
    window.setTimeout(() => URL.revokeObjectURL(url), 1_000)
}

function isImage(attachment: PostAttachment) {
    return previewableImages.has(attachment.contentType)
}

function isPdf(attachment: PostAttachment) {
    return attachment.contentType === "application/pdf"
}

function AttachmentIcon({contentType}: {contentType: string}) {
    if (contentType === "application/zip") return <FileArchive />
    if (contentType.startsWith("image/")) return <ImageIcon />
    return <FileText />
}

function contentTypeLabel(contentType: string, t: ReturnType<typeof useLocale>["t"]) {
    if (contentType === "application/pdf") return t("attachments.pdfDocument")
    if (contentType === "application/zip") return t("attachments.zipArchive")
    if (contentType.startsWith("image/")) return t("attachments.image", {type: contentType.slice("image/".length).toUpperCase()})
    return contentType
}

function ImagePreview({attachment, className}: {
    attachment: PostAttachment
    className: string
}) {
    const [failed, setFailed] = useState(false)
    const {t} = useLocale()

    if (failed) {
        return (
            <span className={cn("flex items-center justify-center bg-muted/30 text-muted-foreground", className)}>
                <ImageOff className="size-8" />
                <span className="sr-only">{t("attachments.previewUnavailable", {name: attachment.fileName})}</span>
            </span>
        )
    }

    return (
        // Presigned URLs must be requested directly by the browser rather than Next's image optimizer.
        // eslint-disable-next-line @next/next/no-img-element
        <img
            className={className}
            src={attachment.presignedUrl}
            alt={t("attachments.previewAlt", {name: attachment.fileName})}
            loading="lazy"
            onError={() => setFailed(true)}
        />
    )
}

export default function PostAttachments({
    attachments,
    postTitle,
    viewerTitle,
    viewerDescription,
    onRemove,
    removingAttachmentId = null,
    removeDisabled = false,
}: {
    attachments: PostAttachment[]
    postTitle: string
    viewerTitle?: string
    viewerDescription?: string
    onRemove?: (attachment: PostAttachment) => void
    removingAttachmentId?: string | null
    removeDisabled?: boolean
}) {
    const [viewerOpen, setViewerOpen] = useState(false)
    const [pdfPreview, setPdfPreview] = useState<PostAttachment | null>(null)
    const [downloadingAttachmentIDs, setDownloadingAttachmentIDs] = useState<string[]>([])
    const [downloadAllProgress, setDownloadAllProgress] = useState<{completed: number; total: number} | null>(null)
    const {locale, t} = useLocale()

    if (attachments.length === 0) return null

    const visibleAttachments = attachments.slice(0, 3)
    const attachmentPlural = pluralCategory(locale, attachments.length)
    const attachmentCountKey = attachmentPlural === "one"
        ? "attachments.one"
        : attachmentPlural === "few" ? "attachments.few" : "attachments.other"
    const fileCountKey = attachmentPlural === "one"
        ? "attachments.filesOne"
        : attachmentPlural === "few" ? "attachments.filesFew" : "attachments.filesOther"
    const previewPdf = (attachment: PostAttachment) => {
        setViewerOpen(false)
        setPdfPreview(attachment)
    }

    async function saveAttachment(attachment: PostAttachment) {
        setDownloadingAttachmentIDs((ids) => ids.includes(attachment.id) ? ids : [...ids, attachment.id])
        try {
            const response = await fetch(attachment.presignedUrl, {
                cache: "no-store",
                credentials: "omit",
                referrerPolicy: "no-referrer",
            })
            if (!response.ok) return false
            saveBlob(await response.blob(), attachment.fileName)
            return true
        } catch {
            return false
        } finally {
            setDownloadingAttachmentIDs((ids) => ids.filter((id) => id !== attachment.id))
        }
    }

    async function downloadAttachment(attachment: PostAttachment) {
        if (downloadAllProgress || downloadingAttachmentIDs.includes(attachment.id)) return
        if (!await saveAttachment(attachment)) {
            toast.error(t("attachments.downloadFailed", {name: attachment.fileName}))
        }
    }

    async function downloadAll() {
        if (downloadAllProgress || downloadingAttachmentIDs.length > 0) return
        const files = [...attachments]
        let failed = 0
        setDownloadAllProgress({completed: 0, total: files.length})

        for (let index = 0; index < files.length; index += 1) {
            if (!await saveAttachment(files[index])) failed += 1
            setDownloadAllProgress({completed: index + 1, total: files.length})
        }

        setDownloadAllProgress(null)
        if (failed > 0) {
            toast.error(t("attachments.downloadAllPartial", {failed, total: files.length}))
        } else {
            toast.success(t("attachments.downloadAllComplete", {count: files.length}))
        }
    }

    return (
        <div className="mt-4 border-t pt-4">
            <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
                <div className="flex items-center gap-2 text-xs font-medium text-muted-foreground">
                    <Paperclip className="size-3.5" />
                    {t(attachmentCountKey, {count: attachments.length})}
                </div>

                <div className="grid w-full grid-cols-2 gap-1 sm:flex sm:w-auto sm:items-center">
                    <Button
                        className="min-w-0"
                        size="xs"
                        variant="ghost"
                        disabled={downloadAllProgress !== null || downloadingAttachmentIDs.length > 0}
                        onClick={() => void downloadAll()}
                    >
                        {downloadAllProgress ? <LoaderCircle className="animate-spin" /> : <Download />}
                        {downloadAllProgress
                            ? t("attachments.downloadingAll", downloadAllProgress)
                            : t("attachments.downloadAll")}
                    </Button>
                    <Dialog open={viewerOpen} onOpenChange={setViewerOpen}>
                        <DialogTrigger render={<Button className="min-w-0" size="xs" variant="ghost" />}>{t("attachments.viewAll")}</DialogTrigger>
                        <DialogPopup className="sm:max-w-3xl">
                        <DialogHeader>
                            <DialogTitle>{viewerTitle ?? t("attachments.title")}</DialogTitle>
                            <DialogDescription>
                                {viewerDescription ?? t(fileCountKey, {
                                    count: attachments.length,
                                    title: postTitle,
                                })}
                            </DialogDescription>
                        </DialogHeader>
                        <DialogPanel>
                            <ScrollArea className="max-h-[65vh]" scrollbarGutter>
                                <div className="grid gap-3 pr-1 sm:grid-cols-2">
                                    {attachments.map((attachment) => (
                                        <div className="overflow-hidden rounded-2xl border bg-card" key={attachment.id}>
                                            {isImage(attachment) ? (
                                                <a
                                                    className="block bg-muted/30"
                                                    href={attachment.presignedUrl}
                                                    target="_blank"
                                                    rel="noreferrer"
                                                    aria-label={t("attachments.open", {name: attachment.fileName})}
                                                >
                                                    <ImagePreview attachment={attachment} className="aspect-video w-full object-contain" />
                                                </a>
                                            ) : isPdf(attachment) ? (
                                                <button
                                                    className="flex aspect-video w-full flex-col items-center justify-center gap-2 bg-muted/30 text-muted-foreground transition-colors hover:bg-muted/50 hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring"
                                                    type="button"
                                                    onClick={() => previewPdf(attachment)}
                                                >
                                                    <FileText className="size-10" />
                                                    <span className="flex items-center gap-1.5 text-xs font-medium"><Eye className="size-3.5" /> {t("attachments.previewPdf")}</span>
                                                </button>
                                            ) : (
                                                <div className="flex aspect-video items-center justify-center bg-muted/30 text-muted-foreground">
                                                    <span className="[&_svg]:size-10"><AttachmentIcon contentType={attachment.contentType} /></span>
                                                </div>
                                            )}
                                            <div className="flex min-w-0 items-center gap-3 p-3">
                                                <span className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-muted text-muted-foreground">
                                                    <AttachmentIcon contentType={attachment.contentType} />
                                                </span>
                                                <div className="min-w-0 flex-1">
                                                    <p className="truncate text-sm font-medium" title={attachment.fileName}>{attachment.fileName}</p>
                                                    <p className="truncate text-xs text-muted-foreground">{contentTypeLabel(attachment.contentType, t)}</p>
                                                </div>
                                                <Button
                                                    size="icon-sm"
                                                    variant="ghost"
                                                    disabled={downloadAllProgress !== null || downloadingAttachmentIDs.includes(attachment.id)}
                                                    aria-label={t("attachments.download", {name: attachment.fileName})}
                                                    onClick={() => void downloadAttachment(attachment)}
                                                >
                                                    {downloadingAttachmentIDs.includes(attachment.id)
                                                        ? <LoaderCircle className="animate-spin" />
                                                        : <Download />}
                                                </Button>
                                                <Button
                                                    size="icon-sm"
                                                    variant="ghost"
                                                    aria-label={t("attachments.open", {name: attachment.fileName})}
                                                    render={<a href={attachment.presignedUrl} target="_blank" rel="noreferrer" />}
                                                >
                                                    <ExternalLink />
                                                </Button>
                                                {onRemove && (
                                                    <Button
                                                        size="icon-sm"
                                                        variant="ghost"
                                                        disabled={removeDisabled || downloadAllProgress !== null || downloadingAttachmentIDs.includes(attachment.id)}
                                                        aria-label={t("attachments.remove", {name: attachment.fileName})}
                                                        onClick={() => onRemove(attachment)}
                                                    >
                                                        {removingAttachmentId === attachment.id
                                                            ? <LoaderCircle className="animate-spin" />
                                                            : <X />}
                                                    </Button>
                                                )}
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </ScrollArea>
                        </DialogPanel>
                        <DialogFooter>
                            <DialogClose render={<Button variant="outline" />}>{t("attachments.close")}</DialogClose>
                        </DialogFooter>
                        </DialogPopup>
                    </Dialog>
                </div>
            </div>

            <div className="grid gap-2 sm:grid-cols-3">
                {visibleAttachments.map((attachment) => {
                    const content = (
                        <>
                            {isImage(attachment) && (
                                <ImagePreview
                                    attachment={attachment}
                                    className="aspect-video w-full border-b object-cover transition-transform duration-200 group-hover:scale-[1.02]"
                                />
                            )}
                            <span className={cn(
                                "flex min-w-0 items-center gap-2 p-2.5",
                                !isImage(attachment) && (onRemove ? "pr-20" : "pr-12"),
                            )}>
                                <span className="text-muted-foreground [&_svg]:size-4">
                                    <AttachmentIcon contentType={attachment.contentType} />
                                </span>
                                <span className="min-w-0 flex-1 truncate text-xs font-medium" title={attachment.fileName}>
                                    {attachment.fileName}
                                </span>
                                {isPdf(attachment)
                                    ? <Eye className="size-3.5 shrink-0 text-muted-foreground" />
                                    : <ExternalLink className="size-3.5 shrink-0 text-muted-foreground" />}
                            </span>
                        </>
                    )

                    return (
                        <div
                            className="group relative min-w-0 overflow-hidden rounded-xl border bg-muted/20 transition-colors hover:border-primary/30 hover:bg-muted/40"
                            key={attachment.id}
                        >
                            {isPdf(attachment) ? (
                                <button
                                    className="block w-full text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring"
                                    type="button"
                                    onClick={() => previewPdf(attachment)}
                                >
                                    {content}
                                </button>
                            ) : (
                                <a
                                    className="block focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring"
                                    href={attachment.presignedUrl}
                                    target="_blank"
                                    rel="noreferrer"
                                >
                                    {content}
                                </a>
                            )}
                            <div className="absolute right-2 top-2 flex items-center gap-1">
                                <Button
                                    className="bg-background/90 shadow-sm backdrop-blur-sm hover:bg-background"
                                    size="icon-xs"
                                    variant="outline"
                                    disabled={downloadAllProgress !== null || downloadingAttachmentIDs.includes(attachment.id)}
                                    aria-label={t("attachments.download", {name: attachment.fileName})}
                                    onClick={() => void downloadAttachment(attachment)}
                                >
                                    {downloadingAttachmentIDs.includes(attachment.id)
                                        ? <LoaderCircle className="animate-spin" />
                                        : <Download />}
                                </Button>
                                {onRemove && (
                                    <Button
                                        className="bg-background/90 shadow-sm backdrop-blur-sm hover:bg-background"
                                        size="icon-xs"
                                        variant="outline"
                                        disabled={removeDisabled || downloadAllProgress !== null || downloadingAttachmentIDs.includes(attachment.id)}
                                        aria-label={t("attachments.remove", {name: attachment.fileName})}
                                        onClick={() => onRemove(attachment)}
                                    >
                                        {removingAttachmentId === attachment.id
                                            ? <LoaderCircle className="animate-spin" />
                                            : <X />}
                                    </Button>
                                )}
                            </div>
                        </div>
                    )
                })}
            </div>

            {attachments.length > visibleAttachments.length && (
                <p className="mt-2 text-xs text-muted-foreground">
                    {t("attachments.more", {count: attachments.length - visibleAttachments.length})}
                </p>
            )}

            <Dialog open={pdfPreview !== null} onOpenChange={(open) => {
                if (!open) setPdfPreview(null)
            }}>
                <DialogPopup className="sm:max-w-5xl">
                    <DialogHeader>
                        <DialogTitle className="pr-8">{pdfPreview?.fileName ?? t("attachments.pdfPreview")}</DialogTitle>
                        <DialogDescription>
                            {t("attachments.pdfDescription")}
                        </DialogDescription>
                    </DialogHeader>
                    <DialogPanel className="p-0">
                        {pdfPreview && (
                            <iframe
                                className="h-[65vh] min-h-96 w-full border-y bg-muted/20"
                                src={pdfPreview.presignedUrl}
                                title={t("attachments.previewAlt", {name: pdfPreview.fileName})}
                                // WebKit does not render PDFs in sandboxed iframes.
                                loading="lazy"
                                referrerPolicy="no-referrer"
                            />
                        )}
                    </DialogPanel>
                    <DialogFooter>
                        <DialogClose render={<Button variant="outline" />}>{t("attachments.close")}</DialogClose>
                        {pdfPreview && (
                            <Button
                                variant="outline"
                                disabled={downloadAllProgress !== null || downloadingAttachmentIDs.includes(pdfPreview.id)}
                                onClick={() => void downloadAttachment(pdfPreview)}
                            >
                                {downloadingAttachmentIDs.includes(pdfPreview.id)
                                    ? <LoaderCircle className="animate-spin" />
                                    : <Download />}
                                {t("attachments.downloadFile")}
                            </Button>
                        )}
                        {pdfPreview && (
                            <Button render={<a href={pdfPreview.presignedUrl} target="_blank" rel="noreferrer" />}>
                                <ExternalLink /> {t("attachments.openSeparately")}
                            </Button>
                        )}
                    </DialogFooter>
                </DialogPopup>
            </Dialog>
        </div>
    )
}

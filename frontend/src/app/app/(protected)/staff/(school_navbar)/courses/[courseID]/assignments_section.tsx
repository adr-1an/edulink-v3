"use client"

import {useMemo, useState} from "react"
import Link from "next/link"
import {useRouter} from "next/navigation"
import {
    CalendarClock, ClipboardList, Clock3, ExternalLink, FileText, Link2, LockKeyhole,
    Pencil, Plus, RotateCcw, Send, Trash2, TriangleAlert,
} from "lucide-react"
import {toast} from "sonner"
import LocalDateTime from "@/components/local-date-time"
import {
    AlertDialog, AlertDialogClose, AlertDialogDescription, AlertDialogFooter,
    AlertDialogHeader, AlertDialogPopup, AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import {Badge} from "@/components/ui/badge"
import {Button} from "@/components/ui/button"
import {Card, CardContent, CardHeader, CardTitle} from "@/components/ui/card"
import {
    Dialog, DialogClose, DialogDescription, DialogFooter, DialogHeader,
    DialogPanel, DialogPopup, DialogTitle,
} from "@/components/ui/dialog"
import {Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle} from "@/components/ui/empty"
import {Field, FieldDescription, FieldLabel} from "@/components/ui/field"
import {Form} from "@/components/ui/form"
import {Input} from "@/components/ui/input"
import {Select, SelectItem, SelectPopup, SelectTrigger, SelectValue} from "@/components/ui/select"
import {Switch} from "@/components/ui/switch"
import {Textarea} from "@/components/ui/textarea"
import {hasSchoolPermission, type SchoolAccess} from "@/lib/school_access"
import {useLocale} from "@/i18n/provider"
import {
    handleCreateAssignment, handleDeleteAssignment, handleUpdateAssignment,
    type CourseAssignmentInput,
} from "./actions"

export interface CourseAssignment {
    id: string
    referencedPost: {
        id: string
        title: string
    } | null
    title: string
    description: string
    dueDate: string | null
    submissionsEnabled: boolean
    submissionsCloseAt: string | null
    createdAt: string
}

interface ReferencePost {
    id: string
    title: string
    body: string
    authorName: string
    accentColor: string
    createdAtLabel: string
}

interface AssignmentDraft {
    hasReferencedPost: boolean
    referencedPostId: string
    title: string
    description: string
    hasDueDate: boolean
    dueDate: string
    submissionsEnabled: boolean
    hasSubmissionsCloseAt: boolean
    submissionsCloseAt: string
}

const emptyDraft: AssignmentDraft = {
    hasReferencedPost: false,
    referencedPostId: "",
    title: "",
    description: "",
    hasDueDate: false,
    dueDate: "",
    submissionsEnabled: true,
    hasSubmissionsCloseAt: false,
    submissionsCloseAt: "",
}

function normalizeColor(value: string) {
    return /^[0-9A-Fa-f]{6}$/.test(value) ? `#${value}` : "#6366F1"
}

function toISOString(value: string) {
    if (!value) return null
    const date = new Date(value)
    return Number.isNaN(date.getTime()) ? null : date.toISOString()
}

function toDateTimeLocal(value: string | null) {
    if (!value) return ""
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return ""
    const pad = (part: number) => String(part).padStart(2, "0")
    return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function assignmentInput(draft: AssignmentDraft): CourseAssignmentInput | null {
    const dueDate = draft.hasDueDate ? toISOString(draft.dueDate) : null
    const submissionsCloseAt = draft.submissionsEnabled && draft.hasSubmissionsCloseAt
        ? toISOString(draft.submissionsCloseAt)
        : null

    if (draft.hasDueDate && !dueDate) return null
    if (draft.submissionsEnabled && draft.hasSubmissionsCloseAt && !submissionsCloseAt) return null

    return {
        referencedPostId: draft.hasReferencedPost ? draft.referencedPostId : null,
        title: draft.title,
        description: draft.description,
        dueDate,
        submissionsEnabled: draft.submissionsEnabled,
        submissionsCloseAt,
    }
}

export default function AssignmentsSection({
    courseID,
    initialAssignments,
    referencePosts,
    access,
    denied = false,
    error,
}: {
    courseID: string
    initialAssignments: CourseAssignment[]
    referencePosts: ReferencePost[]
    access: SchoolAccess
    denied?: boolean
    error?: string
}) {
    const router = useRouter()
    const {t} = useLocale()
    const [createOpen, setCreateOpen] = useState(false)
    const [editTarget, setEditTarget] = useState<CourseAssignment | null>(null)
    const [deleteTarget, setDeleteTarget] = useState<CourseAssignment | null>(null)
    const [deleteConfirmation, setDeleteConfirmation] = useState("")
    const [saving, setSaving] = useState(false)
    const [deleting, setDeleting] = useState(false)
    const [draft, setDraft] = useState<AssignmentDraft>(emptyDraft)
    const canCreate = hasSchoolPermission(access, "course.assignment.create") && !denied && !error
    const canUpdate = hasSchoolPermission(access, "course.assignment.update") && !denied && !error
    const canDelete = hasSchoolPermission(access, "course.assignment.delete") && !denied && !error
    const canViewPosts = hasSchoolPermission(access, "course.post.view")
    const canListSubmissions = hasSchoolPermission(access, "submission.list")
    const sortedAssignments = useMemo(() => [...initialAssignments].sort((first, second) => {
        if (first.dueDate && second.dueDate) {
            return new Date(first.dueDate).getTime() - new Date(second.dueDate).getTime()
        }
        if (first.dueDate) return -1
        if (second.dueDate) return 1
        return new Date(second.createdAt).getTime() - new Date(first.createdAt).getTime()
    }), [initialAssignments])

    function openCreate() {
        setDraft(emptyDraft)
        setCreateOpen(true)
    }

    function openEdit(assignment: CourseAssignment) {
        setDraft({
            hasReferencedPost: assignment.referencedPost !== null,
            referencedPostId: assignment.referencedPost?.id ?? "",
            title: assignment.title,
            description: assignment.description,
            hasDueDate: assignment.dueDate !== null,
            dueDate: toDateTimeLocal(assignment.dueDate),
            submissionsEnabled: assignment.submissionsEnabled,
            hasSubmissionsCloseAt: assignment.submissionsCloseAt !== null,
            submissionsCloseAt: toDateTimeLocal(assignment.submissionsCloseAt),
        })
        setEditTarget(assignment)
    }

    function openDelete(assignment: CourseAssignment) {
        setDeleteConfirmation("")
        setDeleteTarget(assignment)
    }

    async function createAssignment(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        const input = assignmentInput(draft)
        if (!input) return toast.error("Enter valid assignment dates.")

        setSaving(true)
        const result = await handleCreateAssignment(courseID, input)
        setSaving(false)
        if (!result.ok) return toast.error(result.message)

        setCreateOpen(false)
        toast.success("Assignment created.")
        router.refresh()
    }

    async function updateAssignment(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        if (!editTarget) return
        const input = assignmentInput(draft)
        if (!input) return toast.error("Enter valid assignment dates.")

        setSaving(true)
        const result = await handleUpdateAssignment(editTarget.id, input)
        setSaving(false)
        if (!result.ok) return toast.error(result.message)

        setEditTarget(null)
        toast.success("Assignment updated.")
        router.refresh()
    }

    async function deleteAssignment() {
        if (!deleteTarget || deleteConfirmation !== deleteTarget.title) return

        setDeleting(true)
        const result = await handleDeleteAssignment(deleteTarget.id)
        setDeleting(false)
        if (!result.ok) return toast.error(result.message)

        setDeleteTarget(null)
        setDeleteConfirmation("")
        toast.success("Assignment permanently deleted.")
        router.refresh()
    }

    return (
        <section className="space-y-4" aria-labelledby="course-assignments-title">
            <div className="flex flex-wrap items-end justify-between gap-3">
                <div>
                    <h2 className="text-xl font-semibold" id="course-assignments-title">Assignments</h2>
                    <p className="text-sm text-muted-foreground">Coursework, deadlines, and submission availability</p>
                </div>
                {canCreate && (
                    <Button onClick={openCreate}><Plus /> Create assignment</Button>
                )}
            </div>

            {denied ? (
                <Card>
                    <Empty>
                        <EmptyHeader>
                            <EmptyMedia variant="icon"><LockKeyhole /></EmptyMedia>
                            <EmptyTitle>Assignments aren&apos;t available</EmptyTitle>
                            <EmptyDescription>You don&apos;t have permission to view assignments in this course.</EmptyDescription>
                        </EmptyHeader>
                    </Empty>
                </Card>
            ) : error ? (
                <Card>
                    <Empty>
                        <EmptyHeader>
                            <EmptyMedia variant="icon"><ClipboardList /></EmptyMedia>
                            <EmptyTitle>Couldn&apos;t load assignments</EmptyTitle>
                            <EmptyDescription>{error}</EmptyDescription>
                        </EmptyHeader>
                        <Button variant="outline" onClick={() => router.refresh()}><RotateCcw /> Try again</Button>
                    </Empty>
                </Card>
            ) : sortedAssignments.length === 0 ? (
                <Card>
                    <Empty>
                        <EmptyHeader>
                            <EmptyMedia variant="icon"><ClipboardList /></EmptyMedia>
                            <EmptyTitle>No assignments yet</EmptyTitle>
                            <EmptyDescription>
                                {canCreate
                                    ? "Create the first assignment for this course."
                                    : "Assignments for this course will appear here."}
                            </EmptyDescription>
                        </EmptyHeader>
                        {canCreate && <Button onClick={openCreate}><Plus /> Create assignment</Button>}
                    </Empty>
                </Card>
            ) : (
                <div className="grid gap-3 lg:grid-cols-2">
                    {sortedAssignments.map((assignment) => (
                        <AssignmentCard
                            assignment={assignment}
                            canUpdate={canUpdate}
                            canDelete={canDelete}
                            canViewPost={canViewPosts}
                            canListSubmissions={canListSubmissions}
                            submissionsLabel={t("staff.submissions.openList")}
                            onEdit={() => openEdit(assignment)}
                            onDelete={() => openDelete(assignment)}
                            key={assignment.id}
                        />
                    ))}
                </div>
            )}

            <Dialog open={createOpen} onOpenChange={(open) => {
                if (!saving) setCreateOpen(open)
            }}>
                <DialogPopup className="sm:max-w-2xl">
                    <DialogHeader>
                        <DialogTitle>Create assignment</DialogTitle>
                        <DialogDescription>Add coursework to this course, with an optional connection to an existing post.</DialogDescription>
                    </DialogHeader>
                    <AssignmentForm draft={draft} posts={referencePosts} saving={saving} submitLabel="Create assignment" canViewPosts={canViewPosts} onChange={setDraft} onSubmit={createAssignment} />
                </DialogPopup>
            </Dialog>

            <Dialog open={editTarget !== null} onOpenChange={(open) => {
                if (!open && !saving) setEditTarget(null)
            }}>
                <DialogPopup className="sm:max-w-2xl">
                    <DialogHeader>
                        <DialogTitle>Edit assignment</DialogTitle>
                        <DialogDescription>Update the assignment details, deadlines, and submission settings.</DialogDescription>
                    </DialogHeader>
                    <AssignmentForm draft={draft} posts={referencePosts} saving={saving} submitLabel="Save changes" canViewPosts={canViewPosts} onChange={setDraft} onSubmit={updateAssignment} />
                </DialogPopup>
            </Dialog>

            <AlertDialog open={deleteTarget !== null} onOpenChange={(open) => {
                if (!open && !deleting) {
                    setDeleteTarget(null)
                    setDeleteConfirmation("")
                }
            }}>
                <AlertDialogPopup className="border-destructive/30 sm:max-w-lg">
                    <AlertDialogHeader>
                        <div className="mb-2 flex size-12 items-center justify-center rounded-2xl bg-destructive/10 text-destructive">
                            <TriangleAlert className="size-6" />
                        </div>
                        <AlertDialogTitle className="text-destructive">Permanently delete assignment?</AlertDialogTitle>
                        <AlertDialogDescription className="space-y-3">
                            <span className="block">
                                Deleting <strong className="font-semibold text-foreground">&ldquo;{deleteTarget?.title}&rdquo;</strong> also permanently deletes all student submissions and other records connected to this assignment.
                            </span>
                            <span className="block font-medium text-destructive">This cannot be undone or recovered.</span>
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <Field className="px-6 pb-2">
                        <FieldLabel htmlFor="assignment-delete-confirmation">Type the assignment title to confirm</FieldLabel>
                        <Input
                            id="assignment-delete-confirmation"
                            value={deleteConfirmation}
                            onChange={(event) => setDeleteConfirmation(event.target.value)}
                            placeholder={deleteTarget?.title}
                            disabled={deleting}
                            autoComplete="off"
                        />
                    </Field>
                    <AlertDialogFooter>
                        <AlertDialogClose render={<Button variant="outline" disabled={deleting} />}>Cancel</AlertDialogClose>
                        <Button
                            variant="destructive"
                            loading={deleting}
                            disabled={!deleteTarget || deleteConfirmation !== deleteTarget.title || deleting}
                            onClick={deleteAssignment}
                        >
                            Permanently delete
                        </Button>
                    </AlertDialogFooter>
                </AlertDialogPopup>
            </AlertDialog>
        </section>
    )
}

function AssignmentCard({assignment, canUpdate, canDelete, canViewPost, canListSubmissions, submissionsLabel, onEdit, onDelete}: {
    assignment: CourseAssignment
    canUpdate: boolean
    canDelete: boolean
    canViewPost: boolean
    canListSubmissions: boolean
    submissionsLabel: string
    onEdit: () => void
    onDelete: () => void
}) {
    return (
        <Card className="overflow-hidden">
            <CardHeader className="gap-3">
                <div className="flex flex-wrap items-start justify-between gap-3">
                    <CardTitle className="wrap-break-word text-lg leading-snug">{assignment.title}</CardTitle>
                    <div className="flex items-center gap-1">
                        {canUpdate && <Button size="icon-sm" variant="ghost" aria-label={`Edit ${assignment.title}`} onClick={onEdit}><Pencil /></Button>}
                        {canDelete && <Button size="icon-sm" variant="destructive-outline" aria-label={`Delete ${assignment.title}`} onClick={onDelete}><Trash2 /></Button>}
                    </div>
                </div>
                <div className="flex flex-wrap gap-2">
                    <Badge variant="secondary">
                        <CalendarClock />
                        {assignment.dueDate ? <>Due <LocalDateTime value={assignment.dueDate} /></> : "No due date"}
                    </Badge>
                    <Badge variant={assignment.submissionsEnabled ? "success" : "outline"}>
                        <Send /> Submissions {assignment.submissionsEnabled ? "enabled" : "disabled"}
                    </Badge>
                    {assignment.submissionsCloseAt && (
                        <Badge variant="outline"><Clock3 /> Close <LocalDateTime value={assignment.submissionsCloseAt} /></Badge>
                    )}
                </div>
            </CardHeader>
            <CardContent className="space-y-4">
                {assignment.description
                    ? <p className="whitespace-pre-wrap wrap-break-word text-sm leading-relaxed">{assignment.description}</p>
                    : <p className="text-sm italic text-muted-foreground">No description provided.</p>}
                <div className="flex items-center gap-2 rounded-xl border bg-muted/25 px-3 py-2.5 text-sm">
                    {assignment.referencedPost ? (
                        <>
                            <Link2 className="size-4 shrink-0 text-muted-foreground" />
                            <span className="min-w-0 truncate"><span className="text-muted-foreground">Referenced post:</span> {assignment.referencedPost.title}</span>
                            {canViewPost && (
                                <Button className="ml-auto shrink-0" size="icon-xs" variant="ghost" aria-label={`Open ${assignment.referencedPost.title}`} render={<Link href={`/app/staff/posts/${assignment.referencedPost.id}`} />}>
                                    <ExternalLink />
                                </Button>
                            )}
                        </>
                    ) : (
                        <>
                            <FileText className="size-4 shrink-0 text-muted-foreground" />
                            <span className="text-muted-foreground">No course post is referenced.</span>
                        </>
                    )}
                </div>
                <div className="flex flex-wrap items-center justify-between gap-3">
                    <p className="text-xs text-muted-foreground">Created <LocalDateTime value={assignment.createdAt} /></p>
                    {canListSubmissions && (
                        <Button
                            size="sm"
                            variant="outline"
                            render={<Link href={`/app/staff/assignments/${assignment.id}/submissions`} />}
                        >
                            <ClipboardList /> {submissionsLabel}
                        </Button>
                    )}
                </div>
            </CardContent>
        </Card>
    )
}

function AssignmentForm({draft, posts, saving, submitLabel, canViewPosts, onChange, onSubmit}: {
    draft: AssignmentDraft
    posts: ReferencePost[]
    saving: boolean
    submitLabel: string
    canViewPosts: boolean
    onChange: (draft: AssignmentDraft) => void
    onSubmit: (event: React.SubmitEvent<HTMLFormElement>) => void
}) {
    const selectedPost = draft.hasReferencedPost
        ? posts.find((post) => post.id === draft.referencedPostId)
        : undefined
    const dueDate = draft.hasDueDate ? toISOString(draft.dueDate) : null
    const closeAt = draft.submissionsEnabled && draft.hasSubmissionsCloseAt ? toISOString(draft.submissionsCloseAt) : null
    const valid = (!draft.hasReferencedPost || /^\d+$/.test(draft.referencedPostId))
        && draft.title.trim().length >= 3
        && draft.title.trim().length <= 64
        && draft.description.trim().length <= 4096
        && (!draft.hasDueDate || dueDate !== null)
        && (!draft.submissionsEnabled || !draft.hasSubmissionsCloseAt || closeAt !== null)
        && (!dueDate || !closeAt || new Date(closeAt).getTime() >= new Date(dueDate).getTime())

    return (
        <Form className="contents" onSubmit={onSubmit}>
            <DialogPanel className="grid gap-4 sm:grid-cols-2">
                <Field className="sm:col-span-2">
                    <div className="flex items-center justify-between gap-4 rounded-xl border bg-muted/25 p-4">
                        <div>
                            <FieldLabel>Reference a course post</FieldLabel>
                            <FieldDescription className="mt-1">Optionally connect this assignment to an existing post.</FieldDescription>
                        </div>
                        <Switch
                            checked={draft.hasReferencedPost}
                            onCheckedChange={(hasReferencedPost) => onChange({
                                ...draft,
                                hasReferencedPost,
                                referencedPostId: hasReferencedPost ? draft.referencedPostId || posts[0]?.id || "" : "",
                            })}
                            disabled={posts.length === 0 || saving}
                            aria-label="Reference a course post"
                        />
                    </div>
                    {posts.length === 0 && <FieldDescription>No course posts are available to reference.</FieldDescription>}
                </Field>

                <Field className="sm:col-span-2">
                    <FieldLabel>Referenced post</FieldLabel>
                    <Select
                        items={posts.map((post) => ({label: `${post.title} — ${post.createdAtLabel}`, value: post.id}))}
                        value={draft.referencedPostId}
                        onValueChange={(referencedPostId) => onChange({...draft, referencedPostId: referencedPostId ?? ""})}
                        disabled={!draft.hasReferencedPost || posts.length === 0 || saving}
                    >
                        <SelectTrigger aria-label="Referenced course post"><SelectValue placeholder="Choose a post" /></SelectTrigger>
                        <SelectPopup>
                            {posts.map((post) => (
                                <SelectItem value={post.id} key={post.id}>
                                    <span className="flex min-w-0 flex-col">
                                        <span className="truncate">{post.title}</span>
                                        <span className="truncate text-xs text-muted-foreground">{post.authorName} · {post.createdAtLabel}</span>
                                    </span>
                                </SelectItem>
                            ))}
                        </SelectPopup>
                    </Select>
                    <FieldDescription>{draft.hasReferencedPost ? "This can point students to fuller instructions or an announcement." : "Enable post referencing above to choose a post."}</FieldDescription>
                    {selectedPost && (
                        <div className="relative mt-1 overflow-hidden rounded-xl border bg-muted/20 p-4 pl-5">
                            <span className="absolute inset-y-0 left-0 w-1" style={{backgroundColor: normalizeColor(selectedPost.accentColor)}} />
                            <div className="flex items-start gap-3">
                                <span className="mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-lg bg-background text-muted-foreground shadow-xs">
                                    <FileText className="size-4" />
                                </span>
                                <div className="min-w-0">
                                    <p className="truncate text-sm font-medium">{selectedPost.title}</p>
                                    <p className="mt-0.5 truncate text-xs text-muted-foreground">{selectedPost.authorName} · Posted {selectedPost.createdAtLabel}</p>
                                    <p className="mt-2 line-clamp-3 whitespace-pre-wrap wrap-break-word text-sm leading-relaxed">{selectedPost.body}</p>
                                    {canViewPosts && (
                                        <Button className="mt-2 h-auto px-0" size="sm" variant="link" render={<Link href={`/app/staff/posts/${selectedPost.id}`} target="_blank" rel="noreferrer" />}>
                                            <ExternalLink /> Open full post
                                        </Button>
                                    )}
                                </div>
                            </div>
                        </div>
                    )}
                </Field>

                <Field className="sm:col-span-2">
                    <FieldLabel>Title</FieldLabel>
                    <Input value={draft.title} onChange={(event) => onChange({...draft, title: event.target.value})} minLength={3} maxLength={64} required disabled={saving} autoFocus />
                    <FieldDescription>{draft.title.length}/64 characters</FieldDescription>
                </Field>

                <Field className="sm:col-span-2">
                    <FieldLabel>Description</FieldLabel>
                    <Textarea className="min-h-28 resize-y" value={draft.description} onChange={(event) => onChange({...draft, description: event.target.value})} maxLength={4096} disabled={saving} placeholder="Optional summary or additional instructions..." />
                    <FieldDescription>{draft.description.length}/4096 characters</FieldDescription>
                </Field>

                <Field className="sm:col-span-2">
                    <div className="flex items-center justify-between gap-4 rounded-xl border bg-muted/25 p-4">
                        <div>
                            <FieldLabel>Due date</FieldLabel>
                            <FieldDescription className="mt-1">Set a deadline for this assignment.</FieldDescription>
                        </div>
                        <Switch checked={draft.hasDueDate} onCheckedChange={(hasDueDate) => onChange({...draft, hasDueDate, dueDate: hasDueDate ? draft.dueDate : ""})} disabled={saving} aria-label="Set assignment due date" />
                    </div>
                </Field>
                <Field className="sm:col-span-2">
                    <FieldLabel>Due date and time</FieldLabel>
                    <Input type="datetime-local" value={draft.dueDate} onChange={(event) => onChange({...draft, dueDate: event.target.value})} required={draft.hasDueDate} disabled={!draft.hasDueDate || saving} />
                </Field>

                <Field className="sm:col-span-2">
                    <div className="flex items-center justify-between gap-4 rounded-xl border bg-muted/25 p-4">
                        <div>
                            <FieldLabel>Student submissions</FieldLabel>
                            <FieldDescription className="mt-1">Allow students to submit work for this assignment.</FieldDescription>
                        </div>
                        <Switch checked={draft.submissionsEnabled} onCheckedChange={(submissionsEnabled) => onChange({
                            ...draft,
                            submissionsEnabled,
                            hasSubmissionsCloseAt: submissionsEnabled ? draft.hasSubmissionsCloseAt : false,
                            submissionsCloseAt: submissionsEnabled ? draft.submissionsCloseAt : "",
                        })} disabled={saving} aria-label="Enable student submissions" />
                    </div>
                </Field>

                <Field className="sm:col-span-2">
                    <div className="flex items-center justify-between gap-4 rounded-xl border bg-muted/25 p-4">
                        <div>
                            <FieldLabel>Submissions closing time</FieldLabel>
                            <FieldDescription className="mt-1">Optionally stop accepting submissions at a specific time.</FieldDescription>
                        </div>
                        <Switch checked={draft.hasSubmissionsCloseAt} onCheckedChange={(hasSubmissionsCloseAt) => onChange({
                            ...draft,
                            hasSubmissionsCloseAt,
                            submissionsCloseAt: hasSubmissionsCloseAt ? draft.submissionsCloseAt : "",
                        })} disabled={!draft.submissionsEnabled || saving} aria-label="Set submissions closing time" />
                    </div>
                </Field>
                <Field className="sm:col-span-2">
                    <FieldLabel>Submissions close at</FieldLabel>
                    <Input type="datetime-local" value={draft.submissionsCloseAt} onChange={(event) => onChange({...draft, submissionsCloseAt: event.target.value})} required={draft.submissionsEnabled && draft.hasSubmissionsCloseAt} disabled={!draft.submissionsEnabled || !draft.hasSubmissionsCloseAt || saving} />
                    {dueDate && closeAt && new Date(closeAt).getTime() < new Date(dueDate).getTime() && (
                        <FieldDescription className="text-destructive">Submissions cannot close before the due date.</FieldDescription>
                    )}
                </Field>
            </DialogPanel>
            <DialogFooter>
                <DialogClose render={<Button variant="outline" disabled={saving} />}>Cancel</DialogClose>
                <Button type="submit" loading={saving} disabled={!valid || saving}>{submitLabel}</Button>
            </DialogFooter>
        </Form>
    )
}

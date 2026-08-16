"use client"

import {useEffect, useId, useMemo, useState} from "react"
import Link from "next/link"
import {useRouter} from "next/navigation"
import {
    BookOpen, CalendarClock, ChevronLeft, ClipboardList, ExternalLink, FileText, LockKeyhole, Pencil, Plus,
    Paperclip, RotateCcw, Search, Trash2, TriangleAlert, UserMinus, UserPlus, Users, X,
} from "lucide-react"
import {toast} from "sonner"
import PostAttachments from "@/components/app/post-attachments"
import PageTitle from "@/components/app/page_title"
import UserAvatar from "@/components/app/user_avatar"
import {useLocalDateTimeFormatter} from "@/components/local-date-time"
import {
    AlertDialog, AlertDialogClose, AlertDialogDescription, AlertDialogFooter,
    AlertDialogHeader, AlertDialogPopup, AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import {Avatar, AvatarFallback} from "@/components/ui/avatar"
import {Badge} from "@/components/ui/badge"
import {Button} from "@/components/ui/button"
import {Card, CardContent, CardFooter, CardHeader, CardTitle} from "@/components/ui/card"
import {
    Dialog, DialogClose, DialogDescription, DialogFooter, DialogHeader,
    DialogPanel, DialogPopup, DialogTitle,
} from "@/components/ui/dialog"
import {Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle} from "@/components/ui/empty"
import {Field, FieldDescription, FieldLabel} from "@/components/ui/field"
import {Form} from "@/components/ui/form"
import {Input} from "@/components/ui/input"
import {ScrollArea} from "@/components/ui/scroll-area"
import {Switch} from "@/components/ui/switch"
import {Tabs, TabsList, TabsPanel, TabsTab} from "@/components/ui/tabs"
import {Textarea} from "@/components/ui/textarea"
import {type PostAttachment} from "@/lib/post_attachments"
import {hasSchoolPermission, type SchoolAccess} from "@/lib/school_access"
import {rememberCurrentSchoolAccess} from "@/lib/school_navigation"
import {uploadToPresignedURL} from "@/lib/upload_to_presigned_url"
import {useLocale} from "@/i18n/provider"
import AssignmentsSection, {type CourseAssignment} from "./assignments_section"
import {
    handleAssignCourseStudent, handleCompletePostAttachmentUpload, handleCreatePost, handleDeletePost,
    handleDeletePostAttachment, handleInitPostAttachmentUpload, handleRemoveCourseStudent, handleUpdatePost,
    type CoursePostInput,
} from "./actions"
import PostAttachmentPicker, {type PendingPostAttachment} from "./post_attachment_picker"

export interface CoursePost {
    id: string
    attachments: PostAttachment[]
    authorName: string
    authorProfilePictureURL: string | null
    title: string
    body: string
    accentColor: string
    showUntil: string | null
    editedAt: string | null
    createdAt: string
}

export interface CourseStudent {
    id: string
    name: string
    lastName: string
    email: string
    assigned: boolean
}

interface PostDraft {
    title: string
    body: string
    accentColor: string
    hasExpiry: boolean
    showUntil: string
}

const emptyDraft: PostDraft = {title: "", body: "", accentColor: "6366F1", hasExpiry: false, showUntil: ""}

function normalizeColor(color: string) {
    return /^[0-9A-Fa-f]{6}$/.test(color) ? color.toUpperCase() : "6366F1"
}

function initials(name: string) {
    return name
        .trim()
        .split(/\s+/)
        .slice(0, 2)
        .map((part) => part[0]?.toUpperCase())
        .join("") || "?"
}

function toDateTimeLocal(value: string | null) {
    if (!value) return ""
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return ""
    const pad = (part: number) => String(part).padStart(2, "0")
    return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function draftToInput(draft: PostDraft): CoursePostInput | null {
    const input = {title: draft.title, body: draft.body, accentColor: draft.accentColor}
    if (!draft.hasExpiry) return {...input, showUntil: null}
    if (!draft.showUntil) return null
    const date = new Date(draft.showUntil)
    if (Number.isNaN(date.getTime())) return null
    return {...input, showUntil: date.toISOString()}
}

export default function CourseDashboardClientPage({
    courseID,
    initialPosts,
    initialStudents,
    initialAssignments,
    access,
    postsDenied = false,
    postsError,
    studentsDenied = false,
    studentsError,
    assignmentsDenied = false,
    assignmentsError,
}: {
    courseID: string
    initialPosts: CoursePost[]
    initialStudents: CourseStudent[]
    initialAssignments: CourseAssignment[]
    access: SchoolAccess
    postsDenied?: boolean
    postsError?: string
    studentsDenied?: boolean
    studentsError?: string
    assignmentsDenied?: boolean
    assignmentsError?: string
}) {
    const router = useRouter()
    const {t} = useLocale()
    const formatDate = useLocalDateTimeFormatter()
    const [posts, setPosts] = useState(initialPosts)
    const [createOpen, setCreateOpen] = useState(false)
    const [editTarget, setEditTarget] = useState<CoursePost | null>(null)
    const [deleteTarget, setDeleteTarget] = useState<CoursePost | null>(null)
    const [postDeleteConfirmation, setPostDeleteConfirmation] = useState("")
    const [draft, setDraft] = useState<PostDraft>(emptyDraft)
    const [attachments, setAttachments] = useState<PendingPostAttachment[]>([])
    const [attachmentDeleteTarget, setAttachmentDeleteTarget] = useState<PostAttachment | null>(null)
    const [deletingAttachment, setDeletingAttachment] = useState(false)
    const [saving, setSaving] = useState(false)
    const [deleting, setDeleting] = useState(false)
    const [students, setStudents] = useState(initialStudents)
    const [rosterOpen, setRosterOpen] = useState(false)
    const [studentQuery, setStudentQuery] = useState("")
    const [studentBusy, setStudentBusy] = useState<string | null>(null)
    const canCreate = hasSchoolPermission(access, "course.post.create")
    const canView = hasSchoolPermission(access, "course.post.view")
    const canUpdate = hasSchoolPermission(access, "course.post.update")
    const canDelete = hasSchoolPermission(access, "course.post.delete")
    const canUploadAttachments = hasSchoolPermission(access, "post.attachment.create")
    const canDeleteAttachments = hasSchoolPermission(access, "post.attachment.delete")
    const canAssignStudents = hasSchoolPermission(access, "course.student.assign")
    const canRemoveStudents = hasSchoolPermission(access, "course.student.remove")
    const sortedPosts = useMemo(() => [...posts].sort((first, second) => {
        return new Date(second.createdAt).getTime() - new Date(first.createdAt).getTime()
    }), [posts])
    const timedPostCount = posts.filter((post) => post.showUntil !== null).length
    const alwaysVisiblePostCount = posts.length - timedPostCount
    const canCreatePost = canCreate && !postsDenied && !postsError
    const assignedStudents = useMemo(() => students.filter((student) => student.assigned), [students])
    const availableStudents = useMemo(() => students.filter((student) => !student.assigned), [students])
    const normalizedStudentQuery = studentQuery.trim().toLocaleLowerCase()
    const matchesStudentQuery = (student: CourseStudent) => !normalizedStudentQuery
        || `${student.name} ${student.lastName} ${student.email}`.toLocaleLowerCase().includes(normalizedStudentQuery)
    const visibleAssignedStudents = assignedStudents.filter(matchesStudentQuery)
    const visibleAvailableStudents = availableStudents.filter(matchesStudentQuery)

    useEffect(() => {
        if ((postsDenied || postsError) && (studentsDenied || studentsError) && (assignmentsDenied || assignmentsError)) return
        rememberCurrentSchoolAccess(access)
    }, [access, assignmentsDenied, assignmentsError, postsDenied, postsError, studentsDenied, studentsError])

    function openCreate() {
        setDraft(emptyDraft)
        setAttachments([])
        setCreateOpen(true)
    }

    function openEdit(post: CoursePost) {
        setDraft({
            title: post.title,
            body: post.body,
            accentColor: normalizeColor(post.accentColor),
            hasExpiry: post.showUntil !== null,
            showUntil: toDateTimeLocal(post.showUntil),
        })
        setAttachments([])
        setEditTarget(post)
    }

    async function uploadSelectedAttachments(postID: string) {
        let uploaded = 0
        let failed = 0

        for (const attachment of attachments) {
            setAttachments((current) => current.map((item) => item.id === attachment.id
                ? {...item, status: "preparing", progress: 0, error: undefined}
                : item))

            const initialized = await handleInitPostAttachmentUpload(postID, {
                name: attachment.file.name,
                size: attachment.file.size,
                type: attachment.file.type,
            })
            if (!initialized.ok) {
                failed++
                setAttachments((current) => current.map((item) => item.id === attachment.id
                    ? {...item, status: "error", error: initialized.message}
                    : item))
                continue
            }

            setAttachments((current) => current.map((item) => item.id === attachment.id
                ? {...item, status: "uploading", progress: 0}
                : item))
            try {
                await uploadToPresignedURL(attachment.file, initialized.upload.url, (progress) => {
                    setAttachments((current) => current.map((item) => item.id === attachment.id
                        ? {...item, progress}
                        : item))
                })
            } catch (error) {
                failed++
                const message = error instanceof Error ? error.message : "The file upload failed."
                setAttachments((current) => current.map((item) => item.id === attachment.id
                    ? {...item, status: "error", error: message}
                    : item))
                continue
            }

            setAttachments((current) => current.map((item) => item.id === attachment.id
                ? {...item, status: "finalizing", progress: 100}
                : item))
            const completed = await handleCompletePostAttachmentUpload(
                initialized.upload.id,
                initialized.upload.completionToken,
            )
            if (!completed.ok) {
                failed++
                setAttachments((current) => current.map((item) => item.id === attachment.id
                    ? {...item, status: "error", error: completed.message}
                    : item))
                continue
            }

            uploaded++
            setAttachments((current) => current.map((item) => item.id === attachment.id
                ? {...item, status: "done", progress: 100}
                : item))
        }

        return {uploaded, failed}
    }

    async function createPost(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        const input = draftToInput(draft)
        if (!input) return toast.error("Enter a valid expiry date.")

        setSaving(true)
        const res = await handleCreatePost(courseID, input)
        if (!res.ok) {
            setSaving(false)
            return toast.error(res.message)
        }
        if (!res.id) {
            setSaving(false)
            return toast.error("The post was created, but uploads couldn't start because its ID was missing.")
        }

        const {uploaded, failed} = await uploadSelectedAttachments(res.id)

        setSaving(false)
        setCreateOpen(false)
        if (failed > 0) {
            toast.error("Post created, but some files failed to upload.", {
                description: `${uploaded} uploaded · ${failed} failed`,
            })
        } else if (uploaded > 0) {
            toast.success(`Post created with ${uploaded} ${uploaded === 1 ? "attachment" : "attachments"}.`)
        } else {
            toast.success("Post created.")
        }
        router.refresh()
    }

    async function updatePost(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        if (!editTarget) return
        const input = draftToInput(draft)
        if (!input) return toast.error("Enter a valid expiry date.")

        setSaving(true)
        const res = await handleUpdatePost(editTarget.id, input)
        if (!res.ok) {
            setSaving(false)
            return toast.error(res.message)
        }

        const {uploaded, failed} = await uploadSelectedAttachments(editTarget.id)
        setSaving(false)

        const updatedPost: CoursePost = {
            ...editTarget,
            title: input.title.trim(),
            body: input.body.trim(),
            accentColor: normalizeColor(input.accentColor),
            showUntil: input.showUntil,
            editedAt: new Date().toISOString(),
        }
        setPosts((current) => current.map((post) => post.id === editTarget.id ? updatedPost : post))
        setEditTarget(null)
        setAttachments([])
        if (failed > 0) {
            toast.error("Post updated, but some files failed to upload.", {
                description: `${uploaded} uploaded · ${failed} failed`,
            })
        } else if (uploaded > 0) {
            toast.success(`Post updated with ${uploaded} new ${uploaded === 1 ? "attachment" : "attachments"}.`)
        } else {
            toast.success("Post updated.")
        }
        router.refresh()
    }

    async function deletePost() {
        if (!deleteTarget) return
        if (postDeleteConfirmation !== deleteTarget.title) return
        setDeleting(true)
        const res = await handleDeletePost(deleteTarget.id)
        setDeleting(false)
        if (!res.ok) return toast.error(res.message)

        setPosts((current) => current.filter((post) => post.id !== deleteTarget.id))
        setDeleteTarget(null)
        setPostDeleteConfirmation("")
        toast.success("Post deleted.")
    }

    async function deleteAttachment() {
        if (!editTarget || !attachmentDeleteTarget || deletingAttachment) return

        setDeletingAttachment(true)
        const result = await handleDeletePostAttachment(editTarget.id, attachmentDeleteTarget.id)
        setDeletingAttachment(false)
        if (!result.ok) {
            toast.error(result.message)
            router.refresh()
            return
        }

        const attachmentID = attachmentDeleteTarget.id
        setPosts((current) => current.map((post) => post.id === editTarget.id
            ? {...post, attachments: post.attachments.filter((attachment) => attachment.id !== attachmentID)}
            : post))
        setEditTarget((current) => current
            ? {...current, attachments: current.attachments.filter((attachment) => attachment.id !== attachmentID)}
            : null)
        setAttachmentDeleteTarget(null)
        toast.success("Attachment deleted.")
        router.refresh()
    }

    async function updateStudentAssignment(student: CourseStudent, assigned: boolean) {
        setStudentBusy(student.id)
        const res = assigned
            ? await handleAssignCourseStudent(courseID, student.id)
            : await handleRemoveCourseStudent(courseID, student.id)
        setStudentBusy(null)
        if (!res.ok) return toast.error(res.message)

        setStudents((current) => current.map((item) => item.id === student.id ? {...item, assigned} : item))
        toast.success(assigned
            ? `${student.name} ${student.lastName} was added to the course.`
            : `${student.name} ${student.lastName} was removed from the course.`)
    }

    return (
        <div className="space-y-8">
            <header className="relative overflow-hidden rounded-3xl border bg-card p-5 shadow-xs sm:p-7">
                <div className="pointer-events-none absolute -right-16 -top-20 size-64 rounded-full bg-primary/8 blur-3xl" />
                <div className="relative">
                    <Button className="-ml-2 mb-5" size="sm" variant="ghost" onClick={() => router.back()}>
                        <ChevronLeft /> Back to courses
                    </Button>
                    <div className="flex flex-col justify-between gap-5 sm:flex-row sm:items-start">
                        <div className="flex items-start gap-3">
                            <span className="mt-0.5 flex size-11 shrink-0 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-sm">
                                <BookOpen className="size-5" />
                            </span>
                            <div className="min-w-0">
                                <PageTitle centered={false}>Course dashboard</PageTitle>
                                <p className="mt-1 max-w-2xl text-muted-foreground">Updates, announcements, and everything happening in this course.</p>
                            </div>
                        </div>
                    </div>
                    {((!postsDenied && !postsError && posts.length > 0)
                        || (!assignmentsDenied && !assignmentsError && initialAssignments.length > 0)) && (
                        <div className="mt-6 flex flex-wrap gap-x-6 gap-y-2 border-t pt-4 text-sm text-muted-foreground">
                            {!postsDenied && !postsError && (
                                <>
                                    <span className="flex items-center gap-2"><FileText className="size-4" /> {posts.length} {posts.length === 1 ? "post" : "posts"}</span>
                                    <span className="flex items-center gap-2"><CalendarClock className="size-4" /> {timedPostCount} time-limited</span>
                                    {sortedPosts[0] && <span className="flex items-center gap-2">Latest update {formatDate(sortedPosts[0].createdAt)}</span>}
                                </>
                            )}
                            {!assignmentsDenied && !assignmentsError && (
                                <span className="flex items-center gap-2"><ClipboardList className="size-4" /> {initialAssignments.length} {initialAssignments.length === 1 ? "assignment" : "assignments"}</span>
                            )}
                        </div>
                    )}
                </div>
            </header>

            <Tabs defaultValue="posts">
                <TabsList aria-label="Course dashboard sections">
                    <TabsTab value="posts"><FileText /> Posts {!postsDenied && !postsError && <Badge variant="secondary">{posts.length}</Badge>}</TabsTab>
                    <TabsTab value="assignments"><ClipboardList /> Assignments {!assignmentsDenied && !assignmentsError && <Badge variant="secondary">{initialAssignments.length}</Badge>}</TabsTab>
                </TabsList>

                <TabsPanel className="pt-4" value="posts">
            <div className="grid gap-6 xl:grid-cols-[minmax(0,2fr)_minmax(17rem,1fr)]">
                <section className="min-w-0 space-y-4" aria-labelledby="course-posts-title">
                    <div className="flex flex-wrap items-end justify-between gap-3">
                        <div>
                            <h2 className="text-xl font-semibold" id="course-posts-title">Course feed</h2>
                            <p className="text-sm text-muted-foreground">The latest updates from your course</p>
                        </div>
                        {canCreatePost && <Button onClick={openCreate}><Plus /> Create post</Button>}
                    </div>

                    {postsDenied ? (
                        <Card>
                            <Empty>
                                <EmptyHeader>
                                    <EmptyMedia variant="icon"><LockKeyhole /></EmptyMedia>
                                    <EmptyTitle>Posts aren&apos;t available</EmptyTitle>
                                    <EmptyDescription>You don&apos;t have permission to view posts in this course.</EmptyDescription>
                                </EmptyHeader>
                            </Empty>
                        </Card>
                    ) : postsError ? (
                        <Card>
                            <Empty>
                                <EmptyHeader>
                                    <EmptyMedia variant="icon"><FileText /></EmptyMedia>
                                    <EmptyTitle>Couldn&apos;t load posts</EmptyTitle>
                                    <EmptyDescription>{postsError}</EmptyDescription>
                                </EmptyHeader>
                                <Button variant="outline" onClick={() => router.refresh()}><RotateCcw /> Try again</Button>
                            </Empty>
                        </Card>
                    ) : sortedPosts.length === 0 ? (
                        <Card>
                            <Empty>
                                <EmptyHeader>
                                    <EmptyMedia variant="icon"><FileText /></EmptyMedia>
                                    <EmptyTitle>No posts yet</EmptyTitle>
                                    <EmptyDescription>{canCreate ? "Create the first update for this course." : "Updates for this course will appear here."}</EmptyDescription>
                                </EmptyHeader>
                                {canCreate && <Button onClick={openCreate}><Plus /> Create post</Button>}
                            </Empty>
                        </Card>
                    ) : (
                        <div className="space-y-3">
                            {sortedPosts.map((post) => (
                                <Card className="group overflow-hidden transition-[box-shadow,transform] duration-200 hover:-translate-y-0.5 hover:shadow-md" key={post.id}>
                                    <div className="absolute inset-y-0 left-0 w-1.5" style={{backgroundColor: `#${normalizeColor(post.accentColor)}`}} />
                                    <CardHeader className="gap-3 pl-7">
                                        <div className="flex flex-wrap items-start justify-between gap-3">
                                            <CardTitle className="wrap-break-word text-lg leading-snug">{post.title}</CardTitle>
                                            <div className="flex items-center gap-2">
                                                {canView && <Button size="icon-sm" variant="ghost" aria-label={`Open ${post.title}`} render={<Link href={`/app/staff/posts/${post.id}`} />}><ExternalLink /></Button>}
                                                {canUpdate && <Button size="icon-sm" variant="ghost" aria-label={`Edit ${post.title}`} onClick={() => openEdit(post)}><Pencil /></Button>}
                                                {canDelete && <Button size="icon-sm" variant="destructive-outline" aria-label={`Delete ${post.title}`} onClick={() => setDeleteTarget(post)}><Trash2 /></Button>}
                                            </div>
                                        </div>
                                        <div className="flex items-center gap-2.5">
                                            <UserAvatar
                                                name={post.authorName}
                                                src={post.authorProfilePictureURL}
                                                cacheKey={`course-post:${post.id}:author`}
                                                className="size-8 border"
                                            />
                                            <div className="min-w-0 text-xs text-muted-foreground">
                                                <p className="truncate font-medium text-foreground">{post.authorName}</p>
                                                <p>Posted {formatDate(post.createdAt)}</p>
                                            </div>
                                            {post.editedAt && <Badge className="ml-1" variant="secondary">Edited</Badge>}
                                        </div>
                                    </CardHeader>
                                    <CardContent className="pl-7">
                                        <p className="whitespace-pre-wrap wrap-break-word text-sm leading-relaxed">{post.body}</p>
                                        <PostAttachments attachments={post.attachments} postTitle={post.title} />
                                    </CardContent>
                                    {post.showUntil && (
                                        <CardFooter className="gap-2 border-t bg-muted/25 py-3 pl-7 text-xs text-muted-foreground">
                                            <CalendarClock className="size-3.5" /> Shows until {formatDate(post.showUntil)}
                                        </CardFooter>
                                    )}
                                </Card>
                            ))}
                        </div>
                    )}
                </section>

                {(!postsDenied && !postsError || !studentsDenied) && (
                    <aside className="space-y-4 xl:sticky xl:top-6 xl:self-start" aria-label="Course overview">
                        {!studentsDenied && (
                            <Card>
                                <CardHeader className="gap-1">
                                    <div className="flex items-center justify-between gap-3">
                                        <CardTitle className="flex items-center gap-2 text-base"><Users className="size-4" /> Students</CardTitle>
                                        {!studentsError && <Badge variant="secondary">{assignedStudents.length}</Badge>}
                                    </div>
                                    <p className="text-xs text-muted-foreground">Students currently assigned to this course</p>
                                </CardHeader>
                                <CardContent className="space-y-3">
                                    {studentsError ? (
                                        <div className="rounded-xl border border-dashed p-3 text-sm text-muted-foreground">
                                            <p>{studentsError}</p>
                                            <Button className="mt-2 px-0" size="sm" variant="link" onClick={() => router.refresh()}><RotateCcw /> Try again</Button>
                                        </div>
                                    ) : (
                                        <>
                                            {assignedStudents.length === 0 ? (
                                                <p className="rounded-xl border border-dashed p-3 text-sm text-muted-foreground">No students are assigned yet.</p>
                                            ) : (
                                                <div className="space-y-1">
                                                    {assignedStudents.slice(0, 4).map((student) => <CompactStudent key={student.id} student={student} />)}
                                                    {assignedStudents.length > 4 && <p className="px-2 pt-1 text-xs text-muted-foreground">+{assignedStudents.length - 4} more students</p>}
                                                </div>
                                            )}
                                            <Button className="w-full" variant="outline" onClick={() => setRosterOpen(true)}>
                                                {canAssignStudents || canRemoveStudents ? "Manage roster" : "View roster"}
                                            </Button>
                                        </>
                                    )}
                                </CardContent>
                            </Card>
                        )}

                        {!postsDenied && !postsError && <Card>
                            <CardHeader>
                                <CardTitle className="text-base">At a glance</CardTitle>
                            </CardHeader>
                            <CardContent className="grid grid-cols-2 gap-3 xl:grid-cols-1">
                                <div className="rounded-xl border bg-muted/30 p-3">
                                    <p className="text-2xl font-semibold tabular-nums">{alwaysVisiblePostCount}</p>
                                    <p className="mt-0.5 text-xs text-muted-foreground">Always-visible {alwaysVisiblePostCount === 1 ? "post" : "posts"}</p>
                                </div>
                                <div className="rounded-xl border bg-muted/30 p-3">
                                    <p className="text-2xl font-semibold tabular-nums">{timedPostCount}</p>
                                    <p className="mt-0.5 text-xs text-muted-foreground">Time-limited {timedPostCount === 1 ? "post" : "posts"}</p>
                                </div>
                            </CardContent>
                        </Card>}

                        {!postsDenied && !postsError && sortedPosts.length > 0 && (
                            <Card>
                                <CardHeader className="gap-1">
                                    <CardTitle className="text-base">Recent activity</CardTitle>
                                    <p className="text-xs text-muted-foreground">Newest posts in this course</p>
                                </CardHeader>
                                <CardContent className="space-y-1">
                                    {sortedPosts.slice(0, 3).map((post) => (
                                        <div className="flex gap-3 rounded-xl px-2 py-2.5" key={post.id}>
                                            <span className="mt-1.5 size-2 shrink-0 rounded-full" style={{backgroundColor: `#${normalizeColor(post.accentColor)}`}} />
                                            <div className="min-w-0">
                                                <p className="truncate text-sm font-medium">{post.title}</p>
                                                <p className="mt-0.5 text-xs text-muted-foreground">{formatDate(post.createdAt)}</p>
                                            </div>
                                        </div>
                                    ))}
                                </CardContent>
                            </Card>
                        )}
                    </aside>
                )}
            </div>
                </TabsPanel>

                <TabsPanel className="pt-4" value="assignments">
                    <AssignmentsSection
                        courseID={courseID}
                        initialAssignments={initialAssignments}
                        referencePosts={posts.map(({id, title, body, authorName, accentColor, createdAt}) => ({
                            id,
                            title,
                            body,
                            authorName,
                            accentColor,
                            createdAtLabel: formatDate(createdAt),
                        }))}
                        access={access}
                        denied={assignmentsDenied}
                        error={assignmentsError}
                    />
                </TabsPanel>
            </Tabs>

            <Dialog open={rosterOpen} onOpenChange={(open) => {
                if (studentBusy === null) setRosterOpen(open)
                if (!open) setStudentQuery("")
            }}>
                <DialogPopup className="sm:max-w-2xl">
                    <DialogHeader>
                        <DialogTitle>Course roster</DialogTitle>
                        <DialogDescription>View and manage the students who belong to this course.</DialogDescription>
                    </DialogHeader>
                    <DialogPanel className="min-h-0 space-y-4">
                        <div className="relative">
                            <Search className="pointer-events-none absolute left-3 top-1/2 z-10 size-4 -translate-y-1/2 text-muted-foreground" />
                            <Input className="pl-9 pr-9" value={studentQuery} onChange={(event) => setStudentQuery(event.target.value)} placeholder="Search by name or email..." aria-label="Search course students" />
                            {studentQuery && <Button className="absolute right-1.5 top-1/2 -translate-y-1/2" size="icon-xs" variant="ghost" aria-label="Clear search" onClick={() => setStudentQuery("")}><X /></Button>}
                        </div>
                        <Tabs defaultValue="assigned">
                            <TabsList className="w-full">
                                <TabsTab className="flex-1" value="assigned">Assigned <Badge className="ml-1" variant="secondary">{assignedStudents.length}</Badge></TabsTab>
                                <TabsTab className="flex-1" value="available">Available <Badge className="ml-1" variant="secondary">{availableStudents.length}</Badge></TabsTab>
                            </TabsList>
                            <TabsPanel value="assigned">
                                <StudentRosterList
                                    students={visibleAssignedStudents}
                                    emptyMessage={normalizedStudentQuery ? "No assigned students match your search." : "No students are assigned to this course yet."}
                                    busyID={studentBusy}
                                    action={canRemoveStudents ? "remove" : undefined}
                                    onAction={(student) => updateStudentAssignment(student, false)}
                                />
                            </TabsPanel>
                            <TabsPanel value="available">
                                <StudentRosterList
                                    students={visibleAvailableStudents}
                                    emptyMessage={normalizedStudentQuery ? "No available students match your search." : "Every student is already assigned to this course."}
                                    busyID={studentBusy}
                                    action={canAssignStudents ? "assign" : undefined}
                                    onAction={(student) => updateStudentAssignment(student, true)}
                                />
                            </TabsPanel>
                        </Tabs>
                    </DialogPanel>
                    <DialogFooter>
                        <DialogClose render={<Button variant="outline" disabled={studentBusy !== null} />}>Done</DialogClose>
                    </DialogFooter>
                </DialogPopup>
            </Dialog>

            <Dialog open={createOpen} onOpenChange={(open) => {
                if (!saving) {
                    setCreateOpen(open)
                    if (!open) setAttachments([])
                }
            }}>
                <DialogPopup className="sm:max-w-xl">
                    <DialogHeader>
                        <DialogTitle>Create post</DialogTitle>
                        <DialogDescription>Share an update with this course.</DialogDescription>
                    </DialogHeader>
                    <PostForm
                        draft={draft}
                        saving={saving}
                        submitLabel="Create post"
                        attachments={canUploadAttachments ? attachments : undefined}
                        onAttachmentsChange={canUploadAttachments ? setAttachments : undefined}
                        onChange={setDraft}
                        onSubmit={createPost}
                    />
                </DialogPopup>
            </Dialog>

            <Dialog open={editTarget !== null} onOpenChange={(open) => {
                if (!open && !saving) {
                    setEditTarget(null)
                    setAttachments([])
                }
            }}>
                <DialogPopup className="sm:max-w-xl">
                    <DialogHeader>
                        <DialogTitle>Edit post</DialogTitle>
                        <DialogDescription>Update this course post.</DialogDescription>
                    </DialogHeader>
                    <PostForm
                        draft={draft}
                        saving={saving}
                        submitLabel={attachments.length > 0 ? "Save and upload" : "Save changes"}
                        existingAttachments={editTarget?.attachments}
                        attachments={canUploadAttachments ? attachments : undefined}
                        attachmentLabel="Add attachments"
                        onDeleteAttachment={canDeleteAttachments ? setAttachmentDeleteTarget : undefined}
                        onAttachmentsChange={canUploadAttachments ? setAttachments : undefined}
                        onChange={setDraft}
                        onSubmit={updatePost}
                    />
                </DialogPopup>
            </Dialog>

            <AlertDialog open={attachmentDeleteTarget !== null} onOpenChange={(open) => {
                if (!open && !deletingAttachment) setAttachmentDeleteTarget(null)
            }}>
                <AlertDialogPopup>
                    <AlertDialogHeader>
                        <AlertDialogTitle>Delete attachment?</AlertDialogTitle>
                        <AlertDialogDescription>
                            &ldquo;{attachmentDeleteTarget?.fileName}&rdquo; will be permanently removed from this post.
                            This cannot be undone.
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogClose render={<Button variant="outline" disabled={deletingAttachment} />}>
                            Cancel
                        </AlertDialogClose>
                        <Button
                            variant="destructive"
                            loading={deletingAttachment}
                            disabled={deletingAttachment}
                            onClick={deleteAttachment}
                        >
                            Delete attachment
                        </Button>
                    </AlertDialogFooter>
                </AlertDialogPopup>
            </AlertDialog>

            <AlertDialog open={deleteTarget !== null} onOpenChange={(open) => {
                if (!open && !deleting) {
                    setDeleteTarget(null)
                    setPostDeleteConfirmation("")
                }
            }}>
                <AlertDialogPopup className="border-destructive/30 sm:max-w-lg">
                    <AlertDialogHeader>
                        <div className="mb-2 flex size-12 items-center justify-center rounded-2xl bg-destructive/10 text-destructive">
                            <TriangleAlert className="size-6" />
                        </div>
                        <AlertDialogTitle className="text-destructive">{t("staff.posts.delete.title")}</AlertDialogTitle>
                        <AlertDialogDescription className="space-y-3 text-left">
                            <span className="block">{t("staff.posts.delete.description", {title: deleteTarget?.title ?? ""})}</span>
                            <span className="block font-medium text-destructive">{t("staff.posts.delete.irreversible")}</span>
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <Field className="px-6 pb-2">
                        <FieldLabel htmlFor="post-delete-confirmation">{t("staff.posts.delete.typeTitle")}</FieldLabel>
                        <Input
                            id="post-delete-confirmation"
                            value={postDeleteConfirmation}
                            onChange={(event) => setPostDeleteConfirmation(event.target.value)}
                            placeholder={deleteTarget?.title}
                            autoComplete="off"
                            disabled={deleting}
                            autoFocus
                        />
                    </Field>
                    <AlertDialogFooter>
                        <AlertDialogClose render={<Button variant="outline" disabled={deleting} />}>{t("staff.posts.delete.cancel")}</AlertDialogClose>
                        <Button
                            variant="destructive"
                            loading={deleting}
                            disabled={!deleteTarget || postDeleteConfirmation !== deleteTarget.title || deleting}
                            onClick={deletePost}
                        >
                            {t("staff.posts.delete.confirm")}
                        </Button>
                    </AlertDialogFooter>
                </AlertDialogPopup>
            </AlertDialog>
        </div>
    )
}

function CompactStudent({student}: {student: CourseStudent}) {
    return (
        <div className="flex items-center gap-2.5 rounded-xl px-2 py-2">
            <Avatar className="size-8 border">
                <AvatarFallback className="text-xs">{initials(`${student.name} ${student.lastName}`)}</AvatarFallback>
            </Avatar>
            <div className="min-w-0">
                <p className="truncate text-sm font-medium">{student.name} {student.lastName}</p>
                {student.email && <p className="truncate text-xs text-muted-foreground">{student.email}</p>}
            </div>
        </div>
    )
}

function StudentRosterList({students, emptyMessage, busyID, action, onAction}: {
    students: CourseStudent[]
    emptyMessage: string
    busyID: string | null
    action?: "assign" | "remove"
    onAction: (student: CourseStudent) => void
}) {
    if (students.length === 0) {
        return <div className="flex h-64 items-center justify-center rounded-xl border border-dashed px-6 text-center text-sm text-muted-foreground">{emptyMessage}</div>
    }

    return (
        <ScrollArea className="h-64 pr-2" scrollbarGutter>
            <div className="space-y-1">
                {students.map((student) => (
                    <div className="flex items-center gap-3 rounded-xl border border-transparent p-2.5 transition-colors hover:border-border hover:bg-muted/35" key={student.id}>
                        <Avatar className="size-9 border">
                            <AvatarFallback className="text-xs">{initials(`${student.name} ${student.lastName}`)}</AvatarFallback>
                        </Avatar>
                        <div className="min-w-0 flex-1">
                            <p className="truncate text-sm font-medium">{student.name} {student.lastName}</p>
                            <p className="truncate text-xs text-muted-foreground">{student.email || "No email address"}</p>
                        </div>
                        {action === "assign" && (
                            <Button size="sm" loading={busyID === student.id} disabled={busyID !== null && busyID !== student.id} onClick={() => onAction(student)}><UserPlus /> Add</Button>
                        )}
                        {action === "remove" && (
                            <Button size="sm" variant="destructive-outline" loading={busyID === student.id} disabled={busyID !== null && busyID !== student.id} onClick={() => onAction(student)}><UserMinus /> Remove</Button>
                        )}
                        {!action && <Badge variant="outline">{student.assigned ? "Assigned" : "Available"}</Badge>}
                    </div>
                ))}
            </div>
        </ScrollArea>
    )
}

function PostForm({
    draft,
    saving,
    submitLabel,
    existingAttachments,
    attachments,
    attachmentLabel,
    onDeleteAttachment,
    onAttachmentsChange,
    onChange,
    onSubmit,
}: {
    draft: PostDraft
    saving: boolean
    submitLabel: string
    existingAttachments?: PostAttachment[]
    attachments?: PendingPostAttachment[]
    attachmentLabel?: string
    onDeleteAttachment?: (attachment: PostAttachment) => void
    onAttachmentsChange?: (attachments: PendingPostAttachment[]) => void
    onChange: (draft: PostDraft) => void
    onSubmit: (event: React.SubmitEvent<HTMLFormElement>) => void
}) {
    const expiryID = useId()
    const showUntilID = useId()
    const color = normalizeColor(draft.accentColor)
    const valid = Boolean(draft.title.trim())
        && draft.title.trim().length <= 32
        && Boolean(draft.body.trim())
        && draft.body.trim().length <= 2048
        && /^[0-9A-Fa-f]{6}$/.test(draft.accentColor)
        && (!draft.hasExpiry || Boolean(draft.showUntil))

    return (
        <Form className="contents" onSubmit={onSubmit}>
            <DialogPanel className="grid gap-4">
                <Field>
                    <div className="flex items-center justify-between gap-3">
                        <FieldLabel>Title</FieldLabel>
                        <span className="text-xs tabular-nums text-muted-foreground">{draft.title.length}/32</span>
                    </div>
                    <Input
                        value={draft.title}
                        onChange={(event) => onChange({...draft, title: event.target.value})}
                        placeholder="What’s happening?"
                        maxLength={32}
                        required
                        autoFocus
                    />
                </Field>
                <Field>
                    <div className="flex items-center justify-between gap-3">
                        <FieldLabel>Content</FieldLabel>
                        <span className="text-xs tabular-nums text-muted-foreground">{draft.body.length}/2048</span>
                    </div>
                    <Textarea
                        className="min-h-36 resize-y"
                        value={draft.body}
                        onChange={(event) => onChange({...draft, body: event.target.value})}
                        placeholder="Write an update for the course..."
                        maxLength={2048}
                        required
                    />
                </Field>
                {existingAttachments && existingAttachments.length > 0 && (
                    <Field>
                        <div className="flex items-center justify-between gap-3">
                            <FieldLabel>Current attachments</FieldLabel>
                            <span className="text-xs tabular-nums text-muted-foreground">
                                {existingAttachments.length} {existingAttachments.length === 1 ? "file" : "files"}
                            </span>
                        </div>
                        <div className="max-h-32 space-y-2 overflow-y-auto pr-1">
                            {existingAttachments.map((attachment) => (
                                <div
                                    className="flex min-w-0 items-center gap-1 rounded-xl border bg-muted/20 p-1"
                                    key={attachment.id}
                                >
                                    <a
                                        className="flex min-w-0 flex-1 items-center gap-2.5 rounded-lg p-1.5 text-sm transition-colors hover:bg-muted/50"
                                        href={attachment.presignedUrl}
                                        target="_blank"
                                        rel="noreferrer"
                                    >
                                        <Paperclip className="size-4 shrink-0 text-muted-foreground" />
                                        <span className="min-w-0 flex-1 truncate" title={attachment.fileName}>
                                            {attachment.fileName}
                                        </span>
                                        <ExternalLink className="size-3.5 shrink-0 text-muted-foreground" />
                                    </a>
                                    {onDeleteAttachment && (
                                        <Button
                                            type="button"
                                            size="icon-xs"
                                            variant="destructive-outline"
                                            aria-label={`Delete ${attachment.fileName}`}
                                            onClick={() => onDeleteAttachment(attachment)}
                                        >
                                            <Trash2 />
                                        </Button>
                                    )}
                                </div>
                            ))}
                        </div>
                        <FieldDescription>Existing files stay attached when you edit this post.</FieldDescription>
                    </Field>
                )}
                {attachments && onAttachmentsChange && (
                    <PostAttachmentPicker
                        attachments={attachments}
                        disabled={saving}
                        label={attachmentLabel}
                        onChange={onAttachmentsChange}
                    />
                )}
                <Field>
                    <FieldLabel>Accent color</FieldLabel>
                    <div className="flex items-center gap-2">
                        <Input
                            className="h-9 w-14 cursor-pointer p-1"
                            type="color"
                            value={`#${color}`}
                            onChange={(event) => onChange({...draft, accentColor: event.target.value.slice(1).toUpperCase()})}
                        />
                        <div className="relative flex-1">
                            <span className="pointer-events-none absolute inset-y-0 left-3 z-10 flex items-center text-muted-foreground">#</span>
                            <Input
                                className="pl-7 font-mono uppercase"
                                value={draft.accentColor}
                                onChange={(event) => onChange({...draft, accentColor: event.target.value.replace(/^#/, "").slice(0, 6)})}
                                pattern="[0-9A-Fa-f]{6}"
                                maxLength={6}
                                required
                            />
                        </div>
                    </div>
                </Field>
                <Field>
                    <div className="flex w-full items-center justify-between gap-4 rounded-lg border bg-muted/25 p-3">
                        <label className="min-w-0 flex-1 cursor-pointer" htmlFor={expiryID}>
                            <span className="block text-sm font-medium">Set expiry</span>
                            <span className="mt-0.5 block text-xs text-muted-foreground">Automatically stop showing this post after a date.</span>
                        </label>
                        <Switch
                            id={expiryID}
                            checked={draft.hasExpiry}
                            onCheckedChange={(checked) => onChange({
                                ...draft,
                                hasExpiry: checked,
                                showUntil: checked ? draft.showUntil : "",
                            })}
                            aria-label="Set an expiry date"
                        />
                    </div>
                    <FieldLabel htmlFor={showUntilID}>Show until</FieldLabel>
                    <Input
                        id={showUntilID}
                        type="datetime-local"
                        value={draft.showUntil}
                        onChange={(event) => onChange({...draft, showUntil: event.target.value})}
                        disabled={!draft.hasExpiry}
                        required={draft.hasExpiry}
                    />
                    <FieldDescription>{draft.hasExpiry ? "Choose when this post should stop showing." : "Turn on expiry to choose a date and time."}</FieldDescription>
                </Field>
            </DialogPanel>
            <DialogFooter>
                <DialogClose render={<Button variant="outline" disabled={saving} />}>Cancel</DialogClose>
                <Button type="submit" loading={saving} disabled={!valid}>{submitLabel}</Button>
            </DialogFooter>
        </Form>
    )
}

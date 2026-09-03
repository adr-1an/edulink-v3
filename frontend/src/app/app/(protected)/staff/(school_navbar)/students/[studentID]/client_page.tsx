"use client"

import {useEffect, useMemo, useRef, useState} from "react"
import Link from "next/link"
import {useRouter} from "next/navigation"
import {
    ArrowLeft, CalendarDays, Camera, CheckCircle2, ClipboardCheck, Clock3, Eye, FileCheck2, GraduationCap,
    KeyRound, Mail, NotebookText, Pencil, Phone, Search, SearchX, ShieldCheck, Trash2, TriangleAlert,
    UserRound,
} from "lucide-react"
import {toast} from "sonner"
import LocalDateTime from "@/components/local-date-time"
import UserAvatar from "@/components/app/user_avatar"
import {Badge} from "@/components/ui/badge"
import {Button} from "@/components/ui/button"
import {Card, CardContent, CardDescription, CardHeader, CardTitle} from "@/components/ui/card"
import {
    Dialog, DialogDescription, DialogHeader, DialogPopup, DialogTitle,
} from "@/components/ui/dialog"
import {
    AlertDialog, AlertDialogClose, AlertDialogDescription, AlertDialogFooter,
    AlertDialogHeader, AlertDialogPopup, AlertDialogTitle, AlertDialogTrigger,
} from "@/components/ui/alert-dialog"
import {Field, FieldLabel} from "@/components/ui/field"
import {Input} from "@/components/ui/input"
import {Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle} from "@/components/ui/empty"
import {Progress} from "@/components/ui/progress"
import {Tabs, TabsList, TabsTab} from "@/components/ui/tabs"
import {useLocale} from "@/i18n/provider"
import {type Locale} from "@/i18n/config"
import {type MessageKey} from "@/i18n/messages"
import {type ProfilePicture} from "@/lib/profile_picture"
import {invalidateProfilePictureCache} from "@/lib/profile_picture_cache"
import {hasSchoolPermission, type SchoolAccess} from "@/lib/school_access"
import {getActiveSchoolSnapshot, rememberCurrentSchoolAccess} from "@/lib/school_navigation"
import {uploadToPresignedURL} from "@/lib/upload_to_presigned_url"
import {
    handleDeleteStudent, handleUpdateStudent,
} from "@/app/app/(protected)/staff/(school_navbar)/schools/[id]/students/actions"
import {studentErrorKeys} from "@/app/app/(protected)/staff/(school_navbar)/schools/[id]/students/student_action_errors"
import StudentForm, {
    studentDraftToInput, type StudentDraft,
} from "@/app/app/(protected)/staff/(school_navbar)/schools/[id]/students/student_form"
import {scoreAccent} from "@/lib/score"
import {
    handleCompleteStudentProfilePictureUpload,
    handleInitStudentProfilePictureUpload,
    handleRemoveStudentProfilePicture,
    type StudentProfilePictureErrorCode,
} from "./actions"

export interface StaffStudentProfile {
    id: string
    name: string
    lastName: string
    dateOfBirth: string
    email: string
    phone: string
    notes: string
    accountEnabled: boolean
    createdAt: string
    profilePicture: ProfilePicture | null
}

export interface StudentAssignmentSubmission {
    id: string
    submittedAt: string
    notes: string | null
    score: {scorePercentage: number} | null
    assignment: {
        id: string
        title: string
        description: string | null
        course: {
            id: string
            name: string
            accentColor: string
            grade: {
                id: string
                name: string
                level: number
            }
        }
    }
}

type SubmissionFilter = "all" | "graded" | "ungraded"
const SUBMISSIONS_PAGE_SIZE = 8

const profilePictureErrorKeys = {
    invalid_upload: "staff.studentProfile.profilePicture.error.upload",
    network: "staff.studentProfile.profilePicture.error.network",
    unauthorized: "staff.studentProfile.profilePicture.error.unauthorized",
    forbidden: "staff.studentProfile.profilePicture.error.forbidden",
    validation: "staff.studentProfile.profilePicture.error.validation",
    server: "staff.studentProfile.profilePicture.error.server",
    upload: "staff.studentProfile.profilePicture.error.upload",
    invalid_student: "staff.studentProfile.profilePicture.error.invalidStudent",
    invalid_file: "staff.studentProfile.profilePicture.error.invalidFile",
    conflict: "staff.studentProfile.profilePicture.error.conflict",
} satisfies Record<StudentProfilePictureErrorCode, MessageKey>

function normalizeColor(value: string) {
    return /^[0-9A-Fa-f]{6}$/.test(value) ? `#${value}` : "#6366F1"
}

function formatDateOnly(value: string, locale: Locale) {
    const date = new Date(`${value}T00:00:00Z`)
    if (Number.isNaN(date.getTime())) return value
    return new Intl.DateTimeFormat(locale === "pl" ? "pl-PL" : "en-US", {
        day: "numeric",
        month: "long",
        year: "numeric",
        timeZone: "UTC",
    }).format(date)
}

function studentDraft(student: StaffStudentProfile): StudentDraft {
    return {
        name: student.name,
        lastName: student.lastName,
        dateOfBirth: student.dateOfBirth,
        email: student.email,
        phone: student.phone,
        notes: student.notes,
        accountEnabled: student.accountEnabled,
        password: "",
    }
}

export default function StudentProfileClientPage({initialStudent, assignmentSubmissions, access}: {
    initialStudent: StaffStudentProfile
    assignmentSubmissions: StudentAssignmentSubmission[]
    access: SchoolAccess
}) {
    const router = useRouter()
    const {locale, t} = useLocale()
    const [student, setStudent] = useState(initialStudent)
    const [profilePictureOverride, setProfilePictureOverride] = useState<ProfilePicture | null>()
    const profilePictureInput = useRef<HTMLInputElement>(null)
    const profilePicturePreviewURL = useRef<string | null>(null)
    const [profilePictureUploading, setProfilePictureUploading] = useState(false)
    const [profilePictureRemoving, setProfilePictureRemoving] = useState(false)
    const [profilePictureProgress, setProfilePictureProgress] = useState(0)
    const [editOpen, setEditOpen] = useState(false)
    const [draft, setDraft] = useState<StudentDraft>(() => studentDraft(initialStudent))
    const [saving, setSaving] = useState(false)
    const [deleteOpen, setDeleteOpen] = useState(false)
    const [deleteConfirmation, setDeleteConfirmation] = useState("")
    const [deleting, setDeleting] = useState(false)
    const [submissionSearch, setSubmissionSearch] = useState("")
    const [submissionFilter, setSubmissionFilter] = useState<SubmissionFilter>("all")
    const [visibleSubmissions, setVisibleSubmissions] = useState(SUBMISSIONS_PAGE_SIZE)
    const canUpdate = hasSchoolPermission(access, "student.update")
    const canDelete = hasSchoolPermission(access, "student.delete")
    const canListSubmissions = hasSchoolPermission(access, "submission.list")
    const canViewSubmissions = hasSchoolPermission(access, "submission.view")
    const fullName = `${student.name} ${student.lastName}`
    const profilePicture = profilePictureOverride === undefined
        ? initialStudent.profilePicture
        : profilePictureOverride
    const sortedSubmissions = useMemo(() => [...assignmentSubmissions].sort((left, right) =>
        Date.parse(right.submittedAt) - Date.parse(left.submittedAt)
    ), [assignmentSubmissions])
    const gradedCount = useMemo(() => sortedSubmissions.filter((submission) => submission.score !== null).length, [sortedSubmissions])
    const filteredSubmissions = useMemo(() => {
        const query = submissionSearch.trim().toLocaleLowerCase(locale === "pl" ? "pl-PL" : "en-US")
        return sortedSubmissions.filter((submission) => {
            if (submissionFilter === "graded" && !submission.score) return false
            if (submissionFilter === "ungraded" && submission.score) return false
            if (!query) return true

            return [
                submission.assignment.title,
                submission.assignment.course.name,
                submission.assignment.course.grade.name,
            ].some((value) => value.toLocaleLowerCase(locale === "pl" ? "pl-PL" : "en-US").includes(query))
        })
    }, [locale, sortedSubmissions, submissionFilter, submissionSearch])
    const displayedSubmissions = filteredSubmissions.slice(0, visibleSubmissions)

    useEffect(() => {
        rememberCurrentSchoolAccess(access)
    }, [access])

    useEffect(() => () => {
        if (profilePicturePreviewURL.current) URL.revokeObjectURL(profilePicturePreviewURL.current)
    }, [])

    async function uploadProfilePicture(event: React.ChangeEvent<HTMLInputElement>) {
        const file = event.target.files?.[0]
        event.target.value = ""
        if (!file || !canUpdate || profilePicture || profilePictureUploading) return

        setProfilePictureUploading(true)
        setProfilePictureProgress(0)

        const initialized = await handleInitStudentProfilePictureUpload(student.id, {
            name: file.name,
            size: file.size,
            type: file.type,
        })
        if (!initialized.ok) {
            setProfilePictureUploading(false)
            return toast.error(t(profilePictureErrorKeys[initialized.code]))
        }

        try {
            await uploadToPresignedURL(file, initialized.upload.url, setProfilePictureProgress)
        } catch {
            const cleanup = await handleRemoveStudentProfilePicture(student.id)
            setProfilePictureUploading(false)
            toast.error(t("staff.studentProfile.profilePicture.error.upload"))
            if (!cleanup.ok) toast.error(t("staff.studentProfile.profilePicture.error.cleanup"))
            return
        }

        const completed = await handleCompleteStudentProfilePictureUpload(
            initialized.upload.id,
            initialized.upload.completionToken,
        )
        if (!completed.ok) {
            const cleanup = await handleRemoveStudentProfilePicture(student.id)
            setProfilePictureUploading(false)
            toast.error(t(profilePictureErrorKeys[completed.code]))
            if (!cleanup.ok) toast.error(t("staff.studentProfile.profilePicture.error.cleanup"))
            return
        }

        const previewURL = URL.createObjectURL(file)
        profilePicturePreviewURL.current = previewURL
        setProfilePictureOverride({presignedUrl: previewURL})
        setProfilePictureUploading(false)
        invalidateProfilePictureCache(`student:${student.id}`)
        toast.success(t("staff.studentProfile.profilePicture.uploaded"))
        router.refresh()
    }

    async function removeProfilePicture() {
        if (!canUpdate || profilePictureRemoving) return

        setProfilePictureRemoving(true)
        const result = await handleRemoveStudentProfilePicture(student.id)
        setProfilePictureRemoving(false)
        if (!result.ok) return toast.error(t(profilePictureErrorKeys[result.code]))

        if (profilePicturePreviewURL.current) {
            URL.revokeObjectURL(profilePicturePreviewURL.current)
            profilePicturePreviewURL.current = null
        }
        setProfilePictureOverride(null)
        invalidateProfilePictureCache(`student:${student.id}`)
        toast.success(t("staff.studentProfile.profilePicture.removed"))
        router.refresh()
    }

    function openEdit() {
        setDraft(studentDraft(student))
        setEditOpen(true)
    }

    async function updateStudent(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        if (!canUpdate || saving) return

        const input = studentDraftToInput(draft)
        setSaving(true)
        const result = await handleUpdateStudent(student.id, input)
        setSaving(false)
        if (!result.ok) return toast.error(t(studentErrorKeys[result.code]))

        setStudent((current) => ({
            ...current,
            name: input.name,
            lastName: input.lastName,
            dateOfBirth: input.dob ?? current.dateOfBirth,
            email: input.email,
            phone: input.phone,
            notes: input.notes,
            accountEnabled: input.accountEnabled,
        }))
        setEditOpen(false)
        toast.success(t("staff.students.updated"))
        router.refresh()
    }

    async function deleteStudent() {
        if (!canDelete || deleting || deleteConfirmation !== fullName) return

        setDeleting(true)
        const result = await handleDeleteStudent(student.id)
        setDeleting(false)
        if (!result.ok) return toast.error(t(studentErrorKeys[result.code]))

        setDeleteOpen(false)
        toast.success(t("staff.students.deleted"))
        const schoolID = getActiveSchoolSnapshot()
        router.replace(schoolID ? `/app/staff/schools/${schoolID}/students` : "/app")
        router.refresh()
    }

    return (
        <div className="mx-auto w-full max-w-6xl space-y-6">
            <Button className="-ml-3" variant="ghost" onClick={() => router.back()}>
                <ArrowLeft /> {t("staff.studentProfile.back")}
            </Button>

            <header className="flex flex-col gap-5 sm:flex-row sm:items-start sm:justify-between">
                <div className="flex min-w-0 items-center gap-4">
                    <UserAvatar
                        className="size-16 border shadow-xs"
                        fallbackClassName="text-lg font-semibold"
                        name={fullName}
                        src={profilePicture?.presignedUrl}
                        cacheKey={`student:${student.id}`}
                    />
                    <div className="min-w-0">
                        <div className="flex flex-wrap items-center gap-2">
                            <h1 className="wrap-break-word text-3xl font-semibold tracking-tight sm:text-4xl">{fullName}</h1>
                            <Badge variant={student.accountEnabled ? "success" : "secondary"}>
                                {student.accountEnabled ? <ShieldCheck /> : <KeyRound />}
                                {t(student.accountEnabled ? "staff.students.loginEnabled" : "staff.students.noLogin")}
                            </Badge>
                        </div>
                        <a className="mt-2 inline-flex items-center gap-1.5 text-sm text-muted-foreground underline-offset-4 hover:text-foreground hover:underline" href={`mailto:${student.email}`}>
                            <Mail className="size-4" /> {student.email}
                        </a>
                        {canUpdate && (
                            <div className="mt-2">
                                {!profilePicture ? (
                                    <>
                                        <input
                                            ref={profilePictureInput}
                                            className="sr-only"
                                            type="file"
                                            accept="image/jpeg,image/png,image/gif,image/webp"
                                            disabled={profilePictureUploading}
                                            onChange={uploadProfilePicture}
                                        />
                                        <Button
                                            type="button"
                                            size="xs"
                                            variant="outline"
                                            loading={profilePictureUploading}
                                            disabled={profilePictureUploading}
                                            onClick={() => profilePictureInput.current?.click()}
                                        >
                                            <Camera /> {t("staff.studentProfile.profilePicture.add")}
                                        </Button>
                                    </>
                                ) : (
                                    <AlertDialog>
                                        <AlertDialogTrigger render={<Button type="button" size="xs" variant="destructive-outline" />}>
                                            <Trash2 /> {t("staff.studentProfile.profilePicture.remove")}
                                        </AlertDialogTrigger>
                                        <AlertDialogPopup>
                                            <AlertDialogHeader>
                                                <AlertDialogTitle>{t("staff.studentProfile.profilePicture.removeTitle", {name: fullName})}</AlertDialogTitle>
                                                <AlertDialogDescription>{t("staff.studentProfile.profilePicture.removeDescription")}</AlertDialogDescription>
                                            </AlertDialogHeader>
                                            <AlertDialogFooter>
                                                <AlertDialogClose render={<Button variant="outline" disabled={profilePictureRemoving} />}>
                                                    {t("common.cancel")}
                                                </AlertDialogClose>
                                                <Button
                                                    type="button"
                                                    variant="destructive"
                                                    loading={profilePictureRemoving}
                                                    disabled={profilePictureRemoving}
                                                    onClick={removeProfilePicture}
                                                >
                                                    {t("staff.studentProfile.profilePicture.removeConfirm")}
                                                </Button>
                                            </AlertDialogFooter>
                                        </AlertDialogPopup>
                                    </AlertDialog>
                                )}
                                {profilePictureUploading && (
                                    <div className="mt-2 w-52 max-w-full space-y-1.5">
                                        <Progress value={profilePictureProgress} />
                                        <p className="text-xs text-muted-foreground">
                                            {t(profilePictureProgress < 100
                                                ? "staff.studentProfile.profilePicture.uploading"
                                                : "staff.studentProfile.profilePicture.verifying", {
                                                progress: profilePictureProgress,
                                            })}
                                        </p>
                                    </div>
                                )}
                            </div>
                        )}
                    </div>
                </div>
                {(canUpdate || canDelete) && (
                    <div className="flex shrink-0 flex-wrap gap-2">
                        {canUpdate && <Button variant="outline" onClick={openEdit}><Pencil /> {t("staff.studentProfile.edit")}</Button>}
                        {canDelete && <Button variant="destructive-outline" onClick={() => {
                            setDeleteConfirmation("")
                            setDeleteOpen(true)
                        }}><Trash2 /> {t("staff.studentProfile.delete")}</Button>}
                    </div>
                )}
            </header>

            <div className="grid items-start gap-5 lg:grid-cols-[minmax(0,1.35fr)_minmax(18rem,0.65fr)]">
                <div className="space-y-5">
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2"><UserRound /> {t("staff.studentProfile.personalTitle")}</CardTitle>
                            <CardDescription>{t("staff.studentProfile.personalDescription")}</CardDescription>
                        </CardHeader>
                        <CardContent className="grid gap-5 sm:grid-cols-2">
                            <ProfileField icon={CalendarDays} label={t("staff.students.dateOfBirth")} value={formatDateOnly(student.dateOfBirth, locale)} />
                            <ProfileField icon={Clock3} label={t("staff.studentProfile.createdAt")} value={<LocalDateTime value={student.createdAt} />} />
                            <ProfileField icon={Mail} label={t("staff.students.email")} value={<a className="underline-offset-4 hover:underline" href={`mailto:${student.email}`}>{student.email}</a>} />
                            <ProfileField icon={Phone} label={t("staff.students.phone")} value={student.phone ? <a className="underline-offset-4 hover:underline" href={`tel:${student.phone}`}>{student.phone}</a> : t("staff.studentProfile.notProvided")} />
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2"><NotebookText /> {t("staff.students.notes")}</CardTitle>
                            <CardDescription>{t("staff.studentProfile.notesDescription")}</CardDescription>
                        </CardHeader>
                        <CardContent>
                            {student.notes ? (
                                <p className="whitespace-pre-wrap wrap-break-word rounded-xl border bg-muted/20 p-4 text-sm leading-7">{student.notes}</p>
                            ) : (
                                <p className="rounded-xl border border-dashed p-5 text-sm italic text-muted-foreground">{t("staff.studentProfile.noNotes")}</p>
                            )}
                        </CardContent>
                    </Card>
                </div>

                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2"><KeyRound /> {t("staff.studentProfile.accessTitle")}</CardTitle>
                        <CardDescription>{t("staff.studentProfile.accessDescription")}</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="rounded-xl border bg-muted/20 p-4">
                            <Badge variant={student.accountEnabled ? "success" : "secondary"}>
                                {t(student.accountEnabled ? "staff.studentProfile.enabled" : "staff.studentProfile.disabled")}
                            </Badge>
                            <p className="mt-3 text-sm leading-6 text-muted-foreground">
                                {t(student.accountEnabled ? "staff.studentProfile.enabledDescription" : "staff.studentProfile.disabledDescription")}
                            </p>
                        </div>
                    </CardContent>
                </Card>
            </div>

            {canListSubmissions && (
                <Card>
                    <CardHeader className="gap-4 sm:flex-row sm:items-start sm:justify-between">
                        <div>
                            <CardTitle className="flex items-center gap-2"><ClipboardCheck /> {t("staff.studentProfile.submissionsTitle")}</CardTitle>
                            <CardDescription className="mt-1">
                                {t("staff.studentProfile.submissionsDescription", {name: fullName})}
                            </CardDescription>
                        </div>
                        <div className="flex shrink-0 flex-wrap gap-2">
                            <Badge variant="secondary">{t("staff.studentProfile.submissionsCount", {count: assignmentSubmissions.length})}</Badge>
                            <Badge variant="success">{t("staff.studentProfile.gradedCount", {count: gradedCount})}</Badge>
                        </div>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        {assignmentSubmissions.length > 0 ? (
                            <>
                                <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                                    <div className="relative w-full lg:max-w-sm">
                                        <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                                        <Input
                                            className="pl-9"
                                            value={submissionSearch}
                                            onChange={(event) => {
                                                setSubmissionSearch(event.target.value)
                                                setVisibleSubmissions(SUBMISSIONS_PAGE_SIZE)
                                            }}
                                            placeholder={t("staff.studentProfile.submissionsSearch")}
                                            aria-label={t("staff.studentProfile.submissionsSearchLabel")}
                                        />
                                    </div>
                                    <Tabs value={submissionFilter} onValueChange={(value) => {
                                        if (value !== "all" && value !== "graded" && value !== "ungraded") return
                                        setSubmissionFilter(value)
                                        setVisibleSubmissions(SUBMISSIONS_PAGE_SIZE)
                                    }}>
                                        <TabsList className="w-full lg:w-auto" aria-label={t("staff.studentProfile.submissionsFilterLabel")}>
                                            <TabsTab className="flex-1" value="all">{t("staff.studentProfile.filterAll")}</TabsTab>
                                            <TabsTab className="flex-1" value="graded">{t("staff.studentProfile.filterGraded")}</TabsTab>
                                            <TabsTab className="flex-1" value="ungraded">{t("staff.studentProfile.filterUngraded")}</TabsTab>
                                        </TabsList>
                                    </Tabs>
                                </div>

                                {displayedSubmissions.length > 0 ? (
                                    <div className="grid gap-3">
                                        {displayedSubmissions.map((submission) => {
                                            const score = submission.score?.scorePercentage
                                            const courseColor = normalizeColor(submission.assignment.course.accentColor)
                                            return (
                                                <article className="relative overflow-hidden rounded-2xl border bg-card p-4 sm:p-5" key={submission.id}>
                                                    <span className="absolute inset-y-0 left-0 w-1" style={{backgroundColor: courseColor}} />
                                                    <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                                                        <div className="min-w-0 space-y-2 pl-1">
                                                            <div className="flex flex-wrap items-center gap-2">
                                                                <h3 className="wrap-break-word font-semibold">{submission.assignment.title}</h3>
                                                                {score !== undefined ? (
                                                                    <Badge variant="outline" className="font-semibold" style={{color: scoreAccent(score), borderColor: scoreAccent(score)}}>
                                                                        <CheckCircle2 /> {t("staff.studentProfile.score", {score})}
                                                                    </Badge>
                                                                ) : (
                                                                    <Badge variant="secondary">{t("staff.studentProfile.awaitingGrade")}</Badge>
                                                                )}
                                                            </div>
                                                            <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-muted-foreground">
                                                                <span className="inline-flex items-center gap-1.5"><FileCheck2 className="size-4" /> {submission.assignment.course.name}</span>
                                                                <span className="inline-flex items-center gap-1.5"><GraduationCap className="size-4" /> {submission.assignment.course.grade.name}</span>
                                                            </div>
                                                            {submission.assignment.description && (
                                                                <p className="line-clamp-2 text-sm leading-6 text-muted-foreground">{submission.assignment.description}</p>
                                                            )}
                                                            {submission.notes && (
                                                                <div className="rounded-xl bg-muted/35 px-3 py-2 text-sm leading-6">
                                                                    <span className="font-medium">{t("staff.studentProfile.submissionNote")}: </span>
                                                                    <span className="whitespace-pre-wrap wrap-break-word text-muted-foreground">{submission.notes}</span>
                                                                </div>
                                                            )}
                                                        </div>
                                                        <div className="flex shrink-0 flex-col items-start gap-2 sm:items-end">
                                                            <span className="inline-flex items-center gap-1.5 text-xs text-muted-foreground">
                                                                <Clock3 className="size-3.5" />
                                                                {t("staff.studentProfile.submittedAt")} <LocalDateTime value={submission.submittedAt} />
                                                            </span>
                                                            {canViewSubmissions && (
                                                                <Button size="sm" variant="outline" render={<Link href={`/app/staff/submissions/${submission.id}`} />}>
                                                                    <Eye /> {t("staff.studentProfile.viewSubmission")}
                                                                </Button>
                                                            )}
                                                        </div>
                                                    </div>
                                                </article>
                                            )
                                        })}
                                    </div>
                                ) : (
                                    <Empty className="rounded-2xl border border-dashed py-12 md:py-12">
                                        <EmptyHeader>
                                            <EmptyMedia variant="icon"><SearchX /></EmptyMedia>
                                            <EmptyTitle>{t("staff.studentProfile.noMatchingSubmissions")}</EmptyTitle>
                                            <EmptyDescription>{t("staff.studentProfile.noMatchingSubmissionsDescription")}</EmptyDescription>
                                        </EmptyHeader>
                                        <Button type="button" variant="outline" onClick={() => {
                                            setSubmissionSearch("")
                                            setSubmissionFilter("all")
                                            setVisibleSubmissions(SUBMISSIONS_PAGE_SIZE)
                                        }}>{t("staff.studentProfile.clearSubmissionFilters")}</Button>
                                    </Empty>
                                )}

                                {displayedSubmissions.length < filteredSubmissions.length && (
                                    <div className="flex flex-col items-center gap-2 border-t pt-4">
                                        <p className="text-xs text-muted-foreground">
                                            {t("staff.studentProfile.showingSubmissions", {shown: displayedSubmissions.length, total: filteredSubmissions.length})}
                                        </p>
                                        <Button type="button" variant="outline" onClick={() => setVisibleSubmissions((current) => current + SUBMISSIONS_PAGE_SIZE)}>
                                            {t("staff.studentProfile.showMoreSubmissions")}
                                        </Button>
                                    </div>
                                )}
                            </>
                        ) : (
                            <Empty className="rounded-2xl border border-dashed py-12 md:py-12">
                                <EmptyHeader>
                                    <EmptyMedia variant="icon"><FileCheck2 /></EmptyMedia>
                                    <EmptyTitle>{t("staff.studentProfile.noSubmissions")}</EmptyTitle>
                                    <EmptyDescription>{t("staff.studentProfile.noSubmissionsDescription", {name: fullName})}</EmptyDescription>
                                </EmptyHeader>
                            </Empty>
                        )}
                    </CardContent>
                </Card>
            )}

            <Dialog open={editOpen} onOpenChange={(open) => {
                if (!saving) setEditOpen(open)
            }}>
                <DialogPopup className="sm:max-w-2xl">
                    <DialogHeader>
                        <DialogTitle>{t("staff.students.editTitle")}</DialogTitle>
                        <DialogDescription>{t("staff.students.editDescription")}</DialogDescription>
                    </DialogHeader>
                    <StudentForm
                        draft={draft}
                        saving={saving}
                        mode="edit"
                        wasAccountEnabled={student.accountEnabled}
                        onChange={setDraft}
                        onSubmit={updateStudent}
                    />
                </DialogPopup>
            </Dialog>

            <AlertDialog open={deleteOpen} onOpenChange={(open) => {
                if (!deleting) {
                    setDeleteOpen(open)
                    if (!open) setDeleteConfirmation("")
                }
            }}>
                <AlertDialogPopup className="border-destructive/30 sm:max-w-xl">
                    <AlertDialogHeader>
                        <div className="mb-2 flex items-start gap-3 rounded-xl border border-destructive/30 bg-destructive/8 p-4 text-left">
                            <span className="flex size-10 shrink-0 items-center justify-center rounded-full bg-destructive/15 text-destructive"><TriangleAlert className="size-5" /></span>
                            <div>
                                <p className="font-semibold text-destructive">{t("staff.students.permanentDeletion")}</p>
                                <p className="mt-1 text-sm leading-6 text-muted-foreground">{t("staff.students.permanentDeletionWarning")}</p>
                            </div>
                        </div>
                        <AlertDialogTitle>{t("staff.students.deleteTitle", {name: fullName})}</AlertDialogTitle>
                        <AlertDialogDescription>{t("staff.students.deleteConfirmation")}</AlertDialogDescription>
                    </AlertDialogHeader>
                    <Field className="px-6 pb-2">
                        <FieldLabel htmlFor="student-profile-delete-confirmation">{t("staff.students.typeName", {name: fullName})}</FieldLabel>
                        <Input
                            id="student-profile-delete-confirmation"
                            value={deleteConfirmation}
                            onChange={(event) => setDeleteConfirmation(event.target.value)}
                            autoComplete="off"
                            disabled={deleting}
                            autoFocus
                        />
                    </Field>
                    <AlertDialogFooter>
                        <AlertDialogClose render={<Button variant="outline" disabled={deleting} />}>{t("staff.students.cancel")}</AlertDialogClose>
                        <Button variant="destructive" loading={deleting} disabled={deleting || deleteConfirmation !== fullName} onClick={deleteStudent}>
                            {t("staff.students.deletePermanently")}
                        </Button>
                    </AlertDialogFooter>
                </AlertDialogPopup>
            </AlertDialog>
        </div>
    )
}

function ProfileField({icon: Icon, label, value}: {
    icon: typeof Mail
    label: string
    value: React.ReactNode
}) {
    return (
        <div className="flex min-w-0 items-start gap-3">
            <span className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-muted text-muted-foreground"><Icon className="size-4" /></span>
            <div className="min-w-0">
                <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">{label}</p>
                <div className="mt-1 wrap-break-word text-sm font-medium">{value}</div>
            </div>
        </div>
    )
}

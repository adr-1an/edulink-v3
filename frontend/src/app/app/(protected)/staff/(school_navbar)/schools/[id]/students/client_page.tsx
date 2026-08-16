"use client"

import {useEffect, useMemo, useState} from "react"
import Link from "next/link"
import {useRouter} from "next/navigation"
import {toast} from "sonner"
import {
    CalendarDays, ChevronLeft, ChevronRight, FileUp, GraduationCap, Mail, Pencil, Phone, Plus, Search, Trash2, TriangleAlert, UserRoundCheck, X,
} from "lucide-react"
import PageTitle from "@/components/app/page_title"
import {Avatar, AvatarFallback} from "@/components/ui/avatar"
import {Badge} from "@/components/ui/badge"
import {Button} from "@/components/ui/button"
import {Card, CardContent} from "@/components/ui/card"
import {Dialog, DialogDescription, DialogHeader, DialogPopup, DialogTitle} from "@/components/ui/dialog"
import {
    AlertDialog, AlertDialogClose, AlertDialogDescription, AlertDialogFooter,
    AlertDialogHeader, AlertDialogPopup, AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import {Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle} from "@/components/ui/empty"
import {Field, FieldLabel} from "@/components/ui/field"
import {Input} from "@/components/ui/input"
import {hasSchoolPermission, type SchoolAccess} from "@/lib/school_access"
import {rememberSchoolAccess} from "@/lib/school_navigation"
import {handleCreateStudent, handleDeleteStudent, handleUpdateStudent} from "./actions"
import {useLocale} from "@/i18n/provider"
import {type Locale} from "@/i18n/config"
import StudentForm, {
    emptyStudentDraft, studentDraftToInput, type StudentDraft,
} from "./student_form"
import {studentErrorKeys} from "./student_action_errors"
import BulkStudentImport from "./bulk_student_import"

export interface Student {
    id: string
    name: string
    lastName: string
    dateOfBirth: string | null
    dateOfBirthLabel: string | null
    email: string
    phone: string
    notes: string
    accountEnabled: boolean
    createdAt: string
}

const pageSize = 20

function initials(student: Pick<Student, "name" | "lastName">) {
    return `${student.name[0] ?? ""}${student.lastName[0] ?? ""}`.toUpperCase() || "?"
}

function formatDateOnly(value: string | null, locale: Locale) {
    if (!value) return null
    const date = new Date(`${value}T00:00:00Z`)
    if (Number.isNaN(date.getTime())) return null
    return new Intl.DateTimeFormat(locale === "pl" ? "pl-PL" : "en-US", {day: "numeric", month: "short", year: "numeric", timeZone: "UTC"}).format(date)
}

export default function StudentsClientPage({schoolID, initialStudents, access}: {
    schoolID: string
    initialStudents: Student[]
    access: SchoolAccess
}) {
    const router = useRouter()
    const {locale, t} = useLocale()
    const [students, setStudents] = useState(initialStudents)
    const [query, setQuery] = useState("")
    const [requestedPage, setRequestedPage] = useState(1)
    const [createOpen, setCreateOpen] = useState(false)
    const [importOpen, setImportOpen] = useState(false)
    const [editTarget, setEditTarget] = useState<Student | null>(null)
    const [deleteTarget, setDeleteTarget] = useState<Student | null>(null)
    const [deleteConfirmation, setDeleteConfirmation] = useState("")
    const [draft, setDraft] = useState<StudentDraft>(emptyStudentDraft)
    const [saving, setSaving] = useState(false)
    const [deleting, setDeleting] = useState(false)
    const canView = hasSchoolPermission(access, "student.view")
    const canCreate = hasSchoolPermission(access, "student.create")
    const canUpdate = hasSchoolPermission(access, "student.update")
    const canDelete = hasSchoolPermission(access, "student.delete")

    useEffect(() => {
        rememberSchoolAccess(schoolID, access)
    }, [schoolID, access])

    const filteredStudents = useMemo(() => {
        const search = query.trim().toLocaleLowerCase()
        return [...students]
            .filter((student) => !search || [student.name, student.lastName, student.email, student.phone]
                .some((value) => value.toLocaleLowerCase().includes(search)))
            .sort((first, second) => first.lastName.localeCompare(second.lastName) || first.name.localeCompare(second.name))
    }, [query, students])
    const pageCount = Math.max(1, Math.ceil(filteredStudents.length / pageSize))
    const currentPage = Math.min(requestedPage, pageCount)
    const visibleStudents = filteredStudents.slice((currentPage - 1) * pageSize, currentPage * pageSize)
    const firstVisibleStudent = filteredStudents.length === 0 ? 0 : (currentPage - 1) * pageSize + 1
    const lastVisibleStudent = Math.min(currentPage * pageSize, filteredStudents.length)
    const enabledAccounts = students.filter((student) => student.accountEnabled).length
    const requiredDeleteConfirmation = deleteTarget ? `${deleteTarget.name} ${deleteTarget.lastName}` : ""

    function openCreate() {
        setDraft(emptyStudentDraft)
        setCreateOpen(true)
    }

    function openEdit(student: Student) {
        setDraft({
            name: student.name,
            lastName: student.lastName,
            dateOfBirth: student.dateOfBirth ?? "",
            email: student.email,
            phone: student.phone,
            notes: student.notes,
            accountEnabled: student.accountEnabled,
            password: "",
        })
        setEditTarget(student)
    }

    function openDelete(student: Student) {
        setDeleteConfirmation("")
        setDeleteTarget(student)
    }

    async function createStudent(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        setSaving(true)
        const res = await handleCreateStudent(schoolID, studentDraftToInput(draft))
        setSaving(false)
        if (!res.ok) return toast.error(t(studentErrorKeys[res.code]))

        setCreateOpen(false)
        toast.success(t("staff.students.created"))
        router.refresh()
    }

    async function updateStudent(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        if (!editTarget) return
        const input = studentDraftToInput(draft)
        setSaving(true)
        const res = await handleUpdateStudent(editTarget.id, input)
        setSaving(false)
        if (!res.ok) return toast.error(t(studentErrorKeys[res.code]))

        const updated: Student = {
            ...editTarget,
            name: input.name,
            lastName: input.lastName,
            dateOfBirth: input.dob,
            dateOfBirthLabel: formatDateOnly(input.dob, locale),
            email: input.email,
            phone: input.phone,
            notes: input.notes,
            accountEnabled: input.accountEnabled,
        }
        setStudents((current) => current.map((student) => student.id === editTarget.id ? updated : student))
        setEditTarget(null)
        toast.success(t("staff.students.updated"))
    }

    async function deleteStudent() {
        if (!deleteTarget || deleteConfirmation !== requiredDeleteConfirmation) return
        setDeleting(true)
        const res = await handleDeleteStudent(deleteTarget.id)
        setDeleting(false)
        if (!res.ok) return toast.error(t(studentErrorKeys[res.code]))

        setStudents((current) => current.filter((student) => student.id !== deleteTarget.id))
        setDeleteTarget(null)
        setDeleteConfirmation("")
        toast.success(t("staff.students.deleted"))
    }

    return (
        <div className="space-y-6">
            <div className="flex flex-wrap items-end justify-between gap-3">
                <div>
                    <PageTitle centered={false}>{t("staff.students.title")}</PageTitle>
                    <p className="text-muted-foreground">{t("staff.students.description")}</p>
                </div>
                {canCreate && (
                    <div className="flex flex-wrap gap-2">
                        <Button variant="outline" onClick={() => setImportOpen(true)}><FileUp /> {t("staff.students.import.action")}</Button>
                        <Button onClick={openCreate}><Plus /> {t("staff.students.add")}</Button>
                    </div>
                )}
            </div>

            <div className="grid gap-3 sm:grid-cols-2">
                <SummaryCard icon={GraduationCap} label={t("staff.students.summaryStudents")} value={students.length} />
                <SummaryCard icon={UserRoundCheck} label={t("staff.students.loginEnabled")} value={enabledAccounts} />
            </div>

            {students.length > 0 && (
                <div className="relative max-w-md">
                    <Search className="pointer-events-none absolute left-3 top-1/2 z-10 size-4 -translate-y-1/2 text-muted-foreground" />
                    <Input className="pl-9 pr-9" value={query} onChange={(event) => { setQuery(event.target.value); setRequestedPage(1) }} placeholder={t("staff.students.searchPlaceholder")} aria-label={t("staff.students.search")} />
                    {query && <Button className="absolute right-1.5 top-1/2 -translate-y-1/2" size="icon-xs" variant="ghost" aria-label={t("staff.students.clearSearch")} onClick={() => { setQuery(""); setRequestedPage(1) }}><X /></Button>}
                </div>
            )}

            {students.length === 0 ? (
                <Card>
                    <Empty>
                        <EmptyHeader>
                            <EmptyMedia variant="icon"><GraduationCap /></EmptyMedia>
                            <EmptyTitle>{t("staff.students.empty")}</EmptyTitle>
                            <EmptyDescription>{t(canCreate ? "staff.students.emptyCreate" : "staff.students.emptyRestricted")}</EmptyDescription>
                        </EmptyHeader>
                        {canCreate && (
                            <div className="flex flex-wrap justify-center gap-2">
                                <Button variant="outline" onClick={() => setImportOpen(true)}><FileUp /> {t("staff.students.import.action")}</Button>
                                <Button onClick={openCreate}><Plus /> {t("staff.students.add")}</Button>
                            </div>
                        )}
                    </Empty>
                </Card>
            ) : filteredStudents.length === 0 ? (
                <Card className="p-8 text-center text-sm text-muted-foreground">{t("staff.students.noMatch", {query: query.trim()})}</Card>
            ) : (
                <Card className="overflow-hidden p-0">
                    <div className="divide-y">
                        {visibleStudents.map((student) => (
                            <div className="flex flex-col gap-3 px-4 py-4 sm:flex-row sm:items-center sm:px-5" key={student.id}>
                                <Avatar className="size-10 border"><AvatarFallback>{initials(student)}</AvatarFallback></Avatar>
                                <div className="min-w-0 flex-1">
                                    <div className="flex flex-wrap items-center gap-2">
                                        {canView ? (
                                            <Link className="truncate font-medium underline-offset-4 hover:underline" href={`/app/staff/students/${student.id}`}>
                                                {student.name} {student.lastName}
                                            </Link>
                                        ) : (
                                            <p className="truncate font-medium">{student.name} {student.lastName}</p>
                                        )}
                                        <Badge variant={student.accountEnabled ? "default" : "secondary"}>{t(student.accountEnabled ? "staff.students.loginEnabled" : "staff.students.noLogin")}</Badge>
                                    </div>
                                    <div className="mt-1 flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
                                        <span className="flex items-center gap-1.5"><Mail className="size-3.5" /> {student.email}</span>
                                        {student.phone && <span className="flex items-center gap-1.5"><Phone className="size-3.5" /> {student.phone}</span>}
                                        {student.dateOfBirthLabel && <span className="flex items-center gap-1.5"><CalendarDays className="size-3.5" /> {t("staff.students.born", {date: student.dateOfBirthLabel})}</span>}
                                    </div>
                                </div>
                                <div className="flex shrink-0 gap-2 self-end sm:self-auto">
                                    {canUpdate && <Button size="icon-sm" variant="ghost" aria-label={t("staff.students.editLabel", {name: `${student.name} ${student.lastName}`})} onClick={() => openEdit(student)}><Pencil /></Button>}
                                    {canDelete && <Button size="icon-sm" variant="destructive-outline" aria-label={t("staff.students.deleteLabel", {name: `${student.name} ${student.lastName}`})} onClick={() => openDelete(student)}><Trash2 /></Button>}
                                    {canView && <Button size="icon-sm" variant="ghost" aria-label={t("staff.students.viewLabel", {name: `${student.name} ${student.lastName}`})} render={<Link href={`/app/staff/students/${student.id}`} />}><ChevronRight /></Button>}
                                </div>
                            </div>
                        ))}
                    </div>
                    {filteredStudents.length > pageSize && (
                        <div className="flex flex-col gap-3 border-t bg-muted/20 px-4 py-3 sm:flex-row sm:items-center sm:justify-between sm:px-5">
                            <p className="text-xs text-muted-foreground">
                                {t("staff.students.pagination.showing", {
                                    from: firstVisibleStudent,
                                    to: lastVisibleStudent,
                                    total: filteredStudents.length,
                                })}
                            </p>
                            <div className="flex items-center justify-between gap-2 sm:justify-end">
                                <Button size="sm" variant="outline" disabled={currentPage === 1} onClick={() => setRequestedPage(Math.max(1, currentPage - 1))}>
                                    <ChevronLeft /> {t("staff.students.pagination.previous")}
                                </Button>
                                <span className="min-w-20 text-center text-sm tabular-nums">
                                    {t("staff.students.pagination.page", {current: currentPage, total: pageCount})}
                                </span>
                                <Button size="sm" variant="outline" disabled={currentPage === pageCount} onClick={() => setRequestedPage(Math.min(pageCount, currentPage + 1))}>
                                    {t("staff.students.pagination.next")} <ChevronRight />
                                </Button>
                            </div>
                        </div>
                    )}
                </Card>
            )}

            <Dialog open={createOpen} onOpenChange={(open) => { if (!saving) setCreateOpen(open) }}>
                <DialogPopup className="sm:max-w-2xl">
                    <DialogHeader><DialogTitle>{t("staff.students.createTitle")}</DialogTitle><DialogDescription>{t("staff.students.createDescription")}</DialogDescription></DialogHeader>
                    <StudentForm draft={draft} saving={saving} mode="create" onChange={setDraft} onSubmit={createStudent} />
                </DialogPopup>
            </Dialog>

            {canCreate && (
                <BulkStudentImport
                    open={importOpen}
                    schoolID={schoolID}
                    onOpenChange={setImportOpen}
                    onImported={() => router.refresh()}
                />
            )}

            <Dialog open={editTarget !== null} onOpenChange={(open) => { if (!open && !saving) setEditTarget(null) }}>
                <DialogPopup className="sm:max-w-2xl">
                    <DialogHeader><DialogTitle>{t("staff.students.editTitle")}</DialogTitle><DialogDescription>{t("staff.students.editDescription")}</DialogDescription></DialogHeader>
                    <StudentForm draft={draft} saving={saving} mode="edit" wasAccountEnabled={editTarget?.accountEnabled} onChange={setDraft} onSubmit={updateStudent} />
                </DialogPopup>
            </Dialog>

            <AlertDialog open={deleteTarget !== null} onOpenChange={(open) => {
                if (!open && !deleting) {
                    setDeleteTarget(null)
                    setDeleteConfirmation("")
                }
            }}>
                <AlertDialogPopup className="sm:max-w-xl">
                    <AlertDialogHeader>
                        <div className="mb-2 flex items-start gap-3 rounded-xl border border-destructive/30 bg-destructive/8 p-4 text-left">
                            <span className="flex size-10 shrink-0 items-center justify-center rounded-full bg-destructive/15 text-destructive"><TriangleAlert className="size-5" /></span>
                            <div>
                                <p className="font-semibold text-destructive">{t("staff.students.permanentDeletion")}</p>
                                <p className="mt-1 text-sm leading-6 text-muted-foreground">{t("staff.students.permanentDeletionWarning")}</p>
                            </div>
                        </div>
                        <AlertDialogTitle>{t("staff.students.deleteTitle", {name: requiredDeleteConfirmation})}</AlertDialogTitle>
                        <AlertDialogDescription>{t("staff.students.deleteConfirmation")}</AlertDialogDescription>
                        <Field className="pt-2">
                            <FieldLabel>{t("staff.students.typeName", {name: requiredDeleteConfirmation})}</FieldLabel>
                            <Input value={deleteConfirmation} onChange={(event) => setDeleteConfirmation(event.target.value)} disabled={deleting} autoComplete="off" autoFocus />
                        </Field>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogClose render={<Button variant="outline" disabled={deleting} />}>{t("staff.students.cancel")}</AlertDialogClose>
                        <Button variant="destructive" loading={deleting} disabled={deleteConfirmation !== requiredDeleteConfirmation || deleting} onClick={deleteStudent}>{t("staff.students.deletePermanently")}</Button>
                    </AlertDialogFooter>
                </AlertDialogPopup>
            </AlertDialog>
        </div>
    )
}

function SummaryCard({icon: Icon, label, value}: {icon: typeof GraduationCap; label: string; value: number}) {
    return (
        <Card className="py-4">
            <CardContent className="flex items-center gap-3 px-4">
                <span className="flex size-9 items-center justify-center rounded-lg bg-primary/10 text-primary"><Icon className="size-4.5" /></span>
                <div><p className="text-xl font-semibold tabular-nums">{value}</p><p className="text-xs text-muted-foreground">{label}</p></div>
            </CardContent>
        </Card>
    )
}

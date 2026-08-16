"use client"

import {useEffect, useMemo, useState} from "react"
import Link from "next/link"
import {useRouter} from "next/navigation"
import {BookOpen, ChevronLeft, Pencil, Plus, Search, Trash2, TriangleAlert} from "lucide-react"
import {toast} from "sonner"
import PageTitle from "@/components/app/page_title"
import {
    AlertDialog, AlertDialogClose, AlertDialogDescription, AlertDialogFooter,
    AlertDialogHeader, AlertDialogPopup, AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import {Button} from "@/components/ui/button"
import {Checkbox} from "@/components/ui/checkbox"
import {Card, CardContent, CardFooter, CardHeader, CardTitle} from "@/components/ui/card"
import {
    ContextMenu, ContextMenuItem, ContextMenuPopup, ContextMenuTrigger,
} from "@/components/ui/context-menu"
import {
    Dialog, DialogClose, DialogDescription, DialogFooter, DialogHeader,
    DialogPanel, DialogPopup, DialogTitle,
} from "@/components/ui/dialog"
import {Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle} from "@/components/ui/empty"
import {Field, FieldDescription, FieldLabel} from "@/components/ui/field"
import {Form} from "@/components/ui/form"
import {Input} from "@/components/ui/input"
import {Textarea} from "@/components/ui/textarea"
import {hasSchoolPermission, type SchoolAccess} from "@/lib/school_access"
import {rememberCurrentSchoolAccess} from "@/lib/school_navigation"
import {useLocale} from "@/i18n/provider"
import {handleCreateCourse, handleDeleteCourse, handleUpdateCourse} from "./actions"

interface Course {
    id: string
    name: string
    description: string
    color: string
}

interface CourseDraft {
    name: string
    description: string
    color: string
}

const emptyDraft: CourseDraft = {name: "", description: "", color: "6366F1"}

function normalizeColor(color: string) {
    return /^[0-9A-Fa-f]{6}$/.test(color) ? color.toUpperCase() : "000000"
}

export default function CoursesClientPage({gradeID, courses, access}: {
    gradeID: string
    courses: Course[]
    access: SchoolAccess
}) {
    const router = useRouter()
    const {locale, t} = useLocale()
    const [courseList, setCourseList] = useState(courses)
    const [query, setQuery] = useState("")
    const [createOpen, setCreateOpen] = useState(false)
    const [editTarget, setEditTarget] = useState<Course | null>(null)
    const [deleteTarget, setDeleteTarget] = useState<Course | null>(null)
    const [deleteConfirmation, setDeleteConfirmation] = useState("")
    const [deleteAcknowledged, setDeleteAcknowledged] = useState(false)
    const [draft, setDraft] = useState<CourseDraft>(emptyDraft)
    const [saving, setSaving] = useState(false)
    const [deleting, setDeleting] = useState(false)
    const canCreate = hasSchoolPermission(access, "course.create")
    const canUpdate = hasSchoolPermission(access, "course.update")
    const canDelete = hasSchoolPermission(access, "course.delete")

    useEffect(() => {
        rememberCurrentSchoolAccess(access)
    }, [access])

    const visibleCourses = useMemo(() => {
        const search = query.trim().toLocaleLowerCase()
        return [...courseList]
            .filter((course) => !search
                || course.name.toLocaleLowerCase().includes(search)
                || course.description.toLocaleLowerCase().includes(search)
            )
            .sort((first, second) => first.name.localeCompare(second.name, locale))
    }, [courseList, locale, query])

    function openCreate() {
        setDraft(emptyDraft)
        setCreateOpen(true)
    }

    function openEdit(course: Course) {
        setDraft({
            name: course.name,
            description: course.description,
            color: normalizeColor(course.color),
        })
        setEditTarget(course)
    }

    async function createCourse(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        setSaving(true)
        const res = await handleCreateCourse(gradeID, draft)
        setSaving(false)
        if (!res.ok) return toast.error(res.message)

        setCreateOpen(false)
        toast.success(t("staff.courses.created"))
        router.refresh()
    }

    async function updateCourse(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        if (!editTarget) return
        setSaving(true)
        const res = await handleUpdateCourse(editTarget.id, draft)
        setSaving(false)
        if (!res.ok) return toast.error(res.message)

        const updatedCourse = {
            ...editTarget,
            name: draft.name.trim(),
            description: draft.description.trim(),
            color: normalizeColor(draft.color),
        }
        setCourseList((current) => current.map((course) => course.id === editTarget.id ? updatedCourse : course))
        setEditTarget(null)
        toast.success(t("staff.courses.updated"))
    }

    async function deleteCourse() {
        if (!deleteTarget) return
        if (deleteConfirmation !== `DELETE ${deleteTarget.name}` || !deleteAcknowledged) return
        setDeleting(true)
        const res = await handleDeleteCourse(deleteTarget.id)
        setDeleting(false)
        if (!res.ok) return toast.error(res.message)

        setCourseList((current) => current.filter((course) => course.id !== deleteTarget.id))
        setDeleteTarget(null)
        setDeleteConfirmation("")
        setDeleteAcknowledged(false)
        toast.success(t("staff.courses.deleted"))
    }

    return (
        <div className="space-y-6">
            <div className="flex flex-wrap items-end justify-between gap-3">
                <div>
                    <Button className="mb-2 -ml-2" size="sm" variant="ghost" onClick={() => router.back()}>
                        <ChevronLeft /> {t("staff.courses.back")}
                    </Button>
                    <PageTitle centered={false}>{t("staff.courses.title")}</PageTitle>
                    <p className="text-muted-foreground">{t("staff.courses.description")}</p>
                </div>
                {canCreate && <Button onClick={openCreate}><Plus /> {t("staff.courses.create")}</Button>}
            </div>

            {courseList.length > 0 && (
                <div className="relative max-w-md">
                    <Search className="pointer-events-none absolute top-1/2 left-3 z-10 size-4 -translate-y-1/2 text-muted-foreground" />
                    <Input
                        className="pl-9"
                        value={query}
                        onChange={(event) => setQuery(event.target.value)}
                        placeholder={t("staff.courses.searchPlaceholder")}
                        aria-label={t("staff.courses.search")}
                    />
                </div>
            )}

            {courseList.length === 0 ? (
                <Card>
                    <Empty>
                        <EmptyHeader>
                            <EmptyMedia variant="icon"><BookOpen /></EmptyMedia>
                            <EmptyTitle>{t("staff.courses.empty")}</EmptyTitle>
                            <EmptyDescription>{t("staff.courses.emptyDescription")}</EmptyDescription>
                        </EmptyHeader>
                        {canCreate && <Button onClick={openCreate}><Plus /> {t("staff.courses.create")}</Button>}
                    </Empty>
                </Card>
            ) : visibleCourses.length === 0 ? (
                <Card className="p-8 text-center text-sm text-muted-foreground">
                    No courses match &ldquo;{query.trim()}&rdquo;.
                </Card>
            ) : (
                <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                    {visibleCourses.map((course) => (
                        <ContextMenu key={course.id}>
                            <ContextMenuTrigger className="flex">
                                <Card className="w-full overflow-hidden transition-[border-color,box-shadow,transform] hover:-translate-y-0.5 hover:shadow-md/10">
                                    <Link
                                        className="flex flex-1 flex-col rounded-t-2xl outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset"
                                        href={`/app/staff/courses/${course.id}`}
                                    >
                                        <div className="h-1.5" style={{backgroundColor: `#${normalizeColor(course.color)}`}} />
                                        <CardHeader>
                                            <div className="flex items-start gap-3">
                                                <span
                                                    className="mt-0.5 flex size-9 shrink-0 items-center justify-center rounded-lg text-white shadow-sm"
                                                    style={{backgroundColor: `#${normalizeColor(course.color)}`}}
                                                >
                                                    <BookOpen className="size-4" />
                                                </span>
                                                <div className="min-w-0">
                                                    <CardTitle className="truncate">{course.name}</CardTitle>
                                                </div>
                                            </div>
                                        </CardHeader>
                                        <CardContent className="flex-1">
                                            <p className="line-clamp-3 text-sm leading-relaxed text-muted-foreground">
                                                {course.description || t("staff.courses.noDescription")}
                                            </p>
                                        </CardContent>
                                    </Link>
                                    {(canUpdate || canDelete) && <CardFooter className="justify-end gap-2 border-t bg-muted/25">
                                        {canUpdate && <Button size="sm" variant="outline" onClick={() => openEdit(course)}><Pencil /> {t("staff.courses.edit")}</Button>}
                                        {canDelete && <Button size="icon-sm" variant="destructive-outline" aria-label={t("staff.courses.delete", {name: course.name})} onClick={() => setDeleteTarget(course)}><Trash2 /></Button>}
                                    </CardFooter>}
                                </Card>
                            </ContextMenuTrigger>
                            {(canUpdate || canDelete) && <ContextMenuPopup className="w-48">
                                {canUpdate && <ContextMenuItem onClick={() => openEdit(course)}><Pencil /> {t("staff.courses.editAction")}</ContextMenuItem>}
                                {canDelete && <ContextMenuItem variant="destructive" onClick={() => setDeleteTarget(course)}><Trash2 /> {t("staff.courses.deleteAction")}</ContextMenuItem>}
                            </ContextMenuPopup>}
                        </ContextMenu>
                    ))}
                </div>
            )}

            <Dialog open={createOpen} onOpenChange={(open) => {
                if (!saving) setCreateOpen(open)
            }}>
                <DialogPopup className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle>{t("staff.courses.create")}</DialogTitle>
                        <DialogDescription>{t("staff.courses.createDescription")}</DialogDescription>
                    </DialogHeader>
                    <CourseForm draft={draft} saving={saving} submitLabel={t("staff.courses.create")} onChange={setDraft} onSubmit={createCourse} />
                </DialogPopup>
            </Dialog>

            <Dialog open={editTarget !== null} onOpenChange={(open) => {
                if (!open && !saving) setEditTarget(null)
            }}>
                <DialogPopup className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle>{t("staff.courses.editTitle")}</DialogTitle>
                        <DialogDescription>{t("staff.courses.editDescription")}</DialogDescription>
                    </DialogHeader>
                    <CourseForm draft={draft} saving={saving} submitLabel={t("staff.courses.save")} onChange={setDraft} onSubmit={updateCourse} />
                </DialogPopup>
            </Dialog>

            <AlertDialog open={deleteTarget !== null} onOpenChange={(open) => {
                if (!open && !deleting) {
                    setDeleteTarget(null)
                    setDeleteConfirmation("")
                    setDeleteAcknowledged(false)
                }
            }}>
                <AlertDialogPopup className="border-destructive/30 sm:max-w-xl">
                    <AlertDialogHeader>
                        <div className="mb-2 flex size-12 items-center justify-center rounded-2xl bg-destructive/10 text-destructive">
                            <TriangleAlert className="size-6" />
                        </div>
                        <AlertDialogTitle className="text-destructive">{t("staff.courses.deleteTitle")}</AlertDialogTitle>
                        <AlertDialogDescription className="space-y-3 text-left">
                            {t("staff.courses.deleteDescription", {name: deleteTarget?.name ?? ""})}
                            <span className="block font-medium text-destructive">{t("staff.courses.deleteIrreversible")}</span>
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <div className="grid gap-4 px-6 pb-2">
                        <Field>
                            <FieldLabel htmlFor="course-delete-confirmation">
                                {t("staff.courses.typeToDelete", {value: deleteTarget ? `DELETE ${deleteTarget.name}` : ""})}
                            </FieldLabel>
                            <Input
                                id="course-delete-confirmation"
                                value={deleteConfirmation}
                                onChange={(event) => setDeleteConfirmation(event.target.value)}
                                placeholder={deleteTarget ? `DELETE ${deleteTarget.name}` : ""}
                                autoComplete="off"
                                spellCheck={false}
                                disabled={deleting}
                                autoFocus
                            />
                        </Field>
                        <label className="flex items-start gap-3 text-sm">
                            <Checkbox
                                checked={deleteAcknowledged}
                                onCheckedChange={(value) => setDeleteAcknowledged(value)}
                                disabled={deleting}
                            />
                            <span>{t("staff.courses.deleteAcknowledge")}</span>
                        </label>
                    </div>
                    <AlertDialogFooter>
                        <AlertDialogClose render={<Button variant="outline" disabled={deleting} />}>{t("staff.courses.cancel")}</AlertDialogClose>
                        <Button
                            variant="destructive"
                            loading={deleting}
                            disabled={!deleteTarget
                                || deleteConfirmation !== `DELETE ${deleteTarget.name}`
                                || !deleteAcknowledged
                                || deleting}
                            onClick={deleteCourse}
                        >
                            {t("staff.courses.deletePermanently")}
                        </Button>
                    </AlertDialogFooter>
                </AlertDialogPopup>
            </AlertDialog>
        </div>
    )
}

function CourseForm({draft, saving, submitLabel, onChange, onSubmit}: {
    draft: CourseDraft
    saving: boolean
    submitLabel: string
    onChange: (draft: CourseDraft) => void
    onSubmit: (event: React.SubmitEvent<HTMLFormElement>) => void
}) {
    const {t} = useLocale()
    const color = normalizeColor(draft.color)
    const valid = Boolean(draft.name.trim())
        && draft.name.trim().length <= 32
        && draft.description.trim().length <= 128
        && /^[0-9A-Fa-f]{6}$/.test(draft.color)

    return (
        <Form className="contents" onSubmit={onSubmit}>
            <DialogPanel className="grid gap-4">
                <Field>
                    <FieldLabel>{t("staff.courses.name")}</FieldLabel>
                    <Input
                        value={draft.name}
                        onChange={(event) => onChange({...draft, name: event.target.value})}
                        placeholder={t("staff.courses.namePlaceholder")}
                        maxLength={32}
                        required
                        autoFocus
                    />
                </Field>
                <Field>
                    <div className="flex items-center justify-between gap-3">
                        <FieldLabel>{t("staff.courses.courseDescription")}</FieldLabel>
                        <span className="text-xs tabular-nums text-muted-foreground">{draft.description.length}/128</span>
                    </div>
                    <Textarea
                        value={draft.description}
                        onChange={(event) => onChange({...draft, description: event.target.value})}
                        placeholder={t("staff.courses.descriptionPlaceholder")}
                        maxLength={128}
                    />
                    <FieldDescription>{t("staff.courses.descriptionHelp")}</FieldDescription>
                </Field>
                <Field>
                    <FieldLabel>{t("staff.courses.color")}</FieldLabel>
                    <div className="flex items-center gap-2">
                        <Input
                            className="h-9 w-14 cursor-pointer p-1"
                            type="color"
                            value={`#${color}`}
                            onChange={(event) => onChange({...draft, color: event.target.value.slice(1).toUpperCase()})}
                        />
                        <div className="relative flex-1">
                            <span className="pointer-events-none absolute inset-y-0 left-3 z-10 flex items-center text-muted-foreground">#</span>
                            <Input
                                className="pl-7 font-mono uppercase"
                                value={draft.color}
                                onChange={(event) => onChange({...draft, color: event.target.value.replace(/^#/, "").slice(0, 6)})}
                                pattern="[0-9A-Fa-f]{6}"
                                maxLength={6}
                                required
                            />
                        </div>
                    </div>
                </Field>
                <div className="flex items-center gap-3 rounded-lg border bg-muted/25 p-3">
                    <span className="flex size-9 items-center justify-center rounded-lg text-white" style={{backgroundColor: `#${color}`}}>
                        <BookOpen className="size-4" />
                    </span>
                    <div className="min-w-0">
                        <p className="truncate text-sm font-medium">{draft.name.trim() || t("staff.courses.preview")}</p>
                        <p className="truncate text-xs text-muted-foreground">{draft.description.trim() || t("staff.courses.previewNoDescription")}</p>
                    </div>
                </div>
            </DialogPanel>
            <DialogFooter>
                <DialogClose render={<Button variant="outline" disabled={saving} />}>{t("staff.courses.cancel")}</DialogClose>
                <Button type="submit" loading={saving} disabled={!valid}>{submitLabel}</Button>
            </DialogFooter>
        </Form>
    )
}

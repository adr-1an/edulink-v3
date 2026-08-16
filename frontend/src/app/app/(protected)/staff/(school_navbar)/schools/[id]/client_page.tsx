"use client"

import PageTitle from "@/components/app/page_title"
import LocalDateTime from "@/components/local-date-time"
import React, {useEffect, useState} from "react"
import Link from "next/link"
import {useRouter} from "next/navigation"
import ReactCountryFlag from "react-country-flag"
import {toast} from "sonner"
import {Badge} from "@/components/ui/badge"
import {Card, CardDescription, CardFooter, CardHeader, CardTitle} from "@/components/ui/card"
import {Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle} from "@/components/ui/empty"
import {GraduationCap, LayoutGrid, LayoutList, MapPin, Pencil, Plus, Search, Settings, School as SchoolIcon, Trash2, TriangleAlert, X} from "lucide-react"
import {Button} from "@/components/ui/button"
import {Checkbox} from "@/components/ui/checkbox"
import {Input} from "@/components/ui/input"
import {Field, FieldDescription, FieldLabel} from "@/components/ui/field"
import {Form} from "@/components/ui/form"
import {
    Dialog, DialogClose, DialogDescription, DialogFooter, DialogHeader,
    DialogPanel, DialogPopup, DialogTitle,
} from "@/components/ui/dialog"
import {Select, SelectItem, SelectPopup, SelectTrigger, SelectValue} from "@/components/ui/select"
import {Switch} from "@/components/ui/switch"
import {Tabs, TabsList, TabsPanel, TabsTab} from "@/components/ui/tabs"
import {
    handleCreateGrade, handleDeleteGrade, handleUpdateGrade,
} from "@/app/app/(protected)/staff/(school_navbar)/schools/[id]/actions"
import {
    AlertDialog, AlertDialogClose, AlertDialogDescription, AlertDialogFooter,
    AlertDialogHeader, AlertDialogPopup, AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import {hasSchoolPermission, type SchoolAccess} from "@/lib/school_access"
import {rememberSchoolAccess} from "@/lib/school_navigation"
import {useLocale} from "@/i18n/provider"
import {type Locale} from "@/i18n/config"

interface Grade {
    id: string
    academicYearId: string
    name: string
    level: number
    createdAt: string
}

interface School {
    id: string
    name: string
    regionCode: string
    grades: Grade[] | null
    createdAt: string
    updatedAt: string
}

interface AcademicYear {
    id: string
    startYear: number
    endYear: number
    isActive: boolean
}

interface Props {
    school: School
    regionName?: string
    academicYears: AcademicYear[]
    access: SchoolAccess
    canListGrades: boolean
    canListAcademicYears: boolean
}

type GradeView = "grid" | "list"
type GradeSort = "level" | "az" | "za"

function gradeNameToPattern(grade: Grade) {
    const level = String(grade.level)
    const position = grade.name.lastIndexOf(level)

    if (position === -1) return `${grade.name} {level}`
    return `${grade.name.slice(0, position)}{level}${grade.name.slice(position + level.length)}`
}

function ordinalSuffix(level: number, locale: Locale) {
    if (locale === "pl") return "."

    const lastTwoDigits = Math.abs(level) % 100
    if (lastTwoDigits >= 11 && lastTwoDigits <= 13) return "th"

    if (Math.abs(level) % 10 === 1) return "st"
    if (Math.abs(level) % 10 === 2) return "nd"
    if (Math.abs(level) % 10 === 3) return "rd"
    return "th"
}

function displayGradeName(pattern: string, level: string) {
    return pattern.replace("{level}", level).replace(/\s+/g, " ").trim()
}

type GradeNameLayout = "label-first" | "level-first"

function readGradeNameFormat(pattern: string, level: string) {
    const [rawBefore = "", rawAfter = ""] = pattern.split("{level}", 2)
    const before = rawBefore.endsWith(" ") ? rawBefore.slice(0, -1) : rawBefore
    const ordinalMatch = rawAfter.match(/^(st|nd|rd|th|\.)(?:\s|$)/)
    let afterWithoutOrdinal = ordinalMatch ? rawAfter.slice(ordinalMatch[1].length) : rawAfter
    if (afterWithoutOrdinal.startsWith(" ")) afterWithoutOrdinal = afterWithoutOrdinal.slice(1)
    const hasBefore = before.trim().length > 0
    const hasAfter = afterWithoutOrdinal.trim().length > 0

    if (hasBefore && !hasAfter) {
        return {label: before, layout: "label-first" as const, ordinal: Boolean(ordinalMatch), custom: false}
    }
    if (!hasBefore && hasAfter) {
        return {label: afterWithoutOrdinal, layout: "level-first" as const, ordinal: Boolean(ordinalMatch), custom: false}
    }

    return {
        label: before || afterWithoutOrdinal,
        layout: (hasBefore ? "label-first" : "level-first") as GradeNameLayout,
        ordinal: Boolean(ordinalMatch),
        custom: hasBefore && hasAfter,
        currentName: displayGradeName(pattern, level),
    }
}

function buildGradeNameFormat(label: string, layout: GradeNameLayout, ordinal: boolean, level: number, locale: Locale) {
    const levelPart = `{level}${ordinal ? ordinalSuffix(level, locale) : ""}`
    return layout === "label-first" ? `${label} ${levelPart}` : `${levelPart} ${label}`
}

function normalizeGradeNamePattern(pattern: string) {
    return pattern.replace(/\s+/g, " ").trim()
}

export default function ClientPage({school, regionName, academicYears, access, canListGrades, canListAcademicYears}: Props) {
    const router = useRouter()
    const {t} = useLocale()
    const [creatingGrade, setCreatingGrade] = useState(false)
    const [editingGrade, setEditingGrade] = useState<Grade | null>(null)
    const [deletingGrade, setDeletingGrade] = useState<Grade | null>(null)
    const [gradeDeleteConfirmation, setGradeDeleteConfirmation] = useState("")
    const [gradeDeleteAcknowledged, setGradeDeleteAcknowledged] = useState(false)
    const [savingGrade, setSavingGrade] = useState(false)
    const [gradeName, setGradeName] = useState(t("staff.dashboard.patternPlaceholder"))
    const [gradeLevel, setGradeLevel] = useState("1")
    const [gradeView, setGradeView] = useState<GradeView>("grid")
    const [gradeSort, setGradeSort] = useState<GradeSort>("level")
    const [gradeQuery, setGradeQuery] = useState("")
    const allGrades = school.grades ?? []
    const normalizedGradeQuery = gradeQuery.trim().toLocaleLowerCase()
    const grades = allGrades.filter((grade) => !normalizedGradeQuery
        || grade.name.toLocaleLowerCase().includes(normalizedGradeQuery)
        || String(grade.level).includes(normalizedGradeQuery)
    ).sort((a, b) => {
        if (gradeSort === "az") return a.name.localeCompare(b.name) || a.level - b.level
        if (gradeSort === "za") return b.name.localeCompare(a.name) || a.level - b.level
        return a.level - b.level || a.name.localeCompare(b.name)
    })
    const activeYear = academicYears.find((year) => year.isActive)
    const canCreateGrade = hasSchoolPermission(access, "grade.create")
    const canUpdateGrade = hasSchoolPermission(access, "grade.update")
    const canDeleteGrade = hasSchoolPermission(access, "grade.delete")
    const canListCourses = hasSchoolPermission(access, "course.list")
    const canManageAcademicYears = hasSchoolPermission(access, "academicYear.create")
        || hasSchoolPermission(access, "academicYear.toggleActive")

    useEffect(() => {
        rememberSchoolAccess(school.id, access)
    }, [school.id, access])

    function openCreateGrade() {
        setGradeName(t("staff.dashboard.patternPlaceholder"))
        setGradeLevel("1")
        setCreatingGrade(true)
    }

    function openEditGrade(grade: Grade) {
        setGradeName(gradeNameToPattern(grade))
        setGradeLevel(String(grade.level))
        setEditingGrade(grade)
    }

    async function createGrade(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        if (!activeYear) return toast.error(t("staff.dashboard.activeYearRequired"))

        setSavingGrade(true)
        const res = await handleCreateGrade(activeYear.id, normalizeGradeNamePattern(gradeName), Number(gradeLevel))
        setSavingGrade(false)
        if (!res.ok) return toast.error(res.message)

        setCreatingGrade(false)
        toast.success(t("staff.dashboard.createdToast"))
        router.refresh()
    }

    async function updateGrade(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        if (!editingGrade) return

        setSavingGrade(true)
        const res = await handleUpdateGrade(editingGrade.id, normalizeGradeNamePattern(gradeName), Number(gradeLevel))
        setSavingGrade(false)
        if (!res.ok) return toast.error(res.message)

        setEditingGrade(null)
        toast.success(t("staff.dashboard.updatedToast"))
        router.refresh()
    }

    async function deleteGrade() {
        if (!deletingGrade) return
        const requiredConfirmation = `DELETE ${deletingGrade.name}`
        if (gradeDeleteConfirmation !== requiredConfirmation || !gradeDeleteAcknowledged) return

        setSavingGrade(true)
        const res = await handleDeleteGrade(deletingGrade.id)
        setSavingGrade(false)
        if (!res.ok) return toast.error(res.message)

        setDeletingGrade(null)
        setGradeDeleteConfirmation("")
        setGradeDeleteAcknowledged(false)
        toast.success(t("staff.dashboard.deletedToast"))
        router.refresh()
    }

    return (
        <div className="space-y-6">
            <section className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <div className="flex items-center gap-3">
                    {school.regionCode && (
                        <ReactCountryFlag
                            svg
                            countryCode={school.regionCode}
                            style={{width: "2.5rem", height: "2.5rem", borderRadius: "0.5rem"}}
                        />
                    )}
                    <div>
                        <PageTitle centered={false}>{school.name}</PageTitle>
                        <p className="text-sm text-muted-foreground">
                            {regionName ?? school.regionCode ?? t("staff.dashboard.regionMissing")}
                        </p>
                    </div>
                </div>
            </section>

            <section className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                {canListGrades && <Card>
                    <CardHeader>
                        <CardDescription className="flex items-center gap-2"><GraduationCap /> {t("staff.dashboard.activeGrades")}</CardDescription>
                        <CardTitle className="text-3xl">{allGrades.length}</CardTitle>
                    </CardHeader>
                </Card>}
                <Card>
                    <CardHeader>
                        <CardDescription className="flex items-center gap-2"><MapPin /> {t("staff.dashboard.region")}</CardDescription>
                        <CardTitle>{regionName ?? school.regionCode ?? t("staff.dashboard.regionMissing")}</CardTitle>
                    </CardHeader>
                </Card>
                <Card>
                    <CardHeader>
                        <CardDescription className="flex items-center gap-2"><SchoolIcon /> {t("staff.dashboard.created")}</CardDescription>
                        <CardTitle><LocalDateTime value={school.createdAt} precision="date" /></CardTitle>
                    </CardHeader>
                </Card>
            </section>

            {canListGrades && <section>
                <div className="mb-3 flex flex-wrap items-end justify-between gap-3">
                    <div>
                        <h2 className="text-2xl font-semibold">{t("staff.dashboard.activeGrades")}</h2>
                        <p className="text-sm text-muted-foreground">{t("staff.dashboard.gradesDescription")}</p>
                    </div>
                    <div className="flex flex-wrap items-center justify-end gap-2">
                        <div className="relative min-w-52 flex-1 sm:flex-none">
                            <Search className="pointer-events-none absolute top-1/2 left-3 z-10 size-4 -translate-y-1/2 text-muted-foreground" />
                            <Input
                                className="pr-8 pl-9 sm:w-56"
                                value={gradeQuery}
                                onChange={(event) => setGradeQuery(event.target.value)}
                                placeholder={t("staff.dashboard.searchPlaceholder")}
                                aria-label={t("staff.dashboard.search")}
                            />
                            {gradeQuery && (
                                <Button
                                    className="absolute top-1/2 right-1 -translate-y-1/2"
                                    size="icon-xs"
                                    variant="ghost"
                                    aria-label={t("staff.dashboard.clearSearch")}
                                    onClick={() => setGradeQuery("")}
                                >
                                    <X />
                                </Button>
                            )}
                        </div>
                        <Select
                            items={[
                                {label: t("staff.dashboard.sortLevel"), value: "level"},
                                {label: "A-Z", value: "az"},
                                {label: "Z-A", value: "za"},
                            ]}
                            value={gradeSort}
                            onValueChange={(value) => setGradeSort((value ?? "level") as GradeSort)}
                        >
                            <SelectTrigger className="w-36">
                                <SelectValue placeholder={t("staff.dashboard.sortBy")} />
                            </SelectTrigger>

                            <SelectPopup>
                                <SelectItem value="level">{t("staff.dashboard.sortLevel")}</SelectItem>
                                <SelectItem value="az">A-Z</SelectItem>
                                <SelectItem value="za">Z-A</SelectItem>
                            </SelectPopup>
                        </Select>

                        <div className="flex rounded-lg border bg-background p-0.5">
                            <Button
                                size="icon-sm"
                                variant={gradeView === "grid" ? "secondary" : "ghost"}
                                aria-label={t("staff.dashboard.gridView")}
                                aria-pressed={gradeView === "grid"}
                                onClick={() => setGradeView("grid")}
                            >
                                <LayoutGrid />
                            </Button>
                            <Button
                                size="icon-sm"
                                variant={gradeView === "list" ? "secondary" : "ghost"}
                                aria-label={t("staff.dashboard.listView")}
                                aria-pressed={gradeView === "list"}
                                onClick={() => setGradeView("list")}
                            >
                                <LayoutList />
                            </Button>
                        </div>

                        {canCreateGrade && <Button onClick={openCreateGrade} disabled={!activeYear}>
                            <Plus /> {t("staff.dashboard.createGrade")}
                        </Button>}
                    </div>
                </div>

                {canListAcademicYears && !activeYear && (
                    <div className="mb-4 flex flex-col gap-4 rounded-2xl border border-warning/35 bg-warning/[0.07] p-4 text-warning-foreground sm:flex-row sm:items-center" role="alert">
                        <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-warning/15">
                            <TriangleAlert className="size-5" />
                        </span>
                        <div className="min-w-0 flex-1">
                            <h3 className="font-semibold">{t("staff.dashboard.academicYearWarningTitle")}</h3>
                            <p className="mt-1 text-sm leading-6 text-muted-foreground">
                                {t(canManageAcademicYears
                                    ? "staff.dashboard.academicYearWarningDescription"
                                    : "staff.dashboard.academicYearWarningNoAccess")}
                            </p>
                        </div>
                        {canManageAcademicYears && (
                            <Button className="shrink-0" variant="outline" render={<Link href={`/app/staff/schools/${school.id}/settings`} />}>
                                <Settings /> {t("staff.dashboard.openSchoolSettings")}
                            </Button>
                        )}
                    </div>
                )}

                {grades.length > 0 ? (
                    <div className={gradeView === "grid" ? "grid gap-3 sm:grid-cols-2 lg:grid-cols-3" : "space-y-2"}>
                        {grades.map((grade) => (
                            <Card key={grade.id} className="group/grade overflow-hidden transition-[border-color,box-shadow,transform] hover:-translate-y-0.5 hover:border-primary/25 hover:shadow-md/10">
                                {canListCourses && <Link
                                    href={`/app/staff/grades/${grade.id}`}
                                    aria-label={t("staff.dashboard.openCourses", {name: grade.name})}
                                    className="absolute inset-0 z-[1] rounded-2xl focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset"
                                />}
                                <CardHeader className={`pointer-events-none relative z-10 ${gradeView === "list" ? "grid-cols-1 gap-4 sm:grid-cols-[1fr_auto] sm:items-center" : ""}`}>
                                    <div className="flex items-start justify-between gap-3">
                                        <div>
                                            <CardTitle>{grade.name}</CardTitle>
                                            {gradeView === "list" && (
                                                <CardDescription>{t("staff.dashboard.created")} <LocalDateTime value={grade.createdAt} precision="date" /></CardDescription>
                                            )}
                                        </div>
                                        <Badge variant="outline">{t("staff.dashboard.level", {level: grade.level})}</Badge>
                                    </div>
                                    {gradeView === "list" && (canUpdateGrade || canDeleteGrade) && (
                                        <div className="pointer-events-auto flex shrink-0 gap-2">
                                            {canUpdateGrade && <Button variant="outline" onClick={() => openEditGrade(grade)}>
                                                <Pencil /> {t("staff.dashboard.edit")}
                                            </Button>}
                                            {canDeleteGrade && <Button variant="destructive-outline" onClick={() => setDeletingGrade(grade)}>
                                                <Trash2 />
                                            </Button>}
                                        </div>
                                    )}
                                </CardHeader>
                                {gradeView === "grid" && (canUpdateGrade || canDeleteGrade) && <CardFooter className="pointer-events-none relative z-10 justify-end gap-2">
                                    {canUpdateGrade && <Button className="pointer-events-auto" variant="outline" onClick={() => openEditGrade(grade)}>
                                        <Pencil /> {t("staff.dashboard.edit")}
                                    </Button>}
                                    {canDeleteGrade && <Button className="pointer-events-auto" variant="destructive-outline" onClick={() => setDeletingGrade(grade)}>
                                        <Trash2 />
                                    </Button>}
                                </CardFooter>}
                            </Card>
                        ))}
                    </div>
                ) : normalizedGradeQuery && allGrades.length > 0 ? (
                    <Card>
                        <Empty>
                            <EmptyHeader>
                                <EmptyMedia variant="icon"><Search /></EmptyMedia>
                                <EmptyTitle>{t("staff.dashboard.noGradesFound")}</EmptyTitle>
                                <EmptyDescription>{t("staff.dashboard.noGradesMatch", {query: gradeQuery.trim()})}</EmptyDescription>
                            </EmptyHeader>
                            <Button size="sm" variant="outline" onClick={() => setGradeQuery("")}>{t("staff.schools.clear")}</Button>
                        </Empty>
                    </Card>
                ) : (
                    <Card>
                        <Empty>
                            <EmptyHeader>
                                <EmptyMedia variant="icon"><GraduationCap /></EmptyMedia>
                                <EmptyTitle>{t("staff.dashboard.noActiveGrades")}</EmptyTitle>
                                <EmptyDescription>{t("staff.dashboard.noActiveGradesDescription")}</EmptyDescription>
                            </EmptyHeader>
                        </Empty>
                    </Card>
                )}
            </section>}

            <Dialog open={creatingGrade} onOpenChange={(open) => {
                if (!savingGrade) setCreatingGrade(open)
            }}>
                <DialogPopup className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle>{t("staff.dashboard.createTitle")}</DialogTitle>
                        <DialogDescription>{t("staff.dashboard.createDescription")}</DialogDescription>
                    </DialogHeader>
                    <GradeForm
                        name={gradeName}
                        level={gradeLevel}
                        saving={savingGrade}
                        submitLabel={t("staff.dashboard.createGrade")}
                        onNameChange={setGradeName}
                        onLevelChange={setGradeLevel}
                        onSubmit={createGrade}
                    />
                </DialogPopup>
            </Dialog>

            <Dialog open={editingGrade !== null} onOpenChange={(open) => {
                if (!open && !savingGrade) setEditingGrade(null)
            }}>
                <DialogPopup className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle>{t("staff.dashboard.editTitle")}</DialogTitle>
                        <DialogDescription>{t("staff.dashboard.editDescription")}</DialogDescription>
                    </DialogHeader>
                    <GradeForm
                        name={gradeName}
                        level={gradeLevel}
                        saving={savingGrade}
                        submitLabel={t("staff.dashboard.save")}
                        onNameChange={setGradeName}
                        onLevelChange={setGradeLevel}
                        onSubmit={updateGrade}
                    />
                </DialogPopup>
            </Dialog>

            <AlertDialog open={deletingGrade !== null} onOpenChange={(open) => {
                if (!open && !savingGrade) {
                    setDeletingGrade(null)
                    setGradeDeleteConfirmation("")
                    setGradeDeleteAcknowledged(false)
                }
            }}>
                <AlertDialogPopup className="border-destructive/30 sm:max-w-xl">
                    <AlertDialogHeader>
                        <div className="mb-2 flex size-12 items-center justify-center rounded-2xl bg-destructive/10 text-destructive">
                            <TriangleAlert className="size-6" />
                        </div>
                        <AlertDialogTitle className="text-destructive">{t("staff.dashboard.deleteTitle")}</AlertDialogTitle>
                        <AlertDialogDescription className="space-y-3 text-left">
                            <span className="block">{t("staff.dashboard.deleteDescription", {name: deletingGrade?.name ?? ""})}</span>
                            <span className="block font-medium text-destructive">{t("staff.dashboard.deleteIrreversible")}</span>
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <div className="grid gap-4 px-6 pb-2">
                        <Field>
                            <FieldLabel htmlFor="grade-delete-confirmation">
                                {t("staff.dashboard.typeToDelete", {value: deletingGrade ? `DELETE ${deletingGrade.name}` : ""})}
                            </FieldLabel>
                            <Input
                                id="grade-delete-confirmation"
                                value={gradeDeleteConfirmation}
                                onChange={(event) => setGradeDeleteConfirmation(event.target.value)}
                                placeholder={deletingGrade ? `DELETE ${deletingGrade.name}` : ""}
                                autoComplete="off"
                                spellCheck={false}
                                disabled={savingGrade}
                                autoFocus
                            />
                        </Field>
                        <label className="flex items-start gap-3 text-sm">
                            <Checkbox
                                checked={gradeDeleteAcknowledged}
                                onCheckedChange={(value) => setGradeDeleteAcknowledged(value)}
                                disabled={savingGrade}
                            />
                            <span>{t("staff.dashboard.deleteAcknowledge")}</span>
                        </label>
                    </div>
                    <AlertDialogFooter>
                        <AlertDialogClose render={<Button variant="outline" disabled={savingGrade} />}>{t("staff.dashboard.cancel")}</AlertDialogClose>
                        <Button
                            variant="destructive"
                            loading={savingGrade}
                            disabled={!deletingGrade
                                || gradeDeleteConfirmation !== `DELETE ${deletingGrade.name}`
                                || !gradeDeleteAcknowledged
                                || savingGrade}
                            onClick={deleteGrade}
                        >
                            {t("staff.dashboard.deletePermanently")}
                        </Button>
                    </AlertDialogFooter>
                </AlertDialogPopup>
            </AlertDialog>

        </div>
    )
}

function GradeForm({name, level, saving, submitLabel, onNameChange, onLevelChange, onSubmit}: {
    name: string
    level: string
    saving: boolean
    submitLabel: string
    onNameChange: (value: string) => void
    onLevelChange: (value: string) => void
    onSubmit: (event: React.SubmitEvent<HTMLFormElement>) => void
}) {
    const {locale, t} = useLocale()
    const parsedLevel = Number(level)
    const validLevel = /^\d+$/.test(level) && parsedLevel >= 0 && parsedLevel <= 20
    const format = readGradeNameFormat(name, level)
    const visibleName = displayGradeName(name, level)
    const levelVariableCount = name.split("{level}").length - 1
    const validName = levelVariableCount === 1 && name.length >= 7 && name.length <= 32

    function changeLevel(nextLevel: string) {
        if (!format.custom && /^\d+$/.test(nextLevel)) {
            onNameChange(buildGradeNameFormat(format.label, format.layout, format.ordinal, Number(nextLevel), locale))
        }
        onLevelChange(nextLevel)
    }

    function changeFormat(changes: {label?: string; layout?: GradeNameLayout; ordinal?: boolean}) {
        onNameChange(buildGradeNameFormat(
            changes.label ?? format.label,
            changes.layout ?? format.layout,
            changes.ordinal ?? format.ordinal,
            validLevel ? parsedLevel : 0,
            locale,
        ))
    }

    return (
        <Form className="contents" onSubmit={onSubmit}>
            <DialogPanel>
                <Tabs defaultValue={format.custom ? "manual" : "builder"} className="gap-4">
                    <TabsList className="w-full">
                        <TabsTab className="flex-1" value="builder">{t("staff.dashboard.visualBuilder")}</TabsTab>
                        <TabsTab className="flex-1" value="manual">{t("staff.dashboard.manualEntry")}</TabsTab>
                    </TabsList>

                    <Field>
                        <FieldLabel>{t("staff.dashboard.gradeLevel")}</FieldLabel>
                        <Input type="number" min={0} max={20} value={level} onChange={(event) => changeLevel(event.target.value)} required autoFocus />
                        <FieldDescription>{t("staff.dashboard.levelHelp")}</FieldDescription>
                    </Field>

                    <TabsPanel value="builder">
                    <div>
                    <p className="text-sm font-medium">{t("staff.dashboard.nameFormat")}</p>
                    <p className="mt-1 text-xs text-muted-foreground">{t("staff.dashboard.nameFormatHelp")}</p>

                    {format.custom && (
                        <div className="mt-3 rounded-lg border border-warning/30 bg-warning/8 p-3 text-xs text-warning-foreground">
                            {t("staff.dashboard.customNameNotice", {name: format.currentName ?? visibleName})}
                        </div>
                    )}

                    <Field className="mt-3">
                        <FieldLabel>{t("staff.dashboard.label")}</FieldLabel>
                        <Input
                            value={format.label}
                            onChange={(event) => changeFormat({label: event.target.value})}
                            placeholder={t("staff.dashboard.labelPlaceholder")}
                            maxLength={20}
                        />
                        <FieldDescription>{t("staff.dashboard.labelHelp")}</FieldDescription>
                    </Field>

                    <div className="mt-4">
                        <p className="mb-2 text-xs font-medium">{t("staff.dashboard.chooseAppearance")}</p>
                        <div className="grid grid-cols-2 gap-2">
                            <Button
                                type="button"
                                className="h-12"
                                variant={!format.custom && format.layout === "label-first" ? "secondary" : "outline"}
                                disabled={!validLevel}
                                onClick={() => changeFormat({layout: "label-first"})}
                            >
                                {format.label || t("staff.dashboard.labelPlaceholder")} {validLevel ? `${level}${format.ordinal ? ordinalSuffix(parsedLevel, locale) : ""}` : "—"}
                            </Button>
                            <Button
                                type="button"
                                className="h-12"
                                variant={!format.custom && format.layout === "level-first" ? "secondary" : "outline"}
                                disabled={!validLevel}
                                onClick={() => changeFormat({layout: "level-first"})}
                            >
                                {validLevel ? `${level}${format.ordinal ? ordinalSuffix(parsedLevel, locale) : ""}` : "—"} {format.label || t("staff.dashboard.labelPlaceholder")}
                            </Button>
                        </div>
                    </div>

                    <div className="mt-4 flex items-center justify-between gap-4 rounded-xl border p-3">
                        <div>
                            <p className="text-sm font-medium">{t("staff.dashboard.ordinal")}</p>
                            <p className="text-xs text-muted-foreground">
                                {t("staff.dashboard.ordinalHelp", {
                                    ordinal: validLevel ? `${level}${ordinalSuffix(parsedLevel, locale)}` : locale === "pl" ? "2." : "2nd",
                                    plain: validLevel ? level : "2",
                                })}
                            </p>
                        </div>
                        <Switch
                            checked={format.ordinal}
                            onCheckedChange={(checked) => changeFormat({ordinal: checked})}
                            disabled={!validLevel}
                            aria-label={t("staff.dashboard.ordinalAria")}
                        />
                    </div>

                    <div className="mt-4 rounded-xl border bg-muted/35 p-3">
                        <p className="text-[10px] font-semibold uppercase tracking-[0.14em] text-muted-foreground">{t("staff.dashboard.preview")}</p>
                        <p className="mt-2 truncate font-semibold">{visibleName || t("staff.dashboard.chooseFormat")}</p>
                    </div>

                    {!validName && (
                        <p className="mt-2 text-xs text-destructive">{t("staff.dashboard.nameTooLong")}</p>
                    )}
                    </div>
                    </TabsPanel>

                    <TabsPanel value="manual" className="grid gap-3">
                        <Field>
                            <FieldLabel>{t("staff.dashboard.namePattern")}</FieldLabel>
                            <Input
                                className="font-mono"
                                value={name}
                                onChange={(event) => onNameChange(event.target.value)}
                                minLength={7}
                                maxLength={32}
                                placeholder={t("staff.dashboard.patternPlaceholder")}
                                required
                            />
                            <FieldDescription>
                                {t("staff.dashboard.patternHelp", {variable: "{level}"})}
                            </FieldDescription>
                            {levelVariableCount === 0 && (
                                <p className="text-xs text-destructive">{t("staff.dashboard.patternMissing", {variable: "{level}"})}</p>
                            )}
                            {levelVariableCount > 1 && (
                                <p className="text-xs text-destructive">{t("staff.dashboard.patternRepeated", {variable: "{level}"})}</p>
                            )}
                            {levelVariableCount === 1 && !validName && (
                                <p className="text-xs text-destructive">{t("staff.dashboard.patternTooLong")}</p>
                            )}
                        </Field>

                        <div className="rounded-xl border bg-muted/35 p-3">
                            <p className="text-[10px] font-semibold uppercase tracking-[0.14em] text-muted-foreground">{t("staff.dashboard.preview")}</p>
                            <p className="mt-2 truncate font-semibold">{visibleName || t("staff.dashboard.enterPattern")}</p>
                        </div>
                    </TabsPanel>
                </Tabs>
            </DialogPanel>
            <DialogFooter>
                <DialogClose render={<Button type="button" variant="ghost" disabled={saving} />}>{t("staff.dashboard.cancel")}</DialogClose>
                <Button type="submit" loading={saving} disabled={!validName || !validLevel}>{submitLabel}</Button>
            </DialogFooter>
        </Form>
    )
}

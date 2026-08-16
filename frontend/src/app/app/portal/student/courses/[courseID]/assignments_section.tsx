"use client"

import {useDeferredValue, useMemo, useState} from "react"
import Link from "next/link"
import {
    AlertTriangle, ArrowUpDown, Award, CalendarClock, CheckCircle2, ClipboardList, Clock3,
    Eye, EyeOff, ExternalLink, Grid2X2, Link2, List, ListFilter, RotateCcw, Search, Send, X,
} from "lucide-react"
import {Badge} from "@/components/ui/badge"
import {Button} from "@/components/ui/button"
import {Card, CardContent, CardHeader, CardTitle} from "@/components/ui/card"
import {
    useBrowserTimeZone,
    useLocalDateKey,
    useLocalDateTimeFormatter,
} from "@/components/local-date-time"
import {Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle} from "@/components/ui/empty"
import {Input} from "@/components/ui/input"
import {Select, SelectItem, SelectPopup, SelectTrigger, SelectValue} from "@/components/ui/select"
import {useLocale} from "@/i18n/provider"
import {scoreAccent} from "@/lib/score"
import SubmissionDialog from "./submission_dialog"

export interface PortalSubmissionAttachment {
    id: string
    fileName: string
    fileSize: number
    contentType: string
    presignedUrl: string | null
}

export interface PortalCourseAssignment {
    id: string
    referencedPost: {
        id: string
        title: string
    } | null
    course?: {
        id: string
        name: string
    }
    title: string
    description: string
    dueDate: string | null
    submissionsEnabled: boolean
    submissionsCloseAt: string | null
    submission: {
        id: string
        status: "pending" | "submitted" | "returned"
        grade: {
            score: number
            notes: string | null
            gradedAt: string
            gradedBy: {
                id: string
                name: string
                email: string
                phone: string | null
            }
        } | null
        notes: string
        attachments: PortalSubmissionAttachment[]
        createdAt: string
    } | null
    createdAt: string
}

type AssignmentSort = "due-asc" | "due-desc" | "created-desc" | "created-asc"
type AssignmentFilter = "all" | "has-due" | "no-due" | "submissions" | "no-submissions" | "referenced"
type AssignmentView = "list" | "grid"
type SubmissionView = "not-submitted" | "submitted" | "graded" | "returned"

function relativeDueDate(value: string, referenceDate: string, browserTimeZone: boolean) {
    const dueDate = new Date(value)
    const today = new Date(`${referenceDate}T00:00:00${browserTimeZone ? "" : "Z"}`)
    if (Number.isNaN(dueDate.getTime()) || Number.isNaN(today.getTime())) return null

    const dueDay = browserTimeZone
        ? new Date(dueDate.getFullYear(), dueDate.getMonth(), dueDate.getDate()).getTime()
        : Date.UTC(dueDate.getUTCFullYear(), dueDate.getUTCMonth(), dueDate.getUTCDate())
    return Math.round((dueDay - today.getTime()) / 86_400_000)
}

function submissionsClosed(assignment: PortalCourseAssignment, referenceTime: string) {
    if (!assignment.submissionsEnabled) return true
    if (!assignment.submissionsCloseAt) return false

    const closeTime = new Date(assignment.submissionsCloseAt).getTime()
    const currentTime = new Date(referenceTime).getTime()
    return !Number.isNaN(closeTime) && !Number.isNaN(currentTime) && closeTime <= currentTime
}

export default function AssignmentsSection({assignments, referenceDate, referenceTime, showCourseFilter = false, showHeading = true}: {
    assignments: PortalCourseAssignment[]
    referenceDate: string
    referenceTime: string
    showCourseFilter?: boolean
    showHeading?: boolean
}) {
    const [query, setQuery] = useState("")
    const [sort, setSort] = useState<AssignmentSort>("due-asc")
    const [filter, setFilter] = useState<AssignmentFilter>("all")
    const [course, setCourse] = useState("all")
    const [view, setView] = useState<AssignmentView>("list")
    const [submissionView, setSubmissionView] = useState<SubmissionView>("not-submitted")
    const [showClosedPastDue, setShowClosedPastDue] = useState(false)
    const [submissionAssignment, setSubmissionAssignment] = useState<PortalCourseAssignment | null>(null)
    const {locale, t} = useLocale()
    const sortOptions = [
        {label: t("assignments.sort.dueSoonest"), value: "due-asc"},
        {label: t("assignments.sort.dueLatest"), value: "due-desc"},
        {label: t("assignments.sort.newest"), value: "created-desc"},
        {label: t("assignments.sort.oldest"), value: "created-asc"},
    ]
    const filterOptions = [
        {label: t("assignments.filter.all"), value: "all"},
        {label: t("assignments.filter.hasDue"), value: "has-due"},
        {label: t("assignments.filter.noDue"), value: "no-due"},
        {label: t("assignments.filter.submissions"), value: "submissions"},
        {label: t("assignments.filter.noSubmissions"), value: "no-submissions"},
        {label: t("assignments.filter.referenced"), value: "referenced"},
    ]
    const formatDate = useLocalDateTimeFormatter()
    const browserTimeZone = useBrowserTimeZone()
    const currentDate = useLocalDateKey(referenceDate)
    const deferredQuery = useDeferredValue(query)
    const hasActiveFilters = query !== "" || filter !== "all" || sort !== "due-asc" || course !== "all"
    const hasReturnedAssignments = assignments.some((assignment) => assignment.submission?.status === "returned")
    const courses = useMemo(() => {
        const unique = new Map<string, string>()
        for (const assignment of assignments) {
            if (assignment.course) unique.set(assignment.course.id, assignment.course.name)
        }
        return [...unique].map(([id, name]) => ({id, name})).sort((first, second) => first.name.localeCompare(second.name, locale))
    }, [assignments, locale])

    const visibleAssignments = useMemo(() => {
        const search = deferredQuery.trim().toLocaleLowerCase()
        const filtered = assignments.filter((assignment) => {
            if (course !== "all" && assignment.course?.id !== course) return false
            const matchesSearch = !search || [
                assignment.title,
                assignment.description,
                assignment.referencedPost?.title ?? "",
                assignment.course?.name ?? "",
            ].some((value) => value.toLocaleLowerCase().includes(search))
            if (!matchesSearch) return false

            switch (filter) {
            case "has-due":
                return assignment.dueDate !== null
            case "no-due":
                return assignment.dueDate === null
            case "submissions":
                return assignment.submissionsEnabled
            case "no-submissions":
                return !assignment.submissionsEnabled
            case "referenced":
                return assignment.referencedPost !== null
            default:
                return true
            }
        })

        return filtered.sort((first, second) => {
            if (sort.startsWith("created")) {
                const difference = new Date(first.createdAt).getTime() - new Date(second.createdAt).getTime()
                return sort === "created-asc" ? difference : -difference
            }

            if (!first.dueDate && !second.dueDate) {
                return new Date(second.createdAt).getTime() - new Date(first.createdAt).getTime()
            }
            if (!first.dueDate) return 1
            if (!second.dueDate) return -1
            const difference = new Date(first.dueDate).getTime() - new Date(second.dueDate).getTime()
            return sort === "due-asc" ? difference : -difference
        })
    }, [assignments, course, deferredQuery, filter, sort])
    const groupedAssignments = useMemo(() => {
        const graded = visibleAssignments.filter((assignment) => assignment.submission?.status === "submitted" && assignment.submission.grade !== null)
        const submitted = visibleAssignments.filter((assignment) => assignment.submission?.status === "submitted" && assignment.submission.grade === null)
        const returned = visibleAssignments.filter((assignment) => assignment.submission?.status === "returned")
        const notSubmitted = visibleAssignments.filter((assignment) => assignment.submission?.status !== "submitted" && assignment.submission?.status !== "returned")
        const selected = submissionView === "submitted"
            ? submitted
            : submissionView === "graded"
                ? graded
                : submissionView === "returned" ? returned : notSubmitted
        const current: PortalCourseAssignment[] = []
        const pastDue: PortalCourseAssignment[] = []

        for (const assignment of selected) {
            const days = assignment.dueDate
                ? relativeDueDate(assignment.dueDate, currentDate, browserTimeZone)
                : null
            if (submissionView === "not-submitted" && days !== null && days < 0) pastDue.push(assignment)
            else current.push(assignment)
        }

        const hiddenClosed = pastDue.filter((assignment) => (
            submissionsClosed(assignment, referenceTime)
        ))
        const hiddenClosedIDs = new Set(hiddenClosed.map(({id}) => id))
        const displayedPastDue = showClosedPastDue
            ? pastDue
            : pastDue.filter((assignment) => !hiddenClosedIDs.has(assignment.id))

        return {current, pastDue, displayedPastDue, hiddenClosed, submitted, graded, returned, notSubmitted, selected}
    }, [browserTimeZone, currentDate, referenceTime, showClosedPastDue, submissionView, visibleAssignments])
    const displayedAssignmentCount = groupedAssignments.current.length + groupedAssignments.displayedPastDue.length

    function reset() {
        setQuery("")
        setFilter("all")
        setSort("due-asc")
        setCourse("all")
        setSubmissionView("not-submitted")
        setShowClosedPastDue(false)
    }

    return (
        <section
            className="space-y-4"
            aria-labelledby={showHeading ? "course-assignments-title" : undefined}
            aria-label={showHeading ? undefined : t("assignments.title")}
        >
            <div className="flex flex-wrap items-end justify-between gap-3">
                {showHeading && (
                    <div>
                        <h2 className="text-xl font-semibold" id="course-assignments-title">{t("assignments.title")}</h2>
                        <p className="text-sm text-muted-foreground">{t("assignments.description")}</p>
                    </div>
                )}
                <div className={showHeading ? "ml-auto flex items-center gap-3" : "ml-auto flex w-full flex-wrap items-center justify-between gap-3"}>
                    {assignments.length > 0 && (
                        <p className="text-sm tabular-nums text-muted-foreground">
                            {t("assignments.showing", {visible: displayedAssignmentCount, total: assignments.length})}
                        </p>
                    )}
                    {assignments.length > 0 && (
                        <div className="flex items-center gap-1 rounded-xl border bg-card p-1" role="group" aria-label={t("assignments.view")}>
                            <Button
                                size="sm"
                                variant={view === "list" ? "secondary" : "ghost"}
                                aria-pressed={view === "list"}
                                onClick={() => setView("list")}
                            >
                                <List /> {t("assignments.list")}
                            </Button>
                            <Button
                                size="sm"
                                variant={view === "grid" ? "secondary" : "ghost"}
                                aria-pressed={view === "grid"}
                                onClick={() => setView("grid")}
                            >
                                <Grid2X2 /> {t("assignments.cards")}
                            </Button>
                        </div>
                    )}
                </div>
            </div>

            {assignments.length > 0 && (
                <div className="flex w-fit max-w-full flex-wrap items-center gap-1 rounded-xl border bg-muted/50 p-1" role="tablist" aria-label={t("assignments.statusTabs")}>
                    <Button
                        size="sm"
                        variant={submissionView === "not-submitted" ? "secondary" : "ghost"}
                        role="tab"
                        aria-selected={submissionView === "not-submitted"}
                        onClick={() => {
                            setSubmissionView("not-submitted")
                            setShowClosedPastDue(false)
                        }}
                    >
                        <ClipboardList /> {t("assignments.toSubmit")}
                        <Badge variant="secondary">{groupedAssignments.notSubmitted.length}</Badge>
                    </Button>
                    <Button
                        size="sm"
                        variant={submissionView === "submitted" ? "secondary" : "ghost"}
                        role="tab"
                        aria-selected={submissionView === "submitted"}
                        onClick={() => setSubmissionView("submitted")}
                    >
                        <CheckCircle2 /> {t("assignments.submittedTab")}
                        <Badge variant="secondary">{groupedAssignments.submitted.length}</Badge>
                    </Button>
                    <Button
                        size="sm"
                        variant={submissionView === "graded" ? "secondary" : "ghost"}
                        role="tab"
                        aria-selected={submissionView === "graded"}
                        onClick={() => setSubmissionView("graded")}
                    >
                        <Award /> {t("assignments.gradedTab")}
                        <Badge variant="secondary">{groupedAssignments.graded.length}</Badge>
                    </Button>
                    {hasReturnedAssignments && (
                        <Button
                            size="sm"
                            variant={submissionView === "returned" ? "secondary" : "ghost"}
                            role="tab"
                            aria-selected={submissionView === "returned"}
                            onClick={() => {
                                setSubmissionView("returned")
                                setShowClosedPastDue(false)
                            }}
                        >
                            <RotateCcw /> {t("assignments.returnedTab")}
                            <Badge variant="warning">{groupedAssignments.returned.length}</Badge>
                        </Button>
                    )}
                </div>
            )}

            {assignments.length > 0 && (
                <div className={`grid gap-2 rounded-2xl border bg-card p-3 ${showCourseFilter ? "md:grid-cols-2 xl:grid-cols-[minmax(12rem,1fr)_12rem_12rem_13rem_auto]" : "md:grid-cols-[minmax(12rem,1fr)_12rem_13rem_auto]"}`}>
                    <div className="relative">
                        <Search className="pointer-events-none absolute left-3 top-1/2 z-10 size-4 -translate-y-1/2 text-muted-foreground" />
                        <Input className="pl-9 pr-9" value={query} onChange={(event) => setQuery(event.target.value)} placeholder={t("assignments.searchPlaceholder")} aria-label={t("assignments.searchLabel")} />
                        {query && (
                            <Button className="absolute right-1.5 top-1/2 -translate-y-1/2" size="icon-xs" variant="ghost" aria-label={t("assignments.clearSearch")} onClick={() => setQuery("")}>
                                <X />
                            </Button>
                        )}
                    </div>
                    {showCourseFilter && (
                        <Select
                            items={[{label: t("assignments.allCourses"), value: "all"}, ...courses.map((item) => ({label: item.name, value: item.id}))]}
                            value={course}
                            onValueChange={(value) => setCourse(value ?? "all")}
                        >
                            <SelectTrigger aria-label={t("assignments.filterCourse")}><SelectValue /></SelectTrigger>
                            <SelectPopup>
                                <SelectItem value="all">{t("assignments.allCourses")}</SelectItem>
                                {courses.map((item) => <SelectItem value={item.id} key={item.id}>{item.name}</SelectItem>)}
                            </SelectPopup>
                        </Select>
                    )}
                    <Select items={filterOptions} value={filter} onValueChange={(value) => setFilter((value ?? "all") as AssignmentFilter)}>
                        <SelectTrigger aria-label={t("assignments.filter")}><ListFilter /><SelectValue /></SelectTrigger>
                        <SelectPopup>
                            {filterOptions.map((option) => <SelectItem value={option.value} key={option.value}>{option.label}</SelectItem>)}
                        </SelectPopup>
                    </Select>
                    <Select items={sortOptions} value={sort} onValueChange={(value) => setSort((value ?? "due-asc") as AssignmentSort)}>
                        <SelectTrigger aria-label={t("assignments.sort")}><ArrowUpDown /><SelectValue /></SelectTrigger>
                        <SelectPopup>
                            {sortOptions.map((option) => <SelectItem value={option.value} key={option.value}>{option.label}</SelectItem>)}
                        </SelectPopup>
                    </Select>
                    <Button variant="ghost" disabled={!hasActiveFilters} onClick={reset}>{t("assignments.reset")}</Button>
                </div>
            )}

            {assignments.length === 0 ? (
                <Card>
                    <Empty>
                        <EmptyHeader>
                            <EmptyMedia variant="icon"><ClipboardList /></EmptyMedia>
                            <EmptyTitle>{t("assignments.empty")}</EmptyTitle>
                            <EmptyDescription>{t("assignments.emptyDescription")}</EmptyDescription>
                        </EmptyHeader>
                    </Empty>
                </Card>
            ) : visibleAssignments.length === 0 ? (
                <Card>
                    <Empty>
                        <EmptyHeader>
                            <EmptyMedia variant="icon"><Search /></EmptyMedia>
                            <EmptyTitle>{t("assignments.noMatches")}</EmptyTitle>
                            <EmptyDescription>{t("assignments.noMatchesDescription")}</EmptyDescription>
                        </EmptyHeader>
                        <Button variant="outline" onClick={reset}>{t("assignments.clearFilters")}</Button>
                    </Empty>
                </Card>
            ) : groupedAssignments.selected.length === 0 ? (
                <Card>
                    <Empty>
                        <EmptyHeader>
                            <EmptyMedia variant="icon">{submissionView === "graded" ? <Award /> : submissionView === "submitted" ? <CheckCircle2 /> : submissionView === "returned" ? <RotateCcw /> : <ClipboardList />}</EmptyMedia>
                            <EmptyTitle>{t(submissionView === "graded" ? "assignments.noGraded" : submissionView === "submitted" ? "assignments.noSubmitted" : submissionView === "returned" ? "assignments.noReturned" : "assignments.nothingToSubmit")}</EmptyTitle>
                            <EmptyDescription>{t(submissionView === "graded" ? "assignments.noGradedDescription" : submissionView === "submitted" ? "assignments.noSubmittedDescription" : submissionView === "returned" ? "assignments.noReturnedDescription" : "assignments.nothingToSubmitDescription")}</EmptyDescription>
                        </EmptyHeader>
                    </Empty>
                </Card>
            ) : (
                <div className="space-y-6">
                    {groupedAssignments.current.length > 0 && submissionView !== "returned" && (
                        <AssignmentResults
                            assignments={groupedAssignments.current}
                            view={view}
                            referenceDate={currentDate}
                            referenceTime={referenceTime}
                            browserTimeZone={browserTimeZone}
                            formatDate={formatDate}
                            onSubmission={setSubmissionAssignment}
                        />
                    )}

                    {groupedAssignments.current.length > 0 && submissionView === "returned" && (
                        <section className="space-y-4 rounded-2xl border border-warning/30 bg-warning/[0.055] p-4 sm:p-5" aria-labelledby="returned-assignments-title">
                            <div className="flex items-start gap-3">
                                <span className="flex size-9 shrink-0 items-center justify-center rounded-xl bg-warning/12 text-warning-foreground">
                                    <RotateCcw className="size-4" />
                                </span>
                                <div>
                                    <h3 className="font-semibold text-warning-foreground" id="returned-assignments-title">{t("assignments.returned.title")}</h3>
                                    <p className="mt-0.5 text-sm text-muted-foreground">{t("assignments.returned.description")}</p>
                                </div>
                            </div>
                            <AssignmentResults
                                assignments={groupedAssignments.current}
                                view={view}
                                referenceDate={currentDate}
                                referenceTime={referenceTime}
                                browserTimeZone={browserTimeZone}
                                formatDate={formatDate}
                                onSubmission={setSubmissionAssignment}
                            />
                        </section>
                    )}

                    {groupedAssignments.pastDue.length > 0 && (
                        <section className="space-y-4 rounded-2xl border border-destructive/25 bg-destructive/[0.035] p-4 sm:p-5" aria-labelledby="past-due-assignments-title">
                            <div className="flex flex-wrap items-center justify-between gap-3">
                                <div className="flex items-start gap-3">
                                    <span className="flex size-9 shrink-0 items-center justify-center rounded-xl bg-destructive/10 text-destructive">
                                        <AlertTriangle className="size-4" />
                                    </span>
                                    <div>
                                        <h3 className="font-semibold text-destructive" id="past-due-assignments-title">{t("assignments.pastDue.title")}</h3>
                                        <p className="mt-0.5 text-sm text-muted-foreground">{t("assignments.pastDue.description")}</p>
                                    </div>
                                </div>
                                {groupedAssignments.hiddenClosed.length > 0 && (
                                    <Button size="sm" variant="outline" aria-expanded={showClosedPastDue} onClick={() => setShowClosedPastDue((current) => !current)}>
                                        {showClosedPastDue ? <EyeOff /> : <Eye />}
                                        {t(showClosedPastDue ? "assignments.pastDue.hideClosed" : "assignments.pastDue.showClosed", {
                                            count: groupedAssignments.hiddenClosed.length,
                                        })}
                                    </Button>
                                )}
                            </div>

                            {groupedAssignments.displayedPastDue.length > 0 ? (
                                <AssignmentResults
                                    assignments={groupedAssignments.displayedPastDue}
                                    view={view}
                                    referenceDate={currentDate}
                                    referenceTime={referenceTime}
                                    browserTimeZone={browserTimeZone}
                                    formatDate={formatDate}
                                    onSubmission={setSubmissionAssignment}
                                />
                            ) : (
                                <p className="rounded-xl border border-dashed border-destructive/20 bg-background/70 px-4 py-5 text-center text-sm text-muted-foreground">
                                    {t("assignments.pastDue.closedHidden", {count: groupedAssignments.hiddenClosed.length})}
                                </p>
                            )}
                        </section>
                    )}
                </div>
            )}

            <SubmissionDialog
                assignment={submissionAssignment}
                onOpenChange={(open) => {
                    if (!open) setSubmissionAssignment(null)
                }}
            />
        </section>
    )
}

function AssignmentResults({assignments, view, referenceDate, referenceTime, browserTimeZone, formatDate, onSubmission}: {
    assignments: PortalCourseAssignment[]
    view: AssignmentView
    referenceDate: string
    referenceTime: string
    browserTimeZone: boolean
    formatDate: (value: string) => string
    onSubmission: (assignment: PortalCourseAssignment) => void
}) {
    return (
        <div className={view === "grid" ? "grid gap-3 lg:grid-cols-2" : "space-y-3"}>
            {assignments.map((assignment) => view === "grid"
                ? <AssignmentCard assignment={assignment} referenceDate={referenceDate} referenceTime={referenceTime} browserTimeZone={browserTimeZone} formatDate={formatDate} onSubmission={() => onSubmission(assignment)} key={assignment.id} />
                : <AssignmentListItem assignment={assignment} referenceDate={referenceDate} referenceTime={referenceTime} browserTimeZone={browserTimeZone} formatDate={formatDate} onSubmission={() => onSubmission(assignment)} key={assignment.id} />)}
        </div>
    )
}

function DueBadge({assignment, referenceDate, browserTimeZone, formatDate}: {
    assignment: PortalCourseAssignment
    referenceDate: string
    browserTimeZone: boolean
    formatDate: (value: string) => string
}) {
    const {t} = useLocale()
    if (!assignment.dueDate) {
        return <Badge variant="secondary"><CalendarClock /> {t("assignments.noDue")}</Badge>
    }

    const days = relativeDueDate(assignment.dueDate, referenceDate, browserTimeZone)
    const submitted = assignment.submission?.status === "submitted"
    const label = days === null
        ? t("assignments.unknownDue")
        : days < 0 && submitted
            ? t("assignments.dueAt", {date: formatDate(assignment.dueDate)})
        : days === 0
            ? t("assignments.dueToday")
            : days === 1
                ? t("assignments.dueTomorrow")
                : days > 1
                    ? t("assignments.dueIn", {count: days})
                    : days === -1
                        ? t("assignments.overdueOne")
                        : t("assignments.overdue", {count: Math.abs(days)})
    const overdue = days !== null && days < 0 && !submitted
    return (
        <Badge variant={overdue ? "destructive" : "secondary"} title={t("assignments.dueAt", {date: formatDate(assignment.dueDate)})}>
            <CalendarClock /> {label}
        </Badge>
    )
}

function AssignmentCard({assignment, referenceDate, referenceTime, browserTimeZone, formatDate, onSubmission}: {
    assignment: PortalCourseAssignment
    referenceDate: string
    referenceTime: string
    browserTimeZone: boolean
    formatDate: (value: string) => string
    onSubmission: () => void
}) {
    const {t} = useLocale()
    return (
        <Card className="overflow-hidden">
            <CardHeader className="gap-3">
                {assignment.course && (
                    <Button className="h-auto w-fit px-0 text-xs" variant="link" render={<Link href={`/app/portal/student/courses/${assignment.course.id}`} />}>
                        {assignment.course.name}
                    </Button>
                )}
                <CardTitle className="wrap-break-word text-lg leading-snug">{assignment.title}</CardTitle>
                <AssignmentGradeStatus assignment={assignment} onClick={onSubmission} />
            </CardHeader>
            <CardContent className="flex flex-1 flex-col gap-4">
                {assignment.description
                    ? <p className="whitespace-pre-wrap wrap-break-word text-sm leading-relaxed">{assignment.description}</p>
                    : <p className="text-sm italic text-muted-foreground">{t("assignments.noDescription")}</p>}
                {assignment.referencedPost && (
                    <div className="flex min-w-0 items-center gap-2 rounded-xl border bg-muted/25 px-3 py-2.5 text-sm">
                        <Link2 className="size-4 shrink-0 text-muted-foreground" />
                        <span className="min-w-0 flex-1 truncate"><span className="text-muted-foreground">{t("assignments.referencedPost")}</span> {assignment.referencedPost.title}</span>
                        <Button size="icon-xs" variant="ghost" aria-label={t("assignments.openPost", {title: assignment.referencedPost.title})} render={<Link href={`/app/portal/student/posts/${assignment.referencedPost.id}`} />}>
                            <ExternalLink />
                        </Button>
                    </div>
                )}
                <div className="mt-auto flex flex-col gap-3 border-t pt-4 sm:flex-row sm:items-end sm:justify-between">
                    <div className="space-y-2">
                        <div className="flex flex-wrap gap-2">
                            <DueBadge assignment={assignment} referenceDate={referenceDate} browserTimeZone={browserTimeZone} formatDate={formatDate} />
                            {assignment.submissionsCloseAt && (
                                <Badge variant="outline"><Clock3 /> {t("assignments.closeAt", {date: formatDate(assignment.submissionsCloseAt)})}</Badge>
                            )}
                        </div>
                        <p className="text-xs text-muted-foreground">{t("assignments.assignedAt", {date: formatDate(assignment.createdAt)})}</p>
                    </div>
                    <SubmissionAction assignment={assignment} referenceTime={referenceTime} onSubmission={onSubmission} compact />
                </div>
            </CardContent>
        </Card>
    )
}

function AssignmentListItem({assignment, referenceDate, referenceTime, browserTimeZone, formatDate, onSubmission}: {
    assignment: PortalCourseAssignment
    referenceDate: string
    referenceTime: string
    browserTimeZone: boolean
    formatDate: (value: string) => string
    onSubmission: () => void
}) {
    const {t} = useLocale()
    return (
        <Card>
            <CardContent className="flex flex-col gap-4 p-4 sm:p-5 lg:flex-row lg:items-center">
                <div className="min-w-0 flex-1">
                    {assignment.course && (
                        <Button className="mb-1 h-auto w-fit px-0 text-xs" variant="link" render={<Link href={`/app/portal/student/courses/${assignment.course.id}`} />}>
                            {assignment.course.name}
                        </Button>
                    )}
                    <h3 className="wrap-break-word font-semibold leading-snug">{assignment.title}</h3>
                    {assignment.description && (
                        <p className="mt-1 line-clamp-2 text-sm leading-relaxed text-muted-foreground">{assignment.description}</p>
                    )}
                    <div className="mt-3 flex flex-wrap gap-2">
                        <DueBadge assignment={assignment} referenceDate={referenceDate} browserTimeZone={browserTimeZone} formatDate={formatDate} />
                        {assignment.submissionsCloseAt && (
                            <Badge variant="outline"><Clock3 /> {t("assignments.closeAt", {date: formatDate(assignment.submissionsCloseAt)})}</Badge>
                        )}
                    </div>
                    <p className="mt-2 text-xs text-muted-foreground">{t("assignments.assignedAt", {date: formatDate(assignment.createdAt)})}</p>
                </div>

                <div className="flex shrink-0 flex-wrap items-center gap-2 lg:max-w-[25rem] lg:justify-end">
                    <AssignmentGradeStatus assignment={assignment} compact onClick={onSubmission} />
                    {assignment.referencedPost && (
                        <Button
                            className="max-w-full"
                            size="sm"
                            variant="outline"
                            render={<Link href={`/app/portal/student/posts/${assignment.referencedPost.id}`} />}
                        >
                            <Link2 /> <span className="truncate">{assignment.referencedPost.title}</span>
                            <ExternalLink />
                        </Button>
                    )}
                    <SubmissionAction assignment={assignment} referenceTime={referenceTime} onSubmission={onSubmission} compact />
                </div>
            </CardContent>
        </Card>
    )
}

function AssignmentGradeStatus({assignment, onClick, compact = false}: {
    assignment: PortalCourseAssignment
    onClick: () => void
    compact?: boolean
}) {
    const {t} = useLocale()
    const grade = assignment.submission?.status === "submitted" ? assignment.submission.grade : null
    if (!grade) return null

    const accent = scoreAccent(grade.score)
    return (
        <button
            className={`flex cursor-pointer items-center gap-3 rounded-xl border text-left outline-none transition-[filter,box-shadow] hover:brightness-95 focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 dark:hover:brightness-110 ${compact ? "px-3 py-2" : "w-full p-3"}`}
            type="button"
            title={t("assignments.grade.title", {score: grade.score})}
            aria-label={t("assignments.grade.viewDetails", {score: grade.score})}
            onClick={onClick}
            style={{
                backgroundColor: `color-mix(in oklab, ${accent} 10%, transparent)`,
                borderColor: `color-mix(in oklab, ${accent} 30%, var(--border))`,
            }}
        >
            <span className="flex size-9 shrink-0 items-center justify-center rounded-lg" style={{backgroundColor: `color-mix(in oklab, ${accent} 16%, transparent)`, color: accent}}>
                <Award className="size-4" />
            </span>
            <div className="min-w-0">
                <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">{t("assignments.grade.graded")}</p>
                <p className="text-xl font-semibold leading-tight tabular-nums" style={{color: accent}}>{t("assignments.grade.score", {score: grade.score})}</p>
            </div>
        </button>
    )
}

function SubmissionAction({assignment, referenceTime, onSubmission, compact = false}: {
    assignment: PortalCourseAssignment
    referenceTime: string
    onSubmission: () => void
    compact?: boolean
}) {
    const {t} = useLocale()
    const closed = submissionsClosed(assignment, referenceTime)

    if (assignment.submission?.status === "submitted") {
        const graded = assignment.submission.grade !== null
        return (
            <Button size="sm" variant="outline" onClick={onSubmission}>
                {graded ? <Award /> : <CheckCircle2 className="text-emerald-600" />}
                {t(graded ? "assignments.grade.viewSubmissionDetails" : "assignments.submission.viewSubmitted")}
            </Button>
        )
    }

    if (assignment.submission?.status === "returned") {
        if (closed) {
            return (
                <span className="flex flex-wrap items-center gap-2">
                    <Badge variant="warning"><RotateCcw /> {t("assignments.submission.returned")}</Badge>
                    <Badge variant="outline"><Clock3 /> {t("assignments.submission.closed")}</Badge>
                </span>
            )
        }

        return (
            <Button className={`${compact ? "" : "w-full"} border-warning/35 text-warning-foreground hover:bg-warning/10`} size="sm" variant="outline" onClick={onSubmission}>
                <RotateCcw /> {t("assignments.submission.resubmit")}
            </Button>
        )
    }

    if (closed) {
        return <Badge variant="outline"><Clock3 /> {t("assignments.submission.closed")}</Badge>
    }

    return (
        <Button className={compact ? undefined : "w-full"} size="sm" variant={assignment.submission?.status === "pending" ? "outline" : "default"} onClick={onSubmission}>
            <Send />
            {t(assignment.submission?.status === "pending" ? "assignments.submission.continue" : "assignments.submission.submitWork")}
        </Button>
    )
}

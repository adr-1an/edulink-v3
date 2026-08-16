"use client"

import {useDeferredValue, useMemo, useState} from "react"
import Link from "next/link"
import {useRouter} from "next/navigation"
import {
    ArrowLeft, ChevronLeft, ChevronRight, ClipboardCheck, Eye, RotateCcw, Search, StickyNote, UserRound,
} from "lucide-react"
import LocalDateTime from "@/components/local-date-time"
import StudentProfilePopover, {portalStudentDisplayName, type PortalStudentSummary} from "@/components/app/student-profile-popover"
import {Badge} from "@/components/ui/badge"
import {Button} from "@/components/ui/button"
import {Card, CardContent} from "@/components/ui/card"
import {Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle} from "@/components/ui/empty"
import {InputGroup, InputGroupAddon, InputGroupInput} from "@/components/ui/input-group"
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table"
import {useLocale} from "@/i18n/provider"

export interface AssignmentSubmission {
    id: string
    status: "submitted" | "returned"
    submittedBy: PortalStudentSummary
    notes: string | null
    submittedAt: string
}

const pageSize = 20

export default function SubmissionsClientPage({
    submissions,
    canViewSubmissions,
}: {
    submissions: AssignmentSubmission[]
    canViewSubmissions: boolean
}) {
    const router = useRouter()
    const {locale, t} = useLocale()
    const [query, setQuery] = useState("")
    const [statusView, setStatusView] = useState<AssignmentSubmission["status"]>("submitted")
    const [requestedPage, setRequestedPage] = useState(1)
    const deferredQuery = useDeferredValue(query)
    const submittedCount = submissions.filter(({status}) => status === "submitted").length
    const returnedCount = submissions.filter(({status}) => status === "returned").length
    const filteredSubmissions = useMemo(() => {
        const normalizedQuery = deferredQuery.trim().toLocaleLowerCase(locale === "pl" ? "pl-PL" : "en-US")
        const statusSubmissions = submissions.filter(({status}) => status === statusView)
        if (!normalizedQuery) return statusSubmissions

        return statusSubmissions.filter((submission) => [
            portalStudentDisplayName(submission.submittedBy),
            submission.submittedBy.email,
            submission.submittedBy.phone,
            submission.notes ?? "",
        ].some((value) => value.toLocaleLowerCase(locale === "pl" ? "pl-PL" : "en-US").includes(normalizedQuery)))
    }, [deferredQuery, locale, statusView, submissions])
    const pageCount = Math.max(1, Math.ceil(filteredSubmissions.length / pageSize))
    const currentPage = Math.min(requestedPage, pageCount)
    const visibleSubmissions = filteredSubmissions.slice((currentPage - 1) * pageSize, currentPage * pageSize)
    const firstVisible = filteredSubmissions.length === 0 ? 0 : (currentPage - 1) * pageSize + 1
    const lastVisible = Math.min(currentPage * pageSize, filteredSubmissions.length)

    return (
        <div className="mx-auto w-full max-w-6xl space-y-6">
            <div className="space-y-4">
                <Button variant="ghost" className="-ml-3" onClick={() => router.back()}>
                    <ArrowLeft /> {t("staff.submissions.back")}
                </Button>
                <div className="flex flex-wrap items-start justify-between gap-4">
                    <div>
                        <div className="mb-2 flex items-center gap-2">
                            <Badge variant="secondary"><ClipboardCheck /> {t("staff.submissions.badge")}</Badge>
                            <Badge variant="outline">{t("staff.submissions.total", {count: submissions.length})}</Badge>
                        </div>
                        <h1 className="text-3xl font-semibold tracking-tight">{t("staff.submissions.title")}</h1>
                        <p className="mt-2 max-w-2xl text-sm text-muted-foreground">{t("staff.submissions.description")}</p>
                    </div>
                </div>
            </div>

            {submissions.length === 0 ? (
                <Card>
                    <Empty>
                        <EmptyHeader>
                            <EmptyMedia variant="icon"><ClipboardCheck /></EmptyMedia>
                            <EmptyTitle>{t("staff.submissions.emptyTitle")}</EmptyTitle>
                            <EmptyDescription>{t("staff.submissions.emptyDescription")}</EmptyDescription>
                        </EmptyHeader>
                    </Empty>
                </Card>
            ) : (
                <>
                    <div className="flex w-fit max-w-full items-center gap-1 rounded-xl border bg-muted/50 p-1" role="tablist" aria-label={t("staff.submissions.statusTabs")}>
                        <Button
                            size="sm"
                            variant={statusView === "submitted" ? "secondary" : "ghost"}
                            role="tab"
                            aria-selected={statusView === "submitted"}
                            onClick={() => {
                                setStatusView("submitted")
                                setRequestedPage(1)
                            }}
                        >
                            <ClipboardCheck /> {t("staff.submissions.submittedTab")}
                            <Badge variant="secondary">{submittedCount}</Badge>
                        </Button>
                        <Button
                            size="sm"
                            variant={statusView === "returned" ? "secondary" : "ghost"}
                            role="tab"
                            aria-selected={statusView === "returned"}
                            onClick={() => {
                                setStatusView("returned")
                                setRequestedPage(1)
                            }}
                        >
                            <RotateCcw /> {t("staff.submissions.returnedTab")}
                            <Badge variant="warning">{returnedCount}</Badge>
                        </Button>
                    </div>

                    <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                        <InputGroup className="sm:max-w-md">
                            <InputGroupAddon><Search /></InputGroupAddon>
                            <InputGroupInput
                                value={query}
                                onChange={(event) => {
                                    setQuery(event.target.value)
                                    setRequestedPage(1)
                                }}
                                placeholder={t("staff.submissions.searchPlaceholder")}
                                aria-label={t("staff.submissions.searchLabel")}
                            />
                        </InputGroup>
                        <p className="text-sm text-muted-foreground" aria-live="polite">
                            {t("staff.submissions.showing", {
                                first: firstVisible,
                                last: lastVisible,
                                total: filteredSubmissions.length,
                            })}
                        </p>
                    </div>

                    {filteredSubmissions.length === 0 ? (
                        <Card>
                            <Empty className="py-14 md:py-16">
                                <EmptyHeader>
                                    <EmptyMedia variant="icon"><Search /></EmptyMedia>
                                    <EmptyTitle>{t(query ? "staff.submissions.noMatches" : statusView === "returned" ? "staff.submissions.noReturned" : "staff.submissions.noSubmitted")}</EmptyTitle>
                                    <EmptyDescription>{t(query ? "staff.submissions.noMatchesDescription" : statusView === "returned" ? "staff.submissions.noReturnedDescription" : "staff.submissions.noSubmittedDescription")}</EmptyDescription>
                                </EmptyHeader>
                                {query && <Button variant="outline" onClick={() => setQuery("")}>{t("staff.submissions.clearSearch")}</Button>}
                            </Empty>
                        </Card>
                    ) : (
                        <>
                            <Card className="hidden overflow-hidden md:block">
                                <Table>
                                    <TableHeader>
                                        <TableRow>
                                            <TableHead className="pl-5">{t("staff.submissions.student")}</TableHead>
                                            <TableHead>{t("common.email")}</TableHead>
                                            <TableHead>{t("staff.submissions.notes")}</TableHead>
                                            <TableHead className="pr-5 text-right">{t("staff.submissions.submittedAt")}</TableHead>
                                            {canViewSubmissions && <TableHead><span className="sr-only">{t("staff.submission.open")}</span></TableHead>}
                                        </TableRow>
                                    </TableHeader>
                                    <TableBody>
                                        {visibleSubmissions.map((submission) => (
                                            <TableRow key={submission.id}>
                                                <TableCell className="pl-5"><StudentProfilePopover student={submission.submittedBy} /></TableCell>
                                                <TableCell>
                                                    <a className="text-muted-foreground underline-offset-4 hover:text-foreground hover:underline" href={`mailto:${submission.submittedBy.email}`}>
                                                        {submission.submittedBy.email}
                                                    </a>
                                                </TableCell>
                                                <TableCell className="max-w-sm whitespace-normal leading-relaxed">
                                                    {submission.notes
                                                        ? <span className="line-clamp-2">{submission.notes}</span>
                                                        : <span className="italic text-muted-foreground">{t("staff.submissions.noNotes")}</span>}
                                                </TableCell>
                                                <TableCell className="pr-5 text-right text-muted-foreground">
                                                    <LocalDateTime value={submission.submittedAt} />
                                                </TableCell>
                                                {canViewSubmissions && (
                                                    <TableCell className="w-px pr-5">
                                                        <Button
                                                            size="sm"
                                                            variant="outline"
                                                            aria-label={t("staff.submission.openFor", {name: portalStudentDisplayName(submission.submittedBy)})}
                                                            render={<Link href={`/app/staff/submissions/${submission.id}`} />}
                                                        >
                                                            <Eye /> {t("staff.submission.open")}
                                                        </Button>
                                                    </TableCell>
                                                )}
                                            </TableRow>
                                        ))}
                                    </TableBody>
                                </Table>
                            </Card>

                            <div className="grid gap-3 md:hidden">
                                {visibleSubmissions.map((submission) => (
                                    <Card key={submission.id}>
                                        <CardContent className="space-y-4 p-4">
                                            <div className="flex items-start justify-between gap-3">
                                                <StudentProfilePopover student={submission.submittedBy} showEmail />
                                                <Badge variant={submission.status === "returned" ? "warning" : "success"}>
                                                    {submission.status === "returned" ? <RotateCcw /> : <ClipboardCheck />}
                                                    {t(submission.status === "returned" ? "staff.submissions.returned" : "staff.submissions.submitted")}
                                                </Badge>
                                            </div>
                                            <div className="rounded-xl border bg-muted/20 p-3">
                                                <div className="mb-1.5 flex items-center gap-2 text-xs font-medium text-muted-foreground">
                                                    <StickyNote className="size-3.5" /> {t("staff.submissions.notes")}
                                                </div>
                                                {submission.notes
                                                    ? <p className="whitespace-pre-wrap break-words text-sm leading-relaxed">{submission.notes}</p>
                                                    : <p className="text-sm italic text-muted-foreground">{t("staff.submissions.noNotes")}</p>}
                                            </div>
                                            <p className="flex items-center gap-2 text-xs text-muted-foreground">
                                                <UserRound className="size-3.5" />
                                                {t("staff.submissions.submittedAt")} <LocalDateTime value={submission.submittedAt} />
                                            </p>
                                            {canViewSubmissions && (
                                                <Button className="w-full" variant="outline" render={<Link href={`/app/staff/submissions/${submission.id}`} />}>
                                                    <Eye /> {t("staff.submission.open")}
                                                </Button>
                                            )}
                                        </CardContent>
                                    </Card>
                                ))}
                            </div>

                            {pageCount > 1 && (
                                <nav className="flex items-center justify-between gap-3" aria-label={t("staff.submissions.pagination")}>
                                    <Button
                                        variant="outline"
                                        disabled={currentPage === 1}
                                        onClick={() => setRequestedPage(Math.max(1, currentPage - 1))}
                                    >
                                        <ChevronLeft /> {t("staff.submissions.previous")}
                                    </Button>
                                    <span className="text-sm text-muted-foreground">
                                        {t("staff.submissions.page", {current: currentPage, total: pageCount})}
                                    </span>
                                    <Button
                                        variant="outline"
                                        disabled={currentPage === pageCount}
                                        onClick={() => setRequestedPage(Math.min(pageCount, currentPage + 1))}
                                    >
                                        {t("staff.submissions.next")} <ChevronRight />
                                    </Button>
                                </nav>
                            )}
                        </>
                    )}
                </>
            )}
        </div>
    )
}

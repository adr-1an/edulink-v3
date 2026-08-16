"use client"

import {Fragment, useDeferredValue, useMemo, useRef, useState} from "react"
import {ChevronRight, CircleEllipsis, FilePlus2, FilePenLine, FileX2, Mail, ScrollText, Search} from "lucide-react"
import PageTitle from "@/components/app/page_title"
import LocalDateTime, {useLocalDateTimeFormatter} from "@/components/local-date-time"
import {Avatar, AvatarFallback} from "@/components/ui/avatar"
import {Badge} from "@/components/ui/badge"
import {Button} from "@/components/ui/button"
import {Card} from "@/components/ui/card"
import {Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle} from "@/components/ui/empty"
import {Input} from "@/components/ui/input"
import {Popover, PopoverDescription, PopoverPopup, PopoverTitle, PopoverTrigger} from "@/components/ui/popover"
import {
    Select, SelectItem, SelectPopup, SelectTrigger, SelectValue,
} from "@/components/ui/select"
import {
    Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from "@/components/ui/table"
import {useLocale} from "@/i18n/provider"
import {type MessageKey} from "@/i18n/messages"

export type SchoolLogType = "create" | "edit" | "delete" | "other"

export interface SchoolLog {
    id: string
    user: {
        id: string
        name: string
        email: string
    }
    action: string
    type: SchoolLogType
    title: string
    message: string
    details: string
    createdAt: string
}

type TypeFilter = "all" | SchoolLogType

const PAGE_SIZE = 25
const typeDetails = {
    create: {label: "staff.logs.type.create", icon: FilePlus2, variant: "success" as const},
    edit: {label: "staff.logs.type.edit", icon: FilePenLine, variant: "info" as const},
    delete: {label: "staff.logs.type.delete", icon: FileX2, variant: "error" as const},
    other: {label: "staff.logs.type.other", icon: CircleEllipsis, variant: "secondary" as const},
} satisfies Record<SchoolLogType, {label: MessageKey; icon: typeof FilePlus2; variant: "success" | "info" | "error" | "secondary"}>

function initials(name: string) {
    const parts = name.trim().split(/\s+/).filter(Boolean)
    return parts.slice(0, 2).map((part) => part[0]?.toUpperCase()).join("") || "?"
}

export default function LogsClientPage({logs}: {logs: SchoolLog[]}) {
    const {t} = useLocale()
    const [query, setQuery] = useState("")
    const [typeFilter, setTypeFilter] = useState<TypeFilter>("all")
    const [page, setPage] = useState(1)
    const [expandedLogID, setExpandedLogID] = useState<string | null>(null)
    const tableRef = useRef<HTMLDivElement>(null)
    const deferredQuery = useDeferredValue(query)
    const searchPending = query !== deferredQuery
    const formatDate = useLocalDateTimeFormatter()

    const filteredLogs = useMemo(() => {
        const search = deferredQuery.trim().toLocaleLowerCase()
        return logs.filter((log) => {
            const matchesType = typeFilter === "all" || log.type === typeFilter
            const matchesSearch = !search || [
                log.action,
                log.title,
                log.message,
                log.details,
                log.user.id,
                log.user.name,
                log.user.email,
                formatDate(log.createdAt, "second"),
            ]
                .some((value) => value.toLocaleLowerCase().includes(search))
            return matchesType && matchesSearch
        })
    }, [deferredQuery, formatDate, logs, typeFilter])

    const pageCount = Math.max(1, Math.ceil(filteredLogs.length / PAGE_SIZE))
    const currentPage = Math.min(page, pageCount)
    const visibleLogs = filteredLogs.slice((currentPage - 1) * PAGE_SIZE, currentPage * PAGE_SIZE)

    function updateQuery(value: string) {
        setQuery(value)
        setPage(1)
        setExpandedLogID(null)
    }

    function updateType(value: TypeFilter | null) {
        setTypeFilter(value ?? "all")
        setPage(1)
        setExpandedLogID(null)
    }

    function changePage(nextPage: number) {
        setPage(Math.max(1, Math.min(pageCount, nextPage)))
        setExpandedLogID(null)
        requestAnimationFrame(() => tableRef.current?.scrollIntoView({behavior: "smooth", block: "start"}))
    }

    function toggleLog(logID: string) {
        setExpandedLogID((current) => current === logID ? null : logID)
    }

    return (
        <div className="space-y-6">
            <header>
                <PageTitle centered={false}>{t("staff.logs.title")}</PageTitle>
                <p className="text-muted-foreground">{t("staff.logs.description")}</p>
            </header>

            {logs.length === 0 ? (
                <Card>
                    <Empty>
                        <EmptyHeader>
                            <EmptyMedia variant="icon"><ScrollText /></EmptyMedia>
                            <EmptyTitle>{t("staff.logs.empty")}</EmptyTitle>
                            <EmptyDescription>{t("staff.logs.emptyDescription")}</EmptyDescription>
                        </EmptyHeader>
                    </Empty>
                </Card>
            ) : (
                <>
                    <div className="flex flex-col gap-3 sm:flex-row">
                        <div className="relative w-full max-w-md">
                            <Search className="pointer-events-none absolute top-1/2 left-3 z-10 size-4 -translate-y-1/2 text-muted-foreground" />
                            <Input
                                className="pl-9"
                                value={query}
                                onChange={(event) => updateQuery(event.target.value)}
                                placeholder={t("staff.logs.searchPlaceholder")}
                                aria-label={t("staff.logs.search")}
                            />
                        </div>
                        <Select
                            items={[
                                {label: t("staff.logs.allTypes"), value: "all"},
                                {label: t("staff.logs.type.create"), value: "create"},
                                {label: t("staff.logs.type.edit"), value: "edit"},
                                {label: t("staff.logs.type.delete"), value: "delete"},
                                {label: t("staff.logs.type.other"), value: "other"},
                            ]}
                            value={typeFilter}
                            onValueChange={(value) => updateType(value as TypeFilter | null)}
                        >
                            <SelectTrigger className="w-full sm:w-40" aria-label={t("staff.logs.filterType")}>
                                <SelectValue />
                            </SelectTrigger>
                            <SelectPopup>
                                <SelectItem value="all">{t("staff.logs.allTypes")}</SelectItem>
                                <SelectItem value="create">{t("staff.logs.type.create")}</SelectItem>
                                <SelectItem value="edit">{t("staff.logs.type.edit")}</SelectItem>
                                <SelectItem value="delete">{t("staff.logs.type.delete")}</SelectItem>
                                <SelectItem value="other">{t("staff.logs.type.other")}</SelectItem>
                            </SelectPopup>
                        </Select>
                    </div>

                    {visibleLogs.length === 0 ? (
                        <Card>
                            <Empty className="py-12 md:py-12">
                                <EmptyHeader>
                                    <EmptyMedia variant="icon"><Search /></EmptyMedia>
                                    <EmptyTitle>{t("staff.logs.noMatch")}</EmptyTitle>
                                    <EmptyDescription>{t("staff.logs.noMatchDescription")}</EmptyDescription>
                                </EmptyHeader>
                                <Button variant="outline" onClick={() => {
                                    updateQuery("")
                                    updateType("all")
                                }}>{t("staff.logs.clearFilters")}</Button>
                            </Empty>
                        </Card>
                    ) : (
                        <Card ref={tableRef} className={`scroll-mt-4 overflow-hidden transition-opacity ${searchPending ? "opacity-70" : "opacity-100"}`} aria-busy={searchPending}>
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead className="w-10 pl-4"><span className="sr-only">{t("staff.logs.expand")}</span></TableHead>
                                        <TableHead>{t("staff.logs.time")}</TableHead>
                                        <TableHead>{t("staff.logs.activity")}</TableHead>
                                        <TableHead>{t("staff.logs.summary")}</TableHead>
                                        <TableHead className="pr-4">{t("staff.logs.actor")}</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {visibleLogs.map((log) => {
                                        const details = typeDetails[log.type]
                                        const Icon = details.icon
                                        const expanded = expandedLogID === log.id
                                        return (
                                            <Fragment key={log.id}>
                                                <TableRow
                                                    className="cursor-pointer outline-none focus-visible:bg-muted/40 focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring"
                                                    tabIndex={0}
                                                    aria-expanded={expanded}
                                                    onClick={() => toggleLog(log.id)}
                                                    onKeyDown={(event) => {
                                                        if (event.target !== event.currentTarget || (event.key !== "Enter" && event.key !== " ")) return
                                                        event.preventDefault()
                                                        toggleLog(log.id)
                                                    }}
                                                >
                                                    <TableCell className="pl-4">
                                                        <Button
                                                            size="icon-xs"
                                                            variant="ghost"
                                                            aria-expanded={expanded}
                                                            aria-controls={`log-details-${log.id}`}
                                                            aria-label={t("staff.logs.expandDetails", {
                                                                action: t(expanded ? "staff.logs.collapseAction" : "staff.logs.expandAction"),
                                                                title: log.title,
                                                            })}
                                                            onClick={(event) => {
                                                                event.stopPropagation()
                                                                toggleLog(log.id)
                                                            }}
                                                        >
                                                            <ChevronRight className={`transition-transform ${expanded ? "rotate-90" : ""}`} />
                                                        </Button>
                                                    </TableCell>
                                                    <TableCell className="text-xs text-muted-foreground"><LocalDateTime value={log.createdAt} precision="second" /></TableCell>
                                                    <TableCell>
                                                        <div className="flex items-center gap-2">
                                                            <Badge variant={details.variant}><Icon /> {t(details.label)}</Badge>
                                                            <span className="font-medium">{log.title}</span>
                                                        </div>
                                                    </TableCell>
                                                    <TableCell className="min-w-64 max-w-xl whitespace-normal leading-relaxed text-muted-foreground">{log.message}</TableCell>
                                                    <TableCell className="pr-4" onClick={(event) => event.stopPropagation()}>
                                                        <Popover>
                                                            <PopoverTrigger
                                                                className="rounded-sm text-left text-sm font-medium underline decoration-muted-foreground/50 underline-offset-4 outline-none hover:text-primary focus-visible:ring-2 focus-visible:ring-ring"
                                                                type="button"
                                                            >
                                                                {log.user.name}
                                                            </PopoverTrigger>
                                                            <PopoverPopup className="w-72" align="end">
                                                                <div className="flex min-w-0 items-center gap-3">
                                                                    <Avatar className="size-10 border">
                                                                        <AvatarFallback>{initials(log.user.name)}</AvatarFallback>
                                                                    </Avatar>
                                                                    <div className="min-w-0">
                                                                        <PopoverTitle className="truncate text-base">{log.user.name}</PopoverTitle>
                                                                        <PopoverDescription className="mt-1 flex items-center gap-1.5">
                                                                            <Mail className="size-3.5 shrink-0" />
                                                                            <span className="truncate">{log.user.email}</span>
                                                                        </PopoverDescription>
                                                                    </div>
                                                                </div>
                                                            </PopoverPopup>
                                                        </Popover>
                                                    </TableCell>
                                                </TableRow>
                                                {expanded && (
                                                    <TableRow id={`log-details-${log.id}`} className="bg-muted/20 hover:bg-muted/20">
                                                        <TableCell colSpan={5} className="px-4 py-4 whitespace-normal">
                                                            <div className="ml-9 rounded-lg border bg-background px-4 py-3">
                                                                <p className="mb-1 text-xs font-medium uppercase tracking-wide text-muted-foreground">{t("staff.logs.fullDetails")}</p>
                                                                <p className="whitespace-pre-wrap break-words text-sm leading-relaxed">{log.details}</p>
                                                            </div>
                                                        </TableCell>
                                                    </TableRow>
                                                )}
                                            </Fragment>
                                        )
                                    })}
                                </TableBody>
                            </Table>
                        </Card>
                    )}

                    {filteredLogs.length > 0 && (
                        <div className="flex flex-wrap items-center justify-between gap-3">
                            <p className="text-sm text-muted-foreground" aria-live="polite">
                                {t("staff.logs.showing", {
                                    from: (currentPage - 1) * PAGE_SIZE + 1,
                                    to: Math.min(currentPage * PAGE_SIZE, filteredLogs.length),
                                    total: filteredLogs.length,
                                })}
                            </p>
                            <div className="flex items-center gap-2">
                                <Button variant="outline" size="sm" disabled={currentPage === 1} onClick={() => changePage(currentPage - 1)}>{t("staff.logs.previous")}</Button>
                                <span className="min-w-20 text-center text-sm">{t("staff.logs.page", {current: currentPage, total: pageCount})}</span>
                                <Button variant="outline" size="sm" disabled={currentPage === pageCount} onClick={() => changePage(currentPage + 1)}>{t("staff.logs.next")}</Button>
                            </div>
                        </div>
                    )}
                </>
            )}
        </div>
    )
}

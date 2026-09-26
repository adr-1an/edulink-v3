"use client"

import {useEffect, useMemo, useState} from "react"
import {useRouter} from "next/navigation"
import {toast} from "sonner"
import {
    CalendarDays, ChevronLeft, ChevronRight, Mail, Pencil, Phone, Plus, Search, Trash2,
    TriangleAlert, UserRoundCheck, UsersRound, X,
} from "lucide-react"
import PageTitle from "@/components/app/page_title"
import UserAvatar from "@/components/app/user_avatar"
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
import {useLocale} from "@/i18n/provider"
import {type Locale} from "@/i18n/config"
import {hasSchoolPermission, type SchoolAccess} from "@/lib/school_access"
import {rememberSchoolAccess} from "@/lib/school_navigation"
import {
    emptyGuardianDraft, guardianDraftToInput, type Guardian, type GuardianDraft,
} from "@/lib/guardian"
import {handleCreateGuardian, handleDeleteGuardian, handleUpdateGuardian} from "./actions"
import {guardianErrorKeys} from "./guardian_action_errors"
import GuardianForm from "./guardian_form"

const pageSize = 20

function formatDateOnly(value: string, locale: Locale) {
    const date = new Date(`${value}T00:00:00Z`)
    if (Number.isNaN(date.getTime())) return value
    return new Intl.DateTimeFormat(locale === "pl" ? "pl-PL" : "en-US", {
        day: "numeric",
        month: "short",
        year: "numeric",
        timeZone: "UTC",
    }).format(date)
}

export default function GuardiansClientPage({schoolID, initialGuardians, access}: {
    schoolID: string
    initialGuardians: Guardian[]
    access: SchoolAccess
}) {
    const router = useRouter()
    const {locale, t} = useLocale()
    const [guardians, setGuardians] = useState(initialGuardians)
    const [query, setQuery] = useState("")
    const [requestedPage, setRequestedPage] = useState(1)
    const [createOpen, setCreateOpen] = useState(false)
    const [editTarget, setEditTarget] = useState<Guardian | null>(null)
    const [deleteTarget, setDeleteTarget] = useState<Guardian | null>(null)
    const [deleteConfirmation, setDeleteConfirmation] = useState("")
    const [draft, setDraft] = useState<GuardianDraft>(emptyGuardianDraft)
    const [saving, setSaving] = useState(false)
    const [deleting, setDeleting] = useState(false)
    const canCreate = hasSchoolPermission(access, "guardian.create")
    const canUpdate = hasSchoolPermission(access, "guardian.update")
    const canDelete = hasSchoolPermission(access, "guardian.delete")

    useEffect(() => {
        rememberSchoolAccess(schoolID, access)
    }, [schoolID, access])

    const filteredGuardians = useMemo(() => {
        const search = query.trim().toLocaleLowerCase(locale === "pl" ? "pl-PL" : "en-US")
        return [...guardians]
            .filter((guardian) => !search || [guardian.name, guardian.lastName, guardian.email, guardian.phone]
                .some((value) => value.toLocaleLowerCase(locale === "pl" ? "pl-PL" : "en-US").includes(search)))
            .sort((first, second) => first.lastName.localeCompare(second.lastName, locale === "pl" ? "pl-PL" : "en-US")
                || first.name.localeCompare(second.name, locale === "pl" ? "pl-PL" : "en-US"))
    }, [guardians, locale, query])
    const pageCount = Math.max(1, Math.ceil(filteredGuardians.length / pageSize))
    const currentPage = Math.min(requestedPage, pageCount)
    const visibleGuardians = filteredGuardians.slice((currentPage - 1) * pageSize, currentPage * pageSize)
    const firstVisibleGuardian = filteredGuardians.length === 0 ? 0 : (currentPage - 1) * pageSize + 1
    const lastVisibleGuardian = Math.min(currentPage * pageSize, filteredGuardians.length)
    const enabledAccounts = guardians.filter((guardian) => guardian.accountEnabled).length
    const requiredDeleteConfirmation = deleteTarget ? `${deleteTarget.name} ${deleteTarget.lastName}` : ""

    function openCreate() {
        setDraft(emptyGuardianDraft)
        setCreateOpen(true)
    }

    function openEdit(guardian: Guardian) {
        setDraft({
            name: guardian.name,
            lastName: guardian.lastName,
            dateOfBirth: guardian.dateOfBirth,
            email: guardian.email,
            phone: guardian.phone,
            notes: guardian.notes,
            accountEnabled: guardian.accountEnabled,
            password: "",
        })
        setEditTarget(guardian)
    }

    function openDelete(guardian: Guardian) {
        setDeleteConfirmation("")
        setDeleteTarget(guardian)
    }

    async function createGuardian(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        setSaving(true)
        const res = await handleCreateGuardian(schoolID, guardianDraftToInput(draft))
        setSaving(false)
        if (!res.ok) return toast.error(t(guardianErrorKeys[res.code]))

        setCreateOpen(false)
        toast.success(t("staff.guardians.created"))
        router.refresh()
    }

    async function updateGuardian(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        if (!editTarget) return
        const input = guardianDraftToInput(draft)
        setSaving(true)
        const res = await handleUpdateGuardian(editTarget.id, input)
        setSaving(false)
        if (!res.ok) return toast.error(t(guardianErrorKeys[res.code]))

        const updated: Guardian = {
            ...editTarget,
            name: input.name,
            lastName: input.lastName,
            dateOfBirth: draft.dateOfBirth,
            email: input.email,
            phone: input.phone ?? "",
            notes: input.notes ?? "",
            accountEnabled: input.accountEnabled,
        }
        setGuardians((current) => current.map((guardian) => guardian.id === editTarget.id ? updated : guardian))
        setEditTarget(null)
        toast.success(t("staff.guardians.updated"))
    }

    async function deleteGuardian() {
        if (!deleteTarget || deleteConfirmation !== requiredDeleteConfirmation) return
        setDeleting(true)
        const res = await handleDeleteGuardian(deleteTarget.id)
        setDeleting(false)
        if (!res.ok) return toast.error(t(guardianErrorKeys[res.code]))

        setGuardians((current) => current.filter((guardian) => guardian.id !== deleteTarget.id))
        setDeleteTarget(null)
        setDeleteConfirmation("")
        toast.success(t("staff.guardians.deleted"))
    }

    return (
        <div className="space-y-6">
            <div className="flex flex-wrap items-end justify-between gap-3">
                <div>
                    <PageTitle centered={false}>{t("staff.guardians.title")}</PageTitle>
                    <p className="text-muted-foreground">{t("staff.guardians.description")}</p>
                </div>
                {canCreate && <Button onClick={openCreate}><Plus /> {t("staff.guardians.add")}</Button>}
            </div>

            <div className="grid gap-3 sm:grid-cols-2">
                <SummaryCard icon={UsersRound} label={t("staff.guardians.summaryGuardians")} value={guardians.length} />
                <SummaryCard icon={UserRoundCheck} label={t("staff.guardians.loginEnabled")} value={enabledAccounts} />
            </div>

            {guardians.length > 0 && (
                <div className="relative max-w-md">
                    <Search className="pointer-events-none absolute left-3 top-1/2 z-10 size-4 -translate-y-1/2 text-muted-foreground" />
                    <Input className="pl-9 pr-9" value={query} onChange={(event) => { setQuery(event.target.value); setRequestedPage(1) }} placeholder={t("staff.guardians.searchPlaceholder")} aria-label={t("staff.guardians.search")} />
                    {query && <Button className="absolute right-1.5 top-1/2 -translate-y-1/2" size="icon-xs" variant="ghost" aria-label={t("staff.guardians.clearSearch")} onClick={() => { setQuery(""); setRequestedPage(1) }}><X /></Button>}
                </div>
            )}

            {guardians.length === 0 ? (
                <Card>
                    <Empty>
                        <EmptyHeader>
                            <EmptyMedia variant="icon"><UsersRound /></EmptyMedia>
                            <EmptyTitle>{t("staff.guardians.empty")}</EmptyTitle>
                            <EmptyDescription>{t(canCreate ? "staff.guardians.emptyCreate" : "staff.guardians.emptyRestricted")}</EmptyDescription>
                        </EmptyHeader>
                        {canCreate && <Button onClick={openCreate}><Plus /> {t("staff.guardians.add")}</Button>}
                    </Empty>
                </Card>
            ) : filteredGuardians.length === 0 ? (
                <Card className="p-8 text-center text-sm text-muted-foreground">{t("staff.guardians.noMatch", {query: query.trim()})}</Card>
            ) : (
                <Card className="overflow-hidden p-0">
                    <div className="divide-y">
                        {visibleGuardians.map((guardian) => {
                            const fullName = `${guardian.name} ${guardian.lastName}`
                            return (
                                <div className="flex flex-col gap-3 px-4 py-4 transition-colors hover:bg-muted/35 sm:flex-row sm:items-center sm:px-5" key={guardian.id}>
                                    <UserAvatar className="size-10 border" name={fullName} src={null} cacheKey={`guardian:${guardian.id}`} />
                                    <div className="min-w-0 flex-1">
                                        <div className="flex flex-wrap items-center gap-2">
                                            <p className="truncate font-medium">{fullName}</p>
                                            <Badge variant={guardian.accountEnabled ? "default" : "secondary"}>{t(guardian.accountEnabled ? "staff.guardians.loginEnabled" : "staff.guardians.noLogin")}</Badge>
                                        </div>
                                        <div className="mt-1 flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
                                            <span className="flex items-center gap-1.5"><Mail className="size-3.5" /> {guardian.email}</span>
                                            {guardian.phone && <span className="flex items-center gap-1.5"><Phone className="size-3.5" /> {guardian.phone}</span>}
                                            <span className="flex items-center gap-1.5"><CalendarDays className="size-3.5" /> {t("staff.guardians.born", {date: formatDateOnly(guardian.dateOfBirth, locale)})}</span>
                                        </div>
                                    </div>
                                    <div className="flex shrink-0 gap-2 self-end sm:self-auto">
                                        {canUpdate && <Button size="icon-sm" variant="ghost" aria-label={t("staff.guardians.editLabel", {name: fullName})} onClick={() => openEdit(guardian)}><Pencil /></Button>}
                                        {canDelete && <Button size="icon-sm" variant="destructive-outline" aria-label={t("staff.guardians.deleteLabel", {name: fullName})} onClick={() => openDelete(guardian)}><Trash2 /></Button>}
                                    </div>
                                </div>
                            )
                        })}
                    </div>
                    {filteredGuardians.length > pageSize && (
                        <div className="flex flex-col gap-3 border-t bg-muted/20 px-4 py-3 sm:flex-row sm:items-center sm:justify-between sm:px-5">
                            <p className="text-xs text-muted-foreground">{t("staff.guardians.pagination.showing", {from: firstVisibleGuardian, to: lastVisibleGuardian, total: filteredGuardians.length})}</p>
                            <div className="flex items-center justify-between gap-2 sm:justify-end">
                                <Button size="sm" variant="outline" disabled={currentPage === 1} onClick={() => setRequestedPage(Math.max(1, currentPage - 1))}><ChevronLeft /> {t("staff.guardians.pagination.previous")}</Button>
                                <span className="min-w-20 text-center text-sm tabular-nums">{t("staff.guardians.pagination.page", {current: currentPage, total: pageCount})}</span>
                                <Button size="sm" variant="outline" disabled={currentPage === pageCount} onClick={() => setRequestedPage(Math.min(pageCount, currentPage + 1))}>{t("staff.guardians.pagination.next")} <ChevronRight /></Button>
                            </div>
                        </div>
                    )}
                </Card>
            )}

            <Dialog open={createOpen} onOpenChange={(open) => { if (!saving) setCreateOpen(open) }}>
                <DialogPopup className="sm:max-w-2xl">
                    <DialogHeader><DialogTitle>{t("staff.guardians.createTitle")}</DialogTitle><DialogDescription>{t("staff.guardians.createDescription")}</DialogDescription></DialogHeader>
                    <GuardianForm draft={draft} saving={saving} mode="create" onChange={setDraft} onSubmit={createGuardian} />
                </DialogPopup>
            </Dialog>

            <Dialog open={editTarget !== null} onOpenChange={(open) => { if (!open && !saving) setEditTarget(null) }}>
                <DialogPopup className="sm:max-w-2xl">
                    <DialogHeader><DialogTitle>{t("staff.guardians.editTitle")}</DialogTitle><DialogDescription>{t("staff.guardians.editDescription")}</DialogDescription></DialogHeader>
                    <GuardianForm draft={draft} saving={saving} mode="edit" wasAccountEnabled={editTarget?.accountEnabled} onChange={setDraft} onSubmit={updateGuardian} />
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
                                <p className="font-semibold text-destructive">{t("staff.guardians.permanentDeletion")}</p>
                                <p className="mt-1 text-sm leading-6 text-muted-foreground">{t("staff.guardians.permanentDeletionWarning")}</p>
                            </div>
                        </div>
                        <AlertDialogTitle>{t("staff.guardians.deleteTitle", {name: requiredDeleteConfirmation})}</AlertDialogTitle>
                        <AlertDialogDescription>{t("staff.guardians.deleteConfirmation")}</AlertDialogDescription>
                        <Field className="pt-2">
                            <FieldLabel>{t("staff.guardians.typeName", {name: requiredDeleteConfirmation})}</FieldLabel>
                            <Input value={deleteConfirmation} onChange={(event) => setDeleteConfirmation(event.target.value)} disabled={deleting} autoComplete="off" autoFocus />
                        </Field>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogClose render={<Button variant="outline" disabled={deleting} />}>{t("staff.guardians.cancel")}</AlertDialogClose>
                        <Button variant="destructive" loading={deleting} disabled={deleteConfirmation !== requiredDeleteConfirmation || deleting} onClick={deleteGuardian}>{t("staff.guardians.deletePermanently")}</Button>
                    </AlertDialogFooter>
                </AlertDialogPopup>
            </AlertDialog>
        </div>
    )
}

function SummaryCard({icon: Icon, label, value}: {icon: typeof UsersRound; label: string; value: number}) {
    return (
        <Card className="py-4">
            <CardContent className="flex items-center gap-3 px-4">
                <span className="flex size-9 items-center justify-center rounded-full bg-primary/10 text-primary"><Icon className="size-4.5" /></span>
                <div><p className="text-xl font-semibold tabular-nums">{value}</p><p className="text-xs text-muted-foreground">{label}</p></div>
            </CardContent>
        </Card>
    )
}

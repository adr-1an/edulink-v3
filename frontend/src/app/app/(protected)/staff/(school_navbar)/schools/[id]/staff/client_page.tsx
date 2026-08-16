"use client"

import React, {useMemo, useState} from "react"
import {useRouter} from "next/navigation"
import {toast} from "sonner"
import PageTitle from "@/components/app/page_title"
import UserAvatar from "@/components/app/user_avatar"
import LocalDateTime from "@/components/local-date-time"
import {Button} from "@/components/ui/button"
import {Card} from "@/components/ui/card"
import {Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle} from "@/components/ui/empty"
import {CalendarDays, ChevronRight, Clock3, Inbox, LoaderCircle, Mail, MailPlus, Plus, RotateCcw, Shield, Trash2, UserRoundPlus, Users, X} from "lucide-react"
import {
    AlertDialog, AlertDialogClose, AlertDialogDescription, AlertDialogFooter,
    AlertDialogHeader, AlertDialogPopup, AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import {
    handleAddStaffRole, handleCancelSchoolInvitation, handleDeleteStaffMember, handleInviteStaffMember,
    handleListSchoolInvitations, handleRemoveStaffRole, type EmailImportance, type SchoolInvitation,
    type StaffActionErrorCode,
} from "./actions"
import {
    Dialog, DialogClose, DialogDescription, DialogFooter, DialogHeader,
    DialogPanel, DialogPopup, DialogTitle,
} from "@/components/ui/dialog"
import {Form} from "@/components/ui/form"
import {Field, FieldDescription, FieldLabel} from "@/components/ui/field"
import {Input} from "@/components/ui/input"
import {Combobox, ComboboxEmpty, ComboboxInput, ComboboxItem, ComboboxList, ComboboxPopup} from "@/components/ui/combobox"
import {Select, SelectItem, SelectPopup, SelectTrigger, SelectValue} from "@/components/ui/select"
import {canManagePosition, hasSchoolPermission, type SchoolAccess} from "@/lib/school_access"
import {Badge} from "@/components/ui/badge"
import {Popover, PopoverDescription, PopoverPopup, PopoverTitle, PopoverTrigger} from "@/components/ui/popover"
import {useLocale} from "@/i18n/provider"
import {type MessageKey} from "@/i18n/messages"

export interface UserSummary {id: string; name: string; email: string; profilePictureURL: string | null}
export interface StaffRole {id: string; position: number; name: string; color: string}
export interface StaffMember {
    id: string
    userId: string
    user: UserSummary
    addedBy: UserSummary
    createdAt: string
    roles: StaffRole[]
}
type InvitationView = SchoolInvitation & {expired: boolean}

const errorMessageKeys = {
    invalid_school: "staff.members.error.invalidSchool",
    invalid_invitation: "staff.members.error.invalidInvitation",
    invalid_importance: "staff.members.error.invalidImportance",
    invalid_email: "staff.members.error.invalidEmail",
    invalid_member: "staff.members.error.invalidMember",
    invalid_member_role: "staff.members.error.invalidMemberRole",
    network: "staff.members.error.network",
    unauthorized: "staff.members.error.unauthorized",
    list_invitations_forbidden: "staff.members.error.listInvitationsForbidden",
    cancel_invitation_forbidden: "staff.members.error.cancelInvitationForbidden",
    invite_forbidden: "staff.members.error.inviteForbidden",
    remove_member_forbidden: "staff.members.error.removeMemberForbidden",
    assign_role_forbidden: "staff.members.error.assignRoleForbidden",
    remove_role_forbidden: "staff.members.error.removeRoleForbidden",
    privacy_restricted: "staff.members.error.privacyRestricted",
    invitation_conflict: "staff.members.error.invitationConflict",
    already_staff: "staff.members.error.alreadyStaff",
    member_missing: "staff.members.error.memberMissing",
    member_role_missing: "staff.members.error.memberRoleMissing",
    server: "staff.members.error.server",
    load_invitations: "staff.members.error.loadInvitations",
    cancel_invitation: "staff.members.error.cancelInvitation",
    send_invitation: "staff.members.error.sendInvitation",
    remove_member: "staff.members.error.removeMember",
    assign_role: "staff.members.error.assignRole",
    remove_role: "staff.members.error.removeRole",
} satisfies Record<StaffActionErrorCode, MessageKey>

function roleColor(color: string) {
    return color.startsWith("#") ? color : `#${color}`
}

export default function StaffClientPage({schoolID, staff, availableRoles, canListRoles, access}: {
    schoolID: string
    staff: StaffMember[]
    availableRoles: StaffRole[]
    canListRoles: boolean
    access: SchoolAccess
}) {
    const router = useRouter()
    const {t} = useLocale()
    const [members, setMembers] = useState(staff)
    const [selectedStaffID, setSelectedStaffID] = useState<string | null>(null)
    const [deleteTarget, setDeleteTarget] = useState<StaffMember | null>(null)
    const [deleting, setDeleting] = useState(false)
    const [roleRemovalTarget, setRoleRemovalTarget] = useState<{memberID: string; memberName: string; role: StaffRole} | null>(null)
    const [removingRole, setRemovingRole] = useState(false)
    const [assigningRoleID, setAssigningRoleID] = useState<string | null>(null)
    const [addRoleOpen, setAddRoleOpen] = useState(false)
    const [inviteOpen, setInviteOpen] = useState(false)
    const [inviteEmail, setInviteEmail] = useState("")
    const [inviteImportance, setInviteImportance] = useState<EmailImportance>("normal")
    const [inviting, setInviting] = useState(false)
    const [invitationsOpen, setInvitationsOpen] = useState(false)
    const [invitations, setInvitations] = useState<InvitationView[]>([])
    const [loadingInvitations, setLoadingInvitations] = useState(false)
    const [invitationsError, setInvitationsError] = useState<string | null>(null)
    const [cancelInvitationTarget, setCancelInvitationTarget] = useState<InvitationView | null>(null)
    const [cancelingInvitation, setCancelingInvitation] = useState(false)
    const canInviteStaff = hasSchoolPermission(access, "staff.create")
    const canListInvitations = hasSchoolPermission(access, "school.invite.list")
    const canCancelInvitations = hasSchoolPermission(access, "school.invite.cancel")
    const canDeleteStaff = hasSchoolPermission(access, "staff.delete")
    const hasAssignRolePermission = hasSchoolPermission(access, "staff.role.add")
    const canAssignRoles = canListRoles && hasAssignRolePermission
    const canRemoveRoles = hasSchoolPermission(access, "staff.role.remove")
    const pendingInvitations = useMemo(() => invitations
        .filter((invitation) => invitation.status === "pending")
        .sort((first, second) => new Date(second.createdAt).getTime() - new Date(first.createdAt).getTime()), [invitations])

    const selectedMember = members.find((member) => member.id === selectedStaffID) ?? null
    const unassignedRoles = useMemo(() => {
        if (!selectedMember) return []
        const assigned = new Set(selectedMember.roles.map((role) => role.id))
        return availableRoles
            .filter((role) => !assigned.has(role.id) && canManagePosition(access, role.position))
            .sort((a, b) => b.position - a.position)
    }, [access, availableRoles, selectedMember])
    const canDeleteSelectedMember = selectedMember !== null
        && canDeleteStaff
        && canManagePosition(access, Math.max(-1, ...selectedMember.roles.map((role) => role.position)))

    async function inviteMember(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        setInviting(true)
        const res = await handleInviteStaffMember(schoolID, inviteEmail, inviteImportance)
        setInviting(false)
        if (!res.ok) return toast.error(t(errorMessageKeys[res.code]))

        setInviteEmail("")
        setInviteImportance("normal")
        setInviteOpen(false)
        if (res.outcome === "staff_added") {
            toast.success(t("staff.members.ownerAddedAsStaff"))
            router.refresh()
            return
        }
        toast.success(t("staff.members.invitationSent"))
    }

    async function loadInvitations() {
        setLoadingInvitations(true)
        setInvitationsError(null)
        const res = await handleListSchoolInvitations(schoolID)
        setLoadingInvitations(false)
        if (!res.ok) {
            setInvitationsError(t(errorMessageKeys[res.code]))
            return
        }
        setInvitations(res.data.map((invitation) => ({
            ...invitation,
            expired: new Date(invitation.expiresAt).getTime() <= Date.now(),
        })))
    }

    function openInvitations() {
        setInvitationsOpen(true)
        void loadInvitations()
    }

    async function cancelInvitation() {
        if (!cancelInvitationTarget) return
        setCancelingInvitation(true)
        const res = await handleCancelSchoolInvitation(cancelInvitationTarget.id)
        setCancelingInvitation(false)
        if (!res.ok) return toast.error(t(errorMessageKeys[res.code]))

        setInvitations((current) => current.filter((invitation) => invitation.id !== cancelInvitationTarget.id))
        setCancelInvitationTarget(null)
        toast.success(t("staff.members.invitationCanceled"))
    }

    async function deleteMember() {
        if (!deleteTarget) return
        setDeleting(true)
        const res = await handleDeleteStaffMember(deleteTarget.id)
        setDeleting(false)
        if (!res.ok) return toast.error(t(errorMessageKeys[res.code]))

        setMembers((current) => current.filter((member) => member.id !== deleteTarget.id))
        setSelectedStaffID(null)
        setDeleteTarget(null)
        toast.success(t("staff.members.memberRemoved"))
    }

    async function addRole(role: StaffRole) {
        if (!selectedMember) return
        setAssigningRoleID(role.id)
        const res = await handleAddStaffRole(selectedMember.id, role.id)
        setAssigningRoleID(null)
        if (!res.ok) return toast.error(t(errorMessageKeys[res.code]))

        setMembers((current) => current.map((member) => member.id === selectedMember.id
            ? {...member, roles: [...member.roles, role].sort((a, b) => b.position - a.position)}
            : member
        ))
        setAddRoleOpen(false)
        toast.success(t("staff.members.roleAdded", {role: role.name, name: selectedMember.user.name}))
    }

    async function removeRole() {
        if (!roleRemovalTarget) return
        setRemovingRole(true)
        const res = await handleRemoveStaffRole(roleRemovalTarget.memberID, roleRemovalTarget.role.id)
        setRemovingRole(false)
        if (!res.ok) return toast.error(t(errorMessageKeys[res.code]))

        setMembers((current) => current.map((member) => member.id === roleRemovalTarget.memberID
            ? {...member, roles: member.roles.filter((role) => role.id !== roleRemovalTarget.role.id)}
            : member
        ))
        toast.success(t("staff.members.roleRemoved", {role: roleRemovalTarget.role.name, name: roleRemovalTarget.memberName}))
        setRoleRemovalTarget(null)
    }

    return (
        <div className="space-y-6">
            <div className="flex flex-wrap items-end justify-between gap-3">
                <div>
                    <PageTitle centered={false}>{t("staff.members.title")}</PageTitle>
                    <p className="text-muted-foreground">{t("staff.members.description")}</p>
                </div>
                <div className="flex flex-wrap gap-2">
                    {canListInvitations && <Button variant="outline" onClick={openInvitations}><Inbox /> {t("staff.members.pendingInvitations")}</Button>}
                    {canInviteStaff && <Button onClick={() => setInviteOpen(true)}><MailPlus /> {t("staff.members.invite")}</Button>}
                </div>
            </div>

            {members.length ? (
                <Card className="overflow-hidden p-0">
                    <div className="divide-y">
                        {members.map((member) => (
                            <button
                                key={member.id}
                                type="button"
                                className="group flex w-full items-center gap-3 px-4 py-3 text-left transition-colors hover:bg-accent/55 focus-visible:bg-accent/55 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset sm:px-5"
                                onClick={() => setSelectedStaffID(member.id)}
                            >
                                <UserAvatar
                                    name={member.user.name}
                                    src={member.user.profilePictureURL}
                                    cacheKey={`staff:${member.user.id}`}
                                    className="size-10 ring-1 ring-border"
                                />
                                <div className="min-w-0 flex-1">
                                    <p className="truncate font-medium">{member.user.name}</p>
                                    <div className="mt-1 flex items-center gap-2 text-xs text-muted-foreground">
                                        <div className="flex -space-x-1" aria-hidden="true">
                                            {member.roles.slice(0, 5).map((role) => (
                                                <span key={role.id} className="size-3 rounded-full border-2 border-card" style={{backgroundColor: roleColor(role.color)}} />
                                            ))}
                                        </div>
                                        <span>{member.roles.length
                                            ? t(member.roles.length === 1 ? "staff.members.roleOne" : "staff.members.roleOther", {count: member.roles.length})
                                            : t("staff.members.noRoles")}</span>
                                    </div>
                                </div>
                                <ChevronRight className="size-5 text-muted-foreground transition-transform group-hover:translate-x-0.5" />
                            </button>
                        ))}
                    </div>
                </Card>
            ) : (
                <Card>
                    <Empty>
                        <EmptyHeader>
                            <EmptyMedia variant="icon"><Users /></EmptyMedia>
                            <EmptyTitle>{t("staff.members.empty")}</EmptyTitle>
                            <EmptyDescription>{t("staff.members.emptyDescription")}</EmptyDescription>
                        </EmptyHeader>
                    </Empty>
                </Card>
            )}

            <Dialog open={selectedMember !== null} onOpenChange={(open) => {
                if (!open && !deleting && !removingRole) {
                    setSelectedStaffID(null)
                    setAddRoleOpen(false)
                }
            }}>
                <DialogPopup className="overflow-hidden sm:max-w-xl">
                    {selectedMember && <>
                        <div className="h-20 bg-linear-to-br from-primary/30 via-primary/15 to-transparent" />
                        <DialogHeader className="relative pt-0">
                            <UserAvatar
                                name={selectedMember.user.name}
                                src={selectedMember.user.profilePictureURL}
                                cacheKey={`staff:${selectedMember.user.id}`}
                                className="-mt-9 size-20 border-4 border-popover shadow-sm"
                                fallbackClassName="text-xl"
                            />
                            <div className="pt-1">
                                <DialogTitle>{selectedMember.user.name}</DialogTitle>
                                <DialogDescription>{selectedMember.user.email}</DialogDescription>
                            </div>
                        </DialogHeader>

                        <DialogPanel className="space-y-5">
                            <section className="space-y-2.5">
                                <div className="flex items-center justify-between gap-3">
                                    <h3 className="text-sm font-semibold uppercase tracking-wide text-muted-foreground">{t("staff.members.roles")}</h3>
                                    {canAssignRoles && <div className="w-52">
                                        <Combobox<StaffRole>
                                            items={unassignedRoles}
                                            itemToStringLabel={(role) => role?.name ?? ""}
                                            itemToStringValue={(role) => role?.id ?? ""}
                                            value={null}
                                            open={addRoleOpen}
                                            onOpenChange={setAddRoleOpen}
                                            onValueChange={(role) => {
                                                if (role) void addRole(role)
                                            }}
                                            disabled={assigningRoleID !== null}
                                        >
                                            <ComboboxInput
                                                size="sm"
                                                placeholder={t("staff.members.searchRoles")}
                                                startAddon={assigningRoleID ? <LoaderCircle className="animate-spin" /> : <Plus />}
                                            />
                                            <ComboboxPopup align="end" className="w-72">
                                                <ComboboxEmpty>
                                                    {t(unassignedRoles.length ? "staff.members.noMatchingRoles" : "staff.members.allRolesAssigned")}
                                                </ComboboxEmpty>
                                                <ComboboxList>
                                                    {(role: StaffRole) => (
                                                        <ComboboxItem key={role.id} value={role}>
                                                            <span className="flex min-w-0 items-center gap-2">
                                                                <span className="size-3 shrink-0 rounded-full" style={{backgroundColor: roleColor(role.color)}} />
                                                                <span className="truncate">{role.name}</span>
                                                            </span>
                                                        </ComboboxItem>
                                                    )}
                                                </ComboboxList>
                                            </ComboboxPopup>
                                        </Combobox>
                                    </div>}
                                </div>

                                {selectedMember.roles.length ? (
                                    <div className="flex flex-wrap gap-2">
                                        {selectedMember.roles.map((role) => (
                                            <div key={role.id} className="inline-flex items-center gap-1.5 rounded-md border bg-muted/35 py-1 pr-2.5 pl-1.5 text-sm">
                                                {canRemoveRoles && canManagePosition(access, role.position) ? <button
                                                    type="button"
                                                    className="group relative size-4 shrink-0 rounded-full focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                                                    style={{backgroundColor: roleColor(role.color)}}
                                                    aria-label={t("staff.members.removeRoleLabel", {role: role.name})}
                                                    title={t("staff.members.removeRoleLabel", {role: role.name})}
                                                    onClick={() => setRoleRemovalTarget({memberID: selectedMember.id, memberName: selectedMember.user.name, role})}
                                                >
                                                    <span className="absolute inset-0 rounded-full bg-destructive opacity-0 transition-opacity group-hover:opacity-100 group-focus-visible:opacity-100" />
                                                    <X className="absolute inset-0 m-auto size-3 text-white opacity-0 transition-opacity group-hover:opacity-100 group-focus-visible:opacity-100" />
                                                </button> : <span className="size-4 shrink-0 rounded-full" style={{backgroundColor: roleColor(role.color)}} />}
                                                <span>{role.name}</span>
                                            </div>
                                        ))}
                                    </div>
                                ) : (
                                    <div className="flex items-center gap-2 rounded-lg border border-dashed p-3 text-sm text-muted-foreground">
                                        <Shield className="size-4" /> {t("staff.members.noRolesAssigned")}
                                    </div>
                                )}
                                {!canListRoles && hasAssignRolePermission && <p className="text-xs text-muted-foreground">{t("staff.members.roleListRestricted")}</p>}
                            </section>

                            <section className="space-y-2.5">
                                <h3 className="text-sm font-semibold uppercase tracking-wide text-muted-foreground">{t("staff.members.details")}</h3>
                                <div className="grid gap-2 sm:grid-cols-2">
                                    <div className="flex gap-2.5 rounded-lg border bg-muted/25 p-3">
                                        <Mail className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
                                        <div className="min-w-0">
                                            <p className="text-xs text-muted-foreground">{t("staff.members.email")}</p>
                                            <p className="truncate text-sm">{selectedMember.user.email}</p>
                                        </div>
                                    </div>
                                    <div className="flex gap-2.5 rounded-lg border bg-muted/25 p-3">
                                        <CalendarDays className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
                                        <div>
                                            <p className="text-xs text-muted-foreground">{t("staff.members.joined")}</p>
                                            <p className="text-sm"><LocalDateTime value={selectedMember.createdAt} precision="date" /></p>
                                        </div>
                                    </div>
                                    <div className="flex gap-2.5 rounded-lg border bg-muted/25 p-3 sm:col-span-2">
                                        <UserRoundPlus className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
                                        <div className="min-w-0">
                                            <p className="text-xs text-muted-foreground">{t("staff.members.addedBy")}</p>
                                            <p className="truncate text-sm">{selectedMember.addedBy.name} · {selectedMember.addedBy.email}</p>
                                        </div>
                                    </div>
                                </div>
                            </section>
                        </DialogPanel>

                        <DialogFooter className={canDeleteSelectedMember ? "justify-between sm:justify-between" : undefined}>
                            {canDeleteSelectedMember && <Button variant="destructive-outline" onClick={() => setDeleteTarget(selectedMember)}><Trash2 /> {t("staff.members.removeMember")}</Button>}
                            <DialogClose render={<Button variant="outline" />}>{t("staff.members.close")}</DialogClose>
                        </DialogFooter>
                    </>}
                </DialogPopup>
            </Dialog>

            <Dialog open={inviteOpen} onOpenChange={(open) => {
                if (!inviting) setInviteOpen(open)
            }}>
                <DialogPopup className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle>{t("staff.members.inviteTitle")}</DialogTitle>
                        <DialogDescription>{t("staff.members.inviteDescription")}</DialogDescription>
                    </DialogHeader>
                    <Form className="contents" onSubmit={inviteMember}>
                        <DialogPanel className="grid gap-4">
                            <Field>
                                <FieldLabel>{t("staff.members.emailAddress")}</FieldLabel>
                                <Input
                                    type="email"
                                    value={inviteEmail}
                                    onChange={(event) => setInviteEmail(event.target.value)}
                                    placeholder={t("staff.members.emailPlaceholder")}
                                    required
                                    autoFocus
                                />
                                {access.owner && (
                                    <FieldDescription className="flex items-start gap-2 rounded-lg border border-primary/15 bg-primary/5 p-3 text-foreground">
                                        <UserRoundPlus className="mt-0.5 size-4 shrink-0 text-primary" />
                                        <span>{t("staff.members.ownerSelfInviteHelp")}</span>
                                    </FieldDescription>
                                )}
                            </Field>
                            <Field>
                                <FieldLabel>{t("staff.members.importance")}</FieldLabel>
                                <Select
                                    items={[
                                        {label: t("staff.members.importance.nonUrgent"), value: "non-urgent"},
                                        {label: t("staff.members.importance.low"), value: "low"},
                                        {label: t("staff.members.importance.normal"), value: "normal"},
                                        {label: t("staff.members.importance.high"), value: "high"},
                                        {label: t("staff.members.importance.urgent"), value: "urgent"},
                                    ]}
                                    value={inviteImportance}
                                    onValueChange={(value) => setInviteImportance((value ?? "normal") as EmailImportance)}
                                >
                                    <SelectTrigger aria-label={t("staff.members.importanceAria")}>
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectPopup>
                                        <SelectItem value="non-urgent">{t("staff.members.importance.nonUrgent")}</SelectItem>
                                        <SelectItem value="low">{t("staff.members.importance.low")}</SelectItem>
                                        <SelectItem value="normal">{t("staff.members.importance.normal")}</SelectItem>
                                        <SelectItem value="high">{t("staff.members.importance.high")}</SelectItem>
                                        <SelectItem value="urgent">{t("staff.members.importance.urgent")}</SelectItem>
                                    </SelectPopup>
                                </Select>
                                <FieldDescription>
                                    {t("staff.members.importanceHelp")}
                                </FieldDescription>
                            </Field>
                        </DialogPanel>
                        <DialogFooter>
                            <DialogClose render={<Button variant="outline" disabled={inviting} />}>{t("staff.members.cancel")}</DialogClose>
                            <Button type="submit" loading={inviting} disabled={!inviteEmail.trim()}>{t("staff.members.sendInvitation")}</Button>
                        </DialogFooter>
                    </Form>
                </DialogPopup>
            </Dialog>

            <Dialog open={invitationsOpen} onOpenChange={(open) => {
                if (!cancelingInvitation) setInvitationsOpen(open)
            }}>
                <DialogPopup className="sm:max-w-2xl">
                    <DialogHeader>
                        <DialogTitle>{t("staff.members.invitationsTitle")}</DialogTitle>
                        <DialogDescription>{t("staff.members.invitationsDescription")}</DialogDescription>
                    </DialogHeader>
                    <DialogPanel className="space-y-3">
                        {loadingInvitations ? (
                            <div className="flex min-h-40 items-center justify-center gap-2 text-sm text-muted-foreground">
                                <LoaderCircle className="animate-spin" /> {t("staff.members.loadingInvitations")}
                            </div>
                        ) : invitationsError ? (
                            <div className="flex min-h-40 flex-col items-center justify-center gap-3 rounded-xl border border-destructive/15 bg-destructive/5 p-6 text-center">
                                <p className="text-sm">{invitationsError}</p>
                                <Button size="sm" variant="outline" onClick={() => void loadInvitations()}><RotateCcw /> {t("staff.members.tryAgain")}</Button>
                            </div>
                        ) : pendingInvitations.length === 0 ? (
                            <div className="flex min-h-40 flex-col items-center justify-center gap-2 rounded-xl border border-dashed p-6 text-center">
                                <Inbox className="size-6 text-muted-foreground" />
                                <p className="font-medium">{t("staff.members.noInvitations")}</p>
                                <p className="text-sm text-muted-foreground">{t("staff.members.noInvitationsDescription")}</p>
                            </div>
                        ) : (
                            <div className="divide-y overflow-hidden rounded-xl border">
                                {pendingInvitations.map((invitation) => {
                                    return (
                                        <div className="flex flex-col gap-3 p-4 sm:flex-row sm:items-center" key={invitation.id}>
                                            <div className="min-w-0 flex-1">
                                                <div className="flex flex-wrap items-center gap-2">
                                                    <p className="truncate font-medium">{invitation.userEmail}</p>
                                                    {invitation.expired && <Badge variant="warning">{t("staff.members.expired")}</Badge>}
                                                </div>
                                                <div className="mt-1 flex items-center gap-1 text-xs text-muted-foreground">
                                                    <span>{t("staff.members.sentBy")}</span>
                                                    <Popover>
                                                        <PopoverTrigger
                                                            className="rounded-sm font-medium text-foreground underline decoration-muted-foreground/50 underline-offset-4 outline-none hover:text-primary focus-visible:ring-2 focus-visible:ring-ring"
                                                            type="button"
                                                        >
                                                            {invitation.addedBy.name}
                                                        </PopoverTrigger>
                                                        <PopoverPopup className="w-72" align="start">
                                                            <div className="flex min-w-0 items-center gap-3">
                                                                <UserAvatar
                                                                    name={invitation.addedBy.name}
                                                                    className="size-10 border"
                                                                />
                                                                <div className="min-w-0">
                                                                    <PopoverTitle className="truncate text-base">{invitation.addedBy.name}</PopoverTitle>
                                                                    <PopoverDescription className="mt-1 flex items-center gap-1.5">
                                                                        <Mail className="size-3.5 shrink-0" />
                                                                        <span className="truncate">{invitation.addedBy.email}</span>
                                                                    </PopoverDescription>
                                                                </div>
                                                            </div>
                                                        </PopoverPopup>
                                                    </Popover>
                                                </div>
                                                <p className="mt-1 flex items-center gap-1.5 text-xs text-muted-foreground">
                                                    <MailPlus className="size-3.5" /> {t("staff.members.sent")} <LocalDateTime value={invitation.createdAt} />
                                                </p>
                                                <p className="mt-1 flex items-center gap-1.5 text-xs text-muted-foreground">
                                                    <Clock3 className="size-3.5" /> {t(invitation.expired ? "staff.members.expired" : "staff.members.expires")} <LocalDateTime value={invitation.expiresAt} />
                                                </p>
                                            </div>
                                            {canCancelInvitations && (
                                                <Button size="sm" variant="destructive-outline" onClick={() => setCancelInvitationTarget(invitation)}>
                                                    {t("staff.members.cancelInvitation")}
                                                </Button>
                                            )}
                                        </div>
                                    )
                                })}
                            </div>
                        )}
                    </DialogPanel>
                    <DialogFooter>
                        <DialogClose render={<Button variant="outline" disabled={cancelingInvitation} />}>{t("staff.members.close")}</DialogClose>
                    </DialogFooter>
                </DialogPopup>
            </Dialog>

            <AlertDialog open={roleRemovalTarget !== null} onOpenChange={(open) => {
                if (!open && !removingRole) setRoleRemovalTarget(null)
            }}>
                <AlertDialogPopup>
                    <AlertDialogHeader>
                        <AlertDialogTitle>{t("staff.members.removeRoleTitle")}</AlertDialogTitle>
                        <AlertDialogDescription>{t("staff.members.removeRoleDescription", {
                            name: roleRemovalTarget?.memberName ?? "",
                            role: roleRemovalTarget?.role.name ?? "",
                        })}</AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogClose render={<Button variant="outline" disabled={removingRole} />}>{t("staff.members.cancel")}</AlertDialogClose>
                        <Button variant="destructive" loading={removingRole} onClick={removeRole}>{t("staff.members.removeRole")}</Button>
                    </AlertDialogFooter>
                </AlertDialogPopup>
            </AlertDialog>

            <AlertDialog open={cancelInvitationTarget !== null} onOpenChange={(open) => {
                if (!open && !cancelingInvitation) setCancelInvitationTarget(null)
            }}>
                <AlertDialogPopup>
                    <AlertDialogHeader>
                        <AlertDialogTitle>{t("staff.members.cancelInvitationTitle")}</AlertDialogTitle>
                        <AlertDialogDescription>{t("staff.members.cancelInvitationDescription", {
                            email: cancelInvitationTarget?.userEmail ?? "",
                        })}</AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogClose render={<Button variant="outline" disabled={cancelingInvitation} />}>{t("staff.members.keepInvitation")}</AlertDialogClose>
                        <Button variant="destructive" loading={cancelingInvitation} onClick={cancelInvitation}>{t("staff.members.cancelInvitation")}</Button>
                    </AlertDialogFooter>
                </AlertDialogPopup>
            </AlertDialog>

            <AlertDialog open={deleteTarget !== null} onOpenChange={(open) => {
                if (!open && !deleting) setDeleteTarget(null)
            }}>
                <AlertDialogPopup>
                    <AlertDialogHeader>
                        <AlertDialogTitle>{t("staff.members.removeMemberTitle")}</AlertDialogTitle>
                        <AlertDialogDescription>{t("staff.members.removeMemberDescription", {
                            name: deleteTarget?.user.name ?? "",
                        })}</AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogClose render={<Button variant="outline" disabled={deleting} />}>{t("staff.members.cancel")}</AlertDialogClose>
                        <Button variant="destructive" loading={deleting} onClick={deleteMember}>{t("staff.members.removeMember")}</Button>
                    </AlertDialogFooter>
                </AlertDialogPopup>
            </AlertDialog>
        </div>
    )
}

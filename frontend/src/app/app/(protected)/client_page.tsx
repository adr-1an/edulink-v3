"use client"

import PageTitle from "@/components/app/page_title"
import React, {useMemo, useState} from "react"
import Link from "next/link"
import {
    handleCreateSchool,
    handleFetchStaffInvitations,
    handleLeaveSchool,
    handleStaffInvitation,
} from "@/app/app/(protected)/actions"
import {useRouter} from "next/navigation"
import ReactCountryFlag from "react-country-flag"
import LocalDateTime from "@/components/local-date-time"
import {toast} from "sonner"
import UserAvatar from "@/components/app/user_avatar"
import {
    Sheet, SheetClose,
    SheetDescription,
    SheetFooter,
    SheetHeader,
    SheetPanel,
    SheetPopup,
    SheetTitle
} from "@/components/ui/sheet";
import {Menu, MenuItem, MenuPopup, MenuTrigger} from "@/components/ui/menu";
import {Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle} from "@/components/ui/card";
import {Badge} from "@/components/ui/badge";
import {Input} from "@/components/ui/input";
import {Separator} from "@/components/ui/separator";
import {Button} from "@/components/ui/button";
import {Bell, Plus, RefreshCw, Search, School as SchoolIcon, TriangleAlert, X} from "lucide-react";
import {Form} from "@/components/ui/form";
import {Field, FieldLabel} from "@/components/ui/field";
import countries, {Country} from "world-countries"
import {
    Combobox,
    ComboboxEmpty,
    ComboboxInput,
    ComboboxItem,
    ComboboxList,
    ComboboxPopup
} from "@/components/ui/combobox";
import {TurnstileWidget, useTurnstile} from "@/components/turnstile"
import {
    AlertDialog, AlertDialogClose, AlertDialogDescription, AlertDialogFooter,
    AlertDialogHeader, AlertDialogPopup, AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import {type ProfilePicture} from "@/lib/profile_picture"
import {useLocale} from "@/i18n/provider"

export interface StaffInvitation {
    id: string
    school: {
        id: string
        name: string
        regionCode: string
    }
    sentBy: {
        id: string
        name: string
        email: string
    }
    status: string
    createdAt: string
    expiresAt: string
}

interface User {
    id: string
    name: string
    profilePicture: ProfilePicture | null
}

interface School {
    id: string
    ownerId: string
    name: string
    regionCode: string
}

interface Props {
    user: User
    schools: School[]
    regionNames: Record<string, string>
    initialInvitations: StaffInvitation[]
    invitationsInitiallyUnavailable: boolean
}

export default function ClientPage({
    user,
    schools,
    regionNames,
    initialInvitations,
    invitationsInitiallyUnavailable,
}: Props) {
    const router = useRouter()
    const {locale, t} = useLocale()
    const [leftSchoolIDs, setLeftSchoolIDs] = useState<Set<string>>(() => new Set())
    const [searchQuery, setSearchQuery] = useState("")
    const [invitesOpen, setInvitesOpen] = useState(false)
    const [invites, setInvites] = useState<StaffInvitation[]>(initialInvitations)
    const [fetchingInvites, setFetchingInvites] = useState(false)
    const [invitationsUnavailable, setInvitationsUnavailable] = useState(invitationsInitiallyUnavailable)
    const [updatingInvitation, setUpdatingInvitation] = useState<string | null>(null)
    const [creatingSchool, setCreatingSchool] = useState(false)
    const [submittingSchool, setSubmittingSchool] = useState(false)
    const [leaveTarget, setLeaveTarget] = useState<School | null>(null)
    const [leavingSchool, setLeavingSchool] = useState(false)
    const turnstile = useTurnstile()
    const [country, setCountry] = useState<Country | null>(null)
    const countryOptions = useMemo(() => [...countries].sort((first, second) =>
        (regionNames[first.cca2] ?? first.cca2).localeCompare(regionNames[second.cca2] ?? second.cca2, locale)
    ), [locale, regionNames])
    const normalizedSearch = searchQuery.trim().toLocaleLowerCase()
    const availableSchools = schools.filter((school) => !leftSchoolIDs.has(school.id))
    const filteredSchools = availableSchools.filter((school) => {
        if (!normalizedSearch) return true
        const region = regionNames[school.regionCode] ?? ""
        return school.name.toLocaleLowerCase().includes(normalizedSearch)
            || school.regionCode.toLocaleLowerCase().includes(normalizedSearch)
            || region.toLocaleLowerCase().includes(normalizedSearch)
    })
    const ownedSchools = filteredSchools.filter((school) => school.ownerId === user.id)
    const accessibleSchools = filteredSchools.filter((school) => school.ownerId !== user.id)
    const invitationCountLabel = invites.length
        ? t(invites.length === 1 ? "staff.schools.pendingInvitationOne" : "staff.schools.pendingInvitationOther", {count: invites.length})
        : t("staff.schools.invitations")
    const visibleInvitationCount = invites.length > 99 ? "99+" : String(invites.length)

    async function leaveSchool() {
        if (!leaveTarget) return

        setLeavingSchool(true)
        const res = await handleLeaveSchool(leaveTarget.id)
        setLeavingSchool(false)
        if (!res.ok) return toast.error(res.message)

        setLeftSchoolIDs((current) => new Set(current).add(leaveTarget.id))
        setLeaveTarget(null)
        toast.success(t("staff.schools.left"))
        router.refresh()
    }

    async function createSchool(e: React.SubmitEvent<HTMLFormElement>) {
        e.preventDefault()

        const form = e.currentTarget
        const formData = new FormData(form)
        formData.set("regionCode", country?.cca2 ?? "")
        formData.set("cf-turnstile-response", turnstile.token)
        setSubmittingSchool(true)

        const res = await handleCreateSchool(formData)
        setSubmittingSchool(false)

        if (!res.ok) {
            toast.error(t("staff.schools.genericError"))
            turnstile.reset()
            return
        }

        form.reset()
        setCountry(null)
        turnstile.reset()
        setCreatingSchool(false)
        toast.success(t("staff.schools.created"))
        router.refresh()
    }

    async function fetchStaffInvites() {
        setFetchingInvites(true)

        const res = await handleFetchStaffInvitations()
        setFetchingInvites(false)
        if (!res.ok) {
            setInvitationsUnavailable(true)
            toast.error(t("staff.schools.genericError"))
            return
        }

        setInvites(res.data ?? [])
        setInvitationsUnavailable(false)
    }

    async function updateInvitation(invitationID: string, decision: "accept" | "reject") {
        setUpdatingInvitation(`${invitationID}:${decision}`)

        const res = await handleStaffInvitation(invitationID, decision)
        setUpdatingInvitation(null)

        if (!res.ok) {
            toast.error(t("staff.schools.genericError"))
            return
        }

        setInvites((current) => current.filter((invitation) => invitation.id !== invitationID))
        toast.success(t(decision === "accept" ? "staff.schools.invitationAccepted" : "staff.schools.invitationRejected"))

        if (decision === "accept") router.refresh()
    }

    return (
        <div>
            <div className="relative pt-12 md:pt-0">
                <div className="absolute right-0 top-0 z-50 flex items-start gap-2">
                    <>
                        <Button variant="ghost" aria-label={t("staff.schools.create")} title={t("staff.schools.create")} onClick={() => setCreatingSchool(true)}>
                            <Plus />
                        </Button>
                        <Button
                            className="relative"
                            size="icon"
                            variant="ghost"
                            aria-label={invitationCountLabel}
                            title={invitationCountLabel}
                            onClick={() => setInvitesOpen(true)}
                        >
                            <Bell />
                            {invites.length > 0 && (
                                <span
                                    className="absolute -top-1 -right-1 flex h-5 min-w-5 items-center justify-center rounded-full bg-destructive px-1 text-[10px] font-bold leading-none text-white ring-2 ring-background"
                                    aria-hidden="true"
                                >
                                    {visibleInvitationCount}
                                </span>
                            )}
                        </Button>
                        <Menu>
                            <MenuTrigger
                                render={
                                    <button type="button" className="cursor-pointer rounded-full">
                                        <UserAvatar
                                            name={user.name}
                                            src={user.profilePicture?.presignedUrl}
                                            cacheKey={`staff:${user.id}`}
                                        />
                                    </button>
                                }
                            />

                            <MenuPopup align="end">
                                <MenuItem render={<Link href="/app/staff/profile" />}>
                                    {t("staff.schools.profile")}
                                </MenuItem>

                                <MenuItem
                                    className="justify-between"
                                    onClick={() => setInvitesOpen(true)}
                                >
                                    <span>{t("staff.schools.invitations")}</span>
                                    {invites.length > 0 && <Badge variant="destructive" size="sm">{visibleInvitationCount}</Badge>}
                                </MenuItem>
                            </MenuPopup>
                        </Menu>
                    </>
                    </div>
            </div>

            <PageTitle>{t("staff.schools.welcome", {name: user.name})}</PageTitle>

            <section className="mt-6 space-y-4">
                <div className="mx-auto max-w-xl space-y-2">
                    <div className="relative">
                        <Search className="pointer-events-none absolute top-1/2 left-3 z-10 size-4 -translate-y-1/2 text-muted-foreground" />
                        <Input
                            value={searchQuery}
                            onChange={(event) => setSearchQuery(event.target.value)}
                            placeholder={t("staff.schools.searchPlaceholder")}
                            aria-label={t("staff.schools.search")}
                            className="h-10 pr-10 pl-9"
                        />
                        {searchQuery && (
                            <Button
                                className="absolute top-1/2 right-1.5 -translate-y-1/2"
                                size="icon-xs"
                                variant="ghost"
                                aria-label={t("staff.schools.clearSearch")}
                                onClick={() => setSearchQuery("")}
                            >
                                <X />
                            </Button>
                        )}
                    </div>
                    <p className="text-center text-xs text-muted-foreground">
                        {normalizedSearch
                            ? t("staff.schools.matching", {matching: filteredSchools.length, total: availableSchools.length})
                            : t(availableSchools.length === 1 ? "staff.schools.availableOne" : "staff.schools.availableOther", {count: availableSchools.length})}
                    </p>
                </div>

                {filteredSchools.length ? (
                    <div className="space-y-4">
                        {ownedSchools.length > 0 && <SchoolGroup title={t("staff.schools.owned")} schools={ownedSchools} regionNames={regionNames} owned />}
                        {accessibleSchools.length > 0 && <SchoolGroup title={t("staff.schools.accessible")} schools={accessibleSchools} regionNames={regionNames} onLeave={setLeaveTarget} />}
                    </div>
                ) : (
                    <Card className="py-10 text-center">
                        <CardContent className="space-y-3">
                            <span className="mx-auto flex size-10 items-center justify-center rounded-lg bg-muted text-muted-foreground">
                                {normalizedSearch ? <Search /> : <SchoolIcon />}
                            </span>
                            <div>
                                <p className="font-medium">{t(normalizedSearch ? "staff.schools.noneFound" : "staff.schools.none")}</p>
                                <p className="text-sm text-muted-foreground">
                                    {normalizedSearch
                                        ? t("staff.schools.noneFoundDescription")
                                        : t("staff.schools.noneDescription")}
                                </p>
                            </div>
                            {normalizedSearch
                                ? <Button size="sm" variant="outline" onClick={() => setSearchQuery("")}>{t("staff.schools.clear")}</Button>
                                : <Button size="sm" onClick={() => setCreatingSchool(true)}><Plus /> {t("staff.schools.create")}</Button>}
                        </CardContent>
                    </Card>
                )}
            </section>

            <Sheet open={invitesOpen} onOpenChange={setInvitesOpen}>
                <SheetPopup>
                    <SheetHeader>
                        <SheetTitle className="flex items-center gap-2">
                            <span>{t("staff.schools.invitationTitle")}</span>
                            {invites.length > 0 && <Badge variant="destructive">{visibleInvitationCount}</Badge>}
                        </SheetTitle>
                        <SheetDescription>
                            {t("staff.schools.invitationDescription")}
                        </SheetDescription>

                        <Separator />
                    </SheetHeader>

                        {fetchingInvites ? (
                            <CardDescription className="text-center">
                                {t("staff.schools.loadingInvitations")}
                            </CardDescription>
                        ) : invitationsUnavailable ? (
                            <div className="mx-4 flex flex-col items-center gap-3 rounded-xl border border-destructive/20 bg-destructive/5 px-4 py-8 text-center">
                                <span className="flex size-10 items-center justify-center rounded-full bg-destructive/10 text-destructive">
                                    <TriangleAlert className="size-5" />
                                </span>
                                <div>
                                    <p className="font-medium">{t("staff.schools.invitationLoadError")}</p>
                                    <p className="mt-1 text-sm text-muted-foreground">{t("staff.schools.invitationLoadErrorDescription")}</p>
                                </div>
                                <Button size="sm" variant="outline" onClick={() => void fetchStaffInvites()}>
                                    <RefreshCw /> {t("staff.schools.retryInvitations")}
                                </Button>
                            </div>
                        ) : invites.length > 0 ? (
                            <div className="grid gap-2 px-4">
                                {invites.map((inv) => (
                                    <Card key={inv.id}>
                                        <CardHeader>
                                            <CardTitle className="flex items-center gap-2">
                                                {inv.school.regionCode && (
                                                    <ReactCountryFlag svg countryCode={inv.school.regionCode} />
                                                )}
                                                {inv.school.name}
                                            </CardTitle>
                                            <CardDescription>
                                                {t("staff.schools.invitedBy", {name: inv.sentBy.name, email: inv.sentBy.email})}
                                            </CardDescription>
                                        </CardHeader>

                                        <CardContent>
                                            <CardDescription>
                                                {t("staff.schools.received")} <LocalDateTime value={inv.createdAt} />
                                                <br />
                                                {t("staff.schools.expires")} <LocalDateTime value={inv.expiresAt} />
                                            </CardDescription>
                                        </CardContent>

                                        <CardFooter className="justify-end gap-2">
                                            <Button
                                                variant="destructive-outline"
                                                disabled={updatingInvitation !== null}
                                                loading={updatingInvitation === `${inv.id}:reject`}
                                                onClick={() => updateInvitation(inv.id, "reject")}
                                            >
                                                {t("staff.schools.reject")}
                                            </Button>
                                            <Button
                                                disabled={updatingInvitation !== null}
                                                loading={updatingInvitation === `${inv.id}:accept`}
                                                onClick={() => updateInvitation(inv.id, "accept")}
                                            >
                                                {t("staff.schools.accept")}
                                            </Button>
                                        </CardFooter>
                                    </Card>
                                ))}
                            </div>
                        ) : (
                            <CardDescription className="text-center">
                                {t("staff.schools.noInvitations")}
                            </CardDescription>
                        )}
                </SheetPopup>
            </Sheet>

            <AlertDialog open={leaveTarget !== null} onOpenChange={(open) => {
                if (!open && !leavingSchool) setLeaveTarget(null)
            }}>
                <AlertDialogPopup>
                    <AlertDialogHeader>
                        <AlertDialogTitle>{t("staff.schools.leaveTitle", {name: leaveTarget?.name ?? ""})}</AlertDialogTitle>
                        <AlertDialogDescription>
                            {t("staff.schools.leaveDescription")}
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogClose render={<Button variant="outline" disabled={leavingSchool} />}>{t("staff.schools.cancel")}</AlertDialogClose>
                        <Button variant="destructive" loading={leavingSchool} disabled={leavingSchool} onClick={leaveSchool}>
                            {t("staff.schools.leave")}
                        </Button>
                    </AlertDialogFooter>
                </AlertDialogPopup>
            </AlertDialog>

            <Sheet open={creatingSchool} onOpenChange={(open) => {
                if (!submittingSchool) {
                    setCreatingSchool(open)
                    if (!open) turnstile.reset()
                }
            }}>
                <SheetPopup>
                    <SheetHeader>
                        <SheetTitle>
                            {t("staff.schools.createTitle")}
                        </SheetTitle>
                        <Separator />
                    </SheetHeader>

                    <Form onSubmit={createSchool} className="flex min-h-0 flex-1 flex-col">
                        <SheetPanel className="grid flex-1 content-start gap-4">
                            <Field>
                                <FieldLabel>{t("staff.schools.name")}</FieldLabel>
                                <Input name="name" required autoFocus placeholder={t("staff.schools.namePlaceholder")} />
                            </Field>

                            <Field>
                                <FieldLabel>{t("staff.schools.region")}</FieldLabel>
                                <Combobox
                                    required
                                    items={countryOptions}
                                    itemToStringLabel={(country) => country ? regionNames[country.cca2] ?? country.cca2 : ""}
                                    itemToStringValue={(country) => country?.cca2 ?? ""}
                                    value={country}
                                    onValueChange={(e) => setCountry(e)}
                                >
                                    <ComboboxInput name="regionCode" placeholder={t("staff.schools.countryPlaceholder")} required />
                                    <ComboboxPopup>
                                        <ComboboxEmpty>{t("staff.schools.noCountries")}</ComboboxEmpty>
                                        <ComboboxList>
                                            {(c: Country) => (
                                                <ComboboxItem key={c.cca2} value={c}>
                                                    {regionNames[c.cca2] ?? c.cca2}
                                                </ComboboxItem>
                                            )}
                                        </ComboboxList>
                                    </ComboboxPopup>
                                </Combobox>
                            </Field>

                            <Field>
                                <FieldLabel>{t("staff.schools.verification")}</FieldLabel>
                                {turnstile.configured ? (
                                    <TurnstileWidget
                                        key={turnstile.widgetKey}
                                        action="create-school"
                                        onTokenChange={turnstile.setToken}
                                    />
                                ) : (
                                    <p className="text-sm text-destructive">
                                        {t("staff.schools.turnstileMissing")}
                                    </p>
                                )}
                            </Field>
                        </SheetPanel>

                        <SheetFooter className="mt-auto">
                            <SheetClose render={<Button variant="ghost" disabled={submittingSchool} />}>{t("staff.schools.cancel")}</SheetClose>
                            <Button type="submit" disabled={submittingSchool || !turnstile.configured || !turnstile.token}>
                                {t(submittingSchool ? "staff.schools.creating" : "staff.schools.createSubmit")}
                            </Button>
                        </SheetFooter>
                    </Form>
                </SheetPopup>
            </Sheet>
        </div>
    )
}

function SchoolGroup({title, schools, regionNames, owned = false, onLeave}: {
    title: string
    schools: School[]
    regionNames: Record<string, string>
    owned?: boolean
    onLeave?: (school: School) => void
}) {
    const {t} = useLocale()
    return (
        <section className="space-y-2.5">
            <div className="flex items-center justify-between gap-3 px-1">
                <h2 className="text-lg font-semibold">{title}</h2>
                <Badge variant="secondary">{schools.length}</Badge>
            </div>
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                {schools.map((school) => {
                    const region = regionNames[school.regionCode]
                    return (
                        <Card key={school.id} className="group/school overflow-hidden transition-[border-color,box-shadow,transform] hover:-translate-y-0.5 hover:border-primary/25 hover:shadow-md/10">
                            <Link
                                href={`/app/staff/schools/${school.id}`}
                                aria-label={t("staff.schools.open", {name: school.name})}
                                className="absolute inset-0 z-[1] rounded-2xl focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset"
                            />
                            <CardHeader className="pointer-events-none relative z-10 grid-cols-[auto_1fr] items-center gap-3">
                                <span className="flex size-11 shrink-0 items-center justify-center overflow-hidden rounded-xl border bg-muted shadow-sm">
                                    {school.regionCode ? (
                                        <ReactCountryFlag
                                            svg
                                            countryCode={school.regionCode}
                                            style={{width: "100%", height: "100%", objectFit: "cover"}}
                                        />
                                    ) : <SchoolIcon className="size-5 text-muted-foreground" />}
                                </span>
                                <div className="min-w-0">
                                    <div className="flex items-center gap-2">
                                        <CardTitle className="truncate">{school.name}</CardTitle>
                                        {owned && <Badge variant="outline">{t("staff.schools.owner")}</Badge>}
                                    </div>
                                    <CardDescription className="mt-1 truncate">
                                        {region ?? school.regionCode ?? t("staff.schools.regionMissing")}
                                    </CardDescription>
                                </div>
                            </CardHeader>
                            {onLeave && (
                                <CardFooter className="relative z-10 justify-end pt-0">
                                    <Button size="sm" variant="destructive-outline" onClick={() => onLeave(school)}>
                                        {t("staff.schools.leaveAction")}
                                    </Button>
                                </CardFooter>
                            )}
                        </Card>
                    )
                })}
            </div>
        </section>
    )
}

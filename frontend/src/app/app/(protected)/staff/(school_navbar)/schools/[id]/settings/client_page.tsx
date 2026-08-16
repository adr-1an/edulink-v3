"use client"

import {useState} from "react"
import {useRouter} from "next/navigation"
import {toast} from "sonner"
import PageTitle from "@/components/app/page_title"
import {useLocalDateTimeFormatter} from "@/components/local-date-time"
import {Alert, AlertDescription, AlertTitle} from "@/components/ui/alert"
import {Badge} from "@/components/ui/badge"
import {Button} from "@/components/ui/button"
import {Card, CardContent, CardDescription, CardHeader, CardTitle} from "@/components/ui/card"
import {Checkbox} from "@/components/ui/checkbox"
import {Field, FieldDescription, FieldLabel} from "@/components/ui/field"
import {Form} from "@/components/ui/form"
import {Input} from "@/components/ui/input"
import {CalendarDays, CircleHelp, Plus, School as SchoolIcon, ShieldCheck, Trash2, TrendingUp, TriangleAlert} from "lucide-react"
import {
    handleClearActiveAcademicYear, handleCreateAcademicYear, handleDeleteAcademicYear,
    handleCompleteSchoolDeletion, handleDeleteSchool, handlePromoteSchool, handleUpdateSchool,
    type SchoolDeletionChallenge,
} from "./actions"
import countries, {Country} from "world-countries"
import {
    Combobox, ComboboxEmpty, ComboboxInput, ComboboxItem, ComboboxList, ComboboxPopup,
} from "@/components/ui/combobox"

import React from "react"
import {
    AlertDialog, AlertDialogClose, AlertDialogDescription, AlertDialogFooter,
    AlertDialogHeader, AlertDialogPopup, AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import {
    Dialog, DialogClose, DialogDescription, DialogFooter, DialogHeader,
    DialogPanel, DialogPopup, DialogTitle, DialogTrigger,
} from "@/components/ui/dialog"
import {emptySchoolAccess, hasSchoolPermission, type SchoolAccess} from "@/lib/school_access"
import {useLocale} from "@/i18n/provider"
import {OTPField, OTPFieldInput, OTPFieldSeparator} from "@/components/ui/otp-field"

interface AcademicYear { id: string; startYear: number; endYear: number; isActive: boolean }
interface Props {
    school: {id: string; name: string; regionCode: string}
    academicYears: AcademicYear[]
    canListAcademicYears: boolean
    access: SchoolAccess
    regionNames: Record<string, string>
    currentYear: number
}
interface PromotionPlan {
    from: number
    to: number
    activate: boolean
    transferGrades: boolean
    promoteLevels: boolean
}

export default function SettingsClientPage({school, academicYears, canListAcademicYears, access = emptySchoolAccess, regionNames, currentYear}: Props) {
    const router = useRouter()
    const {t} = useLocale()
    const formatDate = useLocalDateTimeFormatter()
    const [creating, setCreating] = useState(false)
    const [savingDetails, setSavingDetails] = useState(false)
    const [promoting, setPromoting] = useState(false)
    const [promotionPlan, setPromotionPlan] = useState<PromotionPlan | null>(null)
    const [activating, setActivating] = useState<string | null>(null)
    const [activationTarget, setActivationTarget] = useState<AcademicYear | null>(null)
    const [deleteTarget, setDeleteTarget] = useState<AcademicYear | null>(null)
    const [deleting, setDeleting] = useState(false)
    const [confirmingClear, setConfirmingClear] = useState(false)
    const [clearing, setClearing] = useState(false)
    const [confirmingSchoolDelete, setConfirmingSchoolDelete] = useState(false)
    const [schoolDeleteConfirmation, setSchoolDeleteConfirmation] = useState("")
    const [schoolDeleteAcknowledged, setSchoolDeleteAcknowledged] = useState(false)
    const [deletingSchool, setDeletingSchool] = useState(false)
    const [schoolDeletionChallenge, setSchoolDeletionChallenge] = useState<SchoolDeletionChallenge | null>(null)
    const [schoolDeletionCode, setSchoolDeletionCode] = useState("")
    const [verifyingSchoolDeletion, setVerifyingSchoolDeletion] = useState(false)
    const [schoolDeletionError, setSchoolDeletionError] = useState<string | null>(null)
    const [schoolDeletionChallengeInvalid, setSchoolDeletionChallengeInvalid] = useState(false)
    const [activate, setActivate] = useState(true)
    const [transferGrades, setTransferGrades] = useState(true)
    const [promoteLevels, setPromoteLevels] = useState(true)
    const years = [...academicYears].sort((a, b) => b.startYear - a.startYear)
    const activeYear = years.find((year) => year.isActive)
    const [createFrom, setCreateFrom] = useState(String(currentYear))
    const [createTo, setCreateTo] = useState(String(currentYear + 1))
    const [promotionFrom, setPromotionFrom] = useState(String(activeYear?.endYear ?? currentYear))
    const [promotionTo, setPromotionTo] = useState(String((activeYear?.endYear ?? currentYear) + 1))
    const [schoolName, setSchoolName] = useState(school.name)
    const [schoolCountry, setSchoolCountry] = useState<Country | null>(
        countries.find((country) => country.cca2 === school.regionCode) ?? null,
    )
    const requiredSchoolDeleteConfirmation = `DELETE ${school.name}`
    const canDeleteSchool = schoolDeleteAcknowledged
        && schoolDeleteConfirmation === requiredSchoolDeleteConfirmation
    const canUpdateSchool = hasSchoolPermission(access, "school.update")
    const canPromoteSchool = hasSchoolPermission(access, "school.promote")
    const canCreateAcademicYear = hasSchoolPermission(access, "academicYear.create")
    const canToggleAcademicYear = hasSchoolPermission(access, "academicYear.toggleActive")
    const canActivateAcademicYear = canToggleAcademicYear && canUpdateSchool
    const canDeleteAcademicYear = hasSchoolPermission(access, "academicYear.delete")

    function range(form: HTMLFormElement) {
        const data = new FormData(form)
        return {from: Number(data.get("from")), to: Number(data.get("to"))}
    }

    async function createYear(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        const {from, to} = range(event.currentTarget)
        setCreating(true)
        const res = await handleCreateAcademicYear(school.id, from, to)
        setCreating(false)
        if (!res.ok) return toast.error(t("staff.settings.error.action"))
        toast.success(t("staff.settings.yearCreated"))
        router.refresh()
    }

    async function updateSchoolDetails(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        setSavingDetails(true)
        const res = await handleUpdateSchool(school.id, {
            name: schoolName,
            regionCode: schoolCountry?.cca2 ?? "",
        })
        setSavingDetails(false)
        if (!res.ok) return toast.error(t("staff.settings.error.action"))

        toast.success(t("staff.settings.detailsUpdated"))
        router.refresh()
    }

    function preparePromotion(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        const {from, to} = range(event.currentTarget)
        setPromotionPlan({
            from,
            to,
            activate,
            transferGrades,
            promoteLevels: transferGrades && promoteLevels,
        })
    }

    async function promote() {
        if (!promotionPlan) return

        setPromoting(true)
        const res = await handlePromoteSchool(school.id, promotionPlan.from, promotionPlan.to, {
            activateAfterPromotion: promotionPlan.activate,
            transferGrades: promotionPlan.transferGrades,
            promoteGradeLevels: promotionPlan.promoteLevels,
        })
        setPromoting(false)
        if (!res.ok) return toast.error(t("staff.settings.error.action"))
        setPromotionPlan(null)
        toast.success(t("staff.settings.promoted"))
        router.refresh()
    }

    async function activateYear() {
        if (!activationTarget) return

        setActivating(activationTarget.id)
        const res = await handleUpdateSchool(school.id, {
            name: school.name,
            regionCode: school.regionCode,
            activeAcademicYearId: activationTarget.id,
        })
        setActivating(null)
        if (!res.ok) return toast.error(t("staff.settings.error.action"))
        setPromotionFrom(String(activationTarget.endYear))
        setPromotionTo(String(activationTarget.endYear + 1))
        setActivationTarget(null)
        toast.success(t("staff.settings.activeUpdated"))
        router.refresh()
    }

    async function deleteYear() {
        if (!deleteTarget) return

        setDeleting(true)
        const res = await handleDeleteAcademicYear(deleteTarget.id)
        setDeleting(false)
        if (!res.ok) return toast.error(t("staff.settings.error.action"))

        setDeleteTarget(null)
        toast.success(t("staff.settings.yearDeleted"))
        router.refresh()
    }

    async function clearActiveYear() {
        setClearing(true)
        const res = await handleClearActiveAcademicYear(school.id)
        setClearing(false)
        if (!res.ok) return toast.error(t("staff.settings.error.action"))

        setConfirmingClear(false)
        toast.success(t("staff.settings.activeCleared"))
        router.refresh()
    }

    function openSchoolDeleteConfirmation() {
        setSchoolDeleteConfirmation("")
        setSchoolDeleteAcknowledged(false)
        setConfirmingSchoolDelete(true)
    }

    async function deleteSchool() {
        if (!canDeleteSchool) return

        setDeletingSchool(true)
        const res = await handleDeleteSchool(school.id)
        setDeletingSchool(false)
        if (!res.ok) return toast.error(t("staff.settings.error.action"))

        if (res.step === "two_factor") {
            setConfirmingSchoolDelete(false)
            setSchoolDeleteConfirmation("")
            setSchoolDeleteAcknowledged(false)
            setSchoolDeletionChallenge(res.challenge)
            setSchoolDeletionCode("")
            setSchoolDeletionError(null)
            setSchoolDeletionChallengeInvalid(false)
            return
        }

        setConfirmingSchoolDelete(false)
        toast.success(t("staff.settings.schoolDeleted"))
        router.push("/app")
    }

    async function verifySchoolDeletion(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        if (!schoolDeletionChallenge || schoolDeletionChallengeInvalid || schoolDeletionCode.length !== 6) return

        setVerifyingSchoolDeletion(true)
        setSchoolDeletionError(null)
        const result = await handleCompleteSchoolDeletion(schoolDeletionChallenge.token, schoolDeletionCode)

        if (!result.ok) {
            setVerifyingSchoolDeletion(false)
            if (result.code === "expired" || result.code === "invalid_challenge") {
                setSchoolDeletionChallengeInvalid(true)
            }

            const message = result.code === "invalid_code"
                ? t("staff.settings.deleteSchoolTwoFactorInvalid")
                : result.code === "expired"
                    ? t("staff.settings.deleteSchoolTwoFactorExpired")
                    : result.code === "invalid_challenge"
                        ? t("staff.settings.deleteSchoolTwoFactorInvalidChallenge")
                        : result.code === "invalid_input"
                            ? t("staff.settings.deleteSchoolTwoFactorInvalidInput")
                            : result.code === "network"
                                ? t("staff.settings.error.network")
                                : result.code === "forbidden"
                                    ? t("staff.settings.error.forbidden")
                                    : t("staff.settings.deleteSchoolTwoFactorError")
            setSchoolDeletionError(message)
            if (result.code === "invalid_code" || result.code === "invalid_input") {
                setSchoolDeletionCode("")
            }
            return
        }

        setSchoolDeletionChallenge(null)
        toast.success(t("staff.settings.schoolDeleted"))
        router.push("/app")
    }

    function changeSchoolDeletionChallengeOpen(open: boolean) {
        if (open || verifyingSchoolDeletion) return
        setSchoolDeletionChallenge(null)
        setSchoolDeletionCode("")
        setSchoolDeletionError(null)
        setSchoolDeletionChallengeInvalid(false)
    }

    return (
        <div className="space-y-6">
            <div className="flex flex-wrap items-end justify-between gap-3">
                <div>
                    <PageTitle centered={false}>{t("staff.settings.title")}</PageTitle>
                    <p className="text-muted-foreground">
                        {canListAcademicYears
                            ? t("staff.settings.descriptionYears", {name: school.name})
                            : t("staff.settings.description", {name: school.name})}
                    </p>
                </div>
                {canListAcademicYears && <Dialog>
                    <DialogTrigger render={<Button variant="outline" />}>
                        <CircleHelp /> {t("staff.settings.yearHelp")}
                    </DialogTrigger>
                    <DialogPopup className="sm:max-w-2xl">
                        <DialogHeader>
                            <DialogTitle>{t("staff.settings.helpTitle")}</DialogTitle>
                            <DialogDescription>{t("staff.settings.helpDescription")}</DialogDescription>
                        </DialogHeader>
                        <DialogPanel className="space-y-5 text-sm text-muted-foreground">
                            <section className="space-y-2">
                                <h3 className="font-medium text-foreground">{t("staff.settings.helpActiveTitle")}</h3>
                                <p>{t("staff.settings.helpActive1")}</p>
                                <p>{t("staff.settings.helpActive2")}</p>
                            </section>
                            <section className="space-y-2">
                                <h3 className="font-medium text-foreground">{t("staff.settings.helpCreateTitle")}</h3>
                                <p>{t("staff.settings.helpCreate1")}</p>
                                <p>{t("staff.settings.helpCreate2")}</p>
                            </section>
                            <section className="space-y-2">
                                <h3 className="font-medium text-foreground">{t("staff.settings.helpPromoteTitle")}</h3>
                                <p>{t("staff.settings.helpPromote")}</p>
                                <ul className="list-disc space-y-1 pl-5">
                                    <li>{t("staff.settings.helpActivate")}</li>
                                    <li>{t("staff.settings.helpTransfer")}</li>
                                    <li>{t("staff.settings.helpLevels")}</li>
                                </ul>
                            </section>
                            <section className="space-y-2">
                                <h3 className="font-medium text-foreground">{t("staff.settings.helpSafeTitle")}</h3>
                                <p>{t("staff.settings.helpSafe1")}</p>
                                <p>{t("staff.settings.helpSafe2")}</p>
                            </section>
                        </DialogPanel>
                        <DialogFooter>
                            <DialogClose render={<Button />}>{t("staff.settings.gotIt")}</DialogClose>
                        </DialogFooter>
                    </DialogPopup>
                </Dialog>}
            </div>

            <section>
                <h2 className="mb-3 text-2xl font-semibold">{t("staff.settings.schoolDetails")}</h2>
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2"><SchoolIcon /> {t("staff.settings.general")}</CardTitle>
                        <CardDescription>
                            {canUpdateSchool
                                ? t("staff.settings.generalEdit")
                                : t("staff.settings.generalView")}
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        {canUpdateSchool ? <Form className="grid max-w-2xl gap-4" onSubmit={updateSchoolDetails}>
                            <Field>
                                <FieldLabel>{t("staff.settings.name")}</FieldLabel>
                                <Input
                                    value={schoolName}
                                    onChange={(event) => setSchoolName(event.target.value)}
                                    maxLength={64}
                                    required
                                />
                            </Field>
                            <Field>
                                <FieldLabel>{t("staff.settings.schoolRegion")}</FieldLabel>
                                <Combobox
                                    required
                                    items={countries}
                                    itemToStringLabel={(item) => item ? regionNames[item.cca2] ?? item.cca2 : ""}
                                    itemToStringValue={(item) => item?.cca2 ?? ""}
                                    value={schoolCountry}
                                    onValueChange={setSchoolCountry}
                                >
                                    <ComboboxInput placeholder={t("staff.settings.selectCountry")} required />
                                    <ComboboxPopup>
                                        <ComboboxEmpty>{t("staff.settings.noCountries")}</ComboboxEmpty>
                                        <ComboboxList>
                                            {(item: Country) => (
                                                <ComboboxItem key={item.cca2} value={item}>
                                                    {regionNames[item.cca2] ?? item.cca2}
                                                </ComboboxItem>
                                            )}
                                        </ComboboxList>
                                    </ComboboxPopup>
                                </Combobox>
                            </Field>
                            <Button className="justify-self-start" type="submit" loading={savingDetails} disabled={!schoolName.trim() || !schoolCountry}>
                                {t("staff.settings.saveDetails")}
                            </Button>
                        </Form> : <dl className="grid max-w-2xl gap-4 sm:grid-cols-2">
                            <div className="rounded-lg border bg-muted/25 p-3">
                                <dt className="text-xs text-muted-foreground">{t("staff.settings.name")}</dt>
                                <dd className="mt-1 text-sm font-medium">{school.name}</dd>
                            </div>
                            <div className="rounded-lg border bg-muted/25 p-3">
                                <dt className="text-xs text-muted-foreground">{t("staff.settings.region")}</dt>
                                <dd className="mt-1 text-sm font-medium">{regionNames[school.regionCode] ?? school.regionCode}</dd>
                            </div>
                        </dl>}
                    </CardContent>
                </Card>
            </section>

            {canListAcademicYears && (canCreateAcademicYear || canPromoteSchool) && <>
            <section>
                <h2 className="text-2xl font-semibold">{t("staff.settings.yearSettings")}</h2>
                <p className="mb-3 text-sm text-muted-foreground">{t("staff.settings.yearSettingsDescription")}</p>

                <div className="grid gap-4 lg:grid-cols-2">
                {canCreateAcademicYear && <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2"><Plus /> {t("staff.settings.createYear")}</CardTitle>
                        <CardDescription>{t("staff.settings.createYearDescription")}</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <Form className="grid gap-4" onSubmit={createYear}>
                            <YearFields from={createFrom} to={createTo} onFromChange={setCreateFrom} onToChange={setCreateTo} />
                            <Button type="submit" loading={creating}>{t("staff.settings.createYearAction")}</Button>
                        </Form>
                    </CardContent>
                </Card>}

                {canPromoteSchool && <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2"><TrendingUp /> {t("staff.settings.promote")}</CardTitle>
                        <CardDescription>{t("staff.settings.promoteDescription")}</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <Form className="grid gap-4" onSubmit={preparePromotion}>
                            <YearFields from={promotionFrom} to={promotionTo} onFromChange={setPromotionFrom} onToChange={setPromotionTo} />
                            <Check label={t("staff.settings.activateNew")} checked={activate} onChange={setActivate} />
                            <Check label={t("staff.settings.transferGrades")} checked={transferGrades} onChange={setTransferGrades} />
                            <Check label={t("staff.settings.increaseLevels")} checked={promoteLevels} onChange={setPromoteLevels} disabled={!transferGrades} />
                            <Button type="submit" loading={promoting} disabled={!activeYear}>{t("staff.settings.promote")}</Button>
                            {!activeYear && (
                                <Field>
                                    <FieldDescription>
                                        {t("staff.settings.activeRequired")}
                                    </FieldDescription>
                                </Field>
                            )}
                        </Form>
                    </CardContent>
                </Card>}
                </div>
            </section>
            </>}

            {canListAcademicYears && <>
            <section>
                <h2 className="mb-3 text-2xl font-semibold">{t("staff.settings.years")}</h2>
                {years.length ? (
                    <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                        {years.map((year) => (
                            <Card key={year.id}>
                                <CardHeader>
                                    <CardTitle className="flex items-center justify-between gap-2">
                                        <span className="flex items-center gap-2"><CalendarDays /> {year.startYear}–{year.endYear}</span>
                                        {year.isActive && <Badge variant="success">{t("staff.settings.active")}</Badge>}
                                    </CardTitle>
                                </CardHeader>
                                {year.isActive ? (
                                    canToggleAcademicYear ? (
                                    <CardContent>
                                        <Button variant="destructive-outline" disabled={clearing} onClick={() => setConfirmingClear(true)}>
                                            {t("staff.settings.clearActive")}
                                        </Button>
                                    </CardContent>
                                    ) : null
                                ) : (canActivateAcademicYear || canDeleteAcademicYear) ? (
                                    <CardContent className="flex gap-2">
                                        {canActivateAcademicYear && <Button variant="outline" loading={activating === year.id} disabled={activating !== null} onClick={() => setActivationTarget(year)}>
                                            {t("staff.settings.setActive")}
                                        </Button>}
                                        {canDeleteAcademicYear && <Button variant="destructive-outline" disabled={activating !== null} onClick={() => setDeleteTarget(year)}>
                                            <Trash2 />
                                        </Button>}
                                    </CardContent>
                                ) : null}
                            </Card>
                        ))}
                    </div>
                ) : <Card className="p-8 text-center text-muted-foreground">{t("staff.settings.noYears")}</Card>}
            </section>
            </>}

            {access.owner && <section className="space-y-3">
                <div>
                    <h2 className="text-2xl font-semibold text-destructive">{t("staff.settings.dangerZone")}</h2>
                    <p className="text-sm text-muted-foreground">{t("staff.settings.dangerDescription")}</p>
                </div>
                <Card className="border-destructive/50">
                    <CardHeader>
                        <CardTitle>{t("staff.settings.deleteSchoolTitle")}</CardTitle>
                        <CardDescription>{t("staff.settings.deleteSchoolDescription", {name: school.name})}</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <Button variant="destructive" onClick={openSchoolDeleteConfirmation}>
                            <Trash2 /> {t("staff.settings.deleteSchool")}
                        </Button>
                    </CardContent>
                </Card>
            </section>}

            <AlertDialog open={activationTarget !== null} onOpenChange={(open) => {
                if (!open && !activating) setActivationTarget(null)
            }}>
                <AlertDialogPopup>
                    <AlertDialogHeader>
                        <AlertDialogTitle>{t("staff.settings.changeActiveTitle")}</AlertDialogTitle>
                        <AlertDialogDescription className="space-y-3 text-left">
                            {activationTarget && (
                                <>
                                    <span className="block">
                                        {t("staff.settings.willBecomeActive", {year: `${activationTarget.startYear}–${activationTarget.endYear}`})}
                                    </span>
                                    {activeYear && (
                                        <span className="block">
                                            {t("staff.settings.willBecomeInactive", {year: `${activeYear.startYear}–${activeYear.endYear}`})}
                                        </span>
                                    )}
                                    <span className="block">
                                        {t("staff.settings.dashboardSwitch")}
                                    </span>
                                    <span className="block font-medium text-foreground">
                                        {t("staff.settings.noDataChanged")}
                                    </span>
                                </>
                            )}
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogClose render={<Button variant="outline" disabled={activating !== null} />}>{t("staff.settings.cancel")}</AlertDialogClose>
                        <Button loading={activating !== null} onClick={activateYear}>{t("staff.settings.setActiveYear")}</Button>
                    </AlertDialogFooter>
                </AlertDialogPopup>
            </AlertDialog>

            <AlertDialog open={confirmingClear} onOpenChange={(open) => {
                if (!clearing) setConfirmingClear(open)
            }}>
                <AlertDialogPopup>
                    <AlertDialogHeader>
                        <AlertDialogTitle>{t("staff.settings.clearActiveTitle")}</AlertDialogTitle>
                        <AlertDialogDescription className="space-y-3 text-left">
                            {activeYear && (
                                <span className="block">
                                    {t("staff.settings.noLongerActive", {year: `${activeYear.startYear}–${activeYear.endYear}`})}
                                </span>
                            )}
                            <span className="block">
                                {t("staff.settings.noCurrentYear")}
                            </span>
                            <span className="block font-medium text-foreground">
                                {t("staff.settings.yearNotDeleted")}
                            </span>
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogClose render={<Button variant="outline" disabled={clearing} />}>{t("staff.settings.cancel")}</AlertDialogClose>
                        <Button variant="destructive" loading={clearing} onClick={clearActiveYear}>{t("staff.settings.clearActive")}</Button>
                    </AlertDialogFooter>
                </AlertDialogPopup>
            </AlertDialog>

            <AlertDialog open={confirmingSchoolDelete} onOpenChange={(open) => {
                if (!deletingSchool) setConfirmingSchoolDelete(open)
            }}>
                <AlertDialogPopup className="sm:max-w-xl">
                    <AlertDialogHeader>
                        <AlertDialogTitle className="text-destructive">{t("staff.settings.deleteSchoolConfirmTitle")}</AlertDialogTitle>
                        <AlertDialogDescription className="space-y-4 text-left">
                            <span className="block text-base font-medium text-foreground">
                                {t("staff.settings.aboutToDelete", {name: school.name})}
                            </span>
                            <span className="block">
                                {t("staff.settings.schoolDisappears")}
                            </span>
                            <span className="block font-medium text-destructive">
                                {t("staff.settings.cannotUndo")}
                            </span>
                        </AlertDialogDescription>
                    </AlertDialogHeader>

                    <div className="grid gap-4 px-6 pb-2">
                        <Field>
                            <FieldLabel>
                                {t("staff.settings.typeToConfirm", {value: requiredSchoolDeleteConfirmation})}
                            </FieldLabel>
                            <Input
                                value={schoolDeleteConfirmation}
                                onChange={(event) => setSchoolDeleteConfirmation(event.target.value)}
                                placeholder={requiredSchoolDeleteConfirmation}
                                autoComplete="off"
                                spellCheck={false}
                                disabled={deletingSchool}
                                autoFocus
                            />
                        </Field>
                        <label className="flex items-start gap-3 text-sm">
                            <Checkbox
                                checked={schoolDeleteAcknowledged}
                                onCheckedChange={(value) => setSchoolDeleteAcknowledged(value)}
                                disabled={deletingSchool}
                            />
                            <span>{t("staff.settings.deleteAcknowledge")}</span>
                        </label>
                    </div>

                    <AlertDialogFooter>
                        <AlertDialogClose render={<Button variant="outline" disabled={deletingSchool} />}>{t("staff.settings.cancel")}</AlertDialogClose>
                        <Button variant="destructive" loading={deletingSchool} disabled={!canDeleteSchool} onClick={deleteSchool}>
                            {t("staff.settings.deleteSchool")}
                        </Button>
                    </AlertDialogFooter>
                </AlertDialogPopup>
            </AlertDialog>

            <Dialog open={schoolDeletionChallenge !== null} onOpenChange={changeSchoolDeletionChallengeOpen}>
                <DialogPopup className="sm:max-w-md" closeProps={{disabled: verifyingSchoolDeletion}}>
                    <DialogHeader>
                        <DialogTitle>{t("staff.settings.deleteSchoolTwoFactorTitle")}</DialogTitle>
                        <DialogDescription>{t("staff.settings.deleteSchoolTwoFactorDescription", {name: school.name})}</DialogDescription>
                    </DialogHeader>

                    <form className="contents" onSubmit={verifySchoolDeletion}>
                        <DialogPanel className="space-y-5">
                            <Alert variant="warning">
                                <TriangleAlert />
                                <AlertTitle>{t("staff.settings.deleteSchoolTwoFactorWarningTitle")}</AlertTitle>
                                <AlertDescription>{t("staff.settings.deleteSchoolTwoFactorWarning")}</AlertDescription>
                            </Alert>

                            {schoolDeletionError && (
                                <Alert variant="error">
                                    <TriangleAlert />
                                    <AlertTitle>{t("staff.settings.deleteSchoolTwoFactorFailedTitle")}</AlertTitle>
                                    <AlertDescription>{schoolDeletionError}</AlertDescription>
                                </Alert>
                            )}

                            <Field>
                                <FieldLabel htmlFor="school-deletion-two-factor-code">{t("staff.settings.deleteSchoolTwoFactorCode")}</FieldLabel>
                                <OTPField
                                    id="school-deletion-two-factor-code"
                                    className="justify-center sm:justify-start"
                                    length={6}
                                    value={schoolDeletionCode}
                                    disabled={verifyingSchoolDeletion || schoolDeletionChallengeInvalid}
                                    onValueChange={(value) => {
                                        setSchoolDeletionCode(value)
                                        setSchoolDeletionError(null)
                                    }}
                                >
                                    <OTPFieldInput autoFocus aria-label={t("staff.settings.deleteSchoolTwoFactorDigit", {position: 1})} />
                                    <OTPFieldInput aria-label={t("staff.settings.deleteSchoolTwoFactorDigit", {position: 2})} />
                                    <OTPFieldInput aria-label={t("staff.settings.deleteSchoolTwoFactorDigit", {position: 3})} />
                                    <OTPFieldSeparator />
                                    <OTPFieldInput aria-label={t("staff.settings.deleteSchoolTwoFactorDigit", {position: 4})} />
                                    <OTPFieldInput aria-label={t("staff.settings.deleteSchoolTwoFactorDigit", {position: 5})} />
                                    <OTPFieldInput aria-label={t("staff.settings.deleteSchoolTwoFactorDigit", {position: 6})} />
                                </OTPField>
                                {schoolDeletionChallenge && (
                                    <FieldDescription>
                                        {t("staff.settings.deleteSchoolTwoFactorExpires", {date: formatDate(schoolDeletionChallenge.expiresAt)})}
                                    </FieldDescription>
                                )}
                            </Field>
                        </DialogPanel>

                        <DialogFooter>
                            <DialogClose render={<Button type="button" variant="outline" />} disabled={verifyingSchoolDeletion}>
                                {t("staff.settings.cancel")}
                            </DialogClose>
                            <Button
                                type="submit"
                                variant="destructive"
                                loading={verifyingSchoolDeletion}
                                disabled={schoolDeletionCode.length !== 6 || schoolDeletionChallengeInvalid}
                            >
                                <ShieldCheck /> {t("staff.settings.deleteSchoolTwoFactorSubmit")}
                            </Button>
                        </DialogFooter>
                    </form>
                </DialogPopup>
            </Dialog>

            <AlertDialog open={promotionPlan !== null} onOpenChange={(open) => {
                if (!open && !promoting) setPromotionPlan(null)
            }}>
                <AlertDialogPopup>
                    <AlertDialogHeader>
                        <AlertDialogTitle>{t("staff.settings.confirmPromotion")}</AlertDialogTitle>
                        <AlertDialogDescription className="space-y-3 text-left">
                            {promotionPlan && (
                                <>
                                    <span className="block">
                                        {t("staff.settings.newYearCreated", {year: `${promotionPlan.from}–${promotionPlan.to}`})}
                                    </span>
                                    <span className="block">
                                        {promotionPlan.activate
                                            ? t("staff.settings.promotionActivate")
                                            : t("staff.settings.promotionInactive")}
                                    </span>
                                    <span className="block">
                                        {promotionPlan.transferGrades
                                            ? promotionPlan.promoteLevels
                                                ? t("staff.settings.promotionGradesLevels")
                                                : t("staff.settings.promotionGrades")
                                            : t("staff.settings.promotionEmpty")}
                                    </span>
                                    <span className="block font-medium text-foreground">
                                        {t("staff.settings.existingUnchanged")}
                                    </span>
                                </>
                            )}
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogClose render={<Button variant="outline" disabled={promoting} />}>{t("staff.settings.goBack")}</AlertDialogClose>
                        <Button loading={promoting} onClick={promote}>{t("staff.settings.confirmPromotionAction")}</Button>
                    </AlertDialogFooter>
                </AlertDialogPopup>
            </AlertDialog>

            <AlertDialog open={deleteTarget !== null} onOpenChange={(open) => {
                if (!open && !deleting) setDeleteTarget(null)
            }}>
                <AlertDialogPopup>
                    <AlertDialogHeader>
                        <AlertDialogTitle>{t("staff.settings.deleteYearTitle")}</AlertDialogTitle>
                        <AlertDialogDescription>{t("staff.settings.deleteYearDescription", {
                            year: deleteTarget ? `${deleteTarget.startYear}–${deleteTarget.endYear}` : "",
                        })}</AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogClose render={<Button variant="outline" disabled={deleting} />}>{t("staff.settings.cancel")}</AlertDialogClose>
                        <Button variant="destructive" loading={deleting} onClick={deleteYear}>
                            {t("staff.settings.deleteYear")}
                        </Button>
                    </AlertDialogFooter>
                </AlertDialogPopup>
            </AlertDialog>
        </div>
    )
}

function YearFields({from, to, onFromChange, onToChange}: {
    from: string
    to: string
    onFromChange: (value: string) => void
    onToChange: (value: string) => void
}) {
    const {t} = useLocale()
    return <div className="grid grid-cols-2 gap-3">
        <Field>
            <FieldLabel>{t("staff.settings.startYear")}</FieldLabel>
            <Input name="from" type="number" min={1900} max={9999} value={from} onChange={(event) => onFromChange(event.target.value)} required />
        </Field>
        <Field>
            <FieldLabel>{t("staff.settings.endYear")}</FieldLabel>
            <Input name="to" type="number" min={1900} max={9999} value={to} onChange={(event) => onToChange(event.target.value)} required />
        </Field>
    </div>
}

function Check({label, checked, onChange, disabled}: {label: string; checked: boolean; onChange: (value: boolean) => void; disabled?: boolean}) {
    return <label className="flex items-center gap-2 text-sm"><Checkbox checked={checked} onCheckedChange={(value) => onChange(value)} disabled={disabled} /> {label}</label>
}

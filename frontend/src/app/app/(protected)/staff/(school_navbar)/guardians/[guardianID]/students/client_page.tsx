"use client"

import {useState, useTransition} from "react"
import Link from "next/link"
import {useRouter} from "next/navigation"
import {ArrowLeft, CalendarDays, GraduationCap, Unlink, UserRound} from "lucide-react"
import {toast} from "sonner"
import {useLocale} from "@/i18n/provider"
import {hasSchoolPermission, type SchoolAccess} from "@/lib/school_access"
import type {GuardianStudent} from "@/lib/guardian_students"
import {Button} from "@/components/ui/button"
import {Card, CardContent} from "@/components/ui/card"
import {Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle} from "@/components/ui/empty"
import {AlertDialog, AlertDialogClose, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogPopup, AlertDialogTitle} from "@/components/ui/alert-dialog"
import {unassignStudentGuardian} from "@/app/app/(protected)/staff/(school_navbar)/students/[studentID]/guardians/actions"

export default function GuardianStudentsPage({guardianID, students, access}: {guardianID: string; students: GuardianStudent[]; access: SchoolAccess}) {
    const {locale, t} = useLocale()
    const router = useRouter()
    const [removeTarget, setRemoveTarget] = useState<GuardianStudent | null>(null)
    const [removing, setRemoving] = useState(false)
    const [refreshing, startTransition] = useTransition()
    const busy = removing || refreshing
    const canView = hasSchoolPermission(access, "student.view")
    const canUnassign = hasSchoolPermission(access, "student.guardian.unassign")
    const formatDate = new Intl.DateTimeFormat(locale === "pl" ? "pl-PL" : "en-US", {
        day: "numeric", month: "short", year: "numeric", timeZone: "UTC",
    })

    async function unassign() {
        if (!removeTarget || busy || !canUnassign) return
        setRemoving(true)
        try {
            const result = await unassignStudentGuardian(removeTarget.id, guardianID)
            if (!result.ok) return toast.error(t(`staff.studentGuardians.error.${result.code}`))
            setRemoveTarget(null)
            toast.success(t("staff.studentGuardians.removed"))
            startTransition(() => router.refresh())
        } catch {
            toast.error(t("staff.studentGuardians.error.network"))
        } finally {
            setRemoving(false)
        }
    }

    return <div className="space-y-6">
        <Button variant="ghost" className="-ml-3" onClick={() => router.back()}><ArrowLeft /> {t("common.back")}</Button>
        <header>
            <h1 className="text-3xl font-semibold tracking-tight">{t("staff.guardianStudents.title")}</h1>
            <p className="mt-2 text-muted-foreground">{t("staff.guardianStudents.description")}</p>
        </header>
        {students.length === 0 ? <Card><Empty><EmptyHeader>
            <EmptyMedia variant="icon"><GraduationCap /></EmptyMedia>
            <EmptyTitle>{t("staff.guardianStudents.empty")}</EmptyTitle>
            <EmptyDescription>{t("staff.guardianStudents.emptyDescription")}</EmptyDescription>
        </EmptyHeader></Empty></Card> : <Card><CardContent className="divide-y p-0">
            {[...students].sort((a, b) => a.lastName.localeCompare(b.lastName, locale) || a.name.localeCompare(b.name, locale)).map((student) => {
                const name = `${student.name} ${student.lastName}`
                return <div key={student.id} className="flex flex-wrap items-center justify-between gap-3 p-5">
                    <div className="min-w-0">
                        <p className="break-words font-medium">{name}</p>
                        <p className="mt-1 flex items-center gap-2 text-sm text-muted-foreground"><CalendarDays className="size-4" />
                            {t("staff.students.born", {date: formatDate.format(new Date(`${student.dateOfBirth}T00:00:00Z`))})}
                        </p>
                    </div>
                    <div className="flex flex-wrap gap-2">
                        {canView && <Button variant="outline" size="sm" aria-label={t("staff.students.viewLabel", {name})}
                            render={<Link href={`/app/staff/students/${student.id}`} prefetch={false} />}><UserRound /> {t("staff.studentGuardians.profile")}</Button>}
                        {canUnassign && <Button variant="destructive-outline" size="sm" disabled={busy}
                            aria-label={t("staff.guardianStudents.removeLabel", {name})}
                            onClick={() => setRemoveTarget(student)}><Unlink /> {t("staff.studentGuardians.unassign")}</Button>}
                    </div>
                </div>
            })}
        </CardContent></Card>}
        <AlertDialog open={removeTarget !== null} onOpenChange={(open) => { if (!open && !busy) setRemoveTarget(null) }}>
            <AlertDialogPopup><AlertDialogHeader>
                <AlertDialogTitle>{t("staff.studentGuardians.removeTitle")}</AlertDialogTitle>
                <AlertDialogDescription>{t("staff.guardianStudents.removeDescription", {
                    name: removeTarget ? `${removeTarget.name} ${removeTarget.lastName}` : "",
                })}</AlertDialogDescription>
            </AlertDialogHeader><AlertDialogFooter>
                <AlertDialogClose render={<Button variant="outline" disabled={busy} />}>{t("common.cancel")}</AlertDialogClose>
                <Button variant="destructive" loading={removing} disabled={busy} onClick={unassign}>{t("staff.studentGuardians.unassign")}</Button>
            </AlertDialogFooter></AlertDialogPopup>
        </AlertDialog>
    </div>
}

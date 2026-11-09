"use client"

import {useEffect, useMemo, useState} from "react"
import Link from "next/link"
import {useRouter} from "next/navigation"
import {ArrowLeft, Mail, Phone, Plus, UsersRound, Unlink} from "lucide-react"
import {toast} from "sonner"
import {useLocale} from "@/i18n/provider"
import {hasSchoolPermission, type SchoolAccess} from "@/lib/school_access"
import type {StudentGuardian} from "@/lib/student_guardians"
import {rememberSchoolAccess} from "@/lib/school_navigation"
import {Button} from "@/components/ui/button"
import {Card, CardContent} from "@/components/ui/card"
import {Input} from "@/components/ui/input"
import {Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle} from "@/components/ui/empty"
import {Dialog, DialogDescription, DialogFooter, DialogHeader, DialogPanel, DialogPopup, DialogTitle} from "@/components/ui/dialog"
import {AlertDialog, AlertDialogClose, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogPopup, AlertDialogTitle} from "@/components/ui/alert-dialog"
import {assignStudentGuardian, unassignStudentGuardian} from "./actions"

export default function StudentGuardiansPage({studentID, schoolID, studentName, initialGuardians, availableGuardians, candidatesUnavailable, access}: {
    studentID: string
    schoolID: string
    studentName: string | null
    initialGuardians: StudentGuardian[]
    availableGuardians: StudentGuardian[]
    candidatesUnavailable: boolean
    access: SchoolAccess
}) {
    const {t} = useLocale()
    const router = useRouter()
    useEffect(() => {
        rememberSchoolAccess(schoolID, access)
    }, [schoolID, access])

    const [guardians, setGuardians] = useState(initialGuardians)
    const [assignOpen, setAssignOpen] = useState(false)
    const [query, setQuery] = useState("")
    const [busyID, setBusyID] = useState<string | null>(null)
    const [removeTarget, setRemoveTarget] = useState<StudentGuardian | null>(null)
    const canAssign = hasSchoolPermission(access, "student.guardian.assign")
    const canUnassign = hasSchoolPermission(access, "student.guardian.unassign")
    const candidates = useMemo(() => {
        const assignedIDs = new Set(guardians.map((guardian) => guardian.id))
        const search = query.trim().toLocaleLowerCase()
        return availableGuardians.filter((guardian) => !assignedIDs.has(guardian.id)
            && (!search || `${guardian.name} ${guardian.lastName} ${guardian.email} ${guardian.phone}`.toLocaleLowerCase().includes(search)))
            .sort((a, b) => a.lastName.localeCompare(b.lastName) || a.name.localeCompare(b.name))
    }, [availableGuardians, guardians, query])

    async function assign(guardian: StudentGuardian) {
        if (busyID || !canAssign) return
        setBusyID(guardian.id)
        const result = await assignStudentGuardian(studentID, guardian.id)
        setBusyID(null)
        if (!result.ok) return toast.error(t(`staff.studentGuardians.error.${result.code}`))
        setGuardians((current) => [...current, guardian])
        setAssignOpen(false)
        toast.success(t("staff.studentGuardians.assigned"))
        router.refresh()
    }

    async function unassign() {
        if (!removeTarget || busyID || !canUnassign) return
        setBusyID(removeTarget.id)
        const result = await unassignStudentGuardian(studentID, removeTarget.id)
        setBusyID(null)
        if (!result.ok) return toast.error(t(`staff.studentGuardians.error.${result.code}`))
        setGuardians((current) => current.filter((guardian) => guardian.id !== removeTarget.id))
        setRemoveTarget(null)
        toast.success(t("staff.studentGuardians.removed"))
        router.refresh()
    }

    return <div className="space-y-6">
        <Button variant="ghost" className="-ml-3" onClick={() => router.back()}><ArrowLeft /> {t("common.back")}</Button>
        <header className="flex flex-wrap items-start justify-between gap-4">
            <div>
                <h1 className="text-3xl font-semibold tracking-tight">{t("staff.studentGuardians.metaTitle")}</h1>
                <p className="mt-2 text-muted-foreground">{studentName ?? t("staff.studentGuardians.description")}</p>
            </div>
            {canAssign && <Button onClick={() => { setQuery(""); setAssignOpen(true) }}><Plus /> {t("staff.studentGuardians.assign")}</Button>}
        </header>
        {guardians.length === 0 ? <Card><Empty><EmptyHeader>
            <EmptyMedia variant="icon"><UsersRound /></EmptyMedia>
            <EmptyTitle>{t("staff.studentGuardians.empty")}</EmptyTitle>
            <EmptyDescription>{t("staff.studentGuardians.emptyDescription")}</EmptyDescription>
        </EmptyHeader></Empty></Card> : <Card><CardContent className="divide-y p-0">
            {[...guardians].sort((a, b) => a.lastName.localeCompare(b.lastName)).map((guardian) => <div key={guardian.id} className="flex flex-wrap items-center justify-between gap-3 p-5">
                <div className="min-w-0">
                    <p className="font-medium">{guardian.name} {guardian.lastName}</p>
                    <div className="mt-1 flex flex-wrap gap-4 text-sm text-muted-foreground">
                        <span className="flex items-center gap-2"><Mail className="size-4" /> {guardian.email}</span>
                        {guardian.phone && <span className="flex items-center gap-2"><Phone className="size-4" /> {guardian.phone}</span>}
                    </div>
                </div>
                {canUnassign && <Button variant="outline" size="sm" disabled={busyID !== null} onClick={() => setRemoveTarget(guardian)}><Unlink /> {t("staff.studentGuardians.unassign")}</Button>}
            </div>)}
        </CardContent></Card>}
        <Button variant="outline" render={<Link href={`/app/staff/students/${studentID}`} />}>{t("staff.studentGuardians.profile")}</Button>

        <Dialog open={assignOpen} onOpenChange={(open) => { if (!busyID) setAssignOpen(open) }}>
            <DialogPopup className="sm:max-w-xl">
                <DialogHeader><DialogTitle>{t("staff.studentGuardians.assign")}</DialogTitle><DialogDescription>{t("staff.studentGuardians.assignDescription")}</DialogDescription></DialogHeader>
                <DialogPanel className="space-y-4">
                    {candidatesUnavailable ? <p className="text-sm text-muted-foreground">{t("staff.studentGuardians.candidatesUnavailable")}</p> : <>
                        <Input value={query} onChange={(event) => setQuery(event.target.value)} placeholder={t("staff.studentGuardians.search")} aria-label={t("staff.studentGuardians.search")} />
                        <div className="max-h-80 space-y-2 overflow-y-auto">
                            {candidates.length === 0 && <p className="py-4 text-center text-sm text-muted-foreground">{t("staff.studentGuardians.noCandidates")}</p>}
                            {candidates.map((guardian) => <div className="flex items-center justify-between gap-3 rounded-lg border p-3" key={guardian.id}>
                                <div className="min-w-0"><p className="truncate text-sm font-medium">{guardian.name} {guardian.lastName}</p><p className="truncate text-xs text-muted-foreground">{guardian.email}</p></div>
                                <Button size="sm" disabled={busyID !== null} loading={busyID === guardian.id} onClick={() => assign(guardian)}>{t("staff.studentGuardians.assign")}</Button>
                            </div>)}
                        </div>
                    </>}
                </DialogPanel>
                <DialogFooter><Button variant="outline" disabled={busyID !== null} onClick={() => setAssignOpen(false)}>{t("common.cancel")}</Button></DialogFooter>
            </DialogPopup>
        </Dialog>
        <AlertDialog open={removeTarget !== null} onOpenChange={(open) => { if (!open && !busyID) setRemoveTarget(null) }}>
            <AlertDialogPopup><AlertDialogHeader>
                <AlertDialogTitle>{t("staff.studentGuardians.removeTitle")}</AlertDialogTitle>
                <AlertDialogDescription>{t("staff.studentGuardians.removeDescription", {name: `${removeTarget?.name ?? ""} ${removeTarget?.lastName ?? ""}`})}</AlertDialogDescription>
            </AlertDialogHeader><AlertDialogFooter>
                <AlertDialogClose render={<Button variant="outline" disabled={busyID !== null} />}>{t("common.cancel")}</AlertDialogClose>
                <Button loading={busyID !== null} disabled={busyID !== null} onClick={unassign}>{t("staff.studentGuardians.unassign")}</Button>
            </AlertDialogFooter></AlertDialogPopup>
        </AlertDialog>
    </div>
}

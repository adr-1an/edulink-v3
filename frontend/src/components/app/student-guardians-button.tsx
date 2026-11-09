"use client"

import Link from "next/link"
import {UsersRound} from "lucide-react"
import {Button} from "@/components/ui/button"
import {useLocale} from "@/i18n/provider"
import {hasSchoolPermission, type SchoolAccess} from "@/lib/school_access"

export default function StudentGuardiansButton({studentID, access}: {
    studentID: string
    access: SchoolAccess
}) {
    const {t} = useLocale()
    if (!hasSchoolPermission(access, "student.view")) return null
    const href = `/app/staff/students/${studentID}/guardians`
    return <Button size="sm" variant="outline" render={<Link href={href} />}><UsersRound /> {t("staff.studentGuardians.button")}</Button>
}

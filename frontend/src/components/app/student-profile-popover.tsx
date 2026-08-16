"use client"

import {Mail, Phone} from "lucide-react"
import {Avatar, AvatarFallback} from "@/components/ui/avatar"
import {Popover, PopoverDescription, PopoverPopup, PopoverTitle, PopoverTrigger} from "@/components/ui/popover"
import {useLocale} from "@/i18n/provider"

export interface PortalStudentSummary {
    id: string
    email: string
    phone: string
    name: string
    lastName: string
}

export function portalStudentDisplayName(student: PortalStudentSummary) {
    return `${student.name} ${student.lastName}`.trim()
}

function initials(name: string) {
    return name
        .split(/\s+/)
        .filter(Boolean)
        .slice(0, 2)
        .map((part) => part[0])
        .join("")
        .toLocaleUpperCase()
}

export default function StudentProfilePopover({student, showEmail = false}: {
    student: PortalStudentSummary
    showEmail?: boolean
}) {
    const {t} = useLocale()
    const name = portalStudentDisplayName(student)

    return (
        <Popover>
            <PopoverTrigger
                className="group flex cursor-pointer items-center gap-2.5 rounded-lg text-left outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                type="button"
            >
                <Avatar className="size-9 border transition-opacity group-hover:opacity-80">
                    <AvatarFallback>{initials(name)}</AvatarFallback>
                </Avatar>
                <span className="min-w-0">
                    <span className="block truncate font-medium underline-offset-4 group-hover:underline group-focus-visible:underline">{name}</span>
                    {showEmail && <span className="block truncate text-xs text-muted-foreground">{student.email}</span>}
                </span>
            </PopoverTrigger>
            <PopoverPopup className="w-72" align="start">
                <div className="flex min-w-0 items-center gap-3">
                    <Avatar className="size-11 border">
                        <AvatarFallback>{initials(name)}</AvatarFallback>
                    </Avatar>
                    <div className="min-w-0">
                        <PopoverTitle className="truncate text-base">{name}</PopoverTitle>
                        <PopoverDescription>{t("common.student")}</PopoverDescription>
                    </div>
                </div>
                <div className="mt-4 space-y-2 border-t pt-3 text-sm">
                    <a className="flex min-w-0 items-center gap-2 text-muted-foreground underline-offset-4 hover:text-foreground hover:underline" href={`mailto:${student.email}`}>
                        <Mail className="size-4 shrink-0" />
                        <span className="truncate">{student.email}</span>
                    </a>
                    {student.phone ? (
                        <a className="flex min-w-0 items-center gap-2 text-muted-foreground underline-offset-4 hover:text-foreground hover:underline" href={`tel:${student.phone}`}>
                            <Phone className="size-4 shrink-0" />
                            <span className="truncate">{student.phone}</span>
                        </a>
                    ) : (
                        <span className="flex items-center gap-2 text-muted-foreground">
                            <Phone className="size-4" />
                            {t("common.notProvided")}
                        </span>
                    )}
                </div>
            </PopoverPopup>
        </Popover>
    )
}

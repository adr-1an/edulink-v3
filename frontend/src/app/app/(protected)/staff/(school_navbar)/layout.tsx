"use client"

import React, {useEffect, useMemo, useState, useSyncExternalStore} from "react"
import Link from "next/link"
import {useParams, usePathname} from "next/navigation"
import {
    ArrowLeft, GraduationCap, LayoutDashboard, Menu, ScrollText, Settings, ShieldCheck, Users,
    type LucideIcon,
} from "lucide-react"
import {Button} from "@/components/ui/button"
import {Separator} from "@/components/ui/separator"
import {
    Sheet, SheetFooter, SheetHeader, SheetPanel, SheetPopup, SheetTitle, SheetTrigger,
} from "@/components/ui/sheet"
import {hasSchoolPermission, type SchoolAccess} from "@/lib/school_access"
import {
    getActiveSchoolAccessSnapshot,
    getActiveSchoolSnapshot,
    getServerSchoolNavigationSnapshot,
    parseStoredSchoolAccess,
    rememberActiveSchool,
    rememberSchoolAccess,
    subscribeToSchoolNavigation,
} from "@/lib/school_navigation"
import {getSchoolNavigationAccess} from "./school_navigation_actions"
import {useLocale} from "@/i18n/provider"

interface NavigationItem {
    label: string
    href: string
    icon: LucideIcon
    active: boolean
}

function NavigationLinks({items, label, onNavigate}: {items: NavigationItem[]; label: string; onNavigate?: () => void}) {
    return (
        <nav className="flex flex-col gap-1" aria-label={label}>
            {items.map(({label, href, icon: Icon, active}) => (
                <Button
                    className={active
                        ? "w-full justify-start bg-primary/10 font-semibold text-primary hover:bg-primary/15 hover:text-primary"
                        : "w-full justify-start text-muted-foreground hover:text-foreground"}
                    size="sm"
                    variant="ghost"
                    render={<Link href={href} aria-current={active ? "page" : undefined} onClick={onNavigate} />}
                    key={href}
                >
                    <Icon /> {label}
                </Button>
            ))}
        </nav>
    )
}

export default function Layout({children}: {children: React.ReactNode}) {
    const pathname = usePathname()
    const {t} = useLocale()
    const params = useParams<{id?: string}>()
    const routeSchoolID = params.id
    const [mobileNavigationOpen, setMobileNavigationOpen] = useState(false)
    const rememberedSchoolID = useSyncExternalStore(
        subscribeToSchoolNavigation,
        getActiveSchoolSnapshot,
        getServerSchoolNavigationSnapshot,
    )
    const storedAccessRaw = useSyncExternalStore(
        subscribeToSchoolNavigation,
        getActiveSchoolAccessSnapshot,
        getServerSchoolNavigationSnapshot,
    )
    const storedAccess = useMemo(() => parseStoredSchoolAccess(storedAccessRaw), [storedAccessRaw])
    const [loadedAccess, setLoadedAccess] = useState<{schoolID: string; access: SchoolAccess} | null>(null)
    const schoolID = routeSchoolID ?? rememberedSchoolID ?? storedAccess?.schoolID ?? loadedAccess?.schoolID
    const access = loadedAccess && loadedAccess.schoolID === schoolID
        ? loadedAccess.access
        : storedAccess && storedAccess.schoolID === schoolID ? storedAccess.access : null

    useEffect(() => {
        if (routeSchoolID && routeSchoolID !== rememberedSchoolID) rememberActiveSchool(routeSchoolID)
    }, [routeSchoolID, rememberedSchoolID])

    useEffect(() => {
        if (!schoolID) return

        let active = true
        void getSchoolNavigationAccess(schoolID).then((nextAccess) => {
            if (!active || !nextAccess) return
            setLoadedAccess({schoolID, access: nextAccess})
            rememberSchoolAccess(schoolID, nextAccess)
        })

        return () => {
            active = false
        }
    }, [schoolID])

    const canViewStaff = hasSchoolPermission(access, "staff.view")
    const canViewStudents = hasSchoolPermission(access, "student.list")
    const canViewRoles = hasSchoolPermission(access, "role.list")
    const canViewSettings = hasSchoolPermission(access, "school.view")
    const canViewLogs = pathname.includes("/logs") || hasSchoolPermission(access, "log.list")
    const navigation: NavigationItem[] = [
        ...(schoolID ? [
            {
                label: t("navigation.dashboard"),
                href: `/app/staff/schools/${schoolID}`,
                icon: LayoutDashboard,
                active: pathname === `/app/staff/schools/${schoolID}`,
            },
            ...(canViewStaff ? [{
                label: t("navigation.staff"),
                href: `/app/staff/schools/${schoolID}/staff`,
                icon: Users,
                active: pathname.startsWith(`/app/staff/schools/${schoolID}/staff`),
            }] : []),
            ...(canViewStudents ? [{
                label: t("navigation.students"),
                href: `/app/staff/schools/${schoolID}/students`,
                icon: GraduationCap,
                active: pathname.startsWith(`/app/staff/schools/${schoolID}/students`)
                    || pathname.startsWith("/app/staff/students/"),
            }] : []),
            ...(canViewRoles ? [{
                label: t("navigation.roles"),
                href: `/app/staff/schools/${schoolID}/roles`,
                icon: ShieldCheck,
                active: pathname.startsWith(`/app/staff/schools/${schoolID}/roles`),
            }] : []),
            ...(canViewLogs ? [{
                label: t("navigation.auditLogs"),
                href: `/app/staff/schools/${schoolID}/logs`,
                icon: ScrollText,
                active: pathname.startsWith(`/app/staff/schools/${schoolID}/logs`),
            }] : []),
            ...(canViewSettings ? [{
                label: t("navigation.settings"),
                href: `/app/staff/schools/${schoolID}/settings`,
                icon: Settings,
                active: pathname.startsWith(`/app/staff/schools/${schoolID}/settings`),
            }] : []),
        ] : []),
    ]
    return (
        <div className="grid min-w-0 items-start gap-4 lg:grid-cols-[13.5rem_minmax(0,1fr)] lg:gap-7">
            <aside className="sticky top-8 hidden h-[calc(100dvh-4rem)] self-start overflow-y-auto rounded-2xl border bg-card p-3 shadow-xs lg:flex lg:flex-col">
                <NavigationLinks items={navigation} label={t("navigation.school")} />

                <div className="mt-auto pt-4">
                    <Separator className="mb-3" />
                    <Button className="w-full justify-start text-muted-foreground" size="sm" variant="ghost" render={<Link href="/app" />}>
                        <ArrowLeft /> {t("navigation.leaveWorkspace")}
                    </Button>
                </div>
            </aside>

            <div className="flex min-w-0 items-center justify-between gap-3 rounded-2xl border bg-card px-3 py-2.5 shadow-xs lg:hidden">
                <p className="px-1 text-sm font-semibold">{t("navigation.navigation")}</p>

                <Sheet open={mobileNavigationOpen} onOpenChange={setMobileNavigationOpen}>
                    <SheetTrigger render={<Button size="icon-sm" variant="outline" aria-label={t("navigation.open")} />}>
                        <Menu />
                    </SheetTrigger>
                    <SheetPopup className="max-w-xs" side="left">
                        <SheetHeader className="border-b px-5 py-5">
                            <SheetTitle>{t("navigation.menu")}</SheetTitle>
                        </SheetHeader>
                        <SheetPanel className="flex-1 p-3">
                            <NavigationLinks items={navigation} label={t("navigation.school")} onNavigate={() => setMobileNavigationOpen(false)} />
                        </SheetPanel>
                        <SheetFooter className="border-t bg-transparent p-3">
                            <Button
                                className="w-full justify-start text-muted-foreground"
                                variant="ghost"
                                render={<Link href="/app" onClick={() => setMobileNavigationOpen(false)} />}
                            >
                                <ArrowLeft /> {t("navigation.leaveWorkspace")}
                            </Button>
                        </SheetFooter>
                    </SheetPopup>
                </Sheet>
            </div>

            <main className="min-w-0 lg:col-start-2">{children}</main>
        </div>
    )
}

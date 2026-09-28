"use client"

import React, {useEffect, useMemo, useState, useSyncExternalStore} from "react"
import Link from "next/link"
import {useParams, usePathname} from "next/navigation"
import {
    ArrowLeft, GraduationCap, LayoutDashboard, Menu, Moon, ScrollText, Settings, ShieldCheck, Sun, Users, UsersRound,
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
import {useTheme} from "@/components/app/theme_provider"

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
                        ? "h-9 w-full justify-start rounded-lg border-transparent bg-primary px-2 font-medium text-primary-foreground hover:bg-primary/90"
                        : "h-9 w-full justify-start rounded-lg border-transparent px-2 font-normal text-sidebar-foreground/80 hover:bg-sidebar-accent hover:text-sidebar-foreground"}
                    size="sm"
                    variant="ghost"
                    render={<Link href={href} aria-current={active ? "page" : undefined} onClick={onNavigate} />}
                    key={href}
                >
                    <span className={active
                        ? "flex size-6 shrink-0 items-center justify-center rounded-md bg-white/16 text-white"
                        : "flex size-6 shrink-0 items-center justify-center rounded-md bg-background/75 text-sidebar-foreground shadow-sm ring-1 ring-sidebar-border"}>
                        <Icon className="size-3.5" />
                    </span>
                    {label}
                </Button>
            ))}
        </nav>
    )
}

export default function Layout({children}: {children: React.ReactNode}) {
    const pathname = usePathname()
    const {t} = useLocale()
    const {theme, toggleTheme} = useTheme()
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
    const canViewGuardians = hasSchoolPermission(access, "guardian.list")
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
            ...(canViewGuardians ? [{
                label: t("navigation.guardians"),
                href: `/app/staff/schools/${schoolID}/guardians`,
                icon: UsersRound,
                active: pathname.startsWith(`/app/staff/schools/${schoolID}/guardians`),
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
    const activeNavigation = navigation.find((item) => item.active)
    return (
        <div className="flex min-h-svh w-full bg-background">
            <aside className="sticky top-0 hidden h-svh w-60 shrink-0 flex-col overflow-y-auto border-r border-sidebar-border bg-sidebar/95 px-3 py-4 text-sidebar-foreground backdrop-blur-2xl lg:flex">
                <Link className="mb-6 flex items-center gap-2.5 px-2" href="/app">
                    <span className="flex size-8 items-center justify-center rounded-lg bg-primary text-sm font-semibold text-primary-foreground">E</span>
                    <span>
                        <span className="block text-sm font-semibold tracking-[-0.02em]">EduLink</span>
                        <span className="block text-xs text-muted-foreground">{t("navigation.school")}</span>
                    </span>
                </Link>

                <NavigationLinks items={navigation} label={t("navigation.school")} />

                <div className="mt-auto pt-4">
                    <Separator className="mb-3" />
                    <Button className="mb-1 w-full justify-start text-muted-foreground" size="sm" variant="ghost" onClick={toggleTheme}>
                        {theme === "dark" ? <Sun /> : <Moon />} {t(theme === "dark" ? "common.lightMode" : "common.darkMode")}
                    </Button>
                    <Button className="w-full justify-start text-muted-foreground" size="sm" variant="ghost" render={<Link href="/app" />}>
                        <ArrowLeft /> {t("navigation.leaveWorkspace")}
                    </Button>
                </div>
            </aside>

            <div className="min-w-0 flex-1">
                <header className="sticky top-0 z-30 flex h-14 items-center gap-3 border-b border-border bg-background/95 px-3 backdrop-blur-2xl sm:px-5">
                    <Sheet open={mobileNavigationOpen} onOpenChange={setMobileNavigationOpen}>
                        <SheetTrigger className="lg:hidden" render={<Button size="icon-sm" variant="outline" aria-label={t("navigation.open")} />}>
                            <Menu />
                        </SheetTrigger>
                        <SheetPopup className="max-w-xs" side="left">
                            <SheetHeader className="border-b px-5 py-5">
                                <SheetTitle>{t("navigation.menu")}</SheetTitle>
                            </SheetHeader>
                            <SheetPanel className="flex-1 p-3">
                                <NavigationLinks items={navigation} label={t("navigation.school")} onNavigate={() => setMobileNavigationOpen(false)} />
                            </SheetPanel>
                            <SheetFooter className="border-t bg-popover p-3">
                                <Button className="w-full justify-start text-muted-foreground" variant="ghost" onClick={toggleTheme}>
                                    {theme === "dark" ? <Sun /> : <Moon />} {t(theme === "dark" ? "common.lightMode" : "common.darkMode")}
                                </Button>
                                <Button className="w-full justify-start text-muted-foreground" variant="ghost" render={<Link href="/app" onClick={() => setMobileNavigationOpen(false)} />}>
                                    <ArrowLeft /> {t("navigation.leaveWorkspace")}
                                </Button>
                            </SheetFooter>
                        </SheetPopup>
                    </Sheet>
                    <p className="truncate text-sm font-semibold tracking-[-0.015em]">{activeNavigation?.label ?? t("navigation.navigation")}</p>
                    <Button className="ml-auto hidden sm:inline-flex" size="sm" variant="outline" render={<Link href="/app" />}>
                        <ArrowLeft /> {t("navigation.leaveWorkspace")}
                    </Button>
                </header>
                <main className="min-w-0 p-4 sm:p-6 lg:p-8">{children}</main>
            </div>
        </div>
    )
}

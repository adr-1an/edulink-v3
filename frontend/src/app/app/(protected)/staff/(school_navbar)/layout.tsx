"use client"

import React, {useEffect, useState, useSyncExternalStore} from "react"
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
import {
    getActiveSchoolSnapshot,
    getServerSchoolNavigationSnapshot,
    rememberActiveSchool,
    subscribeToSchoolNavigation,
} from "@/lib/school_navigation"
import {getSchoolNavigationItems} from "@/lib/school_navigation_items"
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
                    render={<Link href={href} prefetch={false} aria-current={active ? "page" : undefined} onClick={onNavigate} />}
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
    const schoolID = routeSchoolID ?? rememberedSchoolID

    useEffect(() => {
        if (routeSchoolID && routeSchoolID !== rememberedSchoolID) rememberActiveSchool(routeSchoolID)
    }, [routeSchoolID, rememberedSchoolID])

    const icons = {
        dashboard: LayoutDashboard,
        staff: Users,
        students: GraduationCap,
        guardians: UsersRound,
        roles: ShieldCheck,
        auditLogs: ScrollText,
        settings: Settings,
    }
    const navigation: NavigationItem[] = getSchoolNavigationItems(schoolID, pathname).map(({key, ...item}) => ({
        ...item,
        label: t(`navigation.${key}`),
        icon: icons[key],
    }))
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

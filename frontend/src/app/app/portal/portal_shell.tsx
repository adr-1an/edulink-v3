"use client"

import {useState} from "react"
import Link from "next/link"
import {usePathname, useRouter} from "next/navigation"
import {BookOpen, ClipboardList, GraduationCap, LogOut, Menu, Moon, Sun, UserRound} from "lucide-react"
import {toast, Toaster} from "sonner"
import {useTheme} from "@/components/app/theme_provider"
import {Badge} from "@/components/ui/badge"
import {Button} from "@/components/ui/button"
import {
    Sheet, SheetDescription, SheetFooter, SheetHeader, SheetPanel, SheetPopup, SheetTitle, SheetTrigger,
} from "@/components/ui/sheet"
import {handlePortalLogout} from "./actions"
import {useLocale} from "@/i18n/provider"

export default function PortalShell({accountType, children}: {
    accountType: "student" | "guardian"
    children: React.ReactNode
}) {
    const pathname = usePathname()
    const router = useRouter()
    const {theme, toggleTheme} = useTheme()
    const {t} = useLocale()
    const [menuOpen, setMenuOpen] = useState(false)
    const [loggingOut, setLoggingOut] = useState(false)
    const home = `/app/portal/${accountType}`

    async function logout() {
        setLoggingOut(true)
        const result = await handlePortalLogout()
        if (!result.ok) {
            setLoggingOut(false)
            toast.error(t("portal.logoutError"))
            return
        }
        router.replace("/app/portal")
        router.refresh()
    }

    const navigation = accountType === "student"
        ? [
            {label: t("portal.courses"), href: home, icon: BookOpen},
            {label: t("portal.assignments"), href: `${home}/assignments`, icon: ClipboardList},
            {label: t("common.profile"), href: `${home}/profile`, icon: UserRound},
        ]
        : [
            {label: t("portal.overview"), href: home, icon: GraduationCap},
            {label: t("common.profile"), href: `${home}/profile`, icon: UserRound},
        ]

    function isActive(href: string) {
        if (href === home) {
            return pathname === home
                || pathname.startsWith(`${home}/courses/`)
                || pathname.startsWith(`${home}/posts/`)
        }

        return pathname.startsWith(href)
    }

    return (
        <div className="flex min-h-svh w-full bg-background text-foreground">
            <aside className="sticky top-0 hidden h-svh w-60 shrink-0 flex-col border-r border-sidebar-border bg-sidebar/95 px-3 py-4 text-sidebar-foreground backdrop-blur-2xl md:flex">
                <Link className="mb-6 flex min-w-0 items-center gap-2.5 px-2" href={home}>
                    <span className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-primary text-primary-foreground"><GraduationCap className="size-4" /></span>
                    <span className="min-w-0">
                        <span className="block truncate text-sm font-semibold tracking-[-0.02em]">{t("portal.brand")}</span>
                        <span className="block truncate text-xs capitalize text-muted-foreground">{accountType}</span>
                    </span>
                </Link>
                <nav className="space-y-1" aria-label={t("portal.navigation")}>
                    {navigation.map(({label, href, icon: Icon}) => (
                        <Button className="h-9 w-full justify-start rounded-lg px-2 font-normal" variant={isActive(href) ? "default" : "ghost"} size="sm" render={<Link href={href} />} key={href}>
                            <span className={isActive(href)
                                ? "flex size-6 shrink-0 items-center justify-center rounded-md bg-white/16 text-white"
                                : "flex size-6 shrink-0 items-center justify-center rounded-md bg-background/75 text-sidebar-foreground shadow-sm ring-1 ring-sidebar-border"}>
                                <Icon className="size-3.5" />
                            </span>
                            {label}
                        </Button>
                    ))}
                </nav>
                <div className="mt-auto space-y-1 border-t border-sidebar-border pt-3">
                    <Button className="w-full justify-start" variant="ghost" size="sm" onClick={toggleTheme}>
                        {theme === "dark" ? <Sun /> : <Moon />} {t(theme === "dark" ? "common.lightMode" : "common.darkMode")}
                    </Button>
                    <Button className="w-full justify-start text-muted-foreground" variant="ghost" size="sm" loading={loggingOut} onClick={logout}><LogOut /> {t("common.logOut")}</Button>
                </div>
            </aside>

            <div className="min-w-0 flex-1">
                <header className="sticky top-0 z-40 flex h-14 items-center gap-3 border-b border-border bg-background/95 px-3 backdrop-blur-2xl sm:px-5">
                    <Link className="flex min-w-0 items-center gap-2.5 md:hidden" href={home}>
                        <span className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-primary text-primary-foreground"><GraduationCap className="size-4" /></span>
                        <span className="truncate text-sm font-semibold">{t("portal.brand")}</span>
                    </Link>
                    <p className="hidden truncate text-sm font-semibold md:block">{navigation.find(({href}) => isActive(href))?.label ?? t("portal.overview")}</p>
                    <Badge className="capitalize" variant="secondary">{accountType}</Badge>
                    <div className="ml-auto hidden items-center gap-1 md:flex">
                        <Button size="icon-sm" variant="outline" aria-label={`Switch to ${theme === "dark" ? "light" : "dark"} mode`} onClick={toggleTheme}>
                            {theme === "dark" ? <Sun /> : <Moon />}
                        </Button>
                    </div>
                    <Sheet open={menuOpen} onOpenChange={setMenuOpen}>
                        <SheetTrigger className="ml-auto md:hidden" render={<Button size="icon-sm" variant="outline" aria-label={t("portal.openMenu")} />}><Menu /></SheetTrigger>
                        <SheetPopup className="max-w-xs" side="right">
                            <SheetHeader>
                                <SheetTitle>{t("portal.menu")}</SheetTitle>
                                <SheetDescription className="capitalize">
                                    {t("portal.signedInAs", {accountType: t(accountType === "student" ? "common.student" : "common.guardian")})}
                                </SheetDescription>
                            </SheetHeader>
                            <SheetPanel className="space-y-1 p-3">
                                {navigation.map(({label, href, icon: Icon}) => (
                                    <Button className="w-full justify-start" variant={isActive(href) ? "secondary" : "ghost"} render={<Link href={href} onClick={() => setMenuOpen(false)} />} key={href}><Icon /> {label}</Button>
                                ))}
                                <Button className="w-full justify-start" variant="ghost" onClick={toggleTheme}>
                                    {theme === "dark" ? <Sun /> : <Moon />} {t(theme === "dark" ? "common.lightMode" : "common.darkMode")}
                                </Button>
                            </SheetPanel>
                            <SheetFooter>
                                <Button className="w-full" variant="outline" loading={loggingOut} onClick={logout}><LogOut /> {t("common.logOut")}</Button>
                            </SheetFooter>
                        </SheetPopup>
                    </Sheet>
                </header>

                <main className="mx-auto w-full max-w-7xl p-4 sm:p-6 lg:p-8">{children}</main>
            </div>
            <Toaster richColors visibleToasts={5} position="top-right" theme={theme} />
        </div>
    )
}

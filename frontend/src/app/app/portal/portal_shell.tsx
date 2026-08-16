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
        : [{label: t("portal.overview"), href: home, icon: GraduationCap}]

    function isActive(href: string) {
        if (href === home) {
            return pathname === home
                || pathname.startsWith(`${home}/courses/`)
                || pathname.startsWith(`${home}/posts/`)
        }

        return pathname.startsWith(href)
    }

    return (
        <div className="min-h-screen bg-muted/25 text-foreground">
            <header className="sticky top-0 z-40 border-b bg-background/90 backdrop-blur-xl">
                <div className="mx-auto flex h-16 max-w-7xl items-center gap-4 px-4 sm:px-6 lg:px-8">
                    <Link className="flex min-w-0 items-center gap-2.5" href={home}>
                        <span className="flex size-9 shrink-0 items-center justify-center rounded-xl bg-primary text-primary-foreground"><GraduationCap className="size-4.5" /></span>
                        <span className="truncate font-semibold">{t("portal.brand")}</span>
                    </Link>
                    <Badge className="capitalize" variant="secondary">{accountType}</Badge>

                    <nav className="ml-6 hidden items-center gap-1 md:flex" aria-label={t("portal.navigation")}>
                        {navigation.map(({label, href, icon: Icon}) => (
                            <Button variant={isActive(href) ? "secondary" : "ghost"} size="sm" render={<Link href={href} />} key={href}><Icon /> {label}</Button>
                        ))}
                    </nav>

                    <div className="ml-auto hidden items-center gap-1 md:flex">
                        <Button size="icon-sm" variant="ghost" aria-label={`Switch to ${theme === "dark" ? "light" : "dark"} mode`} onClick={toggleTheme}>
                            {theme === "dark" ? <Sun /> : <Moon />}
                        </Button>
                        <Button variant="ghost" size="sm" loading={loggingOut} onClick={logout}><LogOut /> {t("common.logOut")}</Button>
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
                </div>
            </header>

            <main className="mx-auto w-full max-w-7xl px-4 py-6 sm:px-6 sm:py-8 lg:px-8">{children}</main>
            <Toaster richColors visibleToasts={5} position="top-right" theme={theme} />
        </div>
    )
}

"use client"

import Link from "next/link"
import {GraduationCap} from "lucide-react"
import {useLocale} from "@/i18n/provider"

export default function AuthPageShell({children}: {children: React.ReactNode}) {
    const {t} = useLocale()
    return (
        <div className="flex min-h-svh flex-col bg-background text-foreground">
            <header className="flex items-center justify-between border-b px-5 py-4 sm:px-8">
                <Link className="text-lg font-semibold tracking-tight" href="/">EduLink</Link>
                <Link
                    className="inline-flex items-center gap-2 rounded-full px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-card hover:text-foreground"
                    href="/app/portal"
                >
                    <GraduationCap className="size-4" />
                    <span className="hidden sm:inline">{t("auth.portal")}</span>
                    <span className="sm:hidden">{t("landing.portalShort")}</span>
                </Link>
            </header>

            <main className="flex flex-1 items-center justify-center px-4 py-8 sm:px-6 sm:py-12">
                {children}
            </main>

            <footer className="flex flex-wrap items-center justify-center gap-x-5 gap-y-2 border-t px-6 py-5 text-xs text-muted-foreground">
                <Link className="transition-colors hover:text-foreground" href="/">{t("auth.home")}</Link>
                <Link className="transition-colors hover:text-foreground" href="/legal/terms">{t("common.terms")}</Link>
                <Link className="transition-colors hover:text-foreground" href="/legal/privacy">{t("common.privacy")}</Link>
            </footer>
        </div>
    )
}

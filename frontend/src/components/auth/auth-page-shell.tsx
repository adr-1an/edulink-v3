"use client"

import Link from "next/link"
import {GraduationCap} from "lucide-react"
import {useLocale} from "@/i18n/provider"

export default function AuthPageShell({children}: {children: React.ReactNode}) {
    const {t} = useLocale()
    return (
        <div className="relative flex min-h-svh flex-col overflow-hidden bg-background text-foreground">
            <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_15%_10%,color-mix(in_oklab,var(--primary)_10%,transparent),transparent_34%),radial-gradient(circle_at_85%_85%,color-mix(in_oklab,var(--primary)_7%,transparent),transparent_32%)]" />
            <div className="pointer-events-none absolute inset-0 opacity-35 [background-image:linear-gradient(to_right,var(--border)_1px,transparent_1px),linear-gradient(to_bottom,var(--border)_1px,transparent_1px)] [background-size:64px_64px] [mask-image:linear-gradient(to_bottom,black,transparent_80%)]" />

            <header className="relative z-10 flex items-center justify-between px-5 py-5 sm:px-8">
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

            <main className="relative z-10 flex flex-1 items-center justify-center px-4 py-8 sm:px-6 sm:py-12">
                {children}
            </main>

            <footer className="relative z-10 flex flex-wrap items-center justify-center gap-x-5 gap-y-2 px-6 py-5 text-xs text-muted-foreground">
                <Link className="transition-colors hover:text-foreground" href="/">{t("auth.home")}</Link>
                <Link className="transition-colors hover:text-foreground" href="/legal/terms">{t("common.terms")}</Link>
                <Link className="transition-colors hover:text-foreground" href="/legal/privacy">{t("common.privacy")}</Link>
            </footer>
        </div>
    )
}

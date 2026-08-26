"use client"

import Link from "next/link"
import {ArrowRight} from "lucide-react"
import LanguageSwitcher from "@/components/app/language-switcher"
import {useLocale} from "@/i18n/provider"

export default function LandingPage() {
    const {t} = useLocale()

    return (
        <div className="min-h-screen bg-white text-[#1d1d1f] selection:bg-blue-200">
            <header className="sticky top-0 z-50 border-b border-black/[0.06] bg-white/80 backdrop-blur-xl saturate-150">
                <nav className="mx-auto flex h-12 max-w-5xl items-center justify-between gap-2 px-4 sm:px-6" aria-label={t("landing.navigationLabel")}>
                    <Link className="shrink-0 text-sm font-semibold tracking-[-0.02em] max-[390px]:hidden" href="/">EduLink</Link>

                    <div className="flex min-w-0 items-center justify-end gap-0.5 sm:gap-1.5">
                        <LanguageSwitcher className="border-transparent bg-transparent text-[#424245] shadow-none hover:bg-black/[0.04]" compact />
                        <Link className="inline-flex h-8 shrink-0 items-center rounded-full px-2.5 text-xs text-[#424245] transition-colors hover:text-black sm:px-3" href="/app/portal" prefetch={false}>
                            <span className="sm:hidden">{t("landing.portalShort")}</span>
                            <span className="hidden sm:inline">{t("landing.studentParentLogin")}</span>
                        </Link>
                        <Link className="inline-flex h-8 shrink-0 items-center rounded-full px-2.5 text-xs text-[#424245] transition-colors hover:text-black sm:px-3" href="/auth/login" prefetch={false}>
                            <span className="sm:hidden">{t("landing.staffShort")}</span>
                            <span className="hidden sm:inline">{t("landing.staffLogin")}</span>
                        </Link>
                        <Link className="inline-flex h-8 shrink-0 items-center gap-1 rounded-full bg-[#0071e3] px-3 text-xs font-medium text-white transition-colors hover:bg-[#0077ed] sm:px-3.5" href="/auth/register" prefetch={false}>
                            <span className="hidden min-[420px]:inline">{t("landing.getStarted")}</span>
                            <ArrowRight className="size-3" />
                        </Link>
                    </div>
                </nav>
            </header>

            <main>
                <section>
                    <div className="mx-auto flex min-h-[calc(100svh-3rem)] max-w-6xl items-center justify-center px-5 py-24 sm:px-6 sm:py-32">
                        <div className="max-w-5xl text-center">
                            <p className="text-lg font-semibold tracking-[-0.02em] sm:text-xl">{t("landing.eyebrow")}</p>
                            <h1 className="mx-auto mt-4 text-balance text-[clamp(3.5rem,9vw,6.75rem)] font-semibold leading-[0.94] tracking-[-0.065em]">
                                {t("landing.heroTitle")}
                            </h1>
                            <p className="mx-auto mt-7 max-w-2xl text-pretty text-lg font-medium leading-7 tracking-[-0.015em] text-[#6e6e73] sm:text-xl sm:leading-8">
                                {t("landing.heroDescription")}
                            </p>

                            <div className="mt-9 flex flex-col items-center justify-center gap-4 sm:flex-row sm:gap-6">
                                <Link className="inline-flex h-11 items-center justify-center rounded-full bg-[#0071e3] px-6 text-sm font-medium text-white transition-colors hover:bg-[#0077ed]" href="/auth/register">
                                    {t("landing.createStaffAccount")}
                                </Link>
                                <Link className="group inline-flex items-center gap-1.5 text-sm font-medium text-[#0066cc] hover:underline" href="/app/portal">
                                    {t("landing.openPortal")} <ArrowRight className="size-3.5 transition-transform group-hover:translate-x-0.5" />
                                </Link>
                            </div>
                        </div>
                    </div>
                </section>

                <section className="bg-[#f5f5f7]">
                    <div className="mx-auto max-w-6xl px-5 py-24 sm:px-6 sm:py-32 lg:py-40">
                        <div className="mx-auto max-w-4xl text-center">
                            <p className="text-sm font-semibold text-[#6e6e73]">{t("landing.overviewEyebrow")}</p>
                            <h2 className="mt-4 text-balance text-4xl font-semibold leading-tight tracking-[-0.05em] sm:text-6xl">{t("landing.overviewTitle")}</h2>
                        </div>

                        <div className="mt-20 grid border-t border-black/10 md:grid-cols-3 md:divide-x md:divide-black/10">
                            <OverviewItem title={t("landing.organizationTitle")} description={t("landing.organizationDescription")} />
                            <OverviewItem title={t("landing.courseWorkTitle")} description={t("landing.courseWorkDescription")} />
                            <OverviewItem title={t("landing.portalTitle")} description={t("landing.portalDescription")} />
                        </div>
                    </div>
                </section>

                <section>
                    <div className="mx-auto max-w-6xl px-5 py-24 sm:px-6 sm:py-32 lg:py-40">
                        <div className="mx-auto max-w-4xl text-center">
                            <p className="text-sm font-semibold text-[#6e6e73]">{t("landing.spacesEyebrow")}</p>
                            <h2 className="mt-4 text-balance text-4xl font-semibold leading-tight tracking-[-0.05em] sm:text-6xl">{t("landing.spacesTitle")}</h2>
                        </div>

                        <div className="mt-20 grid gap-12 border-t border-black/10 pt-10 sm:grid-cols-2 sm:gap-16">
                            <div>
                                <h3 className="text-2xl font-semibold tracking-[-0.035em] sm:text-3xl">{t("landing.staffWorkspaceTitle")}</h3>
                                <p className="mt-4 max-w-md text-base font-medium leading-7 text-[#6e6e73] sm:text-lg">{t("landing.staffWorkspaceDescription")}</p>
                            </div>
                            <div>
                                <h3 className="text-2xl font-semibold tracking-[-0.035em] sm:text-3xl">{t("landing.studentPortalTitle")}</h3>
                                <p className="mt-4 max-w-md text-base font-medium leading-7 text-[#6e6e73] sm:text-lg">{t("landing.studentPortalDescription")}</p>
                            </div>
                            <div className="border-t border-black/10 pt-8 sm:col-span-2">
                                <h3 className="text-lg font-semibold tracking-[-0.02em]">{t("landing.supportingTitle")}</h3>
                                <p className="mt-3 max-w-2xl text-base leading-7 text-[#6e6e73]">{t("landing.supportingDescription")}</p>
                            </div>
                        </div>
                    </div>
                </section>

                <section className="bg-[#f5f5f7] px-5 py-24 sm:px-6 sm:py-32">
                    <div className="mx-auto max-w-4xl text-center">
                        <div>
                            <h2 className="text-balance text-4xl font-semibold tracking-[-0.05em] sm:text-5xl">{t("landing.ctaTitle")}</h2>
                            <p className="mx-auto mt-4 max-w-xl text-base font-medium leading-7 text-[#6e6e73] sm:text-lg">{t("landing.ctaDescription")}</p>
                        </div>
                        <Link className="mt-8 inline-flex h-11 shrink-0 items-center justify-center rounded-full bg-[#0071e3] px-6 text-sm font-medium text-white transition-colors hover:bg-[#0077ed]" href="/auth/register">
                            {t("landing.createStaffAccount")}
                        </Link>
                    </div>
                </section>
            </main>

            <footer className="bg-[#f5f5f7] px-5 pb-7 pt-0 sm:px-6">
                <div className="mx-auto flex max-w-5xl flex-col items-center justify-between gap-5 border-t border-black/10 pt-6 sm:flex-row">
                    <div className="text-center sm:text-left">
                        <Link className="text-sm font-semibold" href="/">EduLink</Link>
                        <p className="mt-1 text-xs text-[#6e6e73]">{t("landing.footerTagline")}</p>
                    </div>
                    <div className="flex flex-wrap justify-center gap-x-5 gap-y-2 text-xs text-[#6e6e73]">
                        <Link className="hover:text-black" href="/legal/privacy">{t("common.privacy")}</Link>
                        <Link className="hover:text-black" href="/legal/terms">{t("common.terms")}</Link>
                        <Link className="hover:text-black" href="/auth/login">{t("landing.staffLogin")}</Link>
                        <Link className="hover:text-black" href="/app/portal">{t("landing.studentParentLogin")}</Link>
                    </div>
                </div>
            </footer>
        </div>
    )
}

function OverviewItem({title, description}: {title: string; description: string}) {
    return (
        <article className="border-t border-black/10 py-8 first:border-t-0 md:border-t-0 md:px-8 md:py-10 md:first:pl-0 md:last:pr-0">
            <h3 className="text-xl font-semibold tracking-[-0.025em]">{title}</h3>
            <p className="mt-3 text-base leading-7 text-[#6e6e73]">{description}</p>
        </article>
    )
}

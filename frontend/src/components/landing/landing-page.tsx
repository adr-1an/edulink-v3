"use client"

import Link from "next/link"
import {
    ArrowRight,
    BarChart3,
    BellRing,
    BookOpenCheck,
    CheckCircle2,
    ClipboardCheck,
    GraduationCap,
    HeartHandshake,
    Layers3,
    LockKeyhole,
    School,
    ShieldCheck,
    Sparkles,
    UsersRound,
} from "lucide-react"
import {motion, useReducedMotion} from "framer-motion"
import AppDemos from "@/components/landing/app-demos"
import LanguageSwitcher from "@/components/app/language-switcher"
import {useLocale} from "@/i18n/provider"

const features = [
    {
        icon: GraduationCap,
        title: "Every student, understood",
        description: "Profiles, grades, courses, and progress live together, giving the right people a clear view without the spreadsheet chase.",
        className: "md:col-span-2",
        color: "bg-blue-600",
    },
    {
        icon: BookOpenCheck,
        title: "Courses that feel alive",
        description: "Share updates, organize learning, and keep every class moving in one focused space.",
        className: "",
        color: "bg-violet-600",
    },
    {
        icon: UsersRound,
        title: "Staff, roles, and access",
        description: "Invite your team and give everyone exactly the access they need—nothing more, nothing less.",
        className: "",
        color: "bg-emerald-600",
    },
    {
        icon: BarChart3,
        title: "Grades without the guesswork",
        description: "Make progress visible to staff, students, and parents with a consistent source of truth.",
        className: "md:col-span-2",
        color: "bg-amber-500",
    },
]

const roles = [
    {icon: School, label: "School leaders", copy: "See the whole school clearly and keep teams aligned."},
    {icon: ClipboardCheck, label: "Staff", copy: "Spend less time coordinating and more time teaching."},
    {icon: GraduationCap, label: "Students", copy: "Know what’s happening, what’s next, and how you’re doing."},
    {icon: HeartHandshake, label: "Parents", copy: "Stay meaningfully connected without chasing updates."},
]

export default function LandingPage() {
    const {t} = useLocale()

    return (
        <div className="min-h-screen overflow-hidden bg-[#fbfcff] text-slate-950 selection:bg-blue-200">
            <header className="fixed inset-x-0 top-0 z-50 border-b border-slate-200 bg-gradient-to-b from-white to-slate-50 shadow-sm shadow-slate-950/5">
                <nav className="mx-auto flex h-20 max-w-7xl items-center justify-between gap-2 px-3 sm:px-6 lg:px-8" aria-label="Main navigation">
                    <Link className="group flex shrink-0 items-center gap-2.5 font-semibold tracking-tight" href="/">
                        <span className="text-lg max-[390px]:hidden">EduLink</span>
                    </Link>
                    <div className="flex shrink-0 items-center gap-0.5 sm:gap-1.5">
                        <LanguageSwitcher className="border-slate-200 bg-white/70 text-slate-700 shadow-none" compact />
                        <Link className="inline-flex shrink-0 items-center whitespace-nowrap rounded-full px-2.5 py-2 text-xs font-medium text-slate-600 transition-colors hover:bg-white/70 hover:text-slate-950 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 lg:px-3 lg:text-sm" href="/app/portal" prefetch={false}>
                            <span className="lg:hidden">{t("landing.portalShort")}</span>
                            <span className="hidden lg:inline">{t("landing.studentParentLogin")}</span>
                        </Link>
                        <Link className="inline-flex shrink-0 items-center whitespace-nowrap rounded-full px-2.5 py-2 text-xs font-medium text-slate-600 transition-colors hover:bg-white/70 hover:text-slate-950 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 lg:px-3 lg:text-sm" href="/auth/login" prefetch={false}>
                            <span className="lg:hidden">{t("landing.staffShort")}</span>
                            <span className="hidden lg:inline">{t("landing.staffLogin")}</span>
                        </Link>
                        <Link className="inline-flex shrink-0 items-center gap-1.5 whitespace-nowrap rounded-full bg-slate-950 px-3 py-2 text-xs font-semibold text-white shadow-sm transition-colors hover:bg-blue-600 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 sm:px-4 sm:text-sm max-[390px]:size-9 max-[390px]:justify-center max-[390px]:p-0" href="/auth/register" prefetch={false}>
                            <span className="max-[390px]:sr-only">{t("landing.getStarted")}</span>
                            <ArrowRight className="size-3.5" />
                        </Link>
                    </div>
                </nav>
            </header>

            <main>
                <section className="relative isolate flex min-h-[780px] flex-col items-center justify-center overflow-hidden px-5 pb-28 pt-40 sm:px-8 sm:pt-44">
                    <AmbientBackground />
                    <div className="pointer-events-none absolute inset-x-0 top-1/2 -translate-y-1/2 select-none text-center text-[24vw] font-black leading-none tracking-[-0.09em] text-blue-950/[0.025]" aria-hidden="true">
                        EDULINK
                    </div>
                    <FloatingConcepts />
                    <div className="relative z-10">
                    <div className="flex flex-col items-center text-center">
                        <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-blue-200/80 bg-white px-3 py-1.5 text-xs font-semibold text-blue-700 shadow-sm">
                            <Sparkles className="size-3.5" />
                            One school. Finally connected.
                        </div>
                        <h1 className="max-w-5xl text-balance text-5xl font-semibold leading-[0.98] tracking-[-0.055em] text-slate-950 sm:text-7xl lg:text-[6.4rem]">
                            School life, moving
                            <span className="block bg-gradient-to-r from-blue-600 via-violet-600 to-blue-500 bg-clip-text text-transparent">in the same direction.</span>
                        </h1>
                        <p className="mt-7 max-w-2xl text-balance text-base leading-7 text-slate-600 sm:text-lg">
                            EduLink brings students, staff, courses, grades, and families into one beautifully simple place.
                        </p>
                        <div className="mt-9 grid w-full max-w-3xl gap-3 text-left sm:grid-cols-2">
                            <div className="rounded-2xl border border-slate-200 bg-white p-4 shadow-lg shadow-slate-950/5 sm:p-5">
                                <div className="flex items-center gap-2 text-sm font-semibold text-slate-950">
                                    <School className="size-4 text-blue-600" /> Schools &amp; staff
                                </div>
                                <p className="mt-1.5 text-xs leading-5 text-slate-500">Manage your school or sign in to your staff workspace.</p>
                                <div className="mt-4 flex flex-wrap gap-2">
                                    <Link className="inline-flex h-9 items-center justify-center rounded-full border border-slate-200 bg-white px-4 text-xs font-semibold text-slate-700 transition hover:border-slate-300 hover:text-slate-950 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600" href="/auth/login">
                                        Staff login
                                    </Link>
                                    <Link className="inline-flex h-9 items-center justify-center gap-1.5 rounded-full bg-blue-600 px-4 text-xs font-semibold text-white shadow-sm transition-colors hover:bg-blue-700 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600" href="/auth/register">
                                        Get started <ArrowRight className="size-3.5" />
                                    </Link>
                                </div>
                            </div>
                            <div className="rounded-2xl border border-slate-200 bg-white p-4 shadow-lg shadow-slate-950/5 sm:p-5">
                                <div className="flex items-center gap-2 text-sm font-semibold text-slate-950">
                                    <GraduationCap className="size-4 text-violet-600" /> Students &amp; parents
                                </div>
                                <p className="mt-1.5 text-xs leading-5 text-slate-500">Use the portal account provided by your school.</p>
                                <Link className="mt-4 inline-flex h-9 items-center justify-center gap-1.5 rounded-full bg-slate-950 px-4 text-xs font-semibold text-white shadow-sm transition-colors hover:bg-violet-600 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-violet-600" href="/app/portal">
                                    Open portal <ArrowRight className="size-3.5" />
                                </Link>
                            </div>
                        </div>
                        <a className="mt-4 inline-flex h-10 items-center justify-center rounded-full px-5 text-sm font-semibold text-slate-500 transition hover:bg-white/70 hover:text-slate-950 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600" href="#demo">
                            See EduLink
                        </a>
                        <p className="mt-4 flex items-center gap-2 text-xs text-slate-400"><CheckCircle2 className="size-3.5 text-emerald-500" /> Built for the whole school community</p>
                    </div>
                    </div>

                </section>

                <AppDemos />

                <SchoolFlow />

                <section className="mx-auto max-w-7xl px-5 py-28 sm:px-8 sm:py-36" id="platform">
                    <Reveal className="mx-auto max-w-3xl text-center">
                        <p className="text-xs font-semibold uppercase tracking-[0.22em] text-blue-600">The connected school</p>
                        <h2 className="mt-4 text-balance text-4xl font-semibold tracking-[-0.04em] sm:text-6xl">Less chasing. More knowing.</h2>
                        <p className="mx-auto mt-5 max-w-2xl text-balance text-base leading-7 text-slate-600 sm:text-lg">EduLink turns scattered school work into one calm, shared rhythm—from the office to the classroom to home.</p>
                    </Reveal>

                    <div className="mt-16 grid gap-4 md:grid-cols-3">
                        {features.map((feature, index) => (
                            <Reveal className={feature.className} delay={index * 0.06} key={feature.title}>
                                <div className="relative h-full min-h-72 overflow-hidden rounded-3xl border border-slate-200 bg-white p-7 shadow-[0_16px_60px_-35px_rgba(15,23,42,0.3)]">
                                    <div className={`relative z-10 grid size-10 place-items-center rounded-xl text-white shadow-lg ${feature.color}`}>
                                        <feature.icon className="size-5" />
                                    </div>
                                    <div className="relative z-10 mt-16 max-w-lg">
                                        <h3 className="text-2xl font-semibold tracking-[-0.03em] text-slate-950">{feature.title}</h3>
                                        <p className="mt-3 text-sm leading-6 text-slate-600">{feature.description}</p>
                                    </div>
                                    <div className="absolute -right-12 -top-12 size-44 rounded-full bg-gradient-to-br from-blue-100/80 to-violet-100/20" aria-hidden="true" />
                                </div>
                            </Reveal>
                        ))}
                    </div>
                </section>

                <section className="relative bg-slate-950 px-5 py-28 text-white sm:px-8 sm:py-36">
                    <div className="absolute inset-0 opacity-30 [background-image:radial-gradient(circle_at_50%_0%,rgba(59,130,246,.55),transparent_38%)]" aria-hidden="true" />
                    <div className="relative mx-auto max-w-7xl">
                        <Reveal className="max-w-3xl">
                            <p className="text-xs font-semibold uppercase tracking-[0.22em] text-blue-400">Made for everyone</p>
                            <h2 className="mt-4 text-balance text-4xl font-semibold tracking-[-0.04em] sm:text-6xl">Different days. One shared picture.</h2>
                            <p className="mt-5 max-w-2xl text-base leading-7 text-slate-400 sm:text-lg">Every person sees what matters to them, while the school stays connected underneath.</p>
                        </Reveal>

                        <div className="mt-16 grid gap-px overflow-hidden rounded-3xl border border-white/10 bg-white/10 sm:grid-cols-2 lg:grid-cols-4">
                            {roles.map((role, index) => (
                                <Reveal className="bg-slate-950 p-7 sm:p-8" delay={index * 0.07} key={role.label}>
                                    <role.icon className="size-6 text-blue-400" />
                                    <h3 className="mt-12 text-lg font-semibold">{role.label}</h3>
                                    <p className="mt-2 text-sm leading-6 text-slate-400">{role.copy}</p>
                                </Reveal>
                            ))}
                        </div>
                    </div>
                </section>

                <section className="mx-auto max-w-7xl px-5 py-28 sm:px-8 sm:py-36">
                    <div className="mx-auto max-w-5xl">
                        <Reveal className="text-center">
                            <p className="text-xs font-semibold uppercase tracking-[0.22em] text-blue-600">Quietly powerful</p>
                            <h2 className="mx-auto mt-4 max-w-3xl text-balance text-4xl font-semibold tracking-[-0.04em] sm:text-5xl">The complexity stays behind the scenes.</h2>
                            <p className="mx-auto mt-5 max-w-2xl text-base leading-7 text-slate-600">EduLink gives people a simple experience without giving up the control, visibility, and accountability schools need.</p>
                            <div className="mt-12 grid gap-4 text-left md:grid-cols-3">
                                <TrustItem icon={ShieldCheck} title="Permissions that make sense" copy="Give people access based on their role and responsibilities." />
                                <TrustItem icon={LockKeyhole} title="Privacy by default" copy="Personal choices and school access stay clear and intentional." />
                                <TrustItem icon={BellRing} title="Updates that reach people" copy="Keep the community informed without turning everything into noise." />
                            </div>
                        </Reveal>
                    </div>
                </section>

                <section className="px-5 pb-20 sm:px-8 sm:pb-28">
                    <Reveal className="relative mx-auto max-w-7xl overflow-hidden rounded-[2rem] bg-blue-600 px-6 py-20 text-center text-white shadow-[0_30px_100px_-30px_rgba(37,99,235,.65)] sm:px-12 sm:py-28">
                        <div className="absolute inset-0 [background-image:radial-gradient(circle_at_20%_10%,rgba(255,255,255,.28),transparent_25%),radial-gradient(circle_at_85%_90%,rgba(139,92,246,.65),transparent_30%)]" aria-hidden="true" />
                        <div className="relative mx-auto max-w-3xl">
                            <div className="mx-auto grid size-12 place-items-center rounded-2xl bg-white/15"><Layers3 className="size-6" /></div>
                            <h2 className="mt-7 text-balance text-4xl font-semibold tracking-[-0.045em] sm:text-6xl">Bring your whole school into the loop.</h2>
                            <p className="mx-auto mt-5 max-w-xl text-balance text-base leading-7 text-blue-100 sm:text-lg">Start building a calmer, clearer school experience with EduLink.</p>
                            <div className="mt-9 flex flex-col items-center justify-center gap-3 sm:flex-row">
                                <Link className="inline-flex h-12 items-center justify-center gap-2 rounded-full bg-white px-6 text-sm font-semibold text-blue-700 shadow-xl transition-colors hover:bg-blue-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-white" href="/auth/register">
                                    Get started as staff <ArrowRight className="size-4" />
                                </Link>
                                <Link className="inline-flex h-12 items-center justify-center rounded-full border border-white/30 bg-white/10 px-6 text-sm font-semibold text-white transition-colors hover:bg-white/20 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-white" href="/app/portal">
                                    Student &amp; parent portal
                                </Link>
                            </div>
                        </div>
                    </Reveal>
                </section>
            </main>

            <footer className="border-t border-slate-200 bg-white px-5 py-8 sm:px-8">
                <div className="mx-auto flex max-w-7xl flex-col items-center justify-between gap-5 sm:flex-row">
                    <Link className="text-sm font-semibold" href="/">EduLink</Link>
                    <p className="text-center text-xs text-slate-400">{t("landing.footerTagline")}</p>
                    <div className="flex gap-5 text-xs font-medium text-slate-500">
                        <Link className="hover:text-slate-950" href="/legal/privacy">{t("common.privacy")}</Link>
                        <Link className="hover:text-slate-950" href="/legal/terms">{t("common.terms")}</Link>
                        <Link className="hover:text-slate-950" href="/auth/login">{t("landing.staffLogin")}</Link>
                        <Link className="hover:text-slate-950" href="/app/portal">{t("landing.studentParentLogin")}</Link>
                    </div>
                </div>
            </footer>
        </div>
    )
}

function AmbientBackground() {
    return (
        <div className="pointer-events-none absolute inset-0" aria-hidden="true">
            <div className="absolute inset-0 [background-image:linear-gradient(to_right,rgba(148,163,184,.12)_1px,transparent_1px),linear-gradient(to_bottom,rgba(148,163,184,.12)_1px,transparent_1px)] [background-size:72px_72px] [mask-image:linear-gradient(to_bottom,black,transparent_80%)]" />
            <div className="absolute inset-0 bg-[radial-gradient(circle_at_18%_22%,rgba(147,197,253,0.28),transparent_28%),radial-gradient(circle_at_82%_35%,rgba(196,181,253,0.22),transparent_26%)]" />
            <div className="absolute inset-x-0 top-0 h-48 bg-gradient-to-b from-white to-transparent" />
        </div>
    )
}

function FloatingConcepts() {
    const concepts = [
        {label: "COURSES", className: "left-[7%] top-[31%]"},
        {label: "STUDENTS", className: "right-[6%] top-[27%]"},
        {label: "GRADES", className: "bottom-[22%] left-[12%]"},
        {label: "STAFF", className: "bottom-[26%] right-[10%]"},
    ]

    return (
        <div className="pointer-events-none absolute inset-0 hidden md:block" aria-hidden="true">
            {concepts.map((concept) => (
                <span
                    className={`absolute rounded-full border border-slate-200 bg-white px-3 py-1.5 text-[10px] font-semibold tracking-[0.2em] text-slate-400 shadow-sm ${concept.className}`}
                    key={concept.label}
                >
                    {concept.label}
                </span>
            ))}
        </div>
    )
}

function SchoolFlow() {
    const steps = [
        {icon: BookOpenCheck, title: "Staff share", copy: "Course updates and resources start in one clear place."},
        {icon: GraduationCap, title: "Students know", copy: "The right information reaches the people who need it."},
        {icon: BarChart3, title: "Progress stays visible", copy: "Grades and activity keep the bigger picture current."},
        {icon: HeartHandshake, title: "Everyone stays aligned", copy: "Leaders and families can follow along without chasing updates."},
    ]

    return (
        <section className="border-y border-slate-200 bg-slate-50 px-5 py-24 sm:px-8 sm:py-32">
            <div className="mx-auto max-w-7xl">
                <Reveal className="mx-auto max-w-3xl text-center">
                    <p className="text-xs font-semibold uppercase tracking-[0.22em] text-blue-600">One connected flow</p>
                    <h2 className="mt-4 text-balance text-4xl font-semibold tracking-[-0.04em] sm:text-6xl">An update shouldn&apos;t get lost on the way.</h2>
                    <p className="mx-auto mt-5 max-w-2xl text-balance text-base leading-7 text-slate-600 sm:text-lg">EduLink keeps information moving through the school without turning communication into another job.</p>
                </Reveal>

                <div className="relative mt-16 grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                    <div className="absolute left-[12.5%] right-[12.5%] top-7 hidden h-px bg-gradient-to-r from-blue-200 via-violet-300 to-emerald-200 lg:block" aria-hidden="true" />
                    {steps.map((step, index) => (
                        <Reveal className="relative" delay={index * 0.08} key={step.title}>
                            <div className="relative z-10 grid size-14 place-items-center rounded-2xl border border-slate-200 bg-white text-blue-600 shadow-sm"><step.icon className="size-5" /></div>
                            <div className="mt-7 rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
                                <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-slate-400">Step {index + 1}</p>
                                <h3 className="mt-2 text-lg font-semibold tracking-tight">{step.title}</h3>
                                <p className="mt-2 text-sm leading-6 text-slate-500">{step.copy}</p>
                            </div>
                        </Reveal>
                    ))}
                </div>
            </div>
        </section>
    )
}

function Reveal({children, className = "", delay = 0}: {children: React.ReactNode; className?: string; delay?: number}) {
    const reduceMotion = useReducedMotion()
    return (
        <motion.div
            className={className}
            initial={reduceMotion ? false : {opacity: 0, y: 32}}
            whileInView={{opacity: 1, y: 0}}
            viewport={{once: true, amount: 0.18}}
            transition={{duration: 0.7, delay, ease: [0.22, 1, 0.36, 1]}}
        >
            {children}
        </motion.div>
    )
}

function TrustItem({icon: Icon, title, copy}: {icon: typeof ShieldCheck; title: string; copy: string}) {
    return (
        <div className="flex gap-4 rounded-2xl border border-slate-200 bg-white p-4">
            <div className="grid size-10 shrink-0 place-items-center rounded-xl bg-blue-50 text-blue-600"><Icon className="size-5" /></div>
            <div><h3 className="text-sm font-semibold">{title}</h3><p className="mt-1 text-sm leading-5 text-slate-500">{copy}</p></div>
        </div>
    )
}

"use client"

import {useState} from "react"
import {AnimatePresence, motion, useReducedMotion} from "framer-motion"
import {
    ArrowLeft,
    BookOpen,
    CalendarClock,
    GraduationCap,
    LayoutDashboard,
    MapPin,
    Menu,
    Pencil,
    Plus,
    ScrollText,
    Search,
    Settings,
    ShieldCheck,
    Trash2,
    Users,
} from "lucide-react"
import {Badge} from "@/components/ui/badge"
import {Button} from "@/components/ui/button"
import {Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle} from "@/components/ui/card"
import {Input} from "@/components/ui/input"

type DemoPage = "school" | "courses" | "posts"

const demoTabs: Array<{id: DemoPage; label: string}> = [
    {id: "school", label: "School dashboard"},
    {id: "courses", label: "Grade courses"},
    {id: "posts", label: "Course posts"},
]

export default function AppDemos() {
    const [activePage, setActivePage] = useState<DemoPage>("school")
    const reduceMotion = useReducedMotion()

    return (
        <section className="bg-white px-5 py-28 sm:px-8 sm:py-36" id="demo">
            <div className="mx-auto max-w-7xl">
                <div className="mx-auto max-w-3xl text-center">
                    <p className="text-xs font-semibold uppercase tracking-[0.22em] text-blue-600">Inside EduLink</p>
                    <h2 className="mt-4 text-balance text-4xl font-semibold tracking-[-0.04em] text-slate-950 sm:text-6xl">Familiar from the first click.</h2>
                    <p className="mx-auto mt-5 max-w-2xl text-balance text-base leading-7 text-slate-600 sm:text-lg">The same calm, consistent workspace follows you from the whole school down to a single course.</p>
                </div>

                <div className="mx-auto mt-10 flex w-fit max-w-full gap-1 overflow-x-auto rounded-xl border border-slate-200 bg-slate-100 p-1" role="tablist" aria-label="Sample EduLink pages">
                    {demoTabs.map((tab) => (
                        <button
                            className={`shrink-0 rounded-lg px-3.5 py-2 text-xs font-semibold transition sm:text-sm ${activePage === tab.id ? "bg-white text-slate-950 shadow-sm" : "text-slate-500 hover:text-slate-800"}`}
                            type="button"
                            role="tab"
                            aria-selected={activePage === tab.id}
                            onClick={() => setActivePage(tab.id)}
                            key={tab.id}
                        >
                            {tab.label}
                        </button>
                    ))}
                </div>

                <div className="relative mt-8 overflow-hidden rounded-[1.75rem] border border-slate-200 bg-white p-2 shadow-[0_35px_100px_-45px_rgba(15,23,42,.35)] sm:p-3">
                    <div className="flex h-10 items-center justify-between rounded-t-[1.2rem] border border-b-0 border-slate-200 bg-slate-50 px-4" aria-hidden="true">
                        <div className="flex gap-1.5"><span className="size-2 rounded-full bg-slate-300" /><span className="size-2 rounded-full bg-slate-300" /><span className="size-2 rounded-full bg-slate-300" /></div>
                        <p className="text-[10px] font-medium text-slate-400">Sample workspace</p>
                        <div className="size-5" />
                    </div>

                    <div className="grid h-[640px] overflow-hidden rounded-b-[1.2rem] border border-slate-200 bg-gradient-to-br from-white via-white to-slate-50 text-slate-950 md:grid-cols-[10.5rem_minmax(0,1fr)]">
                        <DemoSidebar activePage={activePage} />
                        <div className="min-w-0 overflow-y-auto p-4 sm:p-6" inert>
                            <MobileDemoNavigation activePage={activePage} />
                            <AnimatePresence mode="wait" initial={false}>
                                <motion.div
                                    key={activePage}
                                    initial={reduceMotion ? false : {opacity: 0, y: 18}}
                                    animate={{opacity: 1, y: 0}}
                                    exit={reduceMotion ? undefined : {opacity: 0, y: -10}}
                                    transition={{duration: 0.28, ease: [0.22, 1, 0.36, 1]}}
                                >
                                    {activePage === "school" && <SchoolDashboardDemo />}
                                    {activePage === "courses" && <CoursesDemo />}
                                    {activePage === "posts" && <PostsDemo />}
                                </motion.div>
                            </AnimatePresence>
                        </div>
                    </div>
                </div>
                <p className="mt-4 text-center text-xs text-slate-400">Interactive sample using the same interface patterns as the app.</p>
            </div>
        </section>
    )
}

function DemoSidebar({activePage}: {activePage: DemoPage}) {
    const navigation = [
        {label: "Dashboard", icon: LayoutDashboard, active: activePage === "school"},
        {label: "Staff", icon: Users},
        {label: "Roles", icon: ShieldCheck},
        {label: "Logs", icon: ScrollText},
        {label: "Settings", icon: Settings},
    ]

    return (
        <aside className="hidden border-r border-slate-200 bg-white px-3 py-5 md:flex md:flex-col" aria-hidden="true">
            <p className="mb-3 px-2 text-[10px] font-semibold uppercase tracking-[0.14em] text-slate-400">School menu</p>
            <div className="space-y-0.5">
                {navigation.map(({label, icon: Icon, active}) => (
                    <div className={`flex items-center gap-2 rounded-md px-2.5 py-2 text-xs font-medium ${active ? "bg-slate-100 text-slate-950" : "text-slate-500"}`} key={label}>
                        <Icon className="size-3.5" /> {label}
                    </div>
                ))}
            </div>
            <div className="mt-auto flex items-center gap-2 px-2.5 py-2 text-xs font-medium text-slate-500"><ArrowLeft className="size-3.5" /> All schools</div>
        </aside>
    )
}

function MobileDemoNavigation({activePage}: {activePage: DemoPage}) {
    const label = activePage === "school" ? "Dashboard" : activePage === "courses" ? "Courses" : "Course dashboard"
    return (
        <div className="mb-5 flex items-center justify-between border-b border-slate-200 pb-3 md:hidden" aria-hidden="true">
            <div><p className="text-[9px] font-semibold uppercase tracking-[0.14em] text-slate-400">School</p><p className="text-xs font-semibold">{label}</p></div>
            <div className="flex items-center gap-1.5 rounded-md border border-slate-200 px-2.5 py-1.5 text-[10px] font-semibold"><Menu className="size-3" /> Menu</div>
        </div>
    )
}

function SchoolDashboardDemo() {
    const grades = [
        {name: "Grade 7", level: 7},
        {name: "Grade 8", level: 8},
        {name: "Grade 9", level: 9},
    ]

    return (
        <div className="space-y-5">
            <div><h3 className="text-2xl font-semibold tracking-tight">Northstar Academy</h3><p className="text-xs text-muted-foreground">United States</p></div>
            <div className="grid gap-3 sm:grid-cols-3">
                <Card><CardHeader className="gap-1"><CardDescription className="flex items-center gap-2 text-xs"><GraduationCap /> Active grades</CardDescription><CardTitle className="text-2xl">3</CardTitle></CardHeader></Card>
                <Card><CardHeader className="gap-1"><CardDescription className="flex items-center gap-2 text-xs"><MapPin /> Region</CardDescription><CardTitle className="text-base">United States</CardTitle></CardHeader></Card>
                <Card><CardHeader className="gap-1"><CardDescription className="flex items-center gap-2 text-xs"><LayoutDashboard /> Created</CardDescription><CardTitle className="text-base">16 Oct 2025</CardTitle></CardHeader></Card>
            </div>
            <section>
                <div className="mb-3 flex flex-wrap items-end justify-between gap-3">
                    <div><h4 className="text-xl font-semibold">Active grades</h4><p className="text-xs text-muted-foreground">Grades in the school&apos;s current academic year</p></div>
                    <Button size="sm"><Plus /> Create grade</Button>
                </div>
                <div className="grid gap-3 sm:grid-cols-3">
                    {grades.map((grade) => (
                        <Card key={grade.level}>
                            <CardHeader><div className="flex items-start justify-between gap-2"><CardTitle className="text-sm">{grade.name}</CardTitle><Badge variant="outline">Level {grade.level}</Badge></div></CardHeader>
                            <CardFooter className="justify-end gap-2"><Button size="sm" variant="outline"><Pencil /> Edit</Button><Button size="icon-sm" variant="destructive-outline"><Trash2 /></Button></CardFooter>
                        </Card>
                    ))}
                </div>
            </section>
        </div>
    )
}

function CoursesDemo() {
    const courses = [
        {name: "Mathematics", color: "#6366F1", description: "Core concepts, problem solving, and weekly practice."},
        {name: "Biology", color: "#10B981", description: "Living systems, lab work, and scientific investigation."},
        {name: "World History", color: "#F59E0B", description: "People, events, and ideas that shaped the modern world."},
    ]

    return (
        <div className="space-y-5">
            <div className="flex flex-wrap items-end justify-between gap-3">
                <div><Button className="-ml-2 mb-1" size="sm" variant="ghost"><ArrowLeft /> Back to grades</Button><h3 className="text-2xl font-semibold tracking-tight">Courses</h3><p className="text-xs text-muted-foreground">Courses assigned to this grade</p></div>
                <Button size="sm"><Plus /> Create course</Button>
            </div>
            <div className="relative max-w-sm"><Search className="pointer-events-none absolute left-3 top-1/2 z-10 size-3.5 -translate-y-1/2 text-muted-foreground" /><Input className="pl-9" placeholder="Search courses..." readOnly /></div>
            <div className="grid gap-3 sm:grid-cols-3">
                {courses.map((course) => (
                    <Card className="overflow-hidden" key={course.name}>
                        <div className="h-1.5" style={{backgroundColor: course.color}} />
                        <CardHeader><div className="flex items-start gap-3"><span className="flex size-9 shrink-0 items-center justify-center rounded-lg text-white shadow-sm" style={{backgroundColor: course.color}}><BookOpen className="size-4" /></span><CardTitle className="truncate text-sm">{course.name}</CardTitle></div></CardHeader>
                        <CardContent className="flex-1"><p className="line-clamp-3 text-xs leading-relaxed text-muted-foreground">{course.description}</p></CardContent>
                        <CardFooter className="justify-end gap-2 border-t bg-muted/25"><Button size="sm" variant="outline"><Pencil /> Edit</Button><Button size="icon-sm" variant="destructive-outline"><Trash2 /></Button></CardFooter>
                    </Card>
                ))}
            </div>
        </div>
    )
}

function PostsDemo() {
    const posts = [
        {title: "Project brief is live", author: "Maya Thompson", date: "16 Oct 2025, 09:00", body: "Bring your first sketches to Thursday’s workshop. The project brief and reference material are now available.", color: "#6366F1", expiry: "23 Oct 2025, 16:00"},
        {title: "Reading notes", author: "Daniel Brooks", date: "15 Oct 2025, 14:30", body: "I’ve added the reading notes and follow-up questions from today’s discussion.", color: "#10B981"},
    ]

    return (
        <div className="space-y-6">
            <header><Button className="-ml-2 mb-2" size="sm" variant="ghost"><ArrowLeft /> Back to courses</Button><div className="flex items-start gap-3"><span className="flex size-10 items-center justify-center rounded-xl bg-primary text-primary-foreground"><BookOpen className="size-4" /></span><div><h3 className="text-2xl font-semibold tracking-tight">Course dashboard</h3><p className="text-xs text-muted-foreground">Updates and resources for this course</p></div></div></header>
            <section className="space-y-3">
                <div className="flex items-end justify-between gap-3"><div><h4 className="text-xl font-semibold">Posts</h4><p className="text-xs text-muted-foreground">Updates and announcements for this course</p></div><Button size="sm"><Plus /> Create post</Button></div>
                {posts.map((post) => (
                    <Card className="overflow-hidden" key={post.title}>
                        <div className="absolute inset-y-0 left-0 w-1" style={{backgroundColor: post.color}} />
                        <CardHeader className="gap-2 pl-7"><div className="flex items-start justify-between gap-3"><CardTitle className="text-sm">{post.title}</CardTitle><div className="flex gap-1"><Button size="icon-sm" variant="ghost"><Pencil /></Button><Button size="icon-sm" variant="destructive-outline"><Trash2 /></Button></div></div><p className="text-[10px] text-muted-foreground">{post.author} · Posted {post.date}</p></CardHeader>
                        <CardContent className="pl-7"><p className="text-xs leading-relaxed">{post.body}</p></CardContent>
                        {post.expiry && <CardFooter className="gap-2 border-t bg-muted/25 py-3 pl-7 text-[10px] text-muted-foreground"><CalendarClock className="size-3" /> Shows until {post.expiry}</CardFooter>}
                    </Card>
                ))}
            </section>
        </div>
    )
}

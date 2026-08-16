"use client"

import {useMemo, useState} from "react"
import Link from "next/link"
import {useRouter} from "next/navigation"
import {ArrowUpRight, BookOpen, Layers3, RotateCcw, Search, Sparkles, X} from "lucide-react"
import {Badge} from "@/components/ui/badge"
import {Button} from "@/components/ui/button"
import {Card, CardContent, CardHeader, CardTitle} from "@/components/ui/card"
import {Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle} from "@/components/ui/empty"
import {Input} from "@/components/ui/input"
import {useLocale} from "@/i18n/provider"
import {pluralCategory} from "@/i18n/config"

export interface PortalCourse {
    id: string
    grade: {
        id: string
        name: string
    }
    name: string
    description: string
    accentColor: string
}

function courseColor(value: string) {
    return /^[0-9A-Fa-f]{6}$/.test(value) ? `#${value}` : "#6366F1"
}

export default function StudentPortalPage({initialCourses, error}: {
    initialCourses: PortalCourse[]
    error?: string
}) {
    const router = useRouter()
    const {locale, t} = useLocale()
    const [query, setQuery] = useState("")
    const normalizedQuery = query.trim().toLocaleLowerCase()
    const gradeCount = new Set(initialCourses.map((course) => course.grade.id)).size
    const groupedCourses = useMemo(() => {
        const visible = initialCourses.filter((course) => !normalizedQuery || [course.name, course.description, course.grade.name]
            .some((value) => value.toLocaleLowerCase().includes(normalizedQuery)))
        const groups = new Map<string, {grade: PortalCourse["grade"]; courses: PortalCourse[]}>()

        for (const course of visible) {
            const existing = groups.get(course.grade.id)
            if (existing) existing.courses.push(course)
            else groups.set(course.grade.id, {grade: course.grade, courses: [course]})
        }

        return [...groups.values()]
            .map((group) => ({...group, courses: group.courses.sort((first, second) => first.name.localeCompare(second.name, locale))}))
            .sort((first, second) => first.grade.name.localeCompare(second.grade.name, locale, {numeric: true}))
    }, [initialCourses, locale, normalizedQuery])
    const visibleCount = groupedCourses.reduce((total, group) => total + group.courses.length, 0)
    const countKey = (kind: "matching" | "assigned" | "course", count: number) => {
        const category = pluralCategory(locale, count)
        const suffix = category === "one" ? "One" : category === "few" ? "Few" : "Other"
        return `portal.student.${kind}${suffix}` as
            | "portal.student.matchingOne" | "portal.student.matchingFew" | "portal.student.matchingOther"
            | "portal.student.assignedOne" | "portal.student.assignedFew" | "portal.student.assignedOther"
            | "portal.student.courseOne" | "portal.student.courseFew" | "portal.student.courseOther"
    }

    return (
        <div className="space-y-7">
            <header className="relative overflow-hidden rounded-3xl border bg-card px-5 py-6 shadow-xs sm:px-7 sm:py-8">
                <div className="pointer-events-none absolute -right-16 -top-24 size-64 rounded-full bg-primary/8 blur-3xl" />
                <div className="relative flex flex-col justify-between gap-6 sm:flex-row sm:items-end">
                    <div>
                        <Badge className="mb-3" variant="secondary"><Sparkles /> {t("portal.student.badge")}</Badge>
                        <h1 className="text-3xl font-semibold tracking-tight sm:text-4xl">{t("portal.student.title")}</h1>
                        <p className="mt-2 max-w-xl text-muted-foreground">{t("portal.student.description")}</p>
                    </div>
                    {!error && initialCourses.length > 0 && (
                        <div className="flex gap-3">
                            <Metric icon={BookOpen} label={t("portal.courses")} value={initialCourses.length} />
                            <Metric icon={Layers3} label={t("portal.student.grades")} value={gradeCount} />
                        </div>
                    )}
                </div>
            </header>

            {error ? (
                <Card>
                    <Empty>
                        <EmptyHeader>
                            <EmptyMedia variant="icon"><BookOpen /></EmptyMedia>
                            <EmptyTitle>{t("portal.student.loadErrorTitle")}</EmptyTitle>
                            <EmptyDescription>{error}</EmptyDescription>
                        </EmptyHeader>
                        <Button variant="outline" onClick={() => router.refresh()}><RotateCcw /> {t("portal.student.tryAgain")}</Button>
                    </Empty>
                </Card>
            ) : initialCourses.length === 0 ? (
                <Card>
                    <Empty>
                        <EmptyHeader>
                            <EmptyMedia variant="icon"><BookOpen /></EmptyMedia>
                            <EmptyTitle>{t("portal.student.emptyTitle")}</EmptyTitle>
                            <EmptyDescription>{t("portal.student.emptyDescription")}</EmptyDescription>
                        </EmptyHeader>
                    </Empty>
                </Card>
            ) : (
                <section className="space-y-6" aria-labelledby="course-list-title">
                    <div className="flex flex-col justify-between gap-3 sm:flex-row sm:items-end">
                        <div>
                            <h2 className="text-lg font-semibold" id="course-list-title">{t("portal.student.listTitle")}</h2>
                            <p className="text-sm text-muted-foreground">
                                {normalizedQuery
                                    ? t(countKey("matching", visibleCount), {count: visibleCount})
                                    : t(countKey("assigned", initialCourses.length), {count: initialCourses.length})}
                            </p>
                        </div>
                        <div className="relative w-full sm:max-w-sm">
                            <Search className="pointer-events-none absolute left-3 top-1/2 z-10 size-4 -translate-y-1/2 text-muted-foreground" />
                            <Input className="pl-9 pr-9" value={query} onChange={(event) => setQuery(event.target.value)} placeholder={t("portal.student.searchPlaceholder")} aria-label={t("portal.student.searchLabel")} />
                            {query && <Button className="absolute right-1.5 top-1/2 -translate-y-1/2" size="icon-xs" variant="ghost" aria-label={t("portal.student.clearSearch")} onClick={() => setQuery("")}><X /></Button>}
                        </div>
                    </div>

                    {groupedCourses.length === 0 ? (
                        <Card className="border-dashed">
                            <Empty>
                                <EmptyHeader>
                                    <EmptyMedia variant="icon"><Search /></EmptyMedia>
                                    <EmptyTitle>{t("portal.student.noMatches")}</EmptyTitle>
                                    <EmptyDescription>{t("portal.student.noMatchesDescription")}</EmptyDescription>
                                </EmptyHeader>
                                <Button variant="outline" onClick={() => setQuery("")}>{t("portal.student.clearSearch")}</Button>
                            </Empty>
                        </Card>
                    ) : groupedCourses.map(({grade, courses}) => (
                        <div className="space-y-3" key={grade.id}>
                            <div className="flex items-center gap-3">
                                <h3 className="font-semibold">{grade.name}</h3>
                                <span className="h-px flex-1 bg-border" />
                                <span className="text-xs tabular-nums text-muted-foreground">
                                    {t(countKey("course", courses.length), {count: courses.length})}
                                </span>
                            </div>
                            <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
                                {courses.map((course) => <CourseCard course={course} noDescription={t("portal.student.noDescription")} key={course.id} />)}
                            </div>
                        </div>
                    ))}
                </section>
            )}
        </div>
    )
}

function Metric({icon: Icon, label, value}: {icon: typeof BookOpen; label: string; value: number}) {
    return (
        <div className="min-w-24 rounded-2xl border bg-background/75 px-4 py-3 backdrop-blur">
            <div className="flex items-center gap-1.5 text-xs text-muted-foreground"><Icon className="size-3.5" /> {label}</div>
            <p className="mt-1 text-2xl font-semibold tabular-nums">{value}</p>
        </div>
    )
}

function CourseCard({course, noDescription}: {course: PortalCourse; noDescription: string}) {
    const color = courseColor(course.accentColor)
    return (
        <Link className="group block rounded-xl outline-none focus-visible:ring-2 focus-visible:ring-ring" href={`/app/portal/student/courses/${course.id}`}>
            <Card className="relative min-h-44 overflow-hidden transition-[transform,box-shadow,border-color] duration-200 group-hover:-translate-y-0.5 group-hover:border-foreground/20 group-hover:shadow-md">
                <div className="absolute inset-y-0 left-0 w-1.5" style={{backgroundColor: color}} />
                <CardHeader className="gap-3 pl-7">
                    <div className="flex items-start justify-between gap-3">
                        <span className="flex size-9 items-center justify-center rounded-xl" style={{backgroundColor: `color-mix(in srgb, ${color} 14%, transparent)`, color}}><BookOpen className="size-4" /></span>
                        <ArrowUpRight className="size-4 text-muted-foreground transition-transform group-hover:-translate-y-0.5 group-hover:translate-x-0.5" />
                    </div>
                    <div>
                        <CardTitle className="text-lg leading-snug">{course.name}</CardTitle>
                        <p className="mt-1 text-xs font-medium text-muted-foreground">{course.grade.name}</p>
                    </div>
                </CardHeader>
                <CardContent className="pl-7">
                    <p className="line-clamp-2 text-sm leading-relaxed text-muted-foreground">{course.description || noDescription}</p>
                </CardContent>
            </Card>
        </Link>
    )
}

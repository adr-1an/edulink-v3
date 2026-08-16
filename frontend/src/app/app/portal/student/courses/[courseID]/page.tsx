import {cookies} from "next/headers"
import Link from "next/link"
import {redirect} from "next/navigation"
import {
    ArrowLeft, BookOpen, CalendarDays, CircleX, ClipboardList,
    ExternalLink, FileText, GlobeX, Layers3, Mail, Pencil, ServerCrash,
} from "lucide-react"
import ErrorPage from "@/components/app/error"
import PostAttachments from "@/components/app/post-attachments"
import UserAvatar from "@/components/app/user_avatar"
import LocalDateTime from "@/components/local-date-time"
import {Badge} from "@/components/ui/badge"
import {Button} from "@/components/ui/button"
import {Card, CardContent, CardHeader, CardTitle} from "@/components/ui/card"
import {Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle} from "@/components/ui/empty"
import {Popover, PopoverDescription, PopoverPopup, PopoverTitle, PopoverTrigger} from "@/components/ui/popover"
import {Tabs, TabsList, TabsPanel, TabsTab} from "@/components/ui/tabs"
import {Tooltip, TooltipPopup, TooltipTrigger} from "@/components/ui/tooltip"
import {normalizePostAttachments, type PostAttachment} from "@/lib/post_attachments"
import {normalizeProfilePictureURL} from "@/lib/profile_picture"
import AssignmentsSection, {type PortalCourseAssignment} from "./assignments_section"
import {getTranslations} from "@/i18n/server"

export async function generateMetadata() {
    const {t} = await getTranslations()
    return {title: t("course.metaTitle")}
}

interface PortalCourseDashboard {
    id: string
    name: string
    description: string
    accentColor: string
    grade: {
        id: string
        name: string
        level: number
    }
}

interface PortalCoursePost {
    id: string
    attachments: PostAttachment[]
    author: {
        id: string
        name: string
        email: string
        profilePictureURL: string | null
    }
    title: string
    body: string
    accentColor: string
    createdAt: string
    editedAt: string | null
}

interface RawPortalCoursePost extends Omit<PortalCoursePost, "attachments" | "author"> {
    attachments?: unknown
    author: Omit<PortalCoursePost["author"], "profilePictureURL"> & {
        profilePictureURL?: unknown
    }
}

interface DashboardResponse {
    course?: PortalCourseDashboard
    posts?: RawPortalCoursePost[] | null
    assignments?: PortalCourseAssignment[] | null
}

function color(value: string) {
    return /^[0-9A-Fa-f]{6}$/.test(value) ? `#${value}` : "#6366F1"
}

export default async function Page({params}: {params: Promise<{courseID: string}>}) {
    const {t} = await getTranslations()
    const {courseID} = await params
    if (!/^\d+$/.test(courseID)) return <ErrorPage message={t("course.error.invalidId")} icon={CircleX} />

    const token = (await cookies()).get("portal_token")?.value
    let res: Response
    try {
        res = await fetch(`${process.env.API_URL}/v1/portal/courses/${courseID}`, {
            headers: {Authorization: `Bearer ${token}`},
            cache: "no-store",
        })
    } catch {
        return <ErrorPage message={t("course.error.network")} icon={GlobeX} />
    }

    if (res.status === 401) redirect("/app/portal")
    if (res.status === 403) return <ErrorPage message={t("course.error.forbidden")} icon={CircleX} />
    if (!res.ok) return <ErrorPage message={res.status === 500 ? t("course.error.server") : t("course.error.load")} icon={res.status === 500 ? ServerCrash : CircleX} />

    let data: DashboardResponse
    try {
        data = await res.json() as DashboardResponse
    } catch {
        return <ErrorPage message={t("course.error.invalidResponse")} icon={CircleX} />
    }

    if (!data.course) return <ErrorPage message={t("course.error.incomplete")} icon={CircleX} />

    const course = data.course
    const posts: PortalCoursePost[] = (Array.isArray(data.posts) ? data.posts : [])
        .map((post) => ({
            ...post,
            attachments: normalizePostAttachments(post.attachments),
            author: {
                ...post.author,
                profilePictureURL: normalizeProfilePictureURL(post.author.profilePictureURL),
            },
        }))
        .sort((first, second) => new Date(second.createdAt).getTime() - new Date(first.createdAt).getTime())
    const assignments = (Array.isArray(data.assignments) ? data.assignments : [])
        .sort((first, second) => {
            if (first.dueDate && second.dueDate) {
                return new Date(first.dueDate).getTime() - new Date(second.dueDate).getTime()
            }
            if (first.dueDate) return -1
            if (second.dueDate) return 1
            return new Date(second.createdAt).getTime() - new Date(first.createdAt).getTime()
        })
    const courseColor = color(course.accentColor)
    const referenceTime = new Date().toISOString()
    const referenceDate = referenceTime.slice(0, 10)

    return (
        <div className="space-y-6">
            <header className="relative overflow-hidden rounded-3xl border bg-card p-5 shadow-xs sm:p-7">
                <div className="pointer-events-none absolute inset-y-0 left-0 w-1.5" style={{backgroundColor: courseColor}} />
                <div className="pointer-events-none absolute -right-16 -top-20 size-64 rounded-full blur-3xl" style={{backgroundColor: `color-mix(in srgb, ${courseColor} 10%, transparent)`}} />
                <div className="relative pl-1">
                    <Button className="-ml-2 mb-5" size="sm" variant="ghost" render={<Link href="/app/portal/student" />}><ArrowLeft /> {t("course.back")}</Button>
                    <div className="flex flex-col justify-between gap-5 sm:flex-row sm:items-start">
                        <div className="flex min-w-0 items-start gap-3.5">
                            <span className="mt-0.5 flex size-12 shrink-0 items-center justify-center rounded-2xl text-white shadow-sm" style={{backgroundColor: courseColor}}><BookOpen className="size-5" /></span>
                            <div className="min-w-0">
                                <div className="mb-2 flex flex-wrap items-center gap-2">
                                    <Badge variant="secondary"><Layers3 /> {course.grade.name}</Badge>
                                    <Badge variant="outline">{t("course.level", {level: course.grade.level})}</Badge>
                                </div>
                                <h1 className="wrap-break-word text-3xl font-semibold tracking-tight sm:text-4xl">{course.name}</h1>
                                {course.description && <p className="mt-2 max-w-2xl text-sm leading-relaxed text-muted-foreground sm:text-base">{course.description}</p>}
                            </div>
                        </div>
                        <div className="flex shrink-0 divide-x rounded-2xl border bg-background/75 backdrop-blur">
                            <div className="flex items-center gap-2 px-4 py-3">
                                <FileText className="size-4 text-muted-foreground" />
                                <div>
                                    <p className="text-lg font-semibold leading-none tabular-nums">{posts.length}</p>
                                    <p className="mt-1 text-xs text-muted-foreground">{t(posts.length === 1 ? "course.postOne" : "course.postOther")}</p>
                                </div>
                            </div>
                            <div className="flex items-center gap-2 px-4 py-3">
                                <ClipboardList className="size-4 text-muted-foreground" />
                                <div>
                                    <p className="text-lg font-semibold leading-none tabular-nums">{assignments.length}</p>
                                    <p className="mt-1 text-xs text-muted-foreground">{t(assignments.length === 1 ? "course.assignmentOne" : "course.assignmentOther")}</p>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </header>

            <Tabs defaultValue="posts">
                <TabsList aria-label={t("course.sections")}>
                    <TabsTab value="posts"><FileText /> {t("course.postOther")} <Badge variant="secondary">{posts.length}</Badge></TabsTab>
                    <TabsTab value="assignments"><ClipboardList /> {t("assignments.title")} <Badge variant="secondary">{assignments.length}</Badge></TabsTab>
                </TabsList>

                <TabsPanel className="pt-4" value="posts">
            <div className="grid items-start gap-6 xl:grid-cols-[minmax(0,2fr)_minmax(17rem,0.8fr)]">
                <section className="min-w-0 space-y-4" aria-labelledby="course-posts-title">
                    <div>
                        <h2 className="text-xl font-semibold" id="course-posts-title">{t("course.postOther")}</h2>
                        <p className="text-sm text-muted-foreground">{t("course.postsDescription")}</p>
                    </div>

                    {posts.length === 0 ? (
                        <Card>
                            <Empty>
                                <EmptyHeader>
                                    <EmptyMedia variant="icon"><FileText /></EmptyMedia>
                                    <EmptyTitle>{t("course.noPosts")}</EmptyTitle>
                                    <EmptyDescription>{t("course.noPostsDescription")}</EmptyDescription>
                                </EmptyHeader>
                            </Empty>
                        </Card>
                    ) : (
                        <div className="space-y-3">
                            {posts.map((post) => <PostCard
                                post={post}
                                labels={{
                                    edited: t("posts.edited"),
                                    posted: t("posts.posted"),
                                    open: t("course.openPost", {title: post.title}),
                                }}
                                key={post.id}
                            />)}
                        </div>
                    )}
                </section>

                <aside className="space-y-4 xl:sticky xl:top-22" aria-label={t("course.details")}>
                    <Card>
                        <CardHeader>
                            <CardTitle className="text-base">{t("course.overview")}</CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <Detail icon={Layers3} label={t("course.grade")} value={course.grade.name} />
                            <Detail icon={BookOpen} label={t("course.course")} value={course.name} />
                            {posts[0] && <Detail icon={CalendarDays} label={t("course.latestUpdate")} value={<LocalDateTime value={posts[0].createdAt} />} />}
                        </CardContent>
                    </Card>

                    <Card className="border-dashed bg-card/60">
                        <CardContent className="py-5">
                            <p className="text-sm font-medium">{t("course.space")}</p>
                            <p className="mt-1 text-xs leading-relaxed text-muted-foreground">{t("course.spaceDescription")}</p>
                        </CardContent>
                    </Card>
                </aside>
            </div>
                </TabsPanel>

                <TabsPanel className="pt-4" value="assignments">
                    <AssignmentsSection assignments={assignments} referenceDate={referenceDate} referenceTime={referenceTime} />
                </TabsPanel>
            </Tabs>
        </div>
    )
}

function PostCard({post, labels}: {
    post: PortalCoursePost
    labels: {edited: string; posted: string; open: string}
}) {
    const postColor = color(post.accentColor)
    return (
        <Card className="relative overflow-hidden">
            <div className="absolute inset-y-0 left-0 w-1.5" style={{backgroundColor: postColor}} />
            <CardHeader className="gap-3 pl-7">
                <div className="flex flex-wrap items-start justify-between gap-3">
                    <CardTitle className="wrap-break-word text-lg leading-snug">{post.title}</CardTitle>
                    <div className="flex items-center gap-2">
                        <Button size="icon-sm" variant="ghost" aria-label={labels.open} render={<Link href={`/app/portal/student/posts/${post.id}`} />}><ExternalLink /></Button>
                        {post.editedAt && (
                            <Tooltip>
                                <TooltipTrigger render={<Badge className="cursor-help" variant="secondary" />}>
                                    <Pencil /> {labels.edited}
                                </TooltipTrigger>
                                <TooltipPopup>{labels.edited} <LocalDateTime value={post.editedAt} /></TooltipPopup>
                            </Tooltip>
                        )}
                    </div>
                </div>
                <div className="flex flex-wrap items-center gap-x-2.5 gap-y-1">
                    <Popover>
                        <PopoverTrigger className="flex min-w-0 cursor-pointer items-center gap-2.5 rounded-lg text-left outline-none hover:[&_[data-author-name]]:underline focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:[&_[data-author-name]]:underline" type="button">
                            <UserAvatar
                                name={post.author.name}
                                src={post.author.profilePictureURL}
                                cacheKey={`portal-course-post:${post.id}:author`}
                                className="size-8 border transition-opacity hover:opacity-80"
                            />
                            <span className="max-w-52 whitespace-nowrap text-xs font-medium underline-offset-2" data-author-name>{post.author.name}</span>
                        </PopoverTrigger>
                        <PopoverPopup className="w-72" align="start">
                            <div className="flex min-w-0 items-center gap-3">
                                <UserAvatar
                                    name={post.author.name}
                                    src={post.author.profilePictureURL}
                                    cacheKey={`portal-course-post:${post.id}:author`}
                                    className="size-10 border"
                                />
                                <div className="min-w-0">
                                    <PopoverTitle className="truncate text-base">{post.author.name}</PopoverTitle>
                                    {post.author.email && (
                                        <PopoverDescription className="mt-1">
                                            <a className="flex min-w-0 items-center gap-1.5 underline-offset-4 hover:text-foreground hover:underline" href={`mailto:${post.author.email}`}>
                                                <Mail className="size-3.5 shrink-0" />
                                                <span className="truncate">{post.author.email}</span>
                                            </a>
                                        </PopoverDescription>
                                    )}
                                </div>
                            </div>
                        </PopoverPopup>
                    </Popover>
                    <span className="text-xs text-muted-foreground">{labels.posted} <LocalDateTime value={post.createdAt} /></span>
                </div>
            </CardHeader>
            <CardContent className="pl-7">
                <p className="whitespace-pre-wrap wrap-break-word text-sm leading-relaxed">{post.body}</p>
                <PostAttachments attachments={post.attachments} postTitle={post.title} />
            </CardContent>
        </Card>
    )
}

function Detail({icon: Icon, label, value}: {icon: typeof BookOpen; label: string; value: React.ReactNode}) {
    return (
        <div className="flex items-start gap-3">
            <span className="flex size-9 shrink-0 items-center justify-center rounded-xl bg-muted text-muted-foreground"><Icon className="size-4" /></span>
            <div className="min-w-0">
                <p className="text-xs text-muted-foreground">{label}</p>
                <p className="mt-0.5 break-words text-sm font-medium">{value}</p>
            </div>
        </div>
    )
}

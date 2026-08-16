import {cookies} from "next/headers"
import {redirect} from "next/navigation"
import {CalendarClock, CircleX, Clock3, GlobeX, Mail, Pencil, ServerCrash} from "lucide-react"
import ErrorPage from "@/components/app/error"
import PostAttachments from "@/components/app/post-attachments"
import LocalDateTime from "@/components/local-date-time"
import {Avatar, AvatarFallback} from "@/components/ui/avatar"
import {Badge} from "@/components/ui/badge"
import {Card, CardContent, CardHeader} from "@/components/ui/card"
import {Popover, PopoverDescription, PopoverPopup, PopoverTitle, PopoverTrigger} from "@/components/ui/popover"
import {Tooltip, TooltipPopup, TooltipTrigger} from "@/components/ui/tooltip"
import {normalizePostAttachments, type PostAttachment} from "@/lib/post_attachments"
import BackButton from "./back_button"

export const metadata = {title: "Post"}

interface CoursePostView {
    attachments: PostAttachment[]
    id: string
    author: {
        id: string
        name: string
        email: string
    }
    title: string
    body: string
    accentColor: string
    showUntil: string | null
    createdAt: string
    editedAt: string | null
}

interface RawCoursePostView extends Omit<CoursePostView, "attachments"> {
    attachments?: unknown
}

interface PostViewResponse {
    posts?: RawCoursePostView[] | null
}

function color(value: string) {
    return /^[0-9A-Fa-f]{6}$/.test(value) ? `#${value}` : "#6366F1"
}

function initials(name: string) {
    return name.trim().split(/\s+/).slice(0, 2).map((part) => part[0]?.toUpperCase()).join("") || "?"
}

export default async function Page({params}: {params: Promise<{postID: string}>}) {
    const {postID} = await params
    if (!/^\d+$/.test(postID)) return <ErrorPage message="This post ID is invalid." icon={CircleX} />

    const token = (await cookies()).get("token")?.value
    let res: Response
    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/course-posts/${postID}`, {
            headers: {Authorization: `Bearer ${token}`},
            cache: "no-store",
        })
    } catch {
        return <ErrorPage message="Network error, please try again." icon={GlobeX} />
    }

    if (res.status === 401) redirect("/auth/login")
    if (res.status === 403) return <ErrorPage message="You don't have permission to view this post." icon={CircleX} />
    if (!res.ok) {
        return <ErrorPage
            message={res.status === 500 ? "The server couldn't load this post." : "This post couldn't be loaded."}
            icon={res.status === 500 ? ServerCrash : CircleX}
        />
    }

    let data: PostViewResponse
    try {
        data = await res.json() as PostViewResponse
    } catch {
        return <ErrorPage message="The server returned an invalid post response." icon={CircleX} />
    }

    const rawPost = Array.isArray(data.posts) ? data.posts[0] : undefined
    if (!rawPost) return <ErrorPage message="This post couldn't be found." icon={CircleX} />

    const post: CoursePostView = {
        ...rawPost,
        attachments: normalizePostAttachments(rawPost.attachments),
    }

    const accentColor = color(post.accentColor)
    return (
        <div className="mx-auto w-full max-w-5xl space-y-5">
            <BackButton />
            <Card className="relative overflow-hidden">
                <div className="absolute inset-x-0 top-0 h-1.5" style={{backgroundColor: accentColor}} />
                <CardHeader className="gap-6 px-5 pb-5 pt-8 sm:px-8 sm:pt-10 lg:px-12">
                    <div className="flex flex-wrap items-start justify-between gap-4">
                        <h1 className="max-w-4xl wrap-break-word text-3xl font-semibold tracking-tight sm:text-4xl lg:text-5xl">{post.title}</h1>
                        {post.editedAt && (
                            <Tooltip>
                                <TooltipTrigger render={<Badge className="cursor-help" variant="secondary" />}>
                                    <Pencil /> Edited
                                </TooltipTrigger>
                                <TooltipPopup>Edited <LocalDateTime value={post.editedAt} /></TooltipPopup>
                            </Tooltip>
                        )}
                    </div>

                    <div className="flex flex-wrap items-center gap-x-4 gap-y-3">
                        <Popover>
                            <PopoverTrigger className="group flex cursor-pointer items-center gap-2.5 rounded-lg text-left outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2" type="button">
                                <Avatar className="size-9 border transition-opacity group-hover:opacity-80">
                                    <AvatarFallback className="text-xs">{initials(post.author.name)}</AvatarFallback>
                                </Avatar>
                                <span className="text-sm font-medium underline-offset-2 group-hover:underline group-focus-visible:underline">{post.author.name}</span>
                            </PopoverTrigger>
                            <PopoverPopup className="w-72" align="start">
                                <div className="flex min-w-0 items-center gap-3">
                                    <Avatar className="size-10 border"><AvatarFallback>{initials(post.author.name)}</AvatarFallback></Avatar>
                                    <div className="min-w-0">
                                        <PopoverTitle className="truncate text-base">{post.author.name}</PopoverTitle>
                                        {post.author.email && (
                                            <PopoverDescription className="mt-1">
                                                <a className="flex min-w-0 items-center gap-1.5 underline-offset-4 hover:text-foreground hover:underline" href={`mailto:${post.author.email}`}>
                                                    <Mail className="size-3.5 shrink-0" /><span className="truncate">{post.author.email}</span>
                                                </a>
                                            </PopoverDescription>
                                        )}
                                    </div>
                                </div>
                            </PopoverPopup>
                        </Popover>
                        <span className="flex items-center gap-1.5 text-sm text-muted-foreground"><Clock3 className="size-4" /> Posted <LocalDateTime value={post.createdAt} /></span>
                        {post.showUntil && (
                            <span className="flex items-center gap-1.5 text-sm text-muted-foreground"><CalendarClock className="size-4" /> Shows until <LocalDateTime value={post.showUntil} /></span>
                        )}
                    </div>
                </CardHeader>
                <CardContent className="border-t px-5 py-8 sm:px-8 sm:py-10 lg:px-12 lg:py-12">
                    <p className="whitespace-pre-wrap wrap-break-word text-base leading-8 sm:text-lg">{post.body}</p>
                    <PostAttachments attachments={post.attachments} postTitle={post.title} />
                </CardContent>
            </Card>
        </div>
    )
}

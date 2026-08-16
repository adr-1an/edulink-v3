import {cookies} from "next/headers"
import {redirect} from "next/navigation"
import {CalendarClock, CircleX, Clock3, GlobeX, Pencil, ServerCrash} from "lucide-react"
import ErrorPage from "@/components/app/error"
import PostAttachments from "@/components/app/post-attachments"
import LocalDateTime from "@/components/local-date-time"
import {Badge} from "@/components/ui/badge"
import {Card, CardContent, CardHeader} from "@/components/ui/card"
import {Tooltip, TooltipPopup, TooltipTrigger} from "@/components/ui/tooltip"
import {normalizePostAttachments, type PostAttachment} from "@/lib/post_attachments"
import BackButton from "./back_button"
import {getTranslations} from "@/i18n/server"

export async function generateMetadata() {
    const {t} = await getTranslations()
    return {title: t("posts.metaTitle")}
}

interface StudentPostView {
    id: string
    attachments: PostAttachment[]
    title: string
    body: string
    accentColor: string
    showUntil: string | null
    createdAt: string
    editedAt: string | null
}

interface PostViewResponse {
    post?: Omit<StudentPostView, "attachments"> & {attachments?: unknown}
}

function color(value: string) {
    return /^[0-9A-Fa-f]{6}$/.test(value) ? `#${value}` : "#6366F1"
}

export default async function Page({params}: {params: Promise<{postID: string}>}) {
    const {t} = await getTranslations()
    const {postID} = await params
    if (!/^\d+$/.test(postID)) return <ErrorPage message={t("posts.error.invalidId")} icon={CircleX} />

    const token = (await cookies()).get("portal_token")?.value
    let res: Response
    try {
        res = await fetch(`${process.env.API_URL}/v1/portal/posts/${postID}`, {
            headers: {Authorization: `Bearer ${token}`},
            cache: "no-store",
        })
    } catch {
        return <ErrorPage message={t("posts.error.network")} icon={GlobeX} />
    }

    if (res.status === 401) redirect("/app/portal")
    if (res.status === 403) return <ErrorPage message={t("posts.error.forbidden")} icon={CircleX} />
    if (!res.ok) {
        return <ErrorPage
            message={res.status === 500 ? t("posts.error.server") : t("posts.error.load")}
            icon={res.status === 500 ? ServerCrash : CircleX}
        />
    }

    let data: PostViewResponse
    try {
        data = await res.json() as PostViewResponse
    } catch {
        return <ErrorPage message={t("posts.error.invalidResponse")} icon={CircleX} />
    }

    const rawPost = data.post
    if (!rawPost) return <ErrorPage message={t("posts.error.notFound")} icon={CircleX} />
    const post: StudentPostView = {
        ...rawPost,
        attachments: normalizePostAttachments(rawPost.attachments),
    }

    return (
        <div className="mx-auto w-full max-w-5xl space-y-5">
            <BackButton />
            <Card className="relative overflow-hidden">
                <div className="absolute inset-x-0 top-0 h-1.5" style={{backgroundColor: color(post.accentColor)}} />
                <CardHeader className="gap-6 px-5 pb-5 pt-8 sm:px-8 sm:pt-10 lg:px-12">
                    <div className="flex flex-wrap items-start justify-between gap-4">
                        <h1 className="max-w-4xl wrap-break-word text-3xl font-semibold tracking-tight sm:text-4xl lg:text-5xl">{post.title}</h1>
                        {post.editedAt && (
                            <Tooltip>
                                <TooltipTrigger render={<Badge className="cursor-help" variant="secondary" />}>
                                    <Pencil /> {t("posts.edited")}
                                </TooltipTrigger>
                                <TooltipPopup>{t("posts.edited")} <LocalDateTime value={post.editedAt} /></TooltipPopup>
                            </Tooltip>
                        )}
                    </div>
                    <div className="flex flex-wrap items-center gap-x-4 gap-y-2 text-sm text-muted-foreground">
                        <span className="flex items-center gap-1.5"><Clock3 className="size-4" /> {t("posts.posted")} <LocalDateTime value={post.createdAt} /></span>
                        {post.showUntil && (
                            <span className="flex items-center gap-1.5"><CalendarClock className="size-4" /> {t("posts.showsUntil")} <LocalDateTime value={post.showUntil} /></span>
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

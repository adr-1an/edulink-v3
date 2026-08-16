import {cookies} from "next/headers"
import {CircleX} from "lucide-react"
import ErrorPage from "@/components/app/error"
import {normalizePostAttachments} from "@/lib/post_attachments"
import {normalizeProfilePictureURL} from "@/lib/profile_picture"
import {emptySchoolAccess, type SchoolAccess} from "@/lib/school_access"
import {type CourseAssignment} from "./assignments_section"
import CourseDashboardClientPage, {type CoursePost, type CourseStudent} from "./client_page"

export const metadata = {title: "Course dashboard"}

interface RawCoursePost extends Omit<CoursePost, "showUntil" | "editedAt" | "attachments" | "authorProfilePictureURL"> {
    attachments?: unknown
    authorProfilePictureURL?: unknown
    showUntil: unknown
    editedAt: unknown
}

interface PostListResponse {
    posts?: RawCoursePost[]
    access?: SchoolAccess
}

interface StudentListResponse {
    students?: CourseStudent[]
    access?: SchoolAccess
}

interface AssignmentListResponse {
    assignments?: CourseAssignment[]
    access?: SchoolAccess
}

function nullableTime(value: unknown): string | null {
    if (typeof value === "string") return Number.isNaN(Date.parse(value)) ? null : value
    if (!value || typeof value !== "object") return null

    const time = "Time" in value ? value.Time : "time" in value ? value.time : null
    const valid = "Valid" in value ? value.Valid : "valid" in value ? value.valid : false
    return valid === true && typeof time === "string" && !Number.isNaN(Date.parse(time)) ? time : null
}

async function loadEndpoint(path: string, token?: string) {
    try {
        return await fetch(`${process.env.API_URL}${path}`, {
            headers: {Authorization: `Bearer ${token}`},
            cache: "no-store",
        })
    } catch {
        return null
    }
}

export default async function Page({params}: {
    params: Promise<{courseID: string}>
}) {
    const {courseID} = await params
    const token = (await cookies()).get("token")?.value
    const [postsRes, studentsRes, assignmentsRes] = await Promise.all([
        loadEndpoint(`/v1/staff/courses/${courseID}/posts`, token),
        loadEndpoint(`/v1/staff/courses/${courseID}/students`, token),
        loadEndpoint(`/v1/staff/courses/${courseID}/assignments`, token),
    ])

    if (postsRes?.status === 400 || studentsRes?.status === 400 || assignmentsRes?.status === 400) {
        return <ErrorPage message="This course ID is invalid." icon={CircleX} />
    }

    let postData: PostListResponse = {}
    let studentData: StudentListResponse = {}
    let assignmentData: AssignmentListResponse = {}
    if (postsRes?.ok) postData = await postsRes.json() as PostListResponse
    if (studentsRes?.ok) studentData = await studentsRes.json() as StudentListResponse
    if (assignmentsRes?.ok) assignmentData = await assignmentsRes.json() as AssignmentListResponse

    const posts: CoursePost[] = (postData.posts ?? []).map((post) => {
        const showUntil = nullableTime(post.showUntil)
        const editedAt = nullableTime(post.editedAt)
        return {
            ...post,
            attachments: normalizePostAttachments(post.attachments),
            authorProfilePictureURL: normalizeProfilePictureURL(post.authorProfilePictureURL),
            showUntil,
            editedAt,
        }
    })

    const postSnapshot = posts
        .map((post) => `${post.id}:${post.editedAt ?? ""}:${post.attachments.map((attachment) => `${attachment.id}:${attachment.fileName}:${attachment.contentType}`).join(",")}`)
        .join("|")
    const studentSnapshot = (studentData.students ?? []).map((student) => `${student.id}:${student.assigned}`).join("|")
    const postsError = !postsRes
        ? "Network error, please try again."
        : !postsRes.ok && postsRes.status !== 403
            ? postsRes.status === 500 ? "The server couldn't load course posts." : "Unable to load course posts."
            : undefined
    const studentsError = !studentsRes
        ? "Network error, please try again."
        : !studentsRes.ok && studentsRes.status !== 403
            ? studentsRes.status === 500 ? "The server couldn't load the course roster." : "Unable to load the course roster."
            : undefined
    const assignmentsError = !assignmentsRes
        ? "Network error, please try again."
        : !assignmentsRes.ok && assignmentsRes.status !== 403
            ? assignmentsRes.status === 500 ? "The server couldn't load course assignments." : "Unable to load course assignments."
            : undefined

    return (
        <CourseDashboardClientPage
            key={`${postSnapshot}/${studentSnapshot}`}
            courseID={courseID}
            initialPosts={posts}
            initialStudents={studentData.students ?? []}
            initialAssignments={assignmentData.assignments ?? []}
            access={postData.access ?? studentData.access ?? assignmentData.access ?? emptySchoolAccess}
            postsDenied={postsRes?.status === 403}
            postsError={postsError}
            studentsDenied={studentsRes?.status === 403}
            studentsError={studentsError}
            assignmentsDenied={assignmentsRes?.status === 403}
            assignmentsError={assignmentsError}
        />
    )
}

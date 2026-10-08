export type StaffSubmissionView = "submitted" | "graded" | "returned"

export function matchesStaffSubmissionView(
    submission: {status: "submitted" | "returned"; graded: boolean},
    view: StaffSubmissionView,
) {
    if (view === "returned") return submission.status === "returned"
    if (submission.status !== "submitted") return false
    return view === "graded" ? submission.graded : !submission.graded
}

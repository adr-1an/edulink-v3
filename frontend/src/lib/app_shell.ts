export function usesImmersiveAppShell(pathname: string) {
    const staffWorkspacePrefixes = [
        "/app/staff/schools/",
        "/app/staff/students/",
        "/app/staff/courses/",
        "/app/staff/grades/",
        "/app/staff/assignments/",
        "/app/staff/submissions/",
    ]

    return staffWorkspacePrefixes.some((prefix) => pathname.startsWith(prefix))
        || pathname.startsWith("/app/portal/student")
        || pathname.startsWith("/app/portal/guardian")
}

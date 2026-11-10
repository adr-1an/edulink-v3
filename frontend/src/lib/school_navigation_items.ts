const sections = [
    {key: "dashboard", path: ""},
    {key: "staff", path: "/staff"},
    {key: "students", path: "/students"},
    {key: "guardians", path: "/guardians"},
    {key: "roles", path: "/roles"},
    {key: "auditLogs", path: "/logs"},
    {key: "settings", path: "/settings"},
] as const

// Navigation is discoverable regardless of permissions; each destination
// handles the backend's permission response when the user opens it.
export function getSchoolNavigationItems(schoolID: string | null | undefined, pathname: string) {
    if (!schoolID) return []

    const base = `/app/staff/schools/${schoolID}`
    return sections.map(({key, path}) => {
        const href = `${base}${path}`
        return {
            key,
            href,
            active: key === "dashboard" ? pathname === href
                : pathname.startsWith(href) || (key === "students" && pathname.startsWith("/app/staff/students/"))
                    || (key === "guardians" && pathname.startsWith("/app/staff/guardians/")),
        }
    })
}

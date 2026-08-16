export interface SchoolAccess {
    owner: boolean
    roles: Array<{
        position: number
        permissions: string[]
    }>
}

export const emptySchoolAccess: SchoolAccess = {owner: false, roles: []}

export function isSchoolAccess(value: unknown): value is SchoolAccess {
    if (!value || typeof value !== "object") return false
    const access = value as Partial<SchoolAccess>
    return typeof access.owner === "boolean"
        && Array.isArray(access.roles)
        && access.roles.every((role) => typeof role?.position === "number"
            && Array.isArray(role.permissions)
            && role.permissions.every((permission) => typeof permission === "string"))
}

export function hasSchoolPermission(access: SchoolAccess | null | undefined, permission: string) {
    return Boolean(access?.owner || access?.roles?.some((role) => role.permissions?.includes(permission)))
}

export function highestRolePosition(access: SchoolAccess | null | undefined) {
    if (access?.owner) return Number.POSITIVE_INFINITY
    return Math.max(-1, ...(access?.roles ?? []).map((role) => role.position))
}

export function canManagePosition(access: SchoolAccess | null | undefined, position: number) {
    return Boolean(access?.owner || highestRolePosition(access) > position)
}

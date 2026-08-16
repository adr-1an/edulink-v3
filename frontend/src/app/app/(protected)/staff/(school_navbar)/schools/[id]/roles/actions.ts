"use server"

import {cookies} from "next/headers"

interface RoleInput {
    name: string
    position: number
    color: string
}

function validateRoleInput(input: RoleInput) {
    const name = input.name.trim()
    const color = input.color.replace(/^#/, "").toUpperCase()
    if (!name || name.length > 32) return {ok: false as const, message: "Role names must be between 1 and 32 characters."}
    if (!Number.isInteger(input.position) || input.position < 0) return {ok: false as const, message: "Position must be a non-negative whole number."}
    if (!/^[0-9A-F]{6}$/.test(color)) return {ok: false as const, message: "Enter a valid six-character hex color."}
    return {ok: true as const, input: {name, position: input.position, color}}
}

export async function handleCreateRole(schoolID: string, input: RoleInput) {
    if (!/^\d+$/.test(schoolID)) return {ok: false, message: "Invalid school."}

    const validated = validateRoleInput(input)
    if (!validated.ok) return validated

    const token = (await cookies()).get("token")?.value
    let res: Response
    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/schools/${schoolID}/roles`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify(validated.input),
        })
    } catch {
        return {ok: false, message: "Network error, please try again."}
    }

    if (res.ok) return {ok: true}
    if (res.status === 400) return {ok: false, message: "The role request was invalid."}
    if (res.status === 401) return {ok: false, message: "Unauthorized."}
    if (res.status === 403) return {ok: false, message: "You don't have permission to create roles."}
    if (res.status === 409) return {ok: false, message: "A role with this name or position already exists."}
    if (res.status === 422) return {ok: false, message: "Check the role name, position, and color."}
    if (res.status === 500) return {ok: false, message: "Internal server error."}
    return {ok: false, message: "Unable to create the role."}
}

export async function handleUpdateRole(roleID: string, input: RoleInput) {
    if (!/^\d+$/.test(roleID)) return {ok: false, message: "Invalid role."}

    const validated = validateRoleInput(input)
    if (!validated.ok) return validated

    const token = (await cookies()).get("token")?.value
    let res: Response
    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/roles/${roleID}`, {
            method: "PATCH",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify(validated.input),
        })
    } catch {
        return {ok: false, message: "Network error, please try again."}
    }

    if (res.ok) return {ok: true}
    if (res.status === 400) return {ok: false, message: "The role request was invalid."}
    if (res.status === 401) return {ok: false, message: "Unauthorized."}
    if (res.status === 403) return {ok: false, message: "You don't have permission to update this role."}
    if (res.status === 409) return {ok: false, message: "A role with this name or position already exists."}
    if (res.status === 422) return {ok: false, message: "Check the role name, position, and color."}
    if (res.status === 500) return {ok: false, message: "Internal server error."}
    return {ok: false, message: "Unable to update the role."}
}

export async function handleDeleteRole(roleID: string) {
    if (!/^\d+$/.test(roleID)) return {ok: false, message: "Invalid role."}

    const token = (await cookies()).get("token")?.value
    let res: Response
    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/roles/${roleID}`, {
            method: "DELETE",
            headers: {Authorization: `Bearer ${token}`},
        })
    } catch {
        return {ok: false, message: "Network error, please try again."}
    }

    if (res.ok) return {ok: true}
    if (res.status === 400) return {ok: false, message: "The role request was invalid."}
    if (res.status === 401) return {ok: false, message: "Unauthorized."}
    if (res.status === 403) return {ok: false, message: "You don't have permission to delete this role."}
    if (res.status === 500) return {ok: false, message: "Internal server error."}
    return {ok: false, message: "Unable to delete the role."}
}

export async function handleSetRolePermission(roleID: string, permission: string, allow: boolean) {
    if (!/^\d+$/.test(roleID) || !permission.trim()) {
        return {ok: false, message: "Invalid role permission."}
    }

    const token = (await cookies()).get("token")?.value
    let res: Response
    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/roles/${roleID}/permissions`, {
            method: "PUT",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({permission: permission.trim(), allow}),
        })
    } catch {
        return {ok: false, message: "Network error, please try again."}
    }

    if (res.ok) return {ok: true}
    if (res.status === 400) return {ok: false, message: "The permission request was invalid."}
    if (res.status === 401) return {ok: false, message: "Unauthorized."}
    if (res.status === 403) return {ok: false, message: "You don't have permission to manage this role."}
    if (res.status === 422) return {ok: false, message: "This permission is not available."}
    if (res.status === 500) return {ok: false, message: "Internal server error."}
    return {ok: false, message: "Unable to update the permission."}
}

interface ReorderRole extends RoleInput {
    id: string
}

export async function handleReorderRoles(
    previousOrder: ReorderRole[],
    orderedRoleIDs: string[],
    draggedRoleID: string,
) {
    const previousIDs = previousOrder.map((role) => role.id)
    if (
        !/^\d+$/.test(draggedRoleID)
        || previousIDs.some((id) => !/^\d+$/.test(id))
        || orderedRoleIDs.some((id) => !/^\d+$/.test(id))
        || new Set(previousIDs).size !== previousIDs.length
        || new Set(orderedRoleIDs).size !== orderedRoleIDs.length
        || previousIDs.length !== orderedRoleIDs.length
        || orderedRoleIDs.some((id) => !previousIDs.includes(id))
        || previousOrder.some((role) => !validateRoleInput(role).ok)
    ) {
        return {ok: false, message: "Invalid role order."}
    }

    const oldIndex = previousIDs.indexOf(draggedRoleID)
    const newIndex = orderedRoleIDs.indexOf(draggedRoleID)
    if (oldIndex === -1 || newIndex === -1) return {ok: false, message: "Invalid role order."}
    if (oldIndex === newIndex) return {ok: true, positions: Object.fromEntries(previousOrder.map((role) => [role.id, role.position]))}

    const draggedRole = previousOrder[oldIndex]
    const targetPosition = orderedRoleIDs.length - 1 - newIndex
    const res = await handleUpdateRole(draggedRole.id, {...draggedRole, position: targetPosition})
    if (!res.ok) return res

    return {
        ok: true,
        positions: Object.fromEntries(orderedRoleIDs.map((id, index) => [id, orderedRoleIDs.length - 1 - index])),
    }
}

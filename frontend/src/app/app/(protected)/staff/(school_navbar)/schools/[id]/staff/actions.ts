"use server"

import {cookies} from "next/headers"

export type EmailImportance = "non-urgent" | "low" | "normal" | "high" | "urgent"
export type StaffActionErrorCode =
    | "invalid_school" | "invalid_invitation" | "invalid_importance" | "invalid_email"
    | "invalid_member" | "invalid_member_role" | "network" | "unauthorized"
    | "list_invitations_forbidden" | "cancel_invitation_forbidden" | "invite_forbidden"
    | "remove_member_forbidden" | "assign_role_forbidden" | "remove_role_forbidden"
    | "privacy_restricted" | "invitation_conflict" | "already_staff"
    | "member_missing" | "member_role_missing" | "server"
    | "load_invitations" | "cancel_invitation" | "send_invitation"
    | "remove_member" | "assign_role" | "remove_role"

function failure<const Code extends StaffActionErrorCode>(code: Code, message: string) {
    return {ok: false as const, code, message}
}

export interface SchoolInvitation {
    id: string
    userEmail: string
    addedBy: {
        id: string
        name: string
        email: string
    }
    status: "pending" | "rejected"
    createdAt: string
    expiresAt: string
}

export async function handleListSchoolInvitations(schoolID: string) {
    if (!/^\d+$/.test(schoolID)) return failure("invalid_school", "Invalid school.")

    const token = (await cookies()).get("token")?.value
    let res: Response
    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/schools/${schoolID}/staff-invitations`, {
            headers: {Authorization: `Bearer ${token}`},
            cache: "no-store",
        })
    } catch {
        return failure("network", "Network error, please try again.")
    }

    if (res.ok) {
        const data = await res.json() as {invitations?: SchoolInvitation[] | null}
        return {ok: true as const, data: Array.isArray(data.invitations) ? data.invitations : []}
    }
    if (res.status === 401) return failure("unauthorized", "Unauthorized.")
    if (res.status === 403) return failure("list_invitations_forbidden", "You don't have permission to view staff invitations.")
    if (res.status === 500) return failure("server", "Internal server error.")
    return failure("load_invitations", "Unable to load staff invitations.")
}

export async function handleCancelSchoolInvitation(invitationID: string) {
    if (!/^\d+$/.test(invitationID)) return failure("invalid_invitation", "Invalid invitation.")

    const token = (await cookies()).get("token")?.value
    let res: Response
    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/staff-invitations/${invitationID}/cancel`, {
            method: "POST",
            headers: {Authorization: `Bearer ${token}`},
        })
    } catch {
        return failure("network", "Network error, please try again.")
    }

    if (res.ok) return {ok: true as const}
    if (res.status === 401) return failure("unauthorized", "Unauthorized.")
    if (res.status === 403) return failure("cancel_invitation_forbidden", "You don't have permission to cancel this invitation.")
    if (res.status === 500) return failure("server", "Internal server error.")
    return failure("cancel_invitation", "Unable to cancel the invitation.")
}

export async function handleInviteStaffMember(schoolID: string, email: string, importance: EmailImportance) {
    if (!/^\d+$/.test(schoolID)) return failure("invalid_school", "Invalid school.")
    if (!["non-urgent", "low", "normal", "high", "urgent"].includes(importance)) {
        return failure("invalid_importance", "Select a valid email importance.")
    }

    const token = (await cookies()).get("token")?.value
    let res
    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/schools/${schoolID}/staff/invitations`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({email, importance: importance === "normal" ? "" : importance}),
        })
    } catch {
        return failure("network", "Network error, please try again.")
    }

    if (res.ok) {
        return {
            ok: true as const,
            outcome: res.status === 201 ? "staff_added" as const : "invitation_sent" as const,
        }
    }

    let code: string | undefined
    try { code = (await res.json()).code } catch {}
    code = code?.toUpperCase()
    if (code === "INVALID_EMAIL") return failure("invalid_email", "Enter a valid email address.")
    if (code === "INVALID_MAIL_IMPORTANCE") return failure("invalid_importance", "Select a valid email importance.")
    if (code === "INVITATION_EMAIL_CONFLICT") return failure("invitation_conflict", "An invitation was already sent to this email in the last 24 hours.")
    if (code === "STAFF_MEMBER_EMAIL_CONFLICT") return failure("already_staff", "This person is already a staff member.")
    if (code === "TARGET_PRIVACY_RESTRICTED" || code === "USER_PRIVACY_RESTRICTED") {
        return failure("privacy_restricted", "This user has disabled staff invitations in their privacy settings.")
    }
    if (res.status === 401) return failure("unauthorized", "Unauthorized.")
    if (res.status === 403) return failure("invite_forbidden", "You don't have permission to invite staff members.")
    if (res.status === 500) return failure("server", "Internal server error.")
    return failure("send_invitation", "Unable to send the invitation.")
}

export async function handleDeleteStaffMember(staffID: string) {
    if (!/^\d+$/.test(staffID)) {
        return failure("invalid_member", "Invalid staff member.")
    }

    const token = (await cookies()).get("token")?.value
    let res
    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/staff-members/${staffID}`, {
            method: "DELETE",
            headers: {Authorization: `Bearer ${token}`},
        })
    } catch {
        return failure("network", "Network error, please try again.")
    }

    if (res.ok) return {ok: true as const}
    if (res.status === 401) return failure("unauthorized", "Unauthorized.")
    if (res.status === 403) return failure("remove_member_forbidden", "You cannot remove this staff member.")
    if (res.status === 404) return failure("member_missing", "This staff member no longer exists.")
    if (res.status === 500) return failure("server", "Internal server error.")
    return failure("remove_member", "Unable to remove the staff member.")
}

async function updateStaffRole(method: "POST" | "DELETE", staffID: string, roleID: string) {
    if (![staffID, roleID].every((id) => /^\d+$/.test(id))) {
        return failure("invalid_member_role", "Invalid staff member or role.")
    }

    const token = (await cookies()).get("token")?.value
    let res
    try {
        res = await fetch(`${process.env.API_URL}/v1/staff/staff-members/${staffID}/roles/${roleID}`, {
            method,
            headers: {Authorization: `Bearer ${token}`},
        })
    } catch {
        return failure("network", "Network error, please try again.")
    }

    if (res.ok) return {ok: true as const}
    if (res.status === 401) return failure("unauthorized", "Unauthorized.")
    if (res.status === 403) return failure(method === "POST" ? "assign_role_forbidden" : "remove_role_forbidden", `You don't have permission to ${method === "POST" ? "assign" : "remove"} this role.`)
    if (res.status === 404) return failure("member_role_missing", "This staff member or role no longer exists.")
    if (res.status === 500) return failure("server", "Internal server error.")
    return failure(method === "POST" ? "assign_role" : "remove_role", `Unable to ${method === "POST" ? "assign" : "remove"} the role.`)
}

export async function handleAddStaffRole(staffID: string, roleID: string) {
    return updateStaffRole("POST", staffID, roleID)
}

export async function handleRemoveStaffRole(staffID: string, roleID: string) {
    return updateStaffRole("DELETE", staffID, roleID)
}

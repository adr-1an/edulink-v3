export type PortalPasswordValidationError = "current_required" | "too_short" | "mismatch"

export function validatePortalPasswordUpdate(
    currentPassword: string,
    newPassword: string,
    confirmation: string,
): PortalPasswordValidationError | null {
    if (!currentPassword) return "current_required"
    if (newPassword.length < 8) return "too_short"
    if (newPassword !== confirmation) return "mismatch"
    return null
}

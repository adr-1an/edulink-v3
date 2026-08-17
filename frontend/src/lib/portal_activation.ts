export type PortalActivationError = "invalid_link" | "expired_link" | "invalid_password" | "generic"

export function validatePortalActivationPasswords(password: string, confirmation: string) {
    if (password.length < 8) return "too_short" as const
    if (password !== confirmation) return "mismatch" as const
    return null
}

export function portalActivationErrorFromResponse(status: number, code: unknown): PortalActivationError {
    if (status === 401 && code === "EXPIRED_TOKEN") return "expired_link"
    if (status === 401 && (code === "INVALID_TOKEN" || code === "NO_TOKEN")) return "invalid_link"
    if (status === 422) return "invalid_password"
    return "generic"
}

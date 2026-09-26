import {type MessageKey} from "@/i18n/messages"
import {type GuardianActionErrorCode} from "./actions"

export const guardianErrorKeys: Record<GuardianActionErrorCode, MessageKey> = {
    invalid_school: "staff.guardians.error.invalidSchool",
    invalid_guardian: "staff.guardians.error.invalidGuardian",
    network: "staff.guardians.error.network",
    invalid_data: "staff.guardians.error.invalidData",
    unauthorized: "staff.guardians.error.unauthorized",
    forbidden: "staff.guardians.error.forbidden",
    missing: "staff.guardians.error.missing",
    conflict: "staff.guardians.error.conflict",
    validation: "staff.guardians.error.validation",
    server: "staff.guardians.error.server",
    save: "staff.guardians.error.save",
}

import {type MessageKey} from "@/i18n/messages"
import {type StudentActionErrorCode} from "./actions"

export const studentErrorKeys = {
    invalid_school: "staff.students.error.invalidSchool",
    invalid_student: "staff.students.error.invalidStudent",
    network: "staff.students.error.network",
    invalid_data: "staff.students.error.invalidData",
    unauthorized: "staff.students.error.unauthorized",
    forbidden: "staff.students.error.forbidden",
    missing: "staff.students.error.missing",
    conflict: "staff.students.error.conflict",
    validation: "staff.students.error.validation",
    server: "staff.students.error.server",
    save: "staff.students.error.save",
} satisfies Record<StudentActionErrorCode, MessageKey>

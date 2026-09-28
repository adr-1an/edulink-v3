import {type MessageKey} from "@/i18n/messages"
import {type PortalUserActivationErrorCode} from "./portal_user_actions"

export const portalUserActivationErrorKeys = {
    invalid_user: "staff.portalUsers.activation.error.invalidUser",
    network: "staff.portalUsers.activation.error.network",
    unauthorized: "staff.portalUsers.activation.error.unauthorized",
    forbidden: "staff.portalUsers.activation.error.forbidden",
    already_active: "staff.portalUsers.activation.error.alreadyActive",
    server: "staff.portalUsers.activation.error.server",
    send: "staff.portalUsers.activation.error.send",
} satisfies Record<PortalUserActivationErrorCode, MessageKey>

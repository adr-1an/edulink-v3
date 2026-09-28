export interface PortalUserActivationLinkState {
    accountEnabled: boolean
    accountActive: boolean
    password: string
    activationLinkSent?: boolean
}

export function canSendPortalUserActivationLink(state: PortalUserActivationLinkState) {
    return state.accountEnabled && !state.accountActive && state.password.length === 0 && !state.activationLinkSent
}

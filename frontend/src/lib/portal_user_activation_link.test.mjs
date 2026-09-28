import assert from "node:assert/strict"
import test from "node:test"
import {canSendPortalUserActivationLink} from "./portal_user_activation_link.ts"

test("allows activation links only for enabled inactive accounts without a new password", () => {
    assert.equal(canSendPortalUserActivationLink({accountEnabled: true, accountActive: false, password: ""}), true)
    assert.equal(canSendPortalUserActivationLink({accountEnabled: false, accountActive: false, password: ""}), false)
    assert.equal(canSendPortalUserActivationLink({accountEnabled: true, accountActive: false, password: "new-password"}), false)
    assert.equal(canSendPortalUserActivationLink({accountEnabled: true, accountActive: true, password: ""}), false)
    assert.equal(canSendPortalUserActivationLink({accountEnabled: true, accountActive: false, password: "", activationLinkSent: true}), false)
})

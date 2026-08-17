import assert from "node:assert/strict"
import test from "node:test"
import {
    portalActivationErrorFromResponse,
    validatePortalActivationPasswords,
} from "./portal_activation.ts"

test("validates activation passwords", () => {
    assert.equal(validatePortalActivationPasswords("short", "short"), "too_short")
    assert.equal(validatePortalActivationPasswords("long-enough", "different"), "mismatch")
    assert.equal(validatePortalActivationPasswords("long-enough", "long-enough"), null)
})

test("maps activation link errors without trusting an arbitrary response code", () => {
    assert.equal(portalActivationErrorFromResponse(401, "EXPIRED_TOKEN"), "expired_link")
    assert.equal(portalActivationErrorFromResponse(401, "INVALID_TOKEN"), "invalid_link")
    assert.equal(portalActivationErrorFromResponse(401, "NO_TOKEN"), "invalid_link")
    assert.equal(portalActivationErrorFromResponse(401, "SOMETHING_ELSE"), "generic")
    assert.equal(portalActivationErrorFromResponse(422, undefined), "invalid_password")
})

import assert from "node:assert/strict"
import test from "node:test"
import {validatePortalPasswordUpdate} from "./portal_password.ts"

test("validates portal password updates", () => {
    assert.equal(validatePortalPasswordUpdate("", "long-enough", "long-enough"), "current_required")
    assert.equal(validatePortalPasswordUpdate("current-pass", "short", "short"), "too_short")
    assert.equal(validatePortalPasswordUpdate("current-pass", "long-enough", "different"), "mismatch")
    assert.equal(validatePortalPasswordUpdate("current-pass", "long-enough", "long-enough"), null)
})

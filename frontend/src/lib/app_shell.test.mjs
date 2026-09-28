import test from "node:test"
import assert from "node:assert/strict"

import {usesImmersiveAppShell} from "./app_shell.ts"

test("uses the edge-to-edge app shell for staff school workspaces", () => {
    assert.equal(usesImmersiveAppShell("/app/staff/schools/42"), true)
    assert.equal(usesImmersiveAppShell("/app/staff/schools/42/students"), true)
    assert.equal(usesImmersiveAppShell("/app/staff/students/7"), true)
    assert.equal(usesImmersiveAppShell("/app/staff/courses/9"), true)
    assert.equal(usesImmersiveAppShell("/app/staff/grades/3"), true)
    assert.equal(usesImmersiveAppShell("/app/staff/assignments/11/submissions"), true)
    assert.equal(usesImmersiveAppShell("/app/staff/submissions/12"), true)
})

test("uses the edge-to-edge app shell for student and guardian portals", () => {
    assert.equal(usesImmersiveAppShell("/app/portal/student"), true)
    assert.equal(usesImmersiveAppShell("/app/portal/guardian/profile"), true)
})

test("keeps regular page spacing outside signed-in workspaces", () => {
    assert.equal(usesImmersiveAppShell("/app"), false)
    assert.equal(usesImmersiveAppShell("/app/staff/profile"), false)
    assert.equal(usesImmersiveAppShell("/auth/login"), false)
})

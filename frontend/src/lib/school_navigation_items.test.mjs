import assert from "node:assert/strict"
import test from "node:test"
import {getSchoolNavigationItems} from "./school_navigation_items.ts"

test("shows every school destination without requiring permission data", () => {
    const items = getSchoolNavigationItems("42", "/app/staff/schools/42/students")
    assert.deepEqual(items.map(({href}) => href), [
        "/app/staff/schools/42",
        "/app/staff/schools/42/staff",
        "/app/staff/schools/42/students",
        "/app/staff/schools/42/guardians",
        "/app/staff/schools/42/roles",
        "/app/staff/schools/42/logs",
        "/app/staff/schools/42/settings",
    ])
    assert.deepEqual(items.filter(({active}) => active).map(({key}) => key), ["students"])
})

test("keeps Students active on student guardian pages", () => {
    const items = getSchoolNavigationItems("42", "/app/staff/students/7/guardians")
    assert.deepEqual(items.filter(({active}) => active).map(({key}) => key), ["students"])
})

test("does not invent school links when there is no school context", () => {
    assert.deepEqual(getSchoolNavigationItems(null, "/app/staff/students/7"), [])
})

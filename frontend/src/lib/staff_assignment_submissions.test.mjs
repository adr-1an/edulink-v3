import assert from "node:assert/strict"
import test from "node:test"
import {matchesStaffSubmissionView} from "./staff_assignment_submissions.ts"

const submitted = {status: "submitted", graded: false}
const graded = {status: "submitted", graded: true}
const returned = {status: "returned", graded: false}

test("separates graded submissions from submissions awaiting a grade", () => {
    assert.equal(matchesStaffSubmissionView(submitted, "submitted"), true)
    assert.equal(matchesStaffSubmissionView(graded, "submitted"), false)
    assert.equal(matchesStaffSubmissionView(graded, "graded"), true)
    assert.equal(matchesStaffSubmissionView(submitted, "graded"), false)
})

test("keeps returned submissions in the returned tab", () => {
    assert.equal(matchesStaffSubmissionView(returned, "returned"), true)
    assert.equal(matchesStaffSubmissionView(returned, "submitted"), false)
    assert.equal(matchesStaffSubmissionView(returned, "graded"), false)
})

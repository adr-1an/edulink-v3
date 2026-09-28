import assert from "node:assert/strict"
import test from "node:test"
import {studentDraftToInput} from "./student.ts"

const validDraft = {
    name: "  Avery ",
    lastName: " Morgan  ",
    dateOfBirth: "2012-04-12",
    email: " AVERY@EXAMPLE.COM ",
    phone: "  +1 555 0100  ",
    notes: "  Needs a locker  ",
    accountEnabled: true,
    accountActive: false,
    password: "long-enough",
}

test("builds student account activation state only when login is enabled", () => {
    assert.equal(studentDraftToInput(validDraft).accountActive, false)
    assert.equal(studentDraftToInput({...validDraft, accountActive: true}).accountActive, true)

    const disabledPayload = studentDraftToInput({...validDraft, accountEnabled: false})
    assert.equal(Object.hasOwn(disabledPayload, "accountActive"), false)
})

import assert from "node:assert/strict"
import test from "node:test"
import {
    guardianDraftToInput,
    isGuardianDraftValid,
    normalizeGuardianList,
} from "./guardian.ts"

const validDraft = {
    name: "  Avery ",
    lastName: " Morgan  ",
    dateOfBirth: "1988-04-12",
    email: " AVERY@EXAMPLE.COM ",
    phone: "  +1 555 0100  ",
    notes: "  Primary contact  ",
    accountEnabled: true,
    activateAccount: false,
    password: "long-enough",
}

test("normalizes the guardian list contract without losing string IDs", () => {
    assert.deepEqual(normalizeGuardianList([{
        id: "9223372036854775806",
        name: "Avery",
        lastName: "Morgan",
        email: "avery@example.com",
        phone: null,
        notes: null,
        dateOfBirth: "1988-04-12T00:00:00Z",
        accountEnabled: true,
        accountActive: false,
    }]), [{
        id: "9223372036854775806",
        name: "Avery",
        lastName: "Morgan",
        email: "avery@example.com",
        phone: "",
        notes: "",
        dateOfBirth: "1988-04-12",
        accountEnabled: true,
        accountActive: false,
    }])
})

test("ignores malformed guardian rows instead of exposing unusable actions", () => {
    assert.deepEqual(normalizeGuardianList([
        null,
        {id: 9223372036854775806, name: "Unsafe numeric ID"},
        {id: "12", name: "Missing fields"},
    ]), [])
})

test("builds the guardian API payload with trimmed nullable values and an RFC 3339 date", () => {
    assert.deepEqual(guardianDraftToInput(validDraft), {
        name: "Avery",
        lastName: "Morgan",
        dateOfBirth: "1988-04-12T00:00:00.000Z",
        email: "avery@example.com",
        phone: "+1 555 0100",
        notes: "Primary contact",
        accountEnabled: true,
        activateAccount: false,
        password: "long-enough",
    })

    assert.equal(guardianDraftToInput({...validDraft, phone: " ", notes: " ", password: ""}).phone, null)
    assert.equal(guardianDraftToInput({...validDraft, phone: " ", notes: " ", password: ""}).notes, null)
    assert.equal(guardianDraftToInput({...validDraft, phone: " ", notes: " ", password: ""}).password, null)
})

test("only requests guardian account activation when login is enabled", () => {
    assert.equal(guardianDraftToInput(validDraft).activateAccount, false)
    assert.equal(guardianDraftToInput({...validDraft, activateAccount: true}).activateAccount, true)

    const disabledPayload = guardianDraftToInput({...validDraft, accountEnabled: false})
    assert.equal(Object.hasOwn(disabledPayload, "activateAccount"), false)
})

test("requires a password only when a guardian login is newly enabled", () => {
    const noPassword = {...validDraft, password: ""}
    assert.equal(isGuardianDraftValid(noPassword, "create", false), false)
    assert.equal(isGuardianDraftValid(noPassword, "edit", false), false)
    assert.equal(isGuardianDraftValid(noPassword, "edit", true), true)
    assert.equal(isGuardianDraftValid({...noPassword, accountEnabled: false}, "create", false), true)
})

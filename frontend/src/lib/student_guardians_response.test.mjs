import assert from "node:assert/strict"
import test from "node:test"
import {parseStudentGuardiansResponseJSON} from "./student_guardians.ts"

test("preserves the numeric schoolId from the updated handler without rounding", () => {
    assert.deepEqual(parseStudentGuardiansResponseJSON('{"schoolId":80710177170391041,"access":{"owner":true,"roles":[]},"guardians":[{"id":"81098371665756161","name":"Anna","lastName":"Smith","dateOfBirth":"1980-01-01","email":"anna@example.com","phone":null}]}'), {
        schoolID: "80710177170391041",
        access: {owner: true, roles: []},
        guardians: [{id: "81098371665756161", name: "Anna", lastName: "Smith", email: "anna@example.com", phone: ""}],
    })
})

test("resolves school context even when the guardian list is null or empty", () => {
    assert.deepEqual(parseStudentGuardiansResponseJSON('{"schoolId":"80710177170391041","guardians":null}'), {
        schoolID: "80710177170391041", access: undefined, guardians: [],
    })
    assert.equal(parseStudentGuardiansResponseJSON('{"schoolID":"80710177170391041","guardians":[]}').schoolID, "80710177170391041")
})

test("rejects absent or malformed school context and malformed guardian lists", () => {
    for (const body of ['{}', 'null', 'not json', '{"guardians":[]}', '{"schoolId":null,"guardians":[]}', '{"schoolId":"../school","guardians":[]}', '{"schoolId":1.5,"guardians":[]}', '{"schoolId":"42","guardians":[{}]}']) {
        assert.equal(parseStudentGuardiansResponseJSON(body), null)
    }
})

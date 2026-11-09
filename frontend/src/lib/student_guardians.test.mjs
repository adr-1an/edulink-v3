import assert from "node:assert/strict"
import test from "node:test"
import {parseStudentGuardiansJSON} from "./student_guardians.ts"

test("preserves int64 guardian IDs from the new Go route", () => {
    assert.deepEqual(parseStudentGuardiansJSON('{"guardians":[{"ID":81098371665756161,"Name":"Anna","LastName":"Smith","Email":"anna@example.com","Phone":null}]}'), [
        {id: "81098371665756161", name: "Anna", lastName: "Smith", email: "anna@example.com", phone: ""},
    ])
})

test("accepts camelCase guardian records and empty lists", () => {
    assert.deepEqual(parseStudentGuardiansJSON('{"guardians":[{"id":"123","name":"Anna","lastName":"Smith","email":"anna@example.com","phone":"123"}]}'), [
        {id: "123", name: "Anna", lastName: "Smith", email: "anna@example.com", phone: "123"},
    ])
    assert.deepEqual(parseStudentGuardiansJSON('{"guardians":null}'), [])
    assert.deepEqual(parseStudentGuardiansJSON('{"guardians":[]}'), [])
})

test("rejects malformed responses and duplicate guardian IDs", () => {
    assert.equal(parseStudentGuardiansJSON('{}'), null)
    assert.equal(parseStudentGuardiansJSON('not json'), null)
    assert.equal(parseStudentGuardiansJSON('{"guardians":[{"ID":123}]}'), null)
    const row = {id: "1", name: "Anna", lastName: "Smith", email: "anna@example.com", phone: null}
    assert.equal(parseStudentGuardiansJSON(JSON.stringify({guardians: [row, row]})), null)
})

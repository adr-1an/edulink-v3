import assert from "node:assert/strict"
import test from "node:test"
import {parseStudentGuardiansResponseJSON} from "./student_guardians.ts"

const guardian = {id: "81098371665756161", name: "Anna", lastName: "Smith", email: "anna@example.com", phone: null}

test("reads embedded student guardians from the student view response", () => {
    const parsed = parseStudentGuardiansResponseJSON(JSON.stringify({
        schoolId: "80710177170391041", access: {owner: true, roles: []},
        student: {id: "7", name: "Student", lastName: "Smith", guardians: [guardian]}, assignmentSubmissions: null,
    }))
    assert.deepEqual(parsed, {
        schoolID: "80710177170391041", access: {owner: true, roles: []},
        guardians: [{...guardian, phone: ""}],
    })
})

test("handles null embedded guardians and repeated guardians from submission joins", () => {
    assert.deepEqual(parseStudentGuardiansResponseJSON('{"schoolId":80710177170391041,"student":{"guardians":null}}')?.guardians, [])
    const parsed = parseStudentGuardiansResponseJSON(JSON.stringify({schoolId: "42", student: {guardians: [guardian, guardian]}}))
    assert.deepEqual(parsed?.guardians, [{...guardian, phone: ""}])
})

test("rejects missing or conflicting embedded guardian records", () => {
    assert.equal(parseStudentGuardiansResponseJSON('{"schoolId":"42","student":{}}'), null)
    assert.equal(parseStudentGuardiansResponseJSON(JSON.stringify({schoolId: "42", student: {guardians: [guardian, {...guardian, email: "other@example.com"}]}})), null)
})

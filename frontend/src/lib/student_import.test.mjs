import assert from "node:assert/strict"
import test from "node:test"
import {parseStudentImportFile} from "./student_import.ts"

test("parses quoted CSV fields and multiline notes", () => {
    const result = parseStudentImportFile('firstName,lastName,dateOfBirth,email,phone,notes\r\nAlex,Johnson,2010-04-23,alex@example.com,,"First line\nSecond, line"')
    assert.equal(result.ok, true)
    if (!result.ok) return
    assert.deepEqual(result.rows, [{
        name: "Alex",
        lastName: "Johnson",
        dateOfBirth: "2010-04-23",
        email: "alex@example.com",
        phone: "",
        notes: "First line\nSecond, line",
    }])
})

test("detects semicolon-delimited CSV with Polish headers", () => {
    const result = parseStudentImportFile("Imię;Nazwisko;Data urodzenia;E-mail;Telefon;Uwagi\nJan;Kowalski;2011-02-03;JAN@EXAMPLE.COM;123;Test")
    assert.equal(result.ok, true)
    if (!result.ok) return
    assert.equal(result.rows[0].name, "Jan")
    assert.equal(result.rows[0].email, "jan@example.com")
})

test("parses TSV and leaves optional columns empty", () => {
    const result = parseStudentImportFile("name\tlastName\tdob\temail\nTaylor\tMorgan\t2009-12-01\ttaylor@example.com")
    assert.equal(result.ok, true)
    if (!result.ok) return
    assert.equal(result.rows[0].phone, "")
    assert.equal(result.rows[0].notes, "")
})

test("reports missing required headers", () => {
    const result = parseStudentImportFile("firstName,lastName,email\nAlex,Johnson,alex@example.com")
    assert.deepEqual(result, {ok: false, error: "missing_headers", missingHeaders: ["dateOfBirth"]})
})

test("rejects unfinished quoted fields", () => {
    const result = parseStudentImportFile('firstName,lastName,dateOfBirth,email\n"Alex,Johnson,2010-04-23,alex@example.com')
    assert.equal(result.ok, false)
    if (result.ok) return
    assert.equal(result.error, "invalid_csv")
})

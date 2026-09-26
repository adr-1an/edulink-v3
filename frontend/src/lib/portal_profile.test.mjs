import assert from "node:assert/strict"
import test from "node:test"
import {parsePortalProfile} from "./portal_profile.ts"

test("parses the complete portal profile contract", () => {
    assert.deepEqual(parsePortalProfile({
        pfpUrl: "https://cdn.example.com/profile.png",
        name: "Avery",
        lastName: "Morgan",
        email: "avery@example.com",
        phone: "+1 555 0100",
        dateOfBirth: "1988-04-12T00:00:00Z",
        school: {
            name: "North High",
            region: "US-CA",
            owner: {name: "Jordan Lee", email: "jordan@example.com"},
        },
    }), {
        profilePictureURL: "https://cdn.example.com/profile.png",
        name: "Avery",
        lastName: "Morgan",
        email: "avery@example.com",
        phone: "+1 555 0100",
        dateOfBirth: "1988-04-12T00:00:00Z",
        school: {
            name: "North High",
            region: "US-CA",
            owner: {name: "Jordan Lee", email: "jordan@example.com"},
        },
    })
})

test("accepts an absent profile picture and empty optional contact fields", () => {
    const profile = parsePortalProfile({
        pfpUrl: null,
        name: "Avery",
        lastName: "Morgan",
        email: "avery@example.com",
        phone: "",
        dateOfBirth: "1988-04-12T00:00:00Z",
        school: {
            name: "North High",
            region: "",
            owner: {name: "Jordan Lee", email: "jordan@example.com"},
        },
    })

    assert.equal(profile?.profilePictureURL, null)
    assert.equal(profile?.phone, "")
    assert.equal(profile?.school.region, "")
})

test("discards unsafe profile picture URLs", () => {
    const profile = parsePortalProfile({
        pfpUrl: "javascript:alert('nope')",
        name: "Avery",
        lastName: "Morgan",
        email: "avery@example.com",
        phone: "",
        dateOfBirth: "1988-04-12T00:00:00Z",
        school: {
            name: "North High",
            region: "",
            owner: {name: "Jordan Lee", email: "jordan@example.com"},
        },
    })

    assert.equal(profile?.profilePictureURL, null)
})

test("rejects incomplete profile and school data", () => {
    assert.equal(parsePortalProfile(null), null)
    assert.equal(parsePortalProfile({name: "Avery"}), null)
    assert.equal(parsePortalProfile({
        pfpUrl: null,
        name: "Avery",
        lastName: "Morgan",
        email: "avery@example.com",
        phone: "",
        dateOfBirth: "1988-04-12T00:00:00Z",
        school: {name: "North High", region: "", owner: null},
    }), null)
})

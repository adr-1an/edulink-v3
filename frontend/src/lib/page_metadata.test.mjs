import assert from "node:assert/strict"
import test from "node:test"
import {formatPageTitle, PAGE_TITLES, SITE_TITLE} from "./page_metadata.ts"

test("page titles use one EduLink suffix without changing title casing", () => {
    assert.equal(SITE_TITLE, "EduLink")
    assert.equal(formatPageTitle("Course Dashboard"), "Course Dashboard | EduLink")
    assert.equal(formatPageTitle(), "EduLink")
})

test("static routes expose consistent descriptive titles", () => {
    assert.deepEqual(PAGE_TITLES, {
        completeRegistration: "Complete Registration",
        courseDashboard: "Course Dashboard",
        coursePost: "Course Post",
        forgotPassword: "Forgot Password",
        privacyPolicy: "Privacy Policy",
        resetPassword: "Reset Password",
        staffInvitation: "Staff Invitation",
        staffLogin: "Staff Login",
        staffProfile: "Staff Profile",
        staffRegistration: "Staff Registration",
        termsOfService: "Terms of Service",
        verifyEmailChange: "Verify Email Change",
    })
})

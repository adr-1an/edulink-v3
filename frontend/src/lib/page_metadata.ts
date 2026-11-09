export const SITE_TITLE = "EduLink"

export const PAGE_TITLES = {
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
} as const

export function formatPageTitle(title?: string) {
    return title ? `${title} | ${SITE_TITLE}` : SITE_TITLE
}

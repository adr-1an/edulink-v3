package staff_helpers

// Verification token purposes

type VerificationTokenPurpose string

const (
	TokenPurposeEmailChange   VerificationTokenPurpose = "email_change"
	TokenPurposePasswordReset VerificationTokenPurpose = "password_reset"
)

// Error codes are returned alongside response codes (e.g. 500) in cases where
// one handler may return the same response code for different errors.

type ResErrCode string

const (
	ErrorCodeInvalidToken             ResErrCode = "INVALID_TOKEN"
	ErrorCodeInvalidRecoveryCode      ResErrCode = "INVALID_RECOVERY_CODE"
	ErrorCodeExpiredToken             ResErrCode = "EXPIRED_TOKEN"
	ErrorCodeInvalidEmail             ResErrCode = "INVALID_EMAIL"
	ErrorCodeNoToken                  ResErrCode = "NO_TOKEN"
	ErrorCodeIncorrectPassword        ResErrCode = "INCORRECT_PASSWORD"
	ErrorCodeInvalidRegionCode        ResErrCode = "INVALID_REGION_CODE"
	ErrorCodeInvalidName              ResErrCode = "INVALID_NAME"
	ErrorCodeInvalidMailImportance    ResErrCode = "INVALID_MAIL_IMPORTANCE"
	ErrorCodeInvitationEmailConflict  ResErrCode = "INVITATION_EMAIL_CONFLICT"
	ErrorCodeStaffMemberEmailConflict ResErrCode = "STAFF_MEMBER_EMAIL_CONFLICT"
	ErrorCodeAcademicYearConflict     ResErrCode = "ACADEMIC_YEAR_CONFLICT"
	ErrorCodeMissingLevelVar          ResErrCode = "MISSING_LEVEL_VAR"
	ErrorCodeLevelOutOfRange          ResErrCode = "LEVEL_OUT_OF_RANGE"
	ErrorCodeNoActiveYear             ResErrCode = "NO_ACTIVE_YEAR"
	ErrorCodeTargetPrivacyRestricted  ResErrCode = "TARGET_PRIVACY_RESTRICTED"
	ErrorPayloadTooLong               ResErrCode = "PAYLOAD_TOO_LONG"
)

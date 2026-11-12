package auth

// Verification token purposes

type VerificationTokenPurpose string

const (
	TokenPurposeEmailChange   VerificationTokenPurpose = "email_change"
	TokenPurposePasswordReset VerificationTokenPurpose = "password_reset"
)

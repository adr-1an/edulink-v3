package auth

type TwoFactorChallengePurpose string

const (
	ChallengePurposeLogin          TwoFactorChallengePurpose = "login"
	ChallengePurposeSchoolDeletion TwoFactorChallengePurpose = "schoolDeletion"
)

func (p TwoFactorChallengePurpose) String() string {
	return string(p)
}

const (
	TwoFAStatusEnabled  string = "enabled"
	TwoFAStatusPending  string = "pending"
	TwoFAStatusDisabled string = "disabled"
)

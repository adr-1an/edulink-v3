package portal_users

import (
	"errors"
	"testing"
)

func TestValidateActivationEmailState(t *testing.T) {
	if err := validateActivationEmailState(true, false); err != nil {
		t.Fatalf("enabled inactive account was rejected: %v", err)
	}
	if err := validateActivationEmailState(false, false); !errors.Is(err, ErrAccountDisabled) {
		t.Fatalf("disabled account error = %v, want %v", err, ErrAccountDisabled)
	}
	if err := validateActivationEmailState(true, true); !errors.Is(err, ErrAccountAlreadyActive) {
		t.Fatalf("active account error = %v, want %v", err, ErrAccountAlreadyActive)
	}
}

package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	helpers2 "app/internal/helpers"

	"github.com/pquerna/otp/totp"
)

type VerifyTwoFactorInput struct {
	UserID int64
	TOTP   string
}
type VerifyTwoFactorResult struct{ Codes map[int]string }

func (s *Service) VerifyTwoFactor(ctx context.Context, i *VerifyTwoFactorInput) (*VerifyTwoFactorResult, error) {
	userID := i.UserID

	i.TOTP = strings.TrimSpace(i.TOTP)

	if len(i.TOTP) != 6 {
		return nil, ErrUnprocessableEntity
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var totpSecret []byte
	if err := tx.QueryRowContext(ctx, `
		SELECT totp_secret FROM users WHERE id = $1 AND two_factor_status = $2
	`, userID, TwoFAStatusPending).Scan(&totpSecret); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrForbidden
		} else {
			log.Println(err)
			return nil, err
		}
	}

	appKey, err := helpers2.LoadAppEncKey()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	secret, err := helpers2.DecryptString(totpSecret, appKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if !totp.Validate(i.TOTP, secret) {
		return nil, ErrUnauthorized
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE users
		SET two_factor_status = $1
		WHERE id = $2
	`, TwoFAStatusEnabled, userID); err != nil {
		log.Println(err)
		return nil, err
	}

	codes := make(map[int]string)
	values := make([]string, 0, 8)
	args := make([]any, 0, 16)

	for i := range 8 {
		code, err := helpers2.GenerateRecoveryCode()
		if err != nil {
			log.Println(err)
			return nil, err
		}

		codeHash := helpers2.MakeHash256(code)
		codes[i] = code

		n := len(args)
		values = append(values, fmt.Sprintf("($%d, $%d)", n+1, n+2))
		args = append(args, codeHash, userID)
	}

	// Workaround to suppress false IDE SQL parsing errors
	query := `
	INSERT INTO two_factor_recovery_codes
		(recovery_code_hash, user_id)
	VALUES ($0, $0)
`
	query = strings.ReplaceAll(query, "($0, $0)", strings.Join(values, ","))

	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		log.Println(err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &VerifyTwoFactorResult{Codes: codes}, nil
}

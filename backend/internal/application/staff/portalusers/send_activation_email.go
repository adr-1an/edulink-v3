package portalusers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"app/internal/cryptoutil"
	"app/internal/mailer"
	"app/internal/permissions"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/wneessen/go-mail"
)

type SendActivationEmailInput struct {
	UserID       int64
	PortalUserID int64
}

func validateActivationEmailState(accountEnabled, accountActive bool) error {
	if !accountEnabled {
		return ErrAccountDisabled
	}
	if accountActive {
		return ErrAccountAlreadyActive
	}
	return nil
}

func (s *Service) SendActivationEmail(ctx context.Context, i *SendActivationEmailInput) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var targetEmail string
	var schoolID int64
	var name string
	var accEnabled bool
	var accActive bool
	if err := tx.QueryRowContext(ctx, `
		SELECT email, school_id, name, account_enabled, account_active
		FROM portal_users
		WHERE id = $1
	`, i.PortalUserID).Scan(&targetEmail, &schoolID, &name, &accEnabled, &accActive); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrForbidden
		}
		log.Println(err)
		return err
	}

	if !permissions.Can(permissions.PermissionPortalUserActivate, i.UserID, schoolID, ctx, s.DB) {
		return ErrForbidden
	}

	if err := validateActivationEmailState(accEnabled, accActive); err != nil {
		return err
	}

	mailQueue := make([]mailer.MailQueue, 0, 1)

	token, err := gonanoid.New(128)
	if err != nil {
		log.Println(err)
		return err
	}
	tokenHash := cryptoutil.MakeHash256(token)

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO portal_account_activation_tokens (token_hash, portal_user_id)
		VALUES ($1, $2)
	`, tokenHash, i.PortalUserID); err != nil {
		log.Println(err)
		return err
	}

	frontendUrl := os.Getenv("FRONTEND_URL")
	if frontendUrl == "" {
		log.Println("FRONTEND_URL env var not set.")
	}
	activationUrl := fmt.Sprintf("%s/app/portal/auth/activate/%s", frontendUrl, token)

	body := fmt.Sprintf(`
Hello %s,
Here's your account activation link:
%s
`, name, activationUrl)

	purpose := mailer.MailPurposePortalAccountActivation
	mailQueue = append(mailQueue, mailer.MailQueue{
		To:          targetEmail,
		Subject:     "Activate your account",
		ContentType: mail.TypeTextPlain,
		Importance:  mail.ImportanceNormal,
		Body:        body,
		SendAt:      nil,
		Purpose:     &purpose,
		MaxRetries:  nil,
	})

	if err := mailer.AppendMailQueue(mailQueue, s.Sf, tx, ctx); err != nil {
		log.Println(err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

package profile

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	helpers2 "app/internal/helpers"
	staffhelpers "app/internal/helpers/staff"

	"github.com/wneessen/go-mail"
)

type UpdateEmailInput struct{ Token string }

func (s *Service) UpdateEmail(ctx context.Context, i *UpdateEmailInput) error {
	tokenHash := helpers2.MakeHash256(i.Token)

	// Get old email
	var oldEmail string
	var newEmail string
	if err := s.DB.QueryRowContext(ctx, `
		SELECT u.email, vt.email_change_new_email
		FROM verification_tokens vt
		JOIN users u ON u.id = vt.user_id
		WHERE token_hash = $1
	`, tokenHash).Scan(&oldEmail, &newEmail); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		} else {
			log.Println(err)
			return err
		}
	}

	// Delete token & update user
	var res string
	err := s.DB.QueryRowContext(ctx, `
		WITH token AS (
		    SELECT user_id, email_change_new_email
		    FROM verification_tokens
		    WHERE token_hash = $1
				AND purpose = $2
				AND created_at >= NOW() - INTERVAL '24 hours'
		),
		email_taken AS (
		    SELECT 1
		    FROM users u
		    JOIN token t ON true
		    WHERE u.email = t.email_change_new_email
		    	AND u.id <> t.user_id
		),
		deleted AS (
		    DELETE FROM verification_tokens
			WHERE token_hash = $1
			  	AND purpose = $2
		        AND EXISTS (SELECT 1 FROM token)
		        AND NOT EXISTS (SELECT 1 FROM email_taken)
		    RETURNING user_id
		),
		updated AS (
		    UPDATE users u
		   	SET email = t.email_change_new_email
		   	FROM token t
	   		JOIN deleted d ON d.user_id = t.user_id
		   	RETURNING 1
		)
		SELECT CASE 
			WHEN EXISTS (SELECT 1 FROM updated) THEN 'updated'
			WHEN EXISTS (SELECT 1 FROM email_taken) THEN 'email_taken'
			ELSE 'invalid_token'
		END AS result
	`, tokenHash, staffhelpers.TokenPurposeEmailChange).Scan(&res)
	if err != nil {
		log.Println(err)
		return err
	}

	switch res {
	case "updated":
	case "email_taken":
		return ErrConflict
	case "invalid_token":
		return ErrNotFound
	default:
		return errors.New("unexpected email update result")
	}

	if res == "updated" {
		msg := fmt.Sprintf(`Your account email has been changed.
Old email: %s
New email: %s`, oldEmail, newEmail)
		emailOld := helpers2.Mail{
			To:          oldEmail,
			Subject:     "Email changed",
			ContentType: mail.TypeTextPlain,
			Importance:  mail.ImportanceNormal,
			Body:        msg,
		}
		emailNew := helpers2.Mail{
			To:          newEmail,
			Subject:     "Email changed",
			ContentType: mail.TypeTextPlain,
			Importance:  mail.ImportanceNormal,
			Body:        msg,
		}

		go func() {
			if err := helpers2.SendMail(emailOld); err != nil {
				log.Println(err)
			}
			if err := helpers2.SendMail(emailNew); err != nil {
				log.Println(err)
			}
		}()
	}

	return nil
}

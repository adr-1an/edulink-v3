package profile

import (
	"context"
	"fmt"
	"log"
	"os"

	helpers2 "app/internal/helpers"

	"github.com/alexedwards/argon2id"
	"github.com/wneessen/go-mail"
)

type ChangePasswordInput struct {
	UserID      int64
	Password    string
	NewPassword string
}

func (s *Service) ChangePassword(ctx context.Context, i *ChangePasswordInput) error {
	userID := i.UserID

	// Validate
	if len(i.NewPassword) < 8 {
		return ErrInvalidData
	}

	// Compare passwords
	var uName string
	var uEmail string
	var currentHash string
	if err := s.DB.QueryRowContext(ctx, `
		SELECT name, email, password_hash FROM users WHERE id = $1
	`, userID).Scan(&uName, &uEmail, &currentHash); err != nil {
		log.Println(err)
		return err
	}

	match, err := argon2id.ComparePasswordAndHash(i.Password, currentHash)
	if err != nil {
		log.Println(err)
		return err
	}

	if !match {
		return ErrUnauthorized
	}

	// Delete all sessions
	if _, err := s.DB.ExecContext(ctx, `
		UPDATE sessions
		SET revoked_at = NOW(), revoke_note = 'Password change'
		WHERE user_id = $1
	`, userID); err != nil {
		log.Println(err)
		return err
	}

	// Generate hash
	newHash, err := argon2id.CreateHash(i.NewPassword, argon2id.DefaultParams)
	if err != nil {
		log.Println(err)
		return err
	}

	// Change password
	if _, err := s.DB.ExecContext(ctx, `
		UPDATE users SET password_hash = $1 WHERE id = $2
	`, newHash, userID); err != nil {
		log.Println(err)
		return err
	}

	// Send email notif
	appName := os.Getenv("APP_NAME")
	msg := fmt.Sprintf(`Hello %s,

Your %s account password has been changed.
If this wasn't you, review your account security immediately.
`, uName, appName)
	email := helpers2.Mail{
		To:          uEmail,
		Subject:     "Password changed",
		ContentType: mail.TypeTextPlain,
		Importance:  mail.ImportanceNormal,
		Body:        msg,
	}
	go func() {
		if err := helpers2.SendMail(email); err != nil {
			log.Println(err)
		}
	}()

	return nil
}

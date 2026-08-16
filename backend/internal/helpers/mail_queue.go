package helpers

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sony/sonyflake/v2"
	"github.com/wneessen/go-mail"
)

type Mail struct {
	To          string
	Subject     string
	ContentType mail.ContentType
	Importance  mail.Importance
	Body        string
}

// SendMail is a helper function for sending custom emails.
func SendMail(email Mail) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpUser := os.Getenv("SMTP_USER")
	smtpFrom := os.Getenv("SMTP_FROM")
	smtpPass := os.Getenv("SMTP_PASS")

	// Create message
	m := mail.NewMsg()
	if err := m.From(smtpFrom); err != nil {
		return err
	}
	if err := m.To(email.To); err != nil {
		return err
	}

	// Set subject, body and importance
	m.Subject(email.Subject)
	m.SetImportance(email.Importance)
	m.SetBodyString(email.ContentType, email.Body)

	// Send email
	client, err := mail.NewClient(smtpHost, mail.WithPort(587),
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
		mail.WithUsername(smtpUser), mail.WithPassword(smtpPass))
	if err != nil {
		return err
	}

	if err := client.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

type MailQueuePurpose string

const (
	MailPurposeStudentAccountCreation MailQueuePurpose = "student_account_creation"
)

type MailQueue struct {
	To          string
	Subject     string
	ContentType mail.ContentType
	Importance  mail.Importance
	Body        string
	SendAt      *time.Time
	Purpose     *MailQueuePurpose
	MaxRetries  *int
}

func AppendMailQueue(emails []MailQueue, sf *sonyflake.Sonyflake, db *sql.Tx, ctx context.Context) error {
	type email struct {
		ID          int64
		To          string
		Subject     string
		ContentType mail.ContentType
		Importance  mail.Importance
		Body        string
		SendAt      *time.Time
		Purpose     *MailQueuePurpose
		MaxRetries  *int
	}
	messages := make([]email, 0, len(emails))

	// The number of arguments one insert query will take
	argsLen := 9

	placeholders := make([]string, 0, len(emails))
	args := make([]any, 0, len(emails)*argsLen)

	for i, e := range emails {
		n := i*argsLen + 1

		// Generate ID
		newID, err := sf.NextID()
		if err != nil {
			return err
		}

		// Add to email list
		m := email{
			ID:          newID,
			To:          e.To,
			Subject:     e.Subject,
			ContentType: e.ContentType,
			Importance:  e.Importance,
			Body:        e.Body,
			SendAt:      e.SendAt,
			Purpose:     e.Purpose,
			MaxRetries:  e.MaxRetries,
		}
		messages = append(messages, m)

		// Generate placeholders & args
		placeholders = append(
			placeholders,
			fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, COALESCE($%d, NOW()), $%d, COALESCE($%d, 1))",
				n, n+1, n+2, n+3, n+4, n+5, n+6, n+7, n+8,
			),
		)

		args = append(args,
			m.ID,
			m.To,
			m.Subject,
			m.ContentType,
			m.Importance,
			m.Body,
			m.SendAt,
			m.Purpose,
			m.MaxRetries,
		)
	}

	query := `
INSERT INTO email_queue (id, send_to, subject, content_type, priority, body, send_at, purpose, max_retries)
VALUES ($0,$0)`
	query = strings.ReplaceAll(query, "($0,$0)", strings.Join(placeholders, ","))

	// Insert queued emails
	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		return err
	}

	return nil
}
